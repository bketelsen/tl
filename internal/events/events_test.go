package events

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aholbreich/tl/internal/repo"
)

func TestAppendSeparatesFromJournalWithoutTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	ledger, err := repo.Init(dir)
	if err != nil {
		t.Fatalf("init ledger: %v", err)
	}
	journal := filepath.Join(ledger, repo.EventsJournal)
	first := `{"time":"2026-01-01T00:00:00Z","event":"created","task_id":"task-one"}`
	if err := os.WriteFile(journal, []byte(first), 0o644); err != nil {
		t.Fatalf("write journal: %v", err)
	}

	if err := Append(ledger, Event{Event: "created", TaskID: "task-two"}); err != nil {
		t.Fatalf("Append: %v", err)
	}

	data, err := os.ReadFile(journal)
	if err != nil {
		t.Fatalf("read journal: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected append to create a separate JSONL line, got %d lines:\n%s", len(lines), data)
	}
	if strings.Contains(string(data), "}{") {
		t.Fatalf("journal contains concatenated JSON objects after append:\n%s", data)
	}
}
