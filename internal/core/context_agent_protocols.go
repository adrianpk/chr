package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
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
	AgentFiles    []authoringcore.ContextDocument
	ChrFiles      []authoringcore.ContextDocument
	Sources       []authoringcore.ContextSourceInventoryEntry
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

type agentProtocolPaths struct {
	ProjectDir string
	TargetPath string
	TargetDir  string
}

// LoadAgentProtocolContext loads the corresponding package data.
func LoadAgentProtocolContext(opts AgentProtocolContextOptions) (AgentProtocolContext, error) {
	paths, err := normalizeAgentProtocolContextOptions(opts)
	if err != nil {
		return AgentProtocolContext{}, err
	}

	agent, err := loadApplicableAgentFiles(paths.ProjectDir, paths.TargetDir)
	if err != nil {
		return AgentProtocolContext{}, err
	}

	protocols, err := loadChrProtocolFiles(paths.ProjectDir)
	if err != nil {
		return AgentProtocolContext{}, err
	}

	sources := []authoringcore.ContextSourceInventoryEntry{
		{
			ID:          defaultAgentIdentityID,
			Kind:        authoringcore.ContextSourceKindAgentIdentity,
			Class:       authoringcore.ContextSourceClassInvariant,
			Description: "Default Go software and infrastructure agent identity.",
			Required:    true,
			Included:    true,
		},
	}
	sources = append(sources, agent.Sources...)
	sources = append(sources, protocols.Sources...)

	return AgentProtocolContext{
		ProjectDir:    paths.ProjectDir,
		TargetPath:    paths.TargetPath,
		AgentIdentity: DefaultAgentIdentity(),
		AgentFiles:    agent.Documents,
		ChrFiles:      protocols.Documents,
		Sources:       sources,
	}, nil
}

// RenderAgentContext renders the corresponding package data.
func RenderAgentContext(identity AgentIdentity) string {
	var output strings.Builder

	output.WriteString("# Agent Context\n\n")
	opsmetadata.WriteMarkdownField(&output, "Identity", identity.Title)
	opsmetadata.WriteMarkdownField(&output, "Role", identity.Role)
	opsmetadata.WriteMarkdownList(&output, "Operating Stance", identity.OperatingStance)
	opsmetadata.WriteMarkdownList(&output, "Domain Expectations", identity.DomainExpectations)
	opsmetadata.WriteMarkdownList(&output, "Quality Bar", identity.QualityBar)
	opsmetadata.WriteMarkdownList(&output, "Language Rules", identity.LanguageRules)
	opsmetadata.WriteMarkdownList(&output, "Forbidden Behaviors", identity.ForbiddenBehaviors)

	return output.String()
}

// RenderProtocolsContext renders the corresponding package data.
func RenderProtocolsContext(ctx AgentProtocolContext) string {
	var output strings.Builder

	output.WriteString("# Protocol Context\n\n")
	opsmetadata.WriteMarkdownField(&output, "Project", ctx.ProjectDir)
	opsmetadata.WriteMarkdownField(&output, "Target", ctx.TargetPath)

	output.WriteString("## Applicable AGENTS.md Files\n\n")
	WriteContextDocuments(&output, ctx.AgentFiles, "No applicable AGENTS.md files were found.")

	output.WriteString("## Selected .chr Context\n\n")
	WriteContextDocuments(&output, ctx.ChrFiles, "No selected .chr protocol files were found.")

	return output.String()
}

func normalizeAgentProtocolContextOptions(opts AgentProtocolContextOptions) (agentProtocolPaths, error) {
	projectDir, err := CleanContextProjectDir(opts.ProjectDir)
	if err != nil {
		return agentProtocolPaths{}, err
	}

	targetPath, err := cleanAbsTargetPath(projectDir, opts.TargetPath)
	if err != nil {
		return agentProtocolPaths{}, err
	}

	if !projectcore.ContainsPath(projectDir, targetPath) {
		return agentProtocolPaths{}, fmt.Errorf("target path is outside project: %s", targetPath)
	}

	return agentProtocolPaths{ProjectDir: projectDir, TargetPath: targetPath, TargetDir: targetDirectory(targetPath)}, nil
}

