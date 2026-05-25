package chcapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chc/internal/chc"
)

type QuizListResult struct {
	Message string
	States  []chc.QuizRuntimeState
	Views   []QuizView
}

type QuizOperationResult struct {
	Message string
	State   chc.QuizRuntimeState
}

type QuizStartRequest struct {
	Scope chc.QuizScope
	Mode  chc.QuizMode
}

type QuizResumeRequest struct {
	Scope chc.QuizScope
}

type QuizAnswerRequest struct {
	Scope  chc.QuizScope
	Answer string
}

type QuizSkipRequest struct {
	Scope chc.QuizScope
}

type QuizCompleteRequest struct {
	Scope chc.QuizScope
}

type QuizCancelRequest struct {
	Scope chc.QuizScope
}

type QuizModeRequest struct {
	Scope chc.QuizScope
	Mode  chc.QuizMode
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
		mode = chc.QuizModeDeep
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

	return s.recordQuizOutcome(s.quizScope(request.Scope), chc.QuizEntryStatusAnswered, answer)
}

func (s service) SkipQuiz(request QuizSkipRequest) (QuizOperationResult, error) {
	return s.recordQuizOutcome(s.quizScope(request.Scope), chc.QuizEntryStatusSkipped, "")
}

func (s service) CompleteQuiz(request QuizCompleteRequest) (QuizOperationResult, error) {
	return s.completeQuiz(s.quizScope(request.Scope))
}

func (s service) CancelQuiz(request QuizCancelRequest) (QuizOperationResult, error) {
	return s.stopQuiz(s.quizScope(request.Scope), chc.QuizRuntimeStatusCancelled, "Quiz cancelled.")
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

	if state.Status == chc.QuizRuntimeStatusIdle && state.Mode == chc.QuizModeFast {
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
	scope chc.QuizScope,
	status chc.QuizEntryStatus,
	answer string,
) (QuizOperationResult, error) {
	repository := s.quizRepository()

	state, err := repository.LoadRuntimeState(s.ctx, scope)
	if err != nil {
		return QuizOperationResult{}, err
	}

	if state.Status != chc.QuizRuntimeStatusActive || state.ActiveQuestion == nil {
		return QuizOperationResult{}, fmt.Errorf("quiz has no active question")
	}

	entry := quizEntryFromState(state, status, answer, s.utcNow())

	_, err = repository.AppendRawCapture(s.ctx, state, entry)
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
		state.Status = chc.QuizRuntimeStatusDone
	}

	state.UpdatedAt = s.utcNow()

	_, err = repository.SaveRuntimeState(s.ctx, state)
	if err != nil {
		return QuizOperationResult{}, err
	}

	return QuizOperationResult{Message: "Quiz answer recorded.", State: state}, nil
}

