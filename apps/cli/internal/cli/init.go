package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	defaultInitDirName = ".lunie"
	composeFileName    = "docker-compose.yml"
	envFileName        = ".env"
	localDatabaseURL   = "postgresql://lunie:lunie@postgres:5432/lunie"
	localRedisURL      = "redis://redis:6379"
)

var (
	initForce               bool
	initForeground          bool
	initDatabaseURL         string
	initRedisURL            string
	initWithLocalDatastores bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a Lunie stack",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateDatastoreFlags(); err != nil {
			return err
		}

		baseDir, err := resolveInitDir()
		if err != nil {
			return err
		}

		if err := os.MkdirAll(baseDir, 0o700); err != nil {
			return err
		}

		composePath := filepath.Join(baseDir, composeFileName)
		envPath := filepath.Join(baseDir, envFileName)

		values, extras, err := readEnvFile(envPath)
		if err != nil {
			if os.IsNotExist(err) {
				values = map[string]string{}
				extras = nil
			} else {
				return err
			}
		}

		if initForce {
			values = map[string]string{}
			extras = nil
		}

		required, err := requiredEnvValues(values)
		if err != nil {
			return err
		}
		for key, val := range required {
			values[key] = val
		}

		if err := writeComposeFile(composePath, initForce, initWithLocalDatastores); err != nil {
			return err
		}

		if err := validateComposeDatastoreMode(composePath); err != nil {
			return err
		}

		if err := writeEnvFile(envPath, values, extras); err != nil {
			return err
		}

		if err := maybeUpdateConfig(values); err != nil {
			return err
		}

		if err := runDockerCompose(baseDir, composePath, "pull"); err != nil {
			return err
		}

		if initWithLocalDatastores {
			if err := runDockerCompose(baseDir, composePath, "up", "-d", "postgres", "redis"); err != nil {
				return err
			}

			if err := waitForPostgres(baseDir, composePath); err != nil {
				return err
			}
		}

		if err := runDockerCompose(
			baseDir,
			composePath,
			"run",
			"--rm",
			"--workdir",
			"/app/apps/server",
			"server",
			"node_modules/.bin/prisma",
			"migrate",
			"deploy",
		); err != nil {
			return err
		}

		upArgs := []string{"up"}
		if !initForeground {
			upArgs = append(upArgs, "-d")
		}
		if err := runDockerCompose(baseDir, composePath, upArgs...); err != nil {
			return err
		}

		fmt.Printf("Lunie initialized in %s\n", baseDir)
		fmt.Printf("Compose file: %s\n", composePath)
		fmt.Printf("Env file: %s\n", envPath)
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&initForce, "force", false, "Overwrite existing init files")
	initCmd.Flags().BoolVar(&initForeground, "foreground", false, "Run services in foreground")
	initCmd.Flags().StringVar(&initDatabaseURL, "database-url", "", "PostgreSQL connection string")
	initCmd.Flags().StringVar(&initRedisURL, "redis-url", "", "Redis connection string")
	initCmd.Flags().BoolVar(
		&initWithLocalDatastores,
		"with-local-datastores",
		false,
		"Run bundled Postgres and Redis containers",
	)
}

func resolveInitDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, defaultInitDirName), nil
}

func requiredEnvValues(existing map[string]string) (map[string]string, error) {
	adminToken := existing["LUNIE_ADMIN_TOKEN"]
	if adminToken == "" {
		adminToken = randomHex(32)
	}

	secretKey := existing["LUNIE_SECRET_KEY"]
	if secretKey == "" {
		secretKey = randomHex(32)
	}

	port := existing["PORT"]
	if strings.TrimSpace(port) == "" {
		port = "3000"
	}

	databaseURL := strings.TrimSpace(initDatabaseURL)
	redisURL := strings.TrimSpace(initRedisURL)

	if initWithLocalDatastores {
		databaseURL = localDatabaseURL
		redisURL = localRedisURL
	} else {
		if databaseURL == "" {
			databaseURL = strings.TrimSpace(existing["DATABASE_URL"])
		}
		if redisURL == "" {
			redisURL = strings.TrimSpace(existing["REDIS_URL"])
		}

		if databaseURL == "" || redisURL == "" {
			return nil, fmt.Errorf("database and redis URLs are required; pass --database-url and --redis-url, or use --with-local-datastores")
		}
	}

	if err := validateDatabaseURL(databaseURL); err != nil {
		return nil, err
	}
	if err := validateRedisURL(redisURL); err != nil {
		return nil, err
	}

	return map[string]string{
		"DATABASE_URL":      databaseURL,
		"REDIS_URL":         redisURL,
		"LUNIE_ADMIN_TOKEN": adminToken,
		"LUNIE_SECRET_KEY":  secretKey,
		"PORT":              port,
	}, nil
}

func randomHex(bytes int) string {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}

func readEnvFile(path string) (map[string]string, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	values := make(map[string]string)
	var extras []string

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			extras = append(extras, line)
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			extras = append(extras, line)
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			extras = append(extras, line)
			continue
		}
		values[key] = value
	}

	return values, extras, nil
}

