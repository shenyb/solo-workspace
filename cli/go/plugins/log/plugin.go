package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	core "github.com/shenyb/solo-workspace/cli/go/internal"
	"github.com/spf13/cobra"

	"gopkg.in/yaml.v3"
)

// LogEntry records one timestamped line.
type LogEntry struct {
	Time time.Time `yaml:"time"`
	Text string    `yaml:"text"`
}

// LogStore holds all log entries.
type LogStore struct {
	Entries []LogEntry `yaml:"entries"`
	path    string
}

var store *LogStore

func logPath() (string, error) {
	dir, err := core.DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "log.yaml"), nil
}

func loadStore() error {
	if store != nil {
		return nil
	}
	path, err := logPath()
	if err != nil {
		return err
	}
	store = &LogStore{path: path}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read log: %w", err)
	}
	if err := yaml.Unmarshal(data, store); err != nil {
		return fmt.Errorf("parse log: %w", err)
	}
	if store.Entries == nil {
		store.Entries = []LogEntry{}
	}
	return nil
}

func saveStore() error {
	data, err := yaml.Marshal(store)
	if err != nil {
		return fmt.Errorf("marshal log: %w", err)
	}
	if err := core.WriteFileAtomic(store.path, data, 0600); err != nil {
		return fmt.Errorf("write log: %w", err)
	}
	return nil
}

// Cmd returns the log command tree.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Quick time log — record what you did",
		Long: `Record timestamped log entries for daily tracking.

Examples:
  sw log "修复登录 OAuth bug"
  sw log add "fixed bug" --at 14:30
  sw log today
  sw log since 3d
  sw log list
`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return loadStore()
		},
	}

	addCmd := &cobra.Command{
		Use:       "add <text>",
		Short:     "Add a log entry with current timestamp",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{},
		RunE: func(cmd *cobra.Command, args []string) error {
			at, _ := cmd.Flags().GetString("at")
			return addLog(args[0], at)
		},
	}
	addCmd.Flags().String("at", "", "Set time today (format: 15:04)")
	cmd.AddCommand(addCmd)

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List recent log entries (last 20)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listLogs("", 20)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "today",
		Short: "Show today's log entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listLogs("today", 0)
		},
	})

	sinceCmd := &cobra.Command{
		Use:   "since <duration>",
		Short: "Show log entries since duration (e.g. 3d, 7d, 24h)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return listLogs(args[0], 0)
		},
	}
	cmd.AddCommand(sinceCmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return addLog(strings.Join(args, " "), "")
	}

	return cmd
}

func addLog(text, at string) error {
	ts := time.Now()
	if at != "" {
		parsed, err := time.Parse("15:04", at)
		if err != nil {
			return fmt.Errorf("invalid --at %q: use format 15:04", at)
		}
		now := time.Now()
		ts = time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), 0, 0, now.Location())
	}

	store.Entries = append(store.Entries, LogEntry{
		Time: ts,
		Text: text,
	})

	if len(store.Entries) > 500 {
		store.Entries = store.Entries[len(store.Entries)-500:]
	}

	if err := saveStore(); err != nil {
		return err
	}
	fmt.Printf("📝 Logged: %s\n", text)
	return nil
}

func startOfToday(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

func listLogs(filter string, limit int) error {
	if len(store.Entries) == 0 {
		fmt.Println("No log entries yet. Use: sw log \"what you did\"")
		return nil
	}

	var filtered []LogEntry
	now := time.Now()
	todayStart := startOfToday(now)

	for _, e := range store.Entries {
		switch {
		case filter == "today":
			if e.Time.Before(todayStart) {
				continue
			}
		case filter != "" && filter != "today":
			d, err := core.ParseDurationDays(filter)
			if err != nil {
				return fmt.Errorf("invalid duration %q, use e.g. 3d, 24h, 7d", filter)
			}
			if now.Sub(e.Time) > d {
				continue
			}
		}
		filtered = append(filtered, e)
	}

	if len(filtered) == 0 {
		fmt.Println("No log entries found for the given filter.")
		return nil
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Time.After(filtered[j].Time)
	})

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	fmt.Println("📋 Log entries:")
	for _, e := range filtered {
		fmt.Printf("  %s  %s\n", e.Time.Format("01-02 15:04"), e.Text)
	}
	return nil
}
