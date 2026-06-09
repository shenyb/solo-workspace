package core

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

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
	// Todos holds task management items.
	Todos map[string]*TodoConfig `yaml:"todos"`
}

// ServerConfig describes how to connect to a server.
type ServerConfig struct {
	Host  string                 `yaml:"host"`
	User  string                 `yaml:"user"`
	Port  int                    `yaml:"port"`
	Extra map[string]interface{} `yaml:"-"` // Dynamic fields
}

// UnmarshalYAML custom unmarshaler for ServerConfig to capture extra fields
func (s *ServerConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	if host, ok := raw["host"].(string); ok {
		s.Host = host
	}
	if user, ok := raw["user"].(string); ok {
		s.User = user
	}
	if port, ok := raw["port"].(int); ok {
		s.Port = port
	}

	s.Extra = make(map[string]interface{})
	for k, v := range raw {
		if k != "host" && k != "user" && k != "port" {
			s.Extra[k] = v
		}
	}
	return nil
}

// MarshalYAML custom marshaler for ServerConfig to include extra fields
func (s *ServerConfig) MarshalYAML() (interface{}, error) {
	m := map[string]interface{}{
		"host": s.Host,
		"user": s.User,
		"port": s.Port,
	}
	// Merge extra fields
	for k, v := range s.Extra {
		m[k] = v
	}
	return m, nil
}

// EmailConfig configures SMTP email delivery.
type EmailConfig struct {
	Enabled  bool     `yaml:"enabled,omitempty"`
	Host     string   `yaml:"host"`
	Port     int      `yaml:"port"`
	Username string   `yaml:"username,omitempty"`
	Password string   `yaml:"password,omitempty"`
	From     string   `yaml:"from"`
	To       []string `yaml:"to"`
	UseTLS   bool     `yaml:"use_tls,omitempty"`
}

// NotifyConfig configures notification channels.
type NotifyConfig struct {
	Webhook string       `yaml:"webhook"`
	Email   *EmailConfig `yaml:"email"`
}

// TodoConfig describes a todo item.
type TodoConfig struct {
	ID          int       `yaml:"id,omitempty"`
	Description string    `yaml:"description,omitempty"`
	Done        bool      `yaml:"done,omitempty"`
	CreatedAt   time.Time `yaml:"created_at,omitempty"`
	UpdatedAt   time.Time `yaml:"updated_at,omitempty"`
}

// ProjectConfig describes a local project.
type ProjectConfig struct {
	ID          int                    `yaml:"id,omitempty"`
	Path        string                 `yaml:"path"`
	Description string                 `yaml:"description"`
	RepoURL     string                 `yaml:"repo_url,omitempty"`
	Extra       map[string]interface{} `yaml:"-"` // Dynamic fields
}

// UnmarshalYAML custom unmarshaler for ProjectConfig to capture extra fields
func (p *ProjectConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	if id, ok := raw["id"].(int); ok {
		p.ID = id
	}
	if path, ok := raw["path"].(string); ok {
		p.Path = path
	}
	if desc, ok := raw["description"].(string); ok {
		p.Description = desc
	}
	if repo, ok := raw["repo_url"].(string); ok {
		p.RepoURL = repo
	}

	p.Extra = make(map[string]interface{})
	for k, v := range raw {
		if k != "id" && k != "path" && k != "description" && k != "repo_url" {
			p.Extra[k] = v
		}
	}
	return nil
}

// MarshalYAML custom marshaler for ProjectConfig to include extra fields
func (p *ProjectConfig) MarshalYAML() (interface{}, error) {
	m := map[string]interface{}{
		"id":          p.ID,
		"path":        p.Path,
		"description": p.Description,
	}
	if p.RepoURL != "" {
		m["repo_url"] = p.RepoURL
	}
	// Merge extra fields
	for k, v := range p.Extra {
		m[k] = v
	}
	return m, nil
}

// DataDir returns the directory for env.yaml, secrets.enc, etc.
// Files live alongside the active config file (same directory).
func DataDir() (string, error) {
	path := ConfigPath
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot find home dir: %w", err)
		}
		return filepath.Join(home, ".solo"), nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve config path: %w", err)
	}
	return filepath.Dir(absPath), nil
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Servers:  make(map[string]*ServerConfig),
		Domains:  []string{},
		Notify:   &NotifyConfig{},
		Projects: make(map[string]*ProjectConfig),
		Todos:    make(map[string]*TodoConfig),
	}
}

// LoadConfig loads config from the given path or finds one automatically.
// Priority: -c flag > ~/.solo/config.yaml > ./.solo.yaml > defaults
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		// Try user home first
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot find home dir: %w", err)
		}
		homePath := home + "/.solo/config.yaml"
		if _, err := os.Stat(homePath); err == nil {
			path = homePath
		} else if _, err := os.Stat(".solo.yaml"); err == nil {
			// Fall back to local
			path = ".solo.yaml"
		} else {
			path = homePath // will hit os.IsNotExist below → DefaultConfig
		}
	}

	ConfigPath = path
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
	EnsureIDs(cfg)
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

// UnmarshalYAML implements custom unmarshalling for EmailConfig to accept
// either a mapping (full config) or a scalar string (recipient address).
func (e *EmailConfig) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var s string
		if err := value.Decode(&s); err != nil {
			return err
		}
		// Interpret a scalar string as a single recipient address.
		e.Enabled = true
		e.To = []string{s}
		return nil
	case yaml.MappingNode:
		type alias EmailConfig
		var a alias
		if err := value.Decode(&a); err != nil {
			return err
		}
		*e = EmailConfig(a)
		return nil
	default:
		return fmt.Errorf("unsupported yaml node for EmailConfig: %v", value.Kind)
	}
}
