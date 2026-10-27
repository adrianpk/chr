package app

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// QuizListResult contains the current quiz list and a status message.
type QuizListResult struct {
	Message string
	States  []quizcore.RuntimeState
	Views   []QuizView
}

// QuizOperationResult is the TUI-facing result for quiz mutations.
type QuizOperationResult struct {
	Message string
	State   quizcore.RuntimeState
}

// QuizStartRequest selects the scope for a new quiz session.
type QuizStartRequest struct {
	Binding core.QuizActionBinding
	Scope   quizcore.Scope
	Mode    quizcore.Mode
}

// QuizResumeRequest selects the scope for a resumed quiz session.
type QuizResumeRequest struct {
	Binding core.QuizActionBinding
	Scope   quizcore.Scope
}

// QuizAnswerRequest submits an answer for the active quiz question.
type QuizAnswerRequest struct {
	Binding core.QuizActionBinding
	Scope   quizcore.Scope
	Answer  string
}

// QuizSkipRequest skips the active quiz question for the selected scope.
type QuizSkipRequest struct {
	Binding core.QuizActionBinding
	Scope   quizcore.Scope
}

// QuizCompleteRequest completes the quiz for the selected scope.
type QuizCompleteRequest struct {
	Binding core.QuizActionBinding
	Scope   quizcore.Scope
}

// QuizDraftSpecRequest selects the quiz scope used to draft a spec.
type QuizDraftSpecRequest struct {
	Scope quizcore.Scope
}

// QuizCancelRequest cancels the quiz for the selected scope.
type QuizCancelRequest struct {
	Binding core.QuizActionBinding
	Scope   quizcore.Scope
}

// QuizModeRequest changes the quiz interaction mode for the selected scope.
type QuizModeRequest struct {
	Binding core.QuizActionBinding
	Scope   quizcore.Scope
	Mode    quizcore.Mode
}

func (s service) ListQuizzes() (QuizListResult, error) {
	states, err := s.listQuizRuntimeStates()
	if err != nil {
		return QuizListResult{}, err
	}

	views, err := s.quizViewsFromStates(states)
	if err != nil {
		return QuizListResult{}, err
	}

	return QuizListResult{Message: "Quizzes loaded.", States: states, Views: views}, nil
}

func (s service) StartQuiz(request QuizStartRequest) (QuizOperationResult, error) {
	mode := request.Mode
	if mode == "" {
		mode = quizcore.ModeDeep
	}

	return s.applyQuizAction(core.QuizActionRequest{Scope: s.quizScope(request.Scope), Binding: request.Binding, Action: "start", Mode: mode})
}

func (s service) ResumeQuiz(request QuizResumeRequest) (QuizOperationResult, error) {
	return s.applyQuizAction(core.QuizActionRequest{Scope: s.quizScope(request.Scope), Binding: request.Binding, Action: "resume"})
}

func (s service) AnswerQuiz(request QuizAnswerRequest) (QuizOperationResult, error) {
	answer := strings.TrimSpace(request.Answer)
	if answer == "" {
		return QuizOperationResult{}, fmt.Errorf("quiz answer is required")
	}

	return s.applyQuizAction(core.QuizActionRequest{Scope: s.quizScope(request.Scope), Binding: request.Binding, Action: "answer", Answer: answer})
}

func (s service) SkipQuiz(request QuizSkipRequest) (QuizOperationResult, error) {
	return s.applyQuizAction(core.QuizActionRequest{Scope: s.quizScope(request.Scope), Binding: request.Binding, Action: "skip"})
}

func (s service) CompleteQuiz(request QuizCompleteRequest) (QuizOperationResult, error) {
	return s.applyQuizAction(core.QuizActionRequest{Scope: s.quizScope(request.Scope), Binding: request.Binding, Action: "done"})
}

func (s service) GenerateQuizDraftSpec(request QuizDraftSpecRequest) (QuizOperationResult, error) {
	return s.generateQuizDraftSpec(s.quizScope(request.Scope))
}

func (s service) CancelQuiz(request QuizCancelRequest) (QuizOperationResult, error) {
	return s.applyQuizAction(core.QuizActionRequest{Scope: s.quizScope(request.Scope), Binding: request.Binding, Action: "cancel"})
}

