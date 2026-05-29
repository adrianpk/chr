package chc

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
)

const (
	AuthoringQuizGenerationStrategyDeep = "deep"

	AuthoringQuizActionGeneration = "authoring_quiz_generation"
)

type ContextualQuizGenerationOptions struct {
	ProjectDir    string
	Session       AuthoringSession
	Now           time.Time
	AllowReplace  bool
	DryRun        bool
	QuestionCount int
	Provider      AuthoringQuizQuestionProvider
}

type ContextualQuizGenerationPlan struct {
	Session                AuthoringSession
	Layout                 AuthoringArtifactLayout
	Request                AuthoringQuizGenerationRequest
	Result                 AuthoringQuizGenerationResult
	Content                string
	WouldWrite             bool
	RequestedQuestionCount int
	GeneratedQuestionCount int
}

type AuthoringQuizQuestionProvider interface {
	GenerateAuthoringQuizQuestions(
		ctx context.Context,
		request AuthoringQuizGenerationRequest,
	) (AuthoringQuizGenerationResult, error)
}

type AuthoringQuizGenerationRequest struct {
	ID              string
	Action          string
	Strategy        string
	ProjectDir      string
	SessionID       string
	Target          AuthoringTarget
	Topic           string
	QuestionCount   int
	ContextSections []AuthoringQuizContextSection
	OutputContract  AuthoringQuizOutputContract
	CreatedAt       time.Time
}

type AuthoringQuizContextSection struct {
	Name      string
	Layer     string
	Authority string
	Content   string
}

type AuthoringQuizOutputContract struct {
	RequiredFields []string
}

type AuthoringQuizGenerationResult struct {
	Provider  string
	Strategy  string
	Questions []AuthoringQuizQuestion
	RawText   string
}

type AuthoringQuizQuestion struct {
	ID                    string
	Text                  string
	Tags                  []string
	FollowUpDepth         int
	CompletionRecommended bool
}

type FakeAuthoringQuizQuestionProvider struct {
	Result   AuthoringQuizGenerationResult
	Requests []AuthoringQuizGenerationRequest
	Err      error
}

type CodexAuthoringQuizQuestionProvider struct {
	Program string
	Runner  AgentRunner
}

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

func WriteContextualQuiz(plan ContextualQuizGenerationPlan) error {
	if plan.Layout.Collision.Exists && !plan.Layout.Collision.Allowed {
		return fmt.Errorf("authoring quiz artifact already exists: %s", plan.Layout.FullPath)
	}

	err := os.MkdirAll(plan.Layout.Dir, 0o755)
	if err != nil {
		return fmt.Errorf("create contextual quiz directory: %w", err)
	}

	err = os.WriteFile(plan.Layout.FullPath, []byte(plan.Content), 0o644)
	if err != nil {
		return fmt.Errorf("write contextual quiz artifact: %w", err)
	}

	return nil
}

