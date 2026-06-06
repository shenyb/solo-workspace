package core

import "github.com/spf13/cobra"

// Plugin is the building block of solo-workspace.
// Each plugin registers its own subcommands into the CLI tree.
// This is the "Lego brick" interface.
type Plugin interface {
	// Name returns the plugin identifier (e.g. "ssl", "server")
	Name() string

	// Register adds the plugin's commands to the root command.
	Register(root *cobra.Command)
}
