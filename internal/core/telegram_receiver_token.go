package core

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TelegramReceiverToken reads an environment reference, then the private file.
// It never includes the token or file content in an error.
func TelegramReceiverToken(env UserEnvironment, config TelegramReceiverConfig) (string, error) {
	if token := os.Getenv(config.BotTokenEnv); token != "" {
		return validateTelegramReceiverToken(token)
	}

	data, err := readTelegramPrivateFile(filepath.Join(env.ConfigDir(), "telegram.env"), 4096)
	if err != nil {
		return "", errors.New("telegram token environment reference is unavailable")
	}

	name, token, ok := strings.Cut(strings.TrimSuffix(string(data), "\n"), "=")
	if !ok || name != config.BotTokenEnv {
		return "", errors.New("telegram private token environment reference does not match")
	}

	return validateTelegramReceiverToken(token)
}

func validateTelegramReceiverToken(token string) (string, error) {
	if !regexp.MustCompile(`^[0-9]{1,20}:[A-Za-z0-9_-]{20,256}$`).MatchString(token) {
		return "", errors.New("telegram token is malformed")
	}

	return token, nil
}

// SaveTelegramReceiverToken stores a masked-input token without replacing another credential implicitly.
func SaveTelegramReceiverToken(env UserEnvironment, config TelegramReceiverConfig, token string, replace bool) error {
	if _, err := validateTelegramReceiverToken(token); err != nil {
		return err
	}

	path := filepath.Join(env.ConfigDir(), "telegram.env")
	wanted := config.BotTokenEnv + "=" + token + "\n"

	prior, err := readTelegramPrivateFile(path, 4096)
	if err == nil && string(prior) == wanted {
		return nil
	}

	if !errors.Is(err, os.ErrNotExist) && !replace {
		return errors.New("telegram credential file exists; use --replace to change it")
	}

	return writeTelegramPrivateFile(path, []byte(wanted))
}
