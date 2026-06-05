package sdk

import "github.com/bketelsen/tl/internal/repo"

// InitLedger creates a new .tl ledger under dir (a tasks/ dir, an empty events
// journal, and a default config) and returns the ledger directory path. It
// errors if a ledger already exists there. This is the programmatic equivalent
// of `tl init`, needed before an external consumer can create tasks.
func InitLedger(dir string) (ledger string, err error) {
	return repo.Init(dir)
}

// LedgerName is the ledger directory name (".tl") within a repository.
const LedgerName = repo.LedgerDir
