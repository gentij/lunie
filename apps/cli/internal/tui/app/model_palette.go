package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gentij/taskforge/apps/cli/internal/tui/styles"
)

func (m Model) paletteState() paletteBuildState {
	scope := m.currentStatusScope(m.view)
	hasScope := supportsStatusScope(m.view) && scope != statusScopeAll
	return paletteBuildState{
		View:         m.view,
		HasSelection: m.selectedRowID() != "",
		HasFilter:    strings.TrimSpace(m.searchQuery) != "" || hasScope,
		HasScope:     supportsStatusScope(m.view),
		Scope:        scope,
		AutoRefresh:  m.autoRefresh,
		Profile:      m.networkProfile,
		HasRecent:    len(m.paletteRecent) > 0,
	}
}

func supportsStatusScope(view ViewID) bool {
	return view == ViewWorkflows || view == ViewTriggers
}

func (m Model) currentStatusScope(view ViewID) statusScope {
	if !supportsStatusScope(view) {
		return statusScopeAll
	}
	scope, ok := m.statusScopeByView[view]
	if !ok {
		return statusScopeAll
	}
	return scope
}

func statusScopeLabel(scope statusScope) string {
	switch scope {
	case statusScopeActive:
		return "ACTIVE"
	case statusScopeInactive:
		return "INACTIVE"
	default:
		return "ALL"
	}
}

func statusScopeFromValue(value string) statusScope {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "active":
		return statusScopeActive
	case "inactive":
		return statusScopeInactive
	default:
		return statusScopeAll
	}
}

func (m *Model) rememberPaletteAction(action paletteAction) {
	if action.Kind == paletteNoop {
		return
	}
	maxRecent := 5
	recent := []paletteAction{action}
	for _, existing := range m.paletteRecent {
		if paletteActionKey(existing) == paletteActionKey(action) {
			continue
		}
		recent = append(recent, existing)
		if len(recent) >= maxRecent {
			break
		}
	}
	m.paletteRecent = recent
}

func paletteActionKey(action paletteAction) string {
	return fmt.Sprintf("%d:%s:%d:%s", action.Kind, action.View, action.Profile, action.Value)
}

