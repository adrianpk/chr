package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chc/internal/core"
)

// QuizListResult contains the current quiz list and a status message.
type QuizListResult struct {
	Message string
	States  []core.QuizRuntimeState
	Views   []QuizView
}

// QuizOperationResult is the TUI-facing result for quiz mutations.
type QuizOperationResult struct {
	Message string
	State   core.QuizRuntimeState
}

// QuizStartRequest selects the scope for a new quiz session.
type QuizStartRequest struct {
	Scope core.QuizScope
	Mode  core.QuizMode
}

// QuizResumeRequest selects the scope for a resumed quiz session.
type QuizResumeRequest struct {
	Scope core.QuizScope
}

// QuizAnswerRequest submits an answer for the active quiz question.
type QuizAnswerRequest struct {
	Scope  core.QuizScope
	Answer string
}

// QuizSkipRequest skips the active quiz question for the selected scope.
type QuizSkipRequest struct {
	Scope core.QuizScope
}

// QuizCompleteRequest completes the quiz for the selected scope.
type QuizCompleteRequest struct {
	Scope core.QuizScope
}

// QuizDraftSpecRequest selects the quiz scope used to draft a spec.
type QuizDraftSpecRequest struct {
	Scope core.QuizScope
}

// QuizCancelRequest cancels the quiz for the selected scope.
type QuizCancelRequest struct {
	Scope core.QuizScope
}

// QuizModeRequest changes the quiz interaction mode for the selected scope.
type QuizModeRequest struct {
	Scope core.QuizScope
	Mode  core.QuizMode
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
	scope := s.quizScope(request.Scope)

	mode := request.Mode
	if mode == "" {
		mode = core.QuizModeDeep
	}

	state, err := s.newQuizRuntimeState(scope, mode)
	if err != nil {
		return QuizOperationResult{}, err
	}

	state, err = s.generateNextQuizQuestion(state)
	if err != nil {
		return QuizOperationResult{}, err
	}

	_, err = s.quizRepository().SaveRuntimeState(s.ctx, state)
	if err != nil {
		return QuizOperationResult{}, err
	}

	return QuizOperationResult{Message: "Quiz started.", State: state}, nil
}

func (s service) ResumeQuiz(request QuizResumeRequest) (QuizOperationResult, error) {
	state, err := s.quizRepository().LoadRuntimeState(s.ctx, s.quizScope(request.Scope))
	if err != nil {
		return QuizOperationResult{}, err
	}

	return QuizOperationResult{Message: "Quiz resumed.", State: state}, nil
}

func (s service) AnswerQuiz(request QuizAnswerRequest) (QuizOperationResult, error) {
	answer := strings.TrimSpace(request.Answer)
	if answer == "" {
		return QuizOperationResult{}, fmt.Errorf("quiz answer is required")
	}

	return s.recordQuizOutcome(s.quizScope(request.Scope), core.QuizEntryStatusAnswered, answer)
}

func (s service) SkipQuiz(request QuizSkipRequest) (QuizOperationResult, error) {
	return s.recordQuizOutcome(s.quizScope(request.Scope), core.QuizEntryStatusSkipped, "")
}

func (s service) CompleteQuiz(request QuizCompleteRequest) (QuizOperationResult, error) {
	return s.completeQuiz(s.quizScope(request.Scope))
}

func (s service) GenerateQuizDraftSpec(request QuizDraftSpecRequest) (QuizOperationResult, error) {
	return s.generateQuizDraftSpec(s.quizScope(request.Scope))
}

func (s service) CancelQuiz(request QuizCancelRequest) (QuizOperationResult, error) {
	return s.stopQuiz(s.quizScope(request.Scope), core.QuizRuntimeStatusCancelled, "Quiz cancelled.")
}

func (s service) SwitchQuizMode(request QuizModeRequest) (QuizOperationResult, error) {
	if !request.Mode.Valid() {
		return QuizOperationResult{}, fmt.Errorf("quiz mode is invalid: %s", request.Mode)
	}

	repository := s.quizRepository()

	state, err := repository.LoadRuntimeState(s.ctx, s.quizScope(request.Scope))
	if err != nil {
		return QuizOperationResult{}, err
	}

	state.Mode = request.Mode
	state.Candidates = nil
	state.UpdatedAt = s.utcNow()

	if state.Status == core.QuizRuntimeStatusIdle && state.Mode == core.QuizModeFast {
		state, err = s.generateNextQuizQuestion(state)
		if err != nil {
			return QuizOperationResult{}, err
		}
	}

	_, err = repository.SaveRuntimeState(s.ctx, state)
	if err != nil {
		return QuizOperationResult{}, err
	}

	return QuizOperationResult{Message: "Quiz mode switched.", State: state}, nil
}

