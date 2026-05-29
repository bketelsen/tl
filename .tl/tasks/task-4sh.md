---
id: task-4sh
title: Design and implement tl doctor command
status: open
priority: high
type: feature
created_at: 2026-05-29T11:16:22Z
updated_at: 2026-05-29T13:33:23Z
created_by: human
assignee: null
depends_on: []
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags:
  - cli
---

## Description


Design and implement the `tl doctor` command — a ledger diagnostic tool that scans task files, events, config, and filesystem for integrity issues.

## Scope (agreed with human)

### Diagnostics categories (all in v1):
- **Frontmatter**: malformed YAML, unknown status/priority/type values, missing required fields
- **Identity**: duplicate task IDs across the ledger
- **Dependencies**: missing depends_on targets, cyclic deps (A→B→A), self-dependency
- **Events**: orphaned events (ref nonexistent task), tasks with no events in journal
- **Claims**: in_progress with no claim, expired claim not released, open with stale claim data
- **Timestamps**: created_at > updated_at, timestamps in the future, claim expiry before claim time
- **Filesystem**: orphaned .md.tmp files, corrupted task files that can't be parsed
- **Config**: invalid or missing config.yaml
- **Body**: merge conflict markers (<<<<<<<, =======, >>>>>>>), malformed notes format
- **Scale warning**: warn when >100 tasks or >1000 events exist

### Output model:
- Human-readable grouped report (grouped by category)
- Exit 0 always (doctor is diagnostic, not a failure); non-zero only if doctor itself fails
- --json support: emit array of diagnostic objects [{severity, category, task_id, message, fixable}]
- --fix mode: attempt auto-repair where possible

### Relationship to other commands:
- Does NOT subsume tl stale — orthogonal

### Process:
1. First: write and agree the Gherkin feature file (features/doctor.feature)
2. Second: implement the cmd/doctor.go and any needed internal/ packages
3. Third: add BDD step definitions
4. Tag @implemented when done

## Notes

- 2026-05-29T11:16:43Z [main-pc] note: Drafted initial Gherkin feature file at features/doctor.feature with 31 scenarios covering all agreed diagnostic categories, JSON output shape, and --fix behavior. Key design points captured in the task body. Ready for human review and agreement on the Gherkin before implementation.
- 2026-05-29T11:23:48Z [main-pc] note: Deep analysis: the artifact-traceability gap and the proposed references field
- 2026-05-29T11:23:51Z [main-pc] note: "Deep analysis: the artifact-traceability gap. See analysis in parent message."
