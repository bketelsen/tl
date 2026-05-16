@implemented
Feature: List tasks in the ledger
  As a developer or agent
  I want to see every task in the ledger
  So that I can understand the full scope of recorded work

  Background:
    Given an initialized TaskLedger repository

  Scenario: Listing tasks shows identifier, status, and title for every task
    Given the following tasks exist:
      | id          | status      | title                        |
      | task-abc123 | open        | Add login form validation    |
      | task-def456 | in_progress | Refactor auth error messages |
      | task-ghi789 | done        | Document login flow          |
    When the developer runs `tl list`
    Then the output lists "task-abc123" with status "open" and title "Add login form validation"
    And the output lists "task-def456" with status "in_progress" and title "Refactor auth error messages"
    And the output lists "task-ghi789" with status "done" and title "Document login flow"

  Scenario: Listing tasks with JSON output returns the full array
    Given the following tasks exist:
      | id          | status | title                     |
      | task-abc123 | open   | Add login form validation |
      | task-def456 | open   | Document login flow       |
    When the developer runs `tl list --json`
    Then the JSON output is an array of 2 tasks
    And the JSON output contains a task with identifier "task-abc123"
    And the JSON output contains a task with identifier "task-def456"

  Scenario: Listing tasks orders tasks by priority and identifier
    Given the following tasks exist:
      | id          | status | priority | title                        |
      | task-ccc333 | open   | low      | Document login flow          |
      | task-bbb222 | open   | high     | Add login form validation    |
      | task-aaa111 | open   | high     | Refactor auth error messages |
      | task-ddd444 | open   | medium   | Review login copy            |
    When the developer runs `tl list`
    Then the listed task identifiers appear in this order:
      | id          |
      | task-aaa111 |
      | task-bbb222 |
      | task-ddd444 |
      | task-ccc333 |
