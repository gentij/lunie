package styles

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name       string
	Background lipgloss.Color
	Surface    lipgloss.Color
	SurfaceAlt lipgloss.Color
	Border     lipgloss.Color
	Text       lipgloss.Color
	Muted      lipgloss.Color
	Accent     lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Error      lipgloss.Color
	Info       lipgloss.Color
	SuccessBg  lipgloss.Color
	WarningBg  lipgloss.Color
	ErrorBg    lipgloss.Color
	InfoBg     lipgloss.Color
	MutedBg    lipgloss.Color
	CRT        bool
	Scanline   lipgloss.Color
	Glow       lipgloss.Color
}

type StyleSet struct {
	Header           lipgloss.Style
	Footer           lipgloss.Style
	PanelBorder      lipgloss.Style
	PanelBorderFocus lipgloss.Style
	PanelTitle       lipgloss.Style
	BorderColor      lipgloss.Color
	Accent           lipgloss.Style
	SidebarTitle     lipgloss.Style
	SidebarMuted     lipgloss.Style
	SidebarSection   lipgloss.Style
	SidebarFill      lipgloss.Style
	Chip             lipgloss.Style
	ChipActive       lipgloss.Style
	TabActive        lipgloss.Style
	TabInactive      lipgloss.Style
	Divider          lipgloss.Style
	PanelFill        lipgloss.Style
	ContextFill      lipgloss.Style
	TableHeader      lipgloss.Style
	TableCell        lipgloss.Style
	TableSelected    lipgloss.Style
	RowAlt           lipgloss.Style
	CardTitle        lipgloss.Style
	CardValue        lipgloss.Style
	BadgeSuccess     lipgloss.Style
	BadgeFailed      lipgloss.Style
	BadgeRunning     lipgloss.Style
	BadgeQueued      lipgloss.Style
	BadgeMuted       lipgloss.Style
	Dim              lipgloss.Style
}

func DefaultTheme() Theme {
	return Theme{
		Name:       "Taskforge",
		Background: lipgloss.Color("#0B0F14"),
		Surface:    lipgloss.Color("#111827"),
		SurfaceAlt: lipgloss.Color("#0F172A"),
		Border:     lipgloss.Color("#1F2937"),
		Text:       lipgloss.Color("#E5E7EB"),
		Muted:      lipgloss.Color("#94A3B8"),
		Accent:     lipgloss.Color("#38BDF8"),
		Success:    lipgloss.Color("#22C55E"),
		Warning:    lipgloss.Color("#F59E0B"),
		Error:      lipgloss.Color("#EF4444"),
		Info:       lipgloss.Color("#60A5FA"),
		SuccessBg:  lipgloss.Color("#0B2E1A"),
		WarningBg:  lipgloss.Color("#2B1A05"),
		ErrorBg:    lipgloss.Color("#2A0F14"),
		InfoBg:     lipgloss.Color("#0B1F33"),
		MutedBg:    lipgloss.Color("#111827"),
		CRT:        false,
		Scanline:   lipgloss.Color("#0B0F14"),
		Glow:       lipgloss.Color("#38BDF8"),
	}
}

