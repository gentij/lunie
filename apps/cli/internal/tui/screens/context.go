package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/gentij/taskforge/apps/cli/internal/tui/data"
	"github.com/gentij/taskforge/apps/cli/internal/tui/utils"
)

func BuildContextContent(view ViewID, store *data.Store, selectedID string) string {
	switch view {
	case ViewDashboard:
		return dashboardContext(store)
	case ViewWorkflows:
		return workflowContext(store, selectedID)
	case ViewRuns:
		return runContext(store, selectedID)
	case ViewTriggers:
		return triggerContext(store, selectedID)
	case ViewEvents:
		return eventContext(store, selectedID)
	case ViewSecrets:
		return secretContext(store, selectedID)
	case ViewTokens:
		return tokenContext(store, selectedID)
	default:
		return ""
	}
}

func BuildContextTabContent(view ViewID, store *data.Store, selectedID string, tab ContextTab) string {
	switch tab {
	case ContextTabJSON:
		return contextJSONContent(view, store, selectedID)
	case ContextTabSteps:
		return contextStepsContent(view, store, selectedID)
	case ContextTabLogs:
		return contextLogsContent(view, store, selectedID)
	default:
		return BuildContextContent(view, store, selectedID)
	}
}

func contextJSONContent(view ViewID, store *data.Store, selectedID string) string {
	switch view {
	case ViewDashboard, ViewRuns:
		run, ok := runByID(store, selectedID)
		if !ok {
			return "No run selected"
		}
		parts := []string{
			"Input JSON",
			utils.Indent(utils.PrettyJSON(run.InputJSON), "  "),
			"",
			"Output JSON",
			utils.Indent(utils.PrettyJSON(run.OutputJSON), "  "),
		}
		if strings.TrimSpace(run.ErrorJSON) != "" {
			parts = append(parts, "", "Error JSON", utils.Indent(utils.PrettyJSON(run.ErrorJSON), "  "))
		}
		return strings.Join(parts, "\n")
	case ViewWorkflows:
		wf, ok := workflowByID(store, selectedID)
		if !ok {
			return "No workflow selected"
		}
		versions := versionsForWorkflow(store, selectedID)
		def := latestDefinition(versions)
		return strings.Join([]string{
			"Workflow",
			utils.Indent(utils.PrettyJSON(fmt.Sprintf(`{"id":"%s","name":"%s","active":%t,"latestVersion":%d}`, wf.ID, wf.Name, wf.Active, wf.LatestVersion)), "  "),
			"",
			"Latest Definition",
			utils.Indent(utils.PrettyJSON(def), "  "),
		}, "\n")
	case ViewTriggers:
		trg, ok := triggerByID(store, selectedID)
		if !ok {
			return "No trigger selected"
		}
		return strings.Join([]string{"Trigger Config", utils.Indent(utils.PrettyJSON(trg.ConfigJSON), "  ")}, "\n")
	case ViewEvents:
		evt, ok := eventByID(store, selectedID)
		if !ok {
			return "No event selected"
		}
		return strings.Join([]string{
			"Payload",
			utils.Indent(utils.PrettyJSON(evt.PayloadJSON), "  "),
			"",
			"Metadata",
			utils.Indent(utils.PrettyJSON(evt.Metadata), "  "),
		}, "\n")
	case ViewSecrets:
		sec, ok := secretByID(store, selectedID)
		if !ok {
			return "No secret selected"
		}
		return strings.Join([]string{
			"Secret Metadata",
			utils.Indent(utils.PrettyJSON(fmt.Sprintf(`{"id":"%s","name":"%s","description":"%s"}`, sec.ID, sec.Name, sec.Description)), "  "),
		}, "\n")
	case ViewTokens:
		tok, ok := tokenByID(store, selectedID)
		if !ok {
			return "No token selected"
		}
		return strings.Join([]string{
			"Token Metadata",
			utils.Indent(utils.PrettyJSON(fmt.Sprintf(`{"id":"%s","name":"%s","revoked":%t}`, tok.ID, tok.Name, tok.Revoked)), "  "),
		}, "\n")
	default:
		return "No JSON view available"
	}
}

