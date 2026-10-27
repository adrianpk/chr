package authoring

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// ProjectMetadataSchemaVersion defines a package boundary constant.
const ProjectMetadataSchemaVersion = 1

// ErrProjectMetadataNotFound identifies a package boundary error.
var ErrProjectMetadataNotFound = errors.New("project metadata not found")

// ProjectMetadata describes a package boundary value.
type ProjectMetadata struct {
	ProjectDumpRef string
	frontmatter    string
	SchemaVersion  int
	Name           string
	Description    string
	Context        string
	SourcePath     string
	SourceHash     string
	Exists         bool
}

// ProjectMetadataService describes a package boundary value.
type ProjectMetadataService struct {
	projectDir string
}

// NewProjectMetadataService constructs the corresponding package service or value.
func NewProjectMetadataService(projectDir string) ProjectMetadataService {
	return ProjectMetadataService{projectDir: projectDir}
}

// Load loads the requested data.
func (s ProjectMetadataService) Load() (ProjectMetadata, error) {
	path := ProjectMetadataPath(s.projectDir)

	relative, err := filepath.Rel(s.projectDir, path)
	if err != nil {
		return ProjectMetadata{}, err
	}

	content, err := (&fileAuthoringSessionRepository{projectDir: s.projectDir}).readAuthoringDocument(relative)
	if errors.Is(err, os.ErrNotExist) {
		return defaultProjectMetadata(s.projectDir), nil
	}

	if err != nil {
		return ProjectMetadata{}, fmt.Errorf("read project metadata: %w", err)
	}

	metadata, err := ParseProjectMetadata(string(content))
	if err != nil {
		return ProjectMetadata{}, err
	}

	metadata.SourcePath = path
	metadata.SourceHash = contenthash.Sum(content)
	metadata.Exists = true

	return metadata, nil
}

// Save persists the requested data.
func (s ProjectMetadataService) Save(metadata ProjectMetadata) (ProjectMetadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return s.SaveContext(ctx, metadata)
}

// SaveContext compares the source image inside the shared authoring owner.
func (s ProjectMetadataService) SaveContext(ctx context.Context, metadata ProjectMetadata) (saved ProjectMetadata, err error) {
	err = WithAuthoringMutation(ctx, s.projectDir, func(context.Context) error {
		saved, err = s.save(metadata, false)

		return err
	})

	return saved, err
}

func (s ProjectMetadataService) save(metadata ProjectMetadata, bindDump bool) (ProjectMetadata, error) {
	metadata.SchemaVersion = ProjectMetadataSchemaVersion

	err := ValidateProjectMetadata(metadata)
	if err != nil {
		return ProjectMetadata{}, err
	}

	current, err := s.Load()
	if err != nil {
		return ProjectMetadata{}, err
	}

	if err := validateProjectMetadataWriteHash(current, metadata); err != nil {
		return ProjectMetadata{}, err
	}

	if !bindDump {
		if metadata.ProjectDumpRef != "" && metadata.ProjectDumpRef != current.ProjectDumpRef {
			return ProjectMetadata{}, fmt.Errorf("%w: project_dump_ref changes require dump adoption or reconciliation", ErrProjectDumpReconciliation)
		}

		metadata.ProjectDumpRef = current.ProjectDumpRef
	}

	metadata.frontmatter = current.frontmatter
	path := ProjectMetadataPath(s.projectDir)

	content := RenderProjectMetadata(metadata)
	if len(content) > workflowstate.FileLimit {
		return ProjectMetadata{}, fmt.Errorf("project metadata exceeds the file limit")
	}

	err = opsfile.WriteAtomic(path, []byte(content))
	if err != nil {
		return ProjectMetadata{}, fmt.Errorf("write project metadata: %w", err)
	}

	metadata.SourcePath = path
	metadata.SourceHash = contenthash.Sum([]byte(content))
	metadata.Exists = true

	return metadata, nil
}

func validateProjectMetadataWriteHash(current, metadata ProjectMetadata) error {
	if metadata.SourceHash == "" {
		return nil
	}

	if !current.Exists {
		return fmt.Errorf("project metadata file changed outside chr: source file is missing")
	}

	if current.SourceHash != metadata.SourceHash {
		return fmt.Errorf("project metadata file changed outside chr: refresh before saving")
	}

	return nil
}

// ProjectMetadataPath executes the corresponding package operation.
func ProjectMetadataPath(projectDir string) string {
	return workspacecore.UserPath(projectDir, "project.md")
}

