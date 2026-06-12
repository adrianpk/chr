package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ProjectMetadataSchemaVersion defines a package boundary constant.
const ProjectMetadataSchemaVersion = 1

// ErrProjectMetadataNotFound identifies a package boundary error.
var ErrProjectMetadataNotFound = errors.New("project metadata not found")

// ProjectMetadata describes a package boundary value.
type ProjectMetadata struct {
	SchemaVersion int
	Name          string
	Description   string
	Context       string
	SourcePath    string
	SourceHash    string
	Exists        bool
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

	content, err := os.ReadFile(path)
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
	metadata.SourceHash = contentHash(content)
	metadata.Exists = true

	return metadata, nil
}

// Save persists the requested data.
func (s ProjectMetadataService) Save(metadata ProjectMetadata) (ProjectMetadata, error) {
	metadata.SchemaVersion = ProjectMetadataSchemaVersion

	err := ValidateProjectMetadata(metadata)
	if err != nil {
		return ProjectMetadata{}, err
	}

	path := ProjectMetadataPath(s.projectDir)

	err = validateProjectMetadataWriteHash(path, metadata)
	if err != nil {
		return ProjectMetadata{}, err
	}

	content := RenderProjectMetadata(metadata)

	err = writeFileAtomically(path, []byte(content))
	if err != nil {
		return ProjectMetadata{}, fmt.Errorf("write project metadata: %w", err)
	}

	metadata.SourcePath = path
	metadata.SourceHash = contentHash([]byte(content))
	metadata.Exists = true

	return metadata, nil
}

func validateProjectMetadataWriteHash(path string, metadata ProjectMetadata) error {
	if metadata.SourceHash == "" {
		return nil
	}

	current, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("project metadata file changed outside CHC: source file is missing")
	}

	if err != nil {
		return fmt.Errorf("read current project metadata before write: %w", err)
	}

	currentHash := contentHash(current)
	if currentHash != metadata.SourceHash {
		return fmt.Errorf("project metadata file changed outside CHC: refresh before saving")
	}

	return nil
}

// ProjectMetadataPath executes the corresponding package operation.
func ProjectMetadataPath(projectDir string) string {
	return UserPath(projectDir, "project.md")
}

// ParseProjectMetadata parses the corresponding package data.
func ParseProjectMetadata(content string) (ProjectMetadata, error) {
	frontmatter, body, err := splitMarkdownFrontmatter(content)
	if err != nil {
		return ProjectMetadata{}, err
	}

	fields := parseSimpleYAMLFields(frontmatter)

	schemaVersion, err := strconv.Atoi(unquoteMetadataString(fields["schema_version"]))
	if err != nil {
		return ProjectMetadata{}, fmt.Errorf("parse project metadata schema_version: %w", err)
	}

	metadata := ProjectMetadata{
		SchemaVersion: schemaVersion,
		Name:          unquoteMetadataString(fields["name"]),
		Description:   unquoteMetadataString(fields["description"]),
		Context:       strings.TrimSpace(body),
	}

	err = ValidateProjectMetadata(metadata)
	if err != nil {
		return ProjectMetadata{}, err
	}

	return metadata, nil
}

// RenderProjectMetadata renders the corresponding package data.
func RenderProjectMetadata(metadata ProjectMetadata) string {
	var output strings.Builder

	output.WriteString("---\n")
	output.WriteString("schema_version: ")
	output.WriteString(strconv.Itoa(ProjectMetadataSchemaVersion))
	output.WriteString("\n")
	output.WriteString("name: ")
	output.WriteString(yamlString(metadata.Name))
	output.WriteString("\n")
	output.WriteString("description: ")
	output.WriteString(yamlString(metadata.Description))
	output.WriteString("\n")
	output.WriteString("---\n\n")
	output.WriteString(strings.TrimSpace(metadata.Context))
	output.WriteString("\n")

	return output.String()
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

	return nil
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

func splitMarkdownFrontmatter(content string) (string, string, error) {
	if !strings.HasPrefix(content, "---\n") {
		return "", "", fmt.Errorf("project metadata frontmatter opening delimiter is required")
	}

	remaining := strings.TrimPrefix(content, "---\n")

	end := strings.Index(remaining, "\n---")
	if end < 0 {
		return "", "", fmt.Errorf("project metadata frontmatter closing delimiter is required")
	}

	body := strings.TrimPrefix(remaining[end:], "\n---")
	body = strings.TrimPrefix(body, "\n")

	return remaining[:end], body, nil
}

func parseSimpleYAMLFields(content string) map[string]string {
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

func writeFileAtomically(path string, content []byte) error {
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}

	tmpPath := path + ".tmp"

	err = os.WriteFile(tmpPath, content, 0o644)
	if err != nil {
		return fmt.Errorf("write temporary file: %w", err)
	}

	err = os.Rename(tmpPath, path)
	if err != nil {
		return fmt.Errorf("rename temporary file: %w", err)
	}

	return nil
}