func NewAuthoringQuizGenerationRequest(
	ctx context.Context,
	projectDir string,
	session AuthoringSession,
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

func LoadAuthoringQuizContextSections(
	ctx context.Context,
	projectDir string,
	session AuthoringSession,
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
	sections = append(sections, authoringQuizPathInventorySection(projectDir, "workspace_artifacts", "workspace", DirectoryName))
	sections = append(sections, authoringQuizPathInventorySection(projectDir, "docs", "docs", "docs"))
	sections = append(sections, authoringQuizTreeSection(projectDir))
	sections = append(sections, authoringQuizGoInventorySection(projectDir))

	return sections, nil
}

func RenderContextualQuizMarkdown(
	session AuthoringSession,
	artifact AuthoringArtifactRef,
	request AuthoringQuizGenerationRequest,
	result AuthoringQuizGenerationResult,
) string {
	var output bytes.Buffer

	output.WriteString("---\n")
	output.WriteString("session_id: " + yamlString(session.ID) + "\n")
	output.WriteString("project_id: " + yamlString(session.ProjectID) + "\n")
	output.WriteString("target: " + yamlString(string(session.Target)) + "\n")
	output.WriteString("artifact_id: " + yamlString(artifact.ID) + "\n")
	output.WriteString("artifact_kind: " + yamlString(string(artifact.Kind)) + "\n")
	output.WriteString("topic: " + yamlString(session.Topic) + "\n")
	output.WriteString("generation_strategy: " + yamlString(result.Strategy) + "\n")
	output.WriteString("provider: " + yamlString(result.Provider) + "\n")
	output.WriteString("requested_questions: " + fmt.Sprintf("%d", request.QuestionCount) + "\n")
	output.WriteString("generated_questions: " + fmt.Sprintf("%d", len(result.Questions)) + "\n")
	output.WriteString("created_at: " + yamlTimestamp(session.UpdatedAt) + "\n")
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

func (p CodexAuthoringQuizQuestionProvider) GenerateAuthoringQuizQuestions(
	ctx context.Context,
	request AuthoringQuizGenerationRequest,
) (AuthoringQuizGenerationResult, error) {
	err := ValidateAuthoringQuizGenerationRequest(request)
	if err != nil {
		return AuthoringQuizGenerationResult{}, err
	}

	program := strings.TrimSpace(p.Program)
	if program == "" {
		program = CodexWorkerProgram
	}

	payload, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return AuthoringQuizGenerationResult{}, fmt.Errorf("marshal authoring quiz request: %w", err)
	}

	prompt := authoringQuizCodexPrompt(string(payload))

	run, err := runCodexAgent(ctx, p.agentRunner(), AgentRunRequest{
		ID:                 request.ID,
		Purpose:            request.Action,
		ProjectDir:         request.ProjectDir,
		Prompt:             prompt,
		OutputName:         "output.json",
		Program:            program,
		DeferResponseTrace: true,
	})
	if err != nil {
		return AuthoringQuizGenerationResult{}, fmt.Errorf("invoke authoring quiz provider: %w", err)
	}

	result, err := ParseAuthoringQuizGenerationResult(run.Output)
	if err != nil {
		recordAgentRunResponseEvent(ctx, AgentRunRequest{
			ID:         request.ID,
			Purpose:    request.Action,
			ProjectDir: request.ProjectDir,
		}, run, "parse_error", err)

		return AuthoringQuizGenerationResult{}, err
	}

	result.Provider = "codex"
	result.Strategy = AuthoringQuizGenerationStrategyDeep
	result.RawText = string(run.Output)
	recordAgentRunResponseEvent(ctx, AgentRunRequest{
		ID:         request.ID,
		Purpose:    request.Action,
		ProjectDir: request.ProjectDir,
	}, run, "ready", nil)

	return result, nil
}

func (p CodexAuthoringQuizQuestionProvider) agentRunner() AgentRunner {
	if p.Runner != nil {
		return p.Runner
	}

	return CodexAgentRunner{}
}

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

	err := json.Unmarshal(extractJSONObject(output), &payload)
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
		Provider:  "codex",
		Strategy:  AuthoringQuizGenerationStrategyDeep,
		Questions: questions,
		RawText:   string(output),
	}, nil
}

func validateContextualQuizGenerationOptions(opts ContextualQuizGenerationOptions) error {
	if opts.ProjectDir == "" {
		return fmt.Errorf("contextual quiz project dir is required")
	}

	err := ValidateAuthoringSession(opts.Session)
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
		Kind:         AuthoringArtifactKindQuiz,
		Role:         AuthoringArtifactRoleOutput,
		Topic:        opts.Session.Topic,
		Description:  "Context-aware questionnaire generated for " + string(opts.Session.Target) + " authoring.",
		AllowReplace: opts.AllowReplace,
	}
}

func appendContextualQuizOutput(
	session AuthoringSession,
	artifact AuthoringArtifactRef,
	recordedAt time.Time,
) AuthoringSession {
	artifact.Role = AuthoringArtifactRoleOutput
	session.Outputs = append(session.Outputs, artifact)
	session.Provenance = append(session.Provenance, AuthoringProvenanceLink{
		Relation:   AuthoringProvenanceRelationGenerated,
		Source:     authoringQuizSourceArtifact(session),
		Target:     artifact,
		Decision:   "Generate a context-aware questionnaire for the authoring target.",
		RecordedAt: recordedAt.UTC(),
	})
	session.UpdatedAt = recordedAt.UTC()

	return session
}

func authoringQuizSourceArtifact(session AuthoringSession) AuthoringArtifactRef {
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

func authoringQuizSessionContext(session AuthoringSession) string {
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

func authoringQuizRelativePaths(
	projectDir string,
	root string,
	include func(string) bool,
) []string {
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

func authoringQuizCodexPrompt(payload string) string {
	return strings.TrimSpace(`
You generate context-aware authoring quiz questions for ChrysoCoord.

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

func extractJSONObject(output []byte) []byte {
	text := strings.TrimSpace(string(output))
	start := strings.Index(text, "{")

	end := strings.LastIndex(text, "}")
	if start < 0 || end < start {
		return output
	}

	return []byte(text[start : end+1])
}
