package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) openRenameWorkflowModalCmd() tea.Cmd {
	if m.mutationPending {
		return m.pushToast(ToastWarn, "Another action is still in progress")
	}
	if m.view != ViewWorkflows {
		return m.pushToast(ToastWarn, "Open Workflows to rename")
	}
	selected := m.selectedRowID()
	wf, ok := workflowByID(&m.store, selected)
	if !ok {
		return m.pushToast(ToastWarn, "Select a workflow first")
	}
	input := newActionInput("name> ", "Workflow name", wf.Name, 120)
	m.action = actionModalState{
		Active:      true,
		Mode:        actionModalRenameWorkflow,
		Title:       "Rename Workflow",
		Description: "Update selected workflow name",
		Primary:     input,
		Focus:       0,
		WorkflowID:  wf.ID,
	}
	m.syncActionModalFocus()
	return nil
}

func (m *Model) openRenameTriggerModalCmd() tea.Cmd {
	if m.mutationPending {
		return m.pushToast(ToastWarn, "Another action is still in progress")
	}
	if m.view != ViewTriggers {
		return m.pushToast(ToastWarn, "Open Triggers to update")
	}
	selected := m.selectedRowID()
	trg, ok := triggerByID(&m.store, selected)
	if !ok {
		return m.pushToast(ToastWarn, "Select a trigger first")
	}
	nameInput := newActionInput("name> ", "Trigger name", trg.Name, 120)
	configInput := newActionInput("config> ", "JSON object", "{}", 2000)
	timezoneInput := newActionInput("tz> ", "UTC", "", 120)
	triggerType := strings.ToUpper(strings.TrimSpace(trg.Type))
	if triggerType == "CRON" {
		cronExpr, timezone := triggerCronFieldsFromJSON(trg.ConfigJSON)
		configInput = newActionInput("cron> ", "*/5 * * * *", cronExpr, 256)
		timezoneInput = newActionInput("tz> ", "UTC", timezone, 120)
	} else {
		configValue := strings.TrimSpace(trg.ConfigJSON)
		if configValue == "" {
			configValue = "{}"
		}
		configInput.SetValue(configValue)
		configInput.CursorEnd()
	}
	m.action = actionModalState{
		Active:        true,
		Mode:          actionModalUpdateTrigger,
		Title:         "Update Trigger",
		Description:   "Edit trigger fields and config",
		Primary:       nameInput,
		Secondary:     configInput,
		Tertiary:      timezoneInput,
		Focus:         0,
		WorkflowID:    trg.WorkflowID,
		TriggerID:     trg.ID,
		TriggerType:   triggerType,
		TriggerActive: trg.Active,
	}
	m.syncActionModalFocus()
	return nil
}

func (m *Model) openCreateTriggerModalCmd() tea.Cmd {
	if m.mutationPending {
		return m.pushToast(ToastWarn, "Another action is still in progress")
	}
	workflowID := m.selectedWorkflowIDForTriggerMutation()
	if workflowID == "" {
		return m.pushToast(ToastWarn, "Select a workflow in Workflows, or a trigger row in Triggers")
	}
	nameInput := newActionInput("name> ", "Trigger name", "", 120)
	configInput := newActionInput("config> ", "JSON object", "{}", 2000)
	timezoneInput := newActionInput("tz> ", "UTC", "UTC", 120)
	m.action = actionModalState{
		Active:        true,
		Mode:          actionModalCreateTrigger,
		Title:         "Create Trigger",
		Description:   "Create a trigger for selected workflow",
		Primary:       nameInput,
		Secondary:     configInput,
		Tertiary:      timezoneInput,
		Focus:         1,
		WorkflowID:    workflowID,
		TriggerType:   "MANUAL",
		TriggerActive: true,
	}
	m.setActionTriggerType("MANUAL")
	m.syncActionModalFocus()
	return nil
}

func (m *Model) openCreateSecretModalCmd() tea.Cmd {
	if m.mutationPending {
		return m.pushToast(ToastWarn, "Another action is still in progress")
	}
	if m.view != ViewSecrets {
		return m.pushToast(ToastWarn, "Open Secrets to create")
	}
	nameInput := newActionInput("name> ", "Secret name", "", 120)
	valueInput := newMaskedActionInput("value> ", 5000)
	descriptionInput := newActionInput("desc> ", "Optional description", "", 500)
	m.action = actionModalState{
		Active:      true,
		Mode:        actionModalCreateSecret,
		Title:       "Create Secret",
		Description: "Secret value is masked and never shown in context.",
		Primary:     nameInput,
		Secondary:   valueInput,
		Tertiary:    descriptionInput,
		Focus:       0,
	}
	m.syncActionModalFocus()
	return nil
}

