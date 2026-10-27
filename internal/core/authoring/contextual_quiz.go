package authoring

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	modelcore "forge.adrianpk.com/chr/internal/core/model"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

const (
	// QuizGenerationStrategyDeep defines a package boundary constant.
	QuizGenerationStrategyDeep = "deep"

	// QuizActionGeneration defines a package boundary constant.
	QuizActionGeneration = "authoring_quiz_generation"
)

// ContextualQuizGenerationOptions describes a package boundary value.
type ContextualQuizGenerationOptions struct {
	ProjectDir    string
	Session       Session
	Now           time.Time
	AllowReplace  bool
	DryRun        bool
	QuestionCount int
	Provider      QuizQuestionProvider
	Context       QuizContextProvider
}

// ContextualQuizGenerationPlan describes a package boundary value.
type ContextualQuizGenerationPlan struct {
	Session                Session
	Layout                 ArtifactLayout
	Request                QuizGenerationRequest
	Result                 QuizGenerationResult
	Content                string
	WouldWrite             bool
	RequestedQuestionCount int
	GeneratedQuestionCount int
}

// QuizQuestionProvider defines a package boundary interface.
type QuizQuestionProvider interface {
	GenerateAuthoringQuizQuestions(
		ctx context.Context,
		request QuizGenerationRequest,
	) (QuizGenerationResult, error)
}

// QuizGenerationRequest describes a package boundary value.
type QuizGenerationRequest struct {
	ID              string
	Action          string
	Strategy        string
	ProjectDir      string
	SessionID       string
	Target          Target
	Topic           string
	QuestionCount   int
	ContextSections []QuizContextSection
	OutputContract  QuizOutputContract
	CreatedAt       time.Time
}

// QuizContextSection describes a package boundary value.
type QuizContextSection struct {
	Name      string
	Layer     string
	Authority string
	Content   string
}

// QuizOutputContract describes a package boundary value.
type QuizOutputContract struct {
	RequiredFields []string
}

// QuizGenerationResult describes a package boundary value.
type QuizGenerationResult struct {
	Provider  string
	Strategy  string
	Questions []QuizQuestion
	RawText   string
}

// QuizQuestion describes a package boundary value.
type QuizQuestion struct {
	ID                    string
	Text                  string
	Tags                  []string
	FollowUpDepth         int
	CompletionRecommended bool
}

// FakeAuthoringQuizQuestionProvider describes a package boundary value.
type FakeAuthoringQuizQuestionProvider struct {
	Result   QuizGenerationResult
	Requests []QuizGenerationRequest
	Err      error
}

// PlanContextualQuizGeneration executes the corresponding package operation.
func PlanContextualQuizGeneration(ctx context.Context, opts ContextualQuizGenerationOptions) (ContextualQuizGenerationPlan, error) {
	err := validateContextualQuizGenerationOptions(opts)
	if err != nil {
		return ContextualQuizGenerationPlan{}, err
	}

	layout, err := PlanArtifactLayout(contextualQuizLayoutOptions(opts))
	if err != nil {
		return ContextualQuizGenerationPlan{}, err
	}

	sections, err := opts.Context.Load(ctx, opts.ProjectDir, opts.Session)
	if err != nil {
		return ContextualQuizGenerationPlan{}, err
	}

	request := NewQuizGenerationRequest(opts.ProjectDir, opts.Session, opts.QuestionCount, opts.Now, sections)

	result, err := opts.Provider.GenerateAuthoringQuizQuestions(ctx, request)
	if err != nil {
		return ContextualQuizGenerationPlan{}, err
	}

	err = ValidateQuizGenerationResult(result)
	if err != nil {
		return ContextualQuizGenerationPlan{}, err
	}

	session := appendContextualQuizOutput(opts.Session, layout.Ref, opts.Now)
	content := RenderContextualQuizMarkdown(session, layout.Ref, request, result)

	return ContextualQuizGenerationPlan{
		Session:                session,
		Layout:                 layout,
		Request:                request,
		Result:                 result,
		Content:                content,
		WouldWrite:             !opts.DryRun,
		RequestedQuestionCount: request.QuestionCount,
		GeneratedQuestionCount: len(result.Questions),
	}, nil
}

// WriteContextualQuiz executes the corresponding package operation.
func writeContextualQuiz(plan ContextualQuizGenerationPlan) error {
	if plan.Layout.Collision.Exists && !plan.Layout.Collision.Allowed {
		return fmt.Errorf("authoring quiz artifact already exists: %s", plan.Layout.FullPath)
	}

	return writeAuthoringArtifactStep(plan.Layout, plan.Content, "contextual quiz")
}

// NewQuizGenerationRequest constructs the original request from observed context.
func NewQuizGenerationRequest(projectDir string, session Session, questionCount int, now time.Time, sections []QuizContextSection) QuizGenerationRequest {
	return QuizGenerationRequest{
		ID:              "quiz-request-" + session.ID,
		Action:          QuizActionGeneration,
		Strategy:        QuizGenerationStrategyDeep,
		ProjectDir:      projectDir,
		SessionID:       session.ID,
		Target:          session.Target,
		Topic:           session.Topic,
		QuestionCount:   normalizedAuthoringQuizQuestionCount(questionCount),
		ContextSections: sections,
		OutputContract: QuizOutputContract{
			RequiredFields: []string{
				"id",
				"question",
				"tags",
				"follow_up_depth",
				"completion_recommended",
			},
		},
		CreatedAt: now.UTC(),
	}
}

