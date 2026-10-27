package core

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// Load assembles the original workspace, session and filesystem context.
func (WorkspaceAuthoringQuizContext) Load(ctx context.Context, projectDir string, session authoringcore.Session) ([]AuthoringQuizContextSection, error) {
	projectDir, err := CleanContextProjectDir(projectDir)
	if err != nil {
		return nil, err
	}

	sections := []AuthoringQuizContextSection{
		{
			Name:      "project",
			Layer:     "project",
			Authority: "project_directory",
			Content:   "project_dir: " + projectDir + "\nproject_name: " + filepath.Base(projectDir),
		},
		{
			Name:      "session",
			Layer:     "authoring",
			Authority: "authoring_session",
			Content:   authoringQuizSessionContext(session),
		},
	}

	sections = append(sections, authoringQuizOptionalFileSection(projectDir, "agent_protocol", "protocols", "AGENTS.md"))
	sections = append(sections, authoringQuizWorkspaceGraphSection(ctx, projectDir))
	sections = append(sections, authoringQuizPathInventorySection(projectDir, "workspace_artifacts", "workspace", workspacecore.DirectoryName))
	sections = append(sections, authoringQuizPathInventorySection(projectDir, "docs", "docs", "docs"))
	sections = append(sections, authoringQuizTreeSection(projectDir))
	sections = append(sections, authoringQuizGoInventorySection(projectDir))

	return sections, nil
}

func authoringQuizSessionContext(session authoringcore.Session) string {
	lines := []string{
		"session_id: " + session.ID,
		"target: " + string(session.Target),
		"topic: " + session.Topic,
		"description: " + session.Description,
	}

	return strings.Join(lines, "\n")
}

func authoringQuizOptionalFileSection(projectDir, name, layer, relPath string) AuthoringQuizContextSection {
	fullPath := filepath.Join(projectDir, filepath.FromSlash(relPath))

	content, err := os.ReadFile(fullPath)
	if err != nil {
		content = []byte("missing: " + relPath)
	}

	return AuthoringQuizContextSection{
		Name:      name,
		Layer:     layer,
		Authority: relPath,
		Content:   string(content),
	}
}

func authoringQuizWorkspaceGraphSection(ctx context.Context, projectDir string) AuthoringQuizContextSection {
	graph, err := NewWorkspaceGraphBuilder(projectDir).Build(ctx)
	if err != nil {
		return AuthoringQuizContextSection{
			Name:      "workspace_graph_context",
			Layer:     "workspace",
			Authority: "WorkspaceGraph",
			Content:   "workspace graph context unavailable: " + err.Error(),
		}
	}

	assembly, err := AssembleWorkspaceContext(NewProjectWideWorkspaceContextRequest(projectDir, graph))
	if err != nil {
		return AuthoringQuizContextSection{
			Name:      "workspace_graph_context",
			Layer:     "workspace",
			Authority: "WorkspaceGraph",
			Content:   "workspace graph context unavailable: " + err.Error(),
		}
	}

	return AuthoringQuizContextSection{
		Name:      "workspace_graph_context",
		Layer:     "workspace",
		Authority: "WorkspaceGraph",
		Content:   RenderWorkspaceContextAssembly(assembly),
	}
}

func authoringQuizPathInventorySection(projectDir, name, layer, relPath string) AuthoringQuizContextSection {
	root := filepath.Join(projectDir, filepath.FromSlash(relPath))
	paths := authoringQuizRelativePaths(projectDir, root, authoringQuizMarkdownPath)

	return AuthoringQuizContextSection{
		Name:      name,
		Layer:     layer,
		Authority: relPath,
		Content:   strings.Join(paths, "\n"),
	}
}

func authoringQuizTreeSection(projectDir string) AuthoringQuizContextSection {
	paths := authoringQuizRelativePaths(projectDir, projectDir, authoringQuizTreePath)

	return AuthoringQuizContextSection{
		Name:      "tree",
		Layer:     "tree",
		Authority: "project_tree",
		Content:   strings.Join(paths, "\n"),
	}
}

func authoringQuizGoInventorySection(projectDir string) AuthoringQuizContextSection {
	paths := authoringQuizRelativePaths(projectDir, projectDir, authoringQuizGoPath)

	return AuthoringQuizContextSection{
		Name:      "go_inventory",
		Layer:     "code",
		Authority: "go_files",
		Content:   strings.Join(paths, "\n"),
	}
}

func authoringQuizRelativePaths(projectDir, root string, include func(string) bool) []string {
	paths := []string{}

	walkErr := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() && authoringQuizSkipDir(entry.Name()) {
			return filepath.SkipDir
		}

		if entry.IsDir() || !include(path) {
			return nil
		}

		relPath, relErr := filepath.Rel(projectDir, path)
		if relErr != nil {
			return relErr
		}

		paths = append(paths, filepath.ToSlash(relPath))
		if len(paths) >= 200 {
			return filepath.SkipAll
		}

		return nil
	})
	if walkErr != nil {
		return []string{}
	}

	sort.Strings(paths)

	return paths
}

func authoringQuizMarkdownPath(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".md")
}

func authoringQuizGoPath(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".go") ||
		strings.EqualFold(filepath.Base(path), "go.mod")
}

func authoringQuizTreePath(path string) bool {
	name := filepath.Base(path)

	return name != "" && !strings.HasPrefix(name, ".")
}

func authoringQuizSkipDir(name string) bool {
	switch name {
	case ".git", "vendor", "node_modules", "runs":
		return true
	default:
		return false
	}
}

// WorkspaceAuthoringQuizContext supplies existing native workspace/index context to authoring.
type WorkspaceAuthoringQuizContext struct{}
