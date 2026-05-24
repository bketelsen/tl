@implemented
Feature: Show a single task in detail
  As a developer or agent
  I want to inspect a specific task
  So that I can understand its full state before acting on it

  Background:
    Given an initialized task ledger repository

  Scenario: Showing a task prints its identifier, title, status, dependencies, claim, and notes
    Given a task "task-abc123" titled "Add login form validation" with status "open"
    And "task-abc123" depends on "task-def456"
    And "task-abc123" has a note from "claude-code:frontend" saying "Looked at the existing form."
    When the developer runs `tl show task-abc123`
    Then the output contains identifier "task-abc123"
    And the output contains title "Add login form validation"
    And the output contains status "open"
    And the output contains dependency "task-def456"
    And the output contains the note from "claude-code:frontend"

  Scenario: Showing a task with JSON output returns the full task object
    Given a task "task-abc123" titled "Add login form validation" with status "open"
    When the developer runs `tl show task-abc123 --json`
    Then the JSON output contains identifier "task-abc123"
    And the JSON output contains title "Add login form validation"
    And the JSON output contains status "open"

  Scenario: Showing a task with forced color highlights status and priority
    Given a task "task-abc123" titled "Fix production outage" with status "blocked" and priority "high"
    When the developer runs `tl --color=always show task-abc123`
    Then the output colorizes "blocked" with "red"
    And the output colorizes "high" with "red"

  Scenario: Showing a task with forced color styles labels, values, and Markdown headings
    Given a task "task-abc123" titled "Fix production outage" with status "open" and priority "medium"
    And "task-abc123" has a description "Investigate failing health checks."
    When the developer runs `tl --color=always show task-abc123`
    Then the output colorizes "ID" with "dim"
    And the output colorizes "task-abc123" with "bold"
    And the output colorizes "## Description" with "bright blue"

  Scenario: Showing a task with JSON output does not emit color
    Given a task "task-abc123" titled "Fix production outage" with status "blocked" and priority "high"
    When the developer runs `tl --color=always show task-abc123 --json`
    Then the output does not contain ANSI color
    And the JSON output contains status "blocked"

  Scenario: NO_COLOR disables forced color for show
    Given environment variable "NO_COLOR" is "1"
    And a task "task-abc123" titled "Fix production outage" with status "blocked" and priority "high"
    When the developer runs `tl --color=always show task-abc123`
    Then the output does not contain ANSI color
    And the output contains status "blocked"

  Scenario: Showing a task by bare short code works like the full identifier
    Given a task "task-abc123" titled "Add login form validation" with status "open"
    When the developer runs `tl show abc123`
    Then the output contains identifier "task-abc123"
    And the output contains title "Add login form validation"

  Scenario: Showing a task that does not exist is rejected
    Given no task with identifier "task-zzz999" exists
    When the developer runs `tl show task-zzz999`
    Then the command exits with code 3
    And the output reports that "task-zzz999" was not found
