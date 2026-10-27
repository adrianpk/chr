package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	modelcore "forge.adrianpk.com/chr/internal/core/model"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/harness"
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
	Session       authoringcore.Session
	Now           time.Time
	AllowReplace  bool
	DryRun        bool
	QuestionCount int
	Provider      AuthoringQuizQuestionProvider
}

// ContextualQuizGenerationPlan describes a package boundary value.
type ContextualQuizGenerationPlan struct {
	Session                authoringcore.Session
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
	Target          authoringcore.Target
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

// HarnessAuthoringQuizQuestionProvider describes a package boundary value.
type HarnessAuthoringQuizQuestionProvider struct {
	Program string
	Runner  harness.AgentRunner
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

	request, err := NewAuthoringQuizGenerationRequest(ctx, opts.ProjectDir, opts.Session, opts.QuestionCount, opts.Now)
	if err != nil {
		return ContextualQuizGenerationPlan{}, err
	}

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

// NewAuthoringQuizGenerationRequest constructs the corresponding package service or value.
func NewAuthoringQuizGenerationRequest(
	ctx context.Context,
	projectDir string,
	session authoringcore.Session,
	questionCount int,
	now time.Time,
) (AuthoringQuizGenerationRequest, error) {
	sections, err := LoadAuthoringQuizContextSections(ctx, projectDir, session)
	if err != nil {
		return AuthoringQuizGenerationRequest{}, err
	}

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
	}, nil
}

// LoadAuthoringQuizContextSections loads the corresponding package data.
func LoadAuthoringQuizContextSections(
	ctx context.Context,
	projectDir string,
	session authoringcore.Session,
) ([]AuthoringQuizContextSection, error) {
	projectDir, err := cleanAbsProjectDir(projectDir)
	if err != nil {
		return nil, err
	}

	sections := []AuthoringQuizContextSection{
		{
			Name:      "project",
			Layer:     "project",
			Authority: "project_directory",
			Content:   "project_dir: " + projectDir + "\nproject_name: " + filepath.Base(projectDir),
		},
		{
			Name:      "session",
			Layer:     "authoring",
			Authority: "authoring_session",
			Content:   authoringQuizSessionContext(session),
		},
	}

	sections = append(sections, authoringQuizOptionalFileSection(projectDir, "agent_protocol", "protocols", "AGENTS.md"))
	sections = append(sections, authoringQuizWorkspaceGraphSection(ctx, projectDir))
	sections = append(sections, authoringQuizPathInventorySection(projectDir, "workspace_artifacts", "workspace", workspacecore.DirectoryName))
	sections = append(sections, authoringQuizPathInventorySection(projectDir, "docs", "docs", "docs"))
	sections = append(sections, authoringQuizTreeSection(projectDir))
	sections = append(sections, authoringQuizGoInventorySection(projectDir))

	return sections, nil
}

// RenderContextualQuizMarkdown renders the corresponding package data.
func RenderContextualQuizMarkdown(
	session authoringcore.Session,
	artifact authoringcore.ArtifactRef,
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
		output.WriteString("Completion recommended: " + markdownBool(question.CompletionRecommended) + "\n\n")
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

// GenerateAuthoringQuizQuestions generates the requested artifact.
func (p HarnessAuthoringQuizQuestionProvider) GenerateAuthoringQuizQuestions(ctx context.Context, request AuthoringQuizGenerationRequest) (AuthoringQuizGenerationResult, error) {
	err := ValidateAuthoringQuizGenerationRequest(request)
	if err != nil {
		return AuthoringQuizGenerationResult{}, err
	}

	program := strings.TrimSpace(p.Program)
	if _, _, bound := HarnessGenerationSelection(ctx); bound {
		program = ""
	}

	payload, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return AuthoringQuizGenerationResult{}, fmt.Errorf("marshal authoring quiz request: %w", err)
	}

	prompt := authoringQuizHarnessPrompt(string(payload))

	run, err := RunAgent(ctx, p.Runner, harness.AgentRunRequest{
		ID:                 request.ID,
		Purpose:            request.Action,
		ProjectDir:         request.ProjectDir,
		Prompt:             prompt,
		OutputName:         "output.json",
		Program:            program,
		ReadOnlyFiles:      true,
		DeferResponseTrace: true,
	})
	if err != nil {
		return AuthoringQuizGenerationResult{}, fmt.Errorf("invoke authoring quiz provider: %w", err)
	}

	result, err := ParseAuthoringQuizGenerationResult(run.Output)
	result.Provider = run.Provider
	result.Strategy = AuthoringQuizGenerationStrategyDeep

	result.RawText = string(run.Output)
	if err == nil {
		err = ValidateAuthoringQuizGenerationResult(result)
	}

	err = ValidateHarnessGenerationOutput(ctx, harness.AgentRunRequest{ID: request.ID, Purpose: request.Action, ProjectDir: request.ProjectDir}, err)
	if err != nil {
		recordAgentRunResponseEvent(ctx, harness.AgentRunRequest{
			ID:         request.ID,
			Purpose:    request.Action,
			ProjectDir: request.ProjectDir,
		}, run, "parse_error", err)

		return AuthoringQuizGenerationResult{}, err
	}

	recordAgentRunResponseEvent(ctx, harness.AgentRunRequest{
		ID:         request.ID,
		Purpose:    request.Action,
		ProjectDir: request.ProjectDir,
	}, run, "ready", nil)

	return result, nil
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

	err := authoringcore.ValidateSession(opts.Session)
	if err != nil {
		return err
	}

	if opts.Provider == nil {
		return fmt.Errorf("authoring quiz question provider is required")
	}

	if opts.Now.IsZero() {
		return fmt.Errorf("contextual quiz generation time is required")
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
		Kind:         authoringcore.ArtifactKindQuiz,
		Role:         authoringcore.ArtifactRoleOutput,
		Topic:        opts.Session.Topic,
		Description:  "Context-aware questionnaire generated for " + string(opts.Session.Target) + " authoring.",
		AllowReplace: opts.AllowReplace,
	}
}

func appendContextualQuizOutput(
	session authoringcore.Session,
	artifact authoringcore.ArtifactRef,
	recordedAt time.Time,
) authoringcore.Session {
	artifact.Role = authoringcore.ArtifactRoleOutput
	session.Outputs = append(session.Outputs, artifact)
	session.Provenance = append(session.Provenance, authoringcore.ProvenanceLink{
		Relation:   authoringcore.ProvenanceRelationGenerated,
		Source:     authoringQuizSourceArtifact(session),
		Target:     artifact,
		Decision:   "Generate a context-aware questionnaire for the authoring target.",
		RecordedAt: recordedAt.UTC(),
	})
	session.UpdatedAt = recordedAt.UTC()

	return session
}

func authoringQuizSourceArtifact(session authoringcore.Session) authoringcore.ArtifactRef {
	if len(session.Inputs) > 0 {
		return session.Inputs[0]
	}

	return authoringcore.ArtifactRef{
		ID:          "session:" + session.ID,
		Kind:        authoringcore.ArtifactKindSession,
		Role:        authoringcore.ArtifactRoleInput,
		Path:        filepath.ToSlash(filepath.Join(workspacecore.DirectoryName, filepath.FromSlash(AuthoringSessionsDirectory), session.ID, "session.md")),
		Description: "Authoring session metadata.",
	}
}

func authoringQuizSessionContext(session authoringcore.Session) string {
	lines := []string{
		"session_id: " + session.ID,
		"target: " + string(session.Target),
		"topic: " + session.Topic,
		"description: " + session.Description,
	}

	return strings.Join(lines, "\n")
}

func authoringQuizOptionalFileSection(projectDir, name, layer, relPath string) AuthoringQuizContextSection {
	fullPath := filepath.Join(projectDir, filepath.FromSlash(relPath))

	content, err := os.ReadFile(fullPath)
	if err != nil {
		content = []byte("missing: " + relPath)
	}

	return AuthoringQuizContextSection{
		Name:      name,
		Layer:     layer,
		Authority: relPath,
		Content:   string(content),
	}
}

func authoringQuizWorkspaceGraphSection(ctx context.Context, projectDir string) AuthoringQuizContextSection {
	graph, err := NewWorkspaceGraphBuilder(projectDir).Build(ctx)
	if err != nil {
		return AuthoringQuizContextSection{
			Name:      "workspace_graph_context",
			Layer:     "workspace",
			Authority: "WorkspaceGraph",
			Content:   "workspace graph context unavailable: " + err.Error(),
		}
	}

	assembly, err := AssembleWorkspaceContext(NewProjectWideWorkspaceContextRequest(projectDir, graph))
	if err != nil {
		return AuthoringQuizContextSection{
			Name:      "workspace_graph_context",
			Layer:     "workspace",
			Authority: "WorkspaceGraph",
			Content:   "workspace graph context unavailable: " + err.Error(),
		}
	}

	return AuthoringQuizContextSection{
		Name:      "workspace_graph_context",
		Layer:     "workspace",
		Authority: "WorkspaceGraph",
		Content:   RenderWorkspaceContextAssembly(assembly),
	}
}

func authoringQuizPathInventorySection(projectDir, name, layer, relPath string) AuthoringQuizContextSection {
	root := filepath.Join(projectDir, filepath.FromSlash(relPath))
	paths := authoringQuizRelativePaths(projectDir, root, authoringQuizMarkdownPath)

	return AuthoringQuizContextSection{
		Name:      name,
		Layer:     layer,
		Authority: relPath,
		Content:   strings.Join(paths, "\n"),
	}
}

func authoringQuizTreeSection(projectDir string) AuthoringQuizContextSection {
	paths := authoringQuizRelativePaths(projectDir, projectDir, authoringQuizTreePath)

	return AuthoringQuizContextSection{
		Name:      "tree",
		Layer:     "tree",
		Authority: "project_tree",
		Content:   strings.Join(paths, "\n"),
	}
}

func authoringQuizGoInventorySection(projectDir string) AuthoringQuizContextSection {
	paths := authoringQuizRelativePaths(projectDir, projectDir, authoringQuizGoPath)

	return AuthoringQuizContextSection{
		Name:      "go_inventory",
		Layer:     "code",
		Authority: "go_files",
		Content:   strings.Join(paths, "\n"),
	}
}

func authoringQuizRelativePaths(projectDir, root string, include func(string) bool) []string {
	paths := []string{}

	walkErr := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() && authoringQuizSkipDir(entry.Name()) {
			return filepath.SkipDir
		}

		if entry.IsDir() || !include(path) {
			return nil
		}

		relPath, relErr := filepath.Rel(projectDir, path)
		if relErr != nil {
			return relErr
		}

		paths = append(paths, filepath.ToSlash(relPath))
		if len(paths) >= 200 {
			return filepath.SkipAll
		}

		return nil
	})
	if walkErr != nil {
		return []string{}
	}

	sort.Strings(paths)

	return paths
}

func authoringQuizMarkdownPath(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".md")
}

func authoringQuizGoPath(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".go") ||
		strings.EqualFold(filepath.Base(path), "go.mod")
}

func authoringQuizTreePath(path string) bool {
	name := filepath.Base(path)

	return name != "" && !strings.HasPrefix(name, ".")
}

func authoringQuizSkipDir(name string) bool {
	switch name {
	case ".git", "vendor", "node_modules", "runs":
		return true
	default:
		return false
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

func authoringQuizHarnessPrompt(payload string) string {
	return strings.TrimSpace(`
You generate context-aware authoring quiz questions for chr.

Use the provided structured request and context sections.
Ask concrete, project-specific questions that help produce the requested target artifact.
Do not ask generic questionnaire questions.
For spec targets, cover problem, actors, workflows, acceptance criteria,
constraints, non-goals, validation evidence, impacted code/docs, and open risks
when the project context supports those areas.
Do not ask for secrets, credentials, tokens, passwords, private keys, or PII.
Return strict JSON only. Do not include Markdown or prose outside JSON.

The JSON response must have exactly one top-level key: "questions".
Each question object must contain:
- id
- question
- tags
- follow_up_depth
- completion_recommended

Request:
`) + "\n\n" + payload + "\n"
}