// RenderContextualQuizMarkdown renders the corresponding package data.
func RenderContextualQuizMarkdown(session Session, artifact ArtifactRef, request QuizGenerationRequest, result QuizGenerationResult) string {
	var output bytes.Buffer

	output.WriteString("---\n")
	output.WriteString("session_id: " + opsmetadata.Quote(session.ID) + "\n")
	output.WriteString("project_id: " + opsmetadata.Quote(session.ProjectID) + "\n")
	output.WriteString("target: " + opsmetadata.Quote(string(session.Target)) + "\n")
	output.WriteString("artifact_id: " + opsmetadata.Quote(artifact.ID) + "\n")
	output.WriteString("artifact_kind: " + opsmetadata.Quote(string(artifact.Kind)) + "\n")
	output.WriteString("topic: " + opsmetadata.Quote(session.Topic) + "\n")
	output.WriteString("generation_strategy: " + opsmetadata.Quote(result.Strategy) + "\n")
	output.WriteString("provider: " + opsmetadata.Quote(result.Provider) + "\n")
	output.WriteString("requested_questions: " + fmt.Sprintf("%d", request.QuestionCount) + "\n")
	output.WriteString("generated_questions: " + fmt.Sprintf("%d", len(result.Questions)) + "\n")
	output.WriteString("created_at: " + opsmetadata.Timestamp(session.UpdatedAt) + "\n")
	output.WriteString("---\n\n")
	output.WriteString("# Context-Aware Quiz\n\n")
	output.WriteString("Topic: " + session.Topic + "\n\n")
	output.WriteString("## Questions\n\n")

	for _, question := range result.Questions {
		output.WriteString("### " + question.ID + "\n\n")
		output.WriteString(question.Text + "\n\n")
		output.WriteString("Tags: " + strings.Join(question.Tags, ", ") + "\n\n")
		output.WriteString("Follow-up depth: " + fmt.Sprintf("%d", question.FollowUpDepth) + "\n\n")
		output.WriteString("Completion recommended: " + opsmetadata.YesNo(question.CompletionRecommended) + "\n\n")
		output.WriteString("Answer:\n\n")
	}

	output.WriteString("## Notes\n")

	return output.String()
}

// ValidateQuizGenerationResult validates the corresponding package boundary value.
func ValidateQuizGenerationResult(result QuizGenerationResult) error {
	if strings.TrimSpace(result.Provider) == "" {
		return fmt.Errorf("authoring quiz provider is required")
	}

	if strings.TrimSpace(result.Strategy) == "" {
		return fmt.Errorf("authoring quiz strategy is required")
	}

	if len(result.Questions) == 0 {
		return fmt.Errorf("authoring quiz generation returned no questions")
	}

	for _, question := range result.Questions {
		err := validateAuthoringQuizQuestion(question)
		if err != nil {
			return err
		}
	}

	return nil
}

// GenerateAuthoringQuizQuestions generates the requested artifact.
func (p *FakeAuthoringQuizQuestionProvider) GenerateAuthoringQuizQuestions(ctx context.Context, request QuizGenerationRequest) (QuizGenerationResult, error) {
	err := ctx.Err()
	if err != nil {
		return QuizGenerationResult{}, err
	}

	p.Requests = append(p.Requests, request)
	if p.Err != nil {
		return QuizGenerationResult{}, p.Err
	}

	return p.Result, nil
}

// ValidateQuizGenerationRequest validates the corresponding package boundary value.
func ValidateQuizGenerationRequest(request QuizGenerationRequest) error {
	required := []struct {
		name  string
		value string
	}{
		{name: "id", value: request.ID},
		{name: "action", value: request.Action},
		{name: "strategy", value: request.Strategy},
		{name: "project dir", value: request.ProjectDir},
		{name: "session id", value: request.SessionID},
		{name: "topic", value: request.Topic},
	}

	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("authoring quiz request %s is required", field.name)
		}
	}

	err := validateAuthoringQuizGenerationRequestShape(request)
	if err != nil {
		return err
	}

	return nil
}

func validateAuthoringQuizGenerationRequestShape(request QuizGenerationRequest) error {
	if !request.Target.Valid() {
		return fmt.Errorf("authoring quiz request target is invalid: %s", request.Target)
	}

	if request.QuestionCount <= 0 {
		return fmt.Errorf("authoring quiz request question count must be greater than zero")
	}

	if len(request.ContextSections) == 0 {
		return fmt.Errorf("authoring quiz request context sections are required")
	}

	if request.CreatedAt.IsZero() {
		return fmt.Errorf("authoring quiz request created time is required")
	}

	return nil
}

