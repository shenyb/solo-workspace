package secret

import (
	"crypto/sha256"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/pbkdf2"
)

func TestWrongPasswordDoesNotClearVault(t *testing.T) {
	dir := t.TempDir()

	v1, err := NewSecretVault(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := v1.SetMasterPassword("correct-password"); err != nil {
		t.Fatal(err)
	}
	if err := v1.Set("api_key", "sk_live_123"); err != nil {
		t.Fatal(err)
	}

	orig, err := os.ReadFile(filepath.Join(dir, "secrets.enc"))
	if err != nil {
		t.Fatal(err)
	}

	v2, err := NewSecretVault(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := v2.SetMasterPassword("wrong-password"); err == nil {
		t.Fatal("expected wrong password to fail")
	}
	if !v2.locked {
		t.Fatal("expected vault to be locked after wrong password")
	}
	if v2.encryptedData == "" {
		t.Fatal("encryptedData must be preserved after wrong password")
	}

	if err := v2.Set("other", "value"); err == nil {
		t.Fatal("expected Set to fail on locked vault")
	}
	if err := v2.Save(); err == nil {
		t.Fatal("expected Save to fail on locked vault")
	}

	after, err := os.ReadFile(filepath.Join(dir, "secrets.enc"))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(orig) {
		t.Fatal("vault file must not change after wrong password")
	}

	v3, err := NewSecretVault(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := v3.SetMasterPassword("correct-password"); err != nil {
		t.Fatalf("correct password should unlock vault: %v", err)
	}
	val, err := v3.Get("api_key")
	if err != nil {
		t.Fatal(err)
	}
	if val != "sk_live_123" {
		t.Fatalf("got %q, want original secret", val)
	}
}

func TestSaveRefusesUndecryptedBlob(t *testing.T) {
	sv := &SecretVault{
		path:          filepath.Join(t.TempDir(), "secrets.enc"),
		salt:          make([]byte, 32),
		masterKey:     make([]byte, 32),
		secrets:       map[string]string{},
		encryptedData: "dGVzdA==",
	}
	if err := sv.Save(); err == nil {
		t.Fatal("expected Save to refuse while encryptedData is still set")
	}
}

func TestEmptyVaultFileFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.enc")
	wrapper := vaultFile{Salt: "c2FsdA==", Data: "encrypted-blob"}
	raw, _ := json.Marshal(wrapper)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	sv, err := NewSecretVault(dir)
	if err != nil {
		t.Fatal(err)
	}
	if sv.encryptedData != "encrypted-blob" {
		t.Fatalf("encryptedData = %q", sv.encryptedData)
	}
}

func TestDeleteWithoutMasterKeyDoesNotMutate(t *testing.T) {
	sv := &SecretVault{
		secrets: map[string]string{"api_key": "sk_live_123"},
	}
	if err := sv.Delete("api_key"); err == nil {
		t.Fatal("expected Delete to fail without master key")
	}
	if _, ok := sv.secrets["api_key"]; !ok {
		t.Fatal("secret must remain in memory after failed Delete")
	}
}

func TestDeletePersistedOnlyAfterSuccessfulSave(t *testing.T) {
	dir := t.TempDir()
	v, err := NewSecretVault(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := v.SetMasterPassword("pw"); err != nil {
		t.Fatal(err)
	}
	if err := v.Set("keep", "a"); err != nil {
		t.Fatal(err)
	}
	if err := v.Set("drop", "b"); err != nil {
		t.Fatal(err)
	}

	v.masterKey = nil
	if err := v.Delete("drop"); err == nil {
		t.Fatal("expected Delete to fail without master key")
	}
	if _, ok := v.secrets["drop"]; !ok {
		t.Fatal("drop should still exist after failed Delete")
	}

	v.masterKey = pbkdf2Key("pw", v.salt)
	if err := v.Delete("drop"); err != nil {
		t.Fatal(err)
	}

	v2, err := NewSecretVault(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := v2.SetMasterPassword("pw"); err != nil {
		t.Fatal(err)
	}
	if _, err := v2.Get("drop"); err == nil {
		t.Fatal("drop should be gone from disk")
	}
	if val, err := v2.Get("keep"); err != nil || val != "a" {
		t.Fatalf("keep = %q err = %v", val, err)
	}
}

func pbkdf2Key(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, 100000, 32, sha256.New)
}
