package app

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/gentij/lunie/apps/cli/internal/config"
	"github.com/gentij/lunie/apps/cli/internal/tui/components"
	"github.com/gentij/lunie/apps/cli/internal/tui/screens"
	"github.com/gentij/lunie/apps/cli/internal/tui/styles"
	"github.com/gentij/lunie/apps/cli/internal/tui/utils"
)

func (m *Model) refreshView() {
	selectedID := m.selectedRowID()
	cursor := m.table.Cursor()
	columns, rows, rowIDs := screens.BuildRowsForView(screens.ViewID(m.view), &m.store, m.styles, max(m.layout.MainWidth-2, 1))
	cfg, ok := m.sortByView[m.view]
	if !ok || cfg.Column < 0 || cfg.Column >= len(columns) {
		cfg = defaultSortConfig(columns)
	}
	serverSortable := serverSortableColumnIndexesForView(m.view, columns)
	if len(serverSortable) > 0 && !containsIndex(serverSortable, cfg.Column) {
		cfg.Column = serverSortable[0]
	}
	m.sortByView[m.view] = cfg
	m.sortColumn = cfg.Column
	m.sortDesc = cfg.Desc
	if shouldUseClientSideSort(m.view, columns) {
		rows, rowIDs = screens.SortRowsForView(screens.ViewID(m.view), &m.store, columns, rows, rowIDs, cfg.Column, cfg.Desc)
	}
	m.columns = columns
	m.baseRows = rows
	m.baseRowIDs = rowIDs
	m.applyFilterWithSelection(selectedID, cursor)
	m.syncSidebarSelection()
}

func (m *Model) applyFilter() {
	selectedID := m.selectedRowID()
	cursor := m.table.Cursor()
	m.applyFilterWithSelection(selectedID, cursor)
}

func (m *Model) applyFilterWithSelection(selectedID string, cursor int) {
	rows, rowIDs := m.scopeRowsForCurrentView(m.baseRows, m.baseRowIDs)
	rows, rowIDs = filterRows(rows, rowIDs, m.searchQuery)
	m.filteredRows = rows
	m.filteredRowIDs = rowIDs
	m.restoreSelection(selectedID, cursor)
	m.syncSurfaceStates()
	m.applyTableRows()
}

func (m *Model) scopeRowsForCurrentView(rows []table.Row, rowIDs []string) ([]table.Row, []string) {
	scope := m.currentStatusScope(m.view)
	if scope == statusScopeAll || !supportsStatusScope(m.view) {
		return rows, rowIDs
	}
	filteredRows := make([]table.Row, 0, len(rows))
	filteredIDs := make([]string, 0, len(rowIDs))
	for i, id := range rowIDs {
		if i >= len(rows) {
			break
		}
		active, ok := m.isRowActiveForScope(m.view, id)
		if !ok {
			continue
		}
		if (scope == statusScopeActive && active) || (scope == statusScopeInactive && !active) {
			filteredRows = append(filteredRows, rows[i])
			filteredIDs = append(filteredIDs, id)
		}
	}
	return filteredRows, filteredIDs
}

func (m *Model) isRowActiveForScope(view ViewID, rowID string) (bool, bool) {
	switch view {
	case ViewWorkflows:
		wf, ok := workflowByID(&m.store, rowID)
		if !ok {
			return false, false
		}
		return wf.Active, true
	case ViewTriggers:
		trg, ok := triggerByID(&m.store, rowID)
		if !ok {
			return false, false
		}
		return trg.Active, true
	default:
		return false, false
	}
}

func (m *Model) restoreSelection(selectedID string, cursor int) {
	if len(m.filteredRowIDs) == 0 {
		m.table.SetCursor(0)
		return
	}
	if selectedID != "" {
		for i, id := range m.filteredRowIDs {
			if id == selectedID {
				m.table.SetCursor(i)
				return
			}
		}
	}
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= len(m.filteredRowIDs) {
		cursor = len(m.filteredRowIDs) - 1
	}
	m.table.SetCursor(cursor)
}

