package env

import (
	"strings"
	"testing"
)

func TestExportVaultLinesMasksNonSecretPrefixByDefault(t *testing.T) {
	names := []string{"secret_db_password", "my_api_key", "smtp_password"}
	values := map[string]string{
		"secret_db_password": "p1",
		"my_api_key":         "sk_live_123",
		"smtp_password":      "p2",
	}
	lines, err := exportVaultLines(names, func(name string) (string, error) {
		return values[name], nil
	}, "", false, false)
	if err != nil {
		t.Fatal(err)
	}
	out := strings.Join(lines, "\n")
	if strings.Contains(out, "sk_live_123") || strings.Contains(out, "p1") || strings.Contains(out, "p2") {
		t.Fatalf("vault secrets leaked in default export:\n%s", out)
	}
	for _, want := range []string{"my_api_key=***masked***", "secret_db_password=***masked***"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestExportVaultLinesUnmasked(t *testing.T) {
	lines, err := exportVaultLines([]string{"my_api_key"}, func(name string) (string, error) {
		return "sk_live_123", nil
	}, "", false, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 || lines[0] != "my_api_key=sk_live_123" {
		t.Fatalf("got %v", lines)
	}
}

func TestExportVaultLinesMaskedFlag(t *testing.T) {
	lines, err := exportVaultLines([]string{"my_api_key"}, func(name string) (string, error) {
		return "sk_live_123", nil
	}, "", true, true)
	if err != nil {
		t.Fatal(err)
	}
	if lines[0] != "my_api_key=***masked***" {
		t.Fatalf("got %v", lines)
	}
}