func ThemeRegistry() map[string]Theme {
	return map[string]Theme{
		"taskforge": DefaultTheme(),
		"tokyo-night": {
			Name:       "Tokyo Night",
			Background: lipgloss.Color("#0B0F1A"),
			Surface:    lipgloss.Color("#151A2C"),
			SurfaceAlt: lipgloss.Color("#101522"),
			Border:     lipgloss.Color("#262B3F"),
			Text:       lipgloss.Color("#DDE3F0"),
			Muted:      lipgloss.Color("#8A93A6"),
			Accent:     lipgloss.Color("#7AA2F7"),
			Success:    lipgloss.Color("#9ECE6A"),
			Warning:    lipgloss.Color("#E0AF68"),
			Error:      lipgloss.Color("#F7768E"),
			Info:       lipgloss.Color("#7DCFFF"),
			SuccessBg:  lipgloss.Color("#1F2A1A"),
			WarningBg:  lipgloss.Color("#2A1F10"),
			ErrorBg:    lipgloss.Color("#2A151A"),
			InfoBg:     lipgloss.Color("#10233A"),
			MutedBg:    lipgloss.Color("#151A2C"),
		},
		"catppuccin": {
			Name:       "Catppuccin",
			Background: lipgloss.Color("#0F0E17"),
			Surface:    lipgloss.Color("#1C1B24"),
			SurfaceAlt: lipgloss.Color("#16151D"),
			Border:     lipgloss.Color("#2A2934"),
			Text:       lipgloss.Color("#E6E1F0"),
			Muted:      lipgloss.Color("#A59BBE"),
			Accent:     lipgloss.Color("#8AADF4"),
			Success:    lipgloss.Color("#A6DA95"),
			Warning:    lipgloss.Color("#EED49F"),
			Error:      lipgloss.Color("#ED8796"),
			Info:       lipgloss.Color("#91D7E3"),
			SuccessBg:  lipgloss.Color("#1A2418"),
			WarningBg:  lipgloss.Color("#2A2216"),
			ErrorBg:    lipgloss.Color("#2A151A"),
			InfoBg:     lipgloss.Color("#13252B"),
			MutedBg:    lipgloss.Color("#1C1B24"),
			CRT:        false,
			Scanline:   lipgloss.Color("#0F0E17"),
			Glow:       lipgloss.Color("#8AADF4"),
		},
		"fallout": {
			Name:       "Fallout",
			Background: lipgloss.Color("#050A05"),
			Surface:    lipgloss.Color("#0B140B"),
			SurfaceAlt: lipgloss.Color("#081108"),
			Border:     lipgloss.Color("#2D7A3C"),
			Text:       lipgloss.Color("#B8FFB0"),
			Muted:      lipgloss.Color("#6FBF6A"),
			Accent:     lipgloss.Color("#7CFF6B"),
			Success:    lipgloss.Color("#7CFF6B"),
			Warning:    lipgloss.Color("#F2E94E"),
			Error:      lipgloss.Color("#FF6B6B"),
			Info:       lipgloss.Color("#7CFF6B"),
			SuccessBg:  lipgloss.Color("#0B2A0B"),
			WarningBg:  lipgloss.Color("#2A2405"),
			ErrorBg:    lipgloss.Color("#2A0F0F"),
			InfoBg:     lipgloss.Color("#0A2A0A"),
			MutedBg:    lipgloss.Color("#0B140B"),
			CRT:        true,
			Scanline:   lipgloss.Color("#0A160A"),
			Glow:       lipgloss.Color("#7CFF6B"),
		},
		"retro-amber": {
			Name:       "Retro Amber",
			Background: lipgloss.Color("#120A05"),
			Surface:    lipgloss.Color("#1A0F08"),
			SurfaceAlt: lipgloss.Color("#140B06"),
			Border:     lipgloss.Color("#8B5C24"),
			Text:       lipgloss.Color("#FCE9C1"),
			Muted:      lipgloss.Color("#C98D4A"),
			Accent:     lipgloss.Color("#F2B24C"),
			Success:    lipgloss.Color("#F2B24C"),
			Warning:    lipgloss.Color("#F59E0B"),
			Error:      lipgloss.Color("#FF9A6B"),
			Info:       lipgloss.Color("#F2B24C"),
			SuccessBg:  lipgloss.Color("#2B1A08"),
			WarningBg:  lipgloss.Color("#2B1A05"),
			ErrorBg:    lipgloss.Color("#2A1208"),
			InfoBg:     lipgloss.Color("#2B1A08"),
			MutedBg:    lipgloss.Color("#1A0F08"),
			CRT:        false,
			Scanline:   lipgloss.Color("#120A05"),
			Glow:       lipgloss.Color("#F2B24C"),
		},
	}
}

func NewStyles(theme Theme) StyleSet {
	return StyleSet{
		Header: lipgloss.NewStyle().
			Background(theme.SurfaceAlt).
			Foreground(theme.Text),
		Footer: lipgloss.NewStyle().
			Background(theme.SurfaceAlt).
			Foreground(theme.Muted),
		PanelBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Muted).
			Background(theme.Surface),
		PanelBorderFocus: lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(theme.Accent).
			Background(theme.Surface),
		PanelTitle: lipgloss.NewStyle().
			Foreground(theme.Text).
			Bold(true),
		BorderColor: theme.Border,
		Accent: lipgloss.NewStyle().
			Foreground(theme.Accent).
			Bold(true),
		SidebarTitle: lipgloss.NewStyle().
			Foreground(theme.Accent).
			Bold(true),
		SidebarMuted: lipgloss.NewStyle().
			Foreground(theme.Muted),
		SidebarSection: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Bold(true),
		SidebarFill: lipgloss.NewStyle().
			Background(theme.SurfaceAlt),
		Chip: lipgloss.NewStyle().
			Foreground(theme.Text).
			Background(theme.SurfaceAlt).
			Padding(0, 1),
		ChipActive: lipgloss.NewStyle().
			Foreground(theme.Background).
			Background(theme.Accent).
			Padding(0, 1).
			Bold(true),
		TabActive: lipgloss.NewStyle().
			Foreground(theme.Background).
			Background(theme.Accent).
			Bold(true),
		TabInactive: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Background(theme.SurfaceAlt),
		Divider: lipgloss.NewStyle().
			Foreground(theme.Border),
		PanelFill: lipgloss.NewStyle().
			Background(theme.Surface),
		ContextFill: lipgloss.NewStyle().
			Background(theme.SurfaceAlt),
		TableHeader: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Bold(true),
		TableCell: lipgloss.NewStyle().
			Foreground(theme.Text),
		TableSelected: lipgloss.NewStyle().
			Foreground(theme.Text).
			Bold(true),
		RowAlt: lipgloss.NewStyle().
			Foreground(theme.Text),
		CardTitle: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Bold(true),
		CardValue: lipgloss.NewStyle().
			Foreground(theme.Text).
			Bold(true),
		BadgeSuccess: lipgloss.NewStyle().
			Foreground(theme.Success).
			Background(theme.SuccessBg).
			Bold(true),
		BadgeFailed: lipgloss.NewStyle().
			Foreground(theme.Error).
			Background(theme.ErrorBg).
			Bold(true),
		BadgeRunning: lipgloss.NewStyle().
			Foreground(theme.Info).
			Background(theme.InfoBg).
			Bold(true),
		BadgeQueued: lipgloss.NewStyle().
			Foreground(theme.Warning).
			Background(theme.WarningBg).
			Bold(true),
		BadgeMuted: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Background(theme.MutedBg).
			Bold(true),
		Dim: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Faint(true),
	}
}

func Badge(style lipgloss.Style, text string) string {
	return style.Render("[" + text + "]")
}
