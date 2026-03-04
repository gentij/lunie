package app

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
)

func newActionInput(prompt string, placeholder string, value string, limit int) textinput.Model {
	input := textinput.New()
	input.Prompt = prompt
	input.Placeholder = placeholder
	input.CharLimit = limit
	input.SetValue(value)
	input.CursorEnd()
	return input
}

func newMaskedActionInput(prompt string, limit int) textinput.Model {
	input := newActionInput(prompt, "", "", limit)
	input.EchoMode = textinput.EchoPassword
	input.EchoCharacter = '*'
	return input
}

func parseJSONObject(raw string) (map[string]any, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		trimmed = "{}"
	}
	var parsed any
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return nil, err
	}
	obj, ok := parsed.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("not an object")
	}
	return obj, nil
}

func jsonMapsEqual(a map[string]any, b map[string]any) bool {
	left, errLeft := json.Marshal(a)
	right, errRight := json.Marshal(b)
	if errLeft != nil || errRight != nil {
		return false
	}
	return string(left) == string(right)
}

func triggerCronFieldsFromJSON(configJSON string) (string, string) {
	config, err := parseJSONObject(configJSON)
	if err != nil {
		return "*/5 * * * *", "UTC"
	}
	cron := "*/5 * * * *"
	if value, ok := config["cron"].(string); ok && strings.TrimSpace(value) != "" {
		cron = strings.TrimSpace(value)
	}
	timezone := "UTC"
	if value, ok := config["timezone"].(string); ok && strings.TrimSpace(value) != "" {
		timezone = strings.TrimSpace(value)
	}
	return cron, timezone
}

func normalizeCronConfigMap(config map[string]any) map[string]any {
	normalized := map[string]any{}
	for key, value := range config {
		normalized[key] = value
	}
	cron := strings.TrimSpace(stringFromAny(normalized["cron"]))
	if cron != "" {
		normalized["cron"] = cron
	}
	timezone := strings.TrimSpace(stringFromAny(normalized["timezone"]))
	if timezone == "" {
		timezone = "UTC"
	}
	normalized["timezone"] = timezone
	if _, ok := normalized["input"]; !ok {
		normalized["input"] = map[string]any{}
	}
	return normalized
}

func stringFromAny(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return fmt.Sprintf("%v", value)
}

func isCronTriggerType(value string) bool {
	return strings.EqualFold(strings.TrimSpace(value), "CRON")
}

func (m *Model) setActionTriggerType(triggerType string) {
	next := strings.ToUpper(strings.TrimSpace(triggerType))
	if !isAllowedTriggerType(next) {
		next = "MANUAL"
	}
	previous := strings.ToUpper(strings.TrimSpace(m.action.TriggerType))
	m.action.TriggerType = next
	if isCronTriggerType(next) {
		m.action.Secondary.Prompt = "cron> "
		m.action.Secondary.Placeholder = "*/5 * * * *"
		if strings.TrimSpace(m.action.Secondary.Value()) == "" || strings.TrimSpace(m.action.Secondary.Value()) == "{}" || !isCronTriggerType(previous) {
			m.action.Secondary.SetValue("*/5 * * * *")
			m.action.Secondary.CursorEnd()
		}
		m.action.Tertiary.Prompt = "tz> "
		m.action.Tertiary.Placeholder = "UTC"
		if strings.TrimSpace(m.action.Tertiary.Value()) == "" {
			m.action.Tertiary.SetValue("UTC")
			m.action.Tertiary.CursorEnd()
		}
		return
	}
	m.action.Secondary.Prompt = "config> "
	m.action.Secondary.Placeholder = "JSON object"
	if isCronTriggerType(previous) {
		m.action.Secondary.SetValue("{}")
		m.action.Secondary.CursorEnd()
	}
	m.action.Tertiary.Prompt = "tz> "
	m.action.Tertiary.Placeholder = "UTC"
}

func (m *Model) triggerConfigFromAction() (map[string]any, error) {
	if isCronTriggerType(m.action.TriggerType) {
		cronExpr := strings.TrimSpace(m.action.Secondary.Value())
		timezone := strings.TrimSpace(m.action.Tertiary.Value())
		if timezone == "" {
			timezone = "UTC"
		}
		return map[string]any{
			"cron":     cronExpr,
			"timezone": timezone,
			"input":    map[string]any{},
		}, nil
	}
	return parseJSONObject(m.action.Secondary.Value())
}

func (m *Model) refreshActionValidation() {
	if !m.action.ShowValidation {
		return
	}
	errMessage := m.actionModalValidationError()
	m.action.Validation = errMessage
	if errMessage == "" {
		m.action.ShowValidation = false
	}
}

