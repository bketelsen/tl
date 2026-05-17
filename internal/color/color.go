// Package color provides tiny ANSI helpers for human-readable CLI output.
package color

import (
	"fmt"
	"io"
	"os"
)

const (
	ModeAuto   = "auto"
	ModeNever  = "never"
	ModeAlways = "always"

	Reset      = "\x1b[0m"
	Bold       = "\x1b[1m"
	Dim        = "\x1b[2m"
	Red        = "\x1b[31m"
	Green      = "\x1b[32m"
	Yellow     = "\x1b[33m"
	Blue       = "\x1b[34m"
	BrightBlue = "\x1b[94m"
	Magenta    = "\x1b[35m"
	Cyan       = "\x1b[36m"
)

// ValidateMode checks the user-facing --color value.
func ValidateMode(mode string) error {
	switch mode {
	case ModeAuto, ModeNever, ModeAlways:
		return nil
	default:
		return fmt.Errorf("invalid --color %q: must be auto, never, or always", mode)
	}
}

// Enabled reports whether ANSI color should be emitted for this mode and output.
func Enabled(mode string, out io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	switch mode {
	case ModeAlways:
		return true
	case ModeNever:
		return false
	case ModeAuto:
		return isTerminal(out)
	default:
		return false
	}
}

func isTerminal(out io.Writer) bool {
	file, ok := out.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// Apply wraps text with an ANSI code and reset.
func Apply(code, text string) string {
	return code + text + Reset
}

func Status(status string) string {
	switch status {
	case "open":
		return Apply(Green, status)
	case "in_progress":
		return Apply(Yellow, status)
	case "blocked":
		return Apply(Red, status)
	case "pending_human":
		return Apply(Magenta, status)
	case "done":
		return Apply(Cyan, status)
	case "cancelled":
		return Apply(Dim, status)
	default:
		return status
	}
}

func Priority(priority string) string {
	switch priority {
	case "high":
		return Apply(Red, priority)
	case "medium":
		return Apply(Yellow, priority)
	case "low":
		return Apply(Dim, priority)
	default:
		return priority
	}
}

func ListPriority(priority string) string {
	switch priority {
	case "high":
		return Apply(Red, priority)
	case "medium":
		return Apply(Yellow, priority)
	case "low":
		return Apply(Blue, priority)
	default:
		return priority
	}
}

func ID(id string) string {
	return Apply(Cyan, id)
}

func FieldLabel(label string) string {
	return Apply(Dim, label)
}

func FieldValue(value string) string {
	return Apply(Bold, value)
}

func MarkdownHeading(heading string) string {
	return Apply(BrightBlue, heading)
}

func ClosedListLine(line string) string {
	return Apply(Dim, line)
}
