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

	data, err := readTelegramPrivateFile(filepath.Join(env.ConfigDir(), "telegram.env"), 64*1024)
	if err != nil {
		return "", errors.New("telegram token environment reference is unavailable")
	}

	values, err := parseTelegramEnvironment(data)
	if err != nil {
		return "", err
	}

	token, ok := values[config.BotTokenEnv]
	if !ok {
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

	prior, err := readTelegramPrivateFile(path, 64*1024)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return errors.New("telegram credential file is unavailable")
	}

	values, err := parseTelegramEnvironment(prior)
	if err != nil {
		return err
	}

	if old, exists := values[config.BotTokenEnv]; exists {
		if old == token {
			return nil
		}

		if !replace {
			return errors.New("telegram credential file exists; use --replace to change it")
		}
	}

	wanted, err := renderTelegramTokenEnvironment(prior, config.BotTokenEnv, token)
	if err != nil {
		return err
	}

	return writeTelegramPrivateFile(path, wanted)
}

func renderTelegramTokenEnvironment(prior []byte, name, token string) ([]byte, error) {
	var lines []string

	for _, line := range strings.Split(strings.TrimSuffix(string(prior), "\n"), "\n") {
		key, _, found := strings.Cut(line, "=")
		if line != "" && (!found || key != name) {
			lines = append(lines, line)
		}
	}

	lines = append(lines, name+"="+token)

	wanted := strings.Join(lines, "\n") + "\n"
	if len(wanted) > 64*1024 {
		return nil, errors.New("telegram environment file exceeds its size limit")
	}

	return []byte(wanted), nil
}

func parseTelegramEnvironment(data []byte) (map[string]string, error) {
	values := map[string]string{}

	for _, line := range strings.Split(string(data), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		name, value, ok := strings.Cut(line, "=")
		if !ok || !regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,127}$`).MatchString(name) || strings.ContainsAny(value, "\x00\r") {
			return nil, errors.New("telegram environment file contains an invalid assignment")
		}

		if _, duplicate := values[name]; duplicate {
			return nil, errors.New("telegram environment file contains a duplicate assignment")
		}

		values[name] = value
	}

	return values, nil
}
