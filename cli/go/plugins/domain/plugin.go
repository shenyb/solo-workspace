package domain

import (
	"fmt"
	"sort"

	core "github.com/shenyb/solo-workspace/cli/go/internal"
	"github.com/spf13/cobra"
)

// Cmd returns the domain command tree.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "domain",
		Short: "Domain management",
		Long:  `List, add, and delete configured domains.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listDomains()
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all domains",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listDomains()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "add <domain>",
		Short: "Add a domain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return addDomain(args[0])
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "delete <domain>",
		Short: "Delete a domain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteDomain(args[0])
		},
	})

	return cmd
}

func listDomains() error {
	cfg := core.CurrentConfig
	if cfg == nil {
		cfg = core.DefaultConfig()
	}
	if len(cfg.Domains) == 0 {
		fmt.Println("No domains configured.")
		return nil
	}

	domains := append([]string(nil), cfg.Domains...)
	sort.Strings(domains)

	columns := []string{"Domain"}
	rows := make([][]string, 0, len(domains))
	for _, domain := range domains {
		rows = append(rows, []string{domain})
	}
	core.Table(columns, rows)
	return nil
}

func addDomain(domain string) error {
	cfg := core.CurrentConfig
	if cfg == nil {
		cfg = core.DefaultConfig()
	}
	for _, existing := range cfg.Domains {
		if existing == domain {
			return fmt.Errorf("domain %q already exists", domain)
		}
	}
	cfg.Domains = append(cfg.Domains, domain)
	core.CurrentConfig = cfg

	if err := core.SaveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("✅ Domain %q added\n", domain)
	return nil
}

func deleteDomain(domain string) error {
	cfg := core.CurrentConfig
	if cfg == nil {
		return fmt.Errorf("domain %q not found", domain)
	}
	index := -1
	for i, existing := range cfg.Domains {
		if existing == domain {
			index = i
			break
		}
	}
	if index == -1 {
		return fmt.Errorf("domain %q not found", domain)
	}
	cfg.Domains = append(cfg.Domains[:index], cfg.Domains[index+1:]...)
	core.CurrentConfig = cfg

	if err := core.SaveConfig(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("✅ Domain %q deleted\n", domain)
	return nil
}
