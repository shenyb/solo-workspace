package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	core "github.com/shenyb/solo-workspace/cli/go/internal"
	"github.com/shenyb/solo-workspace/cli/go/plugins/config"
	"github.com/shenyb/solo-workspace/cli/go/plugins/domain"
	"github.com/shenyb/solo-workspace/cli/go/plugins/env"
	"github.com/shenyb/solo-workspace/cli/go/plugins/log"
	"github.com/shenyb/solo-workspace/cli/go/plugins/notify"
	"github.com/shenyb/solo-workspace/cli/go/plugins/project"
	"github.com/shenyb/solo-workspace/cli/go/plugins/secret"
	"github.com/shenyb/solo-workspace/cli/go/plugins/server"
	"github.com/shenyb/solo-workspace/cli/go/plugins/ssl"
	"github.com/shenyb/solo-workspace/cli/go/plugins/todo"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "sw",
	Short: "Solo Workspace — the indie developer's operating system",
	Long: `A unified command center for indie developers.

Manage servers, domains, SSL certificates, deployments,
Docker containers, and more — all from your terminal.

Run without arguments to show an overview of all resources.
Use "sw tui" to enter the interactive menu.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if isCompletionCommand(cmd) {
			return nil
		}
		var err error
		core.CurrentConfig, err = core.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := core.CurrentConfig
		if cfg == nil {
			cfg = core.DefaultConfig()
		}
		printAll(cfg)
		return nil
	},
}

func init() {
	// Run root PersistentPreRun (config load) before plugin-level hooks (env, secret).
	cobra.EnableTraverseRunHooks = true

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")

	// Register plugins — each is a Lego brick
	rootCmd.AddCommand(project.Cmd())
	rootCmd.AddCommand(ssl.Cmd())
	rootCmd.AddCommand(server.Cmd())
	rootCmd.AddCommand(domain.Cmd())
	rootCmd.AddCommand(todo.Cmd())
	rootCmd.AddCommand(notify.Cmd())
	rootCmd.AddCommand(secret.Cmd())
	rootCmd.AddCommand(config.Cmd())
	rootCmd.AddCommand(env.Cmd())
	rootCmd.AddCommand(log.Cmd())
	rootCmd.AddCommand(allCmd())
	rootCmd.AddCommand(tuiCmd())
	rootCmd.AddCommand(completionCmd())
}

func allCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "all",
		Short: "Show all configured projects, servers, and domains",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := core.CurrentConfig
			if cfg == nil {
				cfg = core.DefaultConfig()
			}
			printAll(cfg)
			return nil
		},
	}
}

func printAll(cfg *core.Config) {
	// Projects
	if len(cfg.Projects) == 0 {
		fmt.Println(core.Warn("No projects configured."))
	} else {
		title := core.Info("Projects:") + " " + core.Success(fmt.Sprintf("(%d)", len(cfg.Projects)))
		fmt.Println(title)
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
	}

	// Todos
	if len(cfg.Todos) == 0 {
		fmt.Println(core.Warn("No todos."))
	} else {
		title := core.Info("Todos:") + " " + core.Success(fmt.Sprintf("(%d)", len(cfg.Todos)))
		fmt.Println(title)
		columns := []string{"ID", "Name", "Status", "Description", "Created", "Updated"}
		rows := make([][]string, 0, len(cfg.Todos))
		for _, entry := range core.SortedTodos(cfg) {
			rows = append(rows, []string{
				fmt.Sprintf("%d", entry.Config.ID),
				entry.Name,
				core.TodoStatus(entry.Config.Done),
				entry.Config.Description,
				core.FormatTodoTime(entry.Config.CreatedAt),
				core.FormatTodoTime(entry.Config.UpdatedAt),
			})
		}
		core.Table(columns, rows)
	}

	// Servers
	if len(cfg.Servers) == 0 {
		fmt.Println(core.Warn("No servers configured."))
	} else {
		title := core.Info("Servers:") + " " + core.Success(fmt.Sprintf("(%d)", len(cfg.Servers)))
		fmt.Println(title)
		columns := []string{"Name", "Host", "User", "Port"}
		names := make([]string, 0, len(cfg.Servers))
		for name := range cfg.Servers {
			names = append(names, name)
		}
		sort.Strings(names)
		rows := make([][]string, 0, len(cfg.Servers))
		for _, name := range names {
			srv := cfg.Servers[name]
			if srv == nil {
				continue
			}
			port := 22
			if srv.Port > 0 {
				port = srv.Port
			}
			rows = append(rows, []string{name, srv.Host, srv.User, fmt.Sprintf("%d", port)})
		}
		core.Table(columns, rows)
	}

	// Domains
	if len(cfg.Domains) == 0 {
		fmt.Println(core.Warn("No domains configured."))
	} else {
		title := core.Info("Domains:") + " " + core.Success(fmt.Sprintf("(%d)", len(cfg.Domains)))
		fmt.Println(title)
		columns := []string{"Domain"}
		domains := append([]string(nil), cfg.Domains...)
		sort.Strings(domains)
		rows := make([][]string, 0, len(domains))
		for _, domain := range domains {
			rows = append(rows, []string{domain})
		}
		core.Table(columns, rows)
	}

	// Notifications
	if cfg.Notify == nil || (cfg.Notify.Webhook == "" && cfg.Notify.Email == nil) {
		fmt.Println(core.Warn("No notification channels configured."))
	} else {
		title := core.Info("Notifications:")
		fmt.Println(title)
		columns := []string{"Type", "Value"}
		rows := make([][]string, 0)
		if cfg.Notify.Webhook != "" {
			rows = append(rows, []string{"webhook", cfg.Notify.Webhook})
		}
		if cfg.Notify.Email != nil {
			var emailVal string
			if cfg.Notify.Email.Enabled {
				emailVal = "enabled"
			} else {
				emailVal = "configured (disabled)"
			}
			rows = append(rows, []string{"email", emailVal})
		}
		core.Table(columns, rows)
	}

	// Environment Variables
	envNames := env.ListEnvVarNames()
	if len(envNames) == 0 {
		fmt.Println(core.Warn("No environment variables."))
	} else {
		title := core.Info("Environment Variables:") + " " + core.Success(fmt.Sprintf("(%d)", len(envNames)))
		fmt.Println(title)
		columns := []string{"Name"}
		rows := make([][]string, 0, len(envNames))
		for _, name := range envNames {
			rows = append(rows, []string{name})
		}
		core.Table(columns, rows)
	}

	// Secrets
	if err := secret.InitVault(); err != nil {
		fmt.Println(core.Warn("Secrets unavailable:"), err)
	} else if secretNames, err := secret.ListSecrets(); err != nil {
		fmt.Println(core.Warn("Secrets unavailable:"), err)
	} else if len(secretNames) == 0 {
		fmt.Println(core.Warn("No secrets."))
	} else {
		title := core.Info("Secrets:") + " " + core.Success(fmt.Sprintf("(%d)", len(secretNames)))
		fmt.Println(title)
		columns := []string{"Name"}
		sort.Strings(secretNames)
		rows := make([][]string, 0, len(secretNames))
		for _, name := range secretNames {
			rows = append(rows, []string{name})
		}
		core.Table(columns, rows)
	}
}

func completionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Manage shell completion scripts",
		Long: `Generate or install shell completion scripts for sw.

