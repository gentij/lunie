package app

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updateActionModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.action.Active {
		return m, nil
	}
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, m.keys.Back) || key.Matches(keyMsg, m.keys.Quit) {
			m.action = actionModalState{}
			return m, nil
		}
		if m.action.Mode == actionModalCLIHandoff {
			if key.Matches(keyMsg, m.keys.Enter) {
				m.action = actionModalState{}
			}
			return m, nil
		}
		if key.Matches(keyMsg, m.keys.Enter) {
			if errMessage := m.actionModalValidationError(); errMessage != "" {
				m.action.Validation = errMessage
				m.action.ShowValidation = true
				return m, nil
			}
			m.action.Validation = ""
			m.action.ShowValidation = false
			return m, m.submitActionModal()
		}
		if key.Matches(keyMsg, m.keys.NextScreen) {
			m.cycleActionModalFocus(1)
			return m, nil
		}
		if key.Matches(keyMsg, m.keys.PrevScreen) {
			m.cycleActionModalFocus(-1)
			return m, nil
		}
		if m.action.Mode == actionModalCreateTrigger || m.action.Mode == actionModalUpdateTrigger {
			switch keyMsg.String() {
			case "left", "h":
				if m.action.Mode == actionModalCreateTrigger && m.action.Focus == 0 {
					m.cycleActionTriggerType(-1)
					m.refreshActionValidation()
					return m, nil
				}
			case "right", "l":
				if m.action.Mode == actionModalCreateTrigger && m.action.Focus == 0 {
					m.cycleActionTriggerType(1)
					m.refreshActionValidation()
					return m, nil
				}
			case " ":
				toggleFocus := 2
				if m.action.Mode == actionModalUpdateTrigger {
					toggleFocus = 1
				}
				if m.action.Focus == toggleFocus {
					m.action.TriggerActive = !m.action.TriggerActive
					m.refreshActionValidation()
					return m, nil
				}
			}
		}
		if m.action.Mode == actionModalConfirmDelete {
			if key.Matches(keyMsg, m.keys.Clear) {
				m.action.Confirm.SetValue("")
				m.action.Confirm.CursorEnd()
				m.refreshActionValidation()
				return m, nil
			}
		}
		if key.Matches(keyMsg, m.keys.Clear) {
			switch m.action.Mode {
			case actionModalRenameWorkflow, actionModalRenameTrigger:
				m.action.Primary.SetValue("")
				m.action.Primary.CursorEnd()
				m.refreshActionValidation()
				return m, nil
			case actionModalCreateTrigger, actionModalUpdateTrigger:
				nameFocus := 1
				configFocus := 3
				timezoneFocus := 4
				if m.action.Mode == actionModalUpdateTrigger {
					nameFocus = 0
					configFocus = 2
					timezoneFocus = 3
				}
				if m.action.Focus == nameFocus {
					m.action.Primary.SetValue("")
					m.action.Primary.CursorEnd()
					m.refreshActionValidation()
					return m, nil
				}
				if m.action.Focus == configFocus {
					if isCronTriggerType(m.action.TriggerType) {
						m.action.Secondary.SetValue("*/5 * * * *")
					} else {
						m.action.Secondary.SetValue("{}")
					}
					m.action.Secondary.CursorEnd()
					m.refreshActionValidation()
					return m, nil
				}
				if isCronTriggerType(m.action.TriggerType) && m.action.Focus == timezoneFocus {
					m.action.Tertiary.SetValue("UTC")
					m.action.Tertiary.CursorEnd()
					m.refreshActionValidation()
					return m, nil
				}
			case actionModalCreateSecret, actionModalUpdateSecret:
				if m.action.Focus == 0 {
					m.action.Primary.SetValue("")
					m.action.Primary.CursorEnd()
					m.refreshActionValidation()
					return m, nil
				}
				if m.action.Focus == 1 {
					m.action.Secondary.SetValue("")
					m.action.Secondary.CursorEnd()
					m.refreshActionValidation()
					return m, nil
				}
				if m.action.Focus == 2 {
					m.action.Tertiary.SetValue("")
					m.action.Tertiary.CursorEnd()
					m.refreshActionValidation()
					return m, nil
				}
			}
		}
	}

	var cmd tea.Cmd
	switch m.action.Mode {
	case actionModalRenameWorkflow, actionModalRenameTrigger:
		m.action.Primary, cmd = m.action.Primary.Update(msg)
		m.refreshActionValidation()
		return m, cmd
	case actionModalCreateTrigger, actionModalUpdateTrigger:
		nameFocus := 1
		configFocus := 3
		timezoneFocus := 4
		if m.action.Mode == actionModalUpdateTrigger {
			nameFocus = 0
			configFocus = 2
			timezoneFocus = 3
		}
		if m.action.Focus == nameFocus {
			m.action.Primary, cmd = m.action.Primary.Update(msg)
			m.refreshActionValidation()
			return m, cmd
		}
		if m.action.Focus == configFocus {
			m.action.Secondary, cmd = m.action.Secondary.Update(msg)
			m.refreshActionValidation()
			return m, cmd
		}
		if isCronTriggerType(m.action.TriggerType) && m.action.Focus == timezoneFocus {
			m.action.Tertiary, cmd = m.action.Tertiary.Update(msg)
			m.refreshActionValidation()
			return m, cmd
		}
	case actionModalCreateSecret, actionModalUpdateSecret:
		if m.action.Focus == 0 {
			m.action.Primary, cmd = m.action.Primary.Update(msg)
			m.refreshActionValidation()
			return m, cmd
		}
		if m.action.Focus == 1 {
			m.action.Secondary, cmd = m.action.Secondary.Update(msg)
			m.refreshActionValidation()
			return m, cmd
		}
		if m.action.Focus == 2 {
			m.action.Tertiary, cmd = m.action.Tertiary.Update(msg)
			m.refreshActionValidation()
			return m, cmd
		}
	case actionModalConfirmDelete:
		m.action.Confirm, cmd = m.action.Confirm.Update(msg)
		m.refreshActionValidation()
		return m, cmd
	}
	return m, nil
}