func (s service) SwitchQuizMode(request QuizModeRequest) (QuizOperationResult, error) {
	return s.applyQuizAction(core.QuizActionRequest{Scope: s.quizScope(request.Scope), Binding: request.Binding, Action: "mode", Mode: request.Mode})
}

func (s service) generateQuizDraftSpec(scope quizcore.Scope) (QuizOperationResult, error) {
	bound, _, err := s.bindGeneration(s.ctx, "", string(quizmodel.ActionDraftSpec))
	if err != nil {
		return QuizOperationResult{}, err
	}

	s = bound
	repository := s.quizRepository()

	state, err := repository.LoadRuntimeState(s.ctx, scope)
	if err != nil {
		return QuizOperationResult{}, err
	}

	paths, err := s.writeQuizCuratedNotes(scope)
	if err != nil {
		return QuizOperationResult{}, err
	}

	err = s.writeQuizDraftSpec(state, paths)
	if err != nil {
		return QuizOperationResult{}, err
	}

	return QuizOperationResult{Message: "Spec generated.", State: state}, nil
}

func (s service) writeQuizCuratedNotes(scope quizcore.Scope) (core.QuizArtifactPaths, error) {
	repository := s.quizRepository()

	entries, err := repository.LoadRawCaptureEntries(s.ctx, scope)
	if err != nil {
		return core.QuizArtifactPaths{}, err
	}

	return repository.WriteCuratedNotes(s.ctx, quizcore.CurateEntries(scope, entries, s.utcNow()))
}

func (s service) newQuizModelRequest(state quizcore.RuntimeState) (quizmodel.Request, error) {
	entries, err := s.quizRepository().LoadRawCaptureEntries(s.ctx, quizcore.Scope{ProjectID: state.ProjectID, CaseSlug: state.CaseSlug})
	if err != nil {
		return quizmodel.Request{}, err
	}

	return s.newQuizModelRequestWithEntries(state, entries)
}

func (s service) newQuizModelRequestWithEntries(state quizcore.RuntimeState, entries []quizcore.Entry) (quizmodel.Request, error) {
	if s.guidedContext != nil {
		return s.guidedQuizModelRequest(state, entries)
	}

	scope := quizcore.Scope{
		ProjectID: state.ProjectID,
		CaseSlug:  state.CaseSlug,
	}

	payload := s.quizModelPayloadWithEntries(state, quizmodel.ActionNextQuestion, entries)

	coreSections, err := quizmodel.ContextSectionsForPayload(payload)
	if err != nil {
		return quizmodel.Request{}, err
	}

	supplementalSections, err := s.appendQuizSupplementalContextSections(nil, scope)
	if err != nil {
		return quizmodel.Request{}, err
	}

	history, err := s.quizHistoryContext(scope, entries)
	if err != nil {
		return quizmodel.Request{}, err
	}

	supplementalSections = append(supplementalSections, history)

	readiness := quizAcceptanceReadinessFromContext(state, append(coreSections, supplementalSections...))
	payload.AcceptanceReadiness = &readiness

	coreSections, err = quizmodel.ContextSectionsForPayload(payload)
	if err != nil {
		return quizmodel.Request{}, err
	}

	return quizmodel.Request{
		ID:              core.HarnessGenerationRequestID(s.ctx, fmt.Sprintf("quiz-request-%d", state.QuestionsAsked+1)),
		Action:          quizmodel.ActionNextQuestion,
		Strategy:        quizModelStrategy(state.Mode),
		Scope:           scope,
		Mode:            state.Mode,
		RuntimeState:    state,
		Payload:         payload,
		ContextSections: append(coreSections, supplementalSections...),
		OutputContract: quizmodel.OutputContract{
			RequiredFields: payload.OutputContract.RequiredFields,
		},
		CreatedAt: s.utcNow(),
	}, nil
}

