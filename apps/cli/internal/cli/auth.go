package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gentij/lunie/apps/cli/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
}

var authToken string

func init() {
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login with an API token",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := loadConfig(cmd.Flags().Changed("server"))
			if err != nil {
				return err
			}

			token := strings.TrimSpace(authToken)
			if token == "" {
				if term.IsTerminal(int(os.Stdin.Fd())) {
					fmt.Fprint(os.Stdout, "API token: ")
					password, err := term.ReadPassword(int(os.Stdin.Fd()))
					fmt.Fprintln(os.Stdout)
					if err != nil {
						return err
					}
					token = strings.TrimSpace(string(password))
				} else {
					reader := bufio.NewReader(os.Stdin)
					line, err := reader.ReadString('\n')
					if err != nil {
						return err
					}
					token = strings.TrimSpace(line)
				}
			}

			if token == "" {
				return fmt.Errorf("token is required")
			}

			cfg.Token = token
			cfg.ServerURL = resolveServerURL(
				cfg.ServerURL,
				serverURL,
				cmd.Flags().Changed("server"),
			)

			if err := saveConfig(path, cfg); err != nil {
				return err
			}

			fmt.Printf("Saved credentials to %s\n", path)
			return nil
		},
	}
	loginCmd.Flags().StringVar(&authToken, "token", "", "API token")

	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Clear saved credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := loadConfig(cmd.Flags().Changed("server"))
			if err != nil {
				return err
			}

			cfg.Token = ""

			if err := saveConfig(path, cfg); err != nil {
				return err
			}

			fmt.Printf("Cleared credentials in %s\n", path)
			return nil
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show auth status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := loadConfig(cmd.Flags().Changed("server"))
			if err != nil {
				return err
			}

			fmt.Printf("Config: %s\n", path)
			fmt.Printf("Server: %s\n", cfg.ServerURL)
			fmt.Printf("Token:  %s\n", tokenStatus(cfg.Token))
			return nil
		},
	}

	whoamiCmd := &cobra.Command{
		Use:   "whoami",
		Short: "Validate token and show identity",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := GetContext(cmd.Context())
			if ctx == nil {
				return fmt.Errorf("missing context")
			}
			if strings.TrimSpace(ctx.Config.Token) == "" {
				return fmt.Errorf("token not set")
			}

			var result struct {
				ID     string   `json:"id"`
				Name   string   `json:"name"`
				Scopes []string `json:"scopes"`
			}
			if err := ctx.Client.GetJSON("/auth/whoami", &result); err != nil {
				return err
			}

			if IsJSON(ctx) {
				return output.PrintJSON(result)
			}
			if ctx.Quiet {
				fmt.Fprintln(os.Stdout, result.ID)
				return nil
			}
			scopes := strings.Join(result.Scopes, ",")
			if scopes == "" {
				scopes = "(none)"
			}
			fmt.Printf("id: %s\n", result.ID)
			fmt.Printf("name: %s\n", result.Name)
			fmt.Printf("scopes: %s\n", scopes)
			return nil
		},
	}

	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(statusCmd)
	authCmd.AddCommand(whoamiCmd)
}

func tokenStatus(token string) string {
	if strings.TrimSpace(token) == "" {
		return "not set"
	}

	return "set"
}
