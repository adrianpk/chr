package codex

import (
	"os"
	"path/filepath"
	"slices"

	"forge.adrianpk.com/chr/internal/harness"
)

// RuntimeMounts returns the host-resolved mounts retained by this adapter.
func (r Runner) RuntimeMounts() []harness.RuntimeMount {
	return slices.Clone(r.Mounts)
}

// CredentialMounts resolves existing legacy credential files during setup.
func CredentialMounts(homeDir string) []harness.RuntimeMount {
	candidates := []struct {
		source string
		target string
	}{
		{
			source: filepath.Join(homeDir, ".codex", "auth.json"),
			target: "/root/.codex/auth.json",
		},
		{
			source: filepath.Join(homeDir, ".codex", "config.toml"),
			target: "/root/.codex/config.toml",
		},
	}

	mounts := make([]harness.RuntimeMount, 0, len(candidates))
	for _, candidate := range candidates {
		_, err := os.Stat(candidate.source)
		if err != nil {
			continue
		}

		mounts = append(mounts, harness.RuntimeMount{
			Source:      candidate.source,
			Target:      candidate.target,
			ReadOnly:    true,
			Description: "codex credential",
		})
	}

	return mounts
}
