package app

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/gentij/taskforge/apps/cli/internal/tui/data"
	"github.com/gentij/taskforge/apps/cli/internal/tui/styles"
	"github.com/gentij/taskforge/apps/cli/internal/tui/utils"
)

func BuildRowsForView(view ViewID, store *data.Store, styleSet styles.StyleSet, width int) ([]table.Column, []table.Row, []string) {
	now := time.Now()
	switch view {
	case ViewDashboard:
		columns := []table.Column{
			{Title: "Run ID", Width: 12},
			{Title: "Workflow", Width: 20},
			{Title: "Status", Width: 12},
			{Title: "Started", Width: 12},
		}
		rows, ids := recentRunRows(store, now, 6)
		return fitColumns(columns, width), rows, ids
	case ViewWorkflows:
		columns := []table.Column{
			{Title: "Name", Width: 20},
			{Title: "Active", Width: 10},
			{Title: "Latest Version", Width: 16},
			{Title: "Triggers", Width: 9},
			{Title: "Last Run", Width: 12},
			{Title: "Updated", Width: 12},
		}
		rows, ids := workflowRows(store, now)
		return fitColumns(columns, width), rows, ids
	case ViewRuns:
		columns := []table.Column{
			{Title: "Run ID", Width: 12},
			{Title: "Workflow", Width: 20},
			{Title: "Status", Width: 12},
			{Title: "Trigger", Width: 10},
			{Title: "Started", Width: 12},
			{Title: "Duration", Width: 10},
		}
		rows, ids := runRows(store, now)
		return fitColumns(columns, width), rows, ids
	case ViewTriggers:
		columns := []table.Column{
			{Title: "Name", Width: 18},
			{Title: "Type", Width: 10},
			{Title: "Workflow", Width: 18},
			{Title: "Active", Width: 10},
			{Title: "Created", Width: 12},
		}
		rows, ids := triggerRows(store, now)
		return fitColumns(columns, width), rows, ids
	case ViewEvents:
		columns := []table.Column{
			{Title: "Event ID", Width: 12},
			{Title: "Trigger", Width: 14},
			{Title: "Type", Width: 10},
			{Title: "Received", Width: 12},
			{Title: "Linked Run", Width: 12},
		}
		rows, ids := eventRows(store, styleSet, now)
		return fitColumns(columns, width), rows, ids
	case ViewSecrets:
		columns := []table.Column{
			{Title: "Name", Width: 18},
			{Title: "Description", Width: 26},
			{Title: "Created", Width: 12},
		}
		rows, ids := secretRows(store, now)
		return fitColumns(columns, width), rows, ids
	case ViewTokens:
		columns := []table.Column{
			{Title: "Name", Width: 16},
			{Title: "Scopes", Width: 24},
			{Title: "Created", Width: 12},
			{Title: "Last Used", Width: 12},
			{Title: "Status", Width: 10},
		}
		rows, ids := tokenRows(store, now)
		return fitColumns(columns, width), rows, ids
	default:
		return nil, nil, nil
	}
}

func SortRowsForView(view ViewID, store *data.Store, columns []table.Column, rows []table.Row, rowIDs []string, columnIdx int, desc bool) ([]table.Row, []string) {
	if len(rows) == 0 || len(rows) != len(rowIDs) || columnIdx < 0 || columnIdx >= len(columns) {
		return rows, rowIDs
	}
	title := strings.TrimSpace(columns[columnIdx].Title)
	indexes := make([]int, len(rows))
	for i := range rows {
		indexes[i] = i
	}
	sort.SliceStable(indexes, func(i int, j int) bool {
		a := indexes[i]
		b := indexes[j]
		less := sortLess(view, store, rowIDs[a], rowIDs[b], title, rows[a], rows[b], columnIdx)
		greater := sortLess(view, store, rowIDs[b], rowIDs[a], title, rows[b], rows[a], columnIdx)
		if !less && !greater {
			return false
		}
		if desc {
			return greater
		}
		return less
	})
	sortedRows := make([]table.Row, len(rows))
	sortedIDs := make([]string, len(rowIDs))
	for i, idx := range indexes {
		sortedRows[i] = rows[idx]
		sortedIDs[i] = rowIDs[idx]
	}
	return sortedRows, sortedIDs
}

