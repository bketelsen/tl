Feature: Diagnose ledger integrity
  As a developer or agent
  I want to scan the ledger for structural and data integrity issues
  So that problems like broken dependencies, malformed files, and stale state are found (and optionally fixed)

  Background:
    Given an initialized task ledger repository

  Scenario: A clean ledger reports zero issues
    Given a task "task-abc" with title "Working task" and status "open"
    When the developer runs `tl doctor`
    Then the doctor reports no issues
    And the doctor exits with code 0

  Scenario: Malformed YAML frontmatter is detected
    Given a task file "task-bad.md" with content that is not valid YAML
    When the developer runs `tl doctor`
    Then the doctor reports a "frontmatter" issue for "task-bad"
    And the doctor reports the issue severity as "error"

  Scenario: Missing required frontmatter fields are detected
    Given a task file "task-xyz.md" with frontmatter missing the "title" field
    When the developer runs `tl doctor`
    Then the doctor reports a "frontmatter" issue for "task-xyz"

  Scenario: Invalid status values are detected
    Given a task "task-abc" with status "super-duper"
    When the developer runs `tl doctor`
    Then the doctor reports a "frontmatter" issue for "task-abc" with message matching "status"

  Scenario: Invalid priority values are detected
    Given a task "task-abc" with priority "urgent"
    When the developer runs `tl doctor`
    Then the doctor reports a "frontmatter" issue for "task-abc" with message matching "priority"

  Scenario: Duplicate task IDs are detected
    Given a task "task-dup" exists
    And a second task file "task-dup.md" exists in the tasks directory
    When the developer runs `tl doctor`
    Then the doctor reports an "identity" issue for "task-dup"

  Scenario: Missing dependency targets are detected
    Given a task "task-abc" with dependency "task-nonexistent"
    When the developer runs `tl doctor`
    Then the doctor reports a "dependency" issue for "task-abc"

  Scenario: Self-referencing dependencies are detected
    Given a task "task-abc" with dependency "task-abc"
    When the developer runs `tl doctor`
    Then the doctor reports a "dependency" issue for "task-abc"

  Scenario: Cyclic dependencies are detected
    Given a task "task-a" with dependency "task-b"
    And a task "task-b" with dependency "task-a"
    When the developer runs `tl doctor`
    Then the doctor reports a "dependency" issue for "task-a"
    And the doctor reports a "dependency" issue for "task-b"

  Scenario: Orphaned events referencing nonexistent tasks are detected
    Given an event in the journal referencing task "task-ghost"
    When the developer runs `tl doctor`
    Then the doctor reports an "events" issue for "task-ghost"

  Scenario: Tasks with no events in the journal are detected
    Given a task "task-abc" exists
    And no events exist for "task-abc" in the journal
    When the developer runs `tl doctor`
    Then the doctor reports an "events" issue for "task-abc"

  Scenario: A task with status in_progress but no claim data is detected
    Given a task "task-abc" with status "in_progress" and no active claim
    When the developer runs `tl doctor`
    Then the doctor reports a "claims" issue for "task-abc"

  Scenario: An expired claim not released is detected
    Given a task "task-abc" claimed by "claude-code:main" with an expired lease
    When the developer runs `tl doctor`
    Then the doctor reports a "claims" issue for "task-abc"

  Scenario: An open task with stale claim data is detected
    Given a task "task-abc" with status "open" that still has claim data set
    When the developer runs `tl doctor`
    Then the doctor reports a "claims" issue for "task-abc"

  Scenario: Timestamps in the future are detected
    Given a task "task-abc" with created_at in the year 2099
    When the developer runs `tl doctor`
    Then the doctor reports a "timestamps" issue for "task-abc"

  Scenario: created_at after updated_at is detected
    Given a task "task-abc" with created_at after updated_at
    When the developer runs `tl doctor`
    Then the doctor reports a "timestamps" issue for "task-abc"

  Scenario: Claim expiry before claim time is detected
    Given a task "task-abc" with claim_expires_at before claim_claimed_at
    When the developer runs `tl doctor`
    Then the doctor reports a "timestamps" issue for "task-abc"

  Scenario: Orphaned .tmp files are detected
    Given an orphaned file "task-abc.md.tmp" exists in the tasks directory
    When the developer runs `tl doctor`
    Then the doctor reports a "filesystem" issue

  Scenario: Merge conflict markers in task body are detected
    Given a task "task-abc" whose body contains "<<<<<<< HEAD"
    When the developer runs `tl doctor`
    Then the doctor reports a "body" issue for "task-abc"

  Scenario: Merge conflict markers with ======= line are detected
    Given a task "task-abc" whose body contains "======="
    When the developer runs `tl doctor`
    Then the doctor reports a "body" issue for "task-abc"

  Scenario: Merge conflict markers with >>>>>>> line are detected
    Given a task "task-abc" whose body contains ">>>>>>> branch-name"
    When the developer runs `tl doctor`
    Then the doctor reports a "body" issue for "task-abc"

  Scenario: Scale warning is emitted for large ledgers
    Given a ledger with 101 tasks
    When the developer runs `tl doctor`
    Then the doctor output contains a scale warning about task count

  Scenario: Scale warning is emitted for large event journals
    Given a ledger with 1001 events
    When the developer runs `tl doctor`
    Then the doctor output contains a scale warning about event count

  Scenario: JSON output returns diagnostics as an array
    Given a task "task-bad" with status "invalid-status"
    When the developer runs `tl doctor --json`
    Then the JSON output is an array of diagnostic objects
    And a diagnostic object has severity "error" and category "frontmatter"
    And a diagnostic object has a non-empty "message" field

  Scenario: A diagnostic object includes a fixable flag
    Given a task "task-bad" with status "invalid-status"
    When the developer runs `tl doctor --json`
    Then a diagnostic object has a "fixable" boolean field

  Scenario: A clean ledger produces an empty JSON array
    Given a task "task-abc" with title "Working task" and status "open"
    When the developer runs `tl doctor --json`
    Then the JSON output is an empty array

  Scenario: --fix repairs a self-referencing dependency
    Given a task "task-abc" with dependency "task-abc"
    When the developer runs `tl doctor --fix`
    Then the doctor reports a "dependency" issue for "task-abc" as fixed
    And "task-abc" no longer depends on "task-abc"

  Scenario: --fix removes orphaned .tmp files
    Given an orphaned file "task-abc.md.tmp" exists in the tasks directory
    When the developer runs `tl doctor --fix`
    Then the doctor reports the orphaned file as removed
    And the file "task-abc.md.tmp" no longer exists

  Scenario: --fix clears stale claim data from an open task
    Given a task "task-abc" with status "open" that still has claim data set
    When the developer runs `tl doctor --fix`
    Then the doctor reports the claim data as cleared
    And "task-abc" has no claim data

  Scenario: --fix does not attempt to repair unfixable issues
    Given a task file "task-bad.md" with content that is not valid YAML
    When the developer runs `tl doctor --fix`
    Then the doctor reports the "frontmatter" issue for "task-bad" as not fixable

  Scenario: Config file missing is detected
    Given the ledger has no config.yaml file
    When the developer runs `tl doctor`
    Then the doctor reports a "config" issue

  Scenario: Config file with invalid YAML is detected
    Given the config.yaml contains content that is not valid YAML
    When the developer runs `tl doctor`
    Then the doctor reports a "config" issue

  Scenario: A task file that is entirely unreadable is detected
    Given a task file "task-corrupted.md" that cannot be read
    When the developer runs `tl doctor`
    Then the doctor reports a "filesystem" issue for "task-corrupted"

  Scenario: Malformed notes format in task body is detected
    Given a task "task-abc" whose Notes section contains lines that do not match the canonical format
    When the developer runs `tl doctor`
    Then the doctor reports a "body" issue for "task-abc"

  Scenario: doctor exits 0 even when issues are found
    Given a task "task-bad" with status "invalid-status"
    When the developer runs `tl doctor`
    Then the command exits with code 0
