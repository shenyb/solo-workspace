package project

import (
	"fmt"
	"sort"

	"github.com/shenyb/solo-workspace/cli/go/internal"
	"github.com/spf13/cobra"
)

// Cmd returns the "project" command tree.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Project management",
		Long:  `List, add, update, and delete your local projects.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listProjects()
		},
	})

	addCmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, _ := cmd.Flags().GetString("path")
			desc, _ := cmd.Flags().GetString("desc")
			return addProject(args[0], path, desc)
		},
	}
	addCmd.Flags().String("path", "", "Project path (default: auto-guess)")
	addCmd.Flags().String("desc", "", "Project description")
	_ = addCmd.MarkFlagRequired("path")
	cmd.AddCommand(addCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteProject(args[0])
		},
	})

	updateCmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a project's path or description",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, _ := cmd.Flags().GetString("path")
			desc, _ := cmd.Flags().GetString("desc")
			return updateProject(args[0], path, desc)
		},
	}
	updateCmd.Flags().String("path", "", "New project path")
	updateCmd.Flags().String("desc", "", "New project description")
	cmd.AddCommand(updateCmd)

	return cmd
}

func listProjects() error {
	cfg := core.CurrentConfig
	if cfg == nil {
		cfg = core.DefaultConfig()
	}
	if len(cfg.Projects) == 0 {
		fmt.Println("No projects configured.")
		return nil
	}

	columns := []string{"Name", "Path", "Description"}
	var rows [][]string

	names := make([]string, 0, len(cfg.Projects))
	for name := range cfg.Projects {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		p := cfg.Projects[name]
		rows = append(rows, []string{name, p.Path, p.Description})
	}
	core.Table(columns, rows)
	return nil
}

func addProject(name, path, desc string) error {
	cfg := core.CurrentConfig
	if cfg == nil {
		cfg = core.DefaultConfig()
	}
	if cfg.Projects == nil {
		cfg.Projects = make(map[string]*core.ProjectConfig)
	}

	if _, exists := cfg.Projects[name]; exists {
		return fmt.Errorf("project %q already exists", name)
	}

	cfg.Projects[name] = &core.ProjectConfig{
		Path:        path,
		Description: desc,
	}
	core.CurrentConfig = cfg

	if err := core.SaveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("✅ Project %q added (%s)\n", name, path)
	return nil
}

func deleteProject(name string) error {
	cfg := core.CurrentConfig
	if cfg == nil || cfg.Projects == nil {
		return fmt.Errorf("project %q not found", name)
	}

	if _, exists := cfg.Projects[name]; !exists {
		return fmt.Errorf("project %q not found", name)
	}

	delete(cfg.Projects, name)

	if err := core.SaveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("✅ Project %q deleted\n", name)
	return nil
}

func updateProject(name, path, desc string) error {
	cfg := core.CurrentConfig
	if cfg == nil || cfg.Projects == nil {
		return fmt.Errorf("project %q not found", name)
	}

	p, exists := cfg.Projects[name]
	if !exists {
		return fmt.Errorf("project %q not found", name)
	}

	if path == "" && desc == "" {
		return fmt.Errorf("at least one of --path or --desc is required")
	}

	if path != "" {
		p.Path = path
	}
	if desc != "" {
		p.Description = desc
	}

	if err := core.SaveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("✅ Project %q updated\n", name)
	return nil
}
