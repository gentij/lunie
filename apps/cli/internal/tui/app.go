package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gentij/taskforge/apps/cli/internal/api"
)

type App struct {
	client    *api.Client
	serverURL string
	tokenSet  bool
}

func NewApp(client *api.Client, serverURL string, tokenSet bool) *App {
	return &App{client: client, serverURL: serverURL, tokenSet: tokenSet}
}

func (a *App) Start() error {
	model := newModel(a.client, a.serverURL, a.tokenSet)
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err := program.Run()
	return err
}

type healthResponse struct {
	Status  string  `json:"status"`
	Version string  `json:"version"`
	Uptime  float64 `json:"uptime"`
	DB      struct {
		Ok bool `json:"ok"`
	} `json:"db"`
}

type healthMsg struct {
	data *healthResponse
	err  error
}

type model struct {
	client      *api.Client
	serverURL   string
	tokenSet    bool
	width       int
	height      int
	spinner     spinner.Model
	loading     bool
	lastUpdated time.Time
	health      *healthResponse
	err         error
}

func newModel(client *api.Client, serverURL string, tokenSet bool) model {
	spin := spinner.New(spinner.WithSpinner(spinner.Dot))
	return model{
		client:    client,
		serverURL: serverURL,
		tokenSet:  tokenSet,
		spinner:   spin,
		loading:   true,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, fetchHealthCmd(m.client))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.loading = true
			m.err = nil
			return m, tea.Batch(m.spinner.Tick, fetchHealthCmd(m.client))
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case healthMsg:
		m.loading = false
		m.lastUpdated = time.Now()
		m.health = msg.data
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	header := lipgloss.NewStyle().Bold(true).Render("Taskforge TUI")
	serverLine := fmt.Sprintf("Server: %s", m.serverURL)
	tokenLine := "Token: missing"
	if m.tokenSet {
		tokenLine = "Token: set"
	}

	statusBlock := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(
		fmt.Sprintf("%s\n%s", serverLine, tokenLine),
	)

	body := ""
	if m.loading {
		body = fmt.Sprintf("%s Loading health...", m.spinner.View())
	} else if m.err != nil {
		body = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(
			fmt.Sprintf("Error: %s", m.err.Error()),
		)
	} else if m.health != nil {
		dbStatus := "down"
		if m.health.DB.Ok {
			dbStatus = "ok"
		}
		body = fmt.Sprintf(
			"Status: %s\nVersion: %s\nDB: %s",
			m.health.Status,
			m.health.Version,
			dbStatus,
		)
	}

	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
		"q quit • r refresh",
	)
	if !m.lastUpdated.IsZero() {
		footer = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(
			fmt.Sprintf("q quit • r refresh • updated %s", m.lastUpdated.Format("15:04:05")),
		)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		statusBlock,
		"",
		body,
		"",
		footer,
	)

	return lipgloss.NewStyle().Padding(1, 2).Render(content)
}

func fetchHealthCmd(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return healthMsg{err: fmt.Errorf("missing API client")}
		}

		var health healthResponse
		err := client.GetJSON("/health", &health)
		if err != nil {
			return healthMsg{err: err}
		}
		return healthMsg{data: &health}
	}
}