func (s service) stopQuiz(
	scope chc.QuizScope,
	status chc.QuizRuntimeStatus,
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

func (s service) completeQuiz(scope chc.QuizScope) (QuizOperationResult, error) {
	repository := s.quizRepository()

	state, err := repository.LoadRuntimeState(s.ctx, scope)
	if err != nil {
		return QuizOperationResult{}, err
	}

	state.Status = chc.QuizRuntimeStatusDone
	state.ActiveQuestion = nil
	state.Candidates = nil
	state.UpdatedAt = s.utcNow()

	entries, err := repository.LoadRawCaptureEntries(s.ctx, scope)
	if err != nil {
		return QuizOperationResult{}, err
	}

	paths, err := repository.WriteCuratedNotes(s.ctx, chc.CurateQuizEntries(scope, entries, state.UpdatedAt))
	if err != nil {
		return QuizOperationResult{}, err
	}

	err = s.writeQuizDraftSpec(state, paths)
	if err != nil {
		return QuizOperationResult{}, err
	}

	_, err = repository.SaveRuntimeState(s.ctx, state)
	if err != nil {
		return QuizOperationResult{}, err
	}

	return QuizOperationResult{Message: "Quiz completed.", State: state}, nil
}

func (s service) generateNextQuizQuestion(state chc.QuizRuntimeState) (chc.QuizRuntimeState, error) {
	if len(state.Candidates) > 0 {
		state = activateQuizCandidate(state, state.Candidates[0], s.utcNow())
		state.Candidates = state.Candidates[1:]

		return state, nil
	}

	request, err := s.newQuizModelRequest(state)
	if err != nil {
		return chc.QuizRuntimeState{}, err
	}

	response, err := chc.InvokeQuizModel(s.ctx, s.quizModel, request)
	if err != nil {
		return chc.QuizRuntimeState{}, err
	}

	if response.Status == chc.QuizModelResponseBlocked {
		state.Status = chc.QuizRuntimeStatusBlocked
		state.ActiveQuestion = nil
		state.UpdatedAt = response.CreatedAt.UTC()

		return state, nil
	}

	state.Status = chc.QuizRuntimeStatusActive
	state = activateQuizQuestion(state, response.Question, response.CreatedAt.UTC())
	state.Candidates = response.Candidates

	return state, nil
}

func (s service) newQuizModelRequest(state chc.QuizRuntimeState) (chc.QuizModelRequest, error) {
	scope := chc.QuizScope{
		ProjectID: state.ProjectID,
		CaseSlug:  state.CaseSlug,
	}

	payload, err := s.newQuizModelPayload(scope, state, chc.QuizModelActionNextQuestion)
	if err != nil {
		return chc.QuizModelRequest{}, err
	}

	sections, err := chc.QuizModelContextSectionsForPayload(payload)
	if err != nil {
		return chc.QuizModelRequest{}, err
	}

	sections, err = s.appendQuizSupplementalContextSections(sections, scope)
	if err != nil {
		return chc.QuizModelRequest{}, err
	}

	return chc.QuizModelRequest{
		ID:              fmt.Sprintf("quiz-request-%d", state.QuestionsAsked+1),
		Action:          chc.QuizModelActionNextQuestion,
		Strategy:        quizModelStrategy(state.Mode),
		Scope:           scope,
		Mode:            state.Mode,
		RuntimeState:    state,
		Payload:         payload,
		ContextSections: sections,
		OutputContract: chc.QuizModelOutputContract{
			RequiredFields: payload.OutputContract.RequiredFields,
		},
		CreatedAt: s.utcNow(),
	}, nil
}

func (s service) writeQuizDraftSpec(
	state chc.QuizRuntimeState,
	paths chc.QuizArtifactPaths,
) error {
	request, err := s.newQuizDraftSpecModelRequest(state, paths)
	if err != nil {
		return err
	}

	response, err := chc.InvokeQuizModel(s.ctx, s.quizModel, request)
	if err != nil {
		return err
	}

	spec := response.DraftSpec
	spec.Scope = chc.QuizScope{ProjectID: state.ProjectID, CaseSlug: state.CaseSlug}
	spec.SourceQuizPath = paths.CuratedRelativePath
	spec.SourceRawPath = paths.RawRelativePath
	spec.SourceDumpPaths = quizDumpSourcePaths(request.ContextSections)
	spec.SourceRequestID = request.ID
	spec.SourceResponseID = response.ID
	spec.CreatedAt = response.CreatedAt.UTC()

	_, err = chc.WriteQuizDraftSpec(s.ctx, s.projectContext.ProjectDir, spec)

	return err
}

func (s service) newQuizDraftSpecModelRequest(
	state chc.QuizRuntimeState,
	paths chc.QuizArtifactPaths,
) (chc.QuizModelRequest, error) {
	scope := chc.QuizScope{
		ProjectID: state.ProjectID,
		CaseSlug:  state.CaseSlug,
	}

	payload, err := s.newQuizModelPayload(scope, state, chc.QuizModelActionDraftSpec)
	if err != nil {
		return chc.QuizModelRequest{}, err
	}

	sections, err := chc.QuizModelContextSectionsForPayload(payload)
	if err != nil {
		return chc.QuizModelRequest{}, err
	}

	sections, err = s.appendQuizSupplementalContextSections(sections, scope)
	if err != nil {
		return chc.QuizModelRequest{}, err
	}

	sections = appendQuizFileContextSection(
		sections,
		s.projectContext.ProjectDir,
		"curated_quiz",
		"quiz",
		paths.CuratedRelativePath,
	)

	return chc.QuizModelRequest{
		ID:              fmt.Sprintf("quiz-draft-spec-%s", paths.ScopeID),
		Action:          chc.QuizModelActionDraftSpec,
		Strategy:        "draft_spec",
		Scope:           scope,
		Mode:            state.Mode,
		RuntimeState:    state,
		Payload:         payload,
		ContextSections: sections,
		OutputContract: chc.QuizModelOutputContract{
			RequiredFields: payload.OutputContract.RequiredFields,
		},
		CreatedAt: s.utcNow(),
	}, nil
}

func (s service) newQuizModelPayload(
	scope chc.QuizScope,
	state chc.QuizRuntimeState,
	action chc.QuizModelAction,
) (chc.QuizModelPayload, error) {
	entries, err := s.quizRepository().LoadRawCaptureEntries(s.ctx, scope)
	if err != nil {
		return chc.QuizModelPayload{}, err
	}

	return chc.QuizModelPayload{
		Project:         s.quizModelProjectPayload(),
		Quiz:            quizModelQuizPayload(state),
		RecentEntries:   quizModelRecentEntryPayloads(entries, 5),
		CurrentQuestion: quizModelCurrentQuestionPayload(state, entries),
		OutputContract:  chc.NewQuizModelOutputContract(action, state.Mode),
	}, nil
}

func (s service) newQuizRuntimeState(scope chc.QuizScope, mode chc.QuizMode) (chc.QuizRuntimeState, error) {
	paths, err := s.quizRepository().Paths(scope)
	if err != nil {
		return chc.QuizRuntimeState{}, err
	}

	now := s.utcNow()

	return chc.QuizRuntimeState{
		ID:             "quiz-" + paths.ScopeID,
		ProjectID:      scope.ProjectID,
		CaseSlug:       scope.CaseSlug,
		Status:         chc.QuizRuntimeStatusIdle,
		Mode:           mode,
		QuestionBudget: chc.DefaultQuizQuestionBudget,
		RawPath:        paths.RawRelativePath,
		CuratedPath:    paths.CuratedRelativePath,
		UpdatedAt:      now,
	}, nil
}

func (s service) quizModelProjectPayload() chc.QuizModelProjectPayload {
	payload := chc.QuizModelProjectPayload{ID: s.projectContext.ProjectID}

	metadata, err := chc.NewProjectMetadataService(s.projectContext.ProjectDir).Load()
	if err != nil {
		return payload
	}

	payload.Name = metadata.Name
	payload.Description = metadata.Description

	return payload
}

func (s service) appendQuizSupplementalContextSections(
	sections []chc.QuizModelContextSection,
	scope chc.QuizScope,
) ([]chc.QuizModelContextSection, error) {
	sections = appendQuizFileContextSection(sections, s.projectContext.ProjectDir, "agent_protocol", "protocol", "AGENTS.md")
	sections = appendQuizFileContextSection(sections, s.projectContext.ProjectDir, "project_document", "project", chc.ProjectMetadataPath(""))

	if scope.CaseSlug != "" {
		sections = appendQuizFileContextSection(
			sections,
			s.projectContext.ProjectDir,
			"use_case_document",
			"use_case",
			chc.CaseMetadataPath("", scope.CaseSlug),
		)
	}

	index, err := s.workspaceIndex()
	if err != nil {
		return nil, err
	}

	for _, nodeID := range index.ByKind[chc.WorkspaceNodeDump] {
		node := index.NodesByID[nodeID]
		if node.CaseSlug != "" && node.CaseSlug != scope.CaseSlug {
			continue
		}

		sections = appendQuizNodeContextSection(sections, s.projectContext.ProjectDir, node)
	}

	return sections, nil
}

func quizModelQuizPayload(state chc.QuizRuntimeState) chc.QuizModelQuizPayload {
	return chc.QuizModelQuizPayload{
		ID:         state.ID,
		UseCaseID:  state.CaseSlug,
		EntryCount: state.AnswersRecorded + state.SkipsRecorded,
		Mode:       string(state.Mode),
	}
}

func quizModelRecentEntryPayloads(entries []chc.QuizEntry, limit int) []chc.QuizModelEntryPayload {
	if limit <= 0 {
		return []chc.QuizModelEntryPayload{}
	}

	start := len(entries) - limit
	if start < 0 {
		start = 0
	}

	payloads := make([]chc.QuizModelEntryPayload, 0, len(entries[start:]))
	for _, entry := range entries[start:] {
		payloads = append(payloads, chc.QuizModelEntryPayload{
			Question: entry.Question,
			Answer:   entry.Answer,
			Status:   string(entry.Status),
			Tags:     entry.Tags,
		})
	}

	return payloads
}

func quizModelCurrentQuestionPayload(
	state chc.QuizRuntimeState,
	entries []chc.QuizEntry,
) chc.QuizModelCurrentQuestionPayload {
	if state.ActiveQuestion != nil {
		return quizModelQuestionPayloadFromQuestion(*state.ActiveQuestion)
	}

	if len(entries) == 0 {
		return chc.QuizModelCurrentQuestionPayload{PromptIntent: map[string]any{}}
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

	return chc.QuizModelCurrentQuestionPayload{
		Question:     entry.Question,
		Tags:         quizModelEntryTags(entry),
		PromptIntent: promptIntent,
	}
}

func quizModelQuestionPayloadFromQuestion(question chc.QuizQuestion) chc.QuizModelCurrentQuestionPayload {
	promptIntent := quizModelPromptIntentFromJSON(question.Rationale)
	if len(promptIntent) == 0 {
		promptIntent = map[string]any{
			"metadata": map[string]any{
				"follow_up_depth":        question.Depth - 1,
				"completion_recommended": !question.Required,
			},
		}
	}

	return chc.QuizModelCurrentQuestionPayload{
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

func quizModelQuestionTags(question chc.QuizQuestion) []string {
	if strings.TrimSpace(question.Topic) == "" {
		return []string{}
	}

	return []string{question.Topic}
}

func quizModelEntryTags(entry chc.QuizEntry) []string {
	if len(entry.Tags) > 0 {
		return entry.Tags
	}

	if strings.TrimSpace(entry.Topic) == "" {
		return []string{}
	}

	return []string{entry.Topic}
}

func quizModelStrategy(mode chc.QuizMode) string {
	if mode == chc.QuizModeFast {
		return chc.QuizModelFastStrategy
	}

	return chc.QuizModelDeepStrategy
}

func (s service) listQuizRuntimeStates() ([]chc.QuizRuntimeState, error) {
	root := filepath.Join(s.projectContext.ProjectDir, chc.DirectoryName, filepath.FromSlash(chc.QuizRuntimeDirectory))

	entries, err := os.ReadDir(root)
	if errors.Is(err, os.ErrNotExist) {
		return []chc.QuizRuntimeState{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("read quiz runtime directory: %w", err)
	}

	states := []chc.QuizRuntimeState{}
	repository := s.quizRepository()

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		scope := chc.QuizScope{
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

func (s service) quizRepository() chc.QuizArtifactRepository {
	return chc.NewQuizArtifactRepository(s.projectContext.ProjectDir)
}

func (s service) quizScope(scope chc.QuizScope) chc.QuizScope {
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
	sections []chc.QuizModelContextSection,
	projectDir string,
	name string,
	layer string,
	path string,
) []chc.QuizModelContextSection {
	if !filepath.IsAbs(path) {
		path = filepath.Join(projectDir, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return sections
	}

	return append(sections, chc.QuizModelContextSection{
		Name:      name,
		Layer:     layer,
		Authority: filepath.ToSlash(path),
		Content:   string(content),
	})
}

func appendQuizNodeContextSection(
	sections []chc.QuizModelContextSection,
	projectDir string,
	node chc.WorkspaceNode,
) []chc.QuizModelContextSection {
	content, err := os.ReadFile(filepath.Join(projectDir, filepath.FromSlash(node.RelativePath)))
	if err != nil {
		return sections
	}

	name := "project_dump"
	if node.CaseSlug != "" {
		name = "use_case_dump"
	}

	return append(sections, chc.QuizModelContextSection{
		Name:      name,
		Layer:     "dump",
		Authority: node.RelativePath,
		Content:   string(content),
	})
}

func quizDumpSourcePaths(sections []chc.QuizModelContextSection) []string {
	paths := []string{}

	for _, section := range sections {
		if section.Layer == "dump" {
			paths = append(paths, section.Authority)
		}
	}

	return paths
}

func quizEntryFromState(
	state chc.QuizRuntimeState,
	status chc.QuizEntryStatus,
	answer string,
	now time.Time,
) chc.QuizEntry {
	question := *state.ActiveQuestion

	return chc.QuizEntry{
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

func updateQuizStateAfterEntry(state chc.QuizRuntimeState, entry chc.QuizEntry) chc.QuizRuntimeState {
	if entry.Status == chc.QuizEntryStatusAnswered {
		state.AnswersRecorded++
	}

	if entry.Status == chc.QuizEntryStatusSkipped {
		state.SkipsRecorded++
	}

	state.Status = chc.QuizRuntimeStatusIdle
	state.ActiveQuestion = nil
	state.ActiveTopic = ""

	if entry.Status == chc.QuizEntryStatusSkipped {
		state.Trail = nil
		state.BranchReturnReason = chc.QuizBranchReturnReasonSkipped
	} else if entry.Depth >= chc.MaxQuizFollowUpDepth {
		state.Trail = nil
		state.BranchReturnReason = chc.QuizBranchReturnReasonMaxDepth
	} else {
		state.BranchReturnReason = ""
	}

	if state.Mode == chc.QuizModeDeep {
		state.Candidates = nil
	}

	return state
}

func activateQuizQuestion(
	state chc.QuizRuntimeState,
	question chc.QuizQuestion,
	now time.Time,
) chc.QuizRuntimeState {
	state.Status = chc.QuizRuntimeStatusActive
	state.ActiveQuestion = &question
	state.ActiveTopic = question.Topic
	state.Trail = quizTrailForQuestion(state.Trail, question)
	state.BranchReturnReason = ""
	state.QuestionsAsked++
	state.UpdatedAt = now

	return state
}

func activateQuizCandidate(
	state chc.QuizRuntimeState,
	candidate chc.QuizCandidate,
	now time.Time,
) chc.QuizRuntimeState {
	question := chc.QuizQuestion{
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

func quizTrailForQuestion(trail []chc.QuizTrailNode, question chc.QuizQuestion) []chc.QuizTrailNode {
	node := chc.QuizTrailNode{
		QuestionID: question.ID,
		Prompt:     question.Prompt,
		Topic:      question.Topic,
		Depth:      question.Depth,
	}

	if question.Depth <= 1 {
		return []chc.QuizTrailNode{node}
	}

	parentIndex := -1

	for index, current := range trail {
		if current.QuestionID == question.ParentID {
			parentIndex = index

			break
		}
	}

	if parentIndex < 0 {
		return []chc.QuizTrailNode{node}
	}

	next := append([]chc.QuizTrailNode(nil), trail[:parentIndex+1]...)
	next = append(next, node)

	return next
}
