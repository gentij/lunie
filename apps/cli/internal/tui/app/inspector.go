package app

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gentij/lunie/apps/cli/internal/tui/data"
	"github.com/gentij/lunie/apps/cli/internal/tui/styles"
	"github.com/gentij/lunie/apps/cli/internal/tui/utils"
)

type inspectorFocus int

const (
	inspectorSteps inspectorFocus = iota
	inspectorLogs
)

type RunInspector struct {
	Active    bool
	Focus     inspectorFocus
	Steps     list.Model
	Logs      viewport.Model
	LogWrap   bool
	Searching bool
	Search    textinput.Model
	RunID     string
	Width     int
	Height    int
	styles    styles.StyleSet
	keys      KeyMap
}

func NewInspector(styleSet styles.StyleSet, keys KeyMap) RunInspector {
	search := textinput.New()
	search.Prompt = "log/ "
	search.Placeholder = "Search logs"
	search.CharLimit = 64
	steps := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	steps.SetShowHelp(false)
	steps.SetShowStatusBar(false)
	steps.SetFilteringEnabled(false)
	steps.SetShowTitle(false)
	steps.SetShowPagination(false)
	logs := viewport.New(0, 0)
	return RunInspector{
		Focus:  inspectorSteps,
		Steps:  steps,
		Logs:   logs,
		Search: search,
		styles: styleSet,
		keys:   keys,
	}
}

func (ri *RunInspector) ApplyStyles(styleSet styles.StyleSet) {
	ri.styles = styleSet
}

func (ri *RunInspector) Resize(width int, height int) {
	ri.Width = width
	ri.Height = height
	modalWidth := min(width-4, 120)
	modalHeight := min(height-4, 30)
	if modalWidth < 40 {
		modalWidth = 40
	}
	if modalHeight < 12 {
		modalHeight = 12
	}

	innerWidth := max(modalWidth-2, 1)
	innerHeight := max(modalHeight-2, 1)
	leftWidth := max(int(float64(innerWidth)*0.3), 18)
	rightWidth := max(innerWidth-leftWidth-2, 1)
	contentHeight := max(innerHeight-1, 1)

	ri.Steps.SetSize(leftWidth, contentHeight)
	ri.Logs.Width = rightWidth
	ri.Logs.Height = contentHeight
}

func (m *Model) openInspector() {
	selected := m.selectedRowID()
	if selected == "" {
		return
	}
	run, ok := runByID(&m.store, selected)
	if !ok {
		return
	}
	steps := stepsForRun(&m.store, run.ID)
	items := make([]list.Item, 0, len(steps))
	for _, step := range steps {
		items = append(items, step)
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	listModel := list.New(items, delegate, m.inspector.Steps.Width(), m.inspector.Steps.Height())
	listModel.SetShowHelp(false)
	listModel.SetShowStatusBar(false)
	listModel.SetFilteringEnabled(false)
	listModel.SetShowTitle(false)
	listModel.SetShowPagination(false)

	m.inspector.Steps = listModel
	m.inspector.RunID = run.ID
	m.inspector.Active = true
	m.inspector.Focus = inspectorSteps
	m.inspector.Searching = false
	m.inspector.Search.SetValue("")
	m.inspector.Resize(m.width, m.height)
	m.inspector.SyncLog(steps)
}

func (m Model) updateInspector(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Back) || key.Matches(msg, m.keys.Quit) {
			m.inspector.Active = false
			return m, nil
		}
		if key.Matches(msg, m.keys.NextScreen) {
			if m.inspector.Focus == inspectorSteps {
				m.inspector.Focus = inspectorLogs
			} else {
				m.inspector.Focus = inspectorSteps
			}
			return m, nil
		}
		if key.Matches(msg, m.keys.ToggleWrap) {
			m.inspector.LogWrap = !m.inspector.LogWrap
			m.inspector.SyncLog(stepsForRun(&m.store, m.inspector.RunID))
			return m, nil
		}
		if key.Matches(msg, m.keys.LogSearch) {
			m.inspector.Searching = true
			m.inspector.Search.SetValue("")
			m.inspector.Search.CursorEnd()
			m.inspector.Search.Focus()
			return m, nil
		}
		if m.inspector.Searching {
			if key.Matches(msg, m.keys.Back) {
				m.inspector.Searching = false
				m.inspector.Search.SetValue("")
				m.inspector.Search.Blur()
				m.inspector.SyncLog(stepsForRun(&m.store, m.inspector.RunID))
				return m, nil
			}
			if key.Matches(msg, m.keys.Enter) {
				m.inspector.Searching = false
				m.inspector.Search.Blur()
				m.inspector.SyncLog(stepsForRun(&m.store, m.inspector.RunID))
				return m, nil
			}
			if key.Matches(msg, m.keys.Clear) {
				m.inspector.Search.SetValue("")
				m.inspector.SyncLog(stepsForRun(&m.store, m.inspector.RunID))
				return m, nil
			}
			var cmd tea.Cmd
			m.inspector.Search, cmd = m.inspector.Search.Update(msg)
			m.inspector.SyncLog(stepsForRun(&m.store, m.inspector.RunID))
			return m, cmd
		}
		if m.inspector.Focus == inspectorSteps {
			var cmd tea.Cmd
			m.inspector.Steps, cmd = m.inspector.Steps.Update(msg)
			m.inspector.SyncLog(stepsForRun(&m.store, m.inspector.RunID))
			return m, cmd
		}
		var cmd tea.Cmd
		m.inspector.Logs, cmd = m.inspector.Logs.Update(msg)
		return m, cmd
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		return m, nil
	}

	return m, nil
}

func (ri *RunInspector) SyncLog(steps []data.StepRun) {
	item, ok := ri.Steps.SelectedItem().(data.StepRun)
	if !ok {
		ri.Logs.SetContent("No logs")
		return
	}
	content := item.Log
	query := strings.TrimSpace(ri.Search.Value())
	if query != "" {
		content = utils.FilterLines(content, query)
	}
	if ri.LogWrap && ri.Logs.Width > 0 {
		content = utils.WrapText(content, ri.Logs.Width)
	}
	ri.Logs.SetContent(content)
}

func (ri RunInspector) Render(width int, height int) string {
	if !ri.Active {
		return ""
	}
	modalWidth := min(width-4, 120)
	modalHeight := min(height-4, 30)
	if modalWidth < 40 {
		modalWidth = 40
	}
	if modalHeight < 12 {
		modalHeight = 12
	}
	innerWidth := max(modalWidth-2, 1)
	innerHeight := max(modalHeight-2, 1)
	leftWidth := max(int(float64(innerWidth)*0.3), 18)
	rightWidth := max(innerWidth-leftWidth-2, 1)
	stepsView := strings.TrimRight(ri.Steps.View(), "\n")
	logsView := strings.TrimRight(ri.Logs.View(), "\n")
	left := inspectorColumn("Steps", stepsView, leftWidth, innerHeight, ri.styles)
	right := inspectorColumn("Logs", logsView, rightWidth, innerHeight, ri.styles)
	row := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
	box := ri.styles.PanelBorder.Width(modalWidth).Height(modalHeight)
	content := box.Render(row)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func inspectorColumn(title string, content string, width int, height int, styleSet styles.StyleSet) string {
	if width < 1 {
		width = 1
	}
	if height < 2 {
		height = 2
	}
	innerHeight := max(height-1, 1)
	header := styleSet.PanelTitle.Render(title)
	body := lipgloss.Place(width, innerHeight, lipgloss.Left, lipgloss.Top, content)
	return lipgloss.JoinVertical(lipgloss.Left, header, body)
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
