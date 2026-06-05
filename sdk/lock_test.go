package sdk_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bketelsen/tl/internal/repo"
	"github.com/bketelsen/tl/sdk"
)

// TestSDKLockExcludesAndMatchesCLIPath proves the two properties an external
// consumer relies on:
//  1. the SDK lock is mutually exclusive (a held lock blocks a second acquire), and
//  2. it locks the SAME file the CLI locks (<ledger>/.lock), so SDK and CLI
//     writers coordinate rather than each holding a private, useless lock.
func TestSDKLockExcludesAndMatchesCLIPath(t *testing.T) {
	dir := t.TempDir()
	ledger, err := repo.Init(dir)
	if err != nil {
		t.Fatalf("repo.Init: %v", err)
	}

	// Property 2 (path): the SDK exposes the same lock file name the CLI uses.
	if sdk.LockFile != ".lock" {
		t.Fatalf("sdk.LockFile = %q, want %q", sdk.LockFile, ".lock")
	}

	// Acquire via the SDK.
	release, err := sdk.AcquireLock(ledger)
	if err != nil {
		t.Fatalf("AcquireLock: %v", err)
	}

	// The lock file is created at <ledger>/.lock — the exact path the CLI locks.
	lockPath := filepath.Join(ledger, sdk.LockFile)
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("lock file %s should exist while held: %v", lockPath, err)
	}

	// Property 1 (exclusion): a second acquire with a short timeout must FAIL
	// while the first is held, and must actually have waited the timeout.
	start := time.Now()
	r2, err2 := sdk.AcquireLockWithTimeout(ledger, 200*time.Millisecond)
	elapsed := time.Since(start)
	if err2 == nil {
		_ = r2()
		t.Fatalf("second AcquireLock succeeded while the first was held")
	}
	if elapsed < 150*time.Millisecond {
		t.Fatalf("second acquire returned too fast (%s); did it actually wait on the lock?", elapsed)
	}

	// Release the first; a fresh acquire must now succeed.
	if err := release(); err != nil {
		t.Fatalf("release: %v", err)
	}
	r3, err3 := sdk.AcquireLock(ledger)
	if err3 != nil {
		t.Fatalf("AcquireLock after release should succeed: %v", err3)
	}
	if err := r3(); err != nil {
		t.Fatalf("final release: %v", err)
	}
}