func paletteItemFromAction(action paletteAction, state paletteBuildState) paletteItem {
	item := paletteItem{Label: "Action", Detail: "Recent", Action: action, Enabled: true}
	switch action.Kind {
	case paletteGoToView:
		switch action.View {
		case ViewDashboard:
			item.Label = "Go: Dashboard"
		case ViewWorkflows:
			item.Label = "Go: Workflows"
		case ViewRuns:
			item.Label = "Go: Runs"
		case ViewTriggers:
			item.Label = "Go: Triggers"
		case ViewEvents:
			item.Label = "Go: Events"
		case ViewSecrets:
			item.Label = "Go: Secrets"
		case ViewTokens:
			item.Label = "Go: API Tokens"
		}
	case paletteRunWorkflow:
		item.Label = "Action: Run selected workflow"
		if !(state.View == ViewWorkflows && state.HasSelection) {
			item.Enabled = false
			item.Detail = "Unavailable: select workflow row"
			item.DisabledReason = "Select a workflow row in Workflows first"
		}
	case paletteRenameWorkflow:
		item.Label = "Action: Rename selected workflow"
		if !(state.View == ViewWorkflows && state.HasSelection) {
			item.Enabled = false
			item.Detail = "Unavailable: select workflow row"
			item.DisabledReason = "Select a workflow row in Workflows first"
		}
	case paletteCreateTrigger:
		item.Label = "Action: Create trigger"
		if !((state.View == ViewWorkflows || state.View == ViewTriggers) && state.HasSelection) {
			item.Enabled = false
			item.Detail = "Unavailable: select workflow or trigger row"
			item.DisabledReason = "Select a row in Workflows or Triggers first"
		}
	case paletteRenameTrigger:
		item.Label = "Action: Update selected trigger"
		if !(state.View == ViewTriggers && state.HasSelection) {
			item.Enabled = false
			item.Detail = "Unavailable: select trigger row"
			item.DisabledReason = "Select a trigger row in Triggers first"
		}
	case paletteToggleTrigger:
		item.Label = "Action: Toggle selected trigger"
		if !(state.View == ViewTriggers && state.HasSelection) {
			item.Enabled = false
			item.Detail = "Unavailable: select trigger row"
			item.DisabledReason = "Select a trigger row in Triggers first"
		}
	case paletteDeleteWorkflow:
		item.Label = "Action: Archive selected workflow"
		if !(state.View == ViewWorkflows && state.HasSelection) {
			item.Enabled = false
			item.Detail = "Unavailable: select workflow row"
			item.DisabledReason = "Select a workflow row in Workflows first"
		}
	case paletteDeleteTrigger:
		item.Label = "Action: Archive selected trigger"
		if !(state.View == ViewTriggers && state.HasSelection) {
			item.Enabled = false
			item.Detail = "Unavailable: select trigger row"
			item.DisabledReason = "Select a trigger row in Triggers first"
		}
	case paletteCreateSecret:
		item.Label = "Action: Create secret"
		if state.View != ViewSecrets {
			item.Enabled = false
			item.Detail = "Unavailable in this view"
			item.DisabledReason = "Open Secrets first"
		}
	case paletteUpdateSecret:
		item.Label = "Action: Update selected secret"
		if !(state.View == ViewSecrets && state.HasSelection) {
			item.Enabled = false
			item.Detail = "Unavailable: select secret row"
			item.DisabledReason = "Select a secret row in Secrets first"
		}
	case paletteDeleteSecret:
		item.Label = "Action: Delete selected secret"
		if !(state.View == ViewSecrets && state.HasSelection) {
			item.Enabled = false
			item.Detail = "Unavailable: select secret row"
			item.DisabledReason = "Select a secret row in Secrets first"
		}
	case paletteSetStatusScope:
		if !state.HasScope {
			item.Enabled = false
			item.Detail = "Unavailable in this view"
			item.DisabledReason = "Status scope is available in Workflows and Triggers"
		}
		scope := statusScopeFromValue(action.Value)
		suffix := ""
		if state.Scope == scope {
			suffix = " (active)"
		}
		switch scope {
		case statusScopeActive:
			item.Label = "Filter: Active only" + suffix
		default:
			if scope == statusScopeInactive {
				item.Label = "Filter: Inactive only" + suffix
			} else {
				item.Label = "Filter: Show all" + suffix
			}
		}
	case paletteShowCLIHandoff:
		switch action.Value {
		case "workflow-create":
			item.Label = "CLI: Create workflow"
		case "workflow-version-create":
			item.Label = "CLI: Create workflow version"
		default:
			item.Label = "CLI: Open authoring command"
		}
	case paletteClearFilters:
		item.Label = "Action: Clear filters"
		if !state.HasFilter {
			item.Enabled = false
			item.Detail = "Unavailable: no active filters"
			item.DisabledReason = "No search filter is active"
		}
	case paletteToggleRefresh:
		item.Label = "Toggle: Auto refresh"
	case paletteSetTheme:
		item.Label = "Theme"
		switch action.View {
		case ViewID("catppuccin"):
			item.Label = "Theme: Catppuccin"
		case ViewID("tokyo-night"):
			item.Label = "Theme: Tokyo Night"
		case ViewID("fallout"):
			item.Label = "Theme: Fallout (CRT)"
		case ViewID("retro-amber"):
			item.Label = "Theme: Retro Amber"
		}
	case paletteSetNetworkProfile:
		switch action.Profile {
		case NetworkFast:
			item.Label = "Network: Fast"
		case NetworkNormal:
			item.Label = "Network: Normal"
		case NetworkSlow:
			item.Label = "Network: Slow"
		case NetworkFlaky:
			item.Label = "Network: Flaky"
		}
	case paletteClearRecent:
		item.Label = "Action: Clear recent commands"
		if !state.HasRecent {
			item.Enabled = false
			item.Detail = "Unavailable: no recent commands"
			item.DisabledReason = "Run commands from the palette first"
		}
	}
	return item
}

func (m *Model) setStatusScopeForView(view ViewID, scope statusScope) {
	if !supportsStatusScope(view) {
		return
	}
	m.statusScopeByView[view] = scope
	m.applyFilter()
}

