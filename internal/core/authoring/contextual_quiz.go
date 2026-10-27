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
	// AuthoringQuizGenerationStrategyDeep defines a package boundary constant.
	AuthoringQuizGenerationStrategyDeep = "deep"

	// AuthoringQuizActionGeneration defines a package boundary constant.
	AuthoringQuizActionGeneration = "authoring_quiz_generation"
)

// ContextualQuizGenerationOptions describes a package boundary value.
type ContextualQuizGenerationOptions struct {
	ProjectDir    string
	Session       Session
	Now           time.Time
	AllowReplace  bool
	DryRun        bool
	QuestionCount int
	Provider      AuthoringQuizQuestionProvider
	Context       QuizContextProvider
}

// ContextualQuizGenerationPlan describes a package boundary value.
type ContextualQuizGenerationPlan struct {
	Session                Session
	Layout                 AuthoringArtifactLayout
	Request                AuthoringQuizGenerationRequest
	Result                 AuthoringQuizGenerationResult
	Content                string
	WouldWrite             bool
	RequestedQuestionCount int
	GeneratedQuestionCount int
}

// AuthoringQuizQuestionProvider defines a package boundary interface.
type AuthoringQuizQuestionProvider interface {
	GenerateAuthoringQuizQuestions(
		ctx context.Context,
		request AuthoringQuizGenerationRequest,
	) (AuthoringQuizGenerationResult, error)
}

// AuthoringQuizGenerationRequest describes a package boundary value.
type AuthoringQuizGenerationRequest struct {
	ID              string
	Action          string
	Strategy        string
	ProjectDir      string
	SessionID       string
	Target          Target
	Topic           string
	QuestionCount   int
	ContextSections []AuthoringQuizContextSection
	OutputContract  AuthoringQuizOutputContract
	CreatedAt       time.Time
}

// AuthoringQuizContextSection describes a package boundary value.
type AuthoringQuizContextSection struct {
	Name      string
	Layer     string
	Authority string
	Content   string
}

// AuthoringQuizOutputContract describes a package boundary value.
type AuthoringQuizOutputContract struct {
	RequiredFields []string
}

// AuthoringQuizGenerationResult describes a package boundary value.
type AuthoringQuizGenerationResult struct {
	Provider  string
	Strategy  string
	Questions []AuthoringQuizQuestion
	RawText   string
}

// AuthoringQuizQuestion describes a package boundary value.
type AuthoringQuizQuestion struct {
	ID                    string
	Text                  string
	Tags                  []string
	FollowUpDepth         int
	CompletionRecommended bool
}

// FakeAuthoringQuizQuestionProvider describes a package boundary value.
type FakeAuthoringQuizQuestionProvider struct {
	Result   AuthoringQuizGenerationResult
	Requests []AuthoringQuizGenerationRequest
	Err      error
}

