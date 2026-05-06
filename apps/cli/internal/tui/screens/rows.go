package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/gentij/lunie/apps/cli/internal/tui/data"
	"github.com/gentij/lunie/apps/cli/internal/tui/styles"
	"github.com/gentij/lunie/apps/cli/internal/tui/utils"
)

func recentRunRows(store *data.Store, now time.Time, limit int) ([]table.Row, []string) {
	rows := []table.Row{}
	ids := []string{}
	for i, run := range store.Runs {
		if i >= limit {
			break
		}
		rows = append(rows, table.Row{
			runLabel(store, run.ID),
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
			wf.Key,
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
			runLabel(store, run.ID),
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
			trg.Key,
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
			triggerKey(store, evt.TriggerID),
			evt.Type,
			utils.RelativeTime(now, evt.ReceivedAt),
			runLabel(store, linked),
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
