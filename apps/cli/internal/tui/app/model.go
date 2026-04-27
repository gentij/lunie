package app

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentij/lunie/apps/cli/internal/api"
	"github.com/gentij/lunie/apps/cli/internal/config"
	"github.com/gentij/lunie/apps/cli/internal/tui/components"
	"github.com/gentij/lunie/apps/cli/internal/tui/data"
	"github.com/gentij/lunie/apps/cli/internal/tui/layout"
	"github.com/gentij/lunie/apps/cli/internal/tui/styles"
)

type ViewID string

type FocusPane int

const (
	ViewDashboard ViewID = "dashboard"
	ViewWorkflows ViewID = "workflows"
	ViewRuns      ViewID = "runs"
	ViewTriggers  ViewID = "triggers"
	ViewEvents    ViewID = "events"
	ViewSecrets   ViewID = "secrets"
	ViewTokens    ViewID = "tokens"
)

const (
	FocusSidebar FocusPane = iota
	FocusMain
	FocusContext
)

type paletteActionType int

const (
	paletteNoop paletteActionType = iota
	paletteGoToView
	paletteToggleRefresh
	paletteClearFilters
	paletteRunWorkflow
	paletteRenameWorkflow
	paletteCreateTrigger
	paletteRenameTrigger
	paletteToggleTrigger
	paletteDeleteWorkflow
	paletteDeleteTrigger
	paletteCreateSecret
	paletteUpdateSecret
	paletteDeleteSecret
	paletteSetStatusScope
	paletteShowCLIHandoff
	paletteClearRecent
	paletteSetTheme
	paletteSetNetworkProfile
)

type statusScope int

const (
	statusScopeAll statusScope = iota
	statusScopeActive
	statusScopeInactive
)

type pulseMsg struct{}

type toastClearMsg struct {
	id int
}

type SurfaceState int

const (
	SurfaceIdle SurfaceState = iota
	SurfaceLoading
	SurfaceRefreshing
	SurfaceSuccess
	SurfaceError
	SurfaceStale
	SurfaceEmpty
)

type ToastLevel int

const (
	ToastInfo ToastLevel = iota
	ToastSuccess
	ToastWarn
	ToastError
)

type ToastState struct {
	ID      int
	Active  bool
	Level   ToastLevel
	Message string
}

type NetworkProfile int

const (
	NetworkFast NetworkProfile = iota
	NetworkNormal
	NetworkSlow
	NetworkFlaky
)

type paletteAction struct {
	Kind    paletteActionType
	View    ViewID
	Profile NetworkProfile
	Value   string
}

type ContextTab int

const (
	ContextTabOverview ContextTab = iota
	ContextTabJSON
	ContextTabSteps
	ContextTabLogs
)

type actionModalMode int

const (
	actionModalNone actionModalMode = iota
	actionModalRenameWorkflow
	actionModalRenameTrigger
	actionModalCreateTrigger
	actionModalUpdateTrigger
	actionModalCreateSecret
	actionModalUpdateSecret
	actionModalConfirmDelete
	actionModalCLIHandoff
)

type actionModalState struct {
	Active         bool
	Mode           actionModalMode
	Title          string
	Description    string
	Validation     string
	ShowValidation bool
	Primary        textinput.Model
	Secondary      textinput.Model
	Tertiary       textinput.Model
	Confirm        textinput.Model
	Focus          int
	WorkflowID     string
	TriggerID      string
	SecretID       string
	DeleteKind     string
	TriggerType    string
	TriggerActive  bool
	ConfirmPhrase  string
	CLICommand     string
}

type paletteItem struct {
	Label          string
	Detail         string
	Action         paletteAction
	Section        bool
	Enabled        bool
	DisabledReason string
	Keywords       []string
}

type paletteBuildState struct {
	View         ViewID
	HasSelection bool
	HasFilter    bool
	HasScope     bool
	Scope        statusScope
	AutoRefresh  bool
	Profile      NetworkProfile
	HasRecent    bool
}

type SortConfig struct {
	Column int
	Desc   bool
}

func (p paletteItem) FilterValue() string {
	if p.Section {
		return ""
	}
	parts := []string{p.Label, p.Detail, strings.Join(p.Keywords, " ")}
	if !p.Enabled && p.DisabledReason != "" {
		parts = append(parts, p.DisabledReason)
	}
	return strings.Join(parts, " ")
}
func (p paletteItem) Title() string       { return p.Label }
func (p paletteItem) Description() string { return p.Detail }

type navItem struct {
	ID    ViewID
	Label string
}

