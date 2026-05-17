package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/aholbreich/taskledger/internal/store"
	"github.com/aholbreich/taskledger/internal/task"
)

func newListCmd() *cobra.Command {
	var asJSON bool
	var includeAll bool
	var claimedBy string
	var status string
	var mine bool
	var tag string
	c := &cobra.Command{
		Use:   "list",
		Short: "List tasks in the ledger",
		RunE: func(cmd *cobra.Command, args []string) error {
			ledger, err := requireLedger()
			if err != nil {
				return err
			}
			tasks, err := store.List(ledger)
			if err != nil {
				return err
			}
			tasks = filterListTasks(tasks, includeAll, claimedBy, status, mine, tag)

			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(tasks)
			}

			var rendered bytes.Buffer
			tw := tabwriter.NewWriter(&rendered, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "ID\tStatus\tPriority\tClaimed By\tTitle")
			for _, t := range tasks {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", t.ID, t.Status, t.Priority, listClaimActor(t), t.Title)
			}
			if err := tw.Flush(); err != nil {
				return err
			}
			out := rendered.String()
			if includeAll && commandColorEnabled(cmd) {
				out = colorClosedListRows(out, tasks)
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), out)
			return err
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "Emit JSON output")
	c.Flags().BoolVarP(&includeAll, "all", "a", false, "Include closed tasks (done and cancelled)")
	c.Flags().StringVar(&claimedBy, "claimed-by", "", "Only show tasks claimed by this actor")
	c.Flags().StringVar(&status, "status", "", "Only show tasks with this status (overrides default closed hiding)")
	c.Flags().BoolVar(&mine, "mine", false, "Only show tasks claimed by the resolved actor")
	c.Flags().StringVar(&tag, "tag", "", "Only show tasks carrying this tag")
	return c
}

func filterListTasks(tasks []*task.Task, includeAll bool, claimedBy string, status string, mine bool, tag string) []*task.Task {
	if mine {
		resolved := ResolveActor("")
		claimedBy = resolved
	}

	filtered := tasks[:0]
	for _, t := range tasks {
		// --status overrides the default closed-task hiding.
		if status != "" {
			if t.Status != status {
				continue
			}
		} else if !includeAll && isClosedListStatus(t.Status) {
			continue
		}

		if claimedBy != "" && taskClaimActor(t) != claimedBy {
			continue
		}
		if tag != "" && !taskHasTag(t, tag) {
			continue
		}
		filtered = append(filtered, t)
	}
	return filtered
}

func isClosedListStatus(status string) bool {
	return status == "done" || status == "cancelled"
}

func colorClosedListRows(rendered string, tasks []*task.Task) string {
	lines := strings.Split(rendered, "\n")
	var b strings.Builder
	for i, line := range lines {
		if i > 0 && i-1 < len(tasks) && isClosedListStatus(tasks[i-1].Status) && line != "" {
			line = colorClosedListLine(true, line)
		}
		b.WriteString(line)
		if i < len(lines)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func listClaimActor(t *task.Task) string {
	actor := taskClaimActor(t)
	if actor == "" {
		return "-"
	}
	return actor
}

func taskClaimActor(t *task.Task) string {
	if t.Claim.Actor == nil {
		return ""
	}
	return *t.Claim.Actor
}

func taskHasTag(t *task.Task, tag string) bool {
	for _, tg := range t.Tags {
		if tg == tag {
			return true
		}
	}
	return false
}