Examples:
  sw completion bash > /etc/bash_completion.d/sw
  sw completion zsh > ~/.zshrc
  sw completion install zsh
  sw completion install powershell
  sw completion uninstall zsh
`,
		Args: cobra.NoArgs,
	}

	cmd.AddCommand(shellCmd("bash", "Generate Bash completion script", func(w io.Writer) error {
		return rootCmd.GenBashCompletion(w)
	}))
	cmd.AddCommand(shellCmd("zsh", "Generate Zsh completion script", func(w io.Writer) error {
		return rootCmd.GenZshCompletion(w)
	}))
	cmd.AddCommand(shellCmd("fish", "Generate Fish completion script", func(w io.Writer) error {
		return rootCmd.GenFishCompletion(w, true)
	}))
	cmd.AddCommand(shellCmd("powershell", "Generate PowerShell completion script", func(w io.Writer) error {
		return rootCmd.GenPowerShellCompletionWithDesc(w)
	}))
	cmd.AddCommand(installCmd())
	cmd.AddCommand(uninstallCmd())
	return cmd
}

func shellCmd(name, short string, generator func(io.Writer) error) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return generator(os.Stdout)
		},
	}
}

func installCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install <bash|zsh|fish|powershell>",
		Short: "Install shell completion for a supported shell",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return installCompletion(args[0])
		},
	}
}

func uninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall <bash|zsh|fish|powershell>",
		Short: "Remove shell completion for a supported shell",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return uninstallCompletion(args[0])
		},
	}
}

func installCompletion(shell string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot detect home directory: %w", err)
	}

	switch shell {
	case "bash":
		return installBashCompletion(home)
	case "zsh":
		return installZshCompletion(home)
	case "fish":
		return installFishCompletion(home)
	case "powershell":
		return installPowerShellCompletion(home)
	default:
		return fmt.Errorf("unsupported shell %q, must be one of bash, zsh, fish, powershell", shell)
	}
}

func uninstallCompletion(shell string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot detect home directory: %w", err)
	}

	switch shell {
	case "bash":
		return uninstallBashCompletion(home)
	case "zsh":
		return uninstallZshCompletion(home)
	case "fish":
		return uninstallFishCompletion(home)
	case "powershell":
		return uninstallPowerShellCompletion(home)
	default:
		return fmt.Errorf("unsupported shell %q, must be one of bash, zsh, fish, powershell", shell)
	}
}

func installBashCompletion(home string) error {
	path := filepath.Join(home, ".bash_completion.d")
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	file := filepath.Join(path, "sw")
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := rootCmd.GenBashCompletion(f); err != nil {
		return err
	}

	rcPath := bashRCPath(home)
	snippet := `# sw completion bash start
