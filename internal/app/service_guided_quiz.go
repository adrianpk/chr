package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type guidedGenerationContext struct{ session core.AuthoringSession }

// ConductGuidedQuiz applies a scoped quiz start, recovery or current answer.
func (s service) ConductGuidedQuiz(ctx context.Context, command core.GuidedUseCaseCommand) (core.AuthoringSession, error) {
	if !guidedQuizAction(command.Action) {
		return core.AuthoringSession{}, errors.New("unsupported guided quiz command")
	}

	scope, err := s.guidedUseCaseScope(command.RunID, command.Namespace)
	if err != nil {
		return core.AuthoringSession{}, err
	}

	command.GuidedUseCaseScope = scope

	return core.ApplyGuidedUseCaseStep(ctx, s.projectContext.ProjectDir, command, func(ctx context.Context, session *core.AuthoringSession) error {
		bounded := s
		bounded.ctx = ctx
		bounded.guidedContext = &guidedGenerationContext{session: *session}

		return bounded.conductGuidedQuiz(ctx, session)
	})
}

func (s service) conductGuidedQuiz(ctx context.Context, session *core.AuthoringSession) error {
	if session.Guided.Stage != "quiz" {
		return workflowstate.ErrConflict
	}

	checkpoint, err := guidedDumpCheckpoint(*session)
	if err != nil {
		return err
	}

	digest, err := core.GuidedCheckpointDigest(checkpoint)
	if err != nil {
		return err
	}

	if session.Guided.PendingOperation.QuizRequest == nil {
		request, err := s.prepareGuidedQuizAction(ctx, *session, digest)
		if err != nil {
			return err
		}

		session.Guided.PendingOperation.QuizRequest = &request
		if err := core.RetainGuidedUseCaseProgress(ctx, s.projectContext.ProjectDir, session); err != nil {
			return err
		}
	}

	result, err := s.applyQuizAction(*session.Guided.PendingOperation.QuizRequest)
	if err != nil {
		return err
	}

	if s.afterGuidedQuiz != nil {
		if err := s.afterGuidedQuiz(); err != nil {
			return err
		}
	}

	return s.finishGuidedQuizResult(ctx, session, result.State, digest)
}

func (s service) finishGuidedQuizResult(ctx context.Context, session *core.AuthoringSession, state quizcore.RuntimeState, digest string) error {
	session.Guided.Quiz = &core.GuidedQuizState{SessionID: state.SessionID, Revision: state.Revision, InputDigest: digest}
	session.Guided.Status = "waiting_input"

	session.Guided.Blocker = "Question generation needs corrected setup or more context. Retry clarification."
	if state.ContinuationPending {
		session.Guided.Status = "blocked"
	}

	if state.ActiveQuestion != nil {
		session.Guided.Quiz.QuestionID = state.ActiveQuestion.ID
		session.Guided.Status = "waiting_input"
	}

	if state.Status == quizcore.RuntimeStatusDone {
		return core.SealGuidedQuiz(ctx, s.projectContext.ProjectDir, session)
	}

	return nil
}

func (s service) prepareGuidedQuizAction(ctx context.Context, session core.AuthoringSession, digest string) (core.QuizActionRequest, error) {
	scope := quizcore.Scope{ProjectID: session.ProjectID, Namespace: session.Namespace, CaseSlug: session.Guided.CaseSlug}

	observed, err := s.quizRepository().Observe(ctx, scope)
	if err != nil {
		return core.QuizActionRequest{}, err
	}

	operation := session.Guided.PendingOperation
	action := "resume"

	if session.Guided.Quiz == nil {
		if observed.Exists {
			return core.QuizActionRequest{}, errors.New("guided quiz cannot adopt an existing manual session")
		}

		action = "start"
	} else if session.Guided.Quiz.SessionID != observed.State.SessionID || session.Guided.Quiz.Revision != observed.Revision || session.Guided.Quiz.InputDigest != digest {
		return core.QuizActionRequest{}, workflowstate.ErrConflict
	}

	binding := core.QuizActionBinding{SessionID: observed.State.SessionID, ExpectedRevision: observed.Revision, OperationID: "guided-" + operation.Digest}
	if observed.State.ActiveQuestion != nil {
		binding.QuestionID = observed.State.ActiveQuestion.ID
	}

	action, mode := guidedQuizNativeInput(operation, action, observed.State.Mode)

	return core.QuizActionRequest{Scope: scope, Binding: binding, Action: action, Answer: strings.TrimSpace(operation.Content), Mode: mode}, nil
}

