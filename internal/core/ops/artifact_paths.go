package ops

import (
	"path/filepath"
	"strings"

	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
)

const opsArtifactBodyFile = "artifact.md"

// OpsArtifactBodyFile returns the canonical Markdown body filename for shared artifacts.
func OpsArtifactBodyFile() string { return opsArtifactBodyFile }

// DefaultOpsArtifactRoot returns the default namespace root for one artifact type.
func DefaultOpsArtifactRoot(projectDir, artifactType string) string {
	return filepath.Join(projectDir, OpsDirectoryName, opsnamespace.Default, artifactType)
}

// IsOpsArtifactBodyPath reports whether path is an artifact.md body inside ops/.
func IsOpsArtifactBodyPath(projectDir, path string) bool {
	relativePath, err := filepath.Rel(projectDir, path)
	if err != nil {
		return false
	}

	parts := strings.Split(filepath.ToSlash(relativePath), "/")
	if len(parts) != 5 {
		return false
	}

	if parts[0] != OpsDirectoryName || parts[4] != opsArtifactBodyFile {
		return false
	}

	if !opsnamespace.Valid(parts[1]) {
		return false
	}

	_, ok := opsArtifactTypes[parts[2]]

	return ok
}

// OpsArtifactTypeForBodyPath returns the artifact type for a canonical artifact body path.
func OpsArtifactTypeForBodyPath(projectDir, path string) (string, bool) {
	if !IsOpsArtifactBodyPath(projectDir, path) {
		return "", false
	}

	relativePath, err := filepath.Rel(projectDir, path)
	if err != nil {
		return "", false
	}

	parts := strings.Split(filepath.ToSlash(relativePath), "/")

	return parts[2], true
}
