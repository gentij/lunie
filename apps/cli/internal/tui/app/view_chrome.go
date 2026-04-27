package app

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/gentij/lunie/apps/cli/internal/tui/components"
)

func renderMainHeader(m Model, width int) string {
	section := viewTitle(m.view)
	summary := viewSummary(m)
	left := m.styles.PanelTitle.Render(section) + "  " + themedDivider(m) + "  " + summary
	filter := ""
	if m.searchQuery != "" {
		filter = "Filter: " + m.searchQuery
	}
	line1 := joinLeftRight(left, filter, width)
	chips := []string{
		chip(m, "API "+m.apiStatus, m.apiStatus == "CONNECTED"),
		chip(m, refreshChip(m), false),
		chip(m, "Net "+networkProfileLabel(m.networkProfile), m.networkProfile == NetworkFlaky),
	}
	if state := surfaceStateLabel(m.mainState); state != "" {
		chips = append(chips, chip(m, state, m.mainState == SurfaceError || m.mainState == SurfaceStale))
	}
	chips = append(chips, paneFocusTag(m, FocusMain, "MAIN"))
	if m.searchQuery != "" {
		chips = append(chips, chip(m, "Filter", true))
	}
	line2 := strings.Join(chips, " ")

	line1 = ansi.Truncate(line1, width, "")
	line2 = ansi.Truncate(line2, width, "")
	line1 = m.styles.Header.Width(width).Render(line1)
	line2 = m.styles.Header.Width(width).Render(line2)
	return strings.Join([]string{line1, line2}, "\n")
}

func renderMainBody(m Model, width int) string {
	parts := []string{}
	if m.view == ViewDashboard {
		parts = append(parts, renderDashboard(m, width))
	} else {
		label := m.styles.SidebarSection.Render("List")
		meta := renderTableMeta(m, width)
		if m.mainState == SurfaceLoading {
			parts = append(parts, label, meta, renderLoadingState(m, width))
		} else if m.mainState == SurfaceError {
			parts = append(parts, label, meta, renderErrorState(m, width, "Unable to load data", "Press ctrl+r to retry"))
		} else if m.mainState == SurfaceStale && len(m.filteredRows) == 0 {
			parts = append(parts, label, meta, renderErrorState(m, width, "Showing stale data", "Press ctrl+r to refresh"))
		} else if len(m.filteredRows) == 0 {
			parts = append(parts, label, meta, renderEmptyState(m, width))
		} else {
			parts = append(parts, label, meta, strings.TrimRight(m.table.View(), "\n"))
		}
	}
	return strings.Join(parts, "\n")
}

func renderDashboard(m Model, width int) string {
	if m.layout.DashboardCardsHeight == 0 {
		if m.mainState == SurfaceLoading {
			return renderLoadingState(m, width)
		}
		if m.mainState == SurfaceError {
			return renderErrorState(m, width, "Unable to load dashboard", "Press ctrl+r to retry")
		}
		if len(m.filteredRows) == 0 {
			return renderEmptyState(m, width)
		}
		return strings.TrimRight(m.table.View(), "\n")
	}
	active := 0
	for _, wf := range m.store.Workflows {
		if wf.Active {
			active++
		}
	}
	failed := 0
	for _, run := range m.store.Runs {
		if run.Status == "FAILED" {
			failed++
		}
	}
	cards := components.RenderCards([]components.StatCard{
		{Title: "Workflows", Value: itoa(active) + " active", Subtitle: itoa(len(m.store.Workflows)) + " total"},
		{Title: "Runs 24h", Value: itoa(len(m.store.Runs)) + " total", Subtitle: "recent activity"},
		{Title: "Failures", Value: itoa(failed), Subtitle: "needs attention"},
	}, width, m.styles)
	label := m.styles.SidebarSection.Render("Recent Runs")
	meta := renderTableMeta(m, width)
	if m.mainState == SurfaceLoading {
		return strings.Join([]string{cards, label, meta, renderLoadingState(m, width)}, "\n")
	}
	if m.mainState == SurfaceError {
		return strings.Join([]string{cards, label, meta, renderErrorState(m, width, "Unable to load recent runs", "Press ctrl+r to retry")}, "\n")
	}
	if len(m.filteredRows) == 0 {
		return strings.Join([]string{cards, label, meta, renderEmptyState(m, width)}, "\n")
	}
	return strings.Join([]string{cards, label, meta, strings.TrimRight(m.table.View(), "\n")}, "\n")
}

