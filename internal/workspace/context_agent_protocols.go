package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const defaultAgentIdentityID = "default-go-software-infrastructure-agent"

// AgentIdentity describes a package boundary value.
type AgentIdentity struct {
	ID                 string
	Title              string
	Role               string
	OperatingStance    []string
	DomainExpectations []string
	QualityBar         []string
	LanguageRules      []string
	ForbiddenBehaviors []string
}

// AgentProtocolContextOptions describes a package boundary value.
type AgentProtocolContextOptions struct {
	ProjectDir string
	TargetPath string
}

// AgentProtocolContext describes a package boundary value.
type AgentProtocolContext struct {
	ProjectDir    string
	TargetPath    string
	AgentIdentity AgentIdentity
	AgentFiles    []ContextDocument
	ChcFiles      []ContextDocument
	Sources       []ContextSourceInventoryEntry
}

// ContextDocument describes a package boundary value.
type ContextDocument struct {
	ID           string
	Path         string
	RelativePath string
	Content      string
}

// DefaultAgentIdentity executes the corresponding package operation.
func DefaultAgentIdentity() AgentIdentity {
	return AgentIdentity{
		ID:    defaultAgentIdentityID,
		Title: "Default Go Software and Infrastructure Agent",
		Role:  "Senior software engineer focused on Go, testing, infrastructure, and repository hygiene.",
		OperatingStance: []string{
			"Read project protocol before editing.",
			"Use existing repository patterns before introducing new abstractions.",
			"Keep changes scoped to the requested ticket or objective.",
			"Validate behavior with the project commands supplied in context.",
		},
		DomainExpectations: []string{
			"Go implementation and idioms.",
			"CLI and runtime execution boundaries.",
			"Containerized agent execution.",
			"Git-aware project hygiene.",
		},
		QualityBar: []string{
			"Prefer small, reviewable changes.",
			"Preserve deterministic behavior where tests depend on output.",
			"Add or update focused tests for changed behavior.",
			"Report blockers explicitly instead of guessing through missing context.",
		},
		LanguageRules: []string{
			"Write code, comments, and workspace documentation in English.",
			"Keep generated artifacts concise and technically specific.",
		},
		ForbiddenBehaviors: []string{
			"Do not rewrite unrelated files.",
			"Do not perform destructive Git operations.",
			"Do not rely on implicit conversation history.",
		},
	}
}

// LoadAgentProtocolContext loads the corresponding package data.
func LoadAgentProtocolContext(opts AgentProtocolContextOptions) (AgentProtocolContext, error) {
	projectDir, targetPath, targetDir, err := normalizeAgentProtocolContextOptions(opts)
	if err != nil {
		return AgentProtocolContext{}, err
	}

	agentFiles, agentSources, err := loadApplicableAgentFiles(projectDir, targetDir)
	if err != nil {
		return AgentProtocolContext{}, err
	}

	chcFiles, chcSources, err := loadChcProtocolFiles(projectDir)
	if err != nil {
		return AgentProtocolContext{}, err
	}

	sources := []ContextSourceInventoryEntry{
		{
			ID:          defaultAgentIdentityID,
			Kind:        ContextSourceKindAgentIdentity,
			Class:       ContextSourceClassInvariant,
			Description: "Default Go software and infrastructure agent identity.",
			Required:    true,
			Included:    true,
		},
	}
	sources = append(sources, agentSources...)
	sources = append(sources, chcSources...)

	return AgentProtocolContext{
		ProjectDir:    projectDir,
		TargetPath:    targetPath,
		AgentIdentity: DefaultAgentIdentity(),
		AgentFiles:    agentFiles,
		ChcFiles:      chcFiles,
		Sources:       sources,
	}, nil
}

// RenderAgentContext renders the corresponding package data.
func RenderAgentContext(identity AgentIdentity) string {
	var output strings.Builder

	output.WriteString("# Agent Context\n\n")
	writeMarkdownField(&output, "Identity", identity.Title)
	writeMarkdownField(&output, "Role", identity.Role)
	writeMarkdownList(&output, "Operating Stance", identity.OperatingStance)
	writeMarkdownList(&output, "Domain Expectations", identity.DomainExpectations)
	writeMarkdownList(&output, "Quality Bar", identity.QualityBar)
	writeMarkdownList(&output, "Language Rules", identity.LanguageRules)
	writeMarkdownList(&output, "Forbidden Behaviors", identity.ForbiddenBehaviors)

	return output.String()
}

