package cli

import (
	"fmt"

	"github.com/gentij/taskforge/apps/cli/internal/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the Taskforge TUI",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := GetContext(cmd.Context())
		if ctx == nil {
			return fmt.Errorf("missing context")
		}

		app := tui.NewApp(ctx.Client, ctx.Config.ServerURL, ctx.Config.Token != "")
		return app.Start()
	},
}
