package acceptance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

const (
	// Directory defines a package boundary constant.
	Directory = "acceptance"
)

// Artifact describes a generated Gherkin feature artifact.
type Artifact struct {
	Scope            quizcore.Scope
	SpecSlug         string
	SpecPath         string
	UseCaseSlug      string
	Body             string
	SourceRequestID  string
	SourceResponseID string
	CreatedAt        time.Time
}

// Paths describes the canonical feature artifact paths.
type Paths struct {
	ScopeID      string
	RelativePath string
	Path         string
}

// Layout executes the corresponding package operation.
func Layout(projectDir string, scope quizcore.Scope) (Paths, error) {
	err := quizcore.ValidateScope(scope)
	if err != nil {
		return Paths{}, err
	}

	scopeID, err := quizcore.ScopeID(scope)
	if err != nil {
		return Paths{}, err
	}

	body, err := opscore.NamedArtifactBody(projectDir, "spec", scopeID+"-acceptance-feature")
	if err != nil {
		return Paths{}, err
	}

	return Paths{
		ScopeID:      scopeID,
		RelativePath: body.RelativePath,
		Path:         body.Path,
	}, nil
}

// WriteResult retains a confirmed path and the original static-check outcome.
// Failed checks remain observable even when there is no published path.
type WriteResult struct {
	Paths  Paths
	Checks CheckResult
}

// Write executes the corresponding package operation.
func Write(ctx context.Context, projectDir string, feature Artifact) (WriteResult, error) {
	err := ctx.Err()
	if err != nil {
		return WriteResult{}, err
	}

	result, err := Check(feature)
	if err != nil {
		return WriteResult{}, err
	}

	if !result.Valid {
		return WriteResult{Checks: result}, fmt.Errorf("acceptance feature static checks failed")
	}

	paths, err := Layout(projectDir, feature.Scope)
	if err != nil {
		return WriteResult{}, err
	}

	scopeID, err := quizcore.ScopeID(feature.Scope)
	if err != nil {
		return WriteResult{}, err
	}

	if err := opscore.EnsureNamedArtifactMetadata(projectDir, "spec", scopeID+"-acceptance-feature"); err != nil {
		return WriteResult{}, err
	}

	err = os.MkdirAll(filepath.Dir(paths.Path), 0o755)
	if err != nil {
		return WriteResult{}, fmt.Errorf("create acceptance feature directory: %w", err)
	}

	err = os.WriteFile(paths.Path, []byte(Render(feature)), 0o644)
	if err != nil {
		return WriteResult{}, fmt.Errorf("write acceptance feature: %w", err)
	}

	return WriteResult{Paths: paths, Checks: result}, nil
}

// Render renders the corresponding package data.
func Render(feature Artifact) string {
	return strings.TrimSpace(feature.Body) + "\n"
}

// Validate validates the corresponding package boundary value.
func Validate(feature Artifact) error {
	err := quizcore.ValidateScope(feature.Scope)
	if err != nil {
		return err
	}

	if strings.TrimSpace(feature.SpecSlug) == "" {
		return fmt.Errorf("acceptance feature spec slug is required")
	}

	if strings.TrimSpace(feature.SpecPath) == "" {
		return fmt.Errorf("acceptance feature spec path is required")
	}

	if strings.TrimSpace(feature.Body) == "" {
		return fmt.Errorf("acceptance feature body is required")
	}

	if feature.CreatedAt.IsZero() {
		return fmt.Errorf("acceptance feature created time is required")
	}

	return nil
}
