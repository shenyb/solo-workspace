package secret

import (
	"fmt"
	"os"

	core "github.com/shenyb/solo-workspace/cli/go/internal"
	"github.com/spf13/cobra"
)

var (
	globalVault  *SecretVault
	vaultDataDir string
)

// Cmd returns the secret management command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Manage encrypted secrets (API keys, tokens, passwords)",
		Long: `Save and retrieve sensitive information securely.

Secrets are encrypted using AES-256-GCM and stored locally.
Uses a master password (default: current login user) for encryption.

Examples:
  sw secret set api_key "sk_test_123456"
  sw secret get api_key
  sw secret delete api_key
  sw secret list
`,
		PersistentPreRunE: initSecretVault,
		RunE: func(cmd *cobra.Command, args []string) error {
			keys := globalVault.List()
			if len(keys) == 0 {
				fmt.Println(core.Warn("No secrets saved"))
				return nil
			}
			fmt.Println(core.Info("Saved Secrets:"))
			for _, key := range keys {
				fmt.Printf("  • %s\n", key)
			}
			return nil
		},
	}

	cmd.AddCommand(setCmd())
	cmd.AddCommand(getCmd())
	cmd.AddCommand(deleteCmd())
	cmd.AddCommand(listCmd())

	return cmd
}

// InitVault initializes the global secret vault for read/write access.
func InitVault() error {
	dataDir, err := core.DataDir()
	if err != nil {
		return err
	}
	if globalVault != nil && vaultDataDir == dataDir {
		return nil
	}

	vault, err := NewSecretVault(dataDir)
	if err != nil {
		return err
	}
	vault.SetMasterPassword(getMasterPassword())
	globalVault = vault
	vaultDataDir = dataDir
	return nil
}

// initSecretVault initializes the global secret vault
func initSecretVault(cmd *cobra.Command, args []string) error {
	return InitVault()
}

// getMasterPassword returns the master password for the secret vault.
// Priority: SOLO_SECRET_PASSWORD env var > machine-derived key.
// The machine-derived key is unique per device so the default is never
// the same across different machines, unlike a hardcoded fallback.
func getMasterPassword() string {
 if pwd := os.Getenv("SOLO_SECRET_PASSWORD"); pwd != "" {
  return pwd
 }
 hostname, _ := os.Hostname()
 username := os.Getenv("USER")
 if username == "" {
  username = os.Getenv("USERNAME")
 }
 if hostname == "" {
  hostname = "unknown"
 }
 if username == "" {
  username = "unknown"
 }
 fmt.Fprintf(os.Stderr, "%s No SOLO_SECRET_PASSWORD set, using machine-derived key\n", core.Warn("!"))
 return fmt.Sprintf("sw-%s-%s", hostname, username)
}

// setCmd stores a secret
func setCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <name> <value>",
		Short: "Save an encrypted secret",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			value := args[1]

			if err := globalVault.Set(name, value); err != nil {
				return fmt.Errorf("set secret: %w", err)
			}

			fmt.Printf("%s %s secret '%s' stored\n", core.Success("✓"), core.Info("Secret"), name)
			return nil
		},
	}
}

// getCmd retrieves a secret
func getCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "Retrieve a decrypted secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			value, err := globalVault.Get(name)
			if err != nil {
				return fmt.Errorf("get secret: %w", err)
			}

			fmt.Println(value)
			return nil
		},
	}
}

// deleteCmd removes a secret
func deleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := globalVault.Delete(name); err != nil {
				return fmt.Errorf("delete secret: %w", err)
			}

			fmt.Printf("%s Secret '%s' deleted\n", core.Success("✓"), name)
			return nil
		},
	}
}

// listCmd lists all secret keys
func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all secret keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			keys := globalVault.List()

			if len(keys) == 0 {
				fmt.Println(core.Warn("No secrets saved"))
				return nil
			}

			fmt.Println(core.Info("Saved Secrets:"))
			for _, key := range keys {
				fmt.Printf("  • %s\n", key)
			}

			return nil
		},
	}
}

// GetSecret exposes the secret value for use by other plugins (e.g., SMTP password)
func GetSecret(name string) (string, error) {
	if globalVault == nil {
		return "", fmt.Errorf("secret vault not initialized")
	}
	return globalVault.Get(name)
}

// SetSecret exposes the secret setter for programmatic use
func SetSecret(name, value string) error {
	if globalVault == nil {
		return fmt.Errorf("secret vault not initialized")
	}
	return globalVault.Set(name, value)
}

// DeleteSecret exposes secret deletion for use by other plugins
func DeleteSecret(name string) error {
	if globalVault == nil {
		return fmt.Errorf("secret vault not initialized")
	}
	return globalVault.Delete(name)
}

// ListSecrets exposes secret key listing for use by other plugins
func ListSecrets() ([]string, error) {
	if globalVault == nil {
		return nil, fmt.Errorf("secret vault not initialized")
	}
	return globalVault.List(), nil
}
