package notify

import (
	"fmt"

	core "github.com/shenyb/solo-workspace/cli/go/internal"
	"github.com/spf13/cobra"
)

// Cmd returns the notify command tree.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notify",
		Short: "Email notification management",
		Long:  `Send email notifications based on configured SMTP settings.`,
	}

	sendCmd := &cobra.Command{
		Use:   "send <subject> <body>",
		Short: "Send a notification email",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			subject := args[0]
			body := args[1]
			return send(subject, body)
		},
	}

	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Send a test notification email",
		RunE: func(cmd *cobra.Command, args []string) error {
			return send("SW Notification Test", "This is a test email from sw.")
		},
	}

	cmd.AddCommand(sendCmd)
	cmd.AddCommand(testCmd)
	return cmd
}

func send(subject, body string) error {
	if err := core.SendEmail(subject, body); err != nil {
		return fmt.Errorf("send notification: %w", err)
	}
	fmt.Println("✅ Email notification sent")
	return nil
}
