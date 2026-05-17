package cmd

import (
	"strings"
	"testing"
)

func TestMergeAgentsBlockAppendsToExistingContent(t *testing.T) {
	got := mergeAgentsBlock("# Project")
	for _, want := range []string{"# Project\n\n", agentsBeginMarker, "## TaskLedger Workflow", agentsEndMarker} {
		if !strings.Contains(got, want) {
			t.Fatalf("mergeAgentsBlock() missing %q; got:\n%s", want, got)
		}
	}
}

func TestMergeAgentsBlockReplacesManagedBlockIdempotently(t *testing.T) {
	input := "# Project\n\n" + agentsBeginMarker + "\nold workflow\n" + agentsEndMarker + "\n"

	once := mergeAgentsBlock(input)
	twice := mergeAgentsBlock(once)

	if once != twice {
		t.Fatalf("mergeAgentsBlock should be idempotent\nonce:\n%s\ntwice:\n%s", once, twice)
	}
	if strings.Contains(once, "old workflow") {
		t.Fatalf("mergeAgentsBlock did not replace old managed block; got:\n%s", once)
	}
}
