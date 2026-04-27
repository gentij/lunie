package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/gentij/lunie/apps/cli/internal/tui/styles"
)

type StatCard struct {
	Title    string
	Value    string
	Subtitle string
}

func RenderCards(cards []StatCard, width int, styleSet styles.StyleSet) string {
	if len(cards) == 0 || width <= 0 {
		return ""
	}
	gap := 2
	cardWidth := (width - gap*(len(cards)-1)) / len(cards)
	if cardWidth < 18 {
		return renderCardsStack(cards, width, styleSet)
	}

	views := make([]string, 0, len(cards))
	for _, card := range cards {
		views = append(views, renderCard(card, cardWidth, styleSet))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, views...)
}

func renderCardsStack(cards []StatCard, width int, styleSet styles.StyleSet) string {
	rows := make([]string, 0, len(cards))
	for _, card := range cards {
		rows = append(rows, renderCard(card, width, styleSet))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func renderCard(card StatCard, width int, styleSet styles.StyleSet) string {
	if width < 10 {
		width = 10
	}
	innerWidth := width - 2
	if innerWidth < 1 {
		innerWidth = 1
	}
	line1 := styleSet.CardTitle.Render(card.Title)
	line2 := styleSet.Accent.Render(card.Value)
	line3 := styleSet.Dim.Render(card.Subtitle)
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Width(innerWidth).Render(line1),
		lipgloss.NewStyle().Width(innerWidth).Render(line2),
		lipgloss.NewStyle().Width(innerWidth).Render(line3),
	)
	box := styleSet.PanelBorder.Width(innerWidth).Height(3)
	return box.Render(content)
}
