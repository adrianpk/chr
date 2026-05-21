package chc

import "fmt"

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
