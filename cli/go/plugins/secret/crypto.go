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

// vaultFile is the on-disk format for the encrypted secrets.
//
//	salt:  random 32-byte salt (base64), unique per vault
//	data:  entire secrets map serialised to JSON and encrypted with AES-256-GCM (base64)
//
// Storing the salt alongside the encrypted data is safe — its purpose is to
// ensure that identical master passwords produce different encryption keys.
type vaultFile struct {
	Salt string `json:"salt"`
	Data string `json:"data"`
}

// SecretVault manages encrypted secrets.
type SecretVault struct {
	path            string
	salt            []byte
	masterKey       []byte
	encryptedData   string            // raw base64 blob from disk (new format), decrypted once masterKey is set
	legacyEncrypted map[string]string // individually encrypted values from old vault format (migrated on first write)
	secrets         map[string]string
}

// NewSecretVault creates a new secret vault in the given data directory.
// If a vault file already exists its salt and encrypted data are loaded;
// otherwise a fresh random salt is generated.
func NewSecretVault(dataDir string) (*SecretVault, error) {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, fmt.Errorf("create vault dir: %w", err)
	}

	secretsPath := filepath.Join(dataDir, "secrets.enc")
	sv := &SecretVault{
		path:    secretsPath,
		secrets: make(map[string]string),
	}

	err := sv.Load()
	if err != nil {
		if os.IsNotExist(err) {
			// New vault: generate a random salt
			sv.salt = make([]byte, 32)
			if _, err := io.ReadFull(rand.Reader, sv.salt); err != nil {
				return nil, fmt.Errorf("generate salt: %w", err)
			}
		} else {
			return nil, err
		}
	}

	return sv, nil
}

// SetMasterPassword derives the encryption key from the master password and the
// vault's salt, then decrypts any data that was loaded from disk.
//
// Legacy vaults (per-value encryption with hardcoded salt) are migrated on
// the first call: each value is decrypted with the old scheme and a fresh
// random salt is generated for future saves.
func (sv *SecretVault) SetMasterPassword(password string) {
	// ── Legacy vault migration ────────────────────────────────────
	if sv.legacyEncrypted != nil {
		// Derive key with old hardcoded salt (same as pre-fix code).
		oldSalt := []byte("solo-workspace-secret-salt")
		oldKey := pbkdf2.Key([]byte(password), oldSalt, 100000, 32, sha256.New)

		// Decrypt each legacy value.
		for k, encryptedVal := range sv.legacyEncrypted {
			plaintext, err := decryptLegacyValue(oldKey, encryptedVal)
			if err != nil {
				continue // skip undecryptable entries (wrong password, corrupted)
			}
			sv.secrets[k] = plaintext
		}

		// Generate a fresh random salt so the next Save() writes a
		// properly salted vaultFile.
		sv.salt = make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, sv.salt); err != nil {
			// Extremely unlikely — non-fatal fallback.
			copy(sv.salt, oldSalt)
		}

		// Derive new key from the fresh salt.
		sv.masterKey = pbkdf2.Key([]byte(password), sv.salt, 100000, 32, sha256.New)
		sv.legacyEncrypted = nil
		return
	}

	// ── Normal flow (new vaultFile format) ────────────────────────
	sv.masterKey = pbkdf2.Key([]byte(password), sv.salt, 100000, 32, sha256.New)

	// If an encrypted blob was loaded from disk, decrypt it now.
	if sv.encryptedData != "" {
		if decrypted, err := sv.decryptBlob(sv.encryptedData); err == nil {
			sv.secrets = decrypted
		}
		// Decryption failure (wrong password) leaves sv.secrets as an empty map.
		// The user will see "secret not found" on get/list — no data loss.
		sv.encryptedData = ""
	}
}

// Set stores a secret, persists to disk immediately.
func (sv *SecretVault) Set(key, value string) error {
	if len(sv.masterKey) == 0 {
		return fmt.Errorf("master password not set")
	}
	sv.secrets[key] = value
	return sv.Save()
}

// Get retrieves a decrypted secret.
func (sv *SecretVault) Get(key string) (string, error) {
	if len(sv.masterKey) == 0 {
		return "", fmt.Errorf("master password not set")
	}
	value, exists := sv.secrets[key]
	if !exists {
		return "", fmt.Errorf("secret not found: %s", key)
	}
	return value, nil
}

// Delete removes a secret and persists to disk.
func (sv *SecretVault) Delete(key string) error {
	delete(sv.secrets, key)
	return sv.Save()
}

// List returns all secret keys (without values).
func (sv *SecretVault) List() []string {
	keys := make([]string, 0, len(sv.secrets))
	for k := range sv.secrets {
		keys = append(keys, k)
	}
	return keys
}

// encryptBlob serialises the secrets map to JSON and encrypts it with AES-256-GCM.
// Returns base64-encoded ciphertext (nonce || ciphertext).
func (sv *SecretVault) encryptBlob() (string, error) {
	plaintext, err := json.Marshal(sv.secrets)
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

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptBlob decrypts a base64-encoded ciphertext and unmarshals the
// resulting JSON into a secrets map.
func (sv *SecretVault) decryptBlob(encoded string) (map[string]string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}

	block, err := aes.NewCipher(sv.masterKey)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, encrypted := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesgcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	var secrets map[string]string
	if err := json.Unmarshal(plaintext, &secrets); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return secrets, nil
}

// decryptLegacyValue decrypts a single base64-encoded AES-256-GCM ciphertext
// using the given key — the per-value encryption scheme used before the
// vaultFile refactor (salt + encrypted blob).
func decryptLegacyValue(masterKey []byte, encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(masterKey)
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
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plaintext), nil
}

// Save persists the entire secrets map as a single encrypted blob.
// The on-disk format reveals only the salt and the encrypted payload — secret
// names are never stored in plaintext.
func (sv *SecretVault) Save() error {
	encrypted, err := sv.encryptBlob()
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	wrapper := vaultFile{
		Salt: base64.StdEncoding.EncodeToString(sv.salt),
		Data: encrypted,
	}

	out, err := json.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	return os.WriteFile(sv.path, out, 0o600)
}

// Load reads the vault file from disk.
//
// It first tries the new vaultFile format (salt + encrypted blob).  If that
// fails it falls back to the legacy per-value encryption format and schedules
// an automatic migration on the next write.
//
// Decryption is deferred until SetMasterPassword is called.
func (sv *SecretVault) Load() error {
	data, err := os.ReadFile(sv.path)
	if err != nil {
		return err
	}

	// Try new vaultFile format first.
	var wrapper vaultFile
	if err := json.Unmarshal(data, &wrapper); err == nil && wrapper.Data != "" {
		sv.salt, err = base64.StdEncoding.DecodeString(wrapper.Salt)
		if err != nil {
			return fmt.Errorf("decode salt: %w", err)
		}
		sv.encryptedData = wrapper.Data
		return nil
	}

	// Fall back to legacy format: map[string]string of individually encrypted values.
	var legacy map[string]string
	if err := json.Unmarshal(data, &legacy); err != nil {
		return fmt.Errorf("unrecognized vault format: %w", err)
	}
	sv.legacyEncrypted = legacy
	return nil
}
