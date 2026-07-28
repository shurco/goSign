package server

import (
	"context"
	"strconv"

	"github.com/shurco/gosign/internal/queries"
	"github.com/shurco/gosign/pkg/notification"
	"github.com/shurco/gosign/pkg/utils"
)

// initNotificationService registers email (SMTP) and SMS (Twilio) providers
// based on global settings stored in the database.
func initNotificationService(settingQueries *queries.SettingQueries) *notification.Service {
	svc := notification.NewService()
	ctx := context.Background()

	if smtpMap, err := settingQueries.GetGlobalSetting(ctx, "smtp"); err == nil && utils.GetStringFromMap(smtpMap, "provider", "") == "smtp" {
		port, _ := strconv.Atoi(utils.GetStringFromMap(smtpMap, "smtp_port", "1025"))
		if port == 0 {
			port = 1025
		}
		smtpCfg := notification.SMTPConfig{
			Host:      utils.GetStringFromMap(smtpMap, "smtp_host", ""),
			Port:      port,
			User:      utils.GetStringFromMap(smtpMap, "smtp_user", ""),
			Password:  utils.GetStringFromMap(smtpMap, "smtp_pass", ""),
			FromEmail: utils.GetStringFromMap(smtpMap, "from_email", ""),
			FromName:  utils.GetStringFromMap(smtpMap, "from_name", ""),
		}
		if smtpCfg.Host != "" && smtpCfg.FromEmail != "" {
			svc.RegisterProvider(notification.NewEmailProvider(smtpCfg))
		}
	}

	if smsMap, err := settingQueries.GetGlobalSetting(ctx, "sms"); err == nil && utils.GetStringFromMap(smsMap, "twilio_enabled", "") == "true" {
		svc.RegisterProvider(notification.NewSMSProvider(notification.TwilioConfig{
			AccountSID: utils.GetStringFromMap(smsMap, "twilio_account_sid", ""),
			AuthToken:  utils.GetStringFromMap(smsMap, "twilio_auth_token", ""),
			FromNumber: utils.GetStringFromMap(smsMap, "twilio_from_number", ""),
			Enabled:    true,
		}))
	}

	return svc
}
