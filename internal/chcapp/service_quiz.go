package chcapp

import (
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

	return QuizListResult{Message: "Quizzes loaded.", States: states}, nil
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
	return s.stopQuiz(s.quizScope(request.Scope), chc.QuizRuntimeStatusDone, "Quiz completed.")
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

func (s service) generateNextQuizQuestion(state chc.QuizRuntimeState) (chc.QuizRuntimeState, error) {
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
	state.ActiveQuestion = &response.Question
	state.QuestionsAsked++
	state.UpdatedAt = response.CreatedAt.UTC()

	return state, nil
}

func (s service) newQuizModelRequest(state chc.QuizRuntimeState) (chc.QuizModelRequest, error) {
	sections, err := s.quizContextSections(chc.QuizScope{
		ProjectID: state.ProjectID,
		CaseSlug:  state.CaseSlug,
	})
	if err != nil {
		return chc.QuizModelRequest{}, err
	}

	return chc.QuizModelRequest{
		ID:              fmt.Sprintf("quiz-request-%d", state.QuestionsAsked+1),
		Action:          chc.QuizModelActionNextQuestion,
		Strategy:        string(state.Mode),
		Scope:           chc.QuizScope{ProjectID: state.ProjectID, CaseSlug: state.CaseSlug},
		Mode:            state.Mode,
		RuntimeState:    state,
		ContextSections: sections,
		OutputContract: chc.QuizModelOutputContract{
			RequiredFields: []string{"id", "prompt", "topic", "depth", "trunk_id"},
		},
		CreatedAt: s.utcNow(),
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

func (s service) quizContextSections(scope chc.QuizScope) ([]chc.QuizModelContextSection, error) {
	sections := []chc.QuizModelContextSection{}
	sections = appendQuizFileContextSection(sections, s.projectContext.ProjectDir, "agent_protocol", "protocol", "AGENTS.md")
	sections = appendQuizFileContextSection(sections, s.projectContext.ProjectDir, "project", "project", chc.ProjectMetadataPath(""))

	if scope.CaseSlug != "" {
		sections = appendQuizFileContextSection(
			sections,
			s.projectContext.ProjectDir,
			"use_case",
			"use_case",
			chc.CaseMetadataPath("", scope.CaseSlug),
		)
	}

	dumpSections, err := s.quizDumpContextSections(scope)
	if err != nil {
		return nil, err
	}

	sections = append(sections, dumpSections...)
	if len(sections) == 0 {
		sections = append(sections, chc.QuizModelContextSection{
			Name:      "project",
			Layer:     "project",
			Authority: "project_directory",
			Content:   "project_id: " + scope.ProjectID,
		})
	}

	return sections, nil
}

func (s service) quizDumpContextSections(scope chc.QuizScope) ([]chc.QuizModelContextSection, error) {
	index, err := s.workspaceIndex()
	if err != nil {
		return nil, err
	}

	sections := []chc.QuizModelContextSection{}

	for _, nodeID := range index.ByKind[chc.WorkspaceNodeDump] {
		node := index.NodesByID[nodeID]
		if node.CaseSlug != "" && node.CaseSlug != scope.CaseSlug {
			continue
		}

		if node.CaseSlug == "" || scope.CaseSlug == "" || node.CaseSlug == scope.CaseSlug {
			sections = appendQuizNodeContextSection(sections, s.projectContext.ProjectDir, node)
		}
	}

	return sections, nil
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

func quizEntryFromState(
	state chc.QuizRuntimeState,
	status chc.QuizEntryStatus,
	answer string,
	now time.Time,
) chc.QuizEntry {
	question := *state.ActiveQuestion

	return chc.QuizEntry{
		ID:         fmt.Sprintf("entry-%03d", state.QuestionsAsked),
		QuestionID: question.ID,
		Question:   question.Prompt,
		Answer:     strings.TrimSpace(answer),
		Status:     status,
		Topic:      question.Topic,
		Depth:      question.Depth,
		ParentID:   question.ParentID,
		Mode:       state.Mode,
		CreatedAt:  question.CreatedAt,
		AnsweredAt: now,
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

	return state
}
