package server

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/shenyb/solo-workspace/cli/go/internal"
	"github.com/spf13/cobra"
)

// Cmd returns the "server" command tree.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Server management",
		Long:  `List, SSH into, and manage your servers.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all configured servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listServers()
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
		fmt.Println("No servers configured. Edit .solo.yaml or ~/.solo/config.yaml")
		return nil
	}

	columns := []string{"Name", "Host", "User", "Port"}
	var rows [][]string
	for name, srv := range cfg.Servers {
		port := 22
		if srv.Port > 0 {
			port = srv.Port
		}
		rows = append(rows, []string{name, srv.Host, srv.User, fmt.Sprintf("%d", port)})
	}
	core.Table(columns, rows)
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
