package notification

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/shurco/gosign/internal/models"
	"github.com/wneessen/go-mail"
)

// SMTPConfig contains SMTP settings
type SMTPConfig struct {
	Host      string
	Port      int
	User      string
	Password  string
	FromEmail string
	FromName  string
}

// EmailProvider sends email notifications
type EmailProvider struct {
	config SMTPConfig
}

// NewEmailProvider creates new email provider
func NewEmailProvider(config SMTPConfig) *EmailProvider {
	return &EmailProvider{
		config: config,
	}
}

// Send sends email
func (p *EmailProvider) Send(ctx context.Context, notification *models.Notification) error {
	m := mail.NewMsg()
	if err := m.FromFormat(p.config.FromName, p.config.FromEmail); err != nil {
		return fmt.Errorf("failed to set from address: %w", err)
	}
	if err := m.To(notification.Recipient); err != nil {
		return fmt.Errorf("failed to set to address: %w", err)
	}
	m.Subject(notification.Subject)

	// If HTML body exists
	if notification.Context["html"] != nil {
		htmlBody, ok := notification.Context["html"].(string)
		if ok && htmlBody != "" {
			m.SetBodyString(mail.TypeTextHTML, htmlBody)
			// Alternative text format
			if notification.Body != "" {
				m.AddAlternativeString(mail.TypeTextPlain, notification.Body)
			}
		} else {
			m.SetBodyString(mail.TypeTextPlain, notification.Body)
		}
	} else {
		m.SetBodyString(mail.TypeTextPlain, notification.Body)
	}

	c, err := mail.NewClient(
		p.config.Host,
		mail.WithPort(p.config.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(p.config.User),
		mail.WithPassword(p.config.Password),
		mail.WithTLSPolicy(mail.TLSMandatory),
		mail.WithTLSConfig(&tls.Config{InsecureSkipVerify: false}),
	)
	if err != nil {
		return fmt.Errorf("failed to create mail client: %w", err)
	}

	if err := c.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// Type returns provider type
func (p *EmailProvider) Type() models.NotificationType {
	return models.NotificationTypeEmail
}

// Enabled reports whether this provider is usable (configured).
func (p *EmailProvider) Enabled() bool {
	return p.config.Host != "" && p.config.Port > 0 && p.config.FromEmail != ""
}

// TwilioConfig contains Twilio API configuration
type TwilioConfig struct {
	AccountSID string
	AuthToken  string
	FromNumber string
	Enabled    bool
}

// SMSProvider sends SMS notifications via Twilio
type SMSProvider struct {
	config TwilioConfig
}

// NewSMSProvider creates new SMS provider with Twilio configuration
func NewSMSProvider(config TwilioConfig) *SMSProvider {
	return &SMSProvider{
		config: config,
	}
}

// Send sends SMS via Twilio
func (p *SMSProvider) Send(ctx context.Context, notification *models.Notification) error {
	// Check if Twilio is configured and enabled
	if !p.config.Enabled || p.config.AccountSID == "" || p.config.AuthToken == "" {
		return fmt.Errorf("Twilio SMS is not configured or disabled")
	}

	// Validate phone number
	if notification.Recipient == "" {
		return fmt.Errorf("recipient phone number is required")
	}

	if p.config.FromNumber == "" {
		return fmt.Errorf("Twilio from number is required")
	}
	if strings.TrimSpace(notification.Body) == "" {
		return fmt.Errorf("SMS body is required")
	}

	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", p.config.AccountSID)
	form := url.Values{}
	form.Set("From", p.config.FromNumber)
	form.Set("To", notification.Recipient)
	form.Set("Body", notification.Body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create Twilio request: %w", err)
	}
	req.SetBasicAuth(p.config.AccountSID, p.config.AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send SMS via Twilio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("twilio returned %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	return nil
}

// Type returns provider type
func (p *SMSProvider) Type() models.NotificationType {
	return models.NotificationTypeSMS
}

// Enabled reports whether this provider is usable (configured + enabled).
func (p *SMSProvider) Enabled() bool {
	return p.config.Enabled && p.config.AccountSID != "" && p.config.AuthToken != "" && p.config.FromNumber != ""
}

