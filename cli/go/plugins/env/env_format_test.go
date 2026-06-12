package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvValueRoundTrip(t *testing.T) {
	cases := []struct {
		name  string
		value string
	}{
		{"quotes", `hello"world`},
		{"literal_backslash_n", `hello\nworld`},
		{"actual_newline", "hello\nworld"},
		{"hash", "value#comment"},
		{"backslash", `hello\world`},
		{"equals", "a=b"},
		{"spaces", "hello world"},
		{"empty", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			encoded := formatEnvStoredValue(tc.value)
			got, err := parseEnvStoredValue(encoded)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.value {
				t.Fatalf("round-trip = %q, want %q (encoded %q)", got, tc.value, encoded)
			}
		})
	}
}

func TestEnvDBSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	e := &envDB{
		vars: map[string]string{
			"QUOTED":  `say"hello"`,
			"NEWLINE": "line1\nline2",
			"HASH":    "value#tag",
		},
		path: filepath.Join(dir, envFileName),
	}
	if err := e.save(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, envFileLegacy)); err == nil {
		t.Fatal("legacy env.yaml should not be created on fresh save")
	}

	loaded := &envDB{vars: make(map[string]string), path: e.path}
	if err := loaded.load(); err != nil {
		t.Fatal(err)
	}
	for k, want := range e.vars {
		if got := loaded.vars[k]; got != want {
			t.Fatalf("%s = %q, want %q", k, got, want)
		}
	}
}

func TestEnvDBLoadsLegacyYAMLFilename(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, envFileLegacy)
	content := "# comment\nLEGACY=\"a\\\"b\"\n"
	if err := os.WriteFile(legacy, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	e := &envDB{vars: make(map[string]string), path: filepath.Join(dir, envFileName)}
	if err := e.load(); err != nil {
		t.Fatal(err)
	}
	if got := e.vars["LEGACY"]; got != `a"b` {
		t.Fatalf("LEGACY = %q, want %q", got, `a"b`)
	}
}

func TestEnvDBMigratesLegacyFileOnSave(t *testing.T) {
	dir := t.TempDir()
	legacy := filepath.Join(dir, envFileLegacy)
	if err := os.WriteFile(legacy, []byte("OLD=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	e := &envDB{vars: make(map[string]string), path: filepath.Join(dir, envFileName)}
	if err := e.load(); err != nil {
		t.Fatal(err)
	}
	e.vars["NEW"] = "2"
	if err := e.save(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatal("legacy env.yaml should be removed after migration save")
	}
	if _, err := os.Stat(filepath.Join(dir, envFileName)); err != nil {
		t.Fatal("env.local should exist after save")
	}
}
