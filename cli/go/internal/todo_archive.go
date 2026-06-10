package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// TodoArchiveMaxAge is how long a todo may remain active before auto-archiving.
	TodoArchiveMaxAge = 14 * 24 * time.Hour
	todoArchiveFile   = "todos-archive.yaml"
)

// ArchivedTodo extends TodoConfig with archive metadata.
type ArchivedTodo struct {
	ID          int       `yaml:"id,omitempty"`
	Description string    `yaml:"description,omitempty"`
	Done        bool      `yaml:"done,omitempty"`
	CreatedAt   time.Time `yaml:"created_at,omitempty"`
	UpdatedAt   time.Time `yaml:"updated_at,omitempty"`
	ArchivedAt  time.Time `yaml:"archived_at,omitempty"`
}

// TodoArchive holds archived todo items for a config directory.
type TodoArchive struct {
	Todos map[string]*ArchivedTodo `yaml:"todos"`
}

// TodoArchivePath returns the archive file path alongside the active config.
func TodoArchivePath() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, todoArchiveFile), nil
}

// TodoLastActivity returns the most recent timestamp for a todo.
func TodoLastActivity(t *TodoConfig) time.Time {
	if t == nil {
		return time.Time{}
	}
	if !t.UpdatedAt.IsZero() {
		return t.UpdatedAt
	}
	return t.CreatedAt
}

// IsTodoStale reports whether a todo's last activity exceeds maxAge.
func IsTodoStale(t *TodoConfig, maxAge time.Duration) bool {
	last := TodoLastActivity(t)
	if last.IsZero() {
		return false
	}
	return time.Since(last) > maxAge
}

// LoadTodoArchive reads archived todos from disk.
func LoadTodoArchive() (*TodoArchive, error) {
	path, err := TodoArchivePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &TodoArchive{Todos: make(map[string]*ArchivedTodo)}, nil
		}
		return nil, fmt.Errorf("read todo archive %s: %w", path, err)
	}

	archive := &TodoArchive{Todos: make(map[string]*ArchivedTodo)}
	if err := yaml.Unmarshal(data, archive); err != nil {
		return nil, fmt.Errorf("parse todo archive %s: %w", path, err)
	}
	if archive.Todos == nil {
		archive.Todos = make(map[string]*ArchivedTodo)
	}
	return archive, nil
}

// SaveTodoArchive writes archived todos to disk.
func SaveTodoArchive(archive *TodoArchive) error {
	if archive == nil {
		archive = &TodoArchive{Todos: make(map[string]*ArchivedTodo)}
	}
	if archive.Todos == nil {
		archive.Todos = make(map[string]*ArchivedTodo)
	}

	path, err := TodoArchivePath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(archive)
	if err != nil {
		return fmt.Errorf("marshal todo archive: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write todo archive %s: %w", path, err)
	}
	return nil
}

// ArchiveStaleTodos moves todos older than TodoArchiveMaxAge into the archive file.
// Returns the names of archived todos.
func ArchiveStaleTodos(cfg *Config) ([]string, error) {
	if cfg == nil || len(cfg.Todos) == 0 {
		return nil, nil
	}

	var stale []string
	for name, todo := range cfg.Todos {
		if IsTodoStale(todo, TodoArchiveMaxAge) {
			stale = append(stale, name)
		}
	}
	if len(stale) == 0 {
		return nil, nil
	}

	archive, err := LoadTodoArchive()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	archived := make([]string, 0, len(stale))
	for _, name := range stale {
		todo := cfg.Todos[name]
		archive.Todos[name] = &ArchivedTodo{
			ID:          todo.ID,
			Description: todo.Description,
			Done:        todo.Done,
			CreatedAt:   todo.CreatedAt,
			UpdatedAt:   todo.UpdatedAt,
			ArchivedAt:  now,
		}
		delete(cfg.Todos, name)
		archived = append(archived, name)
	}

	if err := SaveTodoArchive(archive); err != nil {
		return nil, err
	}
	return archived, nil
}

// ArchivedTodoEntry pairs a todo name with its archived config.
type ArchivedTodoEntry struct {
	Name   string
	Config *ArchivedTodo
}

// SortedArchivedTodos returns archived todos sorted by archived_at descending.
func SortedArchivedTodos(archive *TodoArchive) []ArchivedTodoEntry {
	if archive == nil || len(archive.Todos) == 0 {
		return nil
	}

	entries := make([]ArchivedTodoEntry, 0, len(archive.Todos))
	for name, t := range archive.Todos {
		entries = append(entries, ArchivedTodoEntry{Name: name, Config: t})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Config.ArchivedAt.After(entries[j].Config.ArchivedAt)
	})
	return entries
}
