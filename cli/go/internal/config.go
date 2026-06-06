package core

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ConfigPath is the default config file location.
// Priority: ./.solo.yaml > ~/.solo/config.yaml
var ConfigPath string

// CurrentConfig is the global config loaded at startup.
// Plugins access it directly.
var CurrentConfig *Config

// Config holds all user configuration.
type Config struct {
	// Servers is a map of server name -> connection info.
	Servers map[string]*ServerConfig `yaml:"servers"`
	// Domains tracked for SSL/certificate management.
	Domains []string `yaml:"domains"`
	// Notify configures alert delivery.
	Notify *NotifyConfig `yaml:"notify"`
	// Projects lists local projects.
	Projects map[string]*ProjectConfig `yaml:"projects"`
}

// ServerConfig describes how to connect to a server.
type ServerConfig struct {
	Host string `yaml:"host"`
	User string `yaml:"user"`
	Port int    `yaml:"port"`
}

// NotifyConfig configures notification channels.
type NotifyConfig struct {
	Webhook string `yaml:"webhook"`
	Email   string `yaml:"email"`
}

// ProjectConfig describes a local project.
type ProjectConfig struct {
	Path        string `yaml:"path"`
	Description string `yaml:"description"`
	RepoURL     string `yaml:"repo_url,omitempty"`
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Servers:  make(map[string]*ServerConfig),
		Domains:  []string{},
		Notify:   &NotifyConfig{},
		Projects: make(map[string]*ProjectConfig),
	}
}

// LoadConfig loads config from the given path or finds one automatically.
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		// Try local first
		if _, err := os.Stat(".solo.yaml"); err == nil {
			path = ".solo.yaml"
		} else {
			// Try user home
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("cannot find home dir: %w", err)
			}
			path = home + "/.solo/config.yaml"
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}

// SaveConfig writes CurrentConfig back to the file it was loaded from.
// If no config file was loaded (ConfigPath is empty), it writes to ./.solo.yaml.
func SaveConfig() error {
	path := ConfigPath
	if path == "" {
		path = ".solo.yaml"
	}
	data, err := yaml.Marshal(CurrentConfig)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}
	ConfigPath = path
	return nil
}
