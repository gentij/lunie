package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/gentij/taskforge/apps/cli/internal/tui/styles"
)

func RenderModal(title string, body string, width int, height int, styleSet styles.StyleSet) string {
	return RenderModalWithHint(
		title,
		body,
		"Type to filter  |  / manual filter  |  enter select  |  esc close",
		width,
		height,
		styleSet,
	)
}

func RenderModalWithHint(title string, body string, hint string, width int, height int, styleSet styles.StyleSet) string {
	contentWidth := min(max(width-16, 36), 72)
	box := styleSet.PanelBorder.Copy().
		Width(contentWidth).
		Padding(1, 2).
		BorderForeground(styleSet.BorderColor)
	hintLine := styleSet.Dim.Render(hint)
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		styleSet.PanelTitle.Render(title),
		hintLine,
		"",
		body,
	)
	return box.Render(content)
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