// RenderProtocolsContext renders the corresponding package data.
func RenderProtocolsContext(ctx AgentProtocolContext) string {
	var output strings.Builder

	output.WriteString("# Protocol Context\n\n")
	writeMarkdownField(&output, "Project", ctx.ProjectDir)
	writeMarkdownField(&output, "Target", ctx.TargetPath)

	output.WriteString("## Applicable AGENTS.md Files\n\n")
	writeContextDocuments(&output, ctx.AgentFiles, "No applicable AGENTS.md files were found.")

	output.WriteString("## Selected .chc Context\n\n")
	writeContextDocuments(&output, ctx.ChcFiles, "No selected .chc protocol files were found.")

	return output.String()
}

func normalizeAgentProtocolContextOptions(opts AgentProtocolContextOptions) (string, string, string, error) {
	projectDir, err := cleanAbsProjectDir(opts.ProjectDir)
	if err != nil {
		return "", "", "", err
	}

	targetPath, err := cleanAbsTargetPath(projectDir, opts.TargetPath)
	if err != nil {
		return "", "", "", err
	}

	if !pathIsInsideProject(projectDir, targetPath) {
		return "", "", "", fmt.Errorf("target path is outside project: %s", targetPath)
	}

	return projectDir, targetPath, targetDirectory(targetPath), nil
}

func cleanAbsProjectDir(projectDir string) (string, error) {
	projectDir = strings.TrimSpace(projectDir)
	if projectDir == "" {
		return "", fmt.Errorf("agent protocol project dir is required")
	}

	projectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return "", fmt.Errorf("resolve project dir: %w", err)
	}

	return filepath.Clean(projectDir), nil
}

func cleanAbsTargetPath(projectDir, targetPath string) (string, error) {
	targetPath = strings.TrimSpace(targetPath)
	if targetPath == "" {
		targetPath = projectDir
	} else if !filepath.IsAbs(targetPath) {
		targetPath = filepath.Join(projectDir, targetPath)
	}

	targetPath, err := filepath.Abs(targetPath)
	if err != nil {
		return "", fmt.Errorf("resolve target path: %w", err)
	}

	return filepath.Clean(targetPath), nil
}

func pathIsInsideProject(projectDir, targetPath string) bool {
	rel, err := filepath.Rel(projectDir, targetPath)
	if err != nil {
		return false
	}

	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func targetDirectory(targetPath string) string {
	info, err := os.Stat(targetPath)
	if err == nil && info.IsDir() {
		return targetPath
	}

	return filepath.Dir(targetPath)
}

func loadApplicableAgentFiles(projectDir, targetDir string) ([]ContextDocument, []ContextSourceInventoryEntry, error) {
	dirs, err := dirsFromProjectToTarget(projectDir, targetDir)
	if err != nil {
		return nil, nil, err
	}

	documents := make([]ContextDocument, 0, len(dirs))
	sources := make([]ContextSourceInventoryEntry, 0, len(dirs))

	for _, dir := range dirs {
		path := filepath.Join(dir, "AGENTS.md")

		relPath, err := filepath.Rel(projectDir, path)
		if err != nil {
			return nil, nil, fmt.Errorf("resolve AGENTS.md relative path: %w", err)
		}

		relPath = filepath.ToSlash(relPath)
		id := "agents:" + relPath

		content, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				sources = append(sources, ContextSourceInventoryEntry{
					ID:          id,
					Kind:        ContextSourceKindProtocol,
					Class:       ContextSourceClassInvariant,
					Path:        relPath,
					Description: "Applicable AGENTS.md protocol file.",
					Required:    dir == projectDir,
					Included:    false,
					Reason:      "file was not found",
				})

				continue
			}

			return nil, nil, fmt.Errorf("read AGENTS.md: %w", err)
		}

		documents = append(documents, ContextDocument{
			ID:           id,
			Path:         path,
			RelativePath: relPath,
			Content:      string(content),
		})
		sources = append(sources, ContextSourceInventoryEntry{
			ID:          id,
			Kind:        ContextSourceKindProtocol,
			Class:       ContextSourceClassInvariant,
			Path:        relPath,
			Description: "Applicable AGENTS.md protocol file.",
			Required:    dir == projectDir,
			Included:    true,
		})
	}

	return documents, sources, nil
}

