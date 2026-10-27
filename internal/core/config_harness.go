package core

import (
	"errors"

	"forge.adrianpk.com/chr/internal/core/configvalue"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
)

func applyHarnessSelection(config *WorkspaceConfig, values map[string]string) error {
	for _, field := range []struct {
		key    string
		target *string
	}{
		{"harnesses.worker", &config.Harnesses.Worker},
		{"harnesses.generation", &config.Harnesses.Generation},
	} {
		value, found := values[field.key]
		if !found || value == "null" {
			continue
		}

		value = configvalue.String(value)

		if !harnesscore.ValidProfileName(value) {
			return errors.New("harness selection requires a named profile or null")
		}

		*field.target = value
	}

	return nil
}
