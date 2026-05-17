package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed agents_snippet.md
var agentsSnippet string

const (
	agentsBeginMarker = "<!-- BEGIN TASKLEDGER WORKFLOW -->"
	agentsEndMarker   = "<!-- END TASKLEDGER WORKFLOW -->"
)

var agentInstructionFiles = []string{"AGENTS.md", "CLAUDE.md", "GEMINI.md"}

func newAgentsCmd() *cobra.Command {
	var update bool
	c := &cobra.Command{
		Use:   "agents",
		Short: "Print recommended AGENTS.md instructions",
		RunE: func(cmd *cobra.Command, args []string) error {
			if update {
				return updateAgentInstructionFiles(cmd)
			}
			_, err := fmt.Fprint(cmd.OutOrStdout(), agentsSnippet)
			return err
		},
	}
	c.Flags().BoolVar(&update, "update", false, "Append or refresh the TaskLedger workflow block in existing agent instruction files")
	return c
}

func updateAgentInstructionFiles(cmd *cobra.Command) error {
	updated := 0
	for _, path := range agentInstructionFiles {
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if info.IsDir() {
			return fmt.Errorf("%s is a directory", path)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(mergeAgentsBlock(string(data))), info.Mode()); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Updated %s\n", path)
		updated++
	}
	if updated == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No existing agent instruction files found")
	}
	return nil
}

func mergeAgentsBlock(content string) string {
	block := managedAgentsBlock()
	start := strings.Index(content, agentsBeginMarker)
	if start >= 0 {
		end := strings.Index(content[start:], agentsEndMarker)
		if end >= 0 {
			end += start + len(agentsEndMarker)
			if strings.HasPrefix(content[end:], "\r\n") {
				end += len("\r\n")
			} else if strings.HasPrefix(content[end:], "\n") {
				end++
			}
			return content[:start] + block + content[end:]
		}
	}

	if strings.TrimSpace(content) == "" {
		return block
	}
	if strings.HasSuffix(content, "\n\n") {
		return content + block
	}
	if strings.HasSuffix(content, "\n") {
		return content + "\n" + block
	}
	return content + "\n\n" + block
}

func managedAgentsBlock() string {
	return agentsBeginMarker + "\n" + strings.TrimRight(agentsSnippet, "\n") + "\n" + agentsEndMarker + "\n"
}
