package cli

import (
	"fmt"
	"strings"

	"github.com/gentij/lunie/apps/cli/internal/config"
)

func loadConfig(serverFlagChanged bool) (config.Config, string, error) {
	path := config.ResolvePath(configPath)
	cfg, err := config.Load(path)
	if err != nil {
		return config.Config{}, path, err
	}

	cfg.ServerURL = resolveServerURL(cfg.ServerURL, serverURL, serverFlagChanged)

	return cfg, path, nil
}

func resolveServerURL(configServerURL string, flagServerURL string, flagChanged bool) string {
	if flagChanged {
		return strings.TrimSpace(flagServerURL)
	}

	if strings.TrimSpace(configServerURL) != "" {
		return strings.TrimSpace(configServerURL)
	}

	return strings.TrimSpace(flagServerURL)
}

func saveConfig(path string, cfg config.Config) error {
	if cfg.ServerURL == "" {
		return fmt.Errorf("server URL is required")
	}

	return config.Save(path, cfg)
}