func renderContextDrawer(m Model, width int) string {
	innerWidth := max(width-2, 1)
	tabs := renderTabs(m, []string{"Overview", "JSON", "Steps", "Logs"}, contextTabLabel(m.contextTab))
	meta := renderContextMeta(m, innerWidth)
	content := strings.TrimRight(m.contextViewport.View(), "\n")
	if m.contextState == SurfaceLoading {
		content = renderContextLoading(m, innerWidth)
	}
	if m.contextState == SurfaceError {
		content = renderContextError(m, innerWidth, "Context unavailable", "Press ctrl+r to retry")
	}
	if m.contextState == SurfaceStale {
		content = renderContextError(m, innerWidth, "Context may be stale", "Press ctrl+r to refresh")
	}
	body := lipgloss.Place(innerWidth, max(m.layout.ContextHeight-4, 1), lipgloss.Left, lipgloss.Top, content)
	inner := lipgloss.JoinVertical(lipgloss.Left, tabs, meta, body)
	inner = applyBackgroundLayer(inner, innerWidth, max(m.layout.ContextHeight-2, 1), m.styles.ContextFill)
	box := m.styles.PanelBorder.Width(innerWidth).Height(max(m.layout.ContextHeight-2, 1))
	if m.focus == FocusContext {
		box = m.styles.PanelBorderFocus.Width(innerWidth).Height(max(m.layout.ContextHeight-2, 1))
	}
	return box.Render(inner)
}

func renderTabs(m Model, tabs []string, active string) string {
	items := make([]string, 0, len(tabs))
	for _, tab := range tabs {
		if tab == active {
			items = append(items, m.styles.TabActive.Render(" "+tab+" "))
		} else {
			items = append(items, m.styles.TabInactive.Render(" "+tab+" "))
		}
	}
	return strings.Join(items, " ")
}

func contextTabLabel(tab ContextTab) string {
	switch tab {
	case ContextTabJSON:
		return "JSON"
	case ContextTabSteps:
		return "Steps"
	case ContextTabLogs:
		return "Logs"
	default:
		return "Overview"
	}
}

func renderFooter(m Model) string {
	left := renderFooterHints(m)
	if m.toast.Active {
		left = toastLabel(m) + "  " + themedDivider(m) + "  " + left
	}
	right := m.paginator.View()
	line := joinLeftRight(left, right, m.width)
	content := m.styles.Footer.Width(m.width).Render(line)
	return clampSection(content, m.width, m.layout.FooterHeight)
}

func renderTableMeta(m Model, width int) string {
	rows := len(m.baseRows)
	filtered := len(m.filteredRows)
	page := m.paginator.Page + 1
	totalPages := m.paginator.TotalPages
	if totalPages < 1 {
		totalPages = 1
	}
	sortLabel := "none"
	if m.sortColumn >= 0 && m.sortColumn < len(m.columns) {
		dir := "asc"
		if m.sortDesc {
			dir = "desc"
		}
		sortLabel = strings.ToLower(strings.TrimSpace(m.columns[m.sortColumn].Title)) + " " + dir
	}
	text := "rows " + itoa(rows) + "  " + themedDivider(m) + "  filtered " + itoa(filtered) + "  " + themedDivider(m) + "  page " + itoa(page) + "/" + itoa(totalPages) + "  " + themedDivider(m) + "  sort " + sortLabel
	if supportsStatusScope(m.view) {
		scope := strings.ToLower(statusScopeLabel(m.currentStatusScope(m.view)))
		text += "  " + themedDivider(m) + "  status " + scope
	}
	text = ansi.Truncate(text, width, "")
	return m.styles.Dim.Width(width).Render(text)
}

func renderContextMeta(m Model, width int) string {
	selected := m.selectedRowID()
	if selected == "" {
		selected = "none"
	}
	left := paneFocusTag(m, FocusContext, "CONTEXT") + " selected " + selected
	right := "tabs [ ] 1-4  " + themedDivider(m) + "  scroll j/k pgup/pgdn"
	line := joinLeftRight(left, right, width)
	return m.styles.Dim.Width(width).Render(line)
}

