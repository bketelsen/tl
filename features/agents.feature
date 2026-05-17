@implemented
Feature: Print recommended agent instructions
  As a developer setting up an agent-friendly repository
  I want to see or install the recommended AGENTS.md snippet
  So that I can give agents consistent TaskLedger workflow instructions

  Background:
    Given an initialized TaskLedger repository

  Scenario: Running agents prints the recommended AGENTS.md snippet to stdout
    When the developer runs `tl agents`
    Then the output contains a "TaskLedger Workflow" heading
    And the output describes the ready, claim, show, note, and close steps
    And the output formats task commands as Markdown code spans

  Scenario: Running agents includes the current coordination workflow
    When the developer runs `tl agents`
    Then the output contains these snippets:
      | snippet                                                                   |
      | Set `TL_ACTOR` once at the start of your session                            |
      | `tl ready --tag <role> --json`                                              |
      | `tl history <task-id>`                                                       |
      | Re-run `tl claim <task-id>` periodically on long work                        |
      | `tl cancel <task-id> -m "<reason>"`                                         |
      | `tl block <task-id> -m "<blocker>"`                                         |
      | `tl pending <task-id> --question "..."`                                     |
      | `tl release <task-id>`                                                       |
      | check the current `@implemented` set with `make bdd`                         |
      | create a follow-up task with `tl create` rather than silently expanding scope |

  Scenario: Running agents does not modify any existing AGENTS.md
    Given the file "AGENTS.md" exists with content "# My Project"
    When the developer runs `tl agents`
    Then the file "AGENTS.md" still has content "# My Project"

  Scenario: Running agents with update appends to existing agent instruction files
    Given the file "AGENTS.md" exists with content "# My Project"
    And the file "CLAUDE.md" exists with content "# Claude Notes"
    When the developer runs `tl agents --update`
    Then the file "AGENTS.md" contains "<!-- BEGIN TASKLEDGER WORKFLOW -->"
    And the file "AGENTS.md" contains "## TaskLedger Workflow"
    And the file "CLAUDE.md" contains "## TaskLedger Workflow"
    And the output contains "Updated AGENTS.md"
    And the output contains "Updated CLAUDE.md"

  Scenario: Running agents with update refreshes an existing managed block
    Given the file "AGENTS.md" exists with content:
      """
      # My Project

      <!-- BEGIN TASKLEDGER WORKFLOW -->
      old workflow text
      <!-- END TASKLEDGER WORKFLOW -->
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
