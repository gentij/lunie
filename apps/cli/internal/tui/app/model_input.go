package app

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.action.Active {
		return m.updateActionModal(msg)
	}
	if m.showHelp {
		if key.Matches(msg, m.keys.Help) || key.Matches(msg, m.keys.Back) {
			m.showHelp = false
			m.help.ShowAll = false
			return m, nil
		}
		return m, nil
	}
	if m.showPalette {
		return m.updatePalette(msg)
	}
	if m.searching {
		return m.updateSearch(msg)
	}
	if m.contextSearching {
		return m.updateContextSearch(msg)
	}

	if key.Matches(msg, m.keys.Quit) {
		return m, tea.Quit
	}
	if key.Matches(msg, m.keys.Help) {
		m.showHelp = true
		m.help.ShowAll = true
		return m, nil
	}
	if key.Matches(msg, m.keys.Retry) && m.canRetry() {
		m.startMockRefresh(true)
		clearToastCmd := m.pushToast(ToastInfo, "Retrying refresh...")
		return m, tea.Batch(fetchSnapshotCmd(m.client, m.profileDelay(), m.profileShouldFail(true)), clearToastCmd)
	}
	if key.Matches(msg, m.keys.Palette) {
		m.palette = buildPalette(m.theme, m.paletteRecent, m.paletteState())
		m.resizePalette()
		m.palette.ResetFilter()
		m.ensurePaletteSelection(true)
		m.showPalette = true
		return m, nil
	}
	if key.Matches(msg, m.keys.Search) {
		m.searching = true
		m.searchInput.SetValue(m.searchQuery)
		m.searchInput.CursorEnd()
		m.searchInput.Focus()
		return m, nil
	}
	if key.Matches(msg, m.keys.ContextSearch) {
		m.contextSearching = true
		m.contextSearchInput.SetValue(m.contextQuery)
		m.contextSearchInput.CursorEnd()
		m.contextSearchInput.Focus()
		return m, nil
	}
	if key.Matches(msg, m.keys.ToggleContext) {
		m.contextCollapsed = !m.contextCollapsed
		if m.contextCollapsed && m.focus == FocusContext {
			m.focus = FocusMain
		}
		m.resize(m.width, m.height)
		m.updateContext()
		return m, nil
	}
	if key.Matches(msg, m.keys.NextScreen) {
		m.focusNext()
		return m, nil
	}
	if key.Matches(msg, m.keys.PrevScreen) {
		m.focusPrev()
		return m, nil
	}
	if msg.String() == "left" {
		if m.focus == FocusContext {
			m.focus = FocusMain
		} else {
			m.focus = FocusSidebar
		}
		return m, nil
	}
	if msg.String() == "right" {
		if m.focus == FocusSidebar {
			m.focus = FocusMain
			return m, nil
		}
		if m.focus == FocusMain && !m.contextCollapsed && m.layout.ContextHeight > 0 {
			m.focus = FocusContext
			return m, nil
		}
		return m, nil
	}
	if m.focus == FocusMain && m.view == ViewRuns && key.Matches(msg, m.keys.Enter) {
		m.openInspector()
		return m, nil
	}
	if m.view == ViewWorkflows && key.Matches(msg, m.keys.RunWorkflow) {
		return m, m.queueRunForSelectedWorkflowCmd()
	}
	if m.view == ViewWorkflows && key.Matches(msg, m.keys.ToggleActive) {
		return m, m.toggleWorkflowActiveCmd()
	}
	if m.view == ViewWorkflows && key.Matches(msg, m.keys.Rename) {
		return m, m.openRenameWorkflowModalCmd()
	}
	if (m.view == ViewWorkflows || m.view == ViewTriggers) && key.Matches(msg, m.keys.CreateTrigger) {
		return m, m.openCreateTriggerModalCmd()
	}
	if m.view == ViewSecrets && key.Matches(msg, m.keys.CreateTrigger) {
		return m, m.openCreateSecretModalCmd()
	}
	if m.view == ViewTriggers && key.Matches(msg, m.keys.Rename) {
		return m, m.openRenameTriggerModalCmd()
	}
	if m.view == ViewSecrets && key.Matches(msg, m.keys.Rename) {
		return m, m.openUpdateSecretModalCmd()
	}
	if m.view == ViewTriggers && key.Matches(msg, m.keys.ToggleActive) {
		return m, m.toggleTriggerActiveCmd()
	}
	if m.view == ViewWorkflows && key.Matches(msg, m.keys.RevokeToken) {
		return m, m.openDeleteWorkflowModalCmd()
	}
	if m.view == ViewTriggers && key.Matches(msg, m.keys.RevokeToken) {
		return m, m.openDeleteTriggerModalCmd()
	}
	if m.view == ViewSecrets && key.Matches(msg, m.keys.RevokeToken) {
		return m, m.openDeleteSecretModalCmd()
	}
	if m.view == ViewTokens && key.Matches(msg, m.keys.RevokeToken) {
		return m, m.pushToast(ToastWarn, "API tokens are not available yet")
	}

	if m.focus == FocusSidebar {
		return m.updateSidebar(msg)
	}
	if m.focus == FocusContext {
		return m.updateContextPane(msg)
	}
	return m.updateMain(msg)
}

