package core

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"
)

// QuizToDraftPromotionOptions describes a package boundary value.
type QuizToDraftPromotionOptions struct {
	ProjectDir   string
	Session      AuthoringSession
	QuizPath     string
	Now          time.Time
	AllowReplace bool
	DryRun       bool
}

// QuizToDraftPromotionPlan describes a package boundary value.
type QuizToDraftPromotionPlan struct {
	Session    AuthoringSession
	Layout     AuthoringArtifactLayout
	Source     AuthoringArtifactRef
	Content    string
	WouldWrite bool
}

// PlanQuizToDraftPromotion executes the corresponding package operation.
func PlanQuizToDraftPromotion(opts QuizToDraftPromotionOptions) (QuizToDraftPromotionPlan, error) {
	err := validateQuizToDraftPromotionOptions(opts)
	if err != nil {
		return QuizToDraftPromotionPlan{}, err
	}

	layout, err := PlanAuthoringArtifactLayout(quizToDraftLayoutOptions(opts))
	if err != nil {
		return QuizToDraftPromotionPlan{}, err
	}

	source := quizPromotionSourceArtifact(opts)
	session := appendQuizToDraftOutput(opts.Session, source, layout.Ref, opts.Now)

	content, err := RenderDraftFromQuiz(session, source, layout.Ref, opts.QuizPath)
	if err != nil {
		return QuizToDraftPromotionPlan{}, err
	}

	return QuizToDraftPromotionPlan{
		Session:    session,
		Layout:     layout,
		Source:     source,
		Content:    content,
		WouldWrite: !opts.DryRun,
	}, nil
}

// WriteQuizToDraftPromotion executes the corresponding package operation.
func WriteQuizToDraftPromotion(plan QuizToDraftPromotionPlan) error {
	if plan.Layout.Collision.Exists && !plan.Layout.Collision.Allowed {
		return fmt.Errorf("authoring draft artifact already exists: %s", plan.Layout.FullPath)
	}

	err := os.MkdirAll(plan.Layout.Dir, 0o755)
	if err != nil {
		return fmt.Errorf("create authoring draft directory: %w", err)
	}

	err = os.WriteFile(plan.Layout.FullPath, []byte(plan.Content), 0o644)
	if err != nil {
		return fmt.Errorf("write authoring draft artifact: %w", err)
	}

	return nil
}

// RenderDraftFromQuiz renders the corresponding package data.
func RenderDraftFromQuiz(
	session AuthoringSession,
	source AuthoringArtifactRef,
	target AuthoringArtifactRef,
	quizPath string,
) (string, error) {
	quizContent, err := os.ReadFile(quizPath)
	if err != nil {
		return "", fmt.Errorf("read quiz artifact: %w", err)
	}

	var output bytes.Buffer

	output.WriteString("---\n")
	output.WriteString("session_id: " + yamlString(session.ID) + "\n")
	output.WriteString("project_id: " + yamlString(session.ProjectID) + "\n")
	output.WriteString("source_artifact_id: " + yamlString(source.ID) + "\n")
	output.WriteString("source_path: " + yamlString(source.Path) + "\n")
	output.WriteString("artifact_id: " + yamlString(target.ID) + "\n")
	output.WriteString("artifact_kind: " + yamlString(string(target.Kind)) + "\n")
	output.WriteString("topic: " + yamlString(session.Topic) + "\n")
	output.WriteString("created_at: " + yamlTimestamp(session.UpdatedAt) + "\n")
	output.WriteString("---\n\n")
	output.WriteString("# Draft\n\n")
	output.WriteString("Topic: " + session.Topic + "\n\n")
	output.WriteString("## Source Quiz\n\n")
	output.WriteString(source.Path + "\n\n")
	output.WriteString("## Curated Answers\n\n")
	output.WriteString(strings.TrimSpace(string(quizContent)) + "\n\n")
	output.WriteString("## Draft Notes\n")

	return output.String(), nil
}

func validateQuizToDraftPromotionOptions(opts QuizToDraftPromotionOptions) error {
	if opts.ProjectDir == "" {
		return fmt.Errorf("quiz to draft project dir is required")
	}

	err := ValidateAuthoringSession(opts.Session)
	if err != nil {
		return err
	}

	if opts.QuizPath == "" {
		return fmt.Errorf("quiz path is required")
	}

	if opts.Now.IsZero() {
		return fmt.Errorf("quiz to draft promotion time is required")
	}

	return nil
}

func quizToDraftLayoutOptions(opts QuizToDraftPromotionOptions) AuthoringArtifactLayoutOptions {
	return AuthoringArtifactLayoutOptions{
		ProjectDir:   opts.ProjectDir,
		SessionID:    opts.Session.ID,
		Kind:         AuthoringArtifactKindDraft,
		Role:         AuthoringArtifactRoleOutput,
		Topic:        opts.Session.Topic,
		Description:  "Draft promoted from a quiz.",
		AllowReplace: opts.AllowReplace,
	}
}

func appendQuizToDraftOutput(
	session AuthoringSession,
	source AuthoringArtifactRef,
	target AuthoringArtifactRef,
	recordedAt time.Time,
) AuthoringSession {
	target.Role = AuthoringArtifactRoleOutput
	session.Outputs = append(session.Outputs, target)
	session.Provenance = append(session.Provenance, AuthoringProvenanceLink{
		Relation:   AuthoringProvenanceRelationPromoted,
		Source:     source,
		Target:     target,
		Decision:   "Promote quiz answers into an editable draft.",
		RecordedAt: recordedAt.UTC(),
	})
	session.UpdatedAt = recordedAt.UTC()

	return session
}

func quizPromotionSourceArtifact(opts QuizToDraftPromotionOptions) AuthoringArtifactRef {
	for _, output := range opts.Session.Outputs {
		if output.Kind == AuthoringArtifactKindQuiz {
			return output
		}
	}

	return AuthoringArtifactRef{
		ID:          "quiz:" + opts.Session.ID,
		Kind:        AuthoringArtifactKindQuiz,
		Role:        AuthoringArtifactRoleInput,
		Path:        opts.QuizPath,
		Description: "Quiz artifact selected for promotion.",
	}
}