func (s service) writeQuizDraftSpec(state quizcore.RuntimeState, paths core.QuizArtifactPaths) error {
	target, err := core.QuizDraftSpecLayout(s.projectContext.ProjectDir, quizcore.Scope{ProjectID: state.ProjectID, CaseSlug: state.CaseSlug})
	if err != nil {
		return err
	}

	revision, err := core.SpecTargetRevision(s.ctx, s.projectContext.ProjectDir, target.RelativePath)
	if err != nil {
		return err
	}

	request, err := s.newQuizDraftSpecModelRequest(state, paths)
	if err != nil {
		return err
	}

	response, err := quizmodel.Invoke(s.ctx, s.quizModel, request)
	if err != nil {
		return err
	}

	spec := response.DraftSpec
	spec.ExpectedRevision = revision
	spec.Scope = quizcore.Scope{ProjectID: state.ProjectID, CaseSlug: state.CaseSlug}
	spec.SourceQuizPath = paths.CuratedRelativePath
	spec.SourceRawPath = paths.RawRelativePath
	spec.SourceDumpPaths = quizDumpSourcePaths(request.ContextSections)
	spec.SourceRequestID = request.ID
	spec.SourceResponseID = response.ID
	spec.CreatedAt = response.CreatedAt.UTC()

	_, err = core.WriteQuizDraftSpec(s.ctx, s.projectContext.ProjectDir, spec)

	return err
}

func (s service) newQuizDraftSpecModelRequest(
	state quizcore.RuntimeState,
	paths core.QuizArtifactPaths,
) (quizmodel.Request, error) {
	scope := quizcore.Scope{
		ProjectID: state.ProjectID,
		CaseSlug:  state.CaseSlug,
	}

	payload, err := s.newQuizModelPayload(scope, state, quizmodel.ActionDraftSpec)
	if err != nil {
		return quizmodel.Request{}, err
	}

	sections, err := quizmodel.ContextSectionsForPayload(payload)
	if err != nil {
		return quizmodel.Request{}, err
	}

	sections, err = s.appendQuizSupplementalContextSections(sections, scope)
	if err != nil {
		return quizmodel.Request{}, err
	}

	sections = appendQuizFileContextSection(
		sections,
		s.projectContext.ProjectDir,
		"curated_quiz",
		"quiz",
		paths.CuratedRelativePath,
	)

	return quizmodel.Request{
		ID:              core.HarnessGenerationRequestID(s.ctx, fmt.Sprintf("quiz-draft-spec-%s", paths.ScopeID)),
		Action:          quizmodel.ActionDraftSpec,
		Strategy:        "draft_spec",
		Scope:           scope,
		Mode:            state.Mode,
		RuntimeState:    state,
		Payload:         payload,
		ContextSections: sections,
		OutputContract: quizmodel.OutputContract{
			RequiredFields: payload.OutputContract.RequiredFields,
		},
		CreatedAt: s.utcNow(),
	}, nil
}

func (s service) newQuizModelPayload(
	scope quizcore.Scope,
	state quizcore.RuntimeState,
	action quizmodel.Action,
) (quizmodel.Payload, error) {
	entries, err := s.quizRepository().LoadRawCaptureEntries(s.ctx, scope)
	if err != nil {
		return quizmodel.Payload{}, err
	}

	return s.quizModelPayloadWithEntries(state, action, entries), nil
}

func (s service) quizModelPayloadWithEntries(state quizcore.RuntimeState, action quizmodel.Action, entries []quizcore.Entry) quizmodel.Payload {
	return quizmodel.Payload{
		Project:         s.quizModelProjectPayload(),
		Quiz:            quizModelQuizPayload(state),
		RecentEntries:   quizModelRecentEntryPayloads(entries, 5),
		CurrentQuestion: quizModelCurrentQuestionPayload(state, entries),
		OutputContract:  quizmodel.NewOutputContract(action, state.Mode),
	}
}

func (s service) newQuizRuntimeState(scope quizcore.Scope, mode quizcore.Mode) (quizcore.RuntimeState, error) {
	paths, err := s.quizRepository().Paths(scope)
	if err != nil {
		return quizcore.RuntimeState{}, err
	}

	now := s.utcNow()

	return quizcore.RuntimeState{
		ID:             "quiz-" + paths.ScopeID,
		SessionID:      "quiz-" + rand.Text(),
		ProjectID:      scope.ProjectID,
		CaseSlug:       scope.CaseSlug,
		Status:         quizcore.RuntimeStatusIdle,
		Mode:           mode,
		QuestionBudget: quizcore.DefaultQuestionBudget,
		RawPath:        paths.RawRelativePath,
		CuratedPath:    paths.CuratedRelativePath,
		UpdatedAt:      now,
	}, nil
}