func sortLess(view ViewID, store *data.Store, aID string, bID string, title string, aRow table.Row, bRow table.Row, col int) bool {
	switch view {
	case ViewDashboard, ViewRuns:
		ar, aok := runByID(store, aID)
		br, bok := runByID(store, bID)
		if aok && bok {
			switch title {
			case "Started":
				return ar.StartedAt.Before(br.StartedAt)
			case "Duration":
				return ar.Duration < br.Duration
			case "Status":
				return ar.Status < br.Status
			case "Workflow":
				return workflowName(store, ar.WorkflowID) < workflowName(store, br.WorkflowID)
			case "Trigger":
				return ar.TriggerType < br.TriggerType
			case "Run ID":
				return ar.ID < br.ID
			}
		}
	case ViewWorkflows:
		aw, aok := workflowByID(store, aID)
		bw, bok := workflowByID(store, bID)
		if aok && bok {
			switch title {
			case "Name":
				return aw.Name < bw.Name
			case "Active":
				return !aw.Active && bw.Active
			case "Latest Version":
				return aw.LatestVersion < bw.LatestVersion
			case "Triggers":
				return countTriggers(store, aw.ID) < countTriggers(store, bw.ID)
			case "Updated":
				return aw.UpdatedAt.Before(bw.UpdatedAt)
			}
		}
	case ViewTriggers:
		at, aok := triggerByID(store, aID)
		bt, bok := triggerByID(store, bID)
		if aok && bok {
			switch title {
			case "Name":
				return at.Name < bt.Name
			case "Type":
				return at.Type < bt.Type
			case "Workflow":
				return workflowName(store, at.WorkflowID) < workflowName(store, bt.WorkflowID)
			case "Active":
				return !at.Active && bt.Active
			case "Created":
				return at.CreatedAt.Before(bt.CreatedAt)
			}
		}
	case ViewEvents:
		ae, aok := eventByID(store, aID)
		be, bok := eventByID(store, bID)
		if aok && bok {
			switch title {
			case "Event ID":
				return ae.ID < be.ID
			case "Trigger":
				return ae.TriggerID < be.TriggerID
			case "Type":
				return ae.Type < be.Type
			case "Received":
				return ae.ReceivedAt.Before(be.ReceivedAt)
			case "Linked Run":
				ar := ""
				br := ""
				if ae.RunID != nil {
					ar = *ae.RunID
				}
				if be.RunID != nil {
					br = *be.RunID
				}
				return ar < br
			}
		}
	case ViewSecrets:
		as, aok := secretByID(store, aID)
		bs, bok := secretByID(store, bID)
		if aok && bok {
			switch title {
			case "Name":
				return as.Name < bs.Name
			case "Description":
				return as.Description < bs.Description
			case "Created":
				return as.CreatedAt.Before(bs.CreatedAt)
			}
		}
	case ViewTokens:
		at, aok := tokenByID(store, aID)
		bt, bok := tokenByID(store, bID)
		if aok && bok {
			switch title {
			case "Name":
				return at.Name < bt.Name
			case "Scopes":
				return strings.Join(at.Scopes, ",") < strings.Join(bt.Scopes, ",")
			case "Created":
				return at.CreatedAt.Before(bt.CreatedAt)
			case "Last Used":
				aTime := time.Time{}
				bTime := time.Time{}
				if at.LastUsedAt != nil {
					aTime = *at.LastUsedAt
				}
				if bt.LastUsedAt != nil {
					bTime = *bt.LastUsedAt
				}
				return aTime.Before(bTime)
			case "Status":
				return at.Revoked && !bt.Revoked
			}
		}
	}
	aVal := ""
	bVal := ""
	if col < len(aRow) {
		aVal = aRow[col]
	}
	if col < len(bRow) {
		bVal = bRow[col]
	}
	return aVal < bVal
}

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

func recentRunRows(store *data.Store, now time.Time, limit int) ([]table.Row, []string) {
	rows := []table.Row{}
	ids := []string{}
	for i, run := range store.Runs {
		if i >= limit {
			break
		}
		rows = append(rows, table.Row{
			run.ID,
			workflowName(store, run.WorkflowID),
			normalizeStatus(run.Status),
			utils.RelativeTime(now, run.StartedAt),
		})
		ids = append(ids, run.ID)
	}
	return rows, ids
}

