package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// newCompletionCmd replaces cobra's auto-generated `completion` command with
// one that adds --install for one-step setup. Without --install, the script
// is emitted to stdout (matching cobra's default behaviour). With --install,
// the script is written to the canonical path for the chosen shell.
func newCompletionCmd() *cobra.Command {
	var install bool
	c := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate or install the shell completion script",
		Long: `Generate the shell completion script for tl.

Without --install, the script is written to stdout — source it in your
shell to enable completion in the current session.

With --install, the script is written to the canonical path for the chosen
shell so completion loads automatically in new shells.

Examples:
  tl completion bash                # print bash script to stdout
  source <(tl completion bash)      # enable completion in current bash
  tl completion --install           # auto-detect shell from $SHELL and install
  tl completion --install bash      # install bash completion explicitly`,
		Args:      cobra.MaximumNArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := ""
			if len(args) == 1 {
				shell = args[0]
			} else if install {
				shell = detectShellFromEnv(os.Getenv("SHELL"))
				if shell == "" {
					return NewExitError(2, "could not detect shell from $SHELL; pass a shell name explicitly: tl completion --install bash|zsh|fish|powershell")
				}
			} else {
				return cmd.Help()
			}

			if !isSupportedShell(shell) {
				return NewExitError(2, "unsupported shell %q (supported: bash, zsh, fish, powershell)", shell)
			}

			if install {
				return installCompletion(cmd, shell)
			}
			return writeCompletionScript(cmd.Root(), cmd.OutOrStdout(), shell)
		},
	}
	c.Flags().BoolVar(&install, "install", false, "Write the completion script to the canonical path for the chosen shell")
	return c
}

func isSupportedShell(shell string) bool {
	switch shell {
	case "bash", "zsh", "fish", "powershell":
		return true
	default:
		return false
	}
}

// detectShellFromEnv returns the supported-shell basename of the given SHELL
// value, or "" if it cannot be mapped. PowerShell is never auto-detected on
// Unix — users on Windows must pass it explicitly.
func detectShellFromEnv(shellPath string) string {
	base := filepath.Base(strings.TrimSpace(shellPath))
	switch base {
	case "bash", "zsh", "fish":
		return base
	default:
		return ""
	}
}

func writeCompletionScript(root *cobra.Command, w io.Writer, shell string) error {
	switch shell {
	case "bash":
		return root.GenBashCompletionV2(w, true)
	case "zsh":
		return root.GenZshCompletion(w)
	case "fish":
		return root.GenFishCompletion(w, true)
	case "powershell":
		return root.GenPowerShellCompletionWithDesc(w)
	}
	return fmt.Errorf("unsupported shell %q", shell)
}

func installCompletion(cmd *cobra.Command, shell string) error {
	if shell == "powershell" {
		fmt.Fprintln(cmd.OutOrStdout(), "PowerShell does not have a single canonical completion path.")
		fmt.Fprintln(cmd.OutOrStdout(), "Append the completion script to your profile:")
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintln(cmd.OutOrStdout(), "  tl completion powershell | Out-String | Invoke-Expression")
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintln(cmd.OutOrStdout(), "or persist it across sessions:")
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintln(cmd.OutOrStdout(), "  tl completion powershell >> $PROFILE")
		return nil
	}

	path, err := installPath(shell)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create completion directory: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("write completion script: %w", err)
	}
	defer f.Close()
	if err := writeCompletionScript(cmd.Root(), f, shell); err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Installed %s completion to %s\n", shell, path)
	switch shell {
	case "bash":
		fmt.Fprintln(out, "Open a new shell, or run `source "+path+"` to activate now.")
		fmt.Fprintln(out, "(Requires the bash-completion package; install via your distro's package manager if TAB still does nothing.)")
	case "zsh":
		dir := filepath.Dir(path)
		fmt.Fprintln(out, "Add the following to ~/.zshrc if it is not already there, then open a new shell:")
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "  fpath=("+dir+" $fpath)")
		fmt.Fprintln(out, "  autoload -Uz compinit && compinit")
	case "fish":
		fmt.Fprintln(out, "Completion loads automatically in new fish shells.")
	}
	return nil
}

// installPath returns the canonical user-local completion script path for a
// shell, honouring XDG_DATA_HOME / XDG_CONFIG_HOME / ZDOTDIR where defined.
func installPath(shell string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch shell {
	case "bash":
		dataHome := os.Getenv("XDG_DATA_HOME")
		if dataHome == "" {
			dataHome = filepath.Join(home, ".local", "share")
		}
		return filepath.Join(dataHome, "bash-completion", "completions", "tl"), nil
	case "zsh":
		zdotdir := os.Getenv("ZDOTDIR")
		if zdotdir == "" {
			zdotdir = filepath.Join(home, ".zsh")
		}
		return filepath.Join(zdotdir, "completions", "_tl"), nil
	case "fish":
		configHome := os.Getenv("XDG_CONFIG_HOME")
		if configHome == "" {
			configHome = filepath.Join(home, ".config")
		}
		return filepath.Join(configHome, "fish", "completions", "tl.fish"), nil
	}
	return "", fmt.Errorf("no canonical install path for %q", shell)
}
