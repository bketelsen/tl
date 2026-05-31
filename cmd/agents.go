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
	agentsBeginMarker       = "<!-- BEGIN TL WORKFLOW -->"
	agentsEndMarker         = "<!-- END TL WORKFLOW -->"
	legacyAgentsBeginMarker = "<!-- BEGIN " + "TASK" + "LEDGER WORKFLOW -->"
	legacyAgentsEndMarker   = "<!-- END " + "TASK" + "LEDGER WORKFLOW -->"
)

var agentInstructionFiles = []string{"AGENTS.md", "CLAUDE.md", "GEMINI.md"}

func newAgentsCmd() *cobra.Command {
	var writeFiles bool
	c := &cobra.Command{
		Use:   "agents",
		Short: "Print recommended AGENTS.md instructions",
		RunE: func(cmd *cobra.Command, args []string) error {
			if writeFiles {
				return updateAgentInstructionFiles(cmd)
			}
			_, err := fmt.Fprint(cmd.OutOrStdout(), agentsSnippet)
			return err
		},
	}
	c.Flags().BoolVar(&writeFiles, "write-files", false, "Write or refresh the tl workflow block in existing agent instruction files")
	c.Flags().BoolVar(&writeFiles, "update", false, "(deprecated: use --write-files)")
	_ = c.Flags().MarkHidden("update")
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
	for _, markers := range [][2]string{
		{agentsBeginMarker, agentsEndMarker},
		{legacyAgentsBeginMarker, legacyAgentsEndMarker},
	} {
		start := strings.Index(content, markers[0])
		if start >= 0 {
			end := strings.Index(content[start:], markers[1])
			if end >= 0 {
				end += start + len(markers[1])
				if strings.HasPrefix(content[end:], "\r\n") {
					end += len("\r\n")
				} else if strings.HasPrefix(content[end:], "\n") {
					end++
				}
				return content[:start] + block + content[end:]
			}
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
