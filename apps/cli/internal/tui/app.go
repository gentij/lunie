package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentij/lunie/apps/cli/internal/api"
	"github.com/gentij/lunie/apps/cli/internal/config"
	"github.com/gentij/lunie/apps/cli/internal/tui/app"
)

type App struct {
	client     *api.Client
	serverURL  string
	tokenSet   bool
	config     config.Config
	configPath string
}

func NewApp(client *api.Client, serverURL string, tokenSet bool, cfg config.Config, configPath string) *App {
	return &App{client: client, serverURL: serverURL, tokenSet: tokenSet, config: cfg, configPath: configPath}
}

func (a *App) Start() error {
	model := app.NewModel(a.client, a.serverURL, a.tokenSet, a.config, a.configPath)
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err := program.Run()
	return err
}
