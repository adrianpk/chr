package ops

import (
	"path/filepath"
	"strings"

	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
)

const artifactBodyFile = "artifact.md"

// ArtifactBodyFile returns the canonical Markdown body filename for shared artifacts.
func ArtifactBodyFile() string { return artifactBodyFile }

// DefaultArtifactRoot returns the default namespace root for one artifact type.
func DefaultArtifactRoot(projectDir, artifactType string) string {
	return filepath.Join(projectDir, DirectoryName, opsnamespace.Default, artifactType)
}

// IsArtifactBodyPath reports whether path is an artifact.md body inside ops/.
func IsArtifactBodyPath(projectDir, path string) bool {
	relativePath, err := filepath.Rel(projectDir, path)
	if err != nil {
		return false
	}

	parts := strings.Split(filepath.ToSlash(relativePath), "/")
	if len(parts) != 5 {
		return false
	}

	if parts[0] != DirectoryName || parts[4] != artifactBodyFile {
		return false
	}

	if !opsnamespace.Valid(parts[1]) {
		return false
	}

	_, ok := artifactTypes[parts[2]]

	return ok
}

// ArtifactTypeForBodyPath returns the artifact type for a canonical artifact body path.
func ArtifactTypeForBodyPath(projectDir, path string) (string, bool) {
	if !IsArtifactBodyPath(projectDir, path) {
		return "", false
	}

	relativePath, err := filepath.Rel(projectDir, path)
	if err != nil {
		return "", false
	}

	parts := strings.Split(filepath.ToSlash(relativePath), "/")

	return parts[2], true
}
