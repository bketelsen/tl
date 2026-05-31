---
id: task-lrv
title: Teach tl doctor to repair concatenated event journal lines
status: done
priority: high
type: bug
created_at: 2026-05-31T18:10:37Z
updated_at: 2026-05-31T18:16:13Z
created_by: pi:doctor-events
assignee: null
depends_on: []
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags:
  - cli
references:
  - internal/doctor/doctor.go
  - internal/doctor/doctor_test.go
  - features/doctor.feature
---

## Description

Detect malformed events.jsonl lines that contain multiple concatenated JSON event objects, report them as events diagnostics, and let tl doctor --fix split them back into one event per line so tl history works again.

## Notes

- 2026-05-31T18:14:30Z [pi:doctor-events] note: Implemented doctor detection/repair for concatenated events.jsonl JSON objects. Added JSON decoder-based event line scan, fix path that rewrites concatenated objects as one object per line, unit tests, BDD scenarios/steps, and ran go run . doctor --fix on this repo to repair the task-4sh history break. Validation: go test ./internal/doctor ./cmd ./bdd, make bdd, make test, git diff --check, and go run . history task-4sh all pass.
- 2026-05-31T18:16:13Z [pi:doctor-events] note: Follow-up during validation: doctor --fix initially rewrote JSONL without a trailing newline, so the next event append could re-concatenate. Fixed internal/events.Append to insert a separator when the journal lacks a trailing newline, and updated doctor normalization/tests to keep a trailing newline. Re-ran doctor --fix; task-4sh history works and all event lines parse.
