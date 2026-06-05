// Package ready holds the task-readiness and dependency-check logic shared by
// the CLI commands and the public SDK, so both agree on the single rule
// "ready = open/unclaimed + all dependencies done" by construction.
//
// This logic was originally inline in cmd/ready.go (isReady) and cmd/claim.go
// (checkDeps); it was promoted here so the SDK can expose the same predicate the
// CLI uses, rather than re-implement (and risk drifting from) it. The CLI-only
// exit-code coupling (cmd.NewExitError) is intentionally NOT part of this
// package — CheckDeps returns a typed *DepError that the CLI translates to an
// exit code at its boundary, and that the SDK can inspect structurally.
package ready

import (
	"errors"
	"fmt"
	"time"

	"github.com/bketelsen/tl/internal/store"
	"github.com/bketelsen/tl/internal/task"
)

// DepError describes why a task's dependencies are not satisfied. Missing is
// true when the dependency task does not exist; otherwise the dependency exists
// but is not yet done (Status holds its current status). Callers that need an
// exit code (the CLI) can switch on Missing; callers that just need the error
// (the SDK / an orchestrator) can use it as a plain error.
type DepError struct {
	TaskID  string // the dependent task
	DepID   string // the unsatisfied dependency
	Missing bool   // true: DepID does not exist; false: DepID exists but is not done
	Status  string // the dependency's current status when it exists (else "")
}

func (e *DepError) Error() string {
	if e.Missing {
		return fmt.Sprintf("task %s depends on %s which does not exist", e.TaskID, e.DepID)
	}
	return fmt.Sprintf("task %s is not ready: dependency %s is not done (status %s)", e.TaskID, e.DepID, e.Status)
}

// IsReady reports whether t can be claimed now: it must be open (or in_progress
// only if its claim has expired), must not be under an active claim, and every
// dependency must be done. now should be a UTC time.
func IsReady(t *task.Task, ledger string, now time.Time) bool {
	// Must be open or in_progress (in_progress is only ready if claim expired).
	switch t.Status {
	case "open", "in_progress":
	default:
		return false // blocked, pending_human, done, cancelled
	}
	// Must not have an active claim (actor set and not expired).
	if t.Claim.Actor != nil && t.Claim.ExpiresAt != nil && t.Claim.ExpiresAt.After(now) {
		return false
	}
	// All dependencies must be done.
	for _, depID := range t.DependsOn {
		dep, err := store.Read(ledger, depID)
		if err != nil || dep.Status != "done" {
			return false
		}
	}
	return true
}

// CheckDeps verifies every dependency of t has status "done". It returns nil
// when all dependencies are satisfied, or a *DepError identifying the first
// unsatisfied one. A read error other than "not found" is returned as-is.
func CheckDeps(ledger string, t *task.Task) error {
	for _, depID := range t.DependsOn {
		dep, err := store.Read(ledger, depID)
		if errors.Is(err, store.ErrTaskNotFound) {
			return &DepError{TaskID: t.ID, DepID: depID, Missing: true}
		}
		if err != nil {
			return err
		}
		if dep.Status != "done" {
			return &DepError{TaskID: t.ID, DepID: depID, Status: dep.Status}
		}
	}
	return nil
}