func (s service) quizModelProjectPayload() quizmodel.ProjectPayload {
	payload := quizmodel.ProjectPayload{ID: s.projectContext.ProjectID}

	metadata, err := core.NewProjectMetadataService(s.projectContext.ProjectDir).Load()
	if err != nil {
		return payload
	}

	payload.Name = metadata.Name
	payload.Description = metadata.Description

	return payload
}

func (s service) appendQuizSupplementalContextSections(
	sections []quizmodel.ContextSection,
	scope quizcore.Scope,
) ([]quizmodel.ContextSection, error) {
	sections = appendQuizFileContextSection(sections, s.projectContext.ProjectDir, "agent_protocol", "protocol", "AGENTS.md")
	sections = appendQuizFileContextSection(sections, s.projectContext.ProjectDir, "project_document", "project", core.ProjectMetadataPath(""))

	if scope.CaseSlug != "" {
		sections = appendQuizFileContextSection(
			sections,
			s.projectContext.ProjectDir,
			"use_case_document",
			"use_case",
			core.CaseMetadataPath("", scope.CaseSlug),
		)
	}

	index, err := s.workspaceIndex()
	if err != nil {
		return nil, err
	}

	for _, nodeID := range index.ByKind[core.WorkspaceNodeDump] {
		node := index.NodesByID[nodeID]
		if core.IsRawDumpNode(node) {
			continue
		}

		if node.CaseSlug != "" && node.CaseSlug != scope.CaseSlug {
			continue
		}

		sections = appendQuizNodeContextSection(sections, s.projectContext.ProjectDir, node)
	}

	return sections, nil
}

func quizModelQuizPayload(state quizcore.RuntimeState) quizmodel.QuizPayload {
	return quizmodel.QuizPayload{
		ID:         state.ID,
		UseCaseID:  state.CaseSlug,
		EntryCount: state.AnswersRecorded + state.SkipsRecorded,
		Mode:       string(state.Mode),
	}
}

func quizModelRecentEntryPayloads(entries []quizcore.Entry, limit int) []quizmodel.EntryPayload {
	if limit <= 0 {
		return []quizmodel.EntryPayload{}
	}

	start := len(entries) - limit
	if start < 0 {
		start = 0
	}

	payloads := make([]quizmodel.EntryPayload, 0, len(entries[start:]))
	for _, entry := range entries[start:] {
		payloads = append(payloads, quizmodel.EntryPayload{
			Question: entry.Question,
			Answer:   entry.Answer,
			Status:   string(entry.Status),
			Tags:     entry.Tags,
		})
	}

	return payloads
}

func quizModelCurrentQuestionPayload(
	state quizcore.RuntimeState,
	entries []quizcore.Entry,
) quizmodel.CurrentQuestionPayload {
	if state.ActiveQuestion != nil {
		return quizModelQuestionPayloadFromQuestion(*state.ActiveQuestion)
	}

	if len(entries) == 0 {
		return quizmodel.CurrentQuestionPayload{PromptIntent: map[string]any{}}
	}

	entry := entries[len(entries)-1]
	promptIntent := quizModelPromptIntentFromJSON(entry.PromptIntent)

	if len(promptIntent) == 0 {
		promptIntent = map[string]any{
			"metadata": map[string]any{
				"follow_up_depth": entry.Depth - 1,
				"status":          string(entry.Status),
			},
		}
	}

	return quizmodel.CurrentQuestionPayload{
		Question:     entry.Question,
		Tags:         quizModelEntryTags(entry),
		PromptIntent: promptIntent,
	}
}

func quizModelQuestionPayloadFromQuestion(question quizcore.Question) quizmodel.CurrentQuestionPayload {
	promptIntent := quizModelPromptIntentFromJSON(question.Rationale)
	if len(promptIntent) == 0 {
		promptIntent = map[string]any{
			"metadata": map[string]any{
				"follow_up_depth":        question.Depth - 1,
				"completion_recommended": !question.Required,
			},
		}
	}

	return quizmodel.CurrentQuestionPayload{
		Question:     question.Prompt,
		Tags:         quizModelQuestionTags(question),
		PromptIntent: promptIntent,
	}
}

func quizModelPromptIntentFromJSON(value string) map[string]any {
	value = strings.TrimSpace(value)
	if value == "" {
		return map[string]any{}
	}

	var promptIntent map[string]any

	err := json.Unmarshal([]byte(value), &promptIntent)
	if err != nil {
		return map[string]any{}
	}

	return promptIntent
}

