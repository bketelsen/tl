Feature: Run a task's verification commands
  As an agent or developer
  I want to run the verification commands defined for a task
  So that completion is proven rather than asserted

  Background:
    Given an initialized TaskLedger repository

  Scenario: Verification succeeds when all commands succeed
    Given a task "task-abc123" with two verification commands that both succeed
    When the developer runs `tl verify task-abc123`
    Then the command reports verification succeeded
    And the command exits with code 0

  Scenario: Verification stops on the first failing command
    Given a task "task-abc123" with two verification commands where the first fails
    When the developer runs `tl verify task-abc123`
    Then the command reports verification failed
    And the second verification command did not run
    And the command exits with code 6

  Scenario: Verification succeeds when the task has no verification commands
    Given a task "task-abc123" with no verification commands
    When the developer runs `tl verify task-abc123`
    Then the command reports verification succeeded
    And the command exits with code 0

  Scenario: Verification results can be retrieved as JSON
    Given a task "task-abc123" with a single verification command that fails
    When the developer runs `tl verify task-abc123 --json`
    Then the JSON output contains task identifier "task-abc123"
    And the JSON output reports overall success of false
    And the JSON output contains a failing command result
