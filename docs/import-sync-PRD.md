# PRD: Import and external sync

**Status:** Draft for discussion  
**Last updated:** 2026-05-29

This document captures design intent, architectural decisions, and open
questions for bringing existing work into `tl` (import) and keeping it in
sync with external systems (sync).

For the core product thesis and non-goals see [`PRD.md`](PRD.md). For
per-command behavior (once implemented) see the `features/` directory.

---

## Table of Contents

1. [Philosophy](#1-philosophy)
2. [Architecture overview](#2-architecture-overview)
3. [Import: JSON pipe](#3-import-json-pipe)
4. [Import: Markdown](#4-import-markdown)
5. [Import: GitHub Issues](#5-import-github-issues)
6. [Import: JIRA, Linear, Trello (one-shot)](#6-import-jira-linear-trello-one-shot)
7. [Sync: Trello bidirectional](#7-sync-trello-bidirectional)
8. [Data model changes](#8-data-model-changes)
9. [Open questions](#9-open-questions)

---

## 1. Philosophy

### Import is migration, sync is explicit

There are two distinct use cases, and they share infrastructure but have
different contracts:

| | Import | Sync |
|---|---|---|
| **Nature** | One-shot migration | Ongoing alignment |
| **State** | Stateless (no tracking file) | Stateful (`.tl/sync.yaml` + `external` field on tasks) |
| **Direction** | External → tl | Bidirectional |
| **Idempotent** | No (re-run creates duplicates) | Yes (tracks remote IDs) |
| **Requires network** | Depends on source | Yes (explicit `tl sync` command) |

### Constraint: no automatic remote push

The PRD (see [`PRD.md`](PRD.md)) says "no automatic remote push." Import
and sync respect this:

- **Import** is a one-shot command (pull from external source).
- **Sync** is an explicit command (`tl sync`). It never runs on its own, in
  a background worker, as a cron job, or as a git hook. The user or agent
  consciously invokes it.

### Constraint: no daemon, no hidden database

- `.tl/sync.yaml` is a checked-in configuration file.
- The `external` field lives in each task's YAML frontmatter (part of the
  versioned task file).
- No separate sync database, no lock-in to a service.

---

## 2. Architecture overview

```
                     ┌─────────────────────────────────────────┐
                     │           tl import engine              │
                     │                                         │
                     │  ┌──────────┐   ┌───────────────────┐   │
                     │  │  Reader  │──▶│  Intermediate     │   │
                     │  │  (one    │   │  Format (JSON)    │   │
                     │  │   per    │   └────────┬──────────┘   │
                     │  │  source) │            │              │
                     │  └──────────┘            ▼              │
                     │                    ┌──────────┐         │
                     │                    │  Writer   │         │
                     │                    │ (creates  │         │
                     │                    │  task     │         │
                     │                    │  files)   │         │
                     │                    └──────────┘         │
                     └─────────────────────────────────────────┘
                                     │
                                     ▼
                     ┌─────────────────────────────────────────┐
                     │           tl sync engine                │
                     │                                         │
                     │  ┌──────────┐   ┌──────────┐   ┌──────┐ │
                     │  │  Match   │──▶│   Diff   │──▶│Apply │ │
                     │  │  (pair   │   │ (field   │   │(per  │ │
                     │  │  tasks ↔ │   │  by      │   │strat-│ │
                     │  │  cards)  │   │  field)  │   │egy)  │ │
                     │  └──────────┘   └──────────┘   └──────┘ │
                     └─────────────────────────────────────────┘
                                      │
                    ┌─────────────────┼──────────────────┐
                    │                 │                  │
                    ▼                 ▼                  ▼
              ┌──────────┐    ┌──────────────┐   ┌──────────────┐
              │  Reader  │    │   Adapter    │   │   Writer     │
              │ (Fetch   │    │ Transform:   │   │ (Apply to   │
              │  remote) │    │  card ↔ task │   │  remote)    │
              └──────────┘    └──────────────┘   └──────────────┘
```

There are two main entry points:

1. **`tl import <source>`** — one-shot pull from external. Reads from an
   external source, converts to the intermediate format, writes task files.
2. **`tl sync`** — bidirectional sync. Matches existing tl tasks to remote
   items, diffs, applies changes.

Both share the same **intermediate format** for task data and the same
**writer** for creating/updating task files.

---

## 3. Import: JSON pipe

The foundation. All other importers are either convenience wrappers around
this, or share the same writer.

### Interface

```bash
# Read from stdin
any-source --produce-json | tl import json

# Read from file
tl import json < tasks.json
tl import json --file tasks.json

# With mapping overrides
tl import json --status-map "todo=open,done=done"
tl import json --priority-map "P0=high,P1=high,P2=medium,P3=low"

# Dry-run: validate and report, write nothing
tl import json --dry-run

# Machine-readable result
tl import json --json
```

### Intermediate format

```json
{
  "version": 1,
  "tasks": [
    {
      "title": "Add login validation",
      "description": "Validate email format and require a password.",
      "status": "open",
      "priority": "high",
      "type": "feature",
      "tags": ["auth", "frontend"],
      "assignee": "alice",
      "created_at": "2026-05-01T00:00:00Z",
      "updated_at": "2026-05-15T00:00:00Z",
      "external_id": "gh:owner/repo#42",
      "notes": [
        {
          "time": "2026-05-10T00:00:00Z",
          "actor": "alice",
          "kind": "note",
          "message": "Started work on this"
        }
      ]
    }
  ]
}
```

**Required fields:** only `title`.  
**Optional fields:** everything else is optional or has sensible defaults.

### Writer behaviour

1. For each input task, generate a new tl ID.
2. Map `status` and `priority` through user-provided or default mappings.
3. Write task file with YAML frontmatter + Markdown body.
4. Append `created` event to `events.jsonl`.
5. If `external_id` is present and `depends_on` references other
   `external_id` values in the same batch, resolve after all tasks are
   created.

### Open questions

- **Q1:** Should `external_id` be persisted in the task? (Currently: no.
    Import is stateless. Re-importing duplicates.)
- **Q2:** Should the intermediate format support `depends_on` by batch
    index (position in the array) as well as by `external_id`?
- **Q3:** What is the max safe input size for stdin? (Gating factor: RAM
    for in-memory dependency resolution.)

---

## 4. Import: Markdown

File-based import for `TODO.md`, `ROADMAP.md`, `CHANGELOG.md`, and similar
free-form Markdown files.

### Interface

```bash
tl import markdown TODO.md
tl import markdown ROADMAP.md --section-to-tag
tl import markdown tasks.md --format gfm-tasks
tl import markdown todo.txt --priority-regex '\(P(\d)\)'
```

### Supported formats

#### GFM task lists

```markdown
## Backend

- [ ] Add login validation
- [x] Refactor auth middleware
- [/] Write API docs

## Frontend

- [ ] Design error page
  - [ ] Mockup wireframe
  - [ ] Implement component
```

Behaviour:
- `- [ ]` → status `open`
- `- [x]` → status `done`
- `- [/]` or `- [~]` → status `in_progress`
- Section headers (`## Backend`) → tags (with `--section-to-tag`) or ignored
- Nested checklists → `depends_on` parent task

#### Plain Markdown lists

```markdown
# Project tasks

Fix login bug (P1) @alice [HIGH]
Refactor auth module
Document API endpoints [DONE]
```

Behaviour:
- `- item` → title, status `open`
- Priority markers configurable via `--priority-regex`
- Assignee markers (`@name`) → assignee field
- Status markers (`[DONE]`, `[BLOCKED]`) → mapped status

### Open questions

- **Q4:** How deeply should nested lists be parsed? Flat only, or
    recursive with `depends_on`?
- **Q5:** Should we support YAML frontmatter in the source markdown file
    (e.g., a Jekyll-style blog post that also has tasks)?
- **Q6:** Should non-task content (prose between list items) be preserved
    as a preamble or skipped?

---

## 5. Import: GitHub Issues

One-shot import of issues from a GitHub repository.

### Interface

```bash
tl import github-issues --repo owner/repo
tl import github-issues --repo owner/repo --state open
tl import github-issues --repo owner/repo --label bug --label security
tl import github-issues --repo owner/repo --include-closed
tl import github-issues --repo owner/repo --include-prs
tl import github-issues --repo owner/repo --milestone "v1.0"
tl import github-issues --repo owner/repo --dry-run
```

### Auth

- Reads `GITHUB_TOKEN` from environment (or `GH_TOKEN`).
- Optional `--token` flag overrides.
- Unauthenticated: 60 requests/hour (not enough for any real repo). Token
  recommended.

### Mapping

| GitHub field | tl mapping | Notes |
|---|---|---|
| `title` | `title` | Direct |
| `body` | Description in `## Description` | |
| `state` (open/closed) | `open` / `done` | Configurable via `--status-map` |
| `labels` | `tags` | Label name only (color is lost) |
| `assignees` | `assignee` | First assignee only |
| `milestone` | Tag or ignored | `--milestone-as-tag` flag |
| `comments` | Notes in `## Notes` | Under `external-actor:` prefix |
| `number` | Stored in body preamble | `Imported from #42` |
| `pull_request` | Skipped (not an issue) | `--include-prs` to override |
| `created_at` / `updated_at` | Direct | |

### Open questions

- **Q7:** How to handle issue cross-references (e.g., "see #42" in body)?
    Heuristic regex replacement in description?
- **Q8:** Should milestone names become tags or separate metadata?
- **Q9:** Rate-limit handling — pause and retry, or abort with partial
    results?

---

## 6. Import: JIRA, Linear, Trello (one-shot)

These are lower-priority import sources. They follow the same pattern as
GitHub Issues — auth, fetch, map, write — but each has source-specific
complexities.

### JIRA

```bash
tl import jira --host mycompany.atlassian.net --project PROJ
tl import jira --host mycompany.atlassian.net --jql "project = PROJ AND status != Done"
tl import jira --host h --project P --status-map "To Do=open,In Progress=in_progress,Done=done"
```

Key challenges:
- Custom workflows mean arbitrary status names — `--status-map` is
  effectively required.
- Custom fields (every JIRA instance is different) — skipped in v1.
- Epics → parent task with `depends_on`.
- Issue links → `depends_on` with configurable link-type filter.
- Auth: email + API token (cloud) or PAT (server/DC).

### Linear

```bash
tl import linear --team engineering
tl import linear --team engineering --state-map "triage=open,backlog=open,todo=open,inProgress=in_progress,done=done,canceled=cancelled"
```

Key challenges:
- Linear priority (0–4) maps well to tl's (0 → none, 1–2 → high, 3 → medium, 4 → low).
- Teams → tags.
- Cycles → tags or ignored.
- Parent issues → `depends_on`.

### Trello (one-shot, no sync)

```bash
tl import trello --board BOARD_ID
tl import trello --board BOARD_ID --list-map "To Do=open,Doing=in_progress,Done=done"
```

Key challenges:
- Lists are user-defined — `--list-map` is required.
- Labels (color + name) → tags.
- Checklists → either merged into description or `--checklists-as-tasks`.
- Due dates → stored in description or as a pending-like field.

### Open questions

- **Q10:** Should the JSON intermediate format be the recommended path for
    unsupported sources (export → transform → `tl import json`) rather
    than building every adapter?
- **Q11:** For JIRA, how do we handle sprint/version/fixVersion mappings?

---

## 7. Sync: Trello bidirectional

The sync feature adds ongoing alignment between tl and an external kanban
board, initially Trello. Unlike import (one-shot), sync is stateful —
it tracks which tl task maps to which remote item and when they were last
aligned.

### Interface

```bash
# ── Setup (one-time) ──────────────────────────────────────────
tl sync init trello \
  --board "https://trello.com/b/abc123/my-board" \
  --list-map "To Do=open,In Progress=in_progress,Done=done" \
  --strategy warn
# Creates .tl/sync.yaml

# ── Bootstrap ──────────────────────────────────────────────────
tl sync --pull
# Imports all Trello cards as shallow tl tasks (with external field)

# ── Regular sync ───────────────────────────────────────────────
tl sync                 # bidirectional (push + pull)
tl sync --pull          # one-way: Trello → tl
tl sync --push          # one-way: tl → Trello
tl sync --json          # machine-readable report

# ── Status ─────────────────────────────────────────────────────
tl sync status          # show sync state for all synced tasks
```

### Data model: `external` field

Every synced task carries an `external` field in its YAML frontmatter:

```yaml
id: task-x3n
title: Refactor auth module
status: open

external:
  source: trello
  id: 64a2f8c1b3e7d90012345678
  url: https://trello.com/c/abc123/refactor-auth
  last_sync: 2026-05-29T12:00:00Z
```

Go struct addition:

```go
type External struct {
    Source   string     `yaml:"source" json:"source"`
    ID       string     `yaml:"id" json:"id"`
    URL      string     `yaml:"url,omitempty" json:"url,omitempty"`
    LastSync *time.Time `yaml:"last_sync,omitempty" json:"last_sync,omitempty"`
}

// Added to Task:
External *External `yaml:"external,omitempty" json:"external,omitempty"`
```

### Config: `.tl/sync.yaml`

```yaml
sources:
  - source: trello
    board_id: abc123
    board_url: https://trello.com/b/abc123/my-board
    list_map:
      "To Do": open
      "In Progress": in_progress
      "Done": done
    strategy: warn
```

This file is checked into git (part of the ledger).

### Field sync matrix

| tl field | Sync? | Direction | Detail |
|---|---|---|---|
| **title** | ✅ | bidirectional | Both sides can edit |
| **status** | ✅ | bidirectional | tl status ↔ Trello list (via `list_map`) |
| **assignee** | ✅ | bidirectional | tl assignee ↔ Trello member |
| **tags** | ✅ | bidirectional | tl tag names ↔ Trello label names |
| **description** | ✅ | bidirectional | tl body `## Description` ↔ Trello card description |
| **notes** | ⬅️ | Trello → tl only | Trello comments become tl notes. Reverse is too noisy |
| **claim** | ⬆️ | tl → Trello only | Claim → move card to mapped "in progress" list. Release → move to "open" list |
| **depends_on** | ❌ | — | No Trello equivalent |
| **priority** | ❌ | — | No Trello equivalent |

### Sync engine

The engine operates in three phases:

**Phase 1 — Match:**
1. Load all tl tasks with `external.source == trello`.
2. Fetch all cards from the Trello board (via API).
3. Pair by `external.id`. Three categories:
   - **Matched:** task ↔ card both exist.
   - **Orphaned tl:** task has `external.id` but card not found in Trello
     (deleted?).
   - **Orphaned remote:** card exists in Trello but no tl task has its ID
     (new card?).

**Phase 2 — Diff:**

For each matched pair, compare field-by-field since `last_sync`. Track
three states:
- **Local change only** — push to Trello.
- **Remote change only** — pull to tl.
- **Both changed** — conflict (strategy decides resolution).

**Phase 3 — Apply:**

Apply changes according to `strategy`:
- `warn` (default): auto-apply non-conflicting changes, report conflicts
  to user. Conflicts require manual resolution.
- `tl-wins`: always apply tl state to Trello on conflict.
- `trello-wins`: always apply Trello state to tl on conflict.

### Conflict example

```yaml
# Last sync: both were status=open (To Do)

# Now:
#   tl:  task-x3n status=in_progress  (agent claimed it)
#   trello: card moved to "Done"      (human moved it)

# Strategy=warn:
#   tl sync reports:
#     CONFLICT task-x3n: status (tl=in_progress, trello=done)
#     Resolve with: tl sync resolve task-x3n --use tl|trello

# Strategy=tl-wins:
#   Card moves back to "In Progress"

# Strategy=trello-wins:
#   Task status changes to done, claim auto-released
```

### Create with sync

```bash
tl create "Refactor auth" --sync trello
```

1. Generate tl task ID.
2. Create local task file with an `external` struct (ID and URL pending).
3. Call Trello API to create the card in the mapped "open" list.
4. Update task file with the Trello card ID and URL.
5. Record `created` and `sync_pushed` events.

**If Trello is unreachable** — the command fails. No task is created.
Rationale: avoid split-brain where a tl task exists but has no remote
counterpart. A future `--allow-offline` flag could relax this.

### Delete edge cases

| Event | Behaviour |
|---|---|
| Card deleted in Trello | `tl sync --pull` reports orphaned tl task. Task stays in tl. User resolves manually |
| Task deleted in tl | `tl sync --push` skips it (task gone). User must delete card manually in Trello |
| `tl close` synced task | Next `--push` sends status → card moves to "Done" list |
| `tl block` synced task | Status → `blocked`. No Trello list for this. Next `--push` warns: "no mapping for status blocked" |
| Claim expires on synced task | `tl stale` includes it as expected. Next `--push` could move card back to "To Do" (open question) |

### New events

| Event | When |
|---|---|
| `sync_pulled` | Remote changes applied locally |
| `sync_pushed` | Local changes applied remotely |
| `sync_conflict` | Conflict detected, user resolution needed |

### Open questions

- **Q12:** Should `tl sync` support `--auto` — a flag that pushes
    automatically after mutating commands (like `--sync` on `tl claim`,
    `tl close`, `tl release`)? This would make `tl claim task-x3n --sync`
    equivalent to `tl claim task-x3n && tl sync --push`.
- **Q13:** How should we handle Trello label color mapping? Trello labels
    have both a name and a color. tl tags have only a name. Should the
    color be encoded (e.g., `label-name:green` → `label-name` tag with
    color lost)?
- **Q14:** What happens when a synced task's status doesn't have a list
    mapping (e.g., `blocked` has no Trello equivalent)?
- **Q15:** Should `--allow-offline` exist for `tl create --sync trello`?
- **Q16:** Is one `external` per task enough, or should a task be
    syncable to multiple sources?
- **Q17:** How should Trello checklist items be handled? As shallow
    subtasks (depends_on)? As markdown in the description?
- **Q18:** Should sync support `TL_SYNC=1` as an environment variable to
    auto-push after mutating commands, without needing `--sync` on every
    individual command?

---

## 8. Data model changes

Summary of all changes to the Task struct:

```go
// New struct
type External struct {
    Source   string     `yaml:"source" json:"source"`
    ID       string     `yaml:"id" json:"id"`
    URL      string     `yaml:"url,omitempty" json:"url,omitempty"`
    LastSync *time.Time `yaml:"last_sync,omitempty" json:"last_sync,omitempty"`
}

// Addition to Task
External *External `yaml:"external,omitempty" json:"external,omitempty"`
```

New files:

| File | Purpose |
|---|---|
| `.tl/sync.yaml` | Sync source configuration |
| `internal/sync/` | Sync engine package |
| `internal/sync/sync.go` | Engine: match, diff, apply |
| `internal/sync/trello.go` | Trello adapter (API + mapping) |
| `cmd/sync.go` | `tl sync` command |
| `cmd/sync_init.go` | `tl sync init` subcommand |
| `cmd/import.go` | `tl import` root command |
| `cmd/import_json.go` | `tl import json` subcommand |
| `cmd/import_markdown.go` | `tl import markdown` subcommand |
| `internal/import/` | Import engine package |
| `features/import-json.feature` | BDD spec |
| `features/import-markdown.feature` | BDD spec |
| `features/sync-trello.feature` | BDD spec |

Modified files:

| File | Change |
|---|---|
| `internal/task/task.go` | Add `External` struct + field |
| `cmd/create.go` | Add `--sync` flag |
| `cmd/refine.go` | May need `--external-id`, `--external-url` (unlikely) |

---

## 9. Open questions

### Import (general)

- **Q1:** Should `external_id` be persisted in the task when importing?
  (Currently: no. Import is stateless. Re-importing duplicates.)
- **Q2:** Should the intermediate format support `depends_on` by batch
  index (position in the array) as well as by `external_id`?
- **Q3:** What is the max safe input size for stdin?
- **Q10:** Should the JSON intermediate format be the recommended path
  for unsupported sources rather than building every adapter?
- **Q11:** For JIRA, how do we handle sprint/version/fixVersion
  mappings?

### Import (markdown)

- **Q4:** How deeply should nested lists be parsed?
- **Q5:** Should we support YAML frontmatter in the source markdown file?
- **Q6:** Should non-task content (prose between list items) be preserved
  as a preamble or skipped?

### Import (GitHub Issues)

- **Q7:** How to handle issue cross-references (e.g., "see #42")?
- **Q8:** Should milestone names become tags or separate metadata?
- **Q9:** Rate-limit handling — pause and retry, or abort with partial
  results?

### Sync (Trello)

- **Q12:** Should `tl sync` support `--auto` (auto-push after mutating
  commands)?
- **Q13:** How should Trello label color mapping work?
- **Q14:** What happens when a synced task's status doesn't have a list
  mapping (e.g., `blocked`)?
- **Q15:** Should `--allow-offline` exist for `tl create --sync trello`?
- **Q16:** Is one `external` per task enough, or should a task be
  syncable to multiple sources?
- **Q17:** How should Trello checklist items be handled?
- **Q18:** Should sync support `TL_SYNC=1` environment variable for
  implicit auto-push?

### Process

- **Q19:** Should the `external` field be the same as the proposed
  `references` field (from the traceability analysis), or are they
  separate concepts? (`external` points to a remote card; `references`
  points to local files like Gherkin specs — they serve different
  purposes.)
- **Q20:** What is the delivery order? Option: `tl import json` →
  `tl import markdown` → references field → sync infrastructure →
  Trello adapter. Each is a separate task with its own Gherkin.

---

## Appendix A: Comparison of sync vs import

| Dimension | Import | Sync |
|---|---|---|
| **Use case** | Migrate existing work | Ongoing alignment |
| **Frequency** | One-time (or occasional re-import) | Regular (on demand) |
| **Direction** | Source → tl | Bidirectional |
| **Idempotent** | No | Yes (via `external.id`) |
| **Stateful** | No | Yes (`.tl/sync.yaml` + `external` field) |
| **Network** | Yes (for API sources) | Yes |
| **Error model** | Skip and report | Strategy-based conflict resolution |
| **Dependencies** | Aligned | Resolved |

## Appendix B: Phased delivery proposal

```
Phase 1 ── tl import json          │ pipe engine (foundation)
          tl import markdown        │ file-only, no deps
                                    │
Phase 2 ── references field         │ traceability (PRD: import-sync-PRD.md)
                                    │
Phase 3 ── tl sync infrastructure   │ match, diff, apply engine
          external field on Task    │ data model
          .tl/sync.yaml             │ config
          tl sync init              │ setup command
                                    │
Phase 4 ── Trello adapter           │ API client, card↔task mapping
          tl sync (trello)          │ bidirectional sync
          tl create --sync trello   │ create with sync
                                    │
Phase 5 ── tl import github-issues  │ API-based import
                                    │
Phase 6 ── tl import jira           │ complex mapping
          tl import linear          │ similar complexity
```