if [ -f "$HOME/.bash_completion.d/sw" ]; then
  . "$HOME/.bash_completion.d/sw"
fi
# sw completion bash end
`
	if err := appendUnique(rcPath, snippet); err != nil {
		return err
	}
	return nil
}

func uninstallBashCompletion(home string) error {
	_ = os.Remove(filepath.Join(home, ".bash_completion.d", "sw"))
	rcPath := bashRCPath(home)
	return removeBlock(rcPath, "# sw completion bash start", "# sw completion bash end")
}

func bashRCPath(home string) string {
	if runtime.GOOS == "windows" && os.Getenv("MSYSTEM") != "" {
		return filepath.Join(home, ".bashrc")
	}

	bashrc := filepath.Join(home, ".bashrc")
	if _, err := os.Stat(bashrc); err == nil {
		return bashrc
	}

	bashProfile := filepath.Join(home, ".bash_profile")
	if _, err := os.Stat(bashProfile); err == nil {
		return bashProfile
	}

	return bashrc
}

func installZshCompletion(home string) error {
	dir := filepath.Join(home, ".zfunc")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	file := filepath.Join(dir, "_sw")
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := rootCmd.GenZshCompletion(f); err != nil {
		return err
	}

	rcPath := filepath.Join(home, ".zshrc")
	snippet := `# sw completion zsh start
