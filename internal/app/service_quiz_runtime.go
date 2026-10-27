package app

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

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