func renderFooterHints(m Model) string {
	hint := "focus: " + focusName(m.focus) + "  " + themedDivider(m) + "  ? help"
	if m.focus == FocusSidebar {
		hint = "focus: sidebar  " + themedDivider(m) + "  ↑/↓ select  " + themedDivider(m) + "  enter/right focus main  " + themedDivider(m) + "  tab next pane"
	} else if m.focus == FocusMain {
		hint = "focus: main  " + themedDivider(m) + "  ↑/↓ select  " + themedDivider(m) + "  s col  S dir  " + themedDivider(m) + "  g/G top/bottom  " + themedDivider(m) + "  tab next pane"
		if m.view == ViewWorkflows {
			hint += "  " + themedDivider(m) + "  r run  e toggle  n rename  c trigger  d archive  f filter"
		}
		if m.view == ViewTriggers {
			hint += "  " + themedDivider(m) + "  e toggle  n update  c create  d archive  f filter"
		}
		if m.view == ViewSecrets {
			hint += "  " + themedDivider(m) + "  c create  n update  d delete"
		}
	} else {
		hint = "focus: context  " + themedDivider(m) + "  j/k scroll  " + themedDivider(m) + "  [/] or 1-4 tabs  " + themedDivider(m) + "  ctrl+f search"
	}
	if m.canRetry() {
		hint += "  " + themedDivider(m) + "  ctrl+r retry"
	}
	return hint
}

func paneFocusTag(m Model, pane FocusPane, label string) string {
	if m.focus == pane {
		return chip(m, ">> "+label, true)
	}
	return m.styles.Dim.Render("   " + strings.ToLower(label))
}

func focusName(pane FocusPane) string {
	switch pane {
	case FocusSidebar:
		return "sidebar"
	case FocusContext:
		return "context"
	default:
		return "main"
	}
}

func themedDivider(m Model) string {
	if m.themeName == "fallout" || m.themeName == "retro-amber" {
		return m.styles.Dim.Render("::")
	}
	return m.styles.Dim.Render("•")
}

func recommendedFont(themeName string) string {
	switch themeName {
	case "simple-dark":
		return "Cascadia Mono"
	case "simple-light":
		return "Cascadia Mono"
	case "dracula":
		return "Fira Code"
	case "one-dark-pro":
		return "JetBrains Mono"
	case "rose-pine-moon":
		return "Iosevka"
	case "solarized-dark":
		return "Source Code Pro"
	case "nord":
		return "Recursive Mono"
	case "gruvbox-dark":
		return "IBM Plex Mono"
	case "solarized-light":
		return "Source Code Pro"
	case "fallout":
		return "IBM Plex Mono"
	case "retro-amber":
		return "Berkeley Mono"
	case "tokyo-night":
		return "JetBrains Mono"
	case "catppuccin":
		return "Iosevka"
	default:
		return "Cascadia Mono"
	}
}

func chip(m Model, text string, active bool) string {
	if active {
		return m.styles.ChipActive.Render(" " + text + " ")
	}
	return m.styles.Chip.Render(" " + text + " ")
}

func refreshLabel(m Model) string {
	if m.autoRefresh {
		return "2s"
	}
	return "OFF"
}

func refreshChip(m Model) string {
	if !m.autoRefresh {
		return "⟳ " + refreshLabel(m)
	}
	if m.pulseOn {
		return "↻ " + refreshLabel(m)
	}
	return "⟳ " + refreshLabel(m)
}

func renderEmptyState(m Model, width int) string {
	title := "No items to display"
	if m.searchQuery != "" {
		title = "No matches for \"" + m.searchQuery + "\""
	}
	hint := "Try a different filter, or open the command palette with ctrl+k"
	line1 := m.styles.Dim.Bold(true).Render(title)
	line2 := m.styles.Dim.Render(hint)
	inner := strings.Join([]string{line1, "", line2}, "\n")
	return applyBackgroundLayer(inner, width, 6, m.styles.ContextFill)
}

func renderLoadingState(m Model, width int) string {
	lines := []string{
		m.styles.Dim.Render("Loading data..."),
		"",
		m.styles.Dim.Render("··························"),
		m.styles.Dim.Render("··················"),
		m.styles.Dim.Render("·····················"),
	}
	return applyBackgroundLayer(strings.Join(lines, "\n"), width, 6, m.styles.ContextFill)
}

func renderErrorState(m Model, width int, title string, hint string) string {
	line1 := m.styles.ChipActive.Render(" " + title + " ")
	line2 := m.styles.Dim.Render(hint)
	inner := strings.Join([]string{line1, "", line2}, "\n")
	return applyBackgroundLayer(inner, width, 6, m.styles.ContextFill)
}

