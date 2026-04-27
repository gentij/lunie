package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	logsFollow      bool
	logsTail        int
	startForeground bool
)

var startCmd = &cobra.Command{
	Use:   "start [service...]",
	Short: "Start the local Lunie stack",
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir, composePath, envPath, err := resolveInitPaths()
		if err != nil {
			return err
		}

		composeArgs := []string{"--env-file", envPath, "up"}
		if !startForeground {
			composeArgs = append(composeArgs, "-d")
		}
		composeArgs = append(composeArgs, args...)

		return runDockerCompose(baseDir, composePath, composeArgs...)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show local Lunie stack status",
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir, composePath, envPath, err := resolveInitPaths()
		if err != nil {
			return err
		}
		return runDockerCompose(baseDir, composePath, "--env-file", envPath, "ps")
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the local Lunie stack",
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir, composePath, envPath, err := resolveInitPaths()
		if err != nil {
			return err
		}
		return runDockerCompose(baseDir, composePath, "--env-file", envPath, "stop")
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs [service...]",
	Short: "Show Lunie stack logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir, composePath, envPath, err := resolveInitPaths()
		if err != nil {
			return err
		}

		composeArgs := []string{"--env-file", envPath, "logs"}
		if logsFollow {
			composeArgs = append(composeArgs, "--follow")
		}
		if logsTail > 0 {
			composeArgs = append(composeArgs, "--tail", strconv.Itoa(logsTail))
		}
		composeArgs = append(composeArgs, args...)
		return runDockerCompose(baseDir, composePath, composeArgs...)
	},
}

func init() {
	startCmd.Flags().BoolVar(&startForeground, "foreground", false, "Run services in foreground")
	logsCmd.Flags().BoolVar(&logsFollow, "follow", false, "Follow log output")
	logsCmd.Flags().IntVar(&logsTail, "tail", 200, "Lines to show from the end of logs")
}

func resolveInitPaths() (string, string, string, error) {
	baseDir, err := resolveInitDir()
	if err != nil {
		return "", "", "", err
	}
	composePath := filepath.Join(baseDir, composeFileName)
	envPath := filepath.Join(baseDir, envFileName)

	if _, err := os.Stat(composePath); err != nil {
		if os.IsNotExist(err) {
			return "", "", "", fmt.Errorf("lunie not initialized; run 'lunie init'")
		}
		return "", "", "", err
	}
	if _, err := os.Stat(envPath); err != nil {
		if os.IsNotExist(err) {
			return "", "", "", fmt.Errorf("missing env file; run 'lunie init'")
		}
		return "", "", "", err
	}

	return baseDir, composePath, envPath, nil
}