func (m *Model) submitActionModal() tea.Cmd {
	switch m.action.Mode {
	case actionModalRenameWorkflow:
		name := strings.TrimSpace(m.action.Primary.Value())
		if name == "" {
			return m.pushToast(ToastWarn, "Workflow name cannot be empty")
		}
		workflowID := m.action.WorkflowID
		m.action = actionModalState{}
		return m.renameWorkflowCmd(workflowID, name)
	case actionModalRenameTrigger, actionModalUpdateTrigger:
		name := strings.TrimSpace(m.action.Primary.Value())
		workflowID := m.action.WorkflowID
		triggerID := m.action.TriggerID
		active := m.action.TriggerActive
		configValue, err := m.triggerConfigFromAction()
		if err != nil {
			return m.pushToast(ToastWarn, "Trigger config must be a valid JSON object")
		}
		m.action = actionModalState{}
		return m.updateTriggerCmd(workflowID, triggerID, name, active, configValue)
	case actionModalCreateTrigger:
		name := strings.TrimSpace(m.action.Primary.Value())
		configValue, err := m.triggerConfigFromAction()
		if err != nil {
			return m.pushToast(ToastWarn, "Trigger config must be a valid JSON object")
		}
		workflowID := m.action.WorkflowID
		triggerType := m.action.TriggerType
		active := m.action.TriggerActive
		m.action = actionModalState{}
		return m.createTriggerCmd(workflowID, triggerType, name, active, configValue)
	case actionModalCreateSecret:
		name := strings.TrimSpace(m.action.Primary.Value())
		value := m.action.Secondary.Value()
		description := strings.TrimSpace(m.action.Tertiary.Value())
		m.action = actionModalState{}
		return m.createSecretCmd(name, value, description)
	case actionModalUpdateSecret:
		secretID := strings.TrimSpace(m.action.SecretID)
		name := strings.TrimSpace(m.action.Primary.Value())
		value := m.action.Secondary.Value()
		description := strings.TrimSpace(m.action.Tertiary.Value())
		m.action = actionModalState{}
		return m.updateSecretCmd(secretID, name, value, description)
	case actionModalConfirmDelete:
		if errMessage := m.actionModalValidationError(); errMessage != "" {
			m.action.Validation = errMessage
			m.action.ShowValidation = true
			return nil
		}
		kind := strings.TrimSpace(strings.ToLower(m.action.DeleteKind))
		workflowID := m.action.WorkflowID
		triggerID := m.action.TriggerID
		secretID := m.action.SecretID
		m.action = actionModalState{}
		if kind == "secret" {
			return m.deleteSecretCmd(secretID)
		}
		if strings.TrimSpace(triggerID) != "" {
			return m.deleteTriggerCmd(workflowID, triggerID)
		}
		return m.deleteWorkflowCmd(workflowID)
	case actionModalCLIHandoff:
		m.action = actionModalState{}
		return nil
	default:
		m.action = actionModalState{}
		return nil
	}
}

