package ssl

import (
	"crypto/x509"
	"testing"
)

func TestCertInfoFromPeerCertsEmptyChain(t *testing.T) {
	info := certInfoFromPeerCerts("example.com", nil)
	if info.Err != "no peer certificate presented" {
		t.Fatalf("Err = %q", info.Err)
	}
	if info.Domain != "example.com" {
		t.Fatalf("Domain = %q", info.Domain)
	}
}

func TestCertInfoFromPeerCertsEmptySlice(t *testing.T) {
	info := certInfoFromPeerCerts("example.com", []*x509.Certificate{})
	if info.Err != "no peer certificate presented" {
		t.Fatalf("Err = %q", info.Err)
	}
}
