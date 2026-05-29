---
id: task-reg
title: Add references field to task frontmatter
status: open
priority: medium
type: feature
created_at: 2026-05-29T11:29:22Z
updated_at: 2026-05-29T11:29:28Z
created_by: human
assignee: null
depends_on: []
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags:
  - schema
  - cli
---

## Description


Implement Option C from the traceability analysis: add an optional 'references' list to the task frontmatter, with CLI flags for create and refine, display support, and validation in 'tl doctor'.

## What to implement

### 1. Schema change
Add `References []string` field to the Task struct in internal/task/task.go with yaml:"references" and json:"references" tags. Optional (omitempty).

### 2. CLI: tl create
- `--ref <path>` flag (repeatable, like `--tag`)
- Sets initial references on the new task

### 3. CLI: tl refine
- `--add-ref <path>` flag (repeatable) — appends a reference
- `--remove-ref <path>` flag (repeatable) — removes a reference
- Follows the same pattern as tl dep add/remove for idempotency

### 4. Display: tl show
- Renders a `References:` section listing all references
- JSON output includes the `references` field

### 5. Validation: tl doctor
- Detect referenced file paths that do not exist (dead link detection)
- Optionally clear dead references with `--fix`

### 6. Events
- `tl refine --add-ref` records event `reference_added`
- `tl refine --remove-ref` records event `reference_removed`

## Process
- Write Gherkin feature file first (features/references.feature)
- Follow BDD approach: Gherkin → implement step defs → implement command
- Tag @implemented when done

## Notes

- 2026-05-29T11:29:26Z [main-pc] note: Created from traceability analysis. Implements Option C: references field as plain string list. Follows BDD approach — Gherkin first, then implementation. Connected to task-4sh (tl doctor design) as references validation will be part of tl doctor.
