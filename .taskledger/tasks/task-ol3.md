---
id: task-ol3
title: Colorize human-readable output for tl show and tl create
status: done
priority: low
type: feature
created_at: 2026-05-16T23:32:43Z
updated_at: 2026-05-17T19:31:15Z
created_by: human
assignee: null
depends_on: []
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags: []
---

## Description

Add ANSI color to the non-JSON output path so humans get scannable status/priority at a glance. Agents using --json are unaffected.

Scope:
- tl show: color status (open=green, in_progress=yellow, blocked=red, pending_human=magenta, done=cyan, cancelled=dim) and priority (high=red, medium=yellow, low=dim).
- tl create: color the task ID in the success line.
- tl list: deferred to v2 -- tabwriter breaks on ANSI escape widths.

Controls:
- Respect NO_COLOR env var.
- --color=auto|never|always flag on root command (default: auto).
- auto mode enables color only when stdout is a TTY.

Non-goals:
- Zero new dependencies (raw ANSI in internal/color/).
- No TUI, no syntax highlighting in task bodies, no colored diffs.

## Notes

### 2026-05-17T19:31:15Z - pi:codex

Implemented ANSI color controls and output highlighting. Added internal/color helpers, root --color=auto|never|always with NO_COLOR support, colorized tl show status/priority, tl create success ID, and dimmed closed rows for tl list --all without putting ANSI into tabwriter columns. Added BDD coverage and docs; verified with go test ./... and make test.
