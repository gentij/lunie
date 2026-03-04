package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) toggleWorkflowActiveCmd() tea.Cmd {
	if m.mutationPending {
		return m.pushToast(ToastWarn, "Another action is still in progress")
	}
	selected := m.selectedRowID()
	if selected == "" {
		return m.pushToast(ToastWarn, "Select a workflow first")
	}
	wf, ok := workflowByID(&m.store, selected)
	if !ok {
		return m.pushToast(ToastWarn, "Select a workflow first")
	}
	client := m.client
	next := !wf.Active
	m.mutationPending = true
	return func() tea.Msg {
		if client == nil {
			return mutationResultMsg{err: fmt.Errorf("api client unavailable")}
		}
		_, err := client.UpdateWorkflow(selected, map[string]any{"isActive": next})
		if err != nil {
			return mutationResultMsg{err: err}
		}
		message := "Workflow archived"
		if next {
			message = "Workflow restored"
		}
		return mutationResultMsg{successMessage: message, refresh: true}
	}
}

func (m *Model) queueRunForSelectedWorkflowCmd() tea.Cmd {
	if m.mutationPending {
		return m.pushToast(ToastWarn, "Another action is still in progress")
	}
	if m.view != ViewWorkflows {
		return m.pushToast(ToastWarn, "Open Workflows to run")
	}
	selected := m.selectedRowID()
	if selected == "" {
		return m.pushToast(ToastWarn, "Select a workflow first")
	}
	client := m.client
	m.mutationPending = true
	return func() tea.Msg {
		if client == nil {
			return mutationResultMsg{err: fmt.Errorf("api client unavailable")}
		}
		result, err := client.RunWorkflow(selected, map[string]any{}, map[string]any{})
		if err != nil {
			return mutationResultMsg{err: err}
		}
		message := "Workflow run queued"
		if id, ok := result["workflowRunId"]; ok && strings.TrimSpace(id) != "" {
			message = "Workflow run queued: " + id
		}
		return mutationResultMsg{successMessage: message, refresh: true}
	}
}

func (m *Model) toggleTriggerActiveCmd() tea.Cmd {
	if m.mutationPending {
		return m.pushToast(ToastWarn, "Another action is still in progress")
	}
	if m.view != ViewTriggers {
		return m.pushToast(ToastWarn, "Open Triggers to toggle")
	}
	selected := m.selectedRowID()
	if selected == "" {
		return m.pushToast(ToastWarn, "Select a trigger first")
	}
	trg, ok := triggerByID(&m.store, selected)
	if !ok {
		return m.pushToast(ToastWarn, "Select a trigger first")
	}
	client := m.client
	next := !trg.Active
	m.mutationPending = true
	return func() tea.Msg {
		if client == nil {
			return mutationResultMsg{err: fmt.Errorf("api client unavailable")}
		}
		_, err := client.UpdateTrigger(trg.WorkflowID, trg.ID, map[string]any{"isActive": next})
		if err != nil {
			return mutationResultMsg{err: err}
		}
		message := "Trigger archived"
		if next {
			message = "Trigger restored"
		}
		return mutationResultMsg{successMessage: message, refresh: true}
	}
}

func (m *Model) deleteWorkflowCmd(workflowID string) tea.Cmd {
	if m.mutationPending {
		return m.pushToast(ToastWarn, "Another action is still in progress")
	}
	workflowID = strings.TrimSpace(workflowID)
	if workflowID == "" {
		return m.pushToast(ToastWarn, "Select a workflow first")
	}
	client := m.client
	m.mutationPending = true
	return func() tea.Msg {
		if client == nil {
			return mutationResultMsg{err: fmt.Errorf("api client unavailable")}
		}
		_, err := client.DeleteWorkflow(workflowID)
		if err != nil {
			return mutationResultMsg{err: err}
		}
		return mutationResultMsg{successMessage: "Workflow archived", refresh: true}
	}
}