func quizAcceptanceReadinessFromContext(
	state quizcore.RuntimeState,
	sections []quizmodel.ContextSection,
) generationcore.AcceptanceReadiness {
	content := strings.ToLower(quizContextText(sections))
	reasons := []string{"The quiz context was inspected for acceptance-test readiness."}

	if quizContextLooksNonAcceptanceTestable(content) {
		return generationcore.AcceptanceReadiness{
			Decision:         generationcore.AcceptanceReadinessNotApplicable,
			Confidence:       0.7,
			Reasons:          append(reasons, "The context appears internal, architectural, documentation-only, or refactor-oriented."),
			InferredElements: map[generationcore.AcceptanceReadinessSlot][]string{},
		}
	}

	inferred := quizAcceptanceInferredElements(content)
	missing := quizAcceptanceMissingSlots(inferred)

	if len(missing) == 0 {
		return generationcore.AcceptanceReadiness{
			Decision:           generationcore.AcceptanceReadinessGenerate,
			Confidence:         0.72,
			Reasons:            append(reasons, "The context includes actor, trigger, initial state, outcome, and success-path signals."),
			InferredElements:   inferred,
			CandidateScenarios: []string{"Primary behavior succeeds and produces the expected observable outcome."},
		}
	}

	return generationcore.AcceptanceReadiness{
		Decision:         generationcore.AcceptanceReadinessNeedsMoreInfo,
		Confidence:       0.64,
		Reasons:          append(reasons, "The context has some acceptance-test signals, but key scenario elements are missing."),
		Missing:          missing,
		InferredElements: inferred,
	}
}

func quizContextText(sections []quizmodel.ContextSection) string {
	var output strings.Builder
	for _, section := range sections {
		output.WriteString(section.Content)
		output.WriteString("\n")
	}

	return output.String()
}

func quizContextLooksNonAcceptanceTestable(content string) bool {
	nonAcceptanceSignals := []string{
		"documentation-only",
		"docs-only",
		"internal refactor",
		"refactor only",
		"architecture cleanup",
		"package boundary",
	}
	for _, signal := range nonAcceptanceSignals {
		if strings.Contains(content, signal) {
			return true
		}
	}

	return false
}

func quizAcceptanceInferredElements(content string) map[generationcore.AcceptanceReadinessSlot][]string {
	inferred := map[generationcore.AcceptanceReadinessSlot][]string{}

	if quizContainsAny(content, "user", "operator", "cli", "api client", "worker", "scheduler", "reviewer") {
		inferred[generationcore.AcceptanceReadinessSlotActor] = []string{"An actor is referenced by the source context."}
	}

	if quizContainsAny(content, "command", "request", "event", "runs", "starts", "trigger", "action") {
		inferred[generationcore.AcceptanceReadinessSlotTrigger] = []string{"A trigger is referenced by the source context."}
	}

	if quizContainsAny(content, "given", "existing", "configured", "before", "state", "project", "ticket", "file") {
		inferred[generationcore.AcceptanceReadinessSlotInitialState] = []string{"An initial state is referenced by the source context."}
	}

	if quizContainsAny(content, "output", "report", "status", "response", "error", "file", "written", "created") {
		inferred[generationcore.AcceptanceReadinessSlotObservableOutcome] = []string{"An observable outcome is referenced by the source context."}
	}

	if quizContainsAny(content, "success", "succeeds", "passes", "accepted", "valid", "expected") {
		inferred[generationcore.AcceptanceReadinessSlotSuccessPath] = []string{"A success path is referenced by the source context."}
	}

	if quizContainsAny(content, "failure", "fails", "invalid", "reject", "error", "missing") {
		inferred[generationcore.AcceptanceReadinessSlotFailurePath] = []string{"A failure path is referenced by the source context."}
	}

	if quizContainsAny(content, "out of scope", "non-goal", "boundary", "must not", "do not") {
		inferred[generationcore.AcceptanceReadinessSlotBoundaries] = []string{"A boundary is referenced by the source context."}
	}

	return inferred
}

