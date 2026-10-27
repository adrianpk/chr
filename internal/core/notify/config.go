package notify

import "fmt"

// ValidateNotificationConfig validates the corresponding package boundary value.
func ValidateNotificationConfig(config NotificationConfig) error {
	if !config.Telegram.Enabled {
		return nil
	}

	if config.Telegram.BotTokenEnv == "" {
		return fmt.Errorf("telegram bot token environment variable is required when telegram notifications are enabled")
	}

	if config.Telegram.ChatID == "" {
		return fmt.Errorf("telegram chat id is required when telegram notifications are enabled")
	}

	return nil
}

// NotificationConfig describes a package boundary value.
type NotificationConfig struct {
	Telegram TelegramNotificationConfig
}

// TelegramNotificationConfig describes a package boundary value.
type TelegramNotificationConfig struct {
	Enabled     bool
	BotTokenEnv string
	ChatID      string
}

// DefaultNotificationConfig executes the corresponding package operation.
func DefaultNotificationConfig() NotificationConfig {
	return NotificationConfig{}
}