// CleanContextProjectDir returns the original trimmed absolute context root.
func CleanContextProjectDir(projectDir string) (string, error) {
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

func targetDirectory(targetPath string) string {
	info, err := os.Stat(targetPath)
	if err == nil && info.IsDir() {
		return targetPath
	}

	return filepath.Dir(targetPath)
}

func loadApplicableAgentFiles(projectDir, targetDir string) (authoringcore.ContextDocumentLoad, error) {
	dirs, err := dirsFromProjectToTarget(projectDir, targetDir)
	if err != nil {
		return authoringcore.ContextDocumentLoad{}, err
	}

	documents := make([]authoringcore.ContextDocument, 0, len(dirs))
	sources := make([]authoringcore.ContextSourceInventoryEntry, 0, len(dirs))

	for _, dir := range dirs {
		path := filepath.Join(dir, "AGENTS.md")

		relPath, err := filepath.Rel(projectDir, path)
		if err != nil {
			return authoringcore.ContextDocumentLoad{}, fmt.Errorf("resolve AGENTS.md relative path: %w", err)
		}

		relPath = filepath.ToSlash(relPath)
		id := "agents:" + relPath

		content, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				sources = append(sources, authoringcore.ContextSourceInventoryEntry{
					ID:          id,
					Kind:        authoringcore.ContextSourceKindProtocol,
					Class:       authoringcore.ContextSourceClassInvariant,
					Path:        relPath,
					Description: "Applicable AGENTS.md protocol file.",
					Required:    dir == projectDir,
					Included:    false,
					Reason:      "file was not found",
				})

				continue
			}

			return authoringcore.ContextDocumentLoad{}, fmt.Errorf("read AGENTS.md: %w", err)
		}

		documents = append(documents, authoringcore.ContextDocument{
			ID:           id,
			Path:         path,
			RelativePath: relPath,
			Content:      string(content),
		})
		sources = append(sources, authoringcore.ContextSourceInventoryEntry{
			ID:          id,
			Kind:        authoringcore.ContextSourceKindProtocol,
			Class:       authoringcore.ContextSourceClassInvariant,
			Path:        relPath,
			Description: "Applicable AGENTS.md protocol file.",
			Required:    dir == projectDir,
			Included:    true,
		})
	}

	return authoringcore.ContextDocumentLoad{Documents: documents, Sources: sources}, nil
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

func loadChrProtocolFiles(projectDir string) (authoringcore.ContextDocumentLoad, error) {
	candidates, err := chrProtocolCandidates(projectDir)
	if err != nil {
		return authoringcore.ContextDocumentLoad{}, err
	}

	return loadProtocolCandidateFiles(projectDir, candidates)
}

func chrProtocolCandidates(projectDir string) ([]string, error) {
	userDir := workspacecore.UserPath(projectDir)
	candidates := stableChrProtocolCandidates(userDir)
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
		return nil, fmt.Errorf("walk .chr user agents: %w", err)
	}

	sort.Strings(candidates)

	return candidates, nil
}

func stableChrProtocolCandidates(userDir string) []string {
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
) (authoringcore.ContextDocumentLoad, error) {
	documents := make([]authoringcore.ContextDocument, 0, len(candidates))
	sources := make([]authoringcore.ContextSourceInventoryEntry, 0, len(candidates))

	for _, path := range candidates {
		relPath, err := filepath.Rel(projectDir, path)
		if err != nil {
			return authoringcore.ContextDocumentLoad{}, fmt.Errorf("resolve .chr relative path: %w", err)
		}

		relPath = filepath.ToSlash(relPath)
		id := "chr:" + relPath

		content, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				sources = append(sources, authoringcore.ContextSourceInventoryEntry{
					ID:          id,
					Kind:        authoringcore.ContextSourceKindProtocol,
					Class:       authoringcore.ContextSourceClassInvariant,
					Path:        relPath,
					Description: "Selected .chr protocol or runtime file.",
					Required:    false,
					Included:    false,
					Reason:      "file was not found",
				})

				continue
			}

			return authoringcore.ContextDocumentLoad{}, fmt.Errorf("read .chr protocol file: %w", err)
		}

		documents = append(documents, authoringcore.ContextDocument{
			ID:           id,
			Path:         path,
			RelativePath: relPath,
			Content:      string(content),
		})
		sources = append(sources, authoringcore.ContextSourceInventoryEntry{
			ID:          id,
			Kind:        authoringcore.ContextSourceKindProtocol,
			Class:       authoringcore.ContextSourceClassInvariant,
			Path:        relPath,
			Description: "Selected .chr protocol or runtime file.",
			Required:    false,
			Included:    true,
		})
	}

	return authoringcore.ContextDocumentLoad{Documents: documents, Sources: sources}, nil
}

// WriteContextDocuments renders the existing document context and empty message.
func WriteContextDocuments(output *strings.Builder, documents []authoringcore.ContextDocument, emptyMessage string) {
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
