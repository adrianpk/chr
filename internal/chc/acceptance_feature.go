package chc

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

const (
	// AcceptanceFeatureDirectory defines a package boundary constant.
	AcceptanceFeatureDirectory = "acceptance"
)

// AcceptanceFeatureArtifact describes a generated Gherkin feature artifact.
type AcceptanceFeatureArtifact struct {
	Scope            QuizScope
	SpecSlug         string
	SpecPath         string
	UseCaseSlug      string
	Body             string
	SourceRequestID  string
	SourceResponseID string
	CreatedAt        time.Time
}

// AcceptanceFeaturePaths describes the canonical feature artifact paths.
type AcceptanceFeaturePaths struct {
	ScopeID      string
	RelativePath string
	Path         string
}

// AcceptanceFeatureLayout executes the corresponding package operation.
func AcceptanceFeatureLayout(projectDir string, scope QuizScope) (AcceptanceFeaturePaths, error) {
	err := ValidateQuizScope(scope)
	if err != nil {
		return AcceptanceFeaturePaths{}, err
	}

	scopeID, err := quizScopeID(scope)
	if err != nil {
		return AcceptanceFeaturePaths{}, err
	}

	relativePath := UserRelativePath(AcceptanceFeatureDirectory, scopeID+".feature")

	return AcceptanceFeaturePaths{
		ScopeID:      scopeID,
		RelativePath: relativePath,
		Path:         filepath.Join(projectDir, filepath.FromSlash(relativePath)),
	}, nil
}

// ValidateAcceptanceFeatureArtifact validates the corresponding package boundary value.
func ValidateAcceptanceFeatureArtifact(feature AcceptanceFeatureArtifact) error {
	err := ValidateQuizScope(feature.Scope)
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