// PlanContextualQuizGeneration executes the corresponding package operation.
func PlanContextualQuizGeneration(
	ctx context.Context,
	opts ContextualQuizGenerationOptions,
) (ContextualQuizGenerationPlan, error) {
	err := validateContextualQuizGenerationOptions(opts)
	if err != nil {
		return ContextualQuizGenerationPlan{}, err
	}

	layout, err := PlanAuthoringArtifactLayout(contextualQuizLayoutOptions(opts))
	if err != nil {
		return ContextualQuizGenerationPlan{}, err
	}

	sections, err := opts.Context.Load(ctx, opts.ProjectDir, opts.Session)
	if err != nil {
		return ContextualQuizGenerationPlan{}, err
	}

	request := NewAuthoringQuizGenerationRequest(opts.ProjectDir, opts.Session, opts.QuestionCount, opts.Now, sections)

	result, err := opts.Provider.GenerateAuthoringQuizQuestions(ctx, request)
	if err != nil {
		return ContextualQuizGenerationPlan{}, err
	}

	err = ValidateAuthoringQuizGenerationResult(result)
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

// NewAuthoringQuizGenerationRequest constructs the original request from observed context.
func NewAuthoringQuizGenerationRequest(projectDir string, session Session, questionCount int, now time.Time, sections []AuthoringQuizContextSection) AuthoringQuizGenerationRequest {
	return AuthoringQuizGenerationRequest{
		ID:              "quiz-request-" + session.ID,
		Action:          AuthoringQuizActionGeneration,
		Strategy:        AuthoringQuizGenerationStrategyDeep,
		ProjectDir:      projectDir,
		SessionID:       session.ID,
		Target:          session.Target,
		Topic:           session.Topic,
		QuestionCount:   normalizedAuthoringQuizQuestionCount(questionCount),
		ContextSections: sections,
		OutputContract: AuthoringQuizOutputContract{
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
func RenderContextualQuizMarkdown(
	session Session,
	artifact ArtifactRef,
	request AuthoringQuizGenerationRequest,
	result AuthoringQuizGenerationResult,
) string {
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

// ValidateAuthoringQuizGenerationResult validates the corresponding package boundary value.
func ValidateAuthoringQuizGenerationResult(result AuthoringQuizGenerationResult) error {
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
func (p *FakeAuthoringQuizQuestionProvider) GenerateAuthoringQuizQuestions(
	ctx context.Context,
	request AuthoringQuizGenerationRequest,
) (AuthoringQuizGenerationResult, error) {
	err := ctx.Err()
	if err != nil {
		return AuthoringQuizGenerationResult{}, err
	}

	p.Requests = append(p.Requests, request)
	if p.Err != nil {
		return AuthoringQuizGenerationResult{}, p.Err
	}

	return p.Result, nil
}

// ValidateAuthoringQuizGenerationRequest validates the corresponding package boundary value.
func ValidateAuthoringQuizGenerationRequest(request AuthoringQuizGenerationRequest) error {
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

func validateAuthoringQuizGenerationRequestShape(request AuthoringQuizGenerationRequest) error {
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

// ParseAuthoringQuizGenerationResult parses the corresponding package data.
func ParseAuthoringQuizGenerationResult(output []byte) (AuthoringQuizGenerationResult, error) {
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
		return AuthoringQuizGenerationResult{}, fmt.Errorf("parse authoring quiz provider response: %w", err)
	}

	questions := make([]AuthoringQuizQuestion, 0, len(payload.Questions))
	for _, question := range payload.Questions {
		questions = append(questions, AuthoringQuizQuestion{
			ID:                    question.ID,
			Text:                  question.Question,
			Tags:                  question.Tags,
			FollowUpDepth:         question.FollowUpDepth,
			CompletionRecommended: question.CompletionRecommended,
		})
	}

	return AuthoringQuizGenerationResult{
		Strategy:  AuthoringQuizGenerationStrategyDeep,
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

func validateAuthoringQuizQuestion(question AuthoringQuizQuestion) error {
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

func contextualQuizLayoutOptions(opts ContextualQuizGenerationOptions) AuthoringArtifactLayoutOptions {
	return AuthoringArtifactLayoutOptions{
		ProjectDir:   opts.ProjectDir,
		SessionID:    opts.Session.ID,
		Kind:         ArtifactKindQuiz,
		Role:         ArtifactRoleOutput,
		Topic:        opts.Session.Topic,
		Description:  "Context-aware questionnaire generated for " + string(opts.Session.Target) + " authoring.",
		AllowReplace: opts.AllowReplace,
	}
}

func appendContextualQuizOutput(
	session Session,
	artifact ArtifactRef,
	recordedAt time.Time,
) Session {
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
		Path:        filepath.ToSlash(filepath.Join(workspacecore.DirectoryName, filepath.FromSlash(AuthoringSessionsDirectory), session.ID, "session.md")),
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
	Load(context.Context, string, Session) ([]AuthoringQuizContextSection, error)
}
