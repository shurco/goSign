package notification

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/shurco/gosign/internal/models"
	mail "gopkg.in/mail.v2"
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
	m := mail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", p.config.FromName, p.config.FromEmail))
	m.SetHeader("To", notification.Recipient)
	m.SetHeader("Subject", notification.Subject)

	// If HTML body exists
	if notification.Context["html"] != nil {
		htmlBody, ok := notification.Context["html"].(string)
		if ok && htmlBody != "" {
			m.SetBody("text/html", htmlBody)
			// Alternative text format
			if notification.Body != "" {
				m.AddAlternative("text/plain", notification.Body)
			}
		} else {
			m.SetBody("text/plain", notification.Body)
		}
	} else {
		m.SetBody("text/plain", notification.Body)
	}

	d := mail.NewDialer(p.config.Host, p.config.Port, p.config.User, p.config.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: false}

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// Type returns provider type
func (p *EmailProvider) Type() models.NotificationType {
	return models.NotificationTypeEmail
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

	// TODO: Implement actual Twilio API call when credentials are provided
	// For now: log and return success message indicating SMS would be sent
	// This is a correct stub that can be easily extended later
	
	// When implementing:
	// 1. Use Twilio REST API: https://api.twilio.com/2010-04-01/Accounts/{AccountSID}/Messages.json
	// 2. POST with Body, From, To parameters
	// 3. Use Basic Auth with AccountSID and AuthToken
	// 4. Handle Twilio error responses
	
	return fmt.Errorf("SMS sending not fully implemented - would send to %s via Twilio (AccountSID: %s)", 
		notification.Recipient, p.config.AccountSID[:8]+"...")
}

// Type returns provider type
func (p *SMSProvider) Type() models.NotificationType {
	return models.NotificationTypeSMS
}

