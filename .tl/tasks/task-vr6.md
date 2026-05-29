---
id: task-vr6
title: Add tl completion --install for one-step shell completion setup
status: done
priority: medium
type: feature
created_at: 2026-05-29T12:31:06Z
updated_at: 2026-05-29T13:36:36Z
created_by: human
assignee: null
depends_on: []
claim:
  actor: null
  claimed_at: null
  expires_at: null
  heartbeat_at: null
tags:
  - dx
  - completion
---

## Description

Currently users have to know to run `source <(tl completion bash)` (or the equivalent for zsh/fish) and either source it per-shell or install it persistently themselves. This is the same friction kubectl/helm/gh users face, but it's the #1 reason `tl show <TAB>` appears "broken" after a fresh install.

**Goal:** `tl completion --install` detects the user's shell and writes the script to the right place. `tl completion --install <shell>` forces a specific shell. `tl completion <shell>` (no flag) keeps emitting to stdout as today.

**Approach:** disable cobra's auto-generated `completion` command via `CompletionOptions.DisableDefaultCmd = true` on root, then add our own `completion` cobra.Command with bash/zsh/fish/powershell subcommands and a persistent `--install` flag. Detect shell from `os.Getenv("SHELL")` basename, fall back to asking. Install paths per shell:

- bash (Linux): `~/.local/share/bash-completion/completions/tl`
- bash (Homebrew): `$(brew --prefix)/etc/bash_completion.d/tl`
- zsh: `~/.zsh/completions/_tl` + print one-line fpath note
- fish: `~/.config/fish/completions/fish.tl`
- powershell: print append-to-profile snippet (no canonical path)

Print the absolute path written and a one-line 'open a new shell to activate' hint.

**Follow-up consideration:** also add a one-line mention in `tl init` output suggesting `tl completion --install` so new repos discover it.

**Test plan:** BDD scenarios for (a) bash install writes to expected path, (b) explicit shell arg overrides $SHELL detection, (c) bare `tl completion bash` still emits script, (d) unknown shell rejected with exit 2. Use a HOME override via env to test path writing without touching the real homedir.

## Notes

- 2026-05-29T13:36:36Z [claude-code] note: Implemented. Files: cmd/completion_cmd.go (custom completion command replacing cobra's auto-generated one); cmd/root.go (CompletionOptions.DisableDefaultCmd=true, AddCommand newCompletionCmd); cmd/init.go (one-line tip about tl completion --install after success); features/completion-install.feature (10 BDD scenarios); bdd/completion_install_test.go (3 new step defs: HOME=tempdir, file-exists-in-tempdir, unsupported-shell); features/init.feature (asserts the tip is printed); README.md (replaced manual source snippets with one-step tl completion --install). Install paths (XDG-aware): - bash: $XDG_DATA_HOME/bash-completion/completions/tl (default ~/.local/share/...) - zsh: $ZDOTDIR/completions/_tl (default ~/.zsh/...) + prints fpath instructions - fish: $XDG_CONFIG_HOME/fish/completions/tl.fish (default ~/.config/...) - powershell: not auto-installable; prints append-to-profile snippet, exits 0 Spec deviation: description said 'fall back to asking' if shell can't be detected; I exit 2 with a helpful message instead. Better for non-interactive use (CI, piped output) and matches the project's no-interactive-prompts pattern. Tests: 160/160 BDD scenarios pass; gofmt/vet clean. make install run so installed binary reflects new behavior. Smoke-verified all paths: explicit shell positional, SHELL detection, unknown shell rejection, powershell instructions, init tip line.
