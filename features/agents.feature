@implemented
Feature: Print recommended agent instructions
  As a developer setting up an agent-friendly repository
  I want to see or install the recommended AGENTS.md snippet
  So that I can give agents consistent tl workflow instructions

  Background:
    Given an initialized task ledger repository

  Scenario: Running agents prints the recommended AGENTS.md snippet to stdout
    When the developer runs `tl agents`
    Then the output contains a "tl workflow" heading
    And the output describes the ready, claim, show, note, and close steps
    And the output formats task commands as Markdown code spans


  Scenario: Running agents does not modify any existing AGENTS.md
    Given the file "AGENTS.md" exists with content "# My Project"
    When the developer runs `tl agents`
    Then the file "AGENTS.md" still has content "# My Project"

  Scenario: Running agents with write-files appends to existing agent instruction files
    Given the file "AGENTS.md" exists with content "# My Project"
    And the file "CLAUDE.md" exists with content "# Claude Notes"
    When the developer runs `tl agents --write-files`
    Then the file "AGENTS.md" contains "<!-- BEGIN TL WORKFLOW -->"
    And the file "AGENTS.md" contains "## tl workflow"
    And the file "CLAUDE.md" contains "## tl workflow"
    And the output contains "Updated AGENTS.md"
    And the output contains "Updated CLAUDE.md"

  Scenario: Running agents with write-files refreshes an existing managed block
    Given the file "AGENTS.md" exists with content:
      """
      # My Project

      <!-- BEGIN TL WORKFLOW -->
      old workflow text
      <!-- END TL WORKFLOW -->
      """
    When the developer runs `tl agents --write-files`
    Then the file "AGENTS.md" contains "`tl ready --tag <role> --json`"
    And the file "AGENTS.md" does not contain "old workflow text"

  Scenario: Running agents with write-files does not create missing instruction files
    When the developer runs `tl agents --write-files`
    Then the file "AGENTS.md" does not exist
    And the file "CLAUDE.md" does not exist
    And the file "GEMINI.md" does not exist
    And the output contains "No existing agent instruction files found"

  Scenario: Running agents with dry-run reports existing and skipped files without modifying them
    Given the file "AGENTS.md" exists with content:
      """
      # My Project

      <!-- BEGIN TL WORKFLOW -->
      old workflow text
      <!-- END TL WORKFLOW -->
      """
    And the file "CLAUDE.md" exists with content "# Claude Notes"
    When the developer runs `tl agents --write-files --dry-run`
    Then the output contains "Would update AGENTS.md (managed block found)"
    And the output contains "Would update CLAUDE.md (no managed block yet, would append)"
    And the output contains "Would skip GEMINI.md (file not found)"
    And the file "AGENTS.md" contains "old workflow text"
    And the file "CLAUDE.md" still has content "# Claude Notes"

  Scenario: Running agents with dry-run and no instruction files reports skips
    When the developer runs `tl agents --write-files --dry-run`
    Then the output contains "Would skip AGENTS.md (file not found)"
    And the output contains "Would skip CLAUDE.md (file not found)"
    And the output contains "Would skip GEMINI.md (file not found)"
    And the file "AGENTS.md" does not exist
    And the file "CLAUDE.md" does not exist
    And the file "GEMINI.md" does not exist

  Scenario: Running agents with dry-run but without write-files is rejected
    When the developer runs `tl agents --dry-run`
    Then the command exits with code 2
    And the output contains "--dry-run requires --write-files"

  Scenario: Running agents with deprecated update alias still writes files
    Given the file "AGENTS.md" exists with content "# My Project"
    When the developer runs `tl agents --update`
    Then the file "AGENTS.md" contains "<!-- BEGIN TL WORKFLOW -->"
    And the file "AGENTS.md" contains "## tl workflow"
    And the output contains "Updated AGENTS.md"
