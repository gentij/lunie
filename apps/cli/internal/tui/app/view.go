package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/gentij/lunie/apps/cli/internal/tui/components"
)

func Render(m Model) string {
	if m.inspector.Active {
		return m.inspector.Render(m.width, m.height)
	}

	sidebar := renderSidebar(m)
	mainPanel := renderMainPanel(m)
	footer := renderFooter(m)

	row := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainPanel)
	row = clampSection(row, m.width, m.layout.MainHeight)
	base := lipgloss.JoinVertical(lipgloss.Left, row, footer)
	base = clampToViewport(base, m.width, m.height)

	output := base
	if m.showPalette {
		output = renderPaletteScreen(m)
	}
	if m.showHelp {
		output = renderHelpScreen(m)
	}
	if m.action.Active {
		modal := renderActionModal(m)
		output = renderOverlay(base, modal, m)
	}
	if m.theme.CRT {
		output = applyScanlines(output, m.width, m.theme.Scanline)
	}
	return output
}

func buildMainContent(m Model) string {
	width := max(m.layout.MainWidth-2, 1)
	header := renderMainHeader(m, width)
	body := renderMainBody(m, width)
	divider := m.styles.Divider.Render(strings.Repeat("─", width))
	sections := []string{header, divider, body}
	return strings.Join(sections, "\n")
}

func renderSidebar(m Model) string {
	width := m.layout.SidebarWidth
	height := m.layout.SidebarHeight
	innerWidth := max(width-2, 1)
	innerHeight := max(height-2, 1)
	border := m.styles.PanelBorder
	if m.focus == FocusSidebar {
		border = m.styles.PanelBorderFocus
	}

	brand := m.styles.SidebarTitle.Render("LUNE")
	workspace := m.styles.SidebarMuted.Render("WS: " + m.workspaceName)
	focus := paneFocusTag(m, FocusSidebar, "SIDEBAR")
	section := m.styles.SidebarSection.Render("Navigation")
	listView := strings.TrimRight(m.sidebar.View(), "\n")

	status := []string{
		m.styles.SidebarSection.Render("Status"),
		chip(m, "API "+m.apiStatus, m.apiStatus == "CONNECTED"),
		chip(m, refreshChip(m), false),
		chip(m, "Net "+networkProfileLabel(m.networkProfile), m.networkProfile == NetworkFlaky),
		chip(m, "Theme "+strings.Title(strings.ReplaceAll(m.themeName, "-", " ")), false),
		m.styles.Dim.Render("Font hint: " + recommendedFont(m.themeName)),
	}

	content := joinSidebarContent(
		[]string{brand, workspace, focus, "", section},
		strings.Split(listView, "\n"),
		status,
		innerHeight,
	)
	filled := applyBackgroundLayer(content, innerWidth, innerHeight, m.styles.SidebarFill)
	body := lipgloss.Place(innerWidth, innerHeight, lipgloss.Left, lipgloss.Top, filled)
	return border.Width(innerWidth).Height(innerHeight).Render(body)
}

func joinSidebarContent(top []string, middle []string, bottom []string, height int) string {
	lines := append([]string{}, top...)
	lines = append(lines, middle...)
	remaining := height - len(lines) - len(bottom)
	if remaining < 0 {
		remaining = 0
	}
	for i := 0; i < remaining; i++ {
		lines = append(lines, "")
	}
	lines = append(lines, bottom...)
	return strings.Join(lines, "\n")
}

