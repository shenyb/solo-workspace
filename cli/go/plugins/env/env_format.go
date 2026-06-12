package env

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	envFileName   = "env.local"
	envFileLegacy = "env.yaml"
)

// envValueNeedsQuoting reports whether a value must be double-quoted on disk.
func envValueNeedsQuoting(value string) bool {
	if value == "" {
		return true
	}
	return strings.ContainsAny(value, " \t\n\r#=\\\"")
}

// formatEnvStoredValue returns the on-disk representation of value.
func formatEnvStoredValue(value string) string {
	if envValueNeedsQuoting(value) {
		return strconv.Quote(value)
	}
	return value
}

// parseEnvStoredValue decodes a value read from the dotenv store.
func parseEnvStoredValue(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		unquoted, err := strconv.Unquote(raw)
		if err != nil {
			return "", fmt.Errorf("parse quoted env value: %w", err)
		}
		return unquoted, nil
	}
	if len(raw) >= 2 && raw[0] == '\'' && raw[len(raw)-1] == '\'' {
		return raw[1 : len(raw)-1], nil
	}
	return raw, nil
}