func (m *Model) updateMain(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, m.keys.CycleStatus) {
			if supportsStatusScope(m.view) {
				return m, m.cycleStatusScopeForCurrentViewCmd()
			}
		}
		if key.Matches(keyMsg, m.keys.SortColumn) {
			m.cycleSortColumn()
			return m, nil
		}
		if key.Matches(keyMsg, m.keys.SortDirection) {
			m.toggleSortDirection()
			return m, nil
		}
		if key.Matches(keyMsg, m.keys.JumpTop) {
			m.table.SetCursor(0)
			m.applyTableRows()
			return m, nil
		}
		if key.Matches(keyMsg, m.keys.JumpBottom) {
			if n := len(m.filteredRows); n > 0 {
				m.table.SetCursor(n - 1)
				m.applyTableRows()
			}
			return m, nil
		}
		if m.handleMainScroll(keyMsg) {
			return m, nil
		}
	}

	prevCursor := m.table.Cursor()
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	if prevCursor != m.table.Cursor() {
		m.applyTableRows()
	}
	return m, cmd
}

func (m *Model) handleContextScroll(msg tea.KeyMsg) bool {
	if m.contextViewport.Height <= 0 {
		return false
	}
	switch msg.String() {
	case "up", "k", "shift+up", "shift+k":
		m.contextViewport.LineUp(1)
		m.contextOffsets[m.contextTab] = m.contextViewport.YOffset
		m.updateMainPanel()
		return true
	case "down", "j", "shift+down", "shift+j":
		m.contextViewport.LineDown(1)
		m.contextOffsets[m.contextTab] = m.contextViewport.YOffset
		m.updateMainPanel()
		return true
	case "pgup", "pageup", "shift+pgup", "shift+pageup", "ctrl+u":
		m.contextViewport.LineUp(m.contextViewport.Height)
		m.contextOffsets[m.contextTab] = m.contextViewport.YOffset
		m.updateMainPanel()
		return true
	case "pgdown", "pagedown", "pgdn", "shift+pgdown", "shift+pagedown", "shift+pgdn", "ctrl+d":
		m.contextViewport.LineDown(m.contextViewport.Height)
		m.contextOffsets[m.contextTab] = m.contextViewport.YOffset
		m.updateMainPanel()
		return true
	case "home":
		m.contextViewport.GotoTop()
		m.contextOffsets[m.contextTab] = m.contextViewport.YOffset
		m.updateMainPanel()
		return true
	case "end":
		m.contextViewport.GotoBottom()
		m.contextOffsets[m.contextTab] = m.contextViewport.YOffset
		m.updateMainPanel()
		return true
	}
	return false
}

func (m *Model) updateContextPane(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "]", "l":
			m.nextContextTab()
			return m, nil
		case "[", "h":
			m.prevContextTab()
			return m, nil
		case "1":
			m.setContextTab(ContextTabOverview)
			return m, nil
		case "2":
			m.setContextTab(ContextTabJSON)
			return m, nil
		case "3":
			m.setContextTab(ContextTabSteps)
			return m, nil
		case "4":
			m.setContextTab(ContextTabLogs)
			return m, nil
		}
		if m.handleContextScroll(keyMsg) {
			return m, nil
		}
	}
	return m, nil
}

