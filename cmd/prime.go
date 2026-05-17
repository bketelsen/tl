package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const primeSnippet = `## TaskLedger Workflow

This repository uses TaskLedger (tl) for local task coordination between humans and agents.

When starting work:

1. Run tl ready --json to find tasks that are open, unblocked, and unclaimed.
2. Claim one task before editing files:
   tl claim <task-id> --actor <your-agent-name>
3. Inspect the task details:
   tl show <task-id>
4. Do the work.
5. Record important context, decisions, blockers, or handoff notes:
   tl note <task-id> --actor <your-agent-name> -m "..."
6. When the task is complete, close it:
   tl close <task-id> --actor <your-agent-name>

Rules:

- Do not work on a task claimed by another active actor unless explicitly told.
- Prefer tasks from tl ready; blocked, pending, done, cancelled, or actively claimed tasks are not ready.
- Leave notes for partial progress, failed approaches, decisions, and handoffs.
- Do not edit .taskledger/events.jsonl manually.
- Set TL_ACTOR when possible so commands can resolve your identity consistently.
- Ask before editing AGENTS.md or other project instruction files.
- If .taskledger/ is missing, ask the human whether to run tl init.
`

func newPrimeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "prime",
		Short: "Print recommended AGENTS.md instructions",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprint(cmd.OutOrStdout(), primeSnippet)
			return err
		},
	}
}
