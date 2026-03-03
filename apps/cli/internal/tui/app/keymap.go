package app

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up            key.Binding
	Down          key.Binding
	NextScreen    key.Binding
	PrevScreen    key.Binding
	ToggleContext key.Binding
	Search        key.Binding
	ContextSearch key.Binding
	PanelScroll   key.Binding
	ContextScroll key.Binding
	ContextTabs   key.Binding
	Palette       key.Binding
	Help          key.Binding
	Quit          key.Binding
	Enter         key.Binding
	Back          key.Binding
	Clear         key.Binding
	Retry         key.Binding
	SortColumn    key.Binding
	SortDirection key.Binding
	JumpTop       key.Binding
	JumpBottom    key.Binding
	RunWorkflow   key.Binding
	ToggleActive  key.Binding
	Rename        key.Binding
	CreateTrigger key.Binding
	ViewVersions  key.Binding
	RevokeToken   key.Binding
	ToggleWrap    key.Binding
	LogSearch     key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:            key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:          key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		NextScreen:    key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next pane")),
		PrevScreen:    key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev pane")),
		ToggleContext: key.NewBinding(key.WithKeys("ctrl+j"), key.WithHelp("ctrl+j", "toggle context")),
		Search:        key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		ContextSearch: key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("ctrl+f", "panel search")),
		PanelScroll:   key.NewBinding(key.WithKeys("alt+up", "alt+down", "pgup", "pgdown"), key.WithHelp("alt+↑/↓", "main scroll")),
		ContextScroll: key.NewBinding(key.WithKeys("j", "k", "pgup", "pgdown", "ctrl+u", "ctrl+d"), key.WithHelp("j/k", "ctx scroll (focus)")),
		ContextTabs:   key.NewBinding(key.WithKeys("[", "]", "1", "2", "3", "4"), key.WithHelp("[/]", "ctx tabs (focus)")),
		Palette:       key.NewBinding(key.WithKeys("ctrl+k"), key.WithHelp("ctrl+k", "command palette")),
		Help:          key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Quit:          key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Enter:         key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
		Back:          key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Clear:         key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "clear")),
		Retry:         key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("ctrl+r", "retry")),
		SortColumn:    key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "sort column")),
		SortDirection: key.NewBinding(key.WithKeys("S"), key.WithHelp("S", "sort dir")),
		JumpTop:       key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "top")),
		JumpBottom:    key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "bottom")),
		RunWorkflow:   key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "run now")),
		ToggleActive:  key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "toggle active")),
		Rename:        key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "rename")),
		CreateTrigger: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "create trigger")),
		ViewVersions:  key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "view versions")),
		RevokeToken:   key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "revoke")),
		ToggleWrap:    key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "wrap logs")),
		LogSearch:     key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search logs")),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.Enter,
		k.NextScreen,
		k.ToggleContext,
		k.Search,
		k.Palette,
		k.Help,
		k.Quit,
	}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Back},
		{k.NextScreen, k.PrevScreen, k.ToggleContext, k.Search, k.ContextSearch},
		{k.PanelScroll, k.ContextScroll, k.ContextTabs, k.Palette, k.Help, k.Quit, k.Clear, k.Retry},
		{k.SortColumn, k.SortDirection, k.JumpTop, k.JumpBottom},
		{k.RunWorkflow, k.ToggleActive, k.Rename, k.CreateTrigger, k.ViewVersions, k.RevokeToken},
		{k.ToggleWrap, k.LogSearch},
	}
}