func (m *Model) nextContextTab() {
	m.contextOffsets[m.contextTab] = m.contextViewport.YOffset
	m.contextTab = (m.contextTab + 1) % 4
	m.updateContext()
}

func (m *Model) prevContextTab() {
	m.contextOffsets[m.contextTab] = m.contextViewport.YOffset
	m.contextTab = (m.contextTab + 3) % 4
	m.updateContext()
}

func (m *Model) setContextTab(tab ContextTab) {
	m.contextOffsets[m.contextTab] = m.contextViewport.YOffset
	m.contextTab = tab
	m.updateContext()
}

func (m *Model) handleMainScroll(msg tea.KeyMsg) bool {
	if m.mainPanel.Height <= 0 {
		return false
	}
	switch msg.String() {
	case "alt+up", "alt+k":
		m.mainPanel.LineUp(1)
		return true
	case "alt+down", "alt+j":
		m.mainPanel.LineDown(1)
		return true
	case "pgup", "pageup":
		m.mainPanel.LineUp(m.mainPanel.Height)
		return true
	case "pgdown", "pagedown", "pgdn":
		m.mainPanel.LineDown(m.mainPanel.Height)
		return true
	case "home":
		m.mainPanel.GotoTop()
		return true
	case "end":
		m.mainPanel.GotoBottom()
		return true
	}
	return false
}

func (m *Model) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, m.keys.Back) {
			m.searching = false
			m.searchInput.Blur()
			return m, nil
		}
		if key.Matches(keyMsg, m.keys.Enter) {
			m.searching = false
			m.searchInput.Blur()
			return m, nil
		}
		if key.Matches(keyMsg, m.keys.Clear) {
			m.searchInput.SetValue("")
			m.searchQuery = ""
			m.applyFilter()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.searchQuery = strings.TrimSpace(m.searchInput.Value())
	m.applyFilter()
	return m, cmd
}

func (m *Model) updateContextSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, m.keys.Back) {
			m.contextSearching = false
			m.contextSearchInput.Blur()
			return m, nil
		}
		if key.Matches(keyMsg, m.keys.Enter) {
			m.contextSearching = false
			m.contextSearchInput.Blur()
			return m, nil
		}
		if key.Matches(keyMsg, m.keys.Clear) {
			m.contextSearchInput.SetValue("")
			m.contextQuery = ""
			m.updateContext()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.contextSearchInput, cmd = m.contextSearchInput.Update(msg)
	m.contextQuery = strings.TrimSpace(m.contextSearchInput.Value())
	m.updateContext()
	return m, cmd
}

func (m *Model) updatePalette(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, m.keys.Back) {
			m.showPalette = false
			return m, nil
		}
		if key.Matches(keyMsg, m.keys.Enter) {
			item, ok := m.palette.SelectedItem().(paletteItem)
			if ok && !item.Section {
				if !item.Enabled {
					msg := "Command unavailable"
					if item.DisabledReason != "" {
						msg = item.DisabledReason
					}
					return m, m.pushToast(ToastWarn, msg)
				}
				cmd := m.runPaletteAction(item.Action)
				m.showPalette = false
				return m, cmd
			}
			return m, nil
		}
		if shouldAutoStartPaletteFilter(keyMsg, m.palette.SettingFilter()) {
			startMsg := tea.KeyMsg(tea.Key{Type: tea.KeyRunes, Runes: []rune{'/'}})
			var startCmd tea.Cmd
			m.palette, startCmd = m.palette.Update(startMsg)
			var typeCmd tea.Cmd
			m.palette, typeCmd = m.palette.Update(keyMsg)
			return m, tea.Batch(startCmd, typeCmd)
		}
	}
	var cmd tea.Cmd
	m.palette, cmd = m.palette.Update(msg)
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "up", "k":
			m.ensurePaletteSelection(false)
		case "down", "j", "pgup", "pageup", "pgdown", "pagedown", "pgdn", "home", "end":
			m.ensurePaletteSelection(true)
		default:
			if m.palette.SettingFilter() || m.palette.IsFiltered() {
				m.ensurePaletteSelection(true)
			}
		}
	}
	return m, cmd
}