func (s service) recordQuizOutcome(
	scope core.QuizScope,
	status core.QuizEntryStatus,
	answer string,
) (QuizOperationResult, error) {
	repository := s.quizRepository()

	state, err := repository.LoadRuntimeState(s.ctx, scope)
	if err != nil {
		return QuizOperationResult{}, err
	}

	if state.Status != core.QuizRuntimeStatusActive || state.ActiveQuestion == nil {
		return QuizOperationResult{}, fmt.Errorf("quiz has no active question")
	}

	entry := quizEntryFromState(state, status, answer, s.utcNow())

	_, err = repository.AppendRawCapture(s.ctx, state, entry)
	if err != nil {
		return QuizOperationResult{}, err
	}

	_, err = s.writeQuizCuratedNotes(scope)
	if err != nil {
		return QuizOperationResult{}, err
	}

	state = updateQuizStateAfterEntry(state, entry)
	if state.QuestionsAsked < state.QuestionBudget {
		state, err = s.generateNextQuizQuestion(state)
		if err != nil {
			return QuizOperationResult{}, err
		}
	} else {
		state.Status = core.QuizRuntimeStatusDone
	}

	state.UpdatedAt = s.utcNow()

	_, err = repository.SaveRuntimeState(s.ctx, state)
	if err != nil {
		return QuizOperationResult{}, err
	}

	return QuizOperationResult{Message: "Quiz answer recorded.", State: state}, nil
}

func (s service) stopQuiz(
	scope core.QuizScope,
	status core.QuizRuntimeStatus,
	message string,
) (QuizOperationResult, error) {
	repository := s.quizRepository()

	state, err := repository.LoadRuntimeState(s.ctx, scope)
	if err != nil {
		return QuizOperationResult{}, err
	}

	state.Status = status
	state.ActiveQuestion = nil
	state.Candidates = nil
	state.UpdatedAt = s.utcNow()

	_, err = repository.SaveRuntimeState(s.ctx, state)
	if err != nil {
		return QuizOperationResult{}, err
	}

	return QuizOperationResult{Message: message, State: state}, nil
}

func (s service) completeQuiz(scope core.QuizScope) (QuizOperationResult, error) {
	repository := s.quizRepository()

	state, err := repository.LoadRuntimeState(s.ctx, scope)
	if err != nil {
		return QuizOperationResult{}, err
	}

	state.Status = core.QuizRuntimeStatusDone
	state.ActiveQuestion = nil
	state.Candidates = nil
	state.UpdatedAt = s.utcNow()

	_, err = s.writeQuizCuratedNotes(scope)
	if err != nil {
		return QuizOperationResult{}, err
	}

	_, err = repository.SaveRuntimeState(s.ctx, state)
	if err != nil {
		return QuizOperationResult{}, err
	}

	return QuizOperationResult{Message: "Quiz completed.", State: state}, nil
}

