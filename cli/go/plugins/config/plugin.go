package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	core "github.com/shenyb/solo-workspace/cli/go/internal"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Cmd returns the config management command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration (import/export/show)",
		Long: `Import, export, and view configuration in different formats.

Supports YAML and JSON formats for flexibility in backing up and sharing configurations.

Examples:
  sw config show                           # Show current config as YAML
  sw config show --format json             # Show current config as JSON
  sw config export > backup.yaml           # Export to file
  sw config export --format json > backup.json
  sw config import backup.json             # Import configuration
  sw config validate config.json           # Validate config file
`,
	}

	cmd.AddCommand(showCmd())
	cmd.AddCommand(exportCmd())
	cmd.AddCommand(importCmd())
	cmd.AddCommand(validateCmd())
	cmd.AddCommand(setCmd())
	cmd.AddCommand(getCmd())
	cmd.AddCommand(delCmd())

	return cmd
}

// showCmd displays current configuration
func showCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")

			cfg := core.CurrentConfig
			if cfg == nil {
				cfg = core.DefaultConfig()
			}

			switch format {
			case "json":
				return showJSON(cfg)
			default:
				return showYAML(cfg)
			}
		},
	}

	cmd.Flags().StringP("format", "f", "yaml", "Output format: yaml|json")
	return cmd
}

// exportCmd exports configuration to a file
func exportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [file]",
		Short: "Export configuration to file (or stdout if no file specified)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")

			cfg := core.CurrentConfig
			if cfg == nil {
				cfg = core.DefaultConfig()
			}

			var data []byte
			var err error

			switch format {
			case "json":
				data, err = json.MarshalIndent(cfg, "", "  ")
			default:
				data, err = yaml.Marshal(cfg)
			}

			if err != nil {
				return fmt.Errorf("marshal config: %w", err)
			}

			// Write to file if specified, otherwise stdout
			if len(args) > 0 {
				if err := core.WriteFileAtomic(args[0], data, 0600); err != nil {
					return fmt.Errorf("write file: %w", err)
				}
				fmt.Printf("%s Configuration exported to %s\n", core.Success("✓"), args[0])
			} else {
				fmt.Println(string(data))
			}

			return nil
		},
	}

	cmd.Flags().StringP("format", "f", "yaml", "Output format: yaml|json")
	return cmd
}

// parseImportConfig unmarshals import file data using the appropriate parser for ext.
func parseImportConfig(data []byte, ext string) (*core.Config, error) {
	var imported *core.Config
	var parseErr error
	switch strings.ToLower(ext) {
	case ".json":
		parseErr = json.Unmarshal(data, &imported)
	case ".yaml", ".yml", "":
		parseErr = yaml.Unmarshal(data, &imported)
	default:
		parseErr = json.Unmarshal(data, &imported)
		if parseErr != nil {
			parseErr = yaml.Unmarshal(data, &imported)
		}
	}
	if parseErr != nil {
		return nil, parseErr
	}
	if imported == nil {
		return nil, fmt.Errorf("imported config is empty")
	}
	return imported, nil
}

// importCmd imports configuration from a file
func importCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import <file>",
		Short: "Import configuration from file (merges with existing)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}

			imported, parseErr := parseImportConfig(data, filepath.Ext(filePath))
			if parseErr != nil {
				return fmt.Errorf("parse config: %w", parseErr)
			}

			// Merge with existing config
			if core.CurrentConfig == nil {
				core.CurrentConfig = imported
			} else {
				mergeConfigs(core.CurrentConfig, imported)
			}

			// Re-assign IDs to avoid duplicates from imported data
			core.EnsureIDs(core.CurrentConfig)

			// Save the merged configuration
			if err := core.SaveConfig(); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Printf("%s Configuration imported from %s\n", core.Success("✓"), filePath)
			return nil
		},
	}
}

