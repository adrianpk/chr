package chc

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type WorkspaceConfig struct {
	TicketKeyPrefix string
}

func LoadWorkspaceConfig(projectDir string) (WorkspaceConfig, error) {
	config := WorkspaceConfig{TicketKeyPrefix: DefaultTicketKeyPrefix}
	path := filepath.Join(projectDir, ".chc", "config.yml")

	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return config, nil
	}

	if err != nil {
		return WorkspaceConfig{}, fmt.Errorf("open workspace config: %w", err)
	}

	defer func() {
		//nolint:errcheck // The config file is read-only here; scanner errors are checked below.
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		keyPrefix, ok := strings.CutPrefix(line, "key_prefix:")
		if !ok {
			continue
		}

		keyPrefix = strings.Trim(strings.TrimSpace(keyPrefix), `"'`)
		if keyPrefix != "" {
			config.TicketKeyPrefix = keyPrefix
		}
	}

	if err := scanner.Err(); err != nil {
		return WorkspaceConfig{}, fmt.Errorf("read workspace config: %w", err)
	}

	return config, nil
}
