@implemented
Feature: Add a dependency between tasks
  As a developer or agent
  I want to record that one task depends on another
  So that the ready queue reflects real ordering constraints

  Background:
    Given an initialized task ledger repository

  Scenario: Adding a dependency links one task to another
    Given a task "task-abc123" with no dependencies
    And a task "task-def456" exists
    When the developer runs `tl dep add task-abc123 --on task-def456`
    Then "task-abc123" depends on "task-def456"
    And an event "dependency_added" is recorded for "task-abc123"

  Scenario: Adding a dependency on a non-existent task is rejected
    Given a task "task-abc123" exists
    And no task with identifier "task-zzz999" exists
    When the developer runs `tl dep add task-abc123 --on task-zzz999`
    Then the command exits with code 3
    And "task-abc123" has no dependencies
