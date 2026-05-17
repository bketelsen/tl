@implemented
Feature: Remove a dependency between tasks
  As a developer or agent
  I want to drop a dependency that no longer applies
  So that the ready queue reflects updated ordering

  Background:
    Given an initialized TaskLedger repository

  Scenario: Removing an existing dependency unlinks the two tasks
    Given a task "task-abc123" with no dependencies
    And a task "task-def456" exists
    And "task-abc123" depends on "task-def456"
    When the developer runs `tl dep remove task-abc123 --on task-def456`
    Then "task-abc123" does not depend on "task-def456"
    And an event "dependency_removed" is recorded for "task-abc123"