func workflowRows(store *data.Store, now time.Time) ([]table.Row, []string) {
	rows := []table.Row{}
	ids := []string{}
	for _, wf := range store.Workflows {
		rows = append(rows, table.Row{
			wf.Name,
			activeLabel(wf.Active),
			fmt.Sprintf("v%d", wf.LatestVersion),
			fmt.Sprintf("%d", countTriggers(store, wf.ID)),
			lastRunStatus(store, wf.ID),
			utils.RelativeTime(now, wf.UpdatedAt),
		})
		ids = append(ids, wf.ID)
	}
	return rows, ids
}

func runRows(store *data.Store, now time.Time) ([]table.Row, []string) {
	rows := []table.Row{}
	ids := []string{}
	for _, run := range store.Runs {
		rows = append(rows, table.Row{
			run.ID,
			workflowName(store, run.WorkflowID),
			normalizeStatus(run.Status),
			run.TriggerType,
			utils.RelativeTime(now, run.StartedAt),
			formatDuration(run.Duration),
		})
		ids = append(ids, run.ID)
	}
	return rows, ids
}

func triggerRows(store *data.Store, now time.Time) ([]table.Row, []string) {
	rows := []table.Row{}
	ids := []string{}
	for _, trg := range store.Triggers {
		rows = append(rows, table.Row{
			trg.Name,
			trg.Type,
			workflowName(store, trg.WorkflowID),
			activeLabel(trg.Active),
			utils.RelativeTime(now, trg.CreatedAt),
		})
		ids = append(ids, trg.ID)
	}
	return rows, ids
}

func eventRows(store *data.Store, styleSet styles.StyleSet, now time.Time) ([]table.Row, []string) {
	rows := []table.Row{}
	ids := []string{}
	for _, evt := range store.Events {
		linked := "-"
		if evt.RunID != nil {
			linked = *evt.RunID
		}
		rows = append(rows, table.Row{
			evt.ID,
			evt.TriggerID,
			evt.Type,
			utils.RelativeTime(now, evt.ReceivedAt),
			linked,
		})
		ids = append(ids, evt.ID)
	}
	return rows, ids
}

func secretRows(store *data.Store, now time.Time) ([]table.Row, []string) {
	rows := []table.Row{}
	ids := []string{}
	for _, sec := range store.Secrets {
		rows = append(rows, table.Row{
			sec.Name,
			sec.Description,
			utils.RelativeTime(now, sec.CreatedAt),
		})
		ids = append(ids, sec.ID)
	}
	return rows, ids
}

func tokenRows(store *data.Store, now time.Time) ([]table.Row, []string) {
	rows := []table.Row{}
	ids := []string{}
	for _, tok := range store.ApiTokens {
		lastUsed := "-"
		if tok.LastUsedAt != nil {
			lastUsed = utils.RelativeTime(now, *tok.LastUsedAt)
		}
		status := "ACTIVE"
		if tok.Revoked {
			status = "REVOKED"
		}
		rows = append(rows, table.Row{
			tok.Name,
			strings.Join(tok.Scopes, ","),
			utils.RelativeTime(now, tok.CreatedAt),
			lastUsed,
			status,
		})
		ids = append(ids, tok.ID)
	}
	return rows, ids
}

func statusBadge(styleSet styles.StyleSet, status string) string {
	status = strings.ToUpper(status)
	switch status {
	case "SUCCEEDED":
		return styles.Badge(styleSet.BadgeSuccess, status)
	case "FAILED":
		return styles.Badge(styleSet.BadgeFailed, status)
	case "RUNNING":
		return styles.Badge(styleSet.BadgeRunning, status)
	case "QUEUED":
		return styles.Badge(styleSet.BadgeQueued, status)
	case "REVOKED":
		return styles.Badge(styleSet.BadgeFailed, status)
	case "ACTIVE":
		return styles.Badge(styleSet.BadgeSuccess, status)
	case "INACTIVE":
		return styles.Badge(styleSet.BadgeMuted, status)
	default:
		return styles.Badge(styleSet.BadgeMuted, status)
	}
}

func activeBadge(styleSet styles.StyleSet, active bool) string {
	if active {
		return statusBadge(styleSet, "ACTIVE")
	}
	return statusBadge(styleSet, "INACTIVE")
}

