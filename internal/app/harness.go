package app

import (
	"os"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/harness"
	"forge.adrianpk.com/chr/internal/harness/codex"
)

func legacyCodexMounts() []harness.RuntimeMount {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	return codex.CredentialMounts(homeDir)
}

func newCodexWorker() codex.Runner {
	return codex.Runner{Mounts: legacyCodexMounts()}
}

func validateCodexRuntime(opts core.RuntimeValidationOptions) (core.RuntimeValidationResult, error) {
	opts.ExtraMounts = legacyCodexMounts()

	return core.ValidateRuntime(opts)
}