func (m *Model) cycleActionModalFocus(delta int) {
	total := 0
	switch m.action.Mode {
	case actionModalRenameWorkflow, actionModalRenameTrigger:
		total = 1
	case actionModalCreateTrigger:
		total = 4
		if isCronTriggerType(m.action.TriggerType) {
			total = 5
		}
	case actionModalUpdateTrigger:
		total = 3
		if isCronTriggerType(m.action.TriggerType) {
			total = 4
		}
	case actionModalCreateSecret, actionModalUpdateSecret:
		total = 3
	case actionModalConfirmDelete:
		total = 1
	default:
		total = 0
	}
	if total <= 1 {
		return
	}
	next := m.action.Focus + delta
	for next < 0 {
		next += total
	}
	m.action.Focus = next % total
	m.syncActionModalFocus()
}

func (m *Model) syncActionModalFocus() {
	m.action.Primary.Blur()
	m.action.Secondary.Blur()
	m.action.Tertiary.Blur()
	m.action.Confirm.Blur()
	switch m.action.Mode {
	case actionModalRenameWorkflow, actionModalRenameTrigger:
		m.action.Primary.Focus()
	case actionModalCreateTrigger:
		if m.action.Focus == 1 {
			m.action.Primary.Focus()
		}
		if m.action.Focus == 3 {
			m.action.Secondary.Focus()
		}
		if isCronTriggerType(m.action.TriggerType) && m.action.Focus == 4 {
			m.action.Tertiary.Focus()
		}
	case actionModalUpdateTrigger:
		if m.action.Focus == 0 {
			m.action.Primary.Focus()
		}
		if m.action.Focus == 2 {
			m.action.Secondary.Focus()
		}
		if isCronTriggerType(m.action.TriggerType) && m.action.Focus == 3 {
			m.action.Tertiary.Focus()
		}
	case actionModalCreateSecret, actionModalUpdateSecret:
		if m.action.Focus == 0 {
			m.action.Primary.Focus()
		}
		if m.action.Focus == 1 {
			m.action.Secondary.Focus()
		}
		if m.action.Focus == 2 {
			m.action.Tertiary.Focus()
		}
	case actionModalConfirmDelete:
		m.action.Confirm.Focus()
	}
}

func (m *Model) cycleActionTriggerType(delta int) {
	order := []string{"MANUAL", "CRON", "WEBHOOK"}
	index := 0
	for i, item := range order {
		if strings.EqualFold(item, m.action.TriggerType) {
			index = i
			break
		}
	}
	next := index + delta
	for next < 0 {
		next += len(order)
	}
	m.setActionTriggerType(order[next%len(order)])
}