func renderMainPanel(m Model) string {
	width := m.layout.MainWidth
	height := m.layout.MainHeight
	innerWidth := max(width-2, 1)
	innerHeight := max(height-2, 1)
	border := m.styles.PanelBorder
	if m.focus == FocusMain {
		border = m.styles.PanelBorderFocus
	}
	panelContent := strings.TrimRight(m.mainPanel.View(), "\n")
	topHeight := innerHeight
	if !m.contextCollapsed && m.layout.ContextHeight > 0 {
		topHeight -= m.layout.ContextHeight
	}
	if topHeight < 1 {
		topHeight = 1
	}
	panelContent = clampSection(panelContent, innerWidth, topHeight)
	top := applyBackgroundLayer(panelContent, innerWidth, topHeight, m.styles.PanelFill)
	contentParts := []string{top}
	if !m.contextCollapsed && m.layout.ContextHeight > 0 {
		contentParts = append(contentParts, renderContextDrawer(m, innerWidth))
	}
	composed := strings.Join(contentParts, "\n")
	composed = clampSection(composed, innerWidth, innerHeight)
	content := lipgloss.Place(innerWidth, innerHeight, lipgloss.Left, lipgloss.Top, composed)
	return border.Width(innerWidth).Height(innerHeight).Render(content)
}

func renderActionModal(m Model) string {
	hint := "enter submit  |  tab next field  |  esc cancel"
	body := ""
	switch m.action.Mode {
	case actionModalRenameWorkflow:
		body = strings.Join([]string{
			"Workflow ID: " + m.action.WorkflowID,
			"",
			m.action.Primary.View(),
		}, "\n")
	case actionModalRenameTrigger:
		body = strings.Join([]string{
			"Trigger ID: " + m.action.TriggerID,
			"",
			m.action.Primary.View(),
		}, "\n")
	case actionModalCreateTrigger:
		typeLine := "Type: " + m.action.TriggerType
		if m.action.Focus == 0 {
			typeLine = m.styles.ChipActive.Render(" " + typeLine + " ")
		}
		activeLabel := "Active: false"
		if m.action.TriggerActive {
			activeLabel = "Active: true"
		}
		if m.action.Focus == 2 {
			activeLabel = m.styles.ChipActive.Render(" " + activeLabel + " ")
		}
		if isCronTriggerType(m.action.TriggerType) {
			hint = "tab next  |  ←/→ type  |  space toggle active  |  enter submit  |  esc cancel"
			body = strings.Join([]string{
				"Workflow ID: " + m.action.WorkflowID,
				"",
				typeLine,
				m.action.Primary.View(),
				activeLabel,
				m.action.Secondary.View(),
				m.action.Tertiary.View(),
			}, "\n")
		} else {
			hint = "tab next  |  ←/→ type  |  space toggle active  |  enter submit  |  esc cancel"
			body = strings.Join([]string{
				"Workflow ID: " + m.action.WorkflowID,
				"",
				typeLine,
				m.action.Primary.View(),
				activeLabel,
				m.action.Secondary.View(),
			}, "\n")
		}
	case actionModalUpdateTrigger:
		activeLabel := "Active: false"
		if m.action.TriggerActive {
			activeLabel = "Active: true"
		}
		if m.action.Focus == 1 {
			activeLabel = m.styles.ChipActive.Render(" " + activeLabel + " ")
		}
		hint = "tab next  |  space toggle active  |  enter submit  |  esc cancel"
		if isCronTriggerType(m.action.TriggerType) {
			body = strings.Join([]string{
				"Trigger ID: " + m.action.TriggerID,
				"Type: " + m.action.TriggerType,
				"",
				m.action.Primary.View(),
				activeLabel,
				m.action.Secondary.View(),
				m.action.Tertiary.View(),
			}, "\n")
		} else {
			body = strings.Join([]string{
				"Trigger ID: " + m.action.TriggerID,
				"Type: " + m.action.TriggerType,
				"",
				m.action.Primary.View(),
				activeLabel,
				m.action.Secondary.View(),
			}, "\n")
		}
	case actionModalCreateSecret:
		hint = "tab next field  |  enter submit  |  esc cancel"
		body = strings.Join([]string{
			m.action.Description,
			"",
			m.action.Primary.View(),
			m.action.Secondary.View(),
			m.action.Tertiary.View(),
		}, "\n")
	case actionModalUpdateSecret:
		hint = "tab next field  |  enter submit  |  esc cancel"
		body = strings.Join([]string{
			"Secret ID: " + m.action.SecretID,
			m.action.Description,
			"",
			m.action.Primary.View(),
			m.action.Secondary.View(),
			m.action.Tertiary.View(),
		}, "\n")
	case actionModalCLIHandoff:
		hint = "enter/esc close"
		body = strings.Join([]string{
			m.action.Description,
			"",
			m.styles.PanelTitle.Render("Command"),
			m.action.CLICommand,
		}, "\n")
	case actionModalConfirmDelete:
		hint = "enter confirm  |  esc cancel"
		body = strings.Join([]string{
			m.action.Description,
			"Type exactly: " + m.action.ConfirmPhrase,
			"",
			m.action.Confirm.View(),
		}, "\n")
	default:
		body = "No action selected"
	}
	if m.action.ShowValidation && strings.TrimSpace(m.action.Validation) != "" {
		errorLine := lipgloss.NewStyle().Foreground(m.theme.Error).Render("Validation: " + m.action.Validation)
		body = body + "\n\n" + errorLine
	}
	return components.RenderModalWithHint(m.action.Title, body, hint, m.width, m.height, m.styles)
}

