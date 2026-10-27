package ops

import (
	"path/filepath"
	"strings"
	"time"
)

// NamedOpsArtifactPath binds absolute and project-relative paths to the same body.
type NamedOpsArtifactPath struct {
	RelativePath string
	Path         string
}

// NamedOpsArtifactBody resolves a deterministic Markdown artifact body path.
func NamedOpsArtifactBody(projectDir, artifactType, name string) (NamedOpsArtifactPath, error) {
	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: projectDir})
	if err != nil {
		return NamedOpsArtifactPath{}, err
	}

	id := "00000-" + OpsArtifactSlugOrDefault(name)

	artifactDir, err := layout.ArtifactDirectory(artifactType, id)
	if err != nil {
		return NamedOpsArtifactPath{}, err
	}

	bodyPath := filepath.Join(artifactDir, OpsArtifactBodyFile())

	relativePath, err := filepath.Rel(projectDir, bodyPath)
	if err != nil {
		return NamedOpsArtifactPath{}, err
	}

	return NamedOpsArtifactPath{RelativePath: filepath.ToSlash(relativePath), Path: bodyPath}, nil
}

// EnsureOpsArtifactBodyMetadata ensures metadata for a canonical ops artifact body.
func EnsureOpsArtifactBodyMetadata(projectDir, relativePath string) error {
	parts := strings.Split(filepath.ToSlash(relativePath), "/")
	if len(parts) != 5 || parts[0] != OpsDirectoryName || parts[4] != OpsArtifactBodyFile() {
		return nil
	}

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: projectDir, Namespace: parts[1]})
	if err != nil {
		return err
	}

	id := parts[3]
	name := strings.TrimPrefix(id, "00000-")

	artifact, err := EnsureOpsArtifactEnvelope(CreateOpsArtifactOptions{
		Layout:      layout,
		Type:        parts[2],
		Title:       name,
		Producer:    "chr",
		ExternalIDs: []string{name},
		Now:         time.Now,
	}, id)
	if err != nil {
		return err
	}

	return RefreshOpsArtifactMetadata(artifact.Dir, "created", time.Now().UTC())
}

// EnsureNamedOpsArtifactMetadata ensures metadata for a deterministic Markdown artifact body.
func EnsureNamedOpsArtifactMetadata(projectDir, artifactType, name string) error {
	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: projectDir})
	if err != nil {
		return err
	}

	id := "00000-" + OpsArtifactSlugOrDefault(name)

	artifact, err := EnsureOpsArtifactEnvelope(CreateOpsArtifactOptions{
		Layout:      layout,
		Type:        artifactType,
		Title:       strings.TrimSpace(name),
		Producer:    "chr",
		ExternalIDs: []string{name},
		Now:         time.Now,
	}, id)
	if err != nil {
		return err
	}

	return RefreshOpsArtifactMetadata(artifact.Dir, "created", time.Now().UTC())
}
