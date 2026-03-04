package app

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gentij/taskforge/apps/cli/internal/tui/styles"
)

func (m *Model) nextView() {
	for i, view := range m.views {
		if view == m.view {
			m.view = m.views[(i+1)%len(m.views)]
			m.refreshView()
			return
		}
	}
}

func (m *Model) prevView() {
	for i, view := range m.views {
		if view == m.view {
			idx := i - 1
			if idx < 0 {
				idx = len(m.views) - 1
			}
			m.view = m.views[idx]
			m.refreshView()
			return
		}
	}
}

func (m *Model) focusNext() {
	if m.focus == FocusSidebar {
		m.focus = FocusMain
		return
	}
	if m.focus == FocusMain {
		if !m.contextCollapsed && m.layout.ContextHeight > 0 {
			m.focus = FocusContext
		} else {
			m.focus = FocusSidebar
		}
		return
	}
	m.focus = FocusSidebar
}

func (m *Model) focusPrev() {
	if m.focus == FocusSidebar {
		if !m.contextCollapsed && m.layout.ContextHeight > 0 {
			m.focus = FocusContext
		} else {
			m.focus = FocusMain
		}
		return
	}
	if m.focus == FocusMain {
		m.focus = FocusSidebar
		return
	}
	m.focus = FocusMain
}

func (m *Model) updateSidebar(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "right", "enter":
			m.focus = FocusMain
			return m, nil
		}
	}
	prev := m.sidebar.Index()
	var cmd tea.Cmd
	m.sidebar, cmd = m.sidebar.Update(msg)
	if m.sidebar.Index() != prev {
		m.syncViewFromSidebar()
	}
	return m, cmd
}

func (m *Model) syncViewFromSidebar() {
	item, ok := m.sidebar.SelectedItem().(navItem)
	if !ok {
		return
	}
	if item.ID == m.view {
		return
	}
	m.view = item.ID
	m.refreshView()
}

func (m *Model) syncSidebarSelection() {
	items := m.sidebar.Items()
	for i, item := range items {
		nav, ok := item.(navItem)
		if !ok {
			continue
		}
		if nav.ID == m.view {
			m.sidebar.Select(i)
			return
		}
	}
}

func buildSidebar(theme styles.Theme, selected ViewID) list.Model {
	items := []list.Item{
		navItem{ID: ViewDashboard, Label: "Dashboard"},
		navItem{ID: ViewWorkflows, Label: "Workflows"},
		navItem{ID: ViewRuns, Label: "Runs"},
		navItem{ID: ViewTriggers, Label: "Triggers"},
		navItem{ID: ViewEvents, Label: "Events"},
		navItem{ID: ViewSecrets, Label: "Secrets"},
		navItem{ID: ViewTokens, Label: "API Tokens"},
	}
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetHeight(1)
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(theme.Background).Background(theme.Accent).Bold(true)
	delegate.Styles.NormalTitle = lipgloss.NewStyle().Foreground(theme.Muted)
	model := list.New(items, delegate, 20, 10)
	model.SetShowHelp(false)
	model.SetShowStatusBar(false)
	model.SetFilteringEnabled(false)
	model.SetShowTitle(false)
	model.SetShowPagination(false)
	model.DisableQuitKeybindings()

	for i, item := range items {
		nav, ok := item.(navItem)
		if ok && nav.ID == selected {
			model.Select(i)
			break
		}
	}
	return model
}
