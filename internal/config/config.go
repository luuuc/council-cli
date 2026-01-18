package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	CouncilDir  = ".council"
	ConfigFile  = "config.yaml"
	ExpertsDir  = "experts"
	CommandsDir = "commands"
)

// Config represents the council configuration
type Config struct {
	Version int           `yaml:"version"`
	AI      AIConfig      `yaml:"ai"`
	Targets []string      `yaml:"targets"`
	Council CouncilConfig `yaml:"council"`
}

// AIConfig holds AI CLI configuration
type AIConfig struct {
	Command string   `yaml:"command"`
	Args    []string `yaml:"args,omitempty"`
	Timeout int      `yaml:"timeout"`
}

// CouncilConfig holds council generation options
type CouncilConfig struct {
	// Commands is the list of slash commands to generate
	// Valid values: council, council-add, council-detect
	Commands []string `yaml:"commands"`
}

// HasCommand checks if a command is enabled
func (c *CouncilConfig) HasCommand(name string) bool {
	for _, cmd := range c.Commands {
		if cmd == name {
			return true
		}
	}
	return false
}

// Default returns a default configuration
func Default() *Config {
	return &Config{
		Version: 1,
		AI: AIConfig{
			Command: "claude",
			Timeout: 120,
		},
		Targets: []string{"claude", "cursor", "opencode"},
		Council: CouncilConfig{
			Commands: []string{"council", "council-add", "council-detect"},
		},
	}
}

// Path returns the full path to a council file or directory
func Path(parts ...string) string {
	all := append([]string{CouncilDir}, parts...)
	return filepath.Join(all...)
}

// Exists checks if the council directory exists
func Exists() bool {
	info, err := os.Stat(CouncilDir)
	return err == nil && info.IsDir()
}

// Load loads the configuration from .council/config.yaml
func Load() (*Config, error) {
	path := Path(ConfigFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("council not initialized: run 'council init' first")
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply defaults for missing values
	cfg.applyDefaults()

	return &cfg, nil
}

// applyDefaults fills in missing configuration with sensible defaults
func (c *Config) applyDefaults() {
	defaults := Default()

	// If no commands specified, use defaults
	if len(c.Council.Commands) == 0 {
		c.Council.Commands = defaults.Council.Commands
	}

	// If no targets specified, use defaults
	if len(c.Targets) == 0 {
		c.Targets = defaults.Targets
	}

	// If no AI command specified, use default
	if c.AI.Command == "" {
		c.AI.Command = defaults.AI.Command
	}

	// If no timeout specified, use default
	if c.AI.Timeout == 0 {
		c.AI.Timeout = defaults.AI.Timeout
	}
}

// Save saves the configuration to .council/config.yaml
func (c *Config) Save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(Path(ConfigFile), data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