func (m *Model) deleteTriggerCmd(workflowID string, triggerID string) tea.Cmd {
	if m.mutationPending {
		return m.pushToast(ToastWarn, "Another action is still in progress")
	}
	workflowID = strings.TrimSpace(workflowID)
	triggerID = strings.TrimSpace(triggerID)
	if workflowID == "" || triggerID == "" {
		return m.pushToast(ToastWarn, "Select a trigger first")
	}
	client := m.client
	m.mutationPending = true
	return func() tea.Msg {
		if client == nil {
			return mutationResultMsg{err: fmt.Errorf("api client unavailable")}
		}
		_, err := client.DeleteTrigger(workflowID, triggerID)
		if err != nil {
			return mutationResultMsg{err: err}
		}
		return mutationResultMsg{successMessage: "Trigger archived", refresh: true}
	}
}

func (m *Model) createSecretCmd(name string, value string, description string) tea.Cmd {
	if m.mutationPending {
		return m.pushToast(ToastWarn, "Another action is still in progress")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return m.pushToast(ToastWarn, "Secret name cannot be empty")
	}
	if strings.TrimSpace(value) == "" {
		return m.pushToast(ToastWarn, "Secret value cannot be empty")
	}
	payload := map[string]any{"name": name, "value": value}
	description = strings.TrimSpace(description)
	if description != "" {
		payload["description"] = description
	}
	client := m.client
	m.mutationPending = true
	return func() tea.Msg {
		if client == nil {
			return mutationResultMsg{err: fmt.Errorf("api client unavailable")}
		}
		_, err := client.CreateSecret(payload)
		if err != nil {
			return mutationResultMsg{err: err}
		}
		return mutationResultMsg{successMessage: "Secret created", refresh: true}
	}
}

func (m *Model) updateSecretCmd(secretID string, name string, value string, description string) tea.Cmd {
	if m.mutationPending {
		return m.pushToast(ToastWarn, "Another action is still in progress")
	}
	secretID = strings.TrimSpace(secretID)
	if secretID == "" {
		return m.pushToast(ToastWarn, "Select a secret first")
	}
	secret, ok := secretByID(&m.store, secretID)
	if !ok {
		return m.pushToast(ToastWarn, "Select a secret first")
	}
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if name == "" {
		return m.pushToast(ToastWarn, "Secret name cannot be empty")
	}
	patch := map[string]any{}
	if name != secret.Name {
		patch["name"] = name
	}
	if description != "" && description != strings.TrimSpace(secret.Description) {
		patch["description"] = description
	}
	if strings.TrimSpace(value) != "" {
		patch["value"] = value
	}
	if len(patch) == 0 {
		return m.pushToast(ToastWarn, "No changes to update")
	}
	client := m.client
	m.mutationPending = true
	return func() tea.Msg {
		if client == nil {
			return mutationResultMsg{err: fmt.Errorf("api client unavailable")}
		}
		_, err := client.UpdateSecret(secretID, patch)
		if err != nil {
			return mutationResultMsg{err: err}
		}
		return mutationResultMsg{successMessage: "Secret updated", refresh: true}
	}
}

func (m *Model) deleteSecretCmd(secretID string) tea.Cmd {
	if m.mutationPending {
		return m.pushToast(ToastWarn, "Another action is still in progress")
	}
	secretID = strings.TrimSpace(secretID)
	if secretID == "" {
		return m.pushToast(ToastWarn, "Select a secret first")
	}
	client := m.client
	m.mutationPending = true
	return func() tea.Msg {
		if client == nil {
			return mutationResultMsg{err: fmt.Errorf("api client unavailable")}
		}
		_, err := client.DeleteSecret(secretID)
		if err != nil {
			return mutationResultMsg{err: err}
		}
		return mutationResultMsg{successMessage: "Secret deleted", refresh: true}
	}
}

