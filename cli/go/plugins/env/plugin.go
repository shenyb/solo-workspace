package env

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	core "github.com/shenyb/solo-workspace/cli/go/internal"
	"github.com/shenyb/solo-workspace/cli/go/plugins/secret"
	"github.com/spf13/cobra"
)

// envDB manages environment variables persistently
// Stored in ~/.solo/env.yaml: name -> value pairs
// Sensitive vars (marked with secret_*) are encrypted via secret plugin
type envDB struct {
	vars map[string]string // name -> value
	path string
}

var db *envDB

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	soloDir := filepath.Join(home, ".solo")
	os.MkdirAll(soloDir, 0700)
	db = &envDB{
		vars: make(map[string]string),
		path: filepath.Join(soloDir, "env.yaml"),
	}
	db.load()
}

// Cmd returns the env management command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Manage environment variables",
		Long: `Save and manage environment variables locally.

Variables prefixed with 'secret_' are encrypted using the secret plugin.
Other variables are saved in plaintext in ~/.solo/env.yaml

Examples:
  sw env set DB_HOST localhost              # Save plaintext var
  sw env set secret_db_password "p@ssw0rd"  # Save encrypted var
  sw env get DB_HOST                        # Retrieve variable
  sw env list                                # Show all variables
  sw env export > .env                      # Export as .env format
  sw env export --prefix secret_            # Export only secret vars
  sw env delete DB_HOST                     # Remove variable
  sw env unset-secret DB_PASSWORD          # Remove from secret vault
`,
	}

	cmd.AddCommand(setCmd())
	cmd.AddCommand(getCmd())
	cmd.AddCommand(deleteCmd())
	cmd.AddCommand(listCmd())
	cmd.AddCommand(exportCmd())
	cmd.AddCommand(unsetSecretCmd())

	return cmd
}

// setCmd saves an environment variable
func setCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <name> <value>",
		Short: "Set environment variable",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			value := args[1]

			if strings.HasPrefix(name, "secret_") {
				// Save in secret vault
				if err := secret.SetSecret(name, value); err != nil {
					return fmt.Errorf("save secret: %w", err)
				}
			} else {
				// Save in plaintext
				db.vars[name] = value
				if err := db.save(); err != nil {
					return fmt.Errorf("save env: %w", err)
				}
			}

			fmt.Printf("%s Environment variable %s set\n", core.Success("✓"), core.Info(name))
			return nil
		},
	}
}

// getCmd retrieves an environment variable
func getCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Get environment variable value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			var value string
			if strings.HasPrefix(name, "secret_") {
				// Retrieve from secret vault
				val, err := secret.GetSecret(name)
				if err != nil {
					return fmt.Errorf("get secret: %w", err)
				}
				value = val
			} else {
				// Retrieve from plaintext
				var ok bool
				value, ok = db.vars[name]
				if !ok {
					return fmt.Errorf("variable %s not found", name)
				}
			}

			fmt.Println(value)
			return nil
		},
	}
}

// deleteCmd removes an environment variable
func deleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete environment variable",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if strings.HasPrefix(name, "secret_") {
				// Delete from secret vault
				if err := secret.DeleteSecret(name); err != nil {
					return fmt.Errorf("delete secret: %w", err)
				}
			} else {
				// Delete from plaintext
				if _, ok := db.vars[name]; !ok {
					return fmt.Errorf("variable %s not found", name)
				}
				delete(db.vars, name)
				if err := db.save(); err != nil {
					return fmt.Errorf("save env: %w", err)
				}
			}

			fmt.Printf("%s Environment variable %s deleted\n", core.Success("✓"), name)
			return nil
		},
	}
}

