package workspace

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// AssistedAnswerCurationOptions describes a package boundary value.
type AssistedAnswerCurationOptions struct {
	Session          AuthoringSession
	SessionPaths     AuthoringSessionPaths
	TranscriptEvents []AssistedTranscriptEvent
	Repository       AuthoringSessionRepository
	Adapter          AssistedModelAdapter
	Now              func() time.Time
}

// AssistedAnswerCurationResult describes a package boundary value.
type AssistedAnswerCurationResult struct {
	Session      AuthoringSession
	Request      AssistedModelRequest
	Response     AssistedModelResponse
	Summary      AssistedCuratedAnswerSummary
	ArtifactRef  AuthoringArtifactRef
	ArtifactPath string
}

// AssistedCuratedAnswerSummary describes a package boundary value.
type AssistedCuratedAnswerSummary struct {
	SessionID     string
	GeneratedAt   time.Time
	ModelSummary  string
	Answers       []AssistedCuratedAnswer
	Decisions     []string
	OpenQuestions []AssistedCuratedOpenQuestion
}

// AssistedCuratedAnswer describes a package boundary value.
type AssistedCuratedAnswer struct {
	QuestionID string
	Question   string
	Answer     string
	AnsweredAt time.Time
}

// AssistedCuratedOpenQuestion describes a package boundary value.
type AssistedCuratedOpenQuestion struct {
	QuestionID string
	Question   string
	TurnIndex  int
}

// CurateAssistedAnswers executes the corresponding package operation.
func CurateAssistedAnswers(
	ctx context.Context,
	opts AssistedAnswerCurationOptions,
) (AssistedAnswerCurationResult, error) {
	err := validateAssistedAnswerCurationOptions(opts)
	if err != nil {
		return AssistedAnswerCurationResult{}, err
	}

	now := assistedAnswerCurationNow(opts.Now)
	draft := BuildAssistedCuratedAnswerSummary(opts.Session.ID, opts.TranscriptEvents, now)
	request := assistedAnswerCurationRequest(opts.Session, opts.TranscriptEvents, draft, now)

	response, err := InvokeAssistedModel(ctx, opts.Adapter, request)
	if err != nil {
		return AssistedAnswerCurationResult{}, err
	}

	draft.ModelSummary = strings.TrimSpace(response.RawText)

	artifactRef, artifactPath, err := writeAssistedAnswerCurationArtifact(opts.Session, opts.SessionPaths, draft)
	if err != nil {
		return AssistedAnswerCurationResult{}, err
	}

	session, err := updateAssistedAnswerCurationSession(ctx, opts.Repository, opts.Session, artifactRef, now)
	if err != nil {
		return AssistedAnswerCurationResult{}, err
	}

	return AssistedAnswerCurationResult{
		Session:      session,
		Request:      request,
		Response:     response,
		Summary:      draft,
		ArtifactRef:  artifactRef,
		ArtifactPath: artifactPath,
	}, nil
}

// BuildAssistedCuratedAnswerSummary executes the corresponding package operation.
func BuildAssistedCuratedAnswerSummary(
	sessionID string,
	events []AssistedTranscriptEvent,
	generatedAt time.Time,
) AssistedCuratedAnswerSummary {
	questions := map[string]AssistedTranscriptEvent{}
	answers := make([]AssistedCuratedAnswer, 0)
	answered := map[string]struct{}{}
	decisions := []string{}

	for _, event := range events {
		switch event.Type {
		case AssistedTranscriptEventQuestion:
			questions[event.Question.ID] = event
		case AssistedTranscriptEventAnswer:
			question := questions[event.Answer.QuestionID]
			answers = append(answers, AssistedCuratedAnswer{
				QuestionID: event.Answer.QuestionID,
				Question:   question.Question.Prompt,
				Answer:     event.Answer.Text,
				AnsweredAt: event.Answer.AnsweredAt,
			})
			answered[event.Answer.QuestionID] = struct{}{}
		case AssistedTranscriptEventResult:
			decision := assistedAnswerCurationDecision(event)
			if decision != "" {
				decisions = append(decisions, decision)
			}
		}
	}

	return AssistedCuratedAnswerSummary{
		SessionID:     sessionID,
		GeneratedAt:   generatedAt.UTC(),
		Answers:       answers,
		Decisions:     decisions,
		OpenQuestions: assistedAnswerCurationOpenQuestions(questions, answered),
	}
}

