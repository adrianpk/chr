package generation

import (
	"fmt"
	"strings"
	"time"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
)

// QuizDraftSpec describes a package boundary value.
type QuizDraftSpec struct {
	ExpectedRevision string `json:"-"`
	Scope            quizcore.Scope
	Title            string
	Body             string
	Feature          AcceptanceFeatureGeneration
	SourceQuizPath   string
	SourceRawPath    string
	SourceDumpPaths  []string
	SourceRequestID  string
	SourceResponseID string
	CreatedAt        time.Time
}

// ValidateQuizDraftSpec validates the corresponding package boundary value.
func ValidateQuizDraftSpec(spec QuizDraftSpec) error {
	err := ValidateQuizDraftSpecModelContent(spec)
	if err != nil {
		return err
	}

	if strings.TrimSpace(spec.SourceQuizPath) == "" {
		return fmt.Errorf("quiz draft spec source quiz path is required")
	}

	if strings.TrimSpace(spec.SourceRawPath) == "" {
		return fmt.Errorf("quiz draft spec source raw path is required")
	}

	if spec.CreatedAt.IsZero() {
		return fmt.Errorf("quiz draft spec created time is required")
	}

	if spec.Feature.HasContent() {
		_, err = ValidateAcceptanceFeatureGeneration(spec.Feature)
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateQuizDraftSpecModelContent validates the corresponding package boundary value.
func ValidateQuizDraftSpecModelContent(spec QuizDraftSpec) error {
	if !spec.HasContent() {
		return fmt.Errorf("quiz draft spec content is required")
	}

	err := quizcore.ValidateScope(spec.Scope)
	if err != nil {
		return err
	}

	if spec.CreatedAt.IsZero() {
		return fmt.Errorf("quiz draft spec created time is required")
	}

	if spec.Feature.HasContent() {
		_, err = ValidateAcceptanceFeatureGeneration(spec.Feature)
		if err != nil {
			return err
		}
	}

	return nil
}

// HasContent executes the corresponding package operation.
func (s QuizDraftSpec) HasContent() bool {
	return strings.TrimSpace(s.Title) != "" && strings.TrimSpace(s.Body) != ""
}
