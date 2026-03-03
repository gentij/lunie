package app

import (
	"testing"
	"time"

	"github.com/gentij/taskforge/apps/cli/internal/config"
	"github.com/gentij/taskforge/apps/cli/internal/tui/data"
)

func TestRefreshView_PreservesSelectionByRowID(t *testing.T) {
	now := time.Now()
	m := NewModel(nil, "", false, config.Config{}, "")
	m.view = ViewWorkflows
	m.store = data.Store{
		Workflows: []data.Workflow{
			{ID: "wf_a", Name: "A", Active: true, LatestVersion: 1, UpdatedAt: now.Add(-3 * time.Hour)},
			{ID: "wf_b", Name: "B", Active: true, LatestVersion: 1, UpdatedAt: now.Add(-2 * time.Hour)},
			{ID: "wf_c", Name: "C", Active: true, LatestVersion: 1, UpdatedAt: now.Add(-1 * time.Hour)},
		},
	}

	m.refreshView()
	if len(m.filteredRowIDs) < 2 {
		t.Fatalf("expected at least 2 rows, got %d", len(m.filteredRowIDs))
	}

	m.table.SetCursor(1)
	selectedID := m.selectedRowID()
	if selectedID == "" {
		t.Fatal("expected selected row id")
	}

	for i := range m.store.Workflows {
		if m.store.Workflows[i].ID == selectedID {
			m.store.Workflows[i].UpdatedAt = now.Add(4 * time.Hour)
		}
	}

	m.refreshView()
	if got := m.selectedRowID(); got != selectedID {
		t.Fatalf("selection not preserved: got %q, want %q", got, selectedID)
	}
}

func TestParseJSONObject_RequiresObject(t *testing.T) {
	if _, err := parseJSONObject(`{"ok":true}`); err != nil {
		t.Fatalf("expected object to parse, got error: %v", err)
	}
	if _, err := parseJSONObject(`[1,2,3]`); err == nil {
		t.Fatal("expected array payload to be rejected")
	}
}
