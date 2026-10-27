// Package notify owns notification state, persistence and retained delivery.
package notify

import "fmt"

// ValidateConfig validates the corresponding package boundary value.
func ValidateConfig(config Config) error {
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

// Config describes a package boundary value.
type Config struct {
	Telegram TelegramConfig
}

// TelegramConfig describes a package boundary value.
type TelegramConfig struct {
	Enabled     bool
	BotTokenEnv string
	ChatID      string
}

// DefaultConfig executes the corresponding package operation.
func DefaultConfig() Config {
	return Config{}
}
