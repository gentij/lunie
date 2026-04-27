package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("lunie %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("date:   %s\n", date)
	},
}
