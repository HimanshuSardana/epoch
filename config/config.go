package config

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	AI      AIConfig     `toml:"ai"`
	Squash  SquashConfig `toml:"squash"`
	Commits CommitConfig `toml:"commits"`
}

type AIConfig struct {
	Provider    string  `toml:"provider"`
	Model       string  `toml:"model"`
	Temperature float64 `toml:"temperature"`
}

type SquashConfig struct {
	WindowMinutes int  `toml:"window_minutes"`
	AlwaysPreview bool `toml:"always_preview"`
}

type CommitConfig struct {
	EnforceConventional bool     `toml:"enforce_conventional"`
	AllowedTypes        []string `toml:"allowed_types"`
}

func DefaultConfig() *Config {
	return &Config{
		AI: AIConfig{
			Provider:    "cerebras",
			Model:       "llama3.1-8b",
			Temperature: 0.3,
		},
		Squash: SquashConfig{
			WindowMinutes: 10,
			AlwaysPreview: true,
		},
		Commits: CommitConfig{
			EnforceConventional: true,
			AllowedTypes:        []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "chore", "build", "ci", "revert"},
		},
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.AI.Provider == "" {
		cfg.AI = DefaultConfig().AI
	}
	if cfg.Squash.WindowMinutes == 0 {
		cfg.Squash = DefaultConfig().Squash
	}
	if len(cfg.Commits.AllowedTypes) == 0 {
		cfg.Commits = DefaultConfig().Commits
	}

	return &cfg, nil
}

func (c *Config) Save(path string) error {
	data, err := toml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
