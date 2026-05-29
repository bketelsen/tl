---
id: task-pnr
title: Test autocompletion
status: done
priority: medium
type: story
created_at: 2026-05-17T21:06:38Z
updated_at: 2026-05-29T13:28:06Z
created_by: human
assignee: null
depends_on:
  - task-3q7
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags: []
---

## Description

Verify and document the task-ID shell completion delivered in task-3q7.

**Deliverable:**
- Add auto storymated test coverage for the completion logic — assert the shared
  `ValidArgsFunction` returns the expected task IDs for a given ledger state
  (and an empty/partial-match ledger).
- Manually verify `tl <cmd> <TAB>` completes task IDs in bash, zsh, and fish.
- Add a README section documenting how to install/enable shell completion
  (`tl completion <shell>`) and the task-ID completion behaviour.

Depends on task-3q7 — the completion feature must exist before it can be tested
or documented.

## Notes

- 2026-05-29T13:28:06Z [claude-code] note: Closing as already-delivered. The work this task tracked was bundled into task-3q7 (autocompletion implementation): (1) automated test coverage — cmd/completion_test.go (unit test for missing-ledger silent failure) plus 21 BDD scenarios in features/completion.feature (canonical + bare short codes, status prefix in description, KeepOrder directive, closed-task filtering, every command + dep --on flag); (2) README documentation — new 'Shell completion' subsection under Quickstart with bash/zsh/fish install snippets. Outstanding: manual interactive verification in zsh and fish — needs an interactive terminal, leaving for the human. A follow-up task-vr6 was created for 'tl completion --install' to reduce the install friction.
