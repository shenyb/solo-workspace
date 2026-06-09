package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/pbkdf2"
)

// SecretVault manages encrypted secrets
type SecretVault struct {
	path      string
	masterKey []byte
	secrets   map[string]string
}

// NewSecretVault creates a new secret vault in the given data directory.
func NewSecretVault(dataDir string) (*SecretVault, error) {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, fmt.Errorf("create vault dir: %w", err)
	}

	secretsPath := filepath.Join(dataDir, "secrets.enc")
	sv := &SecretVault{
		path:    secretsPath,
		secrets: make(map[string]string),
	}

	// Load existing secrets
	if err := sv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return sv, nil
}

// SetMasterPassword sets the master password for encryption/decryption
func (sv *SecretVault) SetMasterPassword(password string) {
	// Derive key from password using PBKDF2
	// Use a fixed salt for now (in production, save salt separately)
	salt := []byte("solo-workspace-secret-salt")
	sv.masterKey = pbkdf2.Key([]byte(password), salt, 100000, 32, sha256.New)
}

// Set stores an encrypted secret
func (sv *SecretVault) Set(key, value string) error {
	if len(sv.masterKey) == 0 {
		return fmt.Errorf("master password not set")
	}

	// Encrypt the value
	encrypted, err := sv.encrypt(value)
	if err != nil {
		return err
	}

	sv.secrets[key] = encrypted

	// Persist to disk
	return sv.Save()
}

// Get retrieves and decrypts a secret
func (sv *SecretVault) Get(key string) (string, error) {
	if len(sv.masterKey) == 0 {
		return "", fmt.Errorf("master password not set")
	}

	encrypted, exists := sv.secrets[key]
	if !exists {
		return "", fmt.Errorf("secret not found: %s", key)
	}

	return sv.decrypt(encrypted)
}

// Delete removes a secret
func (sv *SecretVault) Delete(key string) error {
	delete(sv.secrets, key)
	return sv.Save()
}

// List returns all secret keys (without values)
func (sv *SecretVault) List() []string {
	keys := make([]string, 0, len(sv.secrets))
	for k := range sv.secrets {
		keys = append(keys, k)
	}
	return keys
}

// encrypt encrypts plaintext using AES-256-GCM
func (sv *SecretVault) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(sv.masterKey)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts ciphertext using AES-256-GCM
func (sv *SecretVault) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(sv.masterKey)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesgcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, encrypted := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// Save persists secrets to disk (encrypted)
func (sv *SecretVault) Save() error {
	data, err := json.MarshalIndent(sv.secrets, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(sv.path, data, 0o600)
}

// Load reads secrets from disk
func (sv *SecretVault) Load() error {
	data, err := os.ReadFile(sv.path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &sv.secrets)
}