func (m *Model) actionModalValidationError() string {
	switch m.action.Mode {
	case actionModalRenameWorkflow:
		if strings.TrimSpace(m.action.WorkflowID) == "" {
			return "Select a workflow first"
		}
		if strings.TrimSpace(m.action.Primary.Value()) == "" {
			return "Workflow name cannot be empty"
		}
	case actionModalRenameTrigger:
		if strings.TrimSpace(m.action.WorkflowID) == "" || strings.TrimSpace(m.action.TriggerID) == "" {
			return "Select a trigger first"
		}
		if strings.TrimSpace(m.action.Primary.Value()) == "" {
			return "Trigger name cannot be empty"
		}
	case actionModalUpdateTrigger:
		if strings.TrimSpace(m.action.WorkflowID) == "" || strings.TrimSpace(m.action.TriggerID) == "" {
			return "Select a trigger first"
		}
		if strings.TrimSpace(m.action.Primary.Value()) == "" {
			return "Trigger name cannot be empty"
		}
		if isCronTriggerType(m.action.TriggerType) {
			if len(strings.Fields(strings.TrimSpace(m.action.Secondary.Value()))) != 5 {
				return "Cron expression must have 5 fields"
			}
			if strings.TrimSpace(m.action.Tertiary.Value()) == "" {
				return "Timezone cannot be empty"
			}
		} else {
			if _, err := parseJSONObject(m.action.Secondary.Value()); err != nil {
				return "Trigger config must be a valid JSON object"
			}
		}
	case actionModalCreateTrigger:
		if strings.TrimSpace(m.action.WorkflowID) == "" {
			return "Select a workflow first"
		}
		if strings.TrimSpace(m.action.Primary.Value()) == "" {
			return "Trigger name cannot be empty"
		}
		if !isAllowedTriggerType(m.action.TriggerType) {
			return "Trigger type must be MANUAL, CRON, or WEBHOOK"
		}
		if isCronTriggerType(m.action.TriggerType) {
			if len(strings.Fields(strings.TrimSpace(m.action.Secondary.Value()))) != 5 {
				return "Cron expression must have 5 fields"
			}
			if strings.TrimSpace(m.action.Tertiary.Value()) == "" {
				return "Timezone cannot be empty"
			}
		} else {
			if _, err := parseJSONObject(m.action.Secondary.Value()); err != nil {
				return "Trigger config must be a valid JSON object"
			}
		}
	case actionModalCreateSecret:
		if strings.TrimSpace(m.action.Primary.Value()) == "" {
			return "Secret name cannot be empty"
		}
		if strings.TrimSpace(m.action.Secondary.Value()) == "" {
			return "Secret value cannot be empty"
		}
	case actionModalUpdateSecret:
		if strings.TrimSpace(m.action.SecretID) == "" {
			return "Select a secret first"
		}
		secret, ok := secretByID(&m.store, m.action.SecretID)
		if !ok {
			return "Select a secret first"
		}
		if strings.TrimSpace(m.action.Primary.Value()) == "" {
			return "Secret name cannot be empty"
		}
		nameChanged := strings.TrimSpace(m.action.Primary.Value()) != secret.Name
		description := strings.TrimSpace(m.action.Tertiary.Value())
		descriptionChanged := description != "" && description != strings.TrimSpace(secret.Description)
		valueChanged := strings.TrimSpace(m.action.Secondary.Value()) != ""
		if !nameChanged && !descriptionChanged && !valueChanged {
			return "No changes to update"
		}
	case actionModalConfirmDelete:
		kind := strings.TrimSpace(strings.ToLower(m.action.DeleteKind))
		switch kind {
		case "workflow":
			if strings.TrimSpace(m.action.WorkflowID) == "" {
				return "Missing delete target"
			}
		case "trigger":
			if strings.TrimSpace(m.action.WorkflowID) == "" || strings.TrimSpace(m.action.TriggerID) == "" {
				return "Missing delete target"
			}
		case "secret":
			if strings.TrimSpace(m.action.SecretID) == "" {
				return "Missing delete target"
			}
		default:
			if strings.TrimSpace(m.action.SecretID) == "" && strings.TrimSpace(m.action.WorkflowID) == "" {
				return "Missing delete target"
			}
		}
		if strings.TrimSpace(m.action.ConfirmPhrase) == "" {
			return "Confirmation phrase is required"
		}
		if strings.TrimSpace(m.action.Confirm.Value()) != strings.TrimSpace(m.action.ConfirmPhrase) {
			return "Type the exact confirmation phrase"
		}
	}
	return ""
}

func isAllowedTriggerType(value string) bool {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "MANUAL", "CRON", "WEBHOOK":
		return true
	default:
		return false
	}
}

func (m *Model) openDeleteConfirmModal(title string, description string, phrase string, deleteKind string, workflowID string, triggerID string, secretID string) {
	confirmInput := newActionInput("confirm> ", "", "", 80)
	m.action = actionModalState{
		Active:        true,
		Mode:          actionModalConfirmDelete,
		Title:         title,
		Description:   description,
		Confirm:       confirmInput,
		Focus:         0,
		DeleteKind:    deleteKind,
		WorkflowID:    workflowID,
		TriggerID:     triggerID,
		SecretID:      secretID,
		ConfirmPhrase: phrase,
	}
	m.syncActionModalFocus()
}

func (m *Model) selectedWorkflowIDForTriggerMutation() string {
	selected := m.selectedRowID()
	if selected == "" {
		return ""
	}
	if m.view == ViewWorkflows {
		if _, ok := workflowByID(&m.store, selected); ok {
			return selected
		}
		return ""
	}
	if m.view == ViewTriggers {
		if trg, ok := triggerByID(&m.store, selected); ok {
			return trg.WorkflowID
		}
	}
	return ""
}
