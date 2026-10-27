package ops

import (
	"fmt"
	"path/filepath"
	"strings"

	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
)

const (
	// DirectoryName is the project-local root for shared operational artifacts.
	DirectoryName = "ops"
)

var artifactTypes = map[string]struct{}{
	"dump": {}, "quiz": {}, "spec": {}, "plan": {}, "tracker": {},
	"ticket": {}, "run": {}, "report": {}, "explanation": {}, "map": {},
}

// Layout describes the project-local operational layout.
type Layout struct {
	projectDir string
	namespace  string
}

// LayoutOptions configures shared operational layout resolution.
type LayoutOptions struct {
	ProjectDir string
	Namespace  string
	LookupEnv  func(string) string
}

// ResolveLayout resolves an explicit namespace, OPS_USER, or default.
func ResolveLayout(opts LayoutOptions) (Layout, error) {
	projectDir, err := projectcore.ResolveDirectory(opts.ProjectDir)
	if err != nil {
		return Layout{}, err
	}

	namespace, err := opsnamespace.Resolve(opts.Namespace, opts.LookupEnv)
	if err != nil {
		return Layout{}, err
	}

	return Layout{projectDir: projectDir, namespace: namespace}, nil
}

// ProjectDir returns the absolute project directory.
func (l Layout) ProjectDir() string { return l.projectDir }

// Namespace returns the active shared namespace.
func (l Layout) Namespace() string { return l.namespace }

// Root returns the absolute namespace root.
func (l Layout) Root() string {
	return filepath.Join(l.projectDir, DirectoryName, l.namespace)
}

// RelativeRoot returns the slash-separated namespace root relative to projectDir.
func (l Layout) RelativeRoot() string {
	return filepath.ToSlash(filepath.Join(DirectoryName, l.namespace))
}

// ArtifactRoot returns the absolute directory for one canonical artifact type.
func (l Layout) ArtifactRoot(artifactType string) (string, error) {
	if _, ok := artifactTypes[artifactType]; !ok {
		return "", fmt.Errorf("invalid ops artifact type: %s", artifactType)
	}

	return filepath.Join(l.Root(), artifactType), nil
}

// ArtifactDirectory returns the absolute directory for one canonical artifact.
func (l Layout) ArtifactDirectory(artifactType, id string) (string, error) {
	root, err := l.ArtifactRoot(artifactType)
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(id) == "" || filepath.Base(id) != id {
		return "", fmt.Errorf("invalid ops artifact id: %s", id)
	}

	return filepath.Join(root, id), nil
}

// WorktreeRoot returns the absolute root for code worktrees associated with artifacts.
func (l Layout) WorktreeRoot() string {
	return filepath.Join(l.Root(), "worktree")
}

// WorktreeDirectory returns the absolute code worktree path for one artifact ID.
func (l Layout) WorktreeDirectory(id string) (string, error) {
	if strings.TrimSpace(id) == "" || filepath.Base(id) != id {
		return "", fmt.Errorf("invalid ops worktree id: %s", id)
	}

	return filepath.Join(l.WorktreeRoot(), id), nil
}
