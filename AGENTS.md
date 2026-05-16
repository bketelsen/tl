# Project entry points

- `README.md` — installation, quickstart, and CLI flag reference.
- `docs/PRD.md` — design intent, non-goals, status enum, exit codes.
- `features/` — Gherkin behavioral spec, one `.feature` file per command.

The `@implemented` tag on each feature marks which commands are actually
built; the godog suite under `bdd/` scopes to those. Untagged features still
serve as the binding contract for unimplemented commands.

## Workflow

- `make bdd` runs the godog suite; `make test` runs everything.
- Adding a command: extend its `.feature`, add step defs in `bdd/bdd_test.go`,
  implement the cobra command under `cmd/<name>.go` (plus any needed package
  under `internal/`), then tag the feature `@implemented`.
- Mutating commands append a JSON line to `.taskledger/events.jsonl`. Read
  commands must support `--json`.

## Gherkin / BDD tests

When writing or changing `.feature` files, follow `docs/gherkin-guidelines.md`.

Rules:
- Write behavior, not implementation.
- Use one behavior per scenario.
- Keep scenarios independent.
- Use concrete examples.
- Do not write vague `Then` steps.
- Do not include step definitions in feature files unless explicitly requested.
