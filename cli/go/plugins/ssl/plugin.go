package ssl

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/shenyb/solo-workspace/cli/go/internal"
	"github.com/spf13/cobra"
)

// CertInfo holds SSL certificate check results.
type CertInfo struct {
	Domain   string
	Expires  time.Time
	DaysLeft int
	Issuer   string
	Err      string
}

// Cmd returns the "ssl" command tree.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssl",
		Short: "SSL certificate management",
		Long:  `Check and manage SSL/TLS certificates for your domains.`,
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "check",
		Short: "Check SSL certificate expiration for all domains",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := core.CurrentConfig
			if cfg == nil {
				cfg = core.DefaultConfig()
			}
			return checkCertificates(cfg)
		},
	})

	return cmd
}

func checkCertificates(cfg *core.Config) error {
	if len(cfg.Domains) == 0 {
		fmt.Println("No domains configured. Edit .solo.yaml or ~/.solo/config.yaml")
		return nil
	}

	var results []CertInfo
	for _, domain := range cfg.Domains {
		info := checkDomain(domain)
		results = append(results, info)
	}

	columns := []string{"Domain", "Expires", "Days Left", "Issuer"}
	var rows [][]string
	for _, r := range results {
		if r.Err != "" {
			rows = append(rows, []string{r.Domain, "ERROR", "-", r.Err})
		} else {
			rows = append(rows, []string{
				r.Domain,
				r.Expires.Format("2006-01-02"),
				fmt.Sprintf("%d", r.DaysLeft),
				r.Issuer,
			})
		}
	}
	core.Table(columns, rows)

	soon := false
	for _, r := range results {
		if r.DaysLeft >= 0 && r.DaysLeft <= 30 {
			if !soon {
				fmt.Println("\n⚠️  Certificates expiring within 30 days:")
				soon = true
			}
			fmt.Printf("  %s — %d days left\n", r.Domain, r.DaysLeft)
		}
	}

	return nil
}

func checkDomain(domain string) CertInfo {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", domain+":443", &tls.Config{
		InsecureSkipVerify: false,
	})
	if err != nil {
		return CertInfo{Domain: domain, Err: err.Error()}
	}
	defer conn.Close()

	cert := conn.ConnectionState().PeerCertificates[0]
	daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)

	return CertInfo{
		Domain:   domain,
		Expires:  cert.NotAfter,
		DaysLeft: daysLeft,
		Issuer:   cert.Issuer.CommonName,
	}
}
