package app

import (
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/term"
)

func apiStatus(tokenSet bool) string {
	if tokenSet {
		return "CONNECTED"
	}
	return "OFFLINE"
}

func filterRows(rows []table.Row, rowIDs []string, query string) ([]table.Row, []string) {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return rows, rowIDs
	}
	filtered := make([]table.Row, 0, len(rows))
	filteredIDs := make([]string, 0, len(rows))
	for i, row := range rows {
		joined := strings.ToLower(strings.Join(row, " "))
		if strings.Contains(joined, query) {
			filtered = append(filtered, row)
			if i < len(rowIDs) {
				filteredIDs = append(filteredIDs, rowIDs[i])
			}
		}
	}
	return filtered, filteredIDs
}

func defaultSortConfig(columns []table.Column) SortConfig {
	if len(columns) == 0 {
		return SortConfig{Column: -1, Desc: true}
	}
	preferred := []string{"Started", "Updated", "Received", "Created", "Last Used"}
	for _, target := range preferred {
		for i, col := range columns {
			if strings.EqualFold(strings.TrimSpace(col.Title), target) {
				return SortConfig{Column: i, Desc: true}
			}
		}
	}
	return SortConfig{Column: 0, Desc: true}
}

func columnsWithSortIndicators(columns []table.Column, sortColumn int, desc bool) []table.Column {
	decorated := make([]table.Column, len(columns))
	copy(decorated, columns)
	if sortColumn < 0 || sortColumn >= len(decorated) {
		return decorated
	}
	arrow := " ▲"
	if desc {
		arrow = " ▼"
	}
	decorated[sortColumn].Title = strings.TrimSpace(decorated[sortColumn].Title) + arrow
	return decorated
}

func (m *Model) cycleSortColumn() {
	if len(m.columns) == 0 {
		return
	}
	if m.sortColumn < 0 {
		m.sortColumn = 0
	} else {
		m.sortColumn = (m.sortColumn + 1) % len(m.columns)
	}
	cfg := m.sortByView[m.view]
	cfg.Column = m.sortColumn
	cfg.Desc = m.sortDesc
	m.sortByView[m.view] = cfg
	m.refreshView()
}

func (m *Model) toggleSortDirection() {
	m.sortDesc = !m.sortDesc
	cfg := m.sortByView[m.view]
	cfg.Column = m.sortColumn
	cfg.Desc = m.sortDesc
	m.sortByView[m.view] = cfg
	m.refreshView()
}

func truncateRows(rows []table.Row, columns []table.Column) []table.Row {
	if len(rows) == 0 {
		return rows
	}
	widths := make([]int, 0, len(columns))
	for _, col := range columns {
		widths = append(widths, max(col.Width, 1))
	}
	for i := range rows {
		for c := 0; c < len(rows[i]) && c < len(widths); c++ {
			rows[i][c] = ansi.Truncate(rows[i][c], widths[c], "")
		}
	}
	return rows
}

func initialSize() (int, int) {
	width, height := 80, 24
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err == nil && w > 0 && h > 0 {
		width = w
		height = h
	}
	return width, height
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