func renderContextLoading(m Model, width int) string {
	lines := []string{m.styles.Dim.Render("Loading context..."), "", m.styles.Dim.Render("················")}
	return clampSection(strings.Join(lines, "\n"), width, max(m.layout.ContextHeight-4, 1))
}

func renderContextError(m Model, width int, title string, hint string) string {
	lines := []string{m.styles.ChipActive.Render(" " + title + " "), "", m.styles.Dim.Render(hint)}
	return clampSection(strings.Join(lines, "\n"), width, max(m.layout.ContextHeight-4, 1))
}

func surfaceStateLabel(state SurfaceState) string {
	switch state {
	case SurfaceLoading:
		return "Loading"
	case SurfaceRefreshing:
		return "Refreshing"
	case SurfaceError:
		return "Error"
	case SurfaceStale:
		return "Stale"
	case SurfaceEmpty:
		return "Empty"
	default:
		return ""
	}
}

func toastLabel(m Model) string {
	if !m.toast.Active {
		return ""
	}
	prefix := "i"
	if m.toast.Level == ToastSuccess {
		prefix = "+"
	}
	if m.toast.Level == ToastWarn {
		prefix = "!"
	}
	if m.toast.Level == ToastError {
		prefix = "x"
	}
	return m.styles.ChipActive.Render(" " + prefix + " " + m.toast.Message + " ")
}

func joinLeftRight(left string, right string, width int) string {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if right == "" {
		return ansi.Truncate(left, width, "")
	}
	space := width - lipgloss.Width(left) - lipgloss.Width(right)
	if space < 1 {
		space = 1
	}
	line := left + strings.Repeat(" ", space) + right
	return ansi.Truncate(line, width, "")
}

func itoa(value int) string {
	return strconv.Itoa(value)
}

func viewTitle(view ViewID) string {
	switch view {
	case ViewDashboard:
		return "Dashboard"
	case ViewWorkflows:
		return "Workflows"
	case ViewRuns:
		return "Runs"
	case ViewTriggers:
		return "Triggers"
	case ViewEvents:
		return "Events"
	case ViewSecrets:
		return "Secrets"
	case ViewTokens:
		return "API Tokens"
	default:
		return "Overview"
	}
}

func viewSummary(m Model) string {
	switch m.view {
	case ViewDashboard:
		return "Overview" + "  " + m.styles.Dim.Render("•") + "  Recent runs " + itoa(min(len(m.store.Runs), 6))
	case ViewWorkflows:
		active := 0
		for _, wf := range m.store.Workflows {
			if wf.Active {
				active++
			}
		}
		return "Workflows " + itoa(len(m.store.Workflows)) + " total" + "  " + m.styles.Dim.Render("•") + "  " + itoa(active) + " active"
	case ViewRuns:
		failed := 0
		running := 0
		queued := 0
		for _, run := range m.store.Runs {
			switch run.Status {
			case "FAILED":
				failed++
			case "RUNNING":
				running++
			case "QUEUED":
				queued++
			}
		}
		return "Runs " + itoa(len(m.store.Runs)) + " total" + "  " + m.styles.Dim.Render("•") + "  " + itoa(running) + " running" + "  " + m.styles.Dim.Render("•") + "  " + itoa(failed) + " failed" + "  " + m.styles.Dim.Render("•") + "  " + itoa(queued) + " queued"
	case ViewTriggers:
		active := 0
		for _, trg := range m.store.Triggers {
			if trg.Active {
				active++
			}
		}
		return "Triggers " + itoa(len(m.store.Triggers)) + " total" + "  " + m.styles.Dim.Render("•") + "  " + itoa(active) + " active"
	case ViewEvents:
		return "Events " + itoa(len(m.store.Events)) + " total"
	case ViewSecrets:
		return "Secrets " + itoa(len(m.store.Secrets)) + " total"
	case ViewTokens:
		revoked := 0
		for _, tok := range m.store.ApiTokens {
			if tok.Revoked {
				revoked++
			}
		}
		return "Tokens " + itoa(len(m.store.ApiTokens)) + " total" + "  " + m.styles.Dim.Render("•") + "  " + itoa(revoked) + " revoked"
	default:
		return ""
	}
}