func (m *Model) openUpdateSecretModalCmd() tea.Cmd {
	if m.mutationPending {
		return m.pushToast(ToastWarn, "Another action is still in progress")
	}
	if m.view != ViewSecrets {
		return m.pushToast(ToastWarn, "Open Secrets to update")
	}
	selected := m.selectedRowID()
	secret, ok := secretByID(&m.store, selected)
	if !ok {
		return m.pushToast(ToastWarn, "Select a secret first")
	}
	nameInput := newActionInput("name> ", "Secret name", secret.Name, 120)
	valueInput := newMaskedActionInput("value> ", 5000)
	descriptionInput := newActionInput("desc> ", "Optional description", secret.Description, 500)
	m.action = actionModalState{
		Active:      true,
		Mode:        actionModalUpdateSecret,
		Title:       "Update Secret",
		Description: "Leave value empty to keep existing secret value.",
		Primary:     nameInput,
		Secondary:   valueInput,
		Tertiary:    descriptionInput,
		Focus:       0,
		SecretID:    secret.ID,
	}
	m.syncActionModalFocus()
	return nil
}

func (m *Model) openDeleteSecretModalCmd() tea.Cmd {
	if m.mutationPending {
		return m.pushToast(ToastWarn, "Another action is still in progress")
	}
	if m.view != ViewSecrets {
		return m.pushToast(ToastWarn, "Open Secrets to delete")
	}
	selected := m.selectedRowID()
	secret, ok := secretByID(&m.store, selected)
	if !ok {
		return m.pushToast(ToastWarn, "Select a secret first")
	}
	phrase := "DELETE " + secret.Name
	description := "Permanently delete secret \"" + secret.Name + "\""
	m.openDeleteConfirmModal("Delete Secret", description, phrase, "secret", "", "", secret.ID)
	return nil
}

func (m *Model) openDeleteWorkflowModalCmd() tea.Cmd {
	if m.mutationPending {
		return m.pushToast(ToastWarn, "Another action is still in progress")
	}
	if m.view != ViewWorkflows {
		return m.pushToast(ToastWarn, "Open Workflows to archive")
	}
	selected := m.selectedRowID()
	wf, ok := workflowByID(&m.store, selected)
	if !ok {
		return m.pushToast(ToastWarn, "Select a workflow first")
	}
	if !wf.Active {
		return m.pushToast(ToastInfo, "Workflow already archived; press e to restore")
	}
	phrase := "ARCHIVE " + wf.Key
	description := "Archive workflow \"" + wf.Name + "\" [" + wf.Key + "] and set it inactive"
	m.openDeleteConfirmModal("Archive Workflow", description, phrase, "workflow", wf.ID, "", "")
	return nil
}

func (m *Model) openDeleteTriggerModalCmd() tea.Cmd {
	if m.mutationPending {
		return m.pushToast(ToastWarn, "Another action is still in progress")
	}
	if m.view != ViewTriggers {
		return m.pushToast(ToastWarn, "Open Triggers to archive")
	}
	selected := m.selectedRowID()
	trg, ok := triggerByID(&m.store, selected)
	if !ok {
		return m.pushToast(ToastWarn, "Select a trigger first")
	}
	if !trg.Active {
		return m.pushToast(ToastInfo, "Trigger already archived; press e to restore")
	}
	phrase := "ARCHIVE " + trg.Key
	description := "Archive trigger \"" + trg.Name + "\" [" + trg.Key + "] and set it inactive"
	m.openDeleteConfirmModal("Archive Trigger", description, phrase, "trigger", trg.WorkflowID, trg.ID, "")
	return nil
}

func (m *Model) openCLIHandoffModalCmd(topic string) tea.Cmd {
	description := "Use the CLI for definition authoring"
	command := ""
	title := "CLI Handoff"
	selectedWorkflow := m.selectedWorkflowIDForTriggerMutation()
	if selectedWorkflow == "" && m.view == ViewWorkflows {
		selectedWorkflow = m.selectedRowID()
	}
	workflowTarget := "<workflow-key>"
	if strings.TrimSpace(selectedWorkflow) != "" {
		if workflow, ok := workflowByID(&m.store, selectedWorkflow); ok {
			workflowTarget = workflow.Key
		}
	}
	switch strings.TrimSpace(topic) {
	case "workflow-version-create":
		title = "Create Workflow Version"
		description = "Workflow versions stay CLI-first for JSON authoring and validation"
		command = "lunie workflow version create " + workflowTarget + " --definition ./workflow-definition.json"
	default:
		title = "Create Workflow"
		description = "Workflow creation stays CLI-first for JSON authoring and validation"
		command = "lunie workflow create --name \"my-workflow\" --definition ./workflow-definition.json"
	}
	m.action = actionModalState{
		Active:      true,
		Mode:        actionModalCLIHandoff,
		Title:       title,
		Description: description,
		CLICommand:  command,
	}
	return nil
}
