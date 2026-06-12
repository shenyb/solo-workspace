package core

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestSaveConfigCreatesParentDirectory(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".solo", "config.yaml")

	ConfigPath = cfgPath
	CurrentConfig = DefaultConfig()
	CurrentConfig.Servers = map[string]*ServerConfig{
		"vps": {Host: "1.2.3.4", User: "root", Port: 22},
	}

	if err := SaveConfig(); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	if _, err := os.Stat(cfgPath); err != nil {
		t.Fatalf("config file not created: %v", err)
	}
	if ConfigPath != cfgPath {
		t.Fatalf("ConfigPath = %q, want %q", ConfigPath, cfgPath)
	}
}

func TestServerConfigExtraRoundTrip(t *testing.T) {
	in := `servers:
  my-vps:
    host: 1.2.3.4
    user: root
    port: 22
    jump_host: bastion.example.com
`
	cfg := DefaultConfig()
	if err := yaml.Unmarshal([]byte(in), cfg); err != nil {
		t.Fatal(err)
	}
	srv := cfg.Servers["my-vps"]
	if srv == nil {
		t.Fatal("server not loaded")
	}
	if got := srv.Extra["jump_host"]; got != "bastion.example.com" {
		t.Fatalf("Extra[jump_host] = %v", got)
	}

	out, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	reloaded := DefaultConfig()
	if err := yaml.Unmarshal(out, reloaded); err != nil {
		t.Fatal(err)
	}
	got := reloaded.Servers["my-vps"].Extra["jump_host"]
	if got != "bastion.example.com" {
		t.Fatalf("after round-trip Extra[jump_host] = %v", got)
	}
}

func TestProjectConfigExtraRoundTrip(t *testing.T) {
	in := `projects:
  app:
    id: 1
    path: /code/app
    description: demo
    stack: go
`
	cfg := DefaultConfig()
	if err := yaml.Unmarshal([]byte(in), cfg); err != nil {
		t.Fatal(err)
	}
	proj := cfg.Projects["app"]
	if proj.Extra["stack"] != "go" {
		t.Fatalf("Extra[stack] = %v", proj.Extra["stack"])
	}

	out, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	reloaded := DefaultConfig()
	if err := yaml.Unmarshal(out, reloaded); err != nil {
		t.Fatal(err)
	}
	if reloaded.Projects["app"].Extra["stack"] != "go" {
		t.Fatalf("after round-trip Extra[stack] = %v", reloaded.Projects["app"].Extra["stack"])
	}
}

func TestSaveConfigPreservesServerExtra(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	ConfigPath = cfgPath
	CurrentConfig = DefaultConfig()
	CurrentConfig.Servers = map[string]*ServerConfig{
		"vps": {
			Host: "1.2.3.4",
			User: "root",
			Port: 22,
			Extra: map[string]interface{}{
				"jump_host": "bastion.example.com",
			},
		},
	}
	if err := SaveConfig(); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := loaded.Servers["vps"].Extra["jump_host"]; got != "bastion.example.com" {
		t.Fatalf("Extra[jump_host] = %v", got)
	}
}
