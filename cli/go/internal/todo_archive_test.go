package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestArchiveStaleTodosPreservesNilEntry(t *testing.T) {
	dir := t.TempDir()
	ConfigPath = filepath.Join(dir, "config.yaml")

	cfg := DefaultConfig()
	cfg.Todos = map[string]*TodoConfig{
		"broken": nil,
	}

	names, err := ArchiveStaleTodos(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "broken" {
		t.Fatalf("archived = %v, want [broken]", names)
	}

	archive, err := LoadTodoArchive()
	if err != nil {
		t.Fatal(err)
	}
	if archive.Todos["broken"] == nil {
		t.Fatal("nil todo must be written to archive")
	}
	if archive.Todos["broken"].ArchivedAt.IsZero() {
		t.Fatal("archived entry must have archived_at")
	}

	if _, ok := cfg.Todos["broken"]; ok {
		t.Fatal("nil todo should be removed from active config after archive")
	}
}

func TestIsTodoStaleZeroTimestampNotStale(t *testing.T) {
	if IsTodoStale(&TodoConfig{}, TodoArchiveMaxAge) {
		t.Fatal("todo without timestamps must not be auto-archived")
	}
}

func TestIsTodoStaleNilIsStale(t *testing.T) {
	if !IsTodoStale(nil, TodoArchiveMaxAge) {
		t.Fatal("nil todo must be eligible for archival cleanup")
	}
}

func TestLegacyTodoWithoutTimestampsNotArchived(t *testing.T) {
	dir := t.TempDir()
	ConfigPath = filepath.Join(dir, "config.yaml")

	cfg := DefaultConfig()
	cfg.Todos = map[string]*TodoConfig{
		"legacy": {ID: 1, Description: "old task", Done: true},
	}

	names, err := ArchiveStaleTodos(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Fatalf("legacy todo without timestamps should not archive, got %v", names)
	}
	if _, ok := cfg.Todos["legacy"]; !ok {
		t.Fatal("legacy todo should remain in active config")
	}
}

func TestArchiveStaleTodosNilEntryNotLostOnSaveFailure(t *testing.T) {
	dir := t.TempDir()
	ConfigPath = filepath.Join(dir, "config.yaml")

	archivePath, err := TodoArchivePath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(archivePath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(archivePath, []byte(":\n  invalid: [\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := DefaultConfig()
	cfg.Todos = map[string]*TodoConfig{"broken": nil}

	_, err = ArchiveStaleTodos(cfg)
	if err == nil {
		t.Fatal("expected save failure on corrupt archive file")
	}
	if _, ok := cfg.Todos["broken"]; !ok {
		t.Fatal("active config must keep nil todo when archive save fails")
	}
}

func TestArchiveStaleTodosSkipsPending(t *testing.T) {
	dir := t.TempDir()
	ConfigPath = filepath.Join(dir, "config.yaml")

	old := time.Now().Add(-TodoArchiveMaxAge - time.Hour)
	cfg := DefaultConfig()
	cfg.Todos = map[string]*TodoConfig{
		"pending-old": {ID: 1, Description: "still open", Done: false, CreatedAt: old, UpdatedAt: old},
		"done-old":    {ID: 2, Description: "finished", Done: true, CreatedAt: old, UpdatedAt: old},
	}

	names, err := ArchiveStaleTodos(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "done-old" {
		t.Fatalf("archived = %v, want [done-old]", names)
	}
	if _, ok := cfg.Todos["pending-old"]; !ok {
		t.Fatal("pending todo must remain active")
	}
	if _, ok := cfg.Todos["done-old"]; ok {
		t.Fatal("done stale todo should be removed from active config")
	}
}

func TestRestoreArchivedTodo(t *testing.T) {
	dir := t.TempDir()
	ConfigPath = filepath.Join(dir, "config.yaml")

	old := time.Now().Add(-TodoArchiveMaxAge - time.Hour)
	cfg := DefaultConfig()
	cfg.Todos = map[string]*TodoConfig{
		"done-old": {ID: 5, Description: "finished", Done: true, Note: "n1", CreatedAt: old, UpdatedAt: old},
	}

	if _, err := ArchiveStaleTodos(cfg); err != nil {
		t.Fatal(err)
	}

	name, err := RestoreArchivedTodo(cfg, 5)
	if err != nil {
		t.Fatal(err)
	}
	if name != "done-old" {
		t.Fatalf("name = %q, want done-old", name)
	}

	todo := cfg.Todos["done-old"]
	if todo == nil {
		t.Fatal("restored todo missing from active config")
	}
	if todo.ID != 5 || todo.Description != "finished" || !todo.Done || todo.Note != "n1" {
		t.Fatalf("unexpected restored todo: %+v", todo)
	}
	if todo.UpdatedAt.Before(time.Now().Add(-time.Minute)) {
		t.Fatal("restore should refresh updated_at")
	}

	archive, err := LoadTodoArchive()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := archive.Todos["done-old"]; ok {
		t.Fatal("restored todo should be removed from archive")
	}
}

func TestRestoreArchivedTodoNotFound(t *testing.T) {
	dir := t.TempDir()
	ConfigPath = filepath.Join(dir, "config.yaml")

	_, err := RestoreArchivedTodo(DefaultConfig(), 99)
	if err == nil {
		t.Fatal("expected not found error")
	}
}