func (m *Model) ensurePaletteSelection(forward bool) {
	items := m.palette.VisibleItems()
	if len(items) == 0 {
		return
	}
	for attempts := 0; attempts < len(items); attempts++ {
		item, ok := m.palette.SelectedItem().(paletteItem)
		if !ok || !item.Section {
			return
		}
		if forward {
			m.palette.CursorDown()
		} else {
			m.palette.CursorUp()
		}
	}
}

func shouldAutoStartPaletteFilter(msg tea.KeyMsg, alreadyFiltering bool) bool {
	if alreadyFiltering {
		return false
	}
	if msg.Type != tea.KeyRunes || len(msg.Runes) == 0 || msg.Alt {
		return false
	}
	for _, r := range msg.Runes {
		if r >= 32 && r != '/' {
			return true
		}
	}
	return false
}

func (m *Model) runPaletteAction(action paletteAction) tea.Cmd {
	switch action.Kind {
	case paletteNoop:
		return nil
	case paletteGoToView:
		m.rememberPaletteAction(action)
		m.view = action.View
		m.refreshView()
		return nil
	case paletteToggleRefresh:
		m.rememberPaletteAction(action)
		m.autoRefresh = !m.autoRefresh
		if m.autoRefresh {
			m.lastRefresh = time.Now()
		}
		return m.pushToast(ToastInfo, "Auto refresh toggled")
	case paletteClearFilters:
		m.rememberPaletteAction(action)
		m.clearCurrentViewFilters()
		return m.pushToast(ToastInfo, "Filters cleared")
	case paletteRunWorkflow:
		m.rememberPaletteAction(action)
		return m.queueRunForSelectedWorkflowCmd()
	case paletteRenameWorkflow:
		m.rememberPaletteAction(action)
		return m.openRenameWorkflowModalCmd()
	case paletteCreateTrigger:
		m.rememberPaletteAction(action)
		return m.openCreateTriggerModalCmd()
	case paletteRenameTrigger:
		m.rememberPaletteAction(action)
		return m.openRenameTriggerModalCmd()
	case paletteToggleTrigger:
		m.rememberPaletteAction(action)
		return m.toggleTriggerActiveCmd()
	case paletteDeleteWorkflow:
		m.rememberPaletteAction(action)
		return m.openDeleteWorkflowModalCmd()
	case paletteDeleteTrigger:
		m.rememberPaletteAction(action)
		return m.openDeleteTriggerModalCmd()
	case paletteCreateSecret:
		m.rememberPaletteAction(action)
		return m.openCreateSecretModalCmd()
	case paletteUpdateSecret:
		m.rememberPaletteAction(action)
		return m.openUpdateSecretModalCmd()
	case paletteDeleteSecret:
		m.rememberPaletteAction(action)
		return m.openDeleteSecretModalCmd()
	case paletteSetStatusScope:
		m.rememberPaletteAction(action)
		if !supportsStatusScope(m.view) {
			return m.pushToast(ToastWarn, "Status scope is available in Workflows and Triggers")
		}
		next := statusScopeFromValue(action.Value)
		m.setStatusScopeForView(m.view, next)
		return m.pushToast(ToastInfo, "Status filter: "+strings.ToLower(statusScopeLabel(next)))
	case paletteShowCLIHandoff:
		m.rememberPaletteAction(action)
		return m.openCLIHandoffModalCmd(action.Value)
	case paletteClearRecent:
		m.paletteRecent = nil
		return m.pushToast(ToastInfo, "Recent commands cleared")
	case paletteSetTheme:
		m.rememberPaletteAction(action)
		m.applyTheme(string(action.View), true)
		return m.pushToast(ToastInfo, "Theme switched")
	case paletteSetNetworkProfile:
		m.rememberPaletteAction(action)
		m.setNetworkProfile(action.Profile)
		return m.pushToast(ToastInfo, "Network profile set to "+strings.ToLower(networkProfileLabel(action.Profile)))
	}
	return nil
}
