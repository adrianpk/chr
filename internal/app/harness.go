package app

import (
	"os"

	harnessruntime "forge.adrianpk.com/chr/internal/core/harness/runtime"
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

func validateCodexRuntime(opts harnessruntime.ValidationOptions) (harnessruntime.ValidationResult, error) {
	opts.ExtraMounts = legacyCodexMounts()

	return harnessruntime.Validate(opts)
}