// validateCmd validates a configuration file
func validateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <file>",
		Short: "Validate configuration file syntax",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}

			var cfg *core.Config

			// Try JSON first
			err = json.Unmarshal(data, &cfg)
			if err != nil {
				// Try YAML
				err = yaml.Unmarshal(data, &cfg)
			}

			if err != nil {
				fmt.Printf("%s Configuration is %s\n", core.Warn("✗"), core.Warn("invalid"))
				fmt.Printf("Error: %v\n", err)
				return err
			}

			fmt.Printf("%s Configuration is %s\n", core.Success("✓"), core.Success("valid"))
			fmt.Printf("  Servers: %d\n", len(cfg.Servers))
			fmt.Printf("  Projects: %d\n", len(cfg.Projects))
			fmt.Printf("  Domains: %d\n", len(cfg.Domains))
			fmt.Printf("  Todos: %d\n", len(cfg.Todos))

			return nil
		},
	}
}

// Helper functions

// setCmd sets a config value at a dot-separated path
func setCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <path> <value>",
		Short: "Set a config value by path (e.g. servers.my-vps.host 1.2.3.4)",
		Long: `Dynamically modify any config value using dot-separated path notation.

Supports nested maps (servers.my-vps.host) and array indices (domains.0).

Examples:
  sw config set servers.my-vps.host 192.168.1.1
  sw config set projects.my-app.description "New description"
  sw config set notify.email.enabled true
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			value := args[1]

			cfg := core.CurrentConfig
			if cfg == nil {
				cfg = core.DefaultConfig()
			}

			// Marshal to generic map for path traversal
			raw, err := configToMap(cfg)
			if err != nil {
				return fmt.Errorf("convert config: %w", err)
			}

			if err := setByPath(raw, path, value); err != nil {
				return fmt.Errorf("set %s: %w", path, err)
			}

			// Convert back to Config and save
			if err := mapToConfig(raw, cfg); err != nil {
				return fmt.Errorf("apply changes: %w", err)
			}
			if err := verifyConfigPathPersisted(cfg, path, value); err != nil {
				return err
			}
			if path == "notify.email.enabled" && strings.EqualFold(value, "true") {
				if err := validateEnabledEmail(cfg); err != nil {
					return err
				}
			}
			core.CurrentConfig = cfg
			if err := core.SaveConfig(); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Printf("%s Config %s updated\n", core.Success("✓"), core.Info(path))
			return nil
		},
	}
}

// getCmd gets a config value at a dot-separated path
func getCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <path>",
		Short: "Get a config value by path (e.g. servers.my-vps.host)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			cfg := core.CurrentConfig
			if cfg == nil {
				cfg = core.DefaultConfig()
			}

			raw, err := configToMap(cfg)
			if err != nil {
				return fmt.Errorf("convert config: %w", err)
			}

			val, err := getByPath(raw, path)
			if err != nil {
				return fmt.Errorf("get %s: %w", path, err)
			}

			fmt.Println(val)
			return nil
		},
	}
}

// delCmd deletes a config key at a dot-separated path
func delCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <path>",
		Short: "Delete a config key by path (e.g. servers.my-vps)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			cfg := core.CurrentConfig
			if cfg == nil {
				cfg = core.DefaultConfig()
			}

			raw, err := configToMap(cfg)
			if err != nil {
				return fmt.Errorf("convert config: %w", err)
			}

			if err := delByPath(raw, path); err != nil {
				return fmt.Errorf("delete %s: %w", path, err)
			}

			if err := mapToConfig(raw, cfg); err != nil {
				return fmt.Errorf("apply changes: %w", err)
			}
			core.CurrentConfig = cfg
			if err := core.SaveConfig(); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Printf("%s Config %s deleted\n", core.Success("✓"), path)
			return nil
		},
	}
}

// ── Path traversal helpers ──────────────────────────────

// configToMap marshals a Config to a generic map via YAML round-trip.
func configToMap(cfg *core.Config) (map[string]any, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// mapToConfig unmarshals a generic map back into a Config via YAML round-trip.
func mapToConfig(raw map[string]any, cfg *core.Config) error {
	data, err := yaml.Marshal(raw)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, cfg)
}

// getByPath retrieves a value at a dot-separated path
func getByPath(data map[string]any, path string) (string, error) {
	segments := strings.Split(path, ".")
	cur := any(data)
	for i, seg := range segments {
		isLast := i == len(segments)-1
		switch v := cur.(type) {
		case map[string]any:
			val, ok := v[seg]
			if !ok {
				return "", fmt.Errorf("key %q not found at %s", seg, strings.Join(segments[:i], "."))
			}
			if isLast {
				return fmt.Sprintf("%v", val), nil
			}
			cur = val
		case []any:
			idx, err := strconv.Atoi(seg)
			if err != nil {
				return "", fmt.Errorf("expected array index, got %q at %s", seg, strings.Join(segments[:i], "."))
			}
			if idx < 0 || idx >= len(v) {
				return "", fmt.Errorf("index %d out of range (len=%d)", idx, len(v))
			}
			if isLast {
				return fmt.Sprintf("%v", v[idx]), nil
			}
			cur = v[idx]
		default:
			return "", fmt.Errorf("cannot descend into %T at %s", cur, strings.Join(segments[:i], "."))
		}
	}
	return "", fmt.Errorf("empty path")
}

// setByPath sets a value at a dot-separated path
func setByPath(data map[string]any, path, value string) error {
	segments := strings.Split(path, ".")
	if len(segments) == 0 {
		return fmt.Errorf("empty path")
	}
	return setByPathRecursive(data, segments, value)
}

func setByPathRecursive(data map[string]any, segments []string, value string) error {
	key := segments[0]
	if len(segments) == 1 {
		// Leaf: set the value, auto-detect type
		data[key] = parseValue(value)
		return nil
	}

	// Intermediate: navigate or create
	next, exists := data[key]
	if !exists {
		// Auto-create nested map
		data[key] = make(map[string]any)
		next = data[key]
	}

	switch v := next.(type) {
	case map[string]any:
		return setByPathRecursive(v, segments[1:], value)
	case []any:
		idx, err := strconv.Atoi(segments[1])
		if err != nil {
			return fmt.Errorf("expected array index, got %q", segments[1])
		}
		if idx < 0 || idx >= len(v) {
			return fmt.Errorf("index %d out of range", idx)
		}
		if len(segments) == 2 {
			v[idx] = parseValue(value)
			return nil
		}
		// Descend into array element
		if m, ok := v[idx].(map[string]any); ok {
			return setByPathRecursive(m, segments[2:], value)
		}
		return fmt.Errorf("cannot descend into %T at index %d", v[idx], idx)
	default:
		return fmt.Errorf("cannot descend into %T at %q", next, key)
	}
}

// delByPath deletes a key at a dot-separated path
func delByPath(data map[string]any, path string) error {
	segments := strings.Split(path, ".")
	if len(segments) == 0 {
		return fmt.Errorf("empty path")
	}

	var assignParent []func(any)
	cur := any(data)

	for i := 0; i < len(segments)-1; i++ {
		seg := segments[i]
		switch v := cur.(type) {
		case map[string]any:
			next, ok := v[seg]
			if !ok {
				return fmt.Errorf("key %q not found", seg)
			}
			key := seg
			m := v
			assignParent = append(assignParent, func(val any) { m[key] = val })
			cur = next
		case []any:
			idx, err := strconv.Atoi(seg)
			if err != nil {
				return fmt.Errorf("expected array index, got %q", seg)
			}
			if idx < 0 || idx >= len(v) {
				return fmt.Errorf("index %d out of range", idx)
			}
			s := v
			at := idx
			assignParent = append(assignParent, func(val any) { s[at] = val })
			cur = s[at]
		default:
			return fmt.Errorf("cannot descend into %T", cur)
		}
	}

	lastSeg := segments[len(segments)-1]
	switch v := cur.(type) {
	case map[string]any:
		if _, ok := v[lastSeg]; !ok {
			return fmt.Errorf("key %q not found", lastSeg)
		}
		delete(v, lastSeg)
		return nil
	case []any:
		idx, err := strconv.Atoi(lastSeg)
		if err != nil {
			return fmt.Errorf("expected array index, got %q", lastSeg)
		}
		if idx < 0 || idx >= len(v) {
			return fmt.Errorf("index %d out of range", idx)
		}
		newSlice := append(v[:idx], v[idx+1:]...)
		if len(assignParent) == 0 {
			return fmt.Errorf("cannot delete array element at root")
		}
		assignParent[len(assignParent)-1](newSlice)
		return nil
	default:
		return fmt.Errorf("cannot delete from %T", cur)
	}
}

// parseValue auto-detects and converts a string value to bool/int/float/string
func parseValue(s string) any {
	// bool
	switch strings.ToLower(s) {
	case "true":
		return true
	case "false":
		return false
	}
	// int
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	// float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}

// verifyConfigPathPersisted ensures setByPath values survive the YAML round-trip into Config.
func verifyConfigPathPersisted(cfg *core.Config, path, value string) error {
	raw, err := configToMap(cfg)
	if err != nil {
		return fmt.Errorf("verify config: %w", err)
	}
	got, err := getByPath(raw, path)
	if err != nil {
		return fmt.Errorf("config path %q is not recognized: %w", path, err)
	}
	if !configValuesEqual(parseValue(value), got) {
		return fmt.Errorf("config path %q is not recognized; value was not saved", path)
	}
	return nil
}

func configValuesEqual(want any, got string) bool {
	switch v := want.(type) {
	case bool:
		return got == fmt.Sprintf("%v", v)
	case int:
		return got == fmt.Sprintf("%d", v)
	case float64:
		return got == fmt.Sprintf("%g", v)
	default:
		return got == fmt.Sprintf("%v", v)
	}
}

func validateEnabledEmail(cfg *core.Config) error {
	if cfg == nil || cfg.Notify == nil || cfg.Notify.Email == nil {
		return fmt.Errorf("notify.email.enabled requires email configuration; set notify.email.host, from, and to first")
	}
	email := cfg.Notify.Email
	var missing []string
	if email.Host == "" {
		missing = append(missing, "host")
	}
	if email.From == "" {
		missing = append(missing, "from")
	}
	if len(email.To) == 0 {
		missing = append(missing, "to")
	}
	if len(missing) > 0 {
		return fmt.Errorf(
			"notify.email.enabled is true but missing required field(s): %s (use sw config set notify.email.<field> ...)",
			strings.Join(missing, ", "),
		)
	}
	return nil
}

func showYAML(cfg *core.Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func showJSON(cfg *core.Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// mergeConfigs merges src into dst (src overwrites existing keys)
func mergeConfigs(dst, src *core.Config) {
	if src.Servers != nil {
		if dst.Servers == nil {
			dst.Servers = make(map[string]*core.ServerConfig)
		}
		for k, v := range src.Servers {
			if existing, ok := dst.Servers[k]; ok && existing != nil && v != nil {
				mergeServerConfig(existing, v)
			} else {
				dst.Servers[k] = v
			}
		}
	}

	// Merge domains (append unique)
	if src.Domains != nil {
		domainMap := make(map[string]bool)
		for _, d := range dst.Domains {
			domainMap[d] = true
		}
		for _, d := range src.Domains {
			if !domainMap[d] {
				dst.Domains = append(dst.Domains, d)
			}
		}
	}

	if src.Projects != nil {
		if dst.Projects == nil {
			dst.Projects = make(map[string]*core.ProjectConfig)
		}
		for k, v := range src.Projects {
			if existing, ok := dst.Projects[k]; ok && existing != nil && v != nil {
				mergeProjectConfig(existing, v)
			} else {
				dst.Projects[k] = v
			}
		}
	}

	if src.Todos != nil {
		if dst.Todos == nil {
			dst.Todos = make(map[string]*core.TodoConfig)
		}
		for k, v := range src.Todos {
			if existing, ok := dst.Todos[k]; ok && existing != nil && v != nil {
				mergeTodoConfig(existing, v)
			} else {
				dst.Todos[k] = v
			}
		}
	}

	if src.Notify != nil {
		if dst.Notify == nil {
			dst.Notify = src.Notify
		} else {
			if src.Notify.Webhook != "" {
				dst.Notify.Webhook = src.Notify.Webhook
			}
			if src.Notify.Email != nil {
				if dst.Notify.Email == nil {
					dst.Notify.Email = src.Notify.Email
				} else {
					mergeEmailConfig(dst.Notify.Email, src.Notify.Email)
				}
			}
		}
	}
}

func mergeServerConfig(dst, src *core.ServerConfig) {
	if src.Host != "" {
		dst.Host = src.Host
	}
	if src.User != "" {
		dst.User = src.User
	}
	if src.Port != 0 {
		dst.Port = src.Port
	}
	if src.Extra != nil {
		if dst.Extra == nil {
			dst.Extra = make(map[string]interface{})
		}
		for k, v := range src.Extra {
			dst.Extra[k] = v
		}
	}
}

func mergeProjectConfig(dst, src *core.ProjectConfig) {
	if src.ID != 0 {
		dst.ID = src.ID
	}
	if src.Path != "" {
		dst.Path = src.Path
	}
	if src.Description != "" {
		dst.Description = src.Description
	}
	if src.RepoURL != "" {
		dst.RepoURL = src.RepoURL
	}
	if src.Extra != nil {
		if dst.Extra == nil {
			dst.Extra = make(map[string]interface{})
		}
		for k, v := range src.Extra {
			dst.Extra[k] = v
		}
	}
}

func mergeTodoConfig(dst, src *core.TodoConfig) {
	if src.ID != 0 {
		dst.ID = src.ID
	}
	if src.Description != "" {
		dst.Description = src.Description
	}
	if src.Note != "" {
		dst.Note = src.Note
	}
	dst.Done = src.Done || dst.Done
	if !src.CreatedAt.IsZero() {
		dst.CreatedAt = src.CreatedAt
	}
	if !src.UpdatedAt.IsZero() {
		dst.UpdatedAt = src.UpdatedAt
	}
}

func mergeEmailConfig(dst, src *core.EmailConfig) {
	if src.Host != "" {
		dst.Host = src.Host
	}
	if src.Port != 0 {
		dst.Port = src.Port
	}
	if src.Username != "" {
		dst.Username = src.Username
	}
	if src.Password != "" {
		dst.Password = src.Password
	}
	if src.PasswordSecret != "" {
		dst.PasswordSecret = src.PasswordSecret
	}
	if src.From != "" {
		dst.From = src.From
	}
	if len(src.To) > 0 {
		seen := make(map[string]struct{}, len(dst.To)+len(src.To))
		merged := make([]string, 0, len(dst.To)+len(src.To))
		for _, addr := range dst.To {
			if _, ok := seen[addr]; ok {
				continue
			}
			seen[addr] = struct{}{}
			merged = append(merged, addr)
		}
		for _, addr := range src.To {
			if _, ok := seen[addr]; ok {
				continue
			}
			seen[addr] = struct{}{}
			merged = append(merged, addr)
		}
		dst.To = merged
	}
	dst.Enabled = src.Enabled || dst.Enabled
	dst.UseTLS = src.UseTLS || dst.UseTLS
}
