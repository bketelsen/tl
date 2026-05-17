## TaskLedger Workflow

This repository uses TaskLedger (`tl`) for local task coordination between humans and agents. Treat TaskLedger as the source of truth for all non-trivial work: planning, claiming, progress notes, blockers, handoffs, and completion.

Set `TL_ACTOR` once at the start of your session so you don't need `--actor` on each command:

```sh
export TL_ACTOR=claude-code:<purpose>
```

Mandatory workflow:

1. Start from TaskLedger.
   - Run `tl ready --json` to find unclaimed work, or `tl ready --tag <role> --json` to filter by role-ish tags.
   - If the human hands you a task, run `tl show <task-id>` and `tl history <task-id>` before editing.
   - Do not begin implementation from chat instructions alone if there is no matching TaskLedger task.
2. Ensure every work item has a task.
   - If no suitable task exists, create one with `tl create "<title>" -d "<description>"` before editing files.
   - If your work uncovers a separable follow-up, create it with `tl create` instead of silently expanding scope.
3. Claim before editing.
   - Run `tl claim <task-id>` before making code, doc, config, or test changes.
   - Do **not** work on a task claimed by another active actor unless explicitly told.
4. Re-check context after claiming.
   - Run `tl show <task-id>` to confirm scope, status, dependencies, and notes.
   - Run `tl history <task-id>` if there are prior events, stale claims, or handoff context.
5. Record progress in TaskLedger while working.
   - Re-run `tl claim <task-id>` periodically on long work — it extends the lease (heartbeat pattern).
   - Use `tl note <task-id> -m "..."` for meaningful progress, decisions, failed approaches, blockers, test results, and handoff context.
6. End every session with an explicit TaskLedger state.
   - `tl close <task-id>` — work is done and verified.
   - `tl cancel <task-id> -m "<reason>"` — work won't be done.
   - `tl block <task-id> -m "<blocker>"` — external blocker; claim is released.
   - `tl pending <task-id> --question "..."` — you need a human decision; claim is released.
   - `tl release <task-id>` — you're stepping away cleanly; leave a comprehensive note first.

   (`cancel`, `block`, `pending` are spec'd in `features/`; check the current `@implemented` set with `make bdd` before relying on them.)

Rules:

- Prefer tasks from `tl ready`; blocked, pending, done, cancelled, or actively claimed tasks are not ready.
- Use `--json` for automation and parsing (`tl ready --json`, `tl show <task-id> --json`, `tl history <task-id> --json`).
- Leave notes for partial progress, failed approaches, decisions, test results, blockers, and handoffs.
- Do **not** edit `.taskledger/events.jsonl` manually.
- If `.taskledger/` is missing, ask the human whether to run `tl init`.
