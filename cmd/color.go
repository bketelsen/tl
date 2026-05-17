package cmd

import (
	"github.com/spf13/cobra"

	internalcolor "github.com/aholbreich/taskledger/internal/color"
)

func commandColorEnabled(cmd *cobra.Command) bool {
	mode := internalcolor.ModeAuto
	if flag := cmd.Root().Flag("color"); flag != nil {
		mode = flag.Value.String()
	}
	return internalcolor.Enabled(mode, cmd.OutOrStdout())
}

func colorStatus(enabled bool, status string) string {
	if !enabled {
		return status
	}
	return internalcolor.Status(status)
}

func colorPriority(enabled bool, priority string) string {
	if !enabled {
		return priority
	}
	return internalcolor.Priority(priority)
}

func colorID(enabled bool, id string) string {
	if !enabled {
		return id
	}
	return internalcolor.ID(id)
}

func colorClosedListLine(enabled bool, line string) string {
	if !enabled {
		return line
	}
	return internalcolor.ClosedListLine(line)
}