func fitColumns(columns []table.Column, width int) []table.Column {
	if width <= 0 {
		return columns
	}
	available := width - (len(columns) - 1)
	if available < len(columns) {
		available = len(columns)
	}
	total := 0
	for _, col := range columns {
		total += col.Width
	}
	if total <= available {
		return columns
	}
	overflow := total - available
	minWidth := 6
	for overflow > 0 {
		reduced := false
		for i := len(columns) - 1; i >= 0 && overflow > 0; i-- {
			if columns[i].Width > minWidth {
				columns[i].Width--
				overflow--
				reduced = true
			}
		}
		if !reduced {
			break
		}
	}
	return columns
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

func workflowName(store *data.Store, workflowID string) string {
	if wf, ok := workflowByID(store, workflowID); ok {
		return wf.Name
	}
	return workflowID
}

func workflowByID(store *data.Store, workflowID string) (data.Workflow, bool) {
	for _, wf := range store.Workflows {
		if wf.ID == workflowID {
			return wf, true
		}
	}
	return data.Workflow{}, false
}

func runByID(store *data.Store, runID string) (data.WorkflowRun, bool) {
	for _, run := range store.Runs {
		if run.ID == runID {
			return run, true
		}
	}
	return data.WorkflowRun{}, false
}

func triggerByID(store *data.Store, triggerID string) (data.Trigger, bool) {
	for _, trg := range store.Triggers {
		if trg.ID == triggerID {
			return trg, true
		}
	}
	return data.Trigger{}, false
}

func eventByID(store *data.Store, eventID string) (data.Event, bool) {
	for _, evt := range store.Events {
		if evt.ID == eventID {
			return evt, true
		}
	}
	return data.Event{}, false
}

func secretByID(store *data.Store, secretID string) (data.Secret, bool) {
	for _, sec := range store.Secrets {
		if sec.ID == secretID {
			return sec, true
		}
	}
	return data.Secret{}, false
}

func tokenByID(store *data.Store, tokenID string) (data.ApiToken, bool) {
	for _, tok := range store.ApiTokens {
		if tok.ID == tokenID {
			return tok, true
		}
	}
	return data.ApiToken{}, false
}

func countTriggers(store *data.Store, workflowID string) int {
	count := 0
	for _, trg := range store.Triggers {
		if trg.WorkflowID == workflowID {
			count++
		}
	}
	return count
}

func runsForWorkflow(store *data.Store, workflowID string) []data.WorkflowRun {
	items := []data.WorkflowRun{}
	for _, run := range store.Runs {
		if run.WorkflowID == workflowID {
			items = append(items, run)
		}
	}
	return items
}

func lastRunForWorkflow(store *data.Store, workflowID string) *data.WorkflowRun {
	for _, run := range store.Runs {
		if run.WorkflowID == workflowID {
			return &run
		}
	}
	return nil
}

func lastRunStatus(store *data.Store, workflowID string) string {
	if run := lastRunForWorkflow(store, workflowID); run != nil {
		return normalizeStatus(run.Status)
	}
	return "QUEUED"
}

func versionsForWorkflow(store *data.Store, workflowID string) []data.WorkflowVersion {
	items := []data.WorkflowVersion{}
	for _, v := range store.WorkflowVersions {
		if v.WorkflowID == workflowID {
			items = append(items, v)
		}
	}
	return items
}

func triggersForWorkflow(store *data.Store, workflowID string) []data.Trigger {
	items := []data.Trigger{}
	for _, t := range store.Triggers {
		if t.WorkflowID == workflowID {
			items = append(items, t)
		}
	}
	return items
}

func eventsForTrigger(store *data.Store, triggerID string) []data.Event {
	items := []data.Event{}
	for _, evt := range store.Events {
		if evt.TriggerID == triggerID {
			items = append(items, evt)
		}
	}
	return items
}

func stepsForRun(store *data.Store, runID string) []data.StepRun {
	items := []data.StepRun{}
	for _, step := range store.StepRuns {
		if step.RunID == runID {
			items = append(items, step)
		}
	}
	return items
}

func latestDefinition(versions []data.WorkflowVersion) string {
	if len(versions) == 0 {
		return "{}"
	}
	latest := versions[0]
	for _, v := range versions {
		if v.Version > latest.Version {
			latest = v
		}
	}
	return latest.DefinitionJSON
}

func activeLabel(active bool) string {
	if active {
		return "ACTIVE"
	}
	return "INACTIVE"
}

func normalizeStatus(status string) string {
	status = strings.TrimSpace(strings.ToUpper(status))
	if status == "" {
		return "UNKNOWN"
	}
	return status
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "-"
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
}
