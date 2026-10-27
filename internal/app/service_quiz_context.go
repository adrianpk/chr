package app

import (
	"os"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

func appendQuizFileContextSection(sections []quizmodel.ContextSection, projectDir, name, layer, path string) []quizmodel.ContextSection {
	if !filepath.IsAbs(path) {
		path = filepath.Join(projectDir, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return sections
	}

	return append(sections, quizmodel.ContextSection{
		Name:      name,
		Layer:     layer,
		Authority: filepath.ToSlash(path),
		Content:   string(content),
	})
}

func appendQuizNodeContextSection(sections []quizmodel.ContextSection, projectDir string, node core.WorkspaceNode) []quizmodel.ContextSection {
	content, err := os.ReadFile(filepath.Join(projectDir, filepath.FromSlash(node.RelativePath)))
	if err != nil {
		return sections
	}

	name := "project_dump"
	if node.CaseSlug != "" {
		name = "use_case_dump"
	}

	return append(sections, quizmodel.ContextSection{
		Name:      name,
		Layer:     "dump",
		Authority: node.RelativePath,
		Content:   string(content),
	})
}