func (n navItem) FilterValue() string { return n.Label }
func (n navItem) Title() string       { return n.Label }
func (n navItem) Description() string { return "" }

type Model struct {
	client     *api.Client
	serverURL  string
	tokenSet   bool
	config     config.Config
	configPath string

	focus FocusPane

	width  int
	height int
	layout layout.Layout

	view  ViewID
	views []ViewID

	theme     styles.Theme
	themeName string
	styles    styles.StyleSet

	store data.Store

	columns        []table.Column
	baseRows       []table.Row
	baseRowIDs     []string
	filteredRows   []table.Row
	filteredRowIDs []string

	table table.Model

	searchInput textinput.Model
	searching   bool
	searchQuery string

	contextViewport    viewport.Model
	contextCollapsed   bool
	contextTab         ContextTab
	contextOffsets     map[ContextTab]int
	contextSelectedID  string
	contextSearchInput textinput.Model
	contextSearching   bool
	contextQuery       string

	palette     list.Model
	showPalette bool
	sidebar     list.Model
	mainPanel   viewport.Model

	inspector RunInspector

	showHelp bool
	help     help.Model
	keys     KeyMap
	action   actionModalState

	autoRefresh     bool
	refreshEvery    time.Duration
	lastRefresh     time.Time
	pulseOn         bool
	refreshPending  bool
	refreshCount    int
	mutationPending bool
	networkProfile  NetworkProfile
	workspaceName   string
	apiStatus       string
	paletteRecent   []paletteAction

	mainState    SurfaceState
	contextState SurfaceState
	toast        ToastState
	uiReady      bool

	sortByView map[ViewID]SortConfig
	sortColumn int
	sortDesc   bool

	statusScopeByView map[ViewID]statusScope

	paginator paginator.Model

	snapshotSort snapshotSortOptions
}

func NewModel(client *api.Client, serverURL string, tokenSet bool, cfg config.Config, configPath string) Model {
	now := time.Now()
	store := data.Store{}
	keys := DefaultKeyMap()
	helper := help.New()
	helper.ShowAll = false

	search := textinput.New()
	search.Prompt = "/ "
	search.Placeholder = "Filter"
	search.CharLimit = 64

	contextSearch := textinput.New()
	contextSearch.Prompt = "panel/ "
	contextSearch.Placeholder = "Search context"
	contextSearch.CharLimit = 64

	defaultTheme := styles.DefaultTheme()
	palette := buildPalette(defaultTheme, nil, paletteBuildState{View: ViewDashboard, Profile: NetworkNormal})
	sidebar := buildSidebar(defaultTheme, ViewDashboard)
	styleSet := styles.NewStyles(defaultTheme)

	tableModel := components.NewTable(nil, nil, 0, 0, styleSet)
	contextViewport := viewport.New(0, 0)
	mainPanel := viewport.New(0, 0)

	pager := paginator.New()
	pager.Type = paginator.Arabic
	pager.PerPage = 1
	pager.SetTotalPages(1)

	model := Model{
		client:             client,
		serverURL:          serverURL,
		tokenSet:           tokenSet,
		config:             cfg,
		configPath:         configPath,
		focus:              FocusSidebar,
		view:               ViewDashboard,
		views:              []ViewID{ViewDashboard, ViewWorkflows, ViewRuns, ViewTriggers, ViewEvents, ViewSecrets, ViewTokens},
		theme:              defaultTheme,
		themeName:          "lunie",
		styles:             styleSet,
		store:              store,
		table:              tableModel,
		searchInput:        search,
		contextViewport:    contextViewport,
		contextTab:         ContextTabOverview,
		contextOffsets:     map[ContextTab]int{},
		contextSearchInput: contextSearch,
		palette:            palette,
		sidebar:            sidebar,
		mainPanel:          mainPanel,
		help:               helper,
		keys:               keys,
		action:             actionModalState{},
		refreshEvery:       2 * time.Second,
		lastRefresh:        now,
		pulseOn:            false,
		refreshPending:     false,
		refreshCount:       0,
		mutationPending:    false,
		networkProfile:     NetworkNormal,
		workspaceName:      "personal",
		apiStatus:          apiStatus(tokenSet),
		paginator:          pager,
		inspector:          NewInspector(styleSet, keys),
		mainState:          SurfaceLoading,
		contextState:       SurfaceLoading,
		uiReady:            false,
		sortByView:         map[ViewID]SortConfig{},
		sortColumn:         -1,
		sortDesc:           true,
		statusScopeByView:  map[ViewID]statusScope{},
		snapshotSort:       defaultSnapshotSortOptions(),
	}
	model.setNetworkProfile(NetworkNormal)

	model.applyTheme(cfg.Theme, false)

	width, height := initialSize()
	model.resize(width, height)
	model.refreshView()

	return model
}