func (m *Model) applyTableRows() {
	m.table.SetRows(nil)
	m.table.SetColumns(columnsWithSortIndicators(m.columns, m.sortColumn, m.sortDesc))
	if len(m.filteredRows) == 0 {
		m.updatePaginator()
		m.updateContext()
		return
	}

	truncated := truncateRows(m.filteredRows, m.columns)
	m.filteredRows = truncated

	cursor := m.table.Cursor()
	if cursor >= len(truncated) {
		cursor = len(truncated) - 1
		m.table.SetCursor(cursor)
	}
	if cursor < 0 {
		cursor = 0
		m.table.SetCursor(0)
	}
	styled := components.StyleRows(truncated, m.columns, cursor, m.styles)
	m.table.SetRows(styled)
	m.updatePaginator()
	m.updateContext()
	m.updateMainPanel()
}

func (m *Model) updatePaginator() {
	perPage := max(m.layout.PrimaryTableHeight, 1)
	pages := (len(m.filteredRows) + perPage - 1) / perPage
	if pages < 1 {
		pages = 1
	}
	m.paginator.SetTotalPages(pages)
	page := 0
	if len(m.filteredRows) > 0 {
		page = m.table.Cursor() / perPage
	}
	if page >= pages {
		page = pages - 1
	}
	m.paginator.Page = page
}

func (m *Model) updateContext() {
	selectedID := m.selectedRowID()
	if selectedID != m.contextSelectedID {
		m.contextSelectedID = selectedID
		m.contextOffsets = map[ContextTab]int{}
	}
	content := screens.BuildContextTabContent(screens.ViewID(m.view), &m.store, selectedID, screens.ContextTab(m.contextTab))
	content = utils.FilterLines(content, m.contextQuery)
	if m.contextViewport.Width > 0 {
		content = utils.WrapText(content, m.contextViewport.Width)
	}
	m.contextViewport.SetContent(content)
	offset := m.contextOffsets[m.contextTab]
	m.contextViewport.SetYOffset(offset)
	m.syncSurfaceStates()
	m.updateMainPanel()
}

func (m *Model) updateMainPanel() {
	if m.mainPanel.Width == 0 || m.mainPanel.Height == 0 {
		return
	}
	content := buildMainContent(*m)
	content = sanitizeRenderable(content)
	content = truncateLines(content, max(m.mainPanel.Width, 1))
	content = clampSection(content, max(m.mainPanel.Width, 1), max(m.mainPanel.Height, 1))
	m.mainPanel.SetContent(content)
	m.mainPanel.GotoTop()
}

func (m *Model) applyTheme(themeKey string, persist bool) {
	registry := styles.ThemeRegistry()
	key := strings.ToLower(strings.TrimSpace(themeKey))
	if key == "" {
		key = "lunie"
	}
	selected, ok := registry[key]
	if !ok {
		selected = styles.DefaultTheme()
		key = "lunie"
	}

	m.theme = selected
	m.themeName = key
	m.styles = styles.NewStyles(selected)
	m.table.SetStyles(components.TableStyles(m.styles))
	m.inspector.ApplyStyles(m.styles)
	m.palette = buildPalette(m.theme, m.paletteRecent, m.paletteState())
	m.sidebar = buildSidebar(m.theme, m.view)
	m.resizePalette()

	if persist {
		m.config.Theme = key
		_ = config.Save(m.configPath, m.config)
	}

	m.refreshView()
}

func (m *Model) resizePalette() {
	if m.width <= 0 || m.height <= 0 {
		return
	}
	width := max(m.width-2, 1)
	height := max(m.height-5, 1)
	m.palette.SetSize(width, height)
}

func (m *Model) selectedRowID() string {
	if len(m.filteredRowIDs) == 0 {
		return ""
	}
	idx := m.table.Cursor()
	if idx < 0 || idx >= len(m.filteredRowIDs) {
		return ""
	}
	return m.filteredRowIDs[idx]
}