// listCmd shows all saved environment variables
func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all environment variables",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix, _ := cmd.Flags().GetString("prefix")

			if len(db.vars) == 0 {
				fmt.Println("No environment variables saved")
			} else {
				fmt.Println("Saved Environment Variables:")
				for name := range db.vars {
					if prefix == "" || strings.HasPrefix(name, prefix) {
						fmt.Printf("  • %s\n", name)
					}
				}
			}

			// Show secret variables
			secretList, err := secret.ListSecrets()
			if err == nil {
				for _, name := range secretList {
					if prefix == "" || strings.HasPrefix(name, prefix) {
						fmt.Printf("  • %s (secret)\n", name)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringP("prefix", "p", "", "Filter by variable name prefix")
	return cmd
}

// exportCmd exports environment variables in .env format
func exportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [file]",
		Short: "Export environment variables in .env format",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix, _ := cmd.Flags().GetString("prefix")
			masked, _ := cmd.Flags().GetBool("masked")

			var lines []string

			// Add plaintext variables
			for name, value := range db.vars {
				if prefix != "" && !strings.HasPrefix(name, prefix) {
					continue
				}
				if masked {
					lines = append(lines, fmt.Sprintf("%s=***masked***", name))
				} else {
					lines = append(lines, fmt.Sprintf("%s=%s", name, value))
				}
			}

			// Add secret variables
			secretList, err := secret.ListSecrets()
			if err == nil {
				for _, name := range secretList {
					if prefix != "" && !strings.HasPrefix(name, prefix) {
						continue
					}
					if masked || strings.HasPrefix(name, "secret_") {
						lines = append(lines, fmt.Sprintf("%s=***masked***", name))
					} else {
						val, _ := secret.GetSecret(name)
						lines = append(lines, fmt.Sprintf("%s=%s", name, val))
					}
				}
			}

			output := strings.Join(lines, "\n")
			if len(output) > 0 {
				output += "\n"
			}

			if len(args) > 0 {
				// Write to file
				if err := os.WriteFile(args[0], []byte(output), 0600); err != nil {
					return fmt.Errorf("write file: %w", err)
				}
				fmt.Printf("%s Environment variables exported to %s\n", core.Success("✓"), args[0])
			} else {
				// Write to stdout
				fmt.Print(output)
			}

			return nil
		},
	}

	cmd.Flags().StringP("prefix", "p", "", "Export only variables with this prefix")
	cmd.Flags().BoolP("masked", "m", false, "Mask sensitive values with ***masked***")
	return cmd
}

// unsetSecretCmd removes a variable from the secret vault
func unsetSecretCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unset-secret <name>",
		Short: "Remove variable from secret vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := secret.DeleteSecret(name); err != nil {
				return fmt.Errorf("delete secret: %w", err)
			}

			fmt.Printf("%s Secret variable %s unset\n", core.Success("✓"), name)
			return nil
		},
	}
}

// ── Helpers ────────────────────────────────────────────

// ListEnvVarNames returns all plaintext env var names (for use by other plugins).
func ListEnvVarNames() []string {
	if db == nil {
		return nil
	}
	names := make([]string, 0, len(db.vars))
	for name := range db.vars {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// load reads environment variables from ~/.solo/env.yaml
func (e *envDB) load() error {
	data, err := os.ReadFile(e.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read env file: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
				value = value[1 : len(value)-1]
			}
			e.vars[name] = value
		}
	}

	return scanner.Err()
}

// save writes environment variables to ~/.solo/env.yaml
func (e *envDB) save() error {
	var lines []string
	lines = append(lines, "# Environment variables managed by solo-workspace")

	for name, value := range e.vars {
		// Quote values that contain spaces or special chars
		quotedValue := value
		if strings.ContainsAny(value, " \t\n#=") {
			quotedValue = fmt.Sprintf("%q", value)
		}
		lines = append(lines, fmt.Sprintf("%s=%s", name, quotedValue))
	}

	data := strings.Join(lines, "\n")
	if len(data) > 0 {
		data += "\n"
	}

	if err := os.WriteFile(e.path, []byte(data), 0600); err != nil {
		return fmt.Errorf("write env file: %w", err)
	}

	return nil
}