// RenderAssistedCuratedAnswerSummary renders the corresponding package data.
func RenderAssistedCuratedAnswerSummary(summary AssistedCuratedAnswerSummary) string {
	var output bytes.Buffer

	output.WriteString("---\n")
	output.WriteString("session_id: " + yamlString(summary.SessionID) + "\n")
	output.WriteString("generated_at: " + yamlTimestamp(summary.GeneratedAt) + "\n")
	output.WriteString("---\n\n")
	output.WriteString("# Curated Assisted Answers\n\n")
	output.WriteString("## Model Summary\n\n")
	output.WriteString(emptyMarkdownValue(summary.ModelSummary, "No model summary was returned."))
	output.WriteString("\n\n")
	writeAssistedCuratedAnswers(&output, summary.Answers)
	writeAssistedCuratedDecisions(&output, summary.Decisions)
	writeAssistedCuratedOpenQuestions(&output, summary.OpenQuestions)

	return output.String()
}

func validateAssistedAnswerCurationOptions(opts AssistedAnswerCurationOptions) error {
	err := ValidateAuthoringSession(opts.Session)
	if err != nil {
		return err
	}

	if opts.SessionPaths.SessionDir == "" {
		return fmt.Errorf("assisted answer curation session directory is required")
	}

	if opts.Repository == nil {
		return fmt.Errorf("assisted answer curation repository is required")
	}

	if opts.Adapter == nil {
		return fmt.Errorf("assisted answer curation model adapter is required")
	}

	for _, event := range opts.TranscriptEvents {
		if event.SessionID != opts.Session.ID {
			return fmt.Errorf("assisted answer curation transcript event session id does not match session")
		}
	}

	return nil
}

func assistedAnswerCurationRequest(
	session AuthoringSession,
	events []AssistedTranscriptEvent,
	draft AssistedCuratedAnswerSummary,
	now time.Time,
) AssistedModelRequest {
	return AssistedModelRequest{
		ID:        "curation-" + session.ID,
		SessionID: session.ID,
		TurnID:    "curation",
		Intent:    AssistedAuthoringIntentCurate,
		Target:    session.Target,
		Context:   RenderAssistedCuratedAnswerSummary(draft),
		Prompt:    assistedAnswerCurationPrompt(session, events),
		CreatedAt: now,
	}
}

func assistedAnswerCurationPrompt(session AuthoringSession, events []AssistedTranscriptEvent) string {
	return fmt.Sprintf(
		"Curate the assisted authoring transcript for session %s into a concise reusable summary. Transcript events: %d.",
		session.ID,
		len(events),
	)
}

func writeAssistedAnswerCurationArtifact(
	session AuthoringSession,
	paths AuthoringSessionPaths,
	summary AssistedCuratedAnswerSummary,
) (AuthoringArtifactRef, string, error) {
	artifactPath := filepath.Join(paths.SessionDir, "curated-summary.md")

	_, err := os.Stat(artifactPath)
	if err == nil {
		return AuthoringArtifactRef{}, "", fmt.Errorf("assisted answer curation artifact already exists: %s", artifactPath)
	}

	if err != nil && !os.IsNotExist(err) {
		return AuthoringArtifactRef{}, "", fmt.Errorf("stat assisted answer curation artifact: %w", err)
	}

	err = os.WriteFile(artifactPath, []byte(RenderAssistedCuratedAnswerSummary(summary)), 0o644)
	if err != nil {
		return AuthoringArtifactRef{}, "", fmt.Errorf("write assisted answer curation artifact: %w", err)
	}

	ref := AuthoringArtifactRef{
		ID:          "curated-summary-" + session.ID,
		Kind:        AuthoringArtifactKindDraft,
		Role:        AuthoringArtifactRoleOutput,
		Path:        filepath.ToSlash(artifactPath),
		Description: "Curated assisted authoring answer summary.",
	}

	return ref, artifactPath, nil
}

