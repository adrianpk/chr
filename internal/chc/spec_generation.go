package chc

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const SpecDirectory = "specs"

var generatedSpecChrysoCoordPathPattern = regexp.MustCompile(`\.chc(?:/[^\s\])"'` + "`" + `]+)?`)

type GeneratedSpec struct {
	Scope            QuizScope
	Title            string
	Body             string
	SourcePaths      []string
	SourceRequestID  string
	SourceResponseID string
	CreatedAt        time.Time
}

type GeneratedSpecPaths struct {
	ScopeID      string
	RelativePath string
	Path         string
}

func GeneratedSpecLayout(projectDir string, scope QuizScope) (GeneratedSpecPaths, error) {
	err := ValidateQuizScope(scope)
	if err != nil {
		return GeneratedSpecPaths{}, err
	}

	scopeID, err := quizScopeID(scope)
	if err != nil {
		return GeneratedSpecPaths{}, err
	}

	relativePath := UserRelativePath(SpecDirectory, scopeID+"-spec.md")

	return GeneratedSpecPaths{
		ScopeID:      scopeID,
		RelativePath: relativePath,
		Path:         filepath.Join(projectDir, filepath.FromSlash(relativePath)),
	}, nil
}

func WriteGeneratedSpec(ctx context.Context, projectDir string, spec GeneratedSpec) (GeneratedSpecPaths, error) {
	err := ctx.Err()
	if err != nil {
		return GeneratedSpecPaths{}, err
	}

	spec.Body = NormalizeGeneratedSpecChrysoCoordPathReferences(spec.Body)

	err = ValidateGeneratedSpec(spec)
	if err != nil {
		return GeneratedSpecPaths{}, err
	}

	paths, err := GeneratedSpecLayout(projectDir, spec.Scope)
	if err != nil {
		return GeneratedSpecPaths{}, err
	}

	err = os.MkdirAll(filepath.Dir(paths.Path), 0o755)
	if err != nil {
		return GeneratedSpecPaths{}, fmt.Errorf("create spec directory: %w", err)
	}

	err = os.WriteFile(paths.Path, []byte(RenderGeneratedSpec(spec)), 0o644)
	if err != nil {
		return GeneratedSpecPaths{}, fmt.Errorf("write generated spec: %w", err)
	}

	return paths, nil
}

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

func ValidateGeneratedSpec(spec GeneratedSpec) error {
	if strings.TrimSpace(spec.Title) == "" {
		return fmt.Errorf("generated spec title is required")
	}

	if strings.TrimSpace(spec.Body) == "" {
		return fmt.Errorf("generated spec body is required")
	}

	err := validateGeneratedSpecChrysoCoordPathReferences(spec.Body)
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

func validateGeneratedSpecChrysoCoordPathReferences(body string) error {
	for _, match := range generatedSpecChrysoCoordPathPattern.FindAllString(body, -1) {
		path := strings.TrimRight(match, ".,;:")
		if generatedSpecChrysoCoordPathReferenceAllowed(path) {
			continue
		}

		return fmt.Errorf("generated spec body references unsupported ChrysoCoord path %q", path)
	}

	return nil
}

func NormalizeGeneratedSpecChrysoCoordPathReferences(body string) string {
	return generatedSpecChrysoCoordPathPattern.ReplaceAllStringFunc(body, func(match string) string {
		path := strings.TrimRight(match, ".,;:")
		suffix := strings.TrimPrefix(match, path)

		if generatedSpecChrysoCoordPathReferenceAllowed(path) {
			return match
		}

		if strings.HasPrefix(path, ".chc/") {
			return ".chc/user/" + strings.TrimPrefix(path, ".chc/") + suffix
		}

		return path + suffix
	})
}

func generatedSpecChrysoCoordPathReferenceAllowed(path string) bool {
	if path == ".chc" || path == ".chc/" || path == ".chc/user" || path == ".chc/user/" {
		return true
	}

	switch path {
	case ".chc/user/config.yml", ".chc/user/manifest.md", ".chc/user/project.md":
		return true
	}

	for _, directory := range []string{
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
		prefix := ".chc/user/" + directory
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}

	return false
}
