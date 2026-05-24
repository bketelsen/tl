---
id: task-ro8
title: Rename default ledger directory to .tl
status: done
priority: high
type: task
created_at: 2026-05-24T16:45:24Z
updated_at: 2026-05-24T16:46:20Z
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

Change TaskLedger default storage directory to .tl across code, tests, docs, and feature specs.

## Notes

- 2026-05-24T16:46:20Z [pi:rename-ledger-dir] note: Renamed default ledger directory to .tl across code, BDD/unit tests, docs, agent snippet, gitignore, and the working ledger directory. Verified with make bdd, go test ./..., and make test.