func writeEnvFile(path string, values map[string]string, extras []string) error {
	order := []string{
		"DATABASE_URL",
		"REDIS_URL",
		"LUNIE_ADMIN_TOKEN",
		"LUNIE_SECRET_KEY",
		"PORT",
	}

	var lines []string
	lines = append(lines, "# Lunie local environment")
	for _, key := range order {
		if val, ok := values[key]; ok && val != "" {
			lines = append(lines, fmt.Sprintf("%s=%s", key, val))
		}
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	known := make(map[string]struct{}, len(order))
	for _, key := range order {
		known[key] = struct{}{}
	}

	for _, key := range keys {
		if _, ok := known[key]; ok {
			continue
		}
		val := values[key]
		if val == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s=%s", key, val))
	}

	if len(extras) > 0 {
		lines = append(lines, extras...)
	}

	content := strings.TrimRight(strings.Join(lines, "\n"), "\n") + "\n"
	return os.WriteFile(path, []byte(content), 0o600)
}

func writeComposeFile(path string, force bool, withLocalDatastores bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return nil
		} else if !os.IsNotExist(err) {
			return err
		}
	}

	var content string
	if withLocalDatastores {
		content = strings.TrimLeft(`name: lunie

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: lunie
      POSTGRES_PASSWORD: lunie
      POSTGRES_DB: lunie
    ports:
      - "5432:5432"
    volumes:
      - lunie_pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U lunie -d lunie"]
      interval: 5s
      timeout: 3s
      retries: 10

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - lunie_redisdata:/data
    command: ["redis-server", "--appendonly", "yes"]
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 10

  server:
    image: gentij/lunie-server:latest
    env_file:
      - .env
    ports:
      - "3000:3000"
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

  worker:
    image: gentij/lunie-worker:latest
    env_file:
      - .env
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

volumes:
  lunie_pgdata:
  lunie_redisdata:
`, "\n")
	} else {
		content = strings.TrimLeft(`name: lunie

services:
  server:
    image: gentij/lunie-server:latest
    env_file:
      - .env
    ports:
      - "3000:3000"
    restart: unless-stopped

  worker:
    image: gentij/lunie-worker:latest
    env_file:
      - .env
    restart: unless-stopped
`, "\n")
	}

	return os.WriteFile(path, []byte(content), 0o644)
}

func validateDatastoreFlags() error {
	if initWithLocalDatastores {
		if strings.TrimSpace(initDatabaseURL) != "" || strings.TrimSpace(initRedisURL) != "" {
			return fmt.Errorf("--with-local-datastores cannot be used with --database-url or --redis-url")
		}
	}
	return nil
}

func validateDatabaseURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("--database-url must be a valid Postgres connection string")
	}

	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return fmt.Errorf("--database-url must use postgres:// or postgresql://")
	}

	return nil
}

func validateRedisURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("--redis-url must be a valid Redis connection string")
	}

	if parsed.Scheme != "redis" {
		return fmt.Errorf("--redis-url must use redis://")
	}

	return nil
}

func validateComposeDatastoreMode(composePath string) error {
	data, err := os.ReadFile(composePath)
	if err != nil {
		return err
	}

	compose := string(data)
	hasPostgres := strings.Contains(compose, "\n  postgres:")
	hasRedis := strings.Contains(compose, "\n  redis:")

	if initWithLocalDatastores && (!hasPostgres || !hasRedis) {
		return fmt.Errorf("existing compose file does not define local datastores; rerun 'lunie init --with-local-datastores --force'")
	}

	if !initWithLocalDatastores && hasPostgres && hasRedis && strings.TrimSpace(initDatabaseURL) != "" && strings.TrimSpace(initRedisURL) != "" {
		return fmt.Errorf("existing compose file still includes local datastores; rerun 'lunie init --force --database-url <postgres-url> --redis-url <redis-url>' to switch modes")
	}

	return nil
}

func maybeUpdateConfig(values map[string]string) error {
	cfg, path, err := loadConfig(false)
	if err != nil {
		return err
	}

	if cfg.ServerURL == "" {
		cfg.ServerURL = defaultServerURL
	}

	if strings.TrimSpace(cfg.Token) == "" {
		if token := strings.TrimSpace(values["LUNIE_ADMIN_TOKEN"]); token != "" {
			cfg.Token = token
			return saveConfig(path, cfg)
		}
	}

	return nil
}

func runDockerCompose(baseDir string, composePath string, args ...string) error {
	if hasDockerComposePlugin() {
		fullArgs := append([]string{"compose", "-f", composePath}, args...)
		cmd := exec.Command("docker", fullArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = baseDir
		return cmd.Run()
	}

	fullArgs := append([]string{"-f", composePath}, args...)
	cmd := exec.Command("docker-compose", fullArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = baseDir
	return cmd.Run()
}

func waitForPostgres(baseDir string, composePath string) error {
	for attempt := 1; attempt <= 20; attempt++ {
		err := runDockerCompose(
			baseDir,
			composePath,
			"exec",
			"-T",
			"postgres",
			"pg_isready",
			"-U",
			"lunie",
			"-d",
			"lunie",
		)
		if err == nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("postgres did not become ready in time")
}

func hasDockerComposePlugin() bool {
	cmd := exec.Command("docker", "compose", "version")
	return cmd.Run() == nil
}