func renderPaletteScreen(m Model) string {
	innerWidth := max(m.width-2, 1)
	innerHeight := max(m.height-2, 1)
	headerTitle := m.styles.PanelTitle.Render("Command Palette")
	headerHint := m.styles.Dim.Render("Type to filter  |  / manual filter  |  enter select  |  esc close")
	divider := m.styles.Divider.Render(strings.Repeat("─", innerWidth))

	listHeight := max(innerHeight-3, 1)
	palette := m.palette
	palette.SetSize(innerWidth, listHeight)
	paletteView := strings.TrimRight(palette.View(), "\n")
	paletteView = sanitizeRenderable(paletteView)
	paletteView = lipgloss.Place(innerWidth, listHeight, lipgloss.Left, lipgloss.Top, paletteView)

	body := strings.Join([]string{headerTitle, headerHint, divider, paletteView}, "\n")
	filled := applyBackgroundLayer(body, innerWidth, innerHeight, m.styles.PanelFill)
	content := lipgloss.Place(innerWidth, innerHeight, lipgloss.Left, lipgloss.Top, filled)
	panel := m.styles.PanelBorderFocus.Width(innerWidth).Height(innerHeight).Render(content)
	return clampToViewport(panel, m.width, m.height)
}

func renderHelpScreen(m Model) string {
	innerWidth := max(m.width-2, 1)
	innerHeight := max(m.height-2, 1)
	helpModel := m.help
	helpModel.Width = innerWidth
	if innerWidth < 96 {
		helpModel.ShowAll = false
	}

	headerTitle := m.styles.PanelTitle.Render("Keyboard Help")
	headerHintText := "esc or ? close"
	if innerWidth < 96 {
		headerHintText = "esc or ? close  |  widen terminal for full help"
	}
	headerHint := m.styles.Dim.Render(headerHintText)
	divider := m.styles.Divider.Render(strings.Repeat("─", innerWidth))

	helpHeight := max(innerHeight-3, 1)
	helpView := strings.TrimRight(helpModel.View(m.keys), "\n")
	helpView = sanitizeRenderable(helpView)
	helpView = lipgloss.Place(innerWidth, helpHeight, lipgloss.Left, lipgloss.Top, helpView)

	body := strings.Join([]string{headerTitle, headerHint, divider, helpView}, "\n")
	filled := applyBackgroundLayer(body, innerWidth, innerHeight, m.styles.PanelFill)
	content := lipgloss.Place(innerWidth, innerHeight, lipgloss.Left, lipgloss.Top, filled)
	panel := m.styles.PanelBorderFocus.Width(innerWidth).Height(innerHeight).Render(content)

	return clampToViewport(panel, m.width, m.height)
}
