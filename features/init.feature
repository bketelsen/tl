@implemented
Feature: Initialize a TaskLedger repository
  As a developer adopting TaskLedger
  I want to set up the ledger inside my Git repository
  So that humans and agents can record and coordinate tasks locally

  Scenario: Initializing creates the ledger layout in a fresh repository
    Given the current directory has no TaskLedger ledger
    When the developer runs `tl init`
    Then the directory contains a TaskLedger config file
    And the directory contains an empty tasks folder
    And the directory contains an empty event journal

  Scenario: Initializing an already-initialized repository is rejected
    Given the current directory already has a TaskLedger ledger
    When the developer runs `tl init`
    Then the command reports that the ledger already exists
    And the existing config file is unchanged
