package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/gentij/lunie/apps/cli/internal/api"
	"github.com/gentij/lunie/apps/cli/internal/output"
	"github.com/spf13/cobra"
)

const defaultServerURL = "http://localhost:3000/v1/api"

var (
	configPath string
	serverURL  string
	outputMode string
	quiet      bool
	noColor    bool
)

var rootCmd = &cobra.Command{
	Use:   "lunie",
	Short: "Lunie CLI",
	Long:  "Lunie CLI for managing workflows, runs, triggers, and secrets.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, _, err := loadConfig(cmd.Flags().Changed("server"))
		if err != nil {
			return err
		}

		if outputMode == "" {
			outputMode = "table"
		}
		outputMode = strings.ToLower(outputMode)
		if outputMode != "table" && outputMode != "json" {
			return fmt.Errorf("invalid output format: %s", outputMode)
		}

		ctx := &Context{
			Config:  cfg,
			Client:  api.NewClient(cfg.ServerURL, cfg.Token),
			Output:  outputMode,
			Quiet:   quiet,
			NoColor: noColor,
		}

		output.SetNoColor(noColor)

		cmd.SetContext(WithContext(cmd.Context(), ctx))
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		output.PrintError(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to config file")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", defaultServerURL, "API server URL")
	rootCmd.PersistentFlags().StringVar(
		&outputMode,
		"output",
		"table",
		"Output format (table|json)",
	)
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "Minimal output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(workflowCmd)
	rootCmd.AddCommand(triggerCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(stepCmd)
	rootCmd.AddCommand(secretCmd)
}