func contextStepsContent(view ViewID, store *data.Store, selectedID string) string {
	if view != ViewRuns && view != ViewDashboard {
		return "Steps\n\nUnsupported for this view. Open Runs and select a run."
	}
	steps := stepsForRun(store, selectedID)
	if len(steps) == 0 {
		return "Steps\n\nNo step data for selected run"
	}
	lines := []string{"Steps", strings.Repeat("-", 24)}
	for _, step := range steps {
		lines = append(lines, fmt.Sprintf("%s", step.StepKey))
		lines = append(lines, fmt.Sprintf("  status   %s", strings.ToLower(step.Status)))
		lines = append(lines, fmt.Sprintf("  duration %s", formatDuration(step.Duration)))
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func contextLogsContent(view ViewID, store *data.Store, selectedID string) string {
	if view != ViewRuns && view != ViewDashboard {
		return "Logs\n\nUnsupported for this view. Open Runs and select a run."
	}
	steps := stepsForRun(store, selectedID)
	if len(steps) == 0 {
		return "Logs\n\nNo logs for selected run"
	}
	lines := []string{"Logs", strings.Repeat("-", 24)}
	for _, step := range steps {
		lines = append(lines, "", step.StepKey)
		sep := len(step.StepKey)
		if sep < 8 {
			sep = 8
		}
		lines = append(lines, "  "+strings.Repeat("-", sep))
		lines = append(lines, utils.Indent(step.Log, "  "))
	}
	return strings.Join(lines, "\n")
}

func dashboardContext(store *data.Store) string {
	active := 0
	for _, wf := range store.Workflows {
		if wf.Active {
			active++
		}
	}
	failed := 0
	running := 0
	for _, run := range store.Runs {
		switch run.Status {
		case "FAILED":
			failed++
		case "RUNNING":
			running++
		}
	}
	lines := []string{
		"Overview",
		fmt.Sprintf("Workflows: %d (active %d)", len(store.Workflows), active),
		fmt.Sprintf("Runs (24h): %d", len(store.Runs)),
		fmt.Sprintf("Failed runs: %d", failed),
		fmt.Sprintf("Running runs: %d", running),
	}
	return strings.Join(lines, "\n")
}

func workflowContext(store *data.Store, workflowID string) string {
	wf, ok := workflowByID(store, workflowID)
	if !ok {
		return "No workflow selected"
	}
	versions := versionsForWorkflow(store, workflowID)
	triggers := triggersForWorkflow(store, workflowID)
	latest := "-"
	if wf.LatestVersion > 0 {
		latest = fmt.Sprintf("v%d", wf.LatestVersion)
	}
	lines := []string{
		"Workflow: " + wf.Name,
		"Active: " + strings.ToLower(activeLabel(wf.Active)),
		"Latest version: " + latest,
		"Triggers: " + fmt.Sprintf("%d", len(triggers)),
		"Runs: " + fmt.Sprintf("%d", len(runsForWorkflow(store, workflowID))),
		"",
		"Versions:",
	}
	for _, v := range versions {
		lines = append(lines, fmt.Sprintf("- v%d (%s)", v.Version, utils.RelativeTime(time.Now(), v.CreatedAt)))
	}
	lines = append(lines, "", "Triggers:")
	for _, t := range triggers {
		lines = append(lines, fmt.Sprintf("- %s (%s)", t.Name, t.Type))
	}
	lines = append(lines, "", "Latest definition:")
	lines = append(lines, utils.Indent(utils.PrettyJSON(latestDefinition(versions)), "  "))
	return strings.Join(lines, "\n")
}

func runContext(store *data.Store, runID string) string {
	run, ok := runByID(store, runID)
	if !ok {
		return "No run selected"
	}
	steps := stepsForRun(store, runID)
	lines := []string{
		"Run: " + run.ID,
		"Status: " + run.Status,
		"Workflow: " + workflowName(store, run.WorkflowID),
		"",
		"Input:",
		utils.Indent(utils.PrettyJSON(run.InputJSON), "  "),
		"",
		"Output:",
		utils.Indent(utils.PrettyJSON(run.OutputJSON), "  "),
	}
	if run.ErrorJSON != "" {
		lines = append(lines, "", "Error:", utils.Indent(utils.PrettyJSON(run.ErrorJSON), "  "))
	}
	lines = append(lines, "", "Steps:")
	for _, step := range steps {
		lines = append(lines, fmt.Sprintf("- %s: %s", step.StepKey, step.Status))
	}
	return strings.Join(lines, "\n")
}

func triggerContext(store *data.Store, triggerID string) string {
	trg, ok := triggerByID(store, triggerID)
	if !ok {
		return "No trigger selected"
	}
	relatedEvents := eventsForTrigger(store, triggerID)
	lastRun := lastRunForWorkflow(store, trg.WorkflowID)
	lastRunID := "-"
	if lastRun != nil {
		lastRunID = lastRun.ID
	}
	lines := []string{
		"Trigger: " + trg.Name,
		"Type: " + trg.Type,
		"Active: " + strings.ToLower(activeLabel(trg.Active)),
		"Related events: " + fmt.Sprintf("%d", len(relatedEvents)),
		"Last run: " + lastRunID,
		"",
		"Config:",
		utils.Indent(utils.PrettyJSON(trg.ConfigJSON), "  "),
	}
	return strings.Join(lines, "\n")
}

func eventContext(store *data.Store, eventID string) string {
	evt, ok := eventByID(store, eventID)
	if !ok {
		return "No event selected"
	}
	linked := "-"
	if evt.RunID != nil {
		linked = *evt.RunID
	}
	lines := []string{
		"Event: " + evt.ID,
		"Trigger: " + evt.TriggerID,
		"Linked run: " + linked,
		"",
		"Payload:",
		utils.Indent(utils.PrettyJSON(evt.PayloadJSON), "  "),
		"",
		"Metadata:",
		utils.Indent(utils.PrettyJSON(evt.Metadata), "  "),
	}
	return strings.Join(lines, "\n")
}

func secretContext(store *data.Store, secretID string) string {
	sec, ok := secretByID(store, secretID)
	if !ok {
		return "No secret selected"
	}
	lines := []string{
		"Secret: " + sec.Name,
		"Description: " + sec.Description,
		"Usage: secrets." + strings.ToUpper(sec.Name),
		"",
		"Value: [REDACTED]",
	}
	return strings.Join(lines, "\n")
}

func tokenContext(store *data.Store, tokenID string) string {
	tok, ok := tokenByID(store, tokenID)
	if !ok {
		return "No token selected"
	}
	status := "active"
	if tok.Revoked {
		status = "revoked"
	}
	lines := []string{
		"Token: " + tok.Name,
		"Status: " + status,
		"Scopes:",
	}
	for _, scope := range tok.Scopes {
		lines = append(lines, "- "+scope)
	}
	lines = append(lines, "", "Revoke: press d")
	return strings.Join(lines, "\n")
}