func updateAssistedAnswerCurationSession(
	ctx context.Context,
	repository AuthoringSessionRepository,
	session AuthoringSession,
	ref AuthoringArtifactRef,
	now time.Time,
) (AuthoringSession, error) {
	session.Outputs = append(session.Outputs, ref)
	session.Provenance = append(session.Provenance, AuthoringProvenanceLink{
		Relation: AuthoringProvenanceRelationGenerated,
		Source: AuthoringArtifactRef{
			ID:          "transcript-" + session.ID,
			Kind:        AuthoringArtifactKindTranscript,
			Role:        AuthoringArtifactRoleInput,
			Path:        filepath.ToSlash(AuthoringSessionLayout("", session.ID).TranscriptPath),
			Description: "Assisted authoring transcript.",
		},
		Target:     ref,
		Decision:   "Curated assisted transcript answers into a reusable summary.",
		RecordedAt: now.UTC(),
	})
	session.UpdatedAt = now.UTC()

	_, err := repository.UpdateSession(ctx, session)
	if err != nil {
		return AuthoringSession{}, err
	}

	return session, nil
}

func assistedAnswerCurationDecision(event AssistedTranscriptEvent) string {
	if event.Result.Summary == "" {
		return ""
	}

	if event.Result.CompletionReason != "" {
		return fmt.Sprintf("%s: %s", event.Result.CompletionReason, event.Result.Summary)
	}

	if event.Result.BlockedReason != "" {
		return fmt.Sprintf("%s: %s", event.Result.BlockedReason, event.Result.Summary)
	}

	return event.Result.Summary
}

func assistedAnswerCurationOpenQuestions(
	questions map[string]AssistedTranscriptEvent,
	answered map[string]struct{},
) []AssistedCuratedOpenQuestion {
	openQuestions := make([]AssistedCuratedOpenQuestion, 0)

	for questionID, event := range questions {
		_, ok := answered[questionID]
		if ok {
			continue
		}

		openQuestions = append(openQuestions, AssistedCuratedOpenQuestion{
			QuestionID: questionID,
			Question:   event.Question.Prompt,
			TurnIndex:  event.TurnIndex,
		})
	}

	sort.Slice(openQuestions, func(left, right int) bool {
		return openQuestions[left].TurnIndex < openQuestions[right].TurnIndex
	})

	return openQuestions
}

func writeAssistedCuratedAnswers(output *bytes.Buffer, answers []AssistedCuratedAnswer) {
	output.WriteString("## Answers\n\n")

	if len(answers) == 0 {
		output.WriteString("No answers were captured.\n\n")

		return
	}

	for _, answer := range answers {
		output.WriteString("### ")
		output.WriteString(answer.QuestionID)
		output.WriteString("\n\n")
		output.WriteString("Question: ")
		output.WriteString(answer.Question)
		output.WriteString("\n\n")
		output.WriteString("Answer: ")
		output.WriteString(answer.Answer)
		output.WriteString("\n\n")
		output.WriteString("Answered at: ")
		output.WriteString(answer.AnsweredAt.UTC().Format(time.RFC3339))
		output.WriteString("\n\n")
	}
}

func writeAssistedCuratedDecisions(output *bytes.Buffer, decisions []string) {
	output.WriteString("## Decisions\n\n")

	if len(decisions) == 0 {
		output.WriteString("No explicit decisions were captured.\n\n")

		return
	}

	for _, decision := range decisions {
		output.WriteString("- ")
		output.WriteString(decision)
		output.WriteString("\n")
	}

	output.WriteString("\n")
}

func writeAssistedCuratedOpenQuestions(output *bytes.Buffer, questions []AssistedCuratedOpenQuestion) {
	output.WriteString("## Open Questions\n\n")

	if len(questions) == 0 {
		output.WriteString("No open questions remain.\n")

		return
	}

	for _, question := range questions {
		output.WriteString("- ")
		output.WriteString(question.QuestionID)
		output.WriteString(": ")
		output.WriteString(question.Question)
		output.WriteString("\n")
	}
}

func emptyMarkdownValue(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}

	return value
}

func assistedAnswerCurationNow(now func() time.Time) time.Time {
	if now == nil {
		return time.Now().UTC()
	}

	return now().UTC()
}