// ParseQuizGenerationResult parses the corresponding package data.
func ParseQuizGenerationResult(output []byte) (QuizGenerationResult, error) {
	var payload struct {
		Questions []struct {
			ID                    string   `json:"id"`
			Question              string   `json:"question"`
			Tags                  []string `json:"tags"`
			FollowUpDepth         int      `json:"follow_up_depth"`
			CompletionRecommended bool     `json:"completion_recommended"`
		} `json:"questions"`
	}

	err := json.Unmarshal(modelcore.ExtractJSONObject(output), &payload)
	if err != nil {
		return QuizGenerationResult{}, fmt.Errorf("parse authoring quiz provider response: %w", err)
	}

	questions := make([]QuizQuestion, 0, len(payload.Questions))
	for _, question := range payload.Questions {
		questions = append(questions, QuizQuestion{
			ID:                    question.ID,
			Text:                  question.Question,
			Tags:                  question.Tags,
			FollowUpDepth:         question.FollowUpDepth,
			CompletionRecommended: question.CompletionRecommended,
		})
	}

	return QuizGenerationResult{
		Strategy:  QuizGenerationStrategyDeep,
		Questions: questions,
		RawText:   string(output),
	}, nil
}

func validateContextualQuizGenerationOptions(opts ContextualQuizGenerationOptions) error {
	if opts.ProjectDir == "" {
		return fmt.Errorf("contextual quiz project dir is required")
	}

	err := ValidateSession(opts.Session)
	if err != nil {
		return err
	}

	if opts.Provider == nil {
		return fmt.Errorf("authoring quiz question provider is required")
	}

	if opts.Now.IsZero() {
		return fmt.Errorf("contextual quiz generation time is required")
	}

	if opts.Context == nil {
		return fmt.Errorf("authoring quiz context provider is required")
	}

	return nil
}

func validateAuthoringQuizQuestion(question QuizQuestion) error {
	if strings.TrimSpace(question.ID) == "" {
		return fmt.Errorf("authoring quiz question id is required")
	}

	if strings.TrimSpace(question.Text) == "" {
		return fmt.Errorf("authoring quiz question text is required")
	}

	if len(question.Tags) == 0 {
		return fmt.Errorf("authoring quiz question tags are required")
	}

	if question.FollowUpDepth < 0 {
		return fmt.Errorf("authoring quiz question follow-up depth must not be negative")
	}

	sensitiveValues := append([]string{question.Text}, question.Tags...)
	for _, value := range sensitiveValues {
		if authoringQuizContainsSensitivePrompt(value) {
			return fmt.Errorf("authoring quiz question must not ask for secrets, credentials, tokens, passwords, private keys, or PII")
		}
	}

	return nil
}

func contextualQuizLayoutOptions(opts ContextualQuizGenerationOptions) ArtifactLayoutOptions {
	return ArtifactLayoutOptions{
		ProjectDir:   opts.ProjectDir,
		SessionID:    opts.Session.ID,
		Kind:         ArtifactKindQuiz,
		Role:         ArtifactRoleOutput,
		Topic:        opts.Session.Topic,
		Description:  "Context-aware questionnaire generated for " + string(opts.Session.Target) + " authoring.",
		AllowReplace: opts.AllowReplace,
	}
}

func appendContextualQuizOutput(session Session, artifact ArtifactRef, recordedAt time.Time) Session {
	artifact.Role = ArtifactRoleOutput
	session.Outputs = append(session.Outputs, artifact)
	session.Provenance = append(session.Provenance, ProvenanceLink{
		Relation:   ProvenanceRelationGenerated,
		Source:     authoringQuizSourceArtifact(session),
		Target:     artifact,
		Decision:   "Generate a context-aware questionnaire for the authoring target.",
		RecordedAt: recordedAt.UTC(),
	})
	session.UpdatedAt = recordedAt.UTC()

	return session
}

func authoringQuizSourceArtifact(session Session) ArtifactRef {
	if len(session.Inputs) > 0 {
		return session.Inputs[0]
	}

	return ArtifactRef{
		ID:          "session:" + session.ID,
		Kind:        ArtifactKindSession,
		Role:        ArtifactRoleInput,
		Path:        filepath.ToSlash(filepath.Join(workspacecore.DirectoryName, filepath.FromSlash(SessionsDirectory), session.ID, "session.md")),
		Description: "Authoring session metadata.",
	}
}

func normalizedAuthoringQuizQuestionCount(count int) int {
	if count <= 0 {
		return 8
	}

	return count
}

func authoringQuizContainsSensitivePrompt(value string) bool {
	normalized := strings.ToLower(value)
	sensitive := []string{
		"api key",
		"apikey",
		"credential",
		"password",
		"private key",
		"secret",
		"token",
		"ssn",
	}

	for _, term := range sensitive {
		if strings.Contains(normalized, term) {
			return true
		}
	}

	return false
}

// QuizContextProvider supplies the existing native workspace/index observations.
// Authoring retains request shape, generation validation and artifact decisions.
type QuizContextProvider interface {
	Load(context.Context, string, Session) ([]QuizContextSection, error)
}
