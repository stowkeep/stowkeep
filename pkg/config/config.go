// Package config loads Stowkeep settings from environment variables.
package config

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

// Config holds runtime configuration for the Stowkeep server.
type Config struct {
	HTTPAddr       string `envconfig:"STOWKEEP_HTTP_ADDR" default:":8080"`
	LogLevel       string `envconfig:"STOWKEEP_LOG_LEVEL" default:"info"`
	LogFormat      string `envconfig:"STOWKEEP_LOG_FORMAT" default:"json"`
	LogAddSource   bool   `envconfig:"STOWKEEP_LOG_ADD_SOURCE" default:"false"`
	DatabaseDriver string `envconfig:"STOWKEEP_DATABASE_DRIVER"`
	DatabasePath   string `envconfig:"STOWKEEP_DATABASE_PATH" default:"/data/stowkeep.db"`
	DatabaseURL    string `envconfig:"STOWKEEP_DATABASE_URL"`
	DockerHost     string `envconfig:"STOWKEEP_DOCKER_HOST" default:"unix:///var/run/docker.sock"`
	MasterKey      string `envconfig:"STOWKEEP_MASTER_KEY"`
	Features       string `envconfig:"STOWKEEP_FEATURES"`
	MigrationsDir  string `envconfig:"STOWKEEP_MIGRATIONS_DIR" default:"migrations"`
	Version        string `envconfig:"STOWKEEP_VERSION" default:"dev"`
}

// Load reads configuration from the environment.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL != "" {
		return nil
	}
	driver := strings.ToLower(strings.TrimSpace(c.DatabaseDriver))
	if driver == "" || driver == "sqlite" {
		return nil
	}
	if driver == "postgres" {
		return fmt.Errorf("STOWKEEP_DATABASE_DRIVER=postgres requires STOWKEEP_DATABASE_URL")
	}
	return fmt.Errorf("unsupported STOWKEEP_DATABASE_DRIVER %q", c.DatabaseDriver)
}

// ResolvedDriver returns the effective database driver name.
func (c *Config) ResolvedDriver() string {
	if c.DatabaseURL != "" {
		lower := strings.ToLower(c.DatabaseURL)
		if strings.HasPrefix(lower, "postgres://") || strings.HasPrefix(lower, "postgresql://") {
			return "postgres"
		}
		if strings.HasPrefix(lower, "sqlite://") {
			return "sqlite"
		}
	}
	driver := strings.ToLower(strings.TrimSpace(c.DatabaseDriver))
	if driver == "" {
		return "sqlite"
	}
	return driver
}

// ResolvedSQLitePath returns the SQLite database file path.
func (c *Config) ResolvedSQLitePath() string {
	if c.DatabaseURL != "" && strings.HasPrefix(strings.ToLower(c.DatabaseURL), "sqlite://") {
		return strings.TrimPrefix(c.DatabaseURL, "sqlite://")
	}
	if c.DatabasePath != "" {
		return c.DatabasePath
	}
	return "/data/stowkeep.db"
}

// FeatureSet returns enabled feature flags as a set.
func (c *Config) FeatureSet() map[string]struct{} {
	out := make(map[string]struct{})
	for _, part := range strings.Split(c.Features, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out[part] = struct{}{}
		}
	}
	return out
}
