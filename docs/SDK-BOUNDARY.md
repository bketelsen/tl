# SDK Boundary

This repository is a fork of [`github.com/aholbreich/tl`](https://github.com/aholbreich/tl)
(MIT, Copyright (c) 2026 Alexander Holbreich; see `LICENSE`). The fork's module
path is `github.com/bketelsen/tl`.

The fork's purpose beyond upstream: expose a small **public Go SDK** (`sdk/`) so
that external modules — notably an agent orchestrator (omnius) — can read and
write the same `.tl/` ledger the `tl` CLI manages, sharing **one** implementation
of the on-disk format, the readiness rule, and the file lock, instead of
maintaining a divergent copy that could drift.

## The boundary

```
sdk/                         PUBLIC — the only thing external modules import
  re-exports task + store + events + ready (type aliases + function values)

internal/task                domain model + Markdown/YAML (de)serialization  [public via sdk]
internal/store               task-file I/O: Read/Write/List/NewID/...         [public via sdk]
internal/events              .tl/events.jsonl audit journal                  [public via sdk]
internal/ready               IsReady + CheckDeps (the dependency rule)        [public via sdk]
internal/repo                .tl/ layout constants + config (transitive dep)  [private]
internal/lock                advisory file lock for mutating commands         [private]
internal/color               ANSI presentation helpers                        [private]
internal/doctor              ledger integrity scan/repair                     [private]
cmd/                         cobra CLI; sits on the same core packages        [private]
```

Internal package DAG (acyclic, shallow):

```
task   -> (nothing)
repo   -> (nothing)
events -> repo
store  -> repo, task
ready  -> store, task          # NEW in the fork (promoted from cmd/)
doctor -> task, store, events, repo
```

The transitive closure of the SDK-exposed set `{task, store, events, ready}` is
exactly `{task, store, events, ready, repo}`. `lock`, `color`, `doctor`, and
`cmd/` stay private.

## Why a re-export shim (not a vendored copy, not a bulk rename)

- **Same-module `internal/` import is allowed.** Go's internal rule blocks only
  *cross-module* imports; a package at `github.com/bketelsen/tl/sdk` may import
  `github.com/bketelsen/tl/internal/...` because they share the module root.
  So `sdk/` re-exports the internal packages and external code imports `sdk/`.
- **Upstream files stay unmoved** (with one deliberate exception, below), so a
  future `git merge upstream` (after `git remote add upstream
  https://github.com/aholbreich/tl`) stays clean.
- **Type aliases, not wrappers.** `type Task = task.Task`, so `sdk.Task` *is*
  `task.Task` — values pass between the SDK and the underlying store/events
  functions with no conversion.

## The one promoted piece: the readiness rule

The readiness predicate (`isReady`) and dependency check (`checkDeps`) lived
only in `cmd/` (in `cmd/ready.go` and `cmd/claim.go`) and were unexported. They
are the single most valuable thing for an orchestrator ("what is ready now?"),
so they were **promoted** into `internal/ready` and the CLI rewired to call it.
This is a deliberate trade: it modifies two upstream files (`cmd/ready.go`,
`cmd/claim.go`), so those two files may conflict on a future upstream merge — in
exchange, the CLI and the SDK share **one** definition of the readiness rule and
cannot silently diverge on it (the exact format/semantics skew the SDK fork was
chosen to prevent).

The CLI-only exit-code coupling (`cmd.NewExitError`) is NOT in `internal/ready`.
`CheckDeps` returns a typed `*ready.DepError`; the CLI translates it to an exit
code at its boundary (`cmd/claim.go`), and the SDK exposes it as a plain,
inspectable error.

## Public API (`sdk/`)

Re-exported types (aliases): `Task`, `Claim`, `Pending`, `Note`, `ParsedBody`,
`Event`, `DepError`.

Re-exported functions: `LedgerDir`, `NewID`, `NormalizeID`, `TaskPath`, `Read`,
`Write`, `List`, `AppendEvent`, `ReadEvents`, `IsReady`, `CheckDeps`,
`SetDescription`, `AppendNote`, `ParseBody`. Error sentinels: `ErrLedgerNotFound`,
`ErrTaskNotFound`.

`sdk/sdk_test.go` is an external-style (`package sdk_test`) contract test that
drives create -> ready -> dep-gate -> claim -> close -> audit/list entirely
through `sdk`, with no `internal/` access — the guarantee external consumers buy.

## Maintaining the fork

1. `git remote add upstream https://github.com/aholbreich/tl` (once).
2. To sync: `git fetch upstream && git merge upstream/main` (or rebase).
   Expect clean merges except: the module-path rename touches every `.go`
   import line, and the readiness promotion touches `cmd/ready.go` +
   `cmd/claim.go`. Resolve those by hand; everything else should apply cleanly.
3. Tag releases (`vMAJOR.MINOR.PATCH`); downstream (omnius) pins a tag and
   upgrades deliberately.

## Concurrency note for SDK consumers

The `tl` CLI guards mutating commands with an advisory file lock
(`internal/lock`, `gofrs/flock`). An SDK consumer that writes the ledger
concurrently with the CLI (or another agent) should acquire the same lock to
avoid write races. `internal/lock` is currently private; if a consumer needs
coordinated writes, promote a locking helper into the SDK (future work).
