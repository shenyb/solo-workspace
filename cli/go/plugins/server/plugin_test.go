package server

import (
	"path/filepath"
	"testing"

	core "github.com/shenyb/solo-workspace/cli/go/internal"
)

func testConfig(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	core.ConfigPath = filepath.Join(dir, "config.yaml")
	core.CurrentConfig = core.DefaultConfig()
	if core.CurrentConfig.Servers == nil {
		core.CurrentConfig.Servers = make(map[string]*core.ServerConfig)
	}
}

func TestAddServerRejectsEmptyHost(t *testing.T) {
	testConfig(t)
	if err := addServer("mysrv", "", "root", 22); err == nil {
		t.Fatal("expected error for empty host")
	}
}

func TestAddServerRejectsEmptyUser(t *testing.T) {
	testConfig(t)
	if err := addServer("mysrv", "1.2.3.4", "", 22); err == nil {
		t.Fatal("expected error for empty user")
	}
}

func TestAddServerRejectsWhitespaceHost(t *testing.T) {
	testConfig(t)
	if err := addServer("mysrv", "   ", "root", 22); err == nil {
		t.Fatal("expected error for whitespace host")
	}
}

func TestAddServerStoresTrimmedValues(t *testing.T) {
	testConfig(t)
	if err := addServer("mysrv", " 1.2.3.4 ", " root ", 22); err != nil {
		t.Fatal(err)
	}
	srv := core.CurrentConfig.Servers["mysrv"]
	if srv.Host != "1.2.3.4" || srv.User != "root" {
		t.Fatalf("got host=%q user=%q", srv.Host, srv.User)
	}
}

func TestUpdateServerRejectsEmptyHost(t *testing.T) {
	testConfig(t)
	core.CurrentConfig.Servers["mysrv"] = &core.ServerConfig{Host: "1.2.3.4", User: "root", Port: 22}
	if err := updateServer("mysrv", "", "root", 22, false, true, false); err == nil {
		t.Fatal("expected error when clearing host")
	}
}

func TestUpdateServerRejectsInvalidPort(t *testing.T) {
	testConfig(t)
	core.CurrentConfig.Servers["mysrv"] = &core.ServerConfig{Host: "1.2.3.4", User: "root", Port: 22}
	if err := updateServer("mysrv", "", "", 0, true, false, false); err == nil {
		t.Fatal("expected error for port 0")
	}
}
