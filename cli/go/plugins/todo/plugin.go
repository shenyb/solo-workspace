package todo

import (
	"fmt"

	core "github.com/shenyb/solo-workspace/cli/go/internal"
	"github.com/spf13/cobra"
)

// Cmd returns the todo command tree.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "todo",
		Short: "Todo management",
		Long:  `Manage todo items for your projects and tasks.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listTodos()
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all todo items",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listTodos()
		},
	})

	addCmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a todo item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			desc, _ := cmd.Flags().GetString("desc")
			return addTodo(args[0], desc)
		},
	}
	addCmd.Flags().String("desc", "", "Todo description")
	cmd.AddCommand(addCmd)

	updateCmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a todo item by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := core.ParseID(args[0])
			if err != nil {
				return err
			}
			name, _ := cmd.Flags().GetString("name")
			desc, _ := cmd.Flags().GetString("desc")
			return updateTodo(id, name, desc)
		},
	}
	updateCmd.Flags().String("name", "", "New todo name")
	updateCmd.Flags().String("desc", "", "New todo description")
	cmd.AddCommand(updateCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a todo item by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := core.ParseID(args[0])
			if err != nil {
				return err
			}
			return deleteTodo(id)
		},
	})

	doneCmd := &cobra.Command{
		Use:   "done <id>",
		Short: "Mark a todo item as done by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := core.ParseID(args[0])
			if err != nil {
				return err
			}
			return setTodoDone(id, true)
		},
	}
	cmd.AddCommand(doneCmd)

	reopenCmd := &cobra.Command{
		Use:   "reopen <id>",
		Short: "Mark a todo item as not done by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := core.ParseID(args[0])
			if err != nil {
				return err
			}
			return setTodoDone(id, false)
		},
	}
	cmd.AddCommand(reopenCmd)

	archiveCmd := &cobra.Command{
		Use:   "archive",
		Short: "Archive stale todos and list archived items",
		Long:  `Archive todos inactive for more than 2 weeks to todos-archive.yaml in the config directory.`,
	}
	archiveCmd.AddCommand(&cobra.Command{
		Use:   "run",
		Short: "Manually archive todos older than 2 weeks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runArchive()
		},
	})
	archiveCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List archived todos",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listArchivedTodos()
		},
	})
	cmd.AddCommand(archiveCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "note <id> <text>",
		Short: "Add a note to a todo item",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := core.ParseID(args[0])
			if err != nil {
				return err
			}
			return noteTodo(id, args[1])
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "stats",
		Short: "Show todo statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			return showStats()
		},
	})

	return cmd
}

func listTodos() error {
	cfg := core.CurrentConfig
	if cfg == nil {
		cfg = core.DefaultConfig()
	}
	if len(cfg.Todos) == 0 {
		fmt.Println("No todo items configured.")
		return nil
	}

	columns := []string{"ID", "Name", "Description", "Note", "Status", "Created", "Updated"}
	rows := make([][]string, 0, len(cfg.Todos))
	for _, entry := range core.SortedTodos(cfg) {
		note := ""
		if entry.Config.Note != "" {
			note = "📝"
		}
		rows = append(rows, []string{
			fmt.Sprintf("%d", entry.Config.ID),
			entry.Name,
			entry.Config.Description,
			note,
			core.TodoStatus(entry.Config.Done),
			core.FormatTodoTime(entry.Config.CreatedAt),
			core.FormatTodoTime(entry.Config.UpdatedAt),
		})
	}
	core.Table(columns, rows)
	return nil
}

func addTodo(name, desc string) error {
	cfg := core.CurrentConfig
	if cfg == nil {
		cfg = core.DefaultConfig()
	}
	if cfg.Todos == nil {
		cfg.Todos = make(map[string]*core.TodoConfig)
	}
	if _, exists := cfg.Todos[name]; exists {
		return fmt.Errorf("todo %q already exists", name)
	}
	created, updated := core.NewTodoTimestamps()
	cfg.Todos[name] = &core.TodoConfig{
		ID:          core.NextTodoID(cfg),
		Description: desc,
		CreatedAt:   created,
		UpdatedAt:   updated,
	}
	core.CurrentConfig = cfg
	if err := core.SaveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	fmt.Printf("✅ Todo %q added (id=%d)\n", name, cfg.Todos[name].ID)
	return nil
}

func updateTodo(id int, newName, desc string) error {
	cfg := core.CurrentConfig
	if cfg == nil || cfg.Todos == nil {
		return fmt.Errorf("todo id %d not found", id)
	}

	oldName, todo, err := core.TodoByID(cfg, id)
	if err != nil {
		return err
	}

	if newName == "" && desc == "" {
		return fmt.Errorf("at least one of --name or --desc is required")
	}

	if newName != "" && newName != oldName {
		if _, exists := cfg.Todos[newName]; exists {
			return fmt.Errorf("todo %q already exists", newName)
		}
		delete(cfg.Todos, oldName)
		cfg.Todos[newName] = todo
		oldName = newName
	}

	if desc != "" {
		todo.Description = desc
	}

	core.TouchTodoUpdated(todo)

	if err := core.SaveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	fmt.Printf("✅ Todo %q (id=%d) updated\n", oldName, id)
	return nil
}

func deleteTodo(id int) error {
	cfg := core.CurrentConfig
	if cfg == nil || cfg.Todos == nil {
		return fmt.Errorf("todo id %d not found", id)
	}

	name, _, err := core.TodoByID(cfg, id)
	if err != nil {
		return err
	}

	delete(cfg.Todos, name)
	if err := core.SaveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	fmt.Printf("✅ Todo %q (id=%d) deleted\n", name, id)
	return nil
}

func runArchive() error {
	cfg := core.CurrentConfig
	if cfg == nil {
		cfg = core.DefaultConfig()
	}
	archived, err := core.ArchiveStaleTodos(cfg)
	if err != nil {
		return err
	}
	if len(archived) == 0 {
		fmt.Println("No stale todos to archive.")
		return nil
	}
	core.CurrentConfig = cfg
	if err := core.SaveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	path, err := core.TodoArchivePath()
	if err != nil {
		return err
	}
	fmt.Printf("✅ Archived %d todo(s) to %s\n", len(archived), path)
	for _, name := range archived {
		fmt.Printf("  - %s\n", name)
	}
	return nil
}

func listArchivedTodos() error {
	archive, err := core.LoadTodoArchive()
	if err != nil {
		return err
	}
	if len(archive.Todos) == 0 {
		fmt.Println("No archived todos.")
		return nil
	}

	columns := []string{"ID", "Name", "Description", "Status", "Created", "Updated", "Archived"}
	rows := make([][]string, 0, len(archive.Todos))
	for _, entry := range core.SortedArchivedTodos(archive) {
		rows = append(rows, []string{
			fmt.Sprintf("%d", entry.Config.ID),
			entry.Name,
			entry.Config.Description,
			core.TodoStatus(entry.Config.Done),
			core.FormatTodoTime(entry.Config.CreatedAt),
			core.FormatTodoTime(entry.Config.UpdatedAt),
			core.FormatTodoTime(entry.Config.ArchivedAt),
		})
	}
	core.Table(columns, rows)
	return nil
}

func setTodoDone(id int, done bool) error {
	cfg := core.CurrentConfig
	if cfg == nil || cfg.Todos == nil {
		return fmt.Errorf("todo id %d not found", id)
	}

	name, todo, err := core.TodoByID(cfg, id)
	if err != nil {
		return err
	}

	todo.Done = done
	core.TouchTodoUpdated(todo)
	if err := core.SaveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	if done {
		fmt.Printf("✅ Todo %q (id=%d) marked done\n", name, id)
	} else {
		fmt.Printf("✅ Todo %q (id=%d) reopened\n", name, id)
	}
	return nil
}

func noteTodo(id int, note string) error {
	cfg := core.CurrentConfig
	if cfg == nil || cfg.Todos == nil {
		return fmt.Errorf("todo id %d not found", id)
	}

	name, todo, err := core.TodoByID(cfg, id)
	if err != nil {
		return err
	}

	if todo.Note != "" {
		todo.Note += "\n"
	}
	todo.Note += note
	core.TouchTodoUpdated(todo)
	if err := core.SaveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	fmt.Printf("✅ Note added to todo %q (id=%d)\n", name, id)
	return nil
}

func showStats() error {
	cfg := core.CurrentConfig
	if cfg == nil {
		cfg = core.DefaultConfig()
	}

	total := len(cfg.Todos)
	done := 0
	withNote := 0
	for _, t := range cfg.Todos {
		if t.Done {
			done++
		}
		if t.Note != "" {
			withNote++
		}
	}

	// Count archived
	archived := 0
	archive, err := core.LoadTodoArchive()
	if err == nil {
		archived = len(archive.Todos)
	}

	fmt.Printf("📊 Todo Statistics\n")
	fmt.Printf("  Total:     %d\n", total)
	fmt.Printf("  Pending:   %d\n", total-done)
	fmt.Printf("  Completed: %d\n", done)
	fmt.Printf("  With note: %d\n", withNote)
	fmt.Printf("  Archived:  %d\n", archived)
	return nil
}