func (m *Model) cycleStatusScopeForCurrentViewCmd() tea.Cmd {
	if !supportsStatusScope(m.view) {
		return m.pushToast(ToastWarn, "Status scope is available in Workflows and Triggers")
	}
	current := m.currentStatusScope(m.view)
	next := statusScopeAll
	switch current {
	case statusScopeAll:
		next = statusScopeActive
	case statusScopeActive:
		next = statusScopeInactive
	default:
		next = statusScopeAll
	}
	m.setStatusScopeForView(m.view, next)
	return m.pushToast(ToastInfo, "Status filter: "+strings.ToLower(statusScopeLabel(next)))
}

func (m *Model) clearCurrentViewFilters() {
	m.searchQuery = ""
	m.searchInput.SetValue("")
	if supportsStatusScope(m.view) {
		m.statusScopeByView[m.view] = statusScopeAll
	}
	m.applyFilter()
}

func buildPalette(theme styles.Theme, recentActions []paletteAction, state paletteBuildState) list.Model {
	section := func(label string) paletteItem {
		return paletteItem{Label: label, Section: true, Enabled: false, Action: paletteAction{Kind: paletteNoop}}
	}
	command := func(label string, detail string, action paletteAction, keywords ...string) paletteItem {
		return paletteItem{Label: label, Detail: detail, Action: action, Enabled: true, Keywords: keywords}
	}

	autoStatus := "OFF"
	if state.AutoRefresh {
		autoStatus = "ON"
	}

	runSelected := command("Action: Run selected workflow", "Workflow", paletteAction{Kind: paletteRunWorkflow}, "run", "workflow", "queue")
	if !(state.View == ViewWorkflows && state.HasSelection) {
		runSelected.Enabled = false
		runSelected.Detail = "Unavailable: select workflow row"
		runSelected.DisabledReason = "Select a workflow row in Workflows first"
	}

	renameWorkflow := command("Action: Rename selected workflow", "Workflow", paletteAction{Kind: paletteRenameWorkflow}, "rename", "workflow", "name")
	if !(state.View == ViewWorkflows && state.HasSelection) {
		renameWorkflow.Enabled = false
		renameWorkflow.Detail = "Unavailable: select workflow row"
		renameWorkflow.DisabledReason = "Select a workflow row in Workflows first"
	}

	createTrigger := command("Action: Create trigger", "Trigger", paletteAction{Kind: paletteCreateTrigger}, "create", "trigger", "workflow")
	if !((state.View == ViewWorkflows || state.View == ViewTriggers) && state.HasSelection) {
		createTrigger.Enabled = false
		createTrigger.Detail = "Unavailable: select workflow or trigger row"
		createTrigger.DisabledReason = "Select a row in Workflows or Triggers first"
	}

	renameTrigger := command("Action: Update selected trigger", "Trigger", paletteAction{Kind: paletteRenameTrigger}, "update", "trigger", "name", "config")
	if !(state.View == ViewTriggers && state.HasSelection) {
		renameTrigger.Enabled = false
		renameTrigger.Detail = "Unavailable: select trigger row"
		renameTrigger.DisabledReason = "Select a trigger row in Triggers first"
	}

	toggleTrigger := command("Action: Toggle selected trigger", "Trigger", paletteAction{Kind: paletteToggleTrigger}, "toggle", "trigger", "active")
	if !(state.View == ViewTriggers && state.HasSelection) {
		toggleTrigger.Enabled = false
		toggleTrigger.Detail = "Unavailable: select trigger row"
		toggleTrigger.DisabledReason = "Select a trigger row in Triggers first"
	}

	deleteWorkflow := command("Action: Archive selected workflow", "Workflow", paletteAction{Kind: paletteDeleteWorkflow}, "archive", "workflow", "soft-delete")
	if !(state.View == ViewWorkflows && state.HasSelection) {
		deleteWorkflow.Enabled = false
		deleteWorkflow.Detail = "Unavailable: select workflow row"
		deleteWorkflow.DisabledReason = "Select a workflow row in Workflows first"
	}

	deleteTrigger := command("Action: Archive selected trigger", "Trigger", paletteAction{Kind: paletteDeleteTrigger}, "archive", "trigger", "soft-delete")
	if !(state.View == ViewTriggers && state.HasSelection) {
		deleteTrigger.Enabled = false
		deleteTrigger.Detail = "Unavailable: select trigger row"
		deleteTrigger.DisabledReason = "Select a trigger row in Triggers first"
	}

	createSecret := command("Action: Create secret", "Secret", paletteAction{Kind: paletteCreateSecret}, "create", "secret", "credential")
	if state.View != ViewSecrets {
		createSecret.Enabled = false
		createSecret.Detail = "Unavailable in this view"
		createSecret.DisabledReason = "Open Secrets first"
	}

	updateSecret := command("Action: Update selected secret", "Secret", paletteAction{Kind: paletteUpdateSecret}, "update", "secret", "edit")
	if !(state.View == ViewSecrets && state.HasSelection) {
		updateSecret.Enabled = false
		updateSecret.Detail = "Unavailable: select secret row"
		updateSecret.DisabledReason = "Select a secret row in Secrets first"
	}

	deleteSecret := command("Action: Delete selected secret", "Secret", paletteAction{Kind: paletteDeleteSecret}, "delete", "secret", "hard-delete")
	if !(state.View == ViewSecrets && state.HasSelection) {
		deleteSecret.Enabled = false
		deleteSecret.Detail = "Unavailable: select secret row"
		deleteSecret.DisabledReason = "Select a secret row in Secrets first"
	}

	showAllScope := command("Filter: Show all", "Status scope", paletteAction{Kind: paletteSetStatusScope, Value: "all"}, "filter", "status", "all")
	showActiveScope := command("Filter: Active only", "Status scope", paletteAction{Kind: paletteSetStatusScope, Value: "active"}, "filter", "status", "active")
	showInactiveScope := command("Filter: Inactive only", "Status scope", paletteAction{Kind: paletteSetStatusScope, Value: "inactive"}, "filter", "status", "inactive")
	if !state.HasScope {
		showAllScope.Enabled = false
		showAllScope.Detail = "Unavailable in this view"
		showAllScope.DisabledReason = "Status scope is available in Workflows and Triggers"
		showActiveScope.Enabled = false
		showActiveScope.Detail = "Unavailable in this view"
		showActiveScope.DisabledReason = "Status scope is available in Workflows and Triggers"
		showInactiveScope.Enabled = false
		showInactiveScope.Detail = "Unavailable in this view"
		showInactiveScope.DisabledReason = "Status scope is available in Workflows and Triggers"
	}
	if state.Scope == statusScopeAll {
		showAllScope.Label += " (active)"
	}
	if state.Scope == statusScopeActive {
		showActiveScope.Label += " (active)"
	}
	if state.Scope == statusScopeInactive {
		showInactiveScope.Label += " (active)"
	}

	clearFilters := command("Action: Clear filters", "Table", paletteAction{Kind: paletteClearFilters}, "clear", "filter", "reset")
	if !state.HasFilter {
		clearFilters.Enabled = false
		clearFilters.Detail = "Unavailable: no active filters"
		clearFilters.DisabledReason = "No search filter is active"
	}

	clearRecent := command("Action: Clear recent commands", "System", paletteAction{Kind: paletteClearRecent}, "recent", "history", "clear")
	if !state.HasRecent {
		clearRecent.Enabled = false
		clearRecent.Detail = "Unavailable: no recent commands"
		clearRecent.DisabledReason = "Run commands from the palette first"
	}

	profileItem := func(label string, profile NetworkProfile, active bool) paletteItem {
		detail := "Network"
		if active {
			detail = "Network (active)"
		}
		return command(label, detail, paletteAction{Kind: paletteSetNetworkProfile, Profile: profile}, "network", "latency", "profile")
	}

	items := []list.Item{
		section(":: Navigation"),
		command("Go: Dashboard", "Navigation", paletteAction{Kind: paletteGoToView, View: ViewDashboard}, "dash", "home", "overview"),
		command("Go: Workflows", "Navigation", paletteAction{Kind: paletteGoToView, View: ViewWorkflows}, "wf", "workflow"),
		command("Go: Runs", "Navigation", paletteAction{Kind: paletteGoToView, View: ViewRuns}, "run", "jobs"),
		command("Go: Triggers", "Navigation", paletteAction{Kind: paletteGoToView, View: ViewTriggers}, "trigger"),
		command("Go: Events", "Navigation", paletteAction{Kind: paletteGoToView, View: ViewEvents}, "event", "webhook"),
		command("Go: Secrets", "Navigation", paletteAction{Kind: paletteGoToView, View: ViewSecrets}, "secret", "vault"),
		command("Go: API Tokens", "Navigation", paletteAction{Kind: paletteGoToView, View: ViewTokens}, "token", "auth", "api"),
		section(":: Actions"),
		runSelected,
		renameWorkflow,
		createTrigger,
		renameTrigger,
		toggleTrigger,
		deleteWorkflow,
		deleteTrigger,
		createSecret,
		updateSecret,
		deleteSecret,
		showAllScope,
		showActiveScope,
		showInactiveScope,
		clearFilters,
		command("Toggle: Auto refresh", "System ("+autoStatus+")", paletteAction{Kind: paletteToggleRefresh}, "refresh", "polling", "live"),
		section(":: CLI Handoff"),
		command("CLI: Create workflow", "Workflow authoring", paletteAction{Kind: paletteShowCLIHandoff, Value: "workflow-create"}, "cli", "workflow", "create", "definition"),
		command("CLI: Create workflow version", "Version authoring", paletteAction{Kind: paletteShowCLIHandoff, Value: "workflow-version-create"}, "cli", "version", "definition"),
		clearRecent,
		section(":: Network"),
		profileItem("Network: Fast", NetworkFast, state.Profile == NetworkFast),
		profileItem("Network: Normal", NetworkNormal, state.Profile == NetworkNormal),
		profileItem("Network: Slow", NetworkSlow, state.Profile == NetworkSlow),
		profileItem("Network: Flaky", NetworkFlaky, state.Profile == NetworkFlaky),
		section(":: Themes"),
		command("Theme: Catppuccin", "Theme", paletteAction{Kind: paletteSetTheme, View: ViewID("catppuccin")}, "theme", "pastel"),
		command("Theme: Tokyo Night", "Theme", paletteAction{Kind: paletteSetTheme, View: ViewID("tokyo-night")}, "theme", "blue"),
		command("Theme: Fallout (CRT)", "Theme", paletteAction{Kind: paletteSetTheme, View: ViewID("fallout")}, "theme", "crt", "green"),
		command("Theme: Retro Amber", "Theme", paletteAction{Kind: paletteSetTheme, View: ViewID("retro-amber")}, "theme", "amber", "crt"),
	}
	if len(recentActions) > 0 {
		recent := make([]list.Item, 0, len(recentActions)+1)
		recent = append(recent, section(":: Recent"))
		for _, action := range recentActions {
			recent = append(recent, paletteItemFromAction(action, state))
		}
		items = append(recent, items...)
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetHeight(1)
	delegate.SetSpacing(0)
	selectedBorder := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(theme.Accent)
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(theme.Accent).
		Background(theme.SurfaceAlt).
		Bold(true).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(theme.Accent)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(theme.Text).
		Background(theme.SurfaceAlt).
		Inherit(selectedBorder)
	delegate.Styles.NormalTitle = lipgloss.NewStyle().Foreground(theme.Text)
	delegate.Styles.NormalDesc = lipgloss.NewStyle().Foreground(theme.Muted)
	model := list.New(items, delegate, 64, 16)
	listStyles := list.DefaultStyles()
	listStyles.Title = lipgloss.NewStyle().Foreground(theme.Text).Bold(true)
	listStyles.FilterPrompt = lipgloss.NewStyle().Foreground(theme.Accent)
	listStyles.FilterCursor = lipgloss.NewStyle().Foreground(theme.Accent)
	listStyles.PaginationStyle = lipgloss.NewStyle().Foreground(theme.Muted)
	listStyles.HelpStyle = lipgloss.NewStyle().Foreground(theme.Muted)
	listStyles.NoItems = lipgloss.NewStyle().Foreground(theme.Muted)
	model.Styles = listStyles
	model.SetFilteringEnabled(true)
	model.SetShowHelp(false)
	model.SetShowStatusBar(false)
	model.SetShowTitle(false)
	model.DisableQuitKeybindings()
	return model
}
