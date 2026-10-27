package authoring

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
)

// QuizToDraftPromotionOptions describes a package boundary value.
type QuizToDraftPromotionOptions struct {
	ProjectDir   string
	Session      Session
	QuizPath     string
	Now          time.Time
	AllowReplace bool
	DryRun       bool
}

// QuizToDraftPromotionPlan describes a package boundary value.
type QuizToDraftPromotionPlan struct {
	Session    Session
	Layout     AuthoringArtifactLayout
	Source     ArtifactRef
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
func writeQuizToDraftPromotion(plan QuizToDraftPromotionPlan) error {
	if plan.Layout.Collision.Exists && !plan.Layout.Collision.Allowed {
		return fmt.Errorf("authoring draft artifact already exists: %s", plan.Layout.FullPath)
	}

	return writeAuthoringArtifactStep(plan.Layout, plan.Content, "authoring draft")
}

// RenderDraftFromQuiz renders the corresponding package data.
func RenderDraftFromQuiz(
	session Session,
	source ArtifactRef,
	target ArtifactRef,
	quizPath string,
) (string, error) {
	quizContent, err := os.ReadFile(quizPath)
	if err != nil {
		return "", fmt.Errorf("read quiz artifact: %w", err)
	}

	var output bytes.Buffer

	output.WriteString("---\n")
	output.WriteString("session_id: " + opsmetadata.Quote(session.ID) + "\n")
	output.WriteString("project_id: " + opsmetadata.Quote(session.ProjectID) + "\n")
	output.WriteString("source_artifact_id: " + opsmetadata.Quote(source.ID) + "\n")
	output.WriteString("source_path: " + opsmetadata.Quote(source.Path) + "\n")
	output.WriteString("artifact_id: " + opsmetadata.Quote(target.ID) + "\n")
	output.WriteString("artifact_kind: " + opsmetadata.Quote(string(target.Kind)) + "\n")
	output.WriteString("topic: " + opsmetadata.Quote(session.Topic) + "\n")
	output.WriteString("created_at: " + opsmetadata.Timestamp(session.UpdatedAt) + "\n")
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

	err := ValidateSession(opts.Session)
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
		Kind:         ArtifactKindDraft,
		Role:         ArtifactRoleOutput,
		Topic:        opts.Session.Topic,
		Description:  "Draft promoted from a quiz.",
		AllowReplace: opts.AllowReplace,
	}
}

func appendQuizToDraftOutput(
	session Session,
	source ArtifactRef,
	target ArtifactRef,
	recordedAt time.Time,
) Session {
	target.Role = ArtifactRoleOutput
	session.Outputs = append(session.Outputs, target)
	session.Provenance = append(session.Provenance, ProvenanceLink{
		Relation:   ProvenanceRelationPromoted,
		Source:     source,
		Target:     target,
		Decision:   "Promote quiz answers into an editable draft.",
		RecordedAt: recordedAt.UTC(),
	})
	session.UpdatedAt = recordedAt.UTC()

	return session
}

func quizPromotionSourceArtifact(opts QuizToDraftPromotionOptions) ArtifactRef {
	for _, output := range opts.Session.Outputs {
		if output.Kind == ArtifactKindQuiz {
			return output
		}
	}

	return ArtifactRef{
		ID:          "quiz:" + opts.Session.ID,
		Kind:        ArtifactKindQuiz,
		Role:        ArtifactRoleInput,
		Path:        opts.QuizPath,
		Description: "Quiz artifact selected for promotion.",
	}
}
