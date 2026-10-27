package ops

import (
	"path/filepath"
	"strings"
	"time"
)

// NamedArtifactPath binds absolute and project-relative paths to the same body.
type NamedArtifactPath struct {
	RelativePath string
	Path         string
}

// NamedArtifactBody resolves a deterministic Markdown artifact body path.
func NamedArtifactBody(projectDir, artifactType, name string) (NamedArtifactPath, error) {
	layout, err := ResolveLayout(LayoutOptions{ProjectDir: projectDir})
	if err != nil {
		return NamedArtifactPath{}, err
	}

	id := "00000-" + ArtifactSlugOrDefault(name)

	artifactDir, err := layout.ArtifactDirectory(artifactType, id)
	if err != nil {
		return NamedArtifactPath{}, err
	}

	bodyPath := filepath.Join(artifactDir, ArtifactBodyFile())

	relativePath, err := filepath.Rel(projectDir, bodyPath)
	if err != nil {
		return NamedArtifactPath{}, err
	}

	return NamedArtifactPath{RelativePath: filepath.ToSlash(relativePath), Path: bodyPath}, nil
}

// EnsureArtifactBodyMetadata ensures metadata for a canonical ops artifact body.
func EnsureArtifactBodyMetadata(projectDir, relativePath string) error {
	parts := strings.Split(filepath.ToSlash(relativePath), "/")
	if len(parts) != 5 || parts[0] != DirectoryName || parts[4] != ArtifactBodyFile() {
		return nil
	}

	layout, err := ResolveLayout(LayoutOptions{ProjectDir: projectDir, Namespace: parts[1]})
	if err != nil {
		return err
	}

	id := parts[3]
	name := strings.TrimPrefix(id, "00000-")

	artifact, err := EnsureArtifactEnvelope(ArtifactOptions{
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

	return RefreshArtifactMetadata(artifact.Dir, "created", time.Now().UTC())
}

// EnsureNamedArtifactMetadata ensures metadata for a deterministic Markdown artifact body.
func EnsureNamedArtifactMetadata(projectDir, artifactType, name string) error {
	layout, err := ResolveLayout(LayoutOptions{ProjectDir: projectDir})
	if err != nil {
		return err
	}

	id := "00000-" + ArtifactSlugOrDefault(name)

	artifact, err := EnsureArtifactEnvelope(ArtifactOptions{
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

	return RefreshArtifactMetadata(artifact.Dir, "created", time.Now().UTC())
}
