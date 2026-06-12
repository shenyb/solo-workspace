package config

import (
	"testing"

	core "github.com/shenyb/solo-workspace/cli/go/internal"
)

func TestConfigSetPersistsServerExtra(t *testing.T) {
	cfg := core.DefaultConfig()
	cfg.Servers = map[string]*core.ServerConfig{
		"my-vps": {Host: "1.2.3.4", User: "root", Port: 22},
	}

	raw, err := configToMap(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := setByPath(raw, "servers.my-vps.jump_host", "bastion.example.com"); err != nil {
		t.Fatal(err)
	}
	if err := mapToConfig(raw, cfg); err != nil {
		t.Fatal(err)
	}
	if got := cfg.Servers["my-vps"].Extra["jump_host"]; got != "bastion.example.com" {
		t.Fatalf("Extra[jump_host] = %v", got)
	}

	out, err := configToMap(cfg)
	if err != nil {
		t.Fatal(err)
	}
	server, ok := out["servers"].(map[string]any)["my-vps"].(map[string]any)
	if !ok {
		t.Fatal("servers.my-vps missing from map")
	}
	if server["jump_host"] != "bastion.example.com" {
		t.Fatalf("map jump_host = %v", server["jump_host"])
	}
}

func TestDelByPathNestedArrayElement(t *testing.T) {
	raw := map[string]any{
		"matrix": []any{
			[]any{"a", "b", "c"},
			[]any{"d", "e"},
		},
	}
	if err := delByPath(raw, "matrix.0.1"); err != nil {
		t.Fatal(err)
	}
	matrix := raw["matrix"].([]any)
	row0 := matrix[0].([]any)
	if len(row0) != 2 || row0[0] != "a" || row0[1] != "c" {
		t.Fatalf("matrix[0] = %v", row0)
	}
}

func TestDelByPathTopLevelArray(t *testing.T) {
	raw := map[string]any{
		"domains": []any{"a.example.com", "b.example.com", "c.example.com"},
	}
	if err := delByPath(raw, "domains.1"); err != nil {
		t.Fatal(err)
	}
	domains := raw["domains"].([]any)
	if len(domains) != 2 || domains[0] != "a.example.com" || domains[1] != "c.example.com" {
		t.Fatalf("domains = %v", domains)
	}
}

func TestMergeEmailConfigAppendsUniqueRecipients(t *testing.T) {
	dst := &core.EmailConfig{
		To: []string{"old@example.com", "shared@example.com"},
	}
	src := &core.EmailConfig{
		To: []string{"shared@example.com", "new@example.com"},
	}
	mergeEmailConfig(dst, src)
	want := []string{"old@example.com", "shared@example.com", "new@example.com"}
	if len(dst.To) != len(want) {
		t.Fatalf("To = %v, want %v", dst.To, want)
	}
	for i, addr := range want {
		if dst.To[i] != addr {
			t.Fatalf("To = %v, want %v", dst.To, want)
		}
	}
}

func TestValidateEnabledEmailRequiresFields(t *testing.T) {
	cfg := core.DefaultConfig()
	cfg.Notify.Email = &core.EmailConfig{Enabled: true}
	if err := validateEnabledEmail(cfg); err == nil {
		t.Fatal("expected error when required email fields are missing")
	}
}

func TestSetByPathUnknownEmailFieldRejected(t *testing.T) {
	cfg := core.DefaultConfig()
	cfg.Notify = &core.NotifyConfig{Email: &core.EmailConfig{}}
	raw, err := configToMap(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := setByPath(raw, "notify.email.unknown_field", "x"); err != nil {
		t.Fatal(err)
	}
	if err := mapToConfig(raw, cfg); err != nil {
		t.Fatal(err)
	}
	if err := verifyConfigPathPersisted(cfg, "notify.email.unknown_field", "x"); err == nil {
		t.Fatal("expected unknown notify.email field to be rejected")
	}
}

func TestSetByPathServerExtraPersists(t *testing.T) {
	cfg := core.DefaultConfig()
	cfg.Servers = map[string]*core.ServerConfig{
		"my-vps": {Host: "1.2.3.4", User: "root", Port: 22},
	}
	raw, err := configToMap(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := setByPath(raw, "servers.my-vps.jump_host", "bastion.example.com"); err != nil {
		t.Fatal(err)
	}
	if err := mapToConfig(raw, cfg); err != nil {
		t.Fatal(err)
	}
	if err := verifyConfigPathPersisted(cfg, "servers.my-vps.jump_host", "bastion.example.com"); err != nil {
		t.Fatalf("server Extra field should persist: %v", err)
	}
}

func TestParseImportConfigJSON(t *testing.T) {
	data := []byte(`{"domains":["a.example.com"],"projects":{"app":{"id":1,"path":"/code/app","description":"demo"}}}`)
	imported, err := parseImportConfig(data, ".json")
	if err != nil {
		t.Fatal(err)
	}
	if len(imported.Domains) != 1 || imported.Domains[0] != "a.example.com" {
		t.Fatalf("domains = %v", imported.Domains)
	}
}
