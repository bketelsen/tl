@implemented
Feature: Install shell completion in one step
  As a developer
  I want `tl completion --install` to write the completion script to my shell's canonical path
  So that TAB completion works in new terminals without copy-pasting setup snippets

  Scenario: tl completion bash emits a bash script to stdout
    When the developer runs `tl completion bash`
    Then the command exits with code 0
    And the output contains "# bash completion"

  Scenario: tl completion zsh emits a zsh script to stdout
    When the developer runs `tl completion zsh`
    Then the command exits with code 0
    And the output contains "#compdef tl"

  Scenario: tl completion fish emits a fish script to stdout
    When the developer runs `tl completion fish`
    Then the command exits with code 0
    And the output contains "# fish completion"

  Scenario: tl completion --install bash writes the script to the canonical bash path
    Given environment variable "HOME" is the scenario temp directory
    When the developer runs `tl completion --install bash`
    Then the command exits with code 0
    And the file ".local/share/bash-completion/completions/tl" exists in the scenario temp directory
    And the output contains "Installed bash completion to"

  Scenario: tl completion --install fish writes the script to the canonical fish path
    Given environment variable "HOME" is the scenario temp directory
    When the developer runs `tl completion --install fish`
    Then the command exits with code 0
    And the file ".config/fish/completions/tl.fish" exists in the scenario temp directory

  Scenario: tl completion --install zsh writes the script and prints fpath setup instructions
    Given environment variable "HOME" is the scenario temp directory
    When the developer runs `tl completion --install zsh`
    Then the command exits with code 0
    And the file ".zsh/completions/_tl" exists in the scenario temp directory
    And the output contains "fpath"

  Scenario: tl completion --install detects bash from the SHELL environment variable
    Given environment variable "HOME" is the scenario temp directory
    And environment variable "SHELL" is "/bin/bash"
    When the developer runs `tl completion --install`
    Then the command exits with code 0
    And the file ".local/share/bash-completion/completions/tl" exists in the scenario temp directory

  Scenario: tl completion --install rejects unsupported shells from SHELL with exit code 2
    Given environment variable "HOME" is the scenario temp directory
    And environment variable "SHELL" is "/bin/dash"
    When the developer runs `tl completion --install`
    Then the command exits with code 2
    And the output reports that the shell is unsupported

  Scenario: tl completion --install rejects an unknown explicit shell with exit code 2
    When the developer runs `tl completion --install xenu`
    Then the command exits with code 2
    And the output reports that the shell is unsupported

  Scenario: tl completion --install powershell prints append-to-profile instructions
    When the developer runs `tl completion --install powershell`
    Then the command exits with code 0
    And the output contains "$PROFILE"
