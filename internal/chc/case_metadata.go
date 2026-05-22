package chc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const CaseMetadataSchemaVersion = 1

var ErrCaseMetadataNotFound = errors.New("case metadata not found")

type CaseMetadata struct {
	SchemaVersion int
	Slug          string
	Name          string
	Description   string
	Context       string
	SourcePath    string
	SourceHash    string
	Exists        bool
}

type IndexedCaseMetadata struct {
	ProjectID string
	Case      CaseMetadata
	IndexedAt string
}

type CaseMetadataRepository interface {
	UpsertCaseMetadata(ctx context.Context, record IndexedCaseMetadata) error
	ListProjectCaseMetadata(ctx context.Context, projectID string) ([]IndexedCaseMetadata, error)
	Close() error
}

type sqliteCaseMetadataRepository struct {
	db *sql.DB
}

type caseMetadataScanner interface {
	Scan(dest ...any) error
}

func NewCaseMetadataRepository(
	ctx context.Context,
	env UserEnvironment,
	assetsFS fs.FS,
) (CaseMetadataRepository, error) {
	err := os.MkdirAll(env.DataDir(), 0o755)
	if err != nil {
		return nil, fmt.Errorf("create user data directory: %w", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(env.DataDir(), userDatabaseName))
	if err != nil {
		return nil, fmt.Errorf("open case metadata database: %w", err)
	}

	db.SetMaxOpenConns(1)

	repository, err := NewCaseMetadataRepositoryFromDB(ctx, db, assetsFS)
	if err != nil {
		return nil, closeDBWithError(db, err)
	}

	return repository, nil
}

func NewCaseMetadataRepositoryFromDB(
	ctx context.Context,
	db *sql.DB,
	assetsFS fs.FS,
) (CaseMetadataRepository, error) {
	migrator := NewMigrator(assetsFS)

	err := migrator.Migrate(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("migrate case metadata database: %w", err)
	}

	return &sqliteCaseMetadataRepository{db: db}, nil
}

func (r *sqliteCaseMetadataRepository) UpsertCaseMetadata(
	ctx context.Context,
	record IndexedCaseMetadata,
) error {
	err := validateIndexedCaseMetadata(record)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO indexed_cases (
    project_id,
    slug,
    name,
    description,
    context,
    source_path,
    source_hash,
    schema_version,
    indexed_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(project_id, slug) DO UPDATE SET
    name = excluded.name,
    description = excluded.description,
    context = excluded.context,
    source_path = excluded.source_path,
    source_hash = excluded.source_hash,
    schema_version = excluded.schema_version,
    indexed_at = excluded.indexed_at`,
		record.ProjectID,
		record.Case.Slug,
		record.Case.Name,
		record.Case.Description,
		record.Case.Context,
		record.Case.SourcePath,
		record.Case.SourceHash,
		record.Case.SchemaVersion,
		record.IndexedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert case metadata: %w", err)
	}

	return nil
}

func (r *sqliteCaseMetadataRepository) ListProjectCaseMetadata(
	ctx context.Context,
	projectID string,
) ([]IndexedCaseMetadata, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT
    project_id,
    slug,
    name,
    description,
    context,
    source_path,
    source_hash,
    schema_version,
    indexed_at
FROM indexed_cases
WHERE project_id = ?
ORDER BY name, slug`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list case metadata: %w", err)
	}

	defer func() {
		//nolint:errcheck // Query errors are returned through rows.Err below.
		_ = rows.Close()
	}()

	records := []IndexedCaseMetadata{}

	for rows.Next() {
		record, err := scanIndexedCaseMetadata(rows)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("iterate case metadata: %w", err)
	}

	return records, nil
}

func (r *sqliteCaseMetadataRepository) Close() error {
	err := r.db.Close()
	if err != nil {
		return fmt.Errorf("close case metadata repository: %w", err)
	}

	return nil
}

func scanIndexedCaseMetadata(scanner caseMetadataScanner) (IndexedCaseMetadata, error) {
	var record IndexedCaseMetadata

	err := scanner.Scan(
		&record.ProjectID,
		&record.Case.Slug,
		&record.Case.Name,
		&record.Case.Description,
		&record.Case.Context,
		&record.Case.SourcePath,
		&record.Case.SourceHash,
		&record.Case.SchemaVersion,
		&record.IndexedAt,
	)
	if err != nil {
		return IndexedCaseMetadata{}, err
	}

	record.Case.Exists = true

	return record, nil
}

func validateIndexedCaseMetadata(record IndexedCaseMetadata) error {
	if strings.TrimSpace(record.ProjectID) == "" {
		return fmt.Errorf("case metadata project id is required")
	}

	if strings.TrimSpace(record.IndexedAt) == "" {
		return fmt.Errorf("case metadata indexed_at is required")
	}

	err := ValidateCaseMetadata(record.Case)
	if err != nil {
		return err
	}

	if strings.TrimSpace(record.Case.SourcePath) == "" {
		return fmt.Errorf("case metadata source path is required")
	}

	if strings.TrimSpace(record.Case.SourceHash) == "" {
		return fmt.Errorf("case metadata source hash is required")
	}

	return nil
}

type CaseMetadataService struct {
	projectDir string
}

func NewCaseMetadataService(projectDir string) CaseMetadataService {
	return CaseMetadataService{projectDir: projectDir}
}

func (s CaseMetadataService) LoadAll() ([]CaseMetadata, error) {
	root := CaseMetadataDirectory(s.projectDir)

	entries, err := os.ReadDir(root)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("read case metadata directory: %w", err)
	}

	cases := []CaseMetadata{}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		metadata, err := s.LoadPath(filepath.Join(root, entry.Name()))
		if err != nil {
			return nil, err
		}

		cases = append(cases, metadata)
	}

	sort.SliceStable(cases, func(i, j int) bool {
		if strings.EqualFold(cases[i].Name, cases[j].Name) {
			return cases[i].Slug < cases[j].Slug
		}

		return strings.ToLower(cases[i].Name) < strings.ToLower(cases[j].Name)
	})

	return cases, nil
}

