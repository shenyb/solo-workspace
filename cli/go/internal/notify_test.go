package core

import (
	"net/smtp"
	"testing"
)

func TestSMTPPlainAuthImplicitTLS(t *testing.T) {
	auth := &smtpPlainAuth{
		username:    "user@example.com",
		password:    "secret",
		host:        "smtp.example.com",
		implicitTLS: true,
	}
	mech, _, err := auth.Start(&smtp.ServerInfo{Name: "smtp.example.com", TLS: false})
	if err != nil {
		t.Fatalf("implicit TLS auth should succeed without ServerInfo.TLS: %v", err)
	}
	if mech != "PLAIN" {
		t.Fatalf("mechanism = %q", mech)
	}
}

func TestSMTPPlainAuthRequiresTLSWithoutImplicit(t *testing.T) {
	auth := &smtpPlainAuth{
		username:    "user@example.com",
		password:    "secret",
		host:        "smtp.example.com",
		implicitTLS: false,
	}
	_, _, err := auth.Start(&smtp.ServerInfo{Name: "smtp.example.com", TLS: false})
	if err == nil {
		t.Fatal("expected error for plaintext connection without implicitTLS")
	}
}

func TestSMTPPlainAuthWrongHost(t *testing.T) {
	auth := &smtpPlainAuth{
		username:    "user@example.com",
		password:    "secret",
		host:        "smtp.example.com",
		implicitTLS: true,
	}
	_, _, err := auth.Start(&smtp.ServerInfo{Name: "evil.example.com", TLS: true})
	if err == nil {
		t.Fatal("expected wrong host name error")
	}
}
