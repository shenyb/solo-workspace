package todo

import (
	"fmt"
	"sort"

	core "github.com/shenyb/solo-workspace/cli/go/internal"
	"github.com/spf13/cobra"
)

// Cmd returns the todo command tree.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "todo",
		Short: "Todo management",
		Long:  `Manage todo items for your projects and tasks.`,
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

	cmd.AddCommand(&cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a todo item",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteTodo(args[0])
		},
	})

	doneCmd := &cobra.Command{
		Use:   "done <name>",
		Short: "Mark a todo item as done",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setTodoDone(args[0], true)
		},
	}
	cmd.AddCommand(doneCmd)

	reopenCmd := &cobra.Command{
		Use:   "reopen <name>",
		Short: "Mark a todo item as not done",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setTodoDone(args[0], false)
		},
	}
	cmd.AddCommand(reopenCmd)

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

	names := make([]string, 0, len(cfg.Todos))
	for name := range cfg.Todos {
		names = append(names, name)
	}
	sort.Strings(names)

	columns := []string{"Name", "Description", "Status"}
	rows := make([][]string, 0, len(names))
	for _, name := range names {
		todo := cfg.Todos[name]
		status := "pending"
		if todo.Done {
			status = "done"
		}
		rows = append(rows, []string{name, todo.Description, status})
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
	cfg.Todos[name] = &core.TodoConfig{Description: desc}
	core.CurrentConfig = cfg
	if err := core.SaveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	fmt.Printf("✅ Todo %q added\n", name)
	return nil
}

func deleteTodo(name string) error {
	cfg := core.CurrentConfig
	if cfg == nil || cfg.Todos == nil {
		return fmt.Errorf("todo %q not found", name)
	}
	if _, exists := cfg.Todos[name]; !exists {
		return fmt.Errorf("todo %q not found", name)
	}
	delete(cfg.Todos, name)
	if err := core.SaveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	fmt.Printf("✅ Todo %q deleted\n", name)
	return nil
}

func setTodoDone(name string, done bool) error {
	cfg := core.CurrentConfig
	if cfg == nil || cfg.Todos == nil {
		return fmt.Errorf("todo %q not found", name)
	}
	todo, exists := cfg.Todos[name]
	if !exists {
		return fmt.Errorf("todo %q not found", name)
	}
	todo.Done = done
	if err := core.SaveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	if done {
		fmt.Printf("✅ Todo %q marked done\n", name)
	} else {
		fmt.Printf("✅ Todo %q reopened\n", name)
	}
	return nil
}
