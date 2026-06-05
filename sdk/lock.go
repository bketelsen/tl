package sdk

import (
	"time"

	"github.com/bketelsen/tl/internal/lock"
)

// Locking
//
// The ledger has no transactions; safe concurrent mutation depends on an
// advisory file lock at <ledger>/.lock. The `tl` CLI acquires this same lock
// (via the same internal package and path) before every read-modify-write, so
// an SDK consumer that mutates the ledger concurrently with the CLI — or with
// another agent — MUST hold this lock around its own read-modify-write, or the
// two can clobber each other and defeat the claim-safety guarantee.
//
// Typical use:
//
//	release, err := sdk.AcquireLock(ledger)
//	if err != nil { return err }
//	defer release()
//	t, _ := sdk.Read(ledger, id)
//	// ...mutate t...
//	_ = sdk.Write(ledger, t)
//	_ = sdk.AppendEvent(ledger, ev)
//
// Read-only use (sdk.Read / sdk.List / sdk.IsReady / sdk.CheckDeps / sdk.ReadEvents)
// does not require the lock, since the ledger is plain files and a torn read of a
// single atomically-written task file cannot occur (Write uses temp-file+rename).
// The lock is only needed to serialize writers.

// LockFile is the name of the lock file within a ledger directory (".lock").
// It is the exact path the CLI locks, so SDK and CLI writers coordinate.
const LockFile = lock.LockFile

// DefaultLockTimeout is how long AcquireLock waits before reporting contention.
const DefaultLockTimeout = lock.DefaultTimeout

// AcquireLock takes the ledger's exclusive advisory lock, waiting up to
// DefaultLockTimeout for a competing holder. The returned release func must be
// called (typically via defer) to release it; on Unix the kernel also releases
// the lock if the process exits without calling release. Returns an error on
// timeout (another writer held the lock too long).
func AcquireLock(ledger string) (release func() error, err error) {
	return lock.Acquire(ledger)
}

// AcquireLockWithTimeout is AcquireLock with a caller-chosen wait limit.
func AcquireLockWithTimeout(ledger string, timeout time.Duration) (release func() error, err error) {
	return lock.AcquireWithTimeout(ledger, timeout)
}
