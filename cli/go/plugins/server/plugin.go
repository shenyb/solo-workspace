package server

import (
	"fmt"
	"os"
	"os/exec"
	"sort"

	core "github.com/shenyb/solo-workspace/cli/go/internal"
	"github.com/spf13/cobra"
)

// Cmd returns the "server" command tree.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Server management",
		Long:  `List, add, update, delete, and SSH into your servers.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all configured servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listServers()
		},
	})

	addCmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			host, _ := cmd.Flags().GetString("host")
			user, _ := cmd.Flags().GetString("user")
			port, _ := cmd.Flags().GetInt("port")
			return addServer(args[0], host, user, port)
		},
	}
	addCmd.Flags().String("host", "", "Server host or IP address")
	addCmd.Flags().String("user", "", "SSH username")
	addCmd.Flags().Int("port", 22, "SSH port")
	_ = addCmd.MarkFlagRequired("host")
	_ = addCmd.MarkFlagRequired("user")
	cmd.AddCommand(addCmd)

	updateCmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update a server by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			host, _ := cmd.Flags().GetString("host")
			user, _ := cmd.Flags().GetString("user")
			port, _ := cmd.Flags().GetInt("port")
			portSet := cmd.Flags().Changed("port")
			return updateServer(args[0], host, user, port, portSet)
		},
	}
	updateCmd.Flags().String("host", "", "New host or IP address")
	updateCmd.Flags().String("user", "", "New SSH username")
	updateCmd.Flags().Int("port", 22, "New SSH port")
	cmd.AddCommand(updateCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a server by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteServer(args[0])
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "ssh <name>",
		Short: "SSH into a server by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return sshServer(args[0])
		},
	})

	return cmd
}

func listServers() error {
	cfg := core.CurrentConfig
	if cfg == nil {
		cfg = core.DefaultConfig()
	}
	if len(cfg.Servers) == 0 {
		fmt.Println("No servers configured. Use: sw server add <name> --host <ip> --user <user>")
		return nil
	}

	names := make([]string, 0, len(cfg.Servers))
	for name := range cfg.Servers {
		names = append(names, name)
	}
	sort.Strings(names)

	columns := []string{"Name", "Host", "User", "Port"}
	rows := make([][]string, 0, len(names))
	for _, name := range names {
		srv := cfg.Servers[name]
		port := 22
		if srv.Port > 0 {
			port = srv.Port
		}
		rows = append(rows, []string{name, srv.Host, srv.User, fmt.Sprintf("%d", port)})
	}
	core.Table(columns, rows)
	return nil
}

func addServer(name, host, user string, port int) error {
	cfg := core.CurrentConfig
	if cfg == nil {
		cfg = core.DefaultConfig()
	}
	if cfg.Servers == nil {
		cfg.Servers = make(map[string]*core.ServerConfig)
	}
	if _, exists := cfg.Servers[name]; exists {
		return fmt.Errorf("server %q already exists", name)
	}

	cfg.Servers[name] = &core.ServerConfig{
		Host: host,
		User: user,
		Port: port,
	}
	core.CurrentConfig = cfg

	if err := core.SaveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("✅ Server %q added (%s@%s:%d)\n", name, user, host, port)
	return nil
}

func updateServer(name, host, user string, port int, portSet bool) error {
	cfg := core.CurrentConfig
	if cfg == nil || cfg.Servers == nil {
		return fmt.Errorf("server %q not found", name)
	}

	srv, ok := cfg.Servers[name]
	if !ok {
		return fmt.Errorf("server %q not found", name)
	}

	if host == "" && user == "" && !portSet {
		return fmt.Errorf("at least one of --host, --user, or --port is required")
	}

	if host != "" {
		srv.Host = host
	}
	if user != "" {
		srv.User = user
	}
	if portSet {
		srv.Port = port
	}

	if err := core.SaveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("✅ Server %q updated\n", name)
	return nil
}

func deleteServer(name string) error {
	cfg := core.CurrentConfig
	if cfg == nil || cfg.Servers == nil {
		return fmt.Errorf("server %q not found", name)
	}
	if _, ok := cfg.Servers[name]; !ok {
		return fmt.Errorf("server %q not found", name)
	}

	delete(cfg.Servers, name)

	if err := core.SaveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("✅ Server %q deleted\n", name)
	return nil
}

func sshServer(name string) error {
	cfg := core.CurrentConfig
	if cfg == nil {
		return fmt.Errorf("no config loaded")
	}
	srv, ok := cfg.Servers[name]
	if !ok {
		return fmt.Errorf("server %q not found in config", name)
	}

	port := srv.Port
	if port == 0 {
		port = 22
	}

	sshCmd := exec.Command("ssh", "-p", fmt.Sprintf("%d", port), fmt.Sprintf("%s@%s", srv.User, srv.Host))
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr

	return sshCmd.Run()
}
