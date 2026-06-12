package core

import (
	"fmt"
	"sort"
	"strconv"
	"time"
)

// EnsureIDs assigns auto-increment IDs and resolves duplicate IDs.
func EnsureIDs(cfg *Config) {
	if cfg == nil {
		return
	}
	if cfg.Projects != nil {
		ensureProjectIDs(cfg)
	}
	if cfg.Todos != nil {
		ensureTodoIDs(cfg)
	}
}

func ensureProjectIDs(cfg *Config) {
	used := make(map[int]string)
	names := sortedKeys(cfg.Projects)
	next := maxProjectID(cfg) + 1

	for _, name := range names {
		if cfg.Projects[name] == nil {
			cfg.Projects[name] = &ProjectConfig{}
		}
		if cfg.Projects[name].ID == 0 {
			for used[next] != "" {
				next++
			}
			cfg.Projects[name].ID = next
			used[next] = name
			next++
		}
	}
	for _, name := range names {
		id := cfg.Projects[name].ID
		if owner, ok := used[id]; ok && owner != name {
			for used[next] != "" {
				next++
			}
			cfg.Projects[name].ID = next
			used[next] = name
			next++
		} else {
			used[id] = name
		}
	}
}

func ensureTodoIDs(cfg *Config) {
	used := make(map[int]string)
	names := sortedKeysTodos(cfg)
	next := maxTodoID(cfg) + 1
	for _, name := range names {
		if cfg.Todos[name] == nil {
			cfg.Todos[name] = &TodoConfig{}
		}
		if cfg.Todos[name].ID == 0 {
			for used[next] != "" {
				next++
			}
			cfg.Todos[name].ID = next
			used[next] = name
			next++
		}
	}
	for _, name := range names {
		id := cfg.Todos[name].ID
		if owner, ok := used[id]; ok && owner != name {
			for used[next] != "" {
				next++
			}
			cfg.Todos[name].ID = next
			used[next] = name
			next++
		} else {
			used[id] = name
		}
	}
}

func sortedKeys(m map[string]*ProjectConfig) []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedKeysTodos(cfg *Config) []string {
	names := make([]string, 0, len(cfg.Todos))
	for name := range cfg.Todos {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func maxProjectID(cfg *Config) int {
	max := 0
	for _, p := range cfg.Projects {
		if p == nil {
			continue
		}
		if p.ID > max {
			max = p.ID
		}
	}
	return max
}

func maxTodoID(cfg *Config) int {
	max := 0
	for _, t := range cfg.Todos {
		if t == nil {
			continue
		}
		if t.ID > max {
			max = t.ID
		}
	}
	return max
}

// NextProjectID returns the next available project ID.
func NextProjectID(cfg *Config) int {
	return maxProjectID(cfg) + 1
}

// NextTodoID returns the next available todo ID.
func NextTodoID(cfg *Config) int {
	return maxTodoID(cfg) + 1
}

// ParseID parses a numeric ID argument.
func ParseID(s string) (int, error) {
	id, err := strconv.Atoi(s)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id %q, must be a positive integer", s)
	}
	return id, nil
}

// ProjectByID finds a project by its ID.
func ProjectByID(cfg *Config, id int) (string, *ProjectConfig, error) {
	if cfg == nil || cfg.Projects == nil {
		return "", nil, fmt.Errorf("project id %d not found", id)
	}
	for name, p := range cfg.Projects {
		if p != nil && p.ID == id {
			return name, p, nil
		}
	}
	return "", nil, fmt.Errorf("project id %d not found", id)
}

// TodoByID finds a todo by its ID.
func TodoByID(cfg *Config, id int) (string, *TodoConfig, error) {
	if cfg == nil || cfg.Todos == nil {
		return "", nil, fmt.Errorf("todo id %d not found", id)
	}
	for name, t := range cfg.Todos {
		if t != nil && t.ID == id {
			return name, t, nil
		}
	}
	return "", nil, fmt.Errorf("todo id %d not found", id)
}

// ProjectEntry pairs a project name with its config.
type ProjectEntry struct {
	Name   string
	Config *ProjectConfig
}

// SortedProjects returns projects sorted by ID.
func SortedProjects(cfg *Config) []ProjectEntry {
	if cfg == nil || len(cfg.Projects) == 0 {
		return nil
	}
	entries := make([]ProjectEntry, 0, len(cfg.Projects))
	for name, p := range cfg.Projects {
		if p == nil {
			continue
		}
		entries = append(entries, ProjectEntry{Name: name, Config: p})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Config.ID < entries[j].Config.ID
	})
	return entries
}

// TodoEntry pairs a todo name with its config.
type TodoEntry struct {
	Name   string
	Config *TodoConfig
}

// SortedTodos returns todos sorted by status (pending first), then updated_at descending.
func SortedTodos(cfg *Config) []TodoEntry {
	if cfg == nil || len(cfg.Todos) == 0 {
		return nil
	}
	entries := make([]TodoEntry, 0, len(cfg.Todos))
	for name, t := range cfg.Todos {
		if t == nil {
			continue
		}
		entries = append(entries, TodoEntry{Name: name, Config: t})
	}
	sort.Slice(entries, func(i, j int) bool {
		ai, aj := entries[i].Config, entries[j].Config
		if ai.Done != aj.Done {
			return !ai.Done
		}
		return ai.UpdatedAt.After(aj.UpdatedAt)
	})
	return entries
}

// NewTodoTimestamps returns created_at and updated_at for a new todo.
func NewTodoTimestamps() (time.Time, time.Time) {
	now := time.Now()
	return now, now
}

// TouchTodoUpdated sets updated_at to now, backfilling created_at if missing.
func TouchTodoUpdated(t *TodoConfig) {
	now := time.Now()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	t.UpdatedAt = now
}

// FormatTodoTime formats a todo timestamp for display.
func FormatTodoTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Local().Format("2006-01-02 15:04")
}
