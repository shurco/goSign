package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shurco/gosign/internal/models"
)

// Dispatcher sends webhook events
type Dispatcher struct {
	client     *http.Client
	maxRetries int
	timeout    time.Duration
}

// NewDispatcher creates new dispatcher
func NewDispatcher(maxRetries int, timeout time.Duration) *Dispatcher {
	return &Dispatcher{
		client: &http.Client{
			Timeout: timeout,
		},
		maxRetries: maxRetries,
		timeout:    timeout,
	}
}

// Send sends webhook event
func (d *Dispatcher) Send(ctx context.Context, webhook *models.Webhook, event *models.WebhookEvent) error {
	// Check that webhook is enabled
	if !webhook.Enabled {
		return fmt.Errorf("webhook is disabled")
	}

	// Check that webhook is subscribed to this event
	if !d.isSubscribed(webhook, event.Type) {
		return nil // not an error, just don't send
	}

	// Serialize event
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Generate signature
	signature := d.generateSignature(payload, webhook.Secret)

	// Send with retry
	var lastErr error
	for attempt := 0; attempt <= d.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(attempt*attempt) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		if err := d.sendRequest(ctx, webhook.URL, payload, signature); err != nil {
			lastErr = err
			log.Warn().
				Err(err).
				Str("webhook_id", webhook.ID).
				Str("url", webhook.URL).
				Int("attempt", attempt+1).
				Msg("Webhook delivery failed, retrying...")
			continue
		}

		// Success
		log.Info().
			Str("webhook_id", webhook.ID).
			Str("url", webhook.URL).
			Str("event_type", event.Type).
			Msg("Webhook delivered successfully")
		return nil
	}

	return fmt.Errorf("failed after %d retries: %w", d.maxRetries, lastErr)
}

// sendRequest executes HTTP request
func (d *Dispatcher) sendRequest(ctx context.Context, url string, payload []byte, signature string) error {
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("User-Agent", "goSign-Webhook/1.0")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response (for logging)
	body, _ := io.ReadAll(resp.Body)

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// generateSignature generates HMAC-SHA256 signature
func (d *Dispatcher) generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// isSubscribed checks if webhook is subscribed to event
func (d *Dispatcher) isSubscribed(webhook *models.Webhook, eventType string) bool {
	for _, subscribedEvent := range webhook.Events {
		if subscribedEvent == eventType || subscribedEvent == "*" {
			return true
		}
	}
	return false
}

// VerifySignature verifies webhook signature (for incoming webhooks)
func VerifySignature(payload []byte, signature, secret string) bool {
	expectedSignature := generateSignature(payload, secret)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// generateSignature helper function for signature generation
func generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

