package app

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"forge.adrianpk.com/chc/internal/core"
)

// QuizViewStatusNotStarted marks a quiz scope that has no active session yet.
const QuizViewStatusNotStarted = "not_started"

// QuizView is the TUI-facing projection of a quiz session or available scope.
type QuizView struct {
	Index                 int
	Title                 string
	Scope                 core.QuizScope
	Status                string
	Mode                  core.QuizMode
	RawPath               string
	CuratedPath           string
	RuntimePath           string
	QuestionBudget        int
	QuestionsAsked        int
	AnswersRecorded       int
	SkipsRecorded         int
	CandidateCount        int
	BranchReturnReason    string
	UpdatedAt             time.Time
	ActiveQuestionSummary string
	TrailSummary          string
	CuratedPreview        string
}

func (s service) quizViewsFromStates(states []core.QuizRuntimeState) ([]QuizView, error) {
	scopes, err := s.quizReadScopes()
	if err != nil {
		return nil, err
	}

	stateByScope := quizStateByScope(states)
	knownScopes := map[string]bool{}

	views := make([]QuizView, 0, len(scopes))
	for _, scope := range scopes {
		knownScopes[quizScopeKey(scope.scope)] = true

		view, err := s.quizViewFromScope(scope, stateByScope)
		if err != nil {
			return nil, err
		}

		views = append(views, view)
	}

	views, err = s.appendRuntimeOnlyQuizViews(views, states, knownScopes)
	if err != nil {
		return nil, err
	}

	for index := range views {
		views[index].Index = index
	}

	return views, nil
}

func (s service) quizViewFromScope(
	scope quizReadScope,
	stateByScope map[string]core.QuizRuntimeState,
) (QuizView, error) {
	state, ok := stateByScope[quizScopeKey(scope.scope)]
	if ok {
		return s.quizViewFromState(scope.title, state)
	}

	return s.notStartedQuizView(scope.title, scope.scope)
}

func (s service) appendRuntimeOnlyQuizViews(
	views []QuizView,
	states []core.QuizRuntimeState,
	knownScopes map[string]bool,
) ([]QuizView, error) {
	for _, state := range states {
		scope := core.QuizScope{ProjectID: state.ProjectID, CaseSlug: state.CaseSlug}
		if knownScopes[quizScopeKey(scope)] {
			continue
		}

		view, err := s.quizViewFromState(quizRuntimeOnlyTitle(state), state)
		if err != nil {
			return nil, err
		}

		views = append(views, view)
	}

	return views, nil
}

func (s service) quizReadScopes() ([]quizReadScope, error) {
	projectTitle := s.projectContext.ProjectID

	projectMetadata, err := core.NewProjectMetadataService(s.projectContext.ProjectDir).Load()
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(projectMetadata.Name) != "" {
		projectTitle = projectMetadata.Name
	}

	scopes := []quizReadScope{{
		title: projectTitle,
		scope: core.QuizScope{ProjectID: s.projectContext.ProjectID},
	}}

	cases, err := core.NewCaseMetadataService(s.projectContext.ProjectDir).LoadAll()
	if err != nil {
		return nil, err
	}

	for _, metadata := range cases {
		title := metadata.Name
		if strings.TrimSpace(title) == "" {
			title = metadata.Slug
		}

		scopes = append(scopes, quizReadScope{
			title: title,
			scope: core.QuizScope{
				ProjectID: s.projectContext.ProjectID,
				CaseSlug:  metadata.Slug,
			},
		})
	}

	sort.SliceStable(scopes[1:], func(i, j int) bool {
		left := strings.ToLower(scopes[i+1].title)

		right := strings.ToLower(scopes[j+1].title)
		if left == right {
			return scopes[i+1].scope.CaseSlug < scopes[j+1].scope.CaseSlug
		}

		return left < right
	})

	return scopes, nil
}

func (s service) quizViewFromState(title string, state core.QuizRuntimeState) (QuizView, error) {
	scope := core.QuizScope{
		ProjectID: state.ProjectID,
		CaseSlug:  state.CaseSlug,
	}

	paths, err := s.quizRepository().Paths(scope)
	if err != nil {
		return QuizView{}, err
	}

	return QuizView{
		Title:                 title,
		Scope:                 scope,
		Status:                string(state.Status),
		Mode:                  state.Mode,
		RawPath:               paths.RawRelativePath,
		CuratedPath:           paths.CuratedRelativePath,
		RuntimePath:           paths.RuntimeRelativePath,
		QuestionBudget:        state.QuestionBudget,
		QuestionsAsked:        state.QuestionsAsked,
		AnswersRecorded:       state.AnswersRecorded,
		SkipsRecorded:         state.SkipsRecorded,
		CandidateCount:        len(state.Candidates),
		BranchReturnReason:    state.BranchReturnReason,
		UpdatedAt:             state.UpdatedAt,
		ActiveQuestionSummary: quizActiveQuestionSummary(state),
		TrailSummary:          quizTrailSummary(state),
		CuratedPreview:        s.quizCuratedPreview(paths.CuratedPath),
	}, nil
}

func (s service) notStartedQuizView(title string, scope core.QuizScope) (QuizView, error) {
	paths, err := s.quizRepository().Paths(scope)
	if err != nil {
		return QuizView{}, err
	}

	return QuizView{
		Title:          title,
		Scope:          scope,
		Status:         QuizViewStatusNotStarted,
		Mode:           core.QuizModeDeep,
		RawPath:        paths.RawRelativePath,
		CuratedPath:    paths.CuratedRelativePath,
		RuntimePath:    paths.RuntimeRelativePath,
		QuestionBudget: core.DefaultQuizQuestionBudget,
		CuratedPreview: s.quizCuratedPreview(paths.CuratedPath),
	}, nil
}

func (s service) quizCuratedPreview(path string) string {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return ""
	}

	if err != nil {
		return fmt.Sprintf("Unable to read curated quiz: %s", err)
	}

	lines := []string{}

	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "---" || strings.Contains(line, ": ") {
			continue
		}

		lines = append(lines, line)
		if len(lines) == 4 {
			break
		}
	}

	return strings.Join(lines, "\n")
}

func quizActiveQuestionSummary(state core.QuizRuntimeState) string {
	if state.ActiveQuestion == nil {
		return ""
	}

	return state.ActiveQuestion.Prompt
}

func quizTrailSummary(state core.QuizRuntimeState) string {
	if len(state.Trail) == 0 {
		return ""
	}

	segments := make([]string, 0, len(state.Trail))
	for _, node := range state.Trail {
		label := node.Topic
		if strings.TrimSpace(label) == "" {
			label = node.QuestionID
		}

		segments = append(segments, label)
	}

	return strings.Join(segments, " > ")
}

func quizStateByScope(states []core.QuizRuntimeState) map[string]core.QuizRuntimeState {
	stateByScope := map[string]core.QuizRuntimeState{}
	for _, state := range states {
		stateByScope[quizScopeKey(core.QuizScope{
			ProjectID: state.ProjectID,
			CaseSlug:  state.CaseSlug,
		})] = state
	}

	return stateByScope
}

func quizRuntimeOnlyTitle(state core.QuizRuntimeState) string {
	if strings.TrimSpace(state.CaseSlug) != "" {
		return state.CaseSlug
	}

	return state.ProjectID
}

func quizScopeKey(scope core.QuizScope) string {
	return scope.ProjectID + "/" + scope.CaseSlug
}

type quizReadScope struct {
	title string
	scope core.QuizScope
}
