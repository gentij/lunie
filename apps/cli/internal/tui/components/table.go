package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/x/ansi"
	"github.com/gentij/lunie/apps/cli/internal/tui/styles"
)

func NewTable(columns []table.Column, rows []table.Row, width int, height int, styleSet styles.StyleSet) table.Model {
	model := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)

	model.SetWidth(width)
	model.SetStyles(TableStyles(styleSet))

	return model
}

func TableStyles(styleSet styles.StyleSet) table.Styles {
	style := table.DefaultStyles()
	style.Header = styleSet.TableHeader
	style.Cell = styleSet.TableCell
	style.Selected = styleSet.TableCell
	return style
}

func StyleRows(rows []table.Row, columns []table.Column, selected int, styleSet styles.StyleSet) []table.Row {
	styled := make([]table.Row, len(rows))
	widths := make([]int, len(columns))
	for i, col := range columns {
		widths[i] = max(col.Width, 1)
	}
	for i, row := range rows {
		styledRow := make(table.Row, len(row))
		for j, cell := range row {
			width := 0
			if j < len(widths) {
				width = widths[j]
			}
			effectiveWidth := width
			if j > 0 {
				effectiveWidth = max(width-1, 1)
			}
			if j == 0 {
				marker := "  "
				if i == selected {
					marker = "> "
				}
				cell = marker + padCell(cell, max(effectiveWidth-len(marker), 1))
			} else {
				cell = padCell(cell, effectiveWidth)
			}
			if j > 0 {
				cell = " " + cell
			}
			styledRow[j] = cell
		}
		styled[i] = styledRow
	}
	return styled
}

func padCell(value string, width int) string {
	if width <= 0 {
		return value
	}
	tail := ""
	if ansi.StringWidth(value) > width {
		if width >= 6 {
			tail = "..."
		} else if width >= 2 {
			tail = "."
		}
	}
	value = ansi.Truncate(value, width, tail)
	pad := width - ansi.StringWidth(value)
	if pad < 0 {
		pad = 0
	}
	return value + strings.Repeat(" ", pad)
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
