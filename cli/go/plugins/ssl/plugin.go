package ssl

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"

	core "github.com/shenyb/solo-workspace/cli/go/internal"
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

	var soonResults []CertInfo
	for _, r := range results {
		if r.DaysLeft >= 0 && r.DaysLeft <= 30 {
			soonResults = append(soonResults, r)
		}
	}

	if len(soonResults) > 0 {
		fmt.Println("\n⚠️  Certificates expiring within 30 days:")
		for _, r := range soonResults {
			fmt.Printf("  %s — %d days left\n", r.Domain, r.DaysLeft)
		}

		if cfg.Notify == nil || (cfg.Notify.Webhook == "" && (cfg.Notify.Email == nil || !cfg.Notify.Email.Enabled)) {
			return nil
		}

		var domainNames []string
		for _, r := range soonResults {
			domainNames = append(domainNames, r.Domain)
		}
		pending, err := core.FilterSSLNotifyDomains(domainNames)
		if err != nil {
			fmt.Printf("Warning: ssl notify state: %v\n", err)
			pending = domainNames
		}
		if len(pending) == 0 {
			fmt.Println("(notification already sent today for these domains)")
			return nil
		}

		pendingSet := make(map[string]CertInfo, len(pending))
		for _, r := range soonResults {
			for _, d := range pending {
				if r.Domain == d {
					pendingSet[d] = r
				}
			}
		}

		subject := "[sw] SSL Certificate Expiry Warning"
		body := "The following certificates are expiring soon:\n\n"
		for _, d := range pending {
			r := pendingSet[d]
			body += fmt.Sprintf("- %s: %d days left\n", r.Domain, r.DaysLeft)
		}

		if err := core.NotifyAlert(subject, body); err != nil {
			fmt.Printf("Warning: failed to send notification: %v\n", err)
		} else {
			if err := core.MarkSSLNotifiedBatch(pending); err != nil {
				fmt.Printf("Warning: failed to record ssl notification state: %v\n", err)
			}
			fmt.Println("✅ Notification sent")
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

	return certInfoFromPeerCerts(domain, conn.ConnectionState().PeerCertificates)
}

func certInfoFromPeerCerts(domain string, certs []*x509.Certificate) CertInfo {
	if len(certs) == 0 {
		return CertInfo{Domain: domain, Err: "no peer certificate presented"}
	}
	cert := certs[0]
	daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)

	return CertInfo{
		Domain:   domain,
		Expires:  cert.NotAfter,
		DaysLeft: daysLeft,
		Issuer:   cert.Issuer.CommonName,
	}
}
