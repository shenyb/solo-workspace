package secret

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	core "github.com/shenyb/solo-workspace/cli/go/internal"
	"github.com/spf13/cobra"
)

func init() {
	core.SecretResolver = func(name string) (string, error) {
		if err := InitVault(); err != nil {
			return "", err
		}
		return GetSecret(name)
	}
}

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
	if err := vault.SetMasterPassword(getMasterPassword()); err != nil {
		return err
	}
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
//
// The machine-derived key uses a stable hardware identifier (IOPlatformUUID
// on macOS, /etc/machine-id on Linux, MachineGuid in the Windows registry)
// rather than the hostname, which on macOS changes with network environment
// and breaks decryption of existing vaults. The username is included so
// different accounts on the same machine still get different keys.
func getMasterPassword() string {
	if pwd := os.Getenv("SOLO_SECRET_PASSWORD"); pwd != "" {
		return pwd
	}
	machineID := stableMachineID()
	if machineID == "" {
		machineID = "unknown"
	}
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}
	if username == "" {
		username = "unknown"
	}
	fmt.Fprintf(os.Stderr, "%s No SOLO_SECRET_PASSWORD set, using machine-derived key\n", core.Warn("!"))
	return fmt.Sprintf("sw-%s-%s", machineID, username)
}

// stableMachineID returns a stable, hardware-backed machine identifier.
// Falls back to "" if no identifier can be obtained (caller maps to "unknown").
func stableMachineID() string {
	// macOS: IOPlatformUUID (stable across reboots, network changes, hostname changes)
	out, err := exec.Command("ioreg", "-d2", "-c", "IOPlatformExpertDevice").Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "\"IOPlatformUUID\"") {
				parts := strings.Split(line, "\"")
				if len(parts) >= 4 {
					return parts[3]
				}
			}
		}
	}
	// Linux: /etc/machine-id
	if data, err := os.ReadFile("/etc/machine-id"); err == nil {
		if s := strings.TrimSpace(string(data)); s != "" {
			return s
		}
	}
	return ""
}

// setCmd stores a secret
func setCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <name> <value>",
		Short: "Save an encrypted secret",
		Long: `Save an encrypted secret. Use "-" as value to read from stdin (avoids argv exposure).

Examples:
  sw secret set api_key "sk_test_123"
  echo "sk_test_123" | sw secret set api_key -
  SOLO_SECRET_VALUE=sk_test sw secret set api_key '$SOLO_SECRET_VALUE'
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			value := args[1]

			if v := os.Getenv("SOLO_SECRET_VALUE"); v != "" && value == "$SOLO_SECRET_VALUE" {
				value = v
			}
			if value == "-" {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("read stdin: %w", err)
				}
				value = strings.TrimSpace(string(data))
				if value == "" {
					return fmt.Errorf("empty secret value from stdin")
				}
			}

			if err := globalVault.Set(name, value); err != nil {
				return fmt.Errorf("set secret: %w", err)
			}

			fmt.Printf("%s %s secret '%s' stored\n", core.Success("✓"), core.Info("Secret"), name)
			return nil
		},
	}
	return cmd
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