// ParseProjectMetadata parses the corresponding package data.
func ParseProjectMetadata(content string) (ProjectMetadata, error) {
	markdown, err := SplitMetadataMarkdown(content)
	frontmatter, body := markdown.Frontmatter, markdown.Body

	if err != nil {
		return ProjectMetadata{}, err
	}

	values, err := ProjectDocumentFields(frontmatter)
	if err != nil {
		return ProjectMetadata{}, err
	}

	schemaVersion, err := strconv.Atoi(values["schema_version"])
	if err != nil {
		return ProjectMetadata{}, fmt.Errorf("parse project metadata schema_version: %w", err)
	}

	metadata := ProjectMetadata{
		ProjectDumpRef: values["project_dump_ref"],
		frontmatter:    frontmatter,
		SchemaVersion:  schemaVersion,
		Name:           values["name"],
		Description:    values["description"],
		Context:        strings.TrimSpace(body),
	}

	err = ValidateProjectMetadata(metadata)
	if err != nil {
		return ProjectMetadata{}, err
	}

	return metadata, nil
}

// RenderProjectMetadata renders the corresponding package data.
func RenderProjectMetadata(metadata ProjectMetadata) string {
	return "---\n" + renderProjectMetadataFields(metadata) + "\n---\n\n" + strings.TrimSpace(metadata.Context) + "\n"
}

// ValidateProjectMetadata validates the corresponding package boundary value.
func ValidateProjectMetadata(metadata ProjectMetadata) error {
	if metadata.SchemaVersion != ProjectMetadataSchemaVersion {
		return fmt.Errorf("project metadata schema_version %d is unsupported", metadata.SchemaVersion)
	}

	if strings.TrimSpace(metadata.Name) == "" {
		return fmt.Errorf("project metadata name is required")
	}

	if strings.TrimSpace(metadata.Description) == "" {
		return fmt.Errorf("project metadata description is required")
	}

	return ValidateProjectDumpReference(metadata.ProjectDumpRef)
}

func defaultProjectMetadata(projectDir string) ProjectMetadata {
	name := filepath.Base(projectDir)
	if name == "." || name == string(filepath.Separator) || strings.TrimSpace(name) == "" {
		name = "project"
	}

	description := inferProjectDescriptionFromReadme(projectDir)
	if description == "" {
		description = "Project description not set."
	}

	return ProjectMetadata{
		SchemaVersion: ProjectMetadataSchemaVersion,
		Name:          name,
		Description:   description,
		Context:       "",
		SourcePath:    ProjectMetadataPath(projectDir),
	}
}

func inferProjectDescriptionFromReadme(projectDir string) string {
	content, err := os.ReadFile(filepath.Join(projectDir, "README.md"))
	if err != nil {
		return ""
	}

	return firstMarkdownParagraph(string(content), 220)
}

func firstMarkdownParagraph(content string, limit int) string {
	paragraph := []string{}

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if len(paragraph) > 0 {
				break
			}

			continue
		}

		if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "```") {
			continue
		}

		paragraph = append(paragraph, trimmed)
	}

	description := strings.Join(paragraph, " ")
	if limit > 0 && len([]rune(description)) > limit {
		runes := []rune(description)
		description = strings.TrimSpace(string(runes[:limit-1])) + "…"
	}

	return description
}

// SplitMetadataMarkdown retains the legacy required project header and body slicing.
func SplitMetadataMarkdown(content string) (MetadataMarkdown, error) {
	if !strings.HasPrefix(content, "---\n") {
		return MetadataMarkdown{Frontmatter: "", Body: ""}, fmt.Errorf("project metadata frontmatter opening delimiter is required")
	}

	remaining := strings.TrimPrefix(content, "---\n")

	end := strings.Index(remaining, "\n---")
	if end < 0 {
		return MetadataMarkdown{Frontmatter: "", Body: ""}, fmt.Errorf("project metadata frontmatter closing delimiter is required")
	}

	body := strings.TrimPrefix(remaining[end:], "\n---")
	body = strings.TrimPrefix(body, "\n")

	return MetadataMarkdown{Frontmatter: remaining[:end], Body: body}, nil
}

// ParseSimpleYAMLFields keeps the original permissive project field parsing.
func ParseSimpleYAMLFields(content string) map[string]string {
	fields := map[string]string{}

	for _, line := range strings.Split(content, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}

		fields[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}

	return fields
}
