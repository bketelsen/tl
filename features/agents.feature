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

  Scenario: Running agents includes the current coordination workflow
    When the developer runs `tl agents`
    Then the output contains these snippets:
      | snippet                                                                   |
      | Set `TL_ACTOR` once at the start of your session                            |
      | Treat the task ledger as the source of truth                                      |
      | Do not begin implementation from chat instructions alone                     |
      | `tl ready --tag <role> --json`                                              |
      | `tl create "<title>" -d "<description>"`                                  |
      | `tl history <task-id>`                                                       |
      | `tl claim <task-id>` before making code, doc, config, or test changes        |
      | Re-run `tl claim <task-id>` periodically on long work                        |
      | `tl cancel <task-id> -m "<reason>"`                                         |
      | `tl block <task-id> -m "<blocker>"`                                         |
      | `tl pending <task-id> --question "..."`                                     |
      | `tl release <task-id>`                                                       |
      | check the current `@implemented` set with `make bdd`                         |
      | create it with `tl create` instead of silently expanding scope               |

  Scenario: Running agents does not modify any existing AGENTS.md
    Given the file "AGENTS.md" exists with content "# My Project"
    When the developer runs `tl agents`
    Then the file "AGENTS.md" still has content "# My Project"

  Scenario: Running agents with update appends to existing agent instruction files
    Given the file "AGENTS.md" exists with content "# My Project"
    And the file "CLAUDE.md" exists with content "# Claude Notes"
    When the developer runs `tl agents --update`
    Then the file "AGENTS.md" contains "<!-- BEGIN TL WORKFLOW -->"
    And the file "AGENTS.md" contains "## tl workflow"
    And the file "CLAUDE.md" contains "## tl workflow"
    And the output contains "Updated AGENTS.md"
    And the output contains "Updated CLAUDE.md"

  Scenario: Running agents with update refreshes an existing managed block
    Given the file "AGENTS.md" exists with content:
      """
      # My Project

      <!-- BEGIN TL WORKFLOW -->
      old workflow text
      <!-- END TL WORKFLOW -->
      """
    When the developer runs `tl agents --update`
    Then the file "AGENTS.md" contains "`tl ready --tag <role> --json`"
    And the file "AGENTS.md" does not contain "old workflow text"

  Scenario: Running agents with update does not create missing instruction files
    When the developer runs `tl agents --update`
    Then the file "AGENTS.md" does not exist
    And the file "CLAUDE.md" does not exist
    And the file "GEMINI.md" does not exist
    And the output contains "No existing agent instruction files found"