func (s service) generateQuizDraftSpec(scope core.QuizScope) (QuizOperationResult, error) {
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

func (s service) writeQuizCuratedNotes(scope core.QuizScope) (core.QuizArtifactPaths, error) {
	repository := s.quizRepository()

	entries, err := repository.LoadRawCaptureEntries(s.ctx, scope)
	if err != nil {
		return core.QuizArtifactPaths{}, err
	}

	return repository.WriteCuratedNotes(s.ctx, core.CurateQuizEntries(scope, entries, s.utcNow()))
}

func (s service) generateNextQuizQuestion(state core.QuizRuntimeState) (core.QuizRuntimeState, error) {
	if len(state.Candidates) > 0 {
		state = activateQuizCandidate(state, state.Candidates[0], s.utcNow())
		state.Candidates = state.Candidates[1:]

		return state, nil
	}

	request, err := s.newQuizModelRequest(state)
	if err != nil {
		return core.QuizRuntimeState{}, err
	}

	response, err := core.InvokeQuizModel(s.ctx, s.quizModel, request)
	if err != nil {
		return core.QuizRuntimeState{}, err
	}

	if response.Status == core.QuizModelResponseBlocked {
		state.Status = core.QuizRuntimeStatusBlocked
		state.ActiveQuestion = nil
		state.UpdatedAt = response.CreatedAt.UTC()

		return state, nil
	}

	state.Status = core.QuizRuntimeStatusActive
	state = activateQuizQuestion(state, response.Question, response.CreatedAt.UTC())
	state.Candidates = response.Candidates

	return state, nil
}

func (s service) newQuizModelRequest(state core.QuizRuntimeState) (core.QuizModelRequest, error) {
	scope := core.QuizScope{
		ProjectID: state.ProjectID,
		CaseSlug:  state.CaseSlug,
	}

	payload, err := s.newQuizModelPayload(scope, state, core.QuizModelActionNextQuestion)
	if err != nil {
		return core.QuizModelRequest{}, err
	}

	coreSections, err := core.QuizModelContextSectionsForPayload(payload)
	if err != nil {
		return core.QuizModelRequest{}, err
	}

	supplementalSections, err := s.appendQuizSupplementalContextSections(nil, scope)
	if err != nil {
		return core.QuizModelRequest{}, err
	}

	readiness := quizAcceptanceReadinessFromContext(state, append(coreSections, supplementalSections...))
	payload.AcceptanceReadiness = &readiness

	coreSections, err = core.QuizModelContextSectionsForPayload(payload)
	if err != nil {
		return core.QuizModelRequest{}, err
	}

	return core.QuizModelRequest{
		ID:              fmt.Sprintf("quiz-request-%d", state.QuestionsAsked+1),
		Action:          core.QuizModelActionNextQuestion,
		Strategy:        quizModelStrategy(state.Mode),
		Scope:           scope,
		Mode:            state.Mode,
		RuntimeState:    state,
		Payload:         payload,
		ContextSections: append(coreSections, supplementalSections...),
		OutputContract: core.QuizModelOutputContract{
			RequiredFields: payload.OutputContract.RequiredFields,
		},
		CreatedAt: s.utcNow(),
	}, nil
}

func (s service) writeQuizDraftSpec(
	state core.QuizRuntimeState,
	paths core.QuizArtifactPaths,
) error {
	request, err := s.newQuizDraftSpecModelRequest(state, paths)
	if err != nil {
		return err
	}

	response, err := core.InvokeQuizModel(s.ctx, s.quizModel, request)
	if err != nil {
		return err
	}

	spec := response.DraftSpec
	spec.Scope = core.QuizScope{ProjectID: state.ProjectID, CaseSlug: state.CaseSlug}
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
	state core.QuizRuntimeState,
	paths core.QuizArtifactPaths,
) (core.QuizModelRequest, error) {
	scope := core.QuizScope{
		ProjectID: state.ProjectID,
		CaseSlug:  state.CaseSlug,
	}

	payload, err := s.newQuizModelPayload(scope, state, core.QuizModelActionDraftSpec)
	if err != nil {
		return core.QuizModelRequest{}, err
	}

	sections, err := core.QuizModelContextSectionsForPayload(payload)
	if err != nil {
		return core.QuizModelRequest{}, err
	}

	sections, err = s.appendQuizSupplementalContextSections(sections, scope)
	if err != nil {
		return core.QuizModelRequest{}, err
	}

	sections = appendQuizFileContextSection(
		sections,
		s.projectContext.ProjectDir,
		"curated_quiz",
		"quiz",
		paths.CuratedRelativePath,
	)

	return core.QuizModelRequest{
		ID:              fmt.Sprintf("quiz-draft-spec-%s", paths.ScopeID),
		Action:          core.QuizModelActionDraftSpec,
		Strategy:        "draft_spec",
		Scope:           scope,
		Mode:            state.Mode,
		RuntimeState:    state,
		Payload:         payload,
		ContextSections: sections,
		OutputContract: core.QuizModelOutputContract{
			RequiredFields: payload.OutputContract.RequiredFields,
		},
		CreatedAt: s.utcNow(),
	}, nil
}

func (s service) newQuizModelPayload(
	scope core.QuizScope,
	state core.QuizRuntimeState,
	action core.QuizModelAction,
) (core.QuizModelPayload, error) {
	entries, err := s.quizRepository().LoadRawCaptureEntries(s.ctx, scope)
	if err != nil {
		return core.QuizModelPayload{}, err
	}

	return core.QuizModelPayload{
		Project:         s.quizModelProjectPayload(),
		Quiz:            quizModelQuizPayload(state),
		RecentEntries:   quizModelRecentEntryPayloads(entries, 5),
		CurrentQuestion: quizModelCurrentQuestionPayload(state, entries),
		OutputContract:  core.NewQuizModelOutputContract(action, state.Mode),
	}, nil
}

func (s service) newQuizRuntimeState(scope core.QuizScope, mode core.QuizMode) (core.QuizRuntimeState, error) {
	paths, err := s.quizRepository().Paths(scope)
	if err != nil {
		return core.QuizRuntimeState{}, err
	}

	now := s.utcNow()

	return core.QuizRuntimeState{
		ID:             "quiz-" + paths.ScopeID,
		ProjectID:      scope.ProjectID,
		CaseSlug:       scope.CaseSlug,
		Status:         core.QuizRuntimeStatusIdle,
		Mode:           mode,
		QuestionBudget: core.DefaultQuizQuestionBudget,
		RawPath:        paths.RawRelativePath,
		CuratedPath:    paths.CuratedRelativePath,
		UpdatedAt:      now,
	}, nil
}

func (s service) quizModelProjectPayload() core.QuizModelProjectPayload {
	payload := core.QuizModelProjectPayload{ID: s.projectContext.ProjectID}

	metadata, err := core.NewProjectMetadataService(s.projectContext.ProjectDir).Load()
	if err != nil {
		return payload
	}

	payload.Name = metadata.Name
	payload.Description = metadata.Description

	return payload
}

func (s service) appendQuizSupplementalContextSections(
	sections []core.QuizModelContextSection,
	scope core.QuizScope,
) ([]core.QuizModelContextSection, error) {
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

func quizModelQuizPayload(state core.QuizRuntimeState) core.QuizModelQuizPayload {
	return core.QuizModelQuizPayload{
		ID:         state.ID,
		UseCaseID:  state.CaseSlug,
		EntryCount: state.AnswersRecorded + state.SkipsRecorded,
		Mode:       string(state.Mode),
	}
}

func quizModelRecentEntryPayloads(entries []core.QuizEntry, limit int) []core.QuizModelEntryPayload {
	if limit <= 0 {
		return []core.QuizModelEntryPayload{}
	}

	start := len(entries) - limit
	if start < 0 {
		start = 0
	}

	payloads := make([]core.QuizModelEntryPayload, 0, len(entries[start:]))
	for _, entry := range entries[start:] {
		payloads = append(payloads, core.QuizModelEntryPayload{
			Question: entry.Question,
			Answer:   entry.Answer,
			Status:   string(entry.Status),
			Tags:     entry.Tags,
		})
	}

	return payloads
}

func quizModelCurrentQuestionPayload(
	state core.QuizRuntimeState,
	entries []core.QuizEntry,
) core.QuizModelCurrentQuestionPayload {
	if state.ActiveQuestion != nil {
		return quizModelQuestionPayloadFromQuestion(*state.ActiveQuestion)
	}

	if len(entries) == 0 {
		return core.QuizModelCurrentQuestionPayload{PromptIntent: map[string]any{}}
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

	return core.QuizModelCurrentQuestionPayload{
		Question:     entry.Question,
		Tags:         quizModelEntryTags(entry),
		PromptIntent: promptIntent,
	}
}

func quizModelQuestionPayloadFromQuestion(question core.QuizQuestion) core.QuizModelCurrentQuestionPayload {
	promptIntent := quizModelPromptIntentFromJSON(question.Rationale)
	if len(promptIntent) == 0 {
		promptIntent = map[string]any{
			"metadata": map[string]any{
				"follow_up_depth":        question.Depth - 1,
				"completion_recommended": !question.Required,
			},
		}
	}

	return core.QuizModelCurrentQuestionPayload{
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
	state core.QuizRuntimeState,
	sections []core.QuizModelContextSection,
) core.AcceptanceReadiness {
	content := strings.ToLower(quizContextText(sections))
	reasons := []string{"The quiz context was inspected for acceptance-test readiness."}

	if quizContextLooksNonAcceptanceTestable(content) {
		return core.AcceptanceReadiness{
			Decision:         core.AcceptanceReadinessNotApplicable,
			Confidence:       0.7,
			Reasons:          append(reasons, "The context appears internal, architectural, documentation-only, or refactor-oriented."),
			InferredElements: map[core.AcceptanceReadinessSlot][]string{},
		}
	}

	inferred := quizAcceptanceInferredElements(content)
	missing := quizAcceptanceMissingSlots(inferred)

	if len(missing) == 0 {
		return core.AcceptanceReadiness{
			Decision:           core.AcceptanceReadinessGenerate,
			Confidence:         0.72,
			Reasons:            append(reasons, "The context includes actor, trigger, initial state, outcome, and success-path signals."),
			InferredElements:   inferred,
			CandidateScenarios: []string{"Primary behavior succeeds and produces the expected observable outcome."},
		}
	}

	return core.AcceptanceReadiness{
		Decision:         core.AcceptanceReadinessNeedsMoreInfo,
		Confidence:       0.64,
		Reasons:          append(reasons, "The context has some acceptance-test signals, but key scenario elements are missing."),
		Missing:          missing,
		InferredElements: inferred,
	}
}

func quizContextText(sections []core.QuizModelContextSection) string {
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

func quizAcceptanceInferredElements(content string) map[core.AcceptanceReadinessSlot][]string {
	inferred := map[core.AcceptanceReadinessSlot][]string{}

	if quizContainsAny(content, "user", "operator", "cli", "api client", "worker", "scheduler", "reviewer") {
		inferred[core.AcceptanceReadinessSlotActor] = []string{"An actor is referenced by the source context."}
	}

	if quizContainsAny(content, "command", "request", "event", "runs", "starts", "trigger", "action") {
		inferred[core.AcceptanceReadinessSlotTrigger] = []string{"A trigger is referenced by the source context."}
	}

	if quizContainsAny(content, "given", "existing", "configured", "before", "state", "project", "ticket", "file") {
		inferred[core.AcceptanceReadinessSlotInitialState] = []string{"An initial state is referenced by the source context."}
	}

	if quizContainsAny(content, "output", "report", "status", "response", "error", "file", "written", "created") {
		inferred[core.AcceptanceReadinessSlotObservableOutcome] = []string{"An observable outcome is referenced by the source context."}
	}

	if quizContainsAny(content, "success", "succeeds", "passes", "accepted", "valid", "expected") {
		inferred[core.AcceptanceReadinessSlotSuccessPath] = []string{"A success path is referenced by the source context."}
	}

	if quizContainsAny(content, "failure", "fails", "invalid", "reject", "error", "missing") {
		inferred[core.AcceptanceReadinessSlotFailurePath] = []string{"A failure path is referenced by the source context."}
	}

	if quizContainsAny(content, "out of scope", "non-goal", "boundary", "must not", "do not") {
		inferred[core.AcceptanceReadinessSlotBoundaries] = []string{"A boundary is referenced by the source context."}
	}

	return inferred
}

func quizAcceptanceMissingSlots(inferred map[core.AcceptanceReadinessSlot][]string) []core.AcceptanceReadinessSlot {
	required := []core.AcceptanceReadinessSlot{
		core.AcceptanceReadinessSlotActor,
		core.AcceptanceReadinessSlotTrigger,
		core.AcceptanceReadinessSlotInitialState,
		core.AcceptanceReadinessSlotObservableOutcome,
		core.AcceptanceReadinessSlotSuccessPath,
	}
	missing := []core.AcceptanceReadinessSlot{}

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

func quizModelQuestionTags(question core.QuizQuestion) []string {
	if strings.TrimSpace(question.Topic) == "" {
		return []string{}
	}

	return []string{question.Topic}
}

func quizModelEntryTags(entry core.QuizEntry) []string {
	if len(entry.Tags) > 0 {
		return entry.Tags
	}

	if strings.TrimSpace(entry.Topic) == "" {
		return []string{}
	}

	return []string{entry.Topic}
}

func quizModelStrategy(mode core.QuizMode) string {
	if mode == core.QuizModeFast {
		return core.QuizModelFastStrategy
	}

	return core.QuizModelDeepStrategy
}

func (s service) listQuizRuntimeStates() ([]core.QuizRuntimeState, error) {
	root := core.UserPath(s.projectContext.ProjectDir, filepath.FromSlash(core.QuizRuntimeDirectory))

	entries, err := os.ReadDir(root)
	if errors.Is(err, os.ErrNotExist) {
		return []core.QuizRuntimeState{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("read quiz runtime directory: %w", err)
	}

	states := []core.QuizRuntimeState{}
	repository := s.quizRepository()

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		scope := core.QuizScope{
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

func (s service) quizScope(scope core.QuizScope) core.QuizScope {
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
	sections []core.QuizModelContextSection,
	projectDir string,
	name string,
	layer string,
	path string,
) []core.QuizModelContextSection {
	if !filepath.IsAbs(path) {
		path = filepath.Join(projectDir, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return sections
	}

	return append(sections, core.QuizModelContextSection{
		Name:      name,
		Layer:     layer,
		Authority: filepath.ToSlash(path),
		Content:   string(content),
	})
}

func appendQuizNodeContextSection(
	sections []core.QuizModelContextSection,
	projectDir string,
	node core.WorkspaceNode,
) []core.QuizModelContextSection {
	content, err := os.ReadFile(filepath.Join(projectDir, filepath.FromSlash(node.RelativePath)))
	if err != nil {
		return sections
	}

	name := "project_dump"
	if node.CaseSlug != "" {
		name = "use_case_dump"
	}

	return append(sections, core.QuizModelContextSection{
		Name:      name,
		Layer:     "dump",
		Authority: node.RelativePath,
		Content:   string(content),
	})
}

func quizDumpSourcePaths(sections []core.QuizModelContextSection) []string {
	paths := []string{}

	for _, section := range sections {
		if section.Layer == "dump" {
			paths = append(paths, section.Authority)
		}
	}

	return paths
}

func quizEntryFromState(
	state core.QuizRuntimeState,
	status core.QuizEntryStatus,
	answer string,
	now time.Time,
) core.QuizEntry {
	question := *state.ActiveQuestion

	return core.QuizEntry{
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

func updateQuizStateAfterEntry(state core.QuizRuntimeState, entry core.QuizEntry) core.QuizRuntimeState {
	if entry.Status == core.QuizEntryStatusAnswered {
		state.AnswersRecorded++
	}

	if entry.Status == core.QuizEntryStatusSkipped {
		state.SkipsRecorded++
	}

	state.Status = core.QuizRuntimeStatusIdle
	state.ActiveQuestion = nil
	state.ActiveTopic = ""

	if entry.Status == core.QuizEntryStatusSkipped {
		state.Trail = nil
		state.BranchReturnReason = core.QuizBranchReturnReasonSkipped
	} else if entry.Depth >= core.MaxQuizFollowUpDepth {
		state.Trail = nil
		state.BranchReturnReason = core.QuizBranchReturnReasonMaxDepth
	} else {
		state.BranchReturnReason = ""
	}

	if state.Mode == core.QuizModeDeep {
		state.Candidates = nil
	}

	return state
}

func activateQuizQuestion(
	state core.QuizRuntimeState,
	question core.QuizQuestion,
	now time.Time,
) core.QuizRuntimeState {
	state.Status = core.QuizRuntimeStatusActive
	state.ActiveQuestion = &question
	state.ActiveTopic = question.Topic
	state.Trail = quizTrailForQuestion(state.Trail, question)
	state.BranchReturnReason = ""
	state.QuestionsAsked++
	state.UpdatedAt = now

	return state
}

func activateQuizCandidate(
	state core.QuizRuntimeState,
	candidate core.QuizCandidate,
	now time.Time,
) core.QuizRuntimeState {
	question := core.QuizQuestion{
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

func quizTrailForQuestion(trail []core.QuizTrailNode, question core.QuizQuestion) []core.QuizTrailNode {
	node := core.QuizTrailNode{
		QuestionID: question.ID,
		Prompt:     question.Prompt,
		Topic:      question.Topic,
		Depth:      question.Depth,
	}

	if question.Depth <= 1 {
		return []core.QuizTrailNode{node}
	}

	parentIndex := -1

	for index, current := range trail {
		if current.QuestionID == question.ParentID {
			parentIndex = index

			break
		}
	}

	if parentIndex < 0 {
		return []core.QuizTrailNode{node}
	}

	next := append([]core.QuizTrailNode(nil), trail[:parentIndex+1]...)
	next = append(next, node)

	return next
}
