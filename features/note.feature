@implemented
Feature: Append a note to a task
  As an agent or developer
  I want to record notes and handoff context on a task
  So that other actors can see what was tried, learned, or decided

  Background:
    Given an initialized TaskLedger repository

  Scenario: An actor appends a note that records timestamp, actor, and message
    Given a task "task-abc123" titled "Add login form validation"
    When the agent runs `tl note task-abc123 --actor claude-code:frontend --message "Mock missing in auth tests."`
    Then "task-abc123" has a note from "claude-code:frontend"
    And the note contains the message "Mock missing in auth tests."
    And the note has a timestamp

  Scenario: Appending a note records a note_added event
    Given a task "task-abc123" exists
    When the agent runs `tl note task-abc123 --actor claude-code:frontend --message "Verified locally."`
    Then an event "note_added" is recorded for "task-abc123"

