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

// SMSProvider sends SMS notifications (stub for future implementation)
type SMSProvider struct {
	// TODO: add Twilio configuration when needed
}

// NewSMSProvider creates new SMS provider
func NewSMSProvider() *SMSProvider {
	return &SMSProvider{}
}

// Send sends SMS
func (p *SMSProvider) Send(ctx context.Context, notification *models.Notification) error {
	// TODO: implement sending via Twilio
	return fmt.Errorf("SMS provider not implemented yet")
}

// Type returns provider type
func (p *SMSProvider) Type() models.NotificationType {
	return models.NotificationTypeSMS
}

