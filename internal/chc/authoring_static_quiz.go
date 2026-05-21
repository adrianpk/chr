package chc

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type StaticQuizGenerationOptions struct {
	ProjectDir    string
	Session       AuthoringSession
	Now           time.Time
	AllowReplace  bool
	DryRun        bool
	QuestionCount int
}

type StaticQuizGenerationPlan struct {
	Session    AuthoringSession
	Layout     AuthoringArtifactLayout
	Questions  []StaticQuizQuestion
	Content    string
	WouldWrite bool
}

type StaticQuizQuestion struct {
	ID   string
	Text string
}

func PlanStaticQuizGeneration(opts StaticQuizGenerationOptions) (StaticQuizGenerationPlan, error) {
	err := validateStaticQuizGenerationOptions(opts)
	if err != nil {
		return StaticQuizGenerationPlan{}, err
	}

	layout, err := PlanAuthoringArtifactLayout(staticQuizLayoutOptions(opts))
	if err != nil {
		return StaticQuizGenerationPlan{}, err
	}

	questions := staticQuizQuestions(opts.Session.Target, opts.QuestionCount)
	if len(questions) == 0 {
		return StaticQuizGenerationPlan{}, fmt.Errorf("static quiz target has no question template: %s", opts.Session.Target)
	}

	session := appendStaticQuizOutput(opts.Session, layout.Ref, opts.Now)
	content := RenderStaticQuizMarkdown(session, layout.Ref, questions)

	return StaticQuizGenerationPlan{
		Session:    session,
		Layout:     layout,
		Questions:  questions,
		Content:    content,
		WouldWrite: !opts.DryRun,
	}, nil
}

func WriteStaticQuiz(plan StaticQuizGenerationPlan) error {
	if plan.Layout.Collision.Exists && !plan.Layout.Collision.Allowed {
		return fmt.Errorf("authoring quiz artifact already exists: %s", plan.Layout.FullPath)
	}

	err := os.MkdirAll(plan.Layout.Dir, 0o755)
	if err != nil {
		return fmt.Errorf("create static quiz directory: %w", err)
	}

	err = os.WriteFile(plan.Layout.FullPath, []byte(plan.Content), 0o644)
	if err != nil {
		return fmt.Errorf("write static quiz artifact: %w", err)
	}

	return nil
}

func RenderStaticQuizMarkdown(
	session AuthoringSession,
	artifact AuthoringArtifactRef,
	questions []StaticQuizQuestion,
) string {
	var output bytes.Buffer

	output.WriteString("---\n")
	output.WriteString("session_id: " + yamlString(session.ID) + "\n")
	output.WriteString("project_id: " + yamlString(session.ProjectID) + "\n")
	output.WriteString("target: " + yamlString(string(session.Target)) + "\n")
	output.WriteString("artifact_id: " + yamlString(artifact.ID) + "\n")
	output.WriteString("artifact_kind: " + yamlString(string(artifact.Kind)) + "\n")
	output.WriteString("topic: " + yamlString(session.Topic) + "\n")
	output.WriteString("created_at: " + yamlTimestamp(session.UpdatedAt) + "\n")
	output.WriteString("---\n\n")
	output.WriteString("# Static Quiz\n\n")
	output.WriteString("Topic: " + session.Topic + "\n\n")
	output.WriteString("## Questions\n\n")

	for _, question := range questions {
		output.WriteString("### " + question.ID + "\n\n")
		output.WriteString(question.Text + "\n\n")
		output.WriteString("Answer:\n\n")
	}

	output.WriteString("## Notes\n")

	return output.String()
}

func validateStaticQuizGenerationOptions(opts StaticQuizGenerationOptions) error {
	if opts.ProjectDir == "" {
		return fmt.Errorf("static quiz project dir is required")
	}

	err := ValidateAuthoringSession(opts.Session)
	if err != nil {
		return err
	}

	if opts.Session.Mode != AuthoringSessionModeStatic {
		return fmt.Errorf("static quiz generation requires static authoring mode")
	}

	if opts.Now.IsZero() {
		return fmt.Errorf("static quiz generation time is required")
	}

	return nil
}

