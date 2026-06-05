package sdk_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/bketelsen/tl/internal/repo"
	"github.com/bketelsen/tl/sdk"
)

// newLedger creates an initialized .tl ledger in a temp dir and returns its path.
func newLedger(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	ledger, err := repo.Init(dir)
	if err != nil {
		t.Fatalf("repo.Init: %v", err)
	}
	if _, err := os.Stat(ledger); err != nil {
		t.Fatalf("ledger dir not created: %v", err)
	}
	return ledger
}

// mkTask creates and persists a task purely through the sdk surface.
func mkTask(t *testing.T, ledger, title string, deps ...string) *sdk.Task {
	t.Helper()
	id, err := sdk.NewID(ledger)
	if err != nil {
		t.Fatalf("NewID: %v", err)
	}
	now := time.Now().UTC().Truncate(time.Second)
	task := &sdk.Task{
		ID:        id,
		Title:     title,
		Status:    "open",
		Priority:  "medium",
		Type:      "task",
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: "sdk-test",
		DependsOn: deps,
		Tags:      []string{},
	}
	if err := sdk.Write(ledger, task); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := sdk.AppendEvent(ledger, sdk.Event{Event: "created", TaskID: id, Actor: "sdk-test"}); err != nil {
		t.Fatalf("AppendEvent: %v", err)
	}
	return task
}

// TestLifecycleThroughSDK exercises the create -> ready -> dep-gate -> claim ->
// close flow entirely via the public sdk package, proving an external consumer
// (e.g. omnius) can drive the ledger without importing internal/.
func TestLifecycleThroughSDK(t *testing.T) {
	ledger := newLedger(t)
	now := time.Now().UTC()

	// A blocker task and a dependent task.
	blocker := mkTask(t, ledger, "blocker")
	dependent := mkTask(t, ledger, "needs blocker", blocker.ID)

	// Round-trip read returns the same data.
	got, err := sdk.Read(ledger, blocker.ID)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.Title != "blocker" {
		t.Fatalf("round-trip title = %q, want %q", got.Title, "blocker")
	}

	// Blocker (no deps) is ready; dependent is NOT ready (dep not done).
	if !sdk.IsReady(blocker, ledger, now) {
		t.Errorf("blocker should be ready")
	}
	if sdk.IsReady(dependent, ledger, now) {
		t.Errorf("dependent should NOT be ready while blocker is open")
	}

	// CheckDeps on the dependent returns a typed DepError (not done).
	err = sdk.CheckDeps(ledger, dependent)
	if err == nil {
		t.Fatalf("CheckDeps should fail while blocker is open")
	}
	var de *sdk.DepError
	if !errors.As(err, &de) {
		t.Fatalf("CheckDeps error = %T (%v), want *sdk.DepError", err, err)
	}
	if de.Missing {
		t.Errorf("DepError.Missing = true, want false (dep exists but not done)")
	}
	if de.DepID != blocker.ID {
		t.Errorf("DepError.DepID = %q, want %q", de.DepID, blocker.ID)
	}

	// Close the blocker (mark done), persist, and re-check.
	blocker.Status = "done"
	blocker.UpdatedAt = time.Now().UTC().Truncate(time.Second)
	if err := sdk.Write(ledger, blocker); err != nil {
		t.Fatalf("Write(done blocker): %v", err)
	}
	if err := sdk.AppendEvent(ledger, sdk.Event{Event: "closed", TaskID: blocker.ID, Actor: "sdk-test"}); err != nil {
		t.Fatalf("AppendEvent(closed): %v", err)
	}

	// Now the dependent's deps are satisfied and it is ready.
	if err := sdk.CheckDeps(ledger, dependent); err != nil {
		t.Errorf("CheckDeps should pass after blocker done, got: %v", err)
	}
	if !sdk.IsReady(dependent, ledger, now) {
		t.Errorf("dependent should be ready after blocker done")
	}

	// Claim the dependent purely via field mutation + Write.
	actor := "agent-1"
	claimAt := time.Now().UTC().Truncate(time.Second)
	expires := claimAt.Add(60 * time.Minute)
	dependent.Claim = sdk.Claim{Actor: &actor, ClaimedAt: &claimAt, ExpiresAt: &expires, HeartbeatAt: &claimAt}
	dependent.Status = "in_progress"
	dependent.UpdatedAt = claimAt
	if err := sdk.Write(ledger, dependent); err != nil {
		t.Fatalf("Write(claim): %v", err)
	}

	// While actively claimed, it is no longer "ready" to a different actor.
	reread, err := sdk.Read(ledger, dependent.ID)
	if err != nil {
		t.Fatalf("Read(claimed): %v", err)
	}
	if sdk.IsReady(reread, ledger, claimAt.Add(1*time.Minute)) {
		t.Errorf("actively-claimed task should not be ready")
	}

	// CheckDeps for a missing dependency yields Missing=true.
	orphan := &sdk.Task{ID: "task-zzz", Status: "open", DependsOn: []string{"task-nope"}}
	err = sdk.CheckDeps(ledger, orphan)
	var de2 *sdk.DepError
	if !errors.As(err, &de2) || !de2.Missing {
		t.Errorf("CheckDeps(missing dep) = %v, want DepError{Missing:true}", err)
	}

	// Audit journal recorded our events.
	evs, err := sdk.ReadEvents(ledger)
	if err != nil {
		t.Fatalf("ReadEvents: %v", err)
	}
	if len(evs) < 3 {
		t.Errorf("expected >=3 events (2 created + 1 closed), got %d", len(evs))
	}

	// List returns both tasks.
	all, err := sdk.List(ledger)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("List len = %d, want 2", len(all))
	}
}
