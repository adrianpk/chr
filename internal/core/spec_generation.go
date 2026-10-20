package core

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// SpecDirectory defines a package boundary constant.
const SpecDirectory = "specs"

var generatedSpecChrPathPattern = regexp.MustCompile(`\.chr(?:/[^\s\])"'` + "`" + `]+)?`)

// GeneratedSpec describes a package boundary value.
type GeneratedSpec struct {
	ExpectedRevision string `json:"-"`
	Scope            QuizScope
	Title            string
	Body             string
	SourcePaths      []string
	SourceRequestID  string
	SourceResponseID string
	CreatedAt        time.Time
}

// GeneratedSpecPaths describes a package boundary value.
type GeneratedSpecPaths struct {
	ScopeID      string
	RelativePath string
	Path         string
}

// GeneratedSpecLayout executes the corresponding package operation.
func GeneratedSpecLayout(projectDir string, scope QuizScope) (GeneratedSpecPaths, error) {
	err := ValidateQuizScope(scope)
	if err != nil {
		return GeneratedSpecPaths{}, err
	}

	scopeID, err := quizScopeID(scope)
	if err != nil {
		return GeneratedSpecPaths{}, err
	}

	relativePath, path, err := NamedOpsArtifactBody(projectDir, "spec", scopeID+"-spec")
	if err != nil {
		return GeneratedSpecPaths{}, err
	}

	return GeneratedSpecPaths{
		ScopeID:      scopeID,
		RelativePath: relativePath,
		Path:         path,
	}, nil
}

// WriteGeneratedSpec executes the corresponding package operation.
func writeGeneratedSpec(ctx context.Context, projectDir string, spec GeneratedSpec) (GeneratedSpecPaths, error) {
	err := ctx.Err()
	if err != nil {
		return GeneratedSpecPaths{}, err
	}

	spec.Body = NormalizeGeneratedSpecChrPathReferences(spec.Body)

	err = ValidateGeneratedSpec(spec)
	if err != nil {
		return GeneratedSpecPaths{}, err
	}

	paths, err := GeneratedSpecLayout(projectDir, spec.Scope)
	if err != nil {
		return GeneratedSpecPaths{}, err
	}

	if err := publishGeneratedSpec(ctx, projectDir, paths, spec); err != nil {
		return GeneratedSpecPaths{}, err
	}

	return paths, nil
}

// RenderGeneratedSpec renders the corresponding package data.
func RenderGeneratedSpec(spec GeneratedSpec) string {
	var output bytes.Buffer

	output.WriteString("---\n")
	output.WriteString("title: " + yamlString(spec.Title) + "\n")
	output.WriteString("project_id: " + yamlString(spec.Scope.ProjectID) + "\n")
	output.WriteString("case_slug: " + yamlString(spec.Scope.CaseSlug) + "\n")
	output.WriteString("status: " + yamlString("draft") + "\n")
	output.WriteString("source_request_id: " + yamlString(spec.SourceRequestID) + "\n")
	output.WriteString("source_response_id: " + yamlString(spec.SourceResponseID) + "\n")
	output.WriteString("created_at: " + yamlTimestamp(spec.CreatedAt) + "\n")
	output.WriteString("---\n\n")
	output.WriteString("# " + spec.Title + "\n\n")
	output.WriteString(generatedSpecBodyWithoutDuplicateTitle(spec.Title, spec.Body) + "\n\n")
	output.WriteString("## Provenance\n\n")

	for _, path := range spec.SourcePaths {
		output.WriteString("- source: " + path + "\n")
	}

	return output.String()
}

func generatedSpecBodyWithoutDuplicateTitle(title, body string) string {
	body = strings.TrimSpace(body)
	heading := "# " + strings.TrimSpace(title)

	if strings.HasPrefix(body, heading) {
		remaining := strings.TrimPrefix(body, heading)
		if remaining == "" || strings.HasPrefix(remaining, "\n") || strings.HasPrefix(remaining, "\r\n") {
			return strings.TrimSpace(remaining)
		}
	}

	return body
}

// ValidateGeneratedSpec validates the corresponding package boundary value.
func ValidateGeneratedSpec(spec GeneratedSpec) error {
	if strings.TrimSpace(spec.Title) == "" {
		return fmt.Errorf("generated spec title is required")
	}

	if strings.TrimSpace(spec.Body) == "" {
		return fmt.Errorf("generated spec body is required")
	}

	err := validateGeneratedSpecChrPathReferences(spec.Body)
	if err != nil {
		return err
	}

	err = ValidateQuizScope(spec.Scope)
	if err != nil {
		return err
	}

	if spec.CreatedAt.IsZero() {
		return fmt.Errorf("generated spec created time is required")
	}

	return nil
}

func validateGeneratedSpecChrPathReferences(body string) error {
	for _, match := range generatedSpecChrPathPattern.FindAllString(body, -1) {
		path := strings.TrimRight(match, ".,;:")
		if generatedSpecChrPathReferenceAllowed(path) {
			continue
		}

		return fmt.Errorf("generated spec body references unsupported chr path %q", path)
	}

	return nil
}

// NormalizeGeneratedSpecChrPathReferences executes the corresponding package operation.
func NormalizeGeneratedSpecChrPathReferences(body string) string {
	return generatedSpecChrPathPattern.ReplaceAllStringFunc(body, func(match string) string {
		path := strings.TrimRight(match, ".,;:")
		suffix := strings.TrimPrefix(match, path)

		if generatedSpecChrPathReferenceAllowed(path) {
			return match
		}

		if strings.HasPrefix(path, ".chr/default/") {
			return match
		}

		if strings.HasPrefix(path, ".chr/") {
			return ".chr/default/" + strings.TrimPrefix(path, ".chr/") + suffix
		}

		return path + suffix
	})
}

func generatedSpecChrPathReferenceAllowed(path string) bool {
	if path == ".chr" || path == ".chr/" || path == ".chr/default" || path == ".chr/default/" {
		return true
	}

	switch path {
	case ".chr/default/config.yml", ".chr/default/manifest.md", ".chr/default/project.md":
		return true
	}

	for _, directory := range []string{
		"agents",
		"cases",
		"dumps",
		"quizzes",
		"specs",
		"plans",
		"tickets",
		"reports",
		"runtime",
		"tmp",
	} {
		prefix := ".chr/default/" + directory
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}

	return false
}
