package app

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/gentij/lunie/apps/cli/internal/config"
	"github.com/gentij/lunie/apps/cli/internal/tui/data"
)

func TestRefreshView_PreservesSelectionByRowID(t *testing.T) {
	now := time.Now()
	m := NewModel(nil, "", false, config.Config{}, "")
	m.view = ViewWorkflows
	m.store = data.Store{
		Workflows: []data.Workflow{
			{ID: "wf_a", Key: "a", Name: "A", Active: true, LatestVersion: 1, UpdatedAt: now.Add(-3 * time.Hour)},
			{ID: "wf_b", Key: "b", Name: "B", Active: true, LatestVersion: 1, UpdatedAt: now.Add(-2 * time.Hour)},
			{ID: "wf_c", Key: "c", Name: "C", Active: true, LatestVersion: 1, UpdatedAt: now.Add(-1 * time.Hour)},
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

func TestDeleteConfirmModal_RequiresExactPhrase(t *testing.T) {
	m := NewModel(nil, "", false, config.Config{}, "")
	m.openDeleteConfirmModal("Archive Workflow", "Archive test workflow", "ARCHIVE workflow-key", "workflow", "wf_test", "", "")

	if got := m.actionModalValidationError(); got == "" {
		t.Fatal("expected validation error when phrase is empty")
	}

	m.action.Confirm.SetValue("ARCHIVE something-else")
	if got := m.actionModalValidationError(); got == "" {
		t.Fatal("expected validation error for mismatched phrase")
	}

	m.action.Confirm.SetValue("ARCHIVE workflow-key")
	if got := m.actionModalValidationError(); got != "" {
		t.Fatalf("expected valid confirmation phrase, got error: %q", got)
	}
}

func TestSubmitDeleteConfirmDispatchesDeleteCmd(t *testing.T) {
	m := NewModel(nil, "", false, config.Config{}, "")
	m.openDeleteConfirmModal("Archive Trigger", "Archive test trigger", "ARCHIVE trigger-key", "trigger", "wf_test", "trg_test", "")
	m.action.Confirm.SetValue("ARCHIVE trigger-key")

	cmd := m.submitActionModal()
	if cmd == nil {
		t.Fatal("expected archive command to be returned")
	}

	msg := cmd()
	result, ok := msg.(mutationResultMsg)
	if !ok {
		t.Fatalf("expected mutationResultMsg, got %T", msg)
	}
	if result.err == nil {
		t.Fatal("expected command to fail without API client")
	}
	if !strings.Contains(result.err.Error(), "api client unavailable") {
		t.Fatalf("unexpected error: %v", result.err)
	}
}

func TestCreateSecretValidation_RequiresNameAndValue(t *testing.T) {
	m := NewModel(nil, "", false, config.Config{}, "")
	m.action = actionModalState{
		Active:    true,
		Mode:      actionModalCreateSecret,
		Primary:   newActionInput("name> ", "", "", 120),
		Secondary: newMaskedActionInput("value> ", 5000),
		Tertiary:  newActionInput("desc> ", "", "", 500),
	}
	if got := m.actionModalValidationError(); got == "" {
		t.Fatal("expected validation error for empty secret fields")
	}
	m.action.Primary.SetValue("API_KEY")
	if got := m.actionModalValidationError(); got == "" {
		t.Fatal("expected validation error for empty secret value")
	}
	m.action.Secondary.SetValue("super-secret")
	if got := m.actionModalValidationError(); got != "" {
		t.Fatalf("expected valid secret create payload, got error: %q", got)
	}
}

func TestTriggerConfigFromAction_CronTypedFields(t *testing.T) {
	m := NewModel(nil, "", false, config.Config{}, "")
	m.action = actionModalState{
		Mode:        actionModalCreateTrigger,
		TriggerType: "CRON",
		Secondary:   newActionInput("cron> ", "", "*/5 * * * *", 256),
		Tertiary:    newActionInput("tz> ", "", "UTC", 120),
	}
	config, err := m.triggerConfigFromAction()
	if err != nil {
		t.Fatalf("expected cron config to parse, got error: %v", err)
	}
	if got, ok := config["cron"].(string); !ok || got != "*/5 * * * *" {
		t.Fatalf("unexpected cron field: %#v", config["cron"])
	}
	if got, ok := config["timezone"].(string); !ok || got != "UTC" {
		t.Fatalf("unexpected timezone field: %#v", config["timezone"])
	}
}

func TestTriggerValidation_CronRequiresFiveFields(t *testing.T) {
	m := NewModel(nil, "", false, config.Config{}, "")
	m.action = actionModalState{
		Mode:        actionModalCreateTrigger,
		WorkflowID:  "wf_1",
		TriggerType: "CRON",
		Primary:     newActionInput("name> ", "", "nightly", 120),
		Secondary:   newActionInput("cron> ", "", "*/5 * * *", 256),
		Tertiary:    newActionInput("tz> ", "", "UTC", 120),
	}
	if got := m.actionModalValidationError(); got == "" {
		t.Fatal("expected cron validation error for invalid field count")
	}
}

func TestScopeRowsForCurrentView_ActiveOnlyWorkflows(t *testing.T) {
	now := time.Now()
	m := NewModel(nil, "", false, config.Config{}, "")
	m.view = ViewWorkflows
	m.store = data.Store{
		Workflows: []data.Workflow{
			{ID: "wf_active", Key: "active", Name: "Active", Active: true, UpdatedAt: now},
			{ID: "wf_inactive", Key: "inactive", Name: "Inactive", Active: false, UpdatedAt: now},
		},
	}
	rows := []table.Row{{"Active"}, {"Inactive"}}
	ids := []string{"wf_active", "wf_inactive"}

	m.setStatusScopeForView(ViewWorkflows, statusScopeActive)
	filteredRows, filteredIDs := m.scopeRowsForCurrentView(rows, ids)
	if len(filteredRows) != 1 || len(filteredIDs) != 1 {
		t.Fatalf("expected one active row, got rows=%d ids=%d", len(filteredRows), len(filteredIDs))
	}
	if filteredIDs[0] != "wf_active" {
		t.Fatalf("unexpected row id: %q", filteredIDs[0])
	}
}