func quizAcceptanceMissingSlots(inferred map[generationcore.AcceptanceReadinessSlot][]string) []generationcore.AcceptanceReadinessSlot {
	required := []generationcore.AcceptanceReadinessSlot{
		generationcore.AcceptanceReadinessSlotActor,
		generationcore.AcceptanceReadinessSlotTrigger,
		generationcore.AcceptanceReadinessSlotInitialState,
		generationcore.AcceptanceReadinessSlotObservableOutcome,
		generationcore.AcceptanceReadinessSlotSuccessPath,
	}
	missing := []generationcore.AcceptanceReadinessSlot{}

	for _, slot := range required {
		if len(inferred[slot]) == 0 {
			missing = append(missing, slot)
		}
	}

	return missing
}

func quizContainsAny(content string, values ...string) bool {
	for _, value := range values {
		if strings.Contains(content, value) {
			return true
		}
	}

	return false
}

func quizModelQuestionTags(question quizcore.Question) []string {
	if strings.TrimSpace(question.Topic) == "" {
		return []string{}
	}

	return []string{question.Topic}
}

func quizModelEntryTags(entry quizcore.Entry) []string {
	if len(entry.Tags) > 0 {
		return entry.Tags
	}

	if strings.TrimSpace(entry.Topic) == "" {
		return []string{}
	}

	return []string{entry.Topic}
}

func quizModelStrategy(mode quizcore.Mode) string {
	if mode == quizcore.ModeFast {
		return quizmodel.FastStrategy
	}

	return quizmodel.DeepStrategy
}

func (s service) listQuizRuntimeStates() ([]quizcore.RuntimeState, error) {
	root := core.UserPath(s.projectContext.ProjectDir, filepath.FromSlash(core.QuizRuntimeDirectory))

	entries, err := readQuizRuntimeDirectory(root)
	if errors.Is(err, os.ErrNotExist) {
		return []quizcore.RuntimeState{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("read quiz runtime directory: %w", err)
	}

	states := []quizcore.RuntimeState{}
	repository := s.quizRepository()

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		scope := quizcore.Scope{
			ProjectID: s.projectContext.ProjectID,
			CaseSlug:  strings.TrimSuffix(entry.Name(), ".json"),
		}
		if scope.CaseSlug == "project" {
			scope.CaseSlug = ""
		}

		state, err := repository.LoadRuntimeState(s.ctx, scope)
		if err != nil {
			return nil, err
		}

		states = append(states, state)
	}

	return states, nil
}

func (s service) quizRepository() core.QuizArtifactRepository {
	return core.NewQuizArtifactRepository(s.projectContext.ProjectDir)
}

func (s service) quizScope(scope quizcore.Scope) quizcore.Scope {
	scope.ProjectID = s.projectContext.ProjectID

	return scope
}

func (s service) utcNow() time.Time {
	if s.now == nil {
		return time.Now().UTC()
	}

	return s.now().UTC()
}

