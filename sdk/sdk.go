// Package sdk is the public, importable API of the tl task ledger.
//
// It re-exports the ledger's domain model and operations from tl's internal
// packages so that external Go modules (e.g. an agent orchestrator) can read
// and write the same .tl/ ledger the `tl` CLI manages — sharing one
// implementation of the on-disk format, the readiness rule, and the file lock,
// rather than maintaining a divergent copy.
//
// The on-disk ledger is plain, git-native files: one Markdown+YAML task per
// .tl/tasks/task-<id>.md, an append-only .tl/events.jsonl audit journal, and a
// .tl/config.yaml. There is no database, daemon, or remote service.
//
// Everything here is a thin re-export (type aliases + function values) of the
// internal packages, so an sdk.Task IS an internal task.Task — values pass
// between the SDK and the underlying store/events functions with no conversion.
//
// tl is a fork of github.com/aholbreich/tl (MIT, Copyright (c) 2026 Alexander
// Holbreich); see the repository LICENSE.
package sdk

import (
	"time"

	"github.com/bketelsen/tl/internal/events"
	"github.com/bketelsen/tl/internal/ready"
	"github.com/bketelsen/tl/internal/store"
	"github.com/bketelsen/tl/internal/task"
)

// --- Domain model (type aliases: an sdk.Task is exactly an internal task.Task) ---

// Task is a single ledger task: id, status, priority, dependencies, claim lease,
// tags, and a Markdown body. All fields are exported and may be constructed and
// mutated directly (see store.Write to persist).
type Task = task.Task

// Claim is a time-limited lease on a task (actor + claimed/expires/heartbeat).
type Claim = task.Claim

// Pending is a human-decision gate recorded on a task (question + requester).
type Pending = task.Pending

// Note is a single parsed entry from a task's "## Notes" section.
type Note = task.Note

// ParsedBody is the structured view of a task body: description, notes, sections.
type ParsedBody = task.ParsedBody

// Event is one line of the .tl/events.jsonl audit journal.
type Event = events.Event

// DepError describes an unsatisfied dependency (missing, or not yet done).
type DepError = ready.DepError

// --- Store: locate the ledger and read/write task files ---

// Store error sentinels.
var (
	ErrLedgerNotFound = store.ErrLedgerNotFound
	ErrTaskNotFound   = store.ErrTaskNotFound
)

// LedgerDir finds the nearest .tl ledger directory by walking up from start.
func LedgerDir(start string) (string, error) { return store.LedgerDir(start) }

// NewID generates a fresh, collision-free task id for the ledger.
func NewID(ledger string) (string, error) { return store.NewID(ledger) }

// NormalizeID canonicalizes a user-supplied id (e.g. "29" -> "task-029" form).
func NormalizeID(id string) string { return store.NormalizeID(id) }

// TaskPath is the on-disk path of a task file within the ledger.
func TaskPath(ledger, id string) string { return store.TaskPath(ledger, id) }

// Read loads a task by id from the ledger.
func Read(ledger, id string) (*Task, error) { return store.Read(ledger, id) }

// Write atomically persists a task to the ledger (temp file + rename).
func Write(ledger string, t *Task) error { return store.Write(ledger, t) }

// List returns all tasks in the ledger, sorted by priority then id.
func List(ledger string) ([]*Task, error) { return store.List(ledger) }

// --- Events: the audit journal ---

// AppendEvent appends one event to the ledger's audit journal.
func AppendEvent(ledger string, e Event) error { return events.Append(ledger, e) }

// ReadEvents reads the full audit journal for the ledger.
func ReadEvents(ledger string) ([]Event, error) { return events.ReadAll(ledger) }

// --- Readiness: the dependency-aware orchestration primitive ---

// IsReady reports whether a task can be claimed now: open (or in_progress with
// an expired claim), not under an active claim, and every dependency done.
func IsReady(t *Task, ledger string, now time.Time) bool {
	return ready.IsReady(t, ledger, now)
}

// CheckDeps returns nil if every dependency of t is done, else a *DepError
// (or a non-not-found read error) identifying the first unsatisfied dependency.
func CheckDeps(ledger string, t *Task) error { return ready.CheckDeps(ledger, t) }

// --- Body helpers (re-exported task body construction/parsing) ---

// SetDescription returns body with its "## Description" set to description.
func SetDescription(body, description string) string {
	return task.SetDescription(body, description)
}

// AppendNote returns body with a new note appended to its "## Notes" section.
func AppendNote(body string, when time.Time, actor, kind, message string) string {
	return task.AppendNote(body, when, actor, kind, message)
}

// ParseBody parses a task body into its structured form.
func ParseBody(body string) ParsedBody { return task.ParseBody(body) }
