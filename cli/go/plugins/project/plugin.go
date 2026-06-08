package project

import (
	"fmt"

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
		Use:   "delete <id>",
		Short: "Delete a project by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := core.ParseID(args[0])
			if err != nil {
				return err
			}
			return deleteProject(id)
		},
	})

	updateCmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a project's path or description by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := core.ParseID(args[0])
			if err != nil {
				return err
			}
			path, _ := cmd.Flags().GetString("path")
			desc, _ := cmd.Flags().GetString("desc")
			return updateProject(id, path, desc)
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

	columns := []string{"ID", "Name", "Path", "Description"}
	rows := make([][]string, 0, len(cfg.Projects))
	for _, entry := range core.SortedProjects(cfg) {
		rows = append(rows, []string{
			fmt.Sprintf("%d", entry.Config.ID),
			entry.Name,
			entry.Config.Path,
			entry.Config.Description,
		})
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
		ID:          core.NextProjectID(cfg),
		Path:        path,
		Description: desc,
	}
	core.CurrentConfig = cfg

	if err := core.SaveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("✅ Project %q added (id=%d, %s)\n", name, cfg.Projects[name].ID, path)
	return nil
}

func deleteProject(id int) error {
	cfg := core.CurrentConfig
	if cfg == nil || cfg.Projects == nil {
		return fmt.Errorf("project id %d not found", id)
	}

	name, _, err := core.ProjectByID(cfg, id)
	if err != nil {
		return err
	}

	delete(cfg.Projects, name)

	if err := core.SaveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("✅ Project %q (id=%d) deleted\n", name, id)
	return nil
}

func updateProject(id int, path, desc string) error {
	cfg := core.CurrentConfig
	if cfg == nil || cfg.Projects == nil {
		return fmt.Errorf("project id %d not found", id)
	}

	name, p, err := core.ProjectByID(cfg, id)
	if err != nil {
		return err
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

	fmt.Printf("✅ Project %q (id=%d) updated\n", name, id)
	return nil
}
