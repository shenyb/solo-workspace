package core

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/smtp"
	"strings"
)

// smtpPlainAuth implements PLAIN authentication. For implicit TLS (port 465 / SMTPS),
// net/smtp.Client.TLS stays false because StartTLS() is never called; implicitTLS
// allows auth on an already-encrypted connection.
type smtpPlainAuth struct {
	identity, username, password, host string
	implicitTLS                        bool
}

func (a *smtpPlainAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS && !a.implicitTLS {
		return "", nil, errors.New("unencrypted connection")
	}
	if server.Name != a.host {
		return "", nil, errors.New("wrong host name")
	}
	resp := []byte(a.identity)
	resp = append(resp, 0)
	resp = append(resp, []byte(a.username)...)
	resp = append(resp, 0)
	resp = append(resp, []byte(a.password)...)
	return "PLAIN", resp, nil
}

func (a *smtpPlainAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		return nil, errors.New("unexpected server challenge")
	}
	return nil, nil
}

func smtpPlainAuthFor(email *EmailConfig, password string, implicitTLS bool) smtp.Auth {
	return &smtpPlainAuth{
		identity:    "",
		username:    email.Username,
		password:    password,
		host:        email.Host,
		implicitTLS: implicitTLS,
	}
}

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
	password, err := resolveEmailPassword(email)
	if err != nil {
		return err
	}

	port := email.Port
	if port == 0 {
		port = 587
	}
	addr := fmt.Sprintf("%s:%d", email.Host, port)

	// Port 465 (and explicit SMTPS) use implicit TLS. STARTTLS-based SendMail path
	// does not apply; auth must tolerate Client.TLS == false after tls.Dial.
	if port == 465 || (email.UseTLS && port != 587) {
		return sendSMTPS(email, addr, password, msg)
	}

	var auth smtp.Auth
	if email.Username != "" || password != "" {
		auth = smtp.PlainAuth("", email.Username, password, email.Host)
	}
	if err := smtp.SendMail(addr, auth, email.From, email.To, msg); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	return nil
}

func sendSMTPS(email *EmailConfig, addr, password string, msg []byte) error {
	var auth smtp.Auth
	if email.Username != "" || password != "" {
		auth = smtpPlainAuthFor(email, password, true)
	}

	tlsConfig := &tls.Config{ServerName: email.Host}
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("dial smtp server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, email.Host)
	if err != nil {
		return fmt.Errorf("create smtp client: %w", err)
	}
	defer func() { _ = client.Quit() }()

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
	if _, err := writer.Write(msg); err != nil {
		return fmt.Errorf("smtp write message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("smtp close writer: %w", err)
	}
	return nil
}

func resolveEmailPassword(email *EmailConfig) (string, error) {
	if email.PasswordSecret != "" {
		if SecretResolver == nil {
			return "", fmt.Errorf("password_secret %q configured but secret vault is unavailable", email.PasswordSecret)
		}
		return SecretResolver(email.PasswordSecret)
	}
	return email.Password, nil
}