fpath=("$HOME/.zfunc" $fpath)
autoload -Uz compinit
compinit
# sw completion zsh end
`
	return appendUnique(rcPath, snippet)
}

func uninstallZshCompletion(home string) error {
	_ = os.Remove(filepath.Join(home, ".zfunc", "_sw"))
	rcPath := filepath.Join(home, ".zshrc")
	return removeBlock(rcPath, "# sw completion zsh start", "# sw completion zsh end")
}

func installFishCompletion(home string) error {
	path := filepath.Join(home, ".config", "fish", "completions")
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	file := filepath.Join(path, "sw.fish")
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return rootCmd.GenFishCompletion(f, true)
}

func uninstallFishCompletion(home string) error {
	_ = os.Remove(filepath.Join(home, ".config", "fish", "completions", "sw.fish"))
	return nil
}

func installPowerShellCompletion(home string) error {
	profile := powerShellProfilePath(home)
	if err := os.MkdirAll(filepath.Dir(profile), 0o755); err != nil {
		return err
	}

	var buf strings.Builder
	if err := rootCmd.GenPowerShellCompletionWithDesc(&buf); err != nil {
		return err
	}

	snippet := fmt.Sprintf("# sw completion powershell start\n%s# sw completion powershell end\n", buf.String())
	return appendUnique(profile, snippet)
}

func uninstallPowerShellCompletion(home string) error {
	profile := powerShellProfilePath(home)
	return removeBlock(profile, "# sw completion powershell start", "# sw completion powershell end")
}

func powerShellProfilePath(home string) string {
	if runtime.GOOS == "windows" {
		documents := filepath.Join(home, "Documents")
		return filepath.Join(documents, "PowerShell", "Microsoft.PowerShell_profile.ps1")
	}
	return filepath.Join(home, ".config", "powershell", "Microsoft.PowerShell_profile.ps1")
}

func appendUnique(path, content string) error {
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if strings.Contains(string(data), content) {
		return nil
	}
	if err := os.WriteFile(path, append(data, []byte(content)...), 0o644); err != nil {
		return err
	}
	return nil
}

func removeBlock(path, start, end string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	text := string(data)
	startIndex := strings.Index(text, start)
	if startIndex == -1 {
		return nil
	}
	endIndex := strings.Index(text[startIndex:], end)
	if endIndex == -1 {
		return nil
	}
	endIndex += startIndex + len(end)

	blockStart := startIndex
	if blockStart > 0 && text[blockStart-1] == '\n' {
		blockStart--
	}

	newText := text[:blockStart] + text[endIndex:]
	return os.WriteFile(path, []byte(newText), 0o644)
}

func tuiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Enter interactive terminal menu",
		RunE: func(cmd *cobra.Command, args []string) error {
			return launchTUI()
		},
	}
}

// ── TUI (interactive menu) ────────────────────────────────

// launchTUI builds the menu and runs the interactive TUI
func launchTUI() error {
	self, err := os.Executable()
	if err != nil {
		self = os.Args[0]
	}

	withConfig := func(parts ...string) []string {
		cmd := []string{self}
		if cfgFile != "" {
			cmd = append(cmd, "-c", cfgFile)
		}
		return append(cmd, parts...)
	}

	items := []core.MenuItem{
		{
			Label:       "all",
			Description: "Show overview of all resources",
			Command:     withConfig("all"),
		},
		{
			Label:       "ssl",
			Description: "SSL certificate management",
			Children: []core.MenuItem{
				{Label: "check", Description: "Check all domain SSL certificates", Command: withConfig("ssl", "check")},
				{Label: "..", Description: "Back", IsBack: true},
			},
		},
		{
			Label:       "server",
			Description: "Server management",
			Children: []core.MenuItem{
				{Label: "list", Description: "List configured servers", Command: withConfig("server", "list")},
				{Label: "add", Description: "Add a server", Command: withConfig("server", "add")},
				{Label: "update", Description: "Update a server", Command: withConfig("server", "update")},
				{Label: "delete", Description: "Delete a server", Command: withConfig("server", "delete")},
				{Label: "ssh", Description: "SSH into a server", Command: withConfig("server", "ssh")},
				{Label: "..", Description: "Back", IsBack: true},
			},
		},
		{
			Label:       "domain",
			Description: "Domain management",
			Children: []core.MenuItem{
				{Label: "list", Description: "List configured domains", Command: withConfig("domain", "list")},
				{Label: "..", Description: "Back", IsBack: true},
			},
		},
		{
			Label:       "project",
			Description: "Project management",
			Children: []core.MenuItem{
				{Label: "list", Description: "List projects", Command: withConfig("project", "list")},
				{Label: "..", Description: "Back", IsBack: true},
			},
		},
		{
			Label:       "todo",
			Description: "Todo task management",
			Children: []core.MenuItem{
				{Label: "list", Description: "List todos", Command: withConfig("todo", "list")},
				{Label: "..", Description: "Back", IsBack: true},
			},
		},
		{
			Label:       "notify",
			Description: "Email notification",
			Children: []core.MenuItem{
				{Label: "test", Description: "Send test notification", Command: withConfig("notify", "test")},
				{Label: "..", Description: "Back", IsBack: true},
			},
		},
		{
			Label:       "secret",
			Description: "Encrypted secrets (API keys, tokens)",
			Children: []core.MenuItem{
				{Label: "list", Description: "List stored secrets", Command: withConfig("secret", "list")},
				{Label: "..", Description: "Back", IsBack: true},
			},
		},
		{
			Label:       "config",
			Description: "Config management (show/export/import/set/get)",
			Children: []core.MenuItem{
				{Label: "show", Description: "Show current config", Command: withConfig("config", "show")},
				{Label: "export", Description: "Export config to stdout", Command: withConfig("config", "export")},
				{Label: "..", Description: "Back", IsBack: true},
			},
		},
		{
			Label:       "env",
			Description: "Environment variables (.env)",
			Children: []core.MenuItem{
				{Label: "list", Description: "List stored env vars", Command: withConfig("env", "list")},
				{Label: "export", Description: "Export as .env format", Command: withConfig("env", "export")},
				{Label: "..", Description: "Back", IsBack: true},
			},
		},
		{
			Label:       "log",
			Description: "Time log",
			Children: []core.MenuItem{
				{Label: "list", Description: "List recent log entries", Command: withConfig("log", "list")},
				{Label: "today", Description: "Show today's entries", Command: withConfig("log", "today")},
				{Label: "..", Description: "Back", IsBack: true},
			},
		},
	}

	tui := core.NewTUI("sw · Solo Workspace", items)
	return tui.Run()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func isCompletionCommand(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() == "completion" {
			return true
		}
	}
	return false
}
