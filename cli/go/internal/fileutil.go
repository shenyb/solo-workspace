package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// WriteFileAtomic writes data to path via a temporary file and rename.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create dir %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, ".sw-tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Chmod(perm); err != nil {
		cleanup()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		cleanup()
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}

// ParseYAMLInt coerces YAML/JSON dynamic values to int.
func ParseYAMLInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(n))
		if err != nil {
			return 0, false
		}
		return i, true
	default:
		return 0, false
	}
}

// ParseDurationDays parses Go durations plus day suffixes (e.g. 3d, 7d).
func ParseDurationDays(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil || days < 0 {
			return 0, fmt.Errorf("invalid day duration %q", s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}

// SecretResolver resolves secret names from the encrypted vault.
// Registered by the secret plugin at init time.
var SecretResolver func(name string) (string, error)

// NotifyAlert sends notifications via configured webhook and/or email.
func NotifyAlert(subject, body string) error {
	if CurrentConfig == nil || CurrentConfig.Notify == nil {
		return fmt.Errorf("notification is not configured")
	}

	hasWebhook := CurrentConfig.Notify.Webhook != ""
	hasEmail := CurrentConfig.Notify.Email != nil && CurrentConfig.Notify.Email.Enabled
	if !hasWebhook && !hasEmail {
		return fmt.Errorf("no notification channel enabled (webhook or email)")
	}

	var errs []string
	if hasWebhook {
		if err := SendWebhook(subject, body); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if hasEmail {
		if err := SendEmail(subject, body); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("notification failed: %s", strings.Join(errs, "; "))
	}
	return nil
}
