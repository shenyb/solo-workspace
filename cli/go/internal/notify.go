package core

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

// SendEmail sends an email notification according to the current config.
func SendEmail(subject, body string) error {
	if CurrentConfig == nil || CurrentConfig.Notify == nil || CurrentConfig.Notify.Email == nil {
		return fmt.Errorf("email notification is not configured")
	}

	email := CurrentConfig.Notify.Email
	if !email.Enabled {
		return fmt.Errorf("email notification is disabled")
	}
	if email.Host == "" {
		return fmt.Errorf("smtp host is not configured")
	}
	if email.Port == 0 {
		email.Port = 587
	}
	if email.From == "" {
		return fmt.Errorf("email from address is not configured")
	}
	if len(email.To) == 0 {
		return fmt.Errorf("email recipient list is empty")
	}

	message := composeEmailMessage(email.From, email.To, subject, body)
	return sendSMTP(email, []byte(message))
}

func composeEmailMessage(from string, to []string, subject, body string) string {
	headers := []string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", strings.Join(to, ",")),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=utf-8",
		"",
	}
	headers = append(headers, body)
	return strings.Join(headers, "\r\n")
}

func sendSMTP(email *EmailConfig, msg []byte) error {
	addr := fmt.Sprintf("%s:%d", email.Host, email.Port)
	var auth smtp.Auth
	if email.Username != "" || email.Password != "" {
		auth = smtp.PlainAuth("", email.Username, email.Password, email.Host)
	}

	if email.Port == 465 || email.UseTLS {
		tlsConfig := &tls.Config{
			ServerName: email.Host,
		}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("dial smtp server: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, email.Host)
		if err != nil {
			return fmt.Errorf("create smtp client: %w", err)
		}
		defer client.Quit()

		if auth != nil {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("smtp auth: %w", err)
			}
		}

		if err := client.Mail(email.From); err != nil {
			return fmt.Errorf("smtp mail from: %w", err)
		}
		for _, recipient := range email.To {
			if err := client.Rcpt(recipient); err != nil {
				return fmt.Errorf("smtp rcpt %s: %w", recipient, err)
			}
		}

		writer, err := client.Data()
		if err != nil {
			return fmt.Errorf("smtp data: %w", err)
		}
		_, err = writer.Write(msg)
		if err != nil {
			return fmt.Errorf("smtp write message: %w", err)
		}
		if err := writer.Close(); err != nil {
			return fmt.Errorf("smtp close writer: %w", err)
		}

		return client.Quit()
	}

	if err := smtp.SendMail(addr, auth, email.From, email.To, msg); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	return nil
}