func appendQuizFileContextSection(
	sections []quizmodel.ContextSection,
	projectDir string,
	name string,
	layer string,
	path string,
) []quizmodel.ContextSection {
	if !filepath.IsAbs(path) {
		path = filepath.Join(projectDir, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return sections
	}

	return append(sections, quizmodel.ContextSection{
		Name:      name,
		Layer:     layer,
		Authority: filepath.ToSlash(path),
		Content:   string(content),
	})
}

func appendQuizNodeContextSection(
	sections []quizmodel.ContextSection,
	projectDir string,
	node core.WorkspaceNode,
) []quizmodel.ContextSection {
	content, err := os.ReadFile(filepath.Join(projectDir, filepath.FromSlash(node.RelativePath)))
	if err != nil {
		return sections
	}

	name := "project_dump"
	if node.CaseSlug != "" {
		name = "use_case_dump"
	}

	return append(sections, quizmodel.ContextSection{
		Name:      name,
		Layer:     "dump",
		Authority: node.RelativePath,
		Content:   string(content),
	})
}

func quizDumpSourcePaths(sections []quizmodel.ContextSection) []string {
	paths := []string{}

	for _, section := range sections {
		if section.Layer == "dump" {
			paths = append(paths, section.Authority)
		}
	}

	return paths
}

func quizEntryFromState(
	state quizcore.RuntimeState,
	status quizcore.EntryStatus,
	answer string,
	now time.Time,
) quizcore.Entry {
	question := *state.ActiveQuestion

	return quizcore.Entry{
		ID:           fmt.Sprintf("entry-%03d", state.QuestionsAsked),
		QuestionID:   question.ID,
		Question:     question.Prompt,
		Answer:       strings.TrimSpace(answer),
		Status:       status,
		Topic:        question.Topic,
		Depth:        question.Depth,
		ParentID:     question.ParentID,
		Mode:         state.Mode,
		Tags:         quizModelQuestionTags(question),
		PromptIntent: question.Rationale,
		CreatedAt:    question.CreatedAt,
		AnsweredAt:   now,
	}
}

func updateQuizStateAfterEntry(state quizcore.RuntimeState, entry quizcore.Entry) quizcore.RuntimeState {
	if entry.Status == quizcore.EntryStatusAnswered {
		state.AnswersRecorded++
	}

	if entry.Status == quizcore.EntryStatusSkipped {
		state.SkipsRecorded++
	}

	state.Status = quizcore.RuntimeStatusIdle
	state.ActiveQuestion = nil
	state.ActiveTopic = ""

	if entry.Status == quizcore.EntryStatusSkipped {
		state.Trail = nil
		state.BranchReturnReason = quizcore.BranchReturnReasonSkipped
	} else if entry.Depth >= quizcore.MaxFollowUpDepth {
		state.Trail = nil
		state.BranchReturnReason = quizcore.BranchReturnReasonMaxDepth
	} else {
		state.BranchReturnReason = ""
	}

	if state.Mode == quizcore.ModeDeep {
		state.Candidates = nil
	}

	return state
}

func activateQuizQuestion(
	state quizcore.RuntimeState,
	question quizcore.Question,
	now time.Time,
) quizcore.RuntimeState {
	state.Status = quizcore.RuntimeStatusActive
	state.ActiveQuestion = &question
	state.ActiveTopic = question.Topic
	state.Trail = quizTrailForQuestion(state.Trail, question)
	state.BranchReturnReason = ""
	state.QuestionsAsked++
	state.UpdatedAt = now

	return state
}

func activateQuizCandidate(
	state quizcore.RuntimeState,
	candidate quizcore.Candidate,
	now time.Time,
) quizcore.RuntimeState {
	question := quizcore.Question{
		ID:        candidate.ID,
		Prompt:    candidate.Prompt,
		Required:  true,
		Topic:     candidate.Topic,
		Depth:     candidate.Depth,
		ParentID:  candidate.ParentID,
		TrunkID:   candidate.ID,
		RouteID:   candidate.RouteID,
		CreatedAt: candidate.CreatedAt,
	}
	if question.CreatedAt.IsZero() {
		question.CreatedAt = now
	}

	if question.Depth > 1 {
		question.TrunkID = candidate.RouteID
		if strings.TrimSpace(question.TrunkID) == "" {
			question.TrunkID = candidate.ParentID
		}
	}

	state = activateQuizQuestion(state, question, now)

	return state
}

func quizTrailForQuestion(trail []quizcore.TrailNode, question quizcore.Question) []quizcore.TrailNode {
	node := quizcore.TrailNode{
		QuestionID: question.ID,
		Prompt:     question.Prompt,
		Topic:      question.Topic,
		Depth:      question.Depth,
	}

	if question.Depth <= 1 {
		return []quizcore.TrailNode{node}
	}

	parentIndex := -1

	for index, current := range trail {
		if current.QuestionID == question.ParentID {
			parentIndex = index

			break
		}
	}

	if parentIndex < 0 {
		return []quizcore.TrailNode{node}
	}

	next := append([]quizcore.TrailNode(nil), trail[:parentIndex+1]...)
	next = append(next, node)

	return next
}

func readQuizRuntimeDirectory(root string) ([]os.DirEntry, error) {
	directory, err := os.Open(root)
	if err != nil {
		return nil, err
	}

	entries, readErr := directory.ReadDir(workflowstate.RecordLimit + 1)
	if errors.Is(readErr, io.EOF) {
		readErr = nil
	}

	if err := errors.Join(readErr, directory.Close()); err != nil {
		return nil, err
	}

	if len(entries) > workflowstate.RecordLimit {
		return nil, fmt.Errorf("quiz runtime inventory capacity exceeded")
	}

	return entries, nil
}
