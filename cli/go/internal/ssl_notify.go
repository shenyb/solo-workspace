package core

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const sslNotifyFile = "ssl-notify.yaml"

type sslNotifyState struct {
	LastSent map[string]string `yaml:"last_sent"`
}

func sslNotifyPath() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, sslNotifyFile), nil
}

func loadSSLNotifyState() (*sslNotifyState, error) {
	path, err := sslNotifyPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &sslNotifyState{LastSent: make(map[string]string)}, nil
		}
		return nil, err
	}
	state := &sslNotifyState{LastSent: make(map[string]string)}
	if err := yaml.Unmarshal(data, state); err != nil {
		return nil, err
	}
	if state.LastSent == nil {
		state.LastSent = make(map[string]string)
	}
	return state, nil
}

func saveSSLNotifyState(state *sslNotifyState) error {
	path, err := sslNotifyPath()
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(state)
	if err != nil {
		return err
	}
	return WriteFileAtomic(path, data, 0600)
}

// ShouldNotifySSL reports whether an expiry alert should be sent for domain today.
func ShouldNotifySSL(domain string) (bool, error) {
	state, err := loadSSLNotifyState()
	if err != nil {
		return true, err
	}
	today := time.Now().Format("2006-01-02")
	return state.LastSent[domain] != today, nil
}

// MarkSSLNotified records that an alert was sent for domain today.
func MarkSSLNotified(domain string) error {
	state, err := loadSSLNotifyState()
	if err != nil {
		return err
	}
	if state.LastSent == nil {
		state.LastSent = make(map[string]string)
	}
	state.LastSent[domain] = time.Now().Format("2006-01-02")
	return saveSSLNotifyState(state)
}

// FilterSSLNotifyDomains returns domains that have not been notified today.
func FilterSSLNotifyDomains(domains []string) ([]string, error) {
	state, err := loadSSLNotifyState()
	if err != nil {
		return nil, fmt.Errorf("load ssl notify state: %w", err)
	}
	today := time.Now().Format("2006-01-02")
	var pending []string
	for _, d := range domains {
		if state.LastSent[d] != today {
			pending = append(pending, d)
		}
	}
	return pending, nil
}

// MarkSSLNotifiedBatch marks multiple domains as notified today.
func MarkSSLNotifiedBatch(domains []string) error {
	state, err := loadSSLNotifyState()
	if err != nil {
		return err
	}
	if state.LastSent == nil {
		state.LastSent = make(map[string]string)
	}
	today := time.Now().Format("2006-01-02")
	for _, d := range domains {
		state.LastSent[d] = today
	}
	return saveSSLNotifyState(state)
}
