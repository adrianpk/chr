package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// CaseMetadataSchemaVersion defines a package boundary constant.
const CaseMetadataSchemaVersion = 1

// ErrCaseMetadataNotFound identifies a package boundary error.
var ErrCaseMetadataNotFound = errors.New("case metadata not found")

// CaseMetadata describes a package boundary value.
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

// CaseMetadataService describes a package boundary value.
type CaseMetadataService struct {
	afterMutationReplace func(int) error
	projectDir           string
}

// NewCaseMetadataService constructs the corresponding package service or value.
func NewCaseMetadataService(projectDir string) CaseMetadataService {
	return CaseMetadataService{projectDir: projectDir}
}

// LoadAll loads the requested data.
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

// LoadPath loads the requested data.
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

// Save persists the requested data.
func (s CaseMetadataService) Save(metadata CaseMetadata) (CaseMetadata, error) {
	return s.SaveContext(context.Background(), metadata)
}

// SaveContext preserves cancellation while serializing metadata writes and recovery.
func (s CaseMetadataService) SaveContext(ctx context.Context, metadata CaseMetadata) (saved CaseMetadata, err error) {
	err = s.withMutationLock(ctx, func() error {
		var saveErr error

		saved, saveErr = s.save(metadata)

		return saveErr
	})

	return saved, err
}

func (s CaseMetadataService) save(metadata CaseMetadata) (CaseMetadata, error) {
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
		return fmt.Errorf("case metadata file changed outside chr: source file is missing")
	}

	if err != nil {
		return fmt.Errorf("read current case metadata before write: %w", err)
	}

	currentHash := contentHash(current)
	if currentHash != metadata.SourceHash {
		return fmt.Errorf("case metadata file changed outside chr: refresh before saving")
	}

	return nil
}

// CaseMetadataDirectory executes the corresponding package operation.
func CaseMetadataDirectory(projectDir string) string {
	return UserPath(projectDir, "cases")
}

// CaseMetadataPath executes the corresponding package operation.
func CaseMetadataPath(projectDir, slug string) string {
	return filepath.Join(CaseMetadataDirectory(projectDir), slug+".md")
}

// ParseCaseMetadata parses the corresponding package data.
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

// RenderCaseMetadata renders the corresponding package data.
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

// ValidateCaseMetadata validates the corresponding package boundary value.
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