func (m Model) Init() tea.Cmd {
	windowSizeCmd := func() tea.Msg {
		width, height := initialSize()
		return tea.WindowSizeMsg{Width: width, Height: height}
	}
	return tea.Batch(windowSizeCmd, pulseTick(), fetchSnapshotCmd(m.client, m.snapshotSort, m.profileDelay(), m.profileShouldFail(false)))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.inspector.Active {
		return m.updateInspector(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		m.refreshView()
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	case pulseMsg:
		m.pulseOn = !m.pulseOn
		cmds := []tea.Cmd{pulseTick()}
		if m.autoRefresh && !m.refreshPending && time.Since(m.lastRefresh) >= m.refreshEvery {
			m.startMockRefresh(false)
			cmds = append(cmds, fetchSnapshotCmd(m.client, m.snapshotSort, m.profileDelay(), m.profileShouldFail(false)))
		}
		return m, tea.Batch(cmds...)
	case snapshotLoadedMsg:
		m.refreshPending = false
		m.lastRefresh = time.Now()
		if msg.err != nil {
			m.apiStatus = "OFFLINE"
			if !m.uiReady {
				m.uiReady = true
				m.mainState = SurfaceError
				m.contextState = SurfaceError
				m.refreshView()
			}
			if m.mainState != SurfaceError {
				m.mainState = SurfaceStale
			}
			if m.contextState != SurfaceError {
				m.contextState = SurfaceStale
			}
			return m, m.pushToast(ToastWarn, "Failed to sync data (ctrl+r retry)")
		}
		m.store = msg.store
		m.apiStatus = msg.apiStatus
		m.uiReady = true
		m.mainState = SurfaceIdle
		m.contextState = SurfaceIdle
		m.refreshView()
		m.syncSurfaceStates()
		return m, nil
	case toastClearMsg:
		if m.toast.Active && m.toast.ID == msg.id {
			m.toast = ToastState{}
		}
		return m, nil
	case mutationResultMsg:
		m.mutationPending = false
		if msg.err != nil {
			return m, m.pushToast(ToastError, mutationErrorMessage(msg.err))
		}
		cmds := []tea.Cmd{}
		if strings.TrimSpace(msg.successMessage) != "" {
			cmds = append(cmds, m.pushToast(ToastSuccess, msg.successMessage))
		}
		if msg.refresh {
			m.startMockRefresh(false)
			cmds = append(cmds, fetchSnapshotCmd(m.client, m.snapshotSort, m.profileDelay(), m.profileShouldFail(false)))
		}
		return m, tea.Batch(cmds...)
	}

	if m.searching {
		return m.updateSearch(msg)
	}
	if m.contextSearching {
		return m.updateContextSearch(msg)
	}
	if m.showPalette {
		return m.updatePalette(msg)
	}

	return m.updateMain(msg)
}

func (m Model) View() string {
	return Render(m)
}

func (m *Model) resize(width int, height int) {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	m.width = width
	m.height = height
	m.layout = layout.Compute(width, height, m.contextCollapsed)

	innerWidth := max(m.layout.MainWidth-2, 1)
	innerHeight := max(m.layout.MainHeight-2, 1)
	m.table.SetWidth(innerWidth)
	m.table.SetHeight(m.layout.PrimaryTableHeight)

	m.paginator.PerPage = m.layout.PrimaryTableHeight

	contextBodyHeight := max(m.layout.ContextHeight-4, 1)
	contextWidth := max(innerWidth-2, 1)
	if m.layout.ContextHeight == 0 {
		contextBodyHeight = 0
	}
	m.contextViewport.Width = contextWidth
	m.contextViewport.Height = contextBodyHeight

	m.sidebar.SetSize(max(m.layout.SidebarWidth-2, 1), max(m.layout.SidebarHeight-10, 1))
	mainBodyHeight := innerHeight
	if !m.contextCollapsed && m.layout.ContextHeight > 0 {
		mainBodyHeight -= m.layout.ContextHeight
	}
	if mainBodyHeight < 1 {
		mainBodyHeight = 1
	}
	m.mainPanel.Width = innerWidth
	m.mainPanel.Height = mainBodyHeight
	if (m.contextCollapsed || m.layout.ContextHeight == 0) && m.focus == FocusContext {
		m.focus = FocusMain
	}

	m.inspector.Resize(width, height)
	m.help.Width = max(m.width-2, 1)
	m.resizePalette()
	m.updateMainPanel()
}
