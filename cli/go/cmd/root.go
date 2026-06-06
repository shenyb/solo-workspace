package cmd

import (
	"fmt"
	"os"

	"github.com/shenyb/solo-workspace/cli/go/internal"
	"github.com/shenyb/solo-workspace/cli/go/plugins/project"
	"github.com/shenyb/solo-workspace/cli/go/plugins/server"
	"github.com/shenyb/solo-workspace/cli/go/plugins/ssl"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "sw",
	Short: "Solo Workspace — the indie developer's operating system",
	Long: `A unified command center for indie developers.

Manage servers, domains, SSL certificates, deployments,
Docker containers, and more — all from your terminal.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		core.CurrentConfig, err = core.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")

	// Register plugins — each is a Lego brick
	rootCmd.AddCommand(project.Cmd())
	rootCmd.AddCommand(ssl.Cmd())
	rootCmd.AddCommand(server.Cmd())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
