package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SendWebhook posts a JSON alert to the configured webhook URL.
func SendWebhook(subject, body string) error {
	if CurrentConfig == nil || CurrentConfig.Notify == nil || CurrentConfig.Notify.Webhook == "" {
		return fmt.Errorf("webhook is not configured")
	}

	payload, err := json.Marshal(map[string]string{
		"subject": subject,
		"body":    body,
	})
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(CurrentConfig.Notify.Webhook, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %s", resp.Status)
	}
	return nil
}