func (s CaseMetadataService) LoadPath(path string) (CaseMetadata, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return CaseMetadata{}, fmt.Errorf("read case metadata: %w", err)
	}

	metadata, err := ParseCaseMetadata(string(content))
	if err != nil {
		return CaseMetadata{}, err
	}

	metadata.SourcePath = path
	metadata.SourceHash = contentHash(content)
	metadata.Exists = true

	if strings.TrimSpace(metadata.Slug) == "" {
		metadata.Slug = caseSlugFromPath(path)
	}

	return metadata, nil
}

func (s CaseMetadataService) Save(metadata CaseMetadata) (CaseMetadata, error) {
	metadata.SchemaVersion = CaseMetadataSchemaVersion

	err := ValidateCaseMetadata(metadata)
	if err != nil {
		return CaseMetadata{}, err
	}

	path := metadata.SourcePath
	if strings.TrimSpace(path) == "" {
		path = s.newCaseMetadataPath(metadata.Name)
	}

	if strings.TrimSpace(metadata.Slug) == "" {
		metadata.Slug = caseSlugFromPath(path)
	}

	err = validateCaseMetadataWriteHash(path, metadata)
	if err != nil {
		return CaseMetadata{}, err
	}

	content := RenderCaseMetadata(metadata)

	err = writeFileAtomically(path, []byte(content))
	if err != nil {
		return CaseMetadata{}, fmt.Errorf("write case metadata: %w", err)
	}

	metadata.SourcePath = path
	metadata.SourceHash = contentHash([]byte(content))
	metadata.Exists = true

	return metadata, nil
}

func (s CaseMetadataService) newCaseMetadataPath(name string) string {
	root := CaseMetadataDirectory(s.projectDir)
	baseSlug := slugifyCaseName(name)
	path := filepath.Join(root, baseSlug+".md")

	for suffix := 2; fileExists(path); suffix++ {
		path = filepath.Join(root, fmt.Sprintf("%s-%d.md", baseSlug, suffix))
	}

	return path
}

func validateCaseMetadataWriteHash(path string, metadata CaseMetadata) error {
	if metadata.SourceHash == "" {
		return nil
	}

	current, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("case metadata file changed outside CHC: source file is missing")
	}

	if err != nil {
		return fmt.Errorf("read current case metadata before write: %w", err)
	}

	currentHash := contentHash(current)
	if currentHash != metadata.SourceHash {
		return fmt.Errorf("case metadata file changed outside CHC: resync before saving")
	}

	return nil
}

func CaseMetadataDirectory(projectDir string) string {
	return filepath.Join(projectDir, DirectoryName, "cases")
}

func CaseMetadataPath(projectDir, slug string) string {
	return filepath.Join(CaseMetadataDirectory(projectDir), slug+".md")
}

func ParseCaseMetadata(content string) (CaseMetadata, error) {
	frontmatter, body, err := splitMarkdownFrontmatter(content)
	if err != nil {
		return CaseMetadata{}, fmt.Errorf("parse case metadata: %w", err)
	}

	fields := parseSimpleYAMLFields(frontmatter)

	schemaVersion, err := strconv.Atoi(unquoteMetadataString(fields["schema_version"]))
	if err != nil {
		return CaseMetadata{}, fmt.Errorf("parse case metadata schema_version: %w", err)
	}

	metadata := CaseMetadata{
		SchemaVersion: schemaVersion,
		Name:          unquoteMetadataString(fields["name"]),
		Description:   unquoteMetadataString(fields["description"]),
		Context:       strings.TrimSpace(body),
	}

	err = ValidateCaseMetadata(metadata)
	if err != nil {
		return CaseMetadata{}, err
	}

	return metadata, nil
}

func RenderCaseMetadata(metadata CaseMetadata) string {
	var output strings.Builder

	output.WriteString("---\n")
	output.WriteString("schema_version: ")
	output.WriteString(strconv.Itoa(CaseMetadataSchemaVersion))
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

func ValidateCaseMetadata(metadata CaseMetadata) error {
	if metadata.SchemaVersion != CaseMetadataSchemaVersion {
		return fmt.Errorf("case metadata schema_version %d is unsupported", metadata.SchemaVersion)
	}

	if strings.TrimSpace(metadata.Name) == "" {
		return fmt.Errorf("case metadata name is required")
	}

	return nil
}

func slugifyCaseName(name string) string {
	var output strings.Builder

	previousDash := false

	for _, current := range strings.ToLower(strings.TrimSpace(name)) {
		switch {
		case unicode.IsLetter(current) || unicode.IsDigit(current):
			output.WriteRune(current)

			previousDash = false
		case !previousDash:
			output.WriteRune('-')

			previousDash = true
		}
	}

	slug := strings.Trim(output.String(), "-")
	if slug == "" {
		return "case"
	}

	return slug
}

func caseSlugFromPath(path string) string {
	slug := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if strings.TrimSpace(slug) == "" {
		return "case"
	}

	return slug
}

func fileExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}
