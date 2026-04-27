package cli

import (
	"fmt"

	"github.com/gentij/lunie/apps/cli/internal/config"
	"github.com/gentij/lunie/apps/cli/internal/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the Lunie TUI",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := GetContext(cmd.Context())
		if ctx == nil {
			return fmt.Errorf("missing context")
		}
		configPath := config.ResolvePath(configPath)
		app := tui.NewApp(ctx.Client, ctx.Config.ServerURL, ctx.Config.Token != "", ctx.Config, configPath)
		return app.Start()
	},
}
