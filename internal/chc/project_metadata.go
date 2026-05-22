package chc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const ProjectMetadataSchemaVersion = 1

var ErrProjectMetadataNotFound = errors.New("project metadata not found")

type ProjectMetadata struct {
	SchemaVersion int
	Name          string
	Description   string
	Context       string
	SourcePath    string
	SourceHash    string
	Exists        bool
}

type IndexedProjectMetadata struct {
	ProjectID string
	Metadata  ProjectMetadata
	IndexedAt string
}

type ProjectMetadataRepository interface {
	UpsertProjectMetadata(ctx context.Context, record IndexedProjectMetadata) error
	GetProjectMetadata(ctx context.Context, projectID string) (IndexedProjectMetadata, error)
	Close() error
}

type sqliteProjectMetadataRepository struct {
	db *sql.DB
}

type projectMetadataScanner interface {
	Scan(dest ...any) error
}

func NewProjectMetadataRepository(
	ctx context.Context,
	env UserEnvironment,
	assetsFS fs.FS,
) (ProjectMetadataRepository, error) {
	err := os.MkdirAll(env.DataDir(), 0o755)
	if err != nil {
		return nil, fmt.Errorf("create user data directory: %w", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(env.DataDir(), userDatabaseName))
	if err != nil {
		return nil, fmt.Errorf("open project metadata database: %w", err)
	}

	db.SetMaxOpenConns(1)

	repository, err := NewProjectMetadataRepositoryFromDB(ctx, db, assetsFS)
	if err != nil {
		return nil, closeDBWithError(db, err)
	}

	return repository, nil
}

func NewProjectMetadataRepositoryFromDB(
	ctx context.Context,
	db *sql.DB,
	assetsFS fs.FS,
) (ProjectMetadataRepository, error) {
	migrator := NewMigrator(assetsFS)

	err := migrator.Migrate(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("migrate project metadata database: %w", err)
	}

	return &sqliteProjectMetadataRepository{db: db}, nil
}

func (r *sqliteProjectMetadataRepository) UpsertProjectMetadata(
	ctx context.Context,
	record IndexedProjectMetadata,
) error {
	err := validateIndexedProjectMetadata(record)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO indexed_project_metadata (
    project_id,
    name,
    description,
    context,
    source_path,
    source_hash,
    schema_version,
    indexed_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(project_id) DO UPDATE SET
    name = excluded.name,
    description = excluded.description,
    context = excluded.context,
    source_path = excluded.source_path,
    source_hash = excluded.source_hash,
    schema_version = excluded.schema_version,
    indexed_at = excluded.indexed_at`,
		record.ProjectID,
		record.Metadata.Name,
		record.Metadata.Description,
		record.Metadata.Context,
		record.Metadata.SourcePath,
		record.Metadata.SourceHash,
		record.Metadata.SchemaVersion,
		record.IndexedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert project metadata: %w", err)
	}

	return nil
}

func (r *sqliteProjectMetadataRepository) GetProjectMetadata(
	ctx context.Context,
	projectID string,
) (IndexedProjectMetadata, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT
    project_id,
    name,
    description,
    context,
    source_path,
    source_hash,
    schema_version,
    indexed_at
FROM indexed_project_metadata
WHERE project_id = ?`,
		projectID,
	)

	record, err := scanIndexedProjectMetadata(row)
	if errors.Is(err, sql.ErrNoRows) {
		return IndexedProjectMetadata{}, fmt.Errorf("%w: %s", ErrProjectMetadataNotFound, projectID)
	}

	if err != nil {
		return IndexedProjectMetadata{}, err
	}

	return record, nil
}

func (r *sqliteProjectMetadataRepository) Close() error {
	err := r.db.Close()
	if err != nil {
		return fmt.Errorf("close project metadata repository: %w", err)
	}

	return nil
}

func scanIndexedProjectMetadata(scanner projectMetadataScanner) (IndexedProjectMetadata, error) {
	var record IndexedProjectMetadata

	err := scanner.Scan(
		&record.ProjectID,
		&record.Metadata.Name,
		&record.Metadata.Description,
		&record.Metadata.Context,
		&record.Metadata.SourcePath,
		&record.Metadata.SourceHash,
		&record.Metadata.SchemaVersion,
		&record.IndexedAt,
	)
	if err != nil {
		return IndexedProjectMetadata{}, err
	}

	record.Metadata.Exists = true

	return record, nil
}

func validateIndexedProjectMetadata(record IndexedProjectMetadata) error {
	if strings.TrimSpace(record.ProjectID) == "" {
		return fmt.Errorf("project metadata project id is required")
	}

	if strings.TrimSpace(record.IndexedAt) == "" {
		return fmt.Errorf("project metadata indexed_at is required")
	}

	err := ValidateProjectMetadata(record.Metadata)
	if err != nil {
		return err
	}

	if strings.TrimSpace(record.Metadata.SourcePath) == "" {
		return fmt.Errorf("project metadata source path is required")
	}

	if strings.TrimSpace(record.Metadata.SourceHash) == "" {
		return fmt.Errorf("project metadata source hash is required")
	}

	return nil
}

type ProjectMetadataService struct {
	projectDir string
}

func NewProjectMetadataService(projectDir string) ProjectMetadataService {
	return ProjectMetadataService{projectDir: projectDir}
}

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
		return fmt.Errorf("project metadata file changed outside CHC: resync before saving")
	}

	return nil
}

func ProjectMetadataPath(projectDir string) string {
	return filepath.Join(projectDir, DirectoryName, "project.md")
}

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