func guidedDumpCheckpoint(session core.AuthoringSession) (core.GuidedStageCheckpoint, error) {
	for _, checkpoint := range session.Guided.Checkpoints {
		if checkpoint.Stage == "dump" && len(checkpoint.Artifacts) > 0 {
			return checkpoint, nil
		}
	}

	return core.GuidedStageCheckpoint{}, errors.New("guided quiz requires a retained dump checkpoint")
}

func (s service) guidedQuizModelRequest(state quizcore.RuntimeState, entries []quizcore.Entry) (quizmodel.Request, error) {
	session := s.guidedContext.session

	sections, err := s.guidedCommittedSections(session, "dump")
	if err != nil {
		return quizmodel.Request{}, err
	}

	history, err := json.Marshal(entries)
	if err != nil {
		return quizmodel.Request{}, err
	}

	if len(history) > workflowstate.FileLimit {
		return quizmodel.Request{}, errors.New("guided clarification history exceeds its bound")
	}

	sections = append(sections, quizmodel.ContextSection{Name: "answered clarification", Layer: quizmodel.ContextLayerQuiz, Authority: "canonical owner", Content: string(history)})
	payload := s.quizModelPayloadWithEntries(state, quizmodel.ActionNextQuestion, entries)
	payload.Project = quizmodel.ProjectPayload{ID: session.ProjectID, Name: session.Guided.CaseName}

	coreSections, err := quizmodel.ContextSectionsForPayload(payload)
	if err != nil {
		return quizmodel.Request{}, err
	}

	return quizmodel.Request{
		ID: fmt.Sprintf("guided-%s-%d", session.ID, state.QuestionsAsked+1), Action: quizmodel.ActionNextQuestion, Strategy: quizModelStrategy(state.Mode),
		Scope: quizcore.Scope{ProjectID: session.ProjectID, Namespace: session.Namespace, CaseSlug: session.Guided.CaseSlug}, Mode: state.Mode,
		RuntimeState: state, Payload: payload, ContextSections: append(coreSections, sections...),
		OutputContract: quizmodel.OutputContract{RequiredFields: payload.OutputContract.RequiredFields}, CreatedAt: s.utcNow(),
	}, nil
}

func (s service) guidedCommittedSections(session core.AuthoringSession, stage string) ([]quizmodel.ContextSection, error) {
	sections := []quizmodel.ContextSection{}

	for _, checkpoint := range session.Guided.Checkpoints {
		if checkpoint.Stage != stage {
			continue
		}

		for _, artifact := range checkpoint.Artifacts {
			content, err := core.ReadGuidedCommittedArtifact(s.ctx, s.projectContext.ProjectDir, checkpoint, artifact.Path)
			if err != nil {
				return nil, err
			}

			if strings.HasSuffix(artifact.Path, "/artifact.md") {
				sections = append(sections, quizmodel.ContextSection{Layer: quizmodel.ContextLayerProject, Name: artifact.Path, Content: string(content), Authority: "committed", Metadata: map[string]string{"repository": checkpoint.Repository, "commit": checkpoint.Commit, "path": artifact.Path, "hash": artifact.Hash}})
			}
		}
	}

	if len(sections) == 0 {
		return nil, errors.New("committed guided input is unavailable")
	}

	return sections, nil
}

func guidedQuizAction(action string) bool {
	switch action {
	case "quiz", "answer", "quiz-mode", "quiz-skip", "quiz-done":
		return true
	default:
		return false
	}
}

func guidedQuizNativeInput(operation *core.GuidedUseCaseOperation, action string, mode quizcore.Mode) (string, quizcore.Mode) {
	switch operation.Action {
	case "answer":
		action = "answer"
	case "quiz-mode", "quiz-skip", "quiz-done":
		action = strings.TrimPrefix(operation.Action, "quiz-")
	}

	switch action {
	case "start":
		mode = quizcore.ModeDeep
	case "mode":
		mode = quizcore.Mode(operation.Content)
	}

	return action, mode
}
