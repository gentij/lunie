package app

import (
	"strings"
	"testing"
	"time"

	"github.com/gentij/taskforge/apps/cli/internal/tui/data"
)

func TestRecentRunRows_StatusColumnUsesStatusValue(t *testing.T) {
	now := time.Now()
	store := data.Store{
		Workflows: []data.Workflow{{ID: "wf_1", Name: "My Workflow", Active: true}},
		Runs: []data.WorkflowRun{{
			ID:         "run_1",
			WorkflowID: "wf_1",
			Status:     "FAILED",
			StartedAt:  now.Add(-2 * time.Minute),
		}},
	}

	rows, ids := recentRunRows(&store, now, 10)
	if len(ids) != 1 || ids[0] != "run_1" {
		t.Fatalf("unexpected ids: %#v", ids)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if len(rows[0]) != 4 {
		t.Fatalf("expected 4 columns, got %d", len(rows[0]))
	}
	if got := rows[0][2]; got != "FAILED" {
		t.Fatalf("status column mismatch: got %q, want %q", got, "FAILED")
	}
	if got := rows[0][3]; !strings.Contains(got, "ago") && got != "just now" {
		t.Fatalf("started column should be relative time, got %q", got)
	}
}
