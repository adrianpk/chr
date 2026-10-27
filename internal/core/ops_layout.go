package core

import (
	"fmt"
	"path/filepath"
	"strings"

	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
)

const (
	// OpsDirectoryName is the project-local root for shared operational artifacts.
	OpsDirectoryName = "ops"
)

var opsArtifactTypes = map[string]struct{}{
	"dump": {}, "quiz": {}, "spec": {}, "plan": {}, "tracker": {},
	"ticket": {}, "run": {}, "report": {}, "explanation": {}, "map": {},
}

// OpsLayout describes the project-local operational layout.
type OpsLayout struct {
	projectDir string
	namespace  string
}

// ResolveOpsLayoutOptions configures shared operational layout resolution.
type ResolveOpsLayoutOptions struct {
	ProjectDir string
	Namespace  string
	LookupEnv  func(string) string
}

// ResolveOpsLayout resolves an explicit namespace, OPS_USER, or default.
func ResolveOpsLayout(opts ResolveOpsLayoutOptions) (OpsLayout, error) {
	projectDir, err := projectcore.ResolveDirectory(opts.ProjectDir)
	if err != nil {
		return OpsLayout{}, err
	}

	namespace, err := opsnamespace.Resolve(opts.Namespace, opts.LookupEnv)
	if err != nil {
		return OpsLayout{}, err
	}

	return OpsLayout{projectDir: projectDir, namespace: namespace}, nil
}

// ProjectDir returns the absolute project directory.
func (l OpsLayout) ProjectDir() string { return l.projectDir }

// Namespace returns the active shared namespace.
func (l OpsLayout) Namespace() string { return l.namespace }

// Root returns the absolute namespace root.
func (l OpsLayout) Root() string {
	return filepath.Join(l.projectDir, OpsDirectoryName, l.namespace)
}

// RelativeRoot returns the slash-separated namespace root relative to projectDir.
func (l OpsLayout) RelativeRoot() string {
	return filepath.ToSlash(filepath.Join(OpsDirectoryName, l.namespace))
}

// ArtifactRoot returns the absolute directory for one canonical artifact type.
func (l OpsLayout) ArtifactRoot(artifactType string) (string, error) {
	if _, ok := opsArtifactTypes[artifactType]; !ok {
		return "", fmt.Errorf("invalid ops artifact type: %s", artifactType)
	}

	return filepath.Join(l.Root(), artifactType), nil
}

// ArtifactDirectory returns the absolute directory for one canonical artifact.
func (l OpsLayout) ArtifactDirectory(artifactType, id string) (string, error) {
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
func (l OpsLayout) WorktreeRoot() string {
	return filepath.Join(l.Root(), "worktree")
}

// WorktreeDirectory returns the absolute code worktree path for one artifact ID.
func (l OpsLayout) WorktreeDirectory(id string) (string, error) {
	if strings.TrimSpace(id) == "" || filepath.Base(id) != id {
		return "", fmt.Errorf("invalid ops worktree id: %s", id)
	}

	return filepath.Join(l.WorktreeRoot(), id), nil
}