func (m *Model) renameWorkflowCmd(workflowID string, name string) tea.Cmd {
	if m.mutationPending {
		return m.pushToast(ToastWarn, "Another action is still in progress")
	}
	name = strings.TrimSpace(name)
	if workflowID == "" {
		return m.pushToast(ToastWarn, "Select a workflow first")
	}
	if name == "" {
		return m.pushToast(ToastWarn, "Workflow name cannot be empty")
	}
	client := m.client
	m.mutationPending = true
	return func() tea.Msg {
		if client == nil {
			return mutationResultMsg{err: fmt.Errorf("api client unavailable")}
		}
		_, err := client.UpdateWorkflow(workflowID, map[string]any{"name": name})
		if err != nil {
			return mutationResultMsg{err: err}
		}
		return mutationResultMsg{successMessage: "Workflow renamed", refresh: true}
	}
}

func (m *Model) updateTriggerCmd(workflowID string, triggerID string, name string, active bool, configValue map[string]any) tea.Cmd {
	if m.mutationPending {
		return m.pushToast(ToastWarn, "Another action is still in progress")
	}
	name = strings.TrimSpace(name)
	if workflowID == "" || triggerID == "" {
		return m.pushToast(ToastWarn, "Select a trigger first")
	}
	if name == "" {
		return m.pushToast(ToastWarn, "Trigger name cannot be empty")
	}
	trigger, ok := triggerByID(&m.store, triggerID)
	if !ok {
		return m.pushToast(ToastWarn, "Select a trigger first")
	}
	if trigger.WorkflowID != workflowID {
		return m.pushToast(ToastWarn, "Trigger does not match selected workflow")
	}
	patch := map[string]any{}
	if name != strings.TrimSpace(trigger.Name) {
		patch["name"] = name
	}
	if active != trigger.Active {
		patch["isActive"] = active
	}
	if configValue == nil {
		configValue = map[string]any{}
	}
	existingConfig, err := parseJSONObject(trigger.ConfigJSON)
	if err != nil {
		existingConfig = map[string]any{}
	}
	if isCronTriggerType(trigger.Type) {
		existingConfig = normalizeCronConfigMap(existingConfig)
		configValue = normalizeCronConfigMap(configValue)
	}
	if !jsonMapsEqual(existingConfig, configValue) {
		patch["config"] = configValue
	}
	if len(patch) == 0 {
		return m.pushToast(ToastWarn, "No changes to update")
	}
	client := m.client
	m.mutationPending = true
	return func() tea.Msg {
		if client == nil {
			return mutationResultMsg{err: fmt.Errorf("api client unavailable")}
		}
		_, err := client.UpdateTrigger(workflowID, triggerID, patch)
		if err != nil {
			return mutationResultMsg{err: err}
		}
		return mutationResultMsg{successMessage: "Trigger updated", refresh: true}
	}
}

func (m *Model) createTriggerCmd(workflowID string, triggerType string, name string, active bool, configValue map[string]any) tea.Cmd {
	if m.mutationPending {
		return m.pushToast(ToastWarn, "Another action is still in progress")
	}
	workflowID = strings.TrimSpace(workflowID)
	name = strings.TrimSpace(name)
	if workflowID == "" {
		return m.pushToast(ToastWarn, "Select a workflow first")
	}
	if name == "" {
		return m.pushToast(ToastWarn, "Trigger name cannot be empty")
	}
	if configValue == nil {
		configValue = map[string]any{}
	}
	client := m.client
	m.mutationPending = true
	triggerType = strings.ToUpper(strings.TrimSpace(triggerType))
	return func() tea.Msg {
		if client == nil {
			return mutationResultMsg{err: fmt.Errorf("api client unavailable")}
		}
		payload := map[string]any{
			"type":     triggerType,
			"name":     name,
			"isActive": active,
			"config":   configValue,
		}
		_, err := client.CreateTrigger(workflowID, payload)
		if err != nil {
			return mutationResultMsg{err: err}
		}
		return mutationResultMsg{successMessage: "Trigger created", refresh: true}
	}
}