func dirsFromProjectToTarget(projectDir, targetDir string) ([]string, error) {
	rel, err := filepath.Rel(projectDir, targetDir)
	if err != nil {
		return nil, fmt.Errorf("resolve target directory relative to project: %w", err)
	}

	dirs := []string{projectDir}
	if rel == "." {
		return dirs, nil
	}

	parts := strings.Split(filepath.ToSlash(rel), "/")
	current := projectDir

	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}

		current = filepath.Join(current, part)
		dirs = append(dirs, current)
	}

	return dirs, nil
}

func loadChcProtocolFiles(projectDir string) ([]ContextDocument, []ContextSourceInventoryEntry, error) {
	candidates, err := chcProtocolCandidates(projectDir)
	if err != nil {
		return nil, nil, err
	}

	return loadProtocolCandidateFiles(projectDir, candidates)
}

func chcProtocolCandidates(projectDir string) ([]string, error) {
	userDir := UserPath(projectDir)
	candidates := stableChcProtocolCandidates(userDir)
	agentsDir := filepath.Join(userDir, "agents")

	if !directoryExists(agentsDir) {
		sort.Strings(candidates)

		return candidates, nil
	}

	err := filepath.WalkDir(agentsDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			return nil
		}

		if strings.EqualFold(filepath.Ext(path), ".md") {
			candidates = append(candidates, path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk .chc user agents: %w", err)
	}

	sort.Strings(candidates)

	return candidates, nil
}

func stableChcProtocolCandidates(userDir string) []string {
	return []string{
		filepath.Join(userDir, "manifest.md"),
		filepath.Join(userDir, "config.yml"),
		filepath.Join(userDir, "runtime", "agent", "Dockerfile"),
	}
}

func directoryExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}

func loadProtocolCandidateFiles(
	projectDir string,
	candidates []string,
) ([]ContextDocument, []ContextSourceInventoryEntry, error) {
	documents := make([]ContextDocument, 0, len(candidates))
	sources := make([]ContextSourceInventoryEntry, 0, len(candidates))

	for _, path := range candidates {
		relPath, err := filepath.Rel(projectDir, path)
		if err != nil {
			return nil, nil, fmt.Errorf("resolve .chc relative path: %w", err)
		}

		relPath = filepath.ToSlash(relPath)
		id := "chc:" + relPath

		content, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				sources = append(sources, ContextSourceInventoryEntry{
					ID:          id,
					Kind:        ContextSourceKindProtocol,
					Class:       ContextSourceClassInvariant,
					Path:        relPath,
					Description: "Selected .chc protocol or runtime file.",
					Required:    false,
					Included:    false,
					Reason:      "file was not found",
				})

				continue
			}

			return nil, nil, fmt.Errorf("read .chc protocol file: %w", err)
		}

		documents = append(documents, ContextDocument{
			ID:           id,
			Path:         path,
			RelativePath: relPath,
			Content:      string(content),
		})
		sources = append(sources, ContextSourceInventoryEntry{
			ID:          id,
			Kind:        ContextSourceKindProtocol,
			Class:       ContextSourceClassInvariant,
			Path:        relPath,
			Description: "Selected .chc protocol or runtime file.",
			Required:    false,
			Included:    true,
		})
	}

	return documents, sources, nil
}

func writeMarkdownField(output *strings.Builder, title, value string) {
	output.WriteString("## ")
	output.WriteString(title)
	output.WriteString("\n\n")
	output.WriteString(value)
	output.WriteString("\n\n")
}

func writeMarkdownList(output *strings.Builder, title string, values []string) {
	output.WriteString("## ")
	output.WriteString(title)
	output.WriteString("\n\n")

	for _, value := range values {
		output.WriteString("- ")
		output.WriteString(value)
		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func writeContextDocuments(output *strings.Builder, documents []ContextDocument, emptyMessage string) {
	if len(documents) == 0 {
		output.WriteString(emptyMessage)
		output.WriteString("\n\n")

		return
	}

	for _, document := range documents {
		output.WriteString("### ")
		output.WriteString(document.RelativePath)
		output.WriteString("\n\n")
		output.WriteString("```text\n")
		output.WriteString(strings.TrimSpace(document.Content))
		output.WriteString("\n```\n\n")
	}
}
