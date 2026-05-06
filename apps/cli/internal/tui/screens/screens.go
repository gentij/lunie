package screens

import (
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/gentij/lunie/apps/cli/internal/tui/data"
	"github.com/gentij/lunie/apps/cli/internal/tui/styles"
)

func BuildRowsForView(view ViewID, store *data.Store, styleSet styles.StyleSet, width int) ([]table.Column, []table.Row, []string) {
	now := time.Now()
	switch view {
	case ViewDashboard:
		columns := []table.Column{
			{Title: "Run", Width: 8},
			{Title: "Workflow", Width: 20},
			{Title: "Status", Width: 12},
			{Title: "Started", Width: 12},
		}
		rows, ids := recentRunRows(store, now, 6)
		return fitColumns(columns, width), rows, ids
	case ViewWorkflows:
		columns := []table.Column{
			{Title: "Key", Width: 18},
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
			{Title: "Run", Width: 8},
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
			{Title: "Key", Width: 18},
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
			case "Run":
				return ar.Number < br.Number
			}
		}
	case ViewWorkflows:
		aw, aok := workflowByID(store, aID)
		bw, bok := workflowByID(store, bID)
		if aok && bok {
			switch title {
			case "Key":
				return aw.Key < bw.Key
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
			case "Key":
				return at.Key < bt.Key
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
				return triggerKey(store, ae.TriggerID) < triggerKey(store, be.TriggerID)
			case "Type":
				return ae.Type < be.Type
			case "Received":
				return ae.ReceivedAt.Before(be.ReceivedAt)
			case "Linked Run":
				ar := "-"
				br := "-"
				if ae.RunID != nil {
					ar = runLabel(store, *ae.RunID)
				}
				if be.RunID != nil {
					br = runLabel(store, *be.RunID)
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