func staticQuizLayoutOptions(opts StaticQuizGenerationOptions) AuthoringArtifactLayoutOptions {
	return AuthoringArtifactLayoutOptions{
		ProjectDir:   opts.ProjectDir,
		SessionID:    opts.Session.ID,
		Kind:         AuthoringArtifactKindQuiz,
		Role:         AuthoringArtifactRoleOutput,
		Topic:        opts.Session.Topic,
		Description:  "Static questionnaire generated for " + string(opts.Session.Target) + " authoring.",
		AllowReplace: opts.AllowReplace,
	}
}

func appendStaticQuizOutput(
	session AuthoringSession,
	artifact AuthoringArtifactRef,
	recordedAt time.Time,
) AuthoringSession {
	artifact.Role = AuthoringArtifactRoleOutput
	session.Outputs = append(session.Outputs, artifact)
	session.Provenance = append(session.Provenance, AuthoringProvenanceLink{
		Relation:   AuthoringProvenanceRelationGenerated,
		Source:     staticQuizSourceArtifact(session),
		Target:     artifact,
		Decision:   "Generate a static questionnaire for the authoring target.",
		RecordedAt: recordedAt.UTC(),
	})
	session.UpdatedAt = recordedAt.UTC()

	return session
}

func staticQuizSourceArtifact(session AuthoringSession) AuthoringArtifactRef {
	if len(session.Inputs) > 0 {
		return session.Inputs[0]
	}

	return AuthoringArtifactRef{
		ID:          "session:" + session.ID,
		Kind:        AuthoringArtifactKindSession,
		Role:        AuthoringArtifactRoleInput,
		Path:        filepath.ToSlash(filepath.Join(DirectoryName, filepath.FromSlash(AuthoringSessionsDirectory), session.ID, "session.md")),
		Description: "Authoring session metadata.",
	}
}

func staticQuizQuestions(target AuthoringTarget, limit int) []StaticQuizQuestion {
	questions := staticQuizQuestionTemplates()[target]
	if limit <= 0 || limit >= len(questions) {
		return questions
	}

	return questions[:limit]
}

func staticQuizQuestionTemplates() map[AuthoringTarget][]StaticQuizQuestion {
	common := []StaticQuizQuestion{
		{ID: "Q1", Text: "What problem or context should this artifact capture?"},
		{ID: "Q2", Text: "What outcome should be true when this work is complete?"},
		{ID: "Q3", Text: "Which existing files, commands, or documents are relevant?"},
		{ID: "Q4", Text: "What constraints, risks, or boundaries must be preserved?"},
		{ID: "Q5", Text: "What evidence should prove that the result is correct?"},
	}

	templates := map[AuthoringTarget][]StaticQuizQuestion{}
	for _, target := range staticQuizTemplateTargets() {
		templates[target] = targetStaticQuizQuestions(target, common)
	}

	return templates
}

func staticQuizTemplateTargets() []AuthoringTarget {
	return []AuthoringTarget{
		AuthoringTargetDump,
		AuthoringTargetQuiz,
		AuthoringTargetDraft,
		AuthoringTargetSpec,
		AuthoringTargetPlan,
		AuthoringTargetSliceSet,
		AuthoringTargetSlice,
		AuthoringTargetTracker,
		AuthoringTargetTicket,
	}
}

func targetStaticQuizQuestions(
	target AuthoringTarget,
	common []StaticQuizQuestion,
) []StaticQuizQuestion {
	targetQuestion := StaticQuizQuestion{
		ID:   "Q6",
		Text: "What target-specific details should shape the " + strings.ReplaceAll(string(target), "_", " ") + " artifact?",
	}

	questions := make([]StaticQuizQuestion, 0, len(common)+1)
	questions = append(questions, common...)
	questions = append(questions, targetQuestion)

	return questions
}
