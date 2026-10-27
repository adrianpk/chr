package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// GenerateGuidedSpec generates, saves and approves only the current run's specification.
func (s service) GenerateGuidedSpec(ctx context.Context, command core.GuidedUseCaseCommand) (core.AuthoringSession, error) {
	if command.Action != "spec" {
		return core.AuthoringSession{}, errors.New("unsupported guided specification command")
	}

	scope, err := s.guidedUseCaseScope(command.RunID, command.Namespace)
	if err != nil {
		return core.AuthoringSession{}, err
	}

	command.GuidedUseCaseScope = scope

	return core.ApplyGuidedUseCaseStep(ctx, s.projectContext.ProjectDir, command, func(ctx context.Context, session *core.AuthoringSession) error {
		bounded := s
		bounded.ctx = ctx

		return bounded.generateGuidedSpec(ctx, session)
	})
}

func (s service) generateGuidedSpec(ctx context.Context, session *core.AuthoringSession) error {
	if session.Guided.Stage != "specification" {
		return workflowstate.ErrConflict
	}

	operation := session.Guided.PendingOperation
	if operation.PreparedSpec == nil {
		prepared, err := s.prepareGuidedSpec(ctx, *session)
		if err != nil {
			return err
		}

		operation.PreparedSpec = &prepared

		if err := core.RetainGuidedUseCaseProgress(ctx, s.projectContext.ProjectDir, session); err != nil {
			return err
		}
	}

	prepared := operation.PreparedSpec
	request := core.SpecMutationRequest{OperationID: "guided-" + operation.Digest + "-draft", Action: "create", Namespace: session.Namespace, ProjectID: session.ProjectID, CaseSlug: session.Guided.CaseSlug, Title: prepared.Title, Body: prepared.Body, Status: "draft"}

	draft, err := core.MutateSpec(ctx, s.projectContext.ProjectDir, request)
	if err != nil {
		return err
	}

	if err := s.guidedSpecEffect("draft"); err != nil {
		return err
	}

	hash := sha256.Sum256([]byte(prepared.Body))

	session.Guided.SpecDraft = &core.GuidedSpecDraftReference{Reference: draft.Reference, Revision: draft.Revision, ContentHash: hex.EncodeToString(hash[:])}
	if err := core.RetainGuidedUseCaseProgress(ctx, s.projectContext.ProjectDir, session); err != nil {
		return err
	}

	return s.approveGuidedSpec(ctx, session, request, draft)
}

func (s service) approveGuidedSpec(ctx context.Context, session *core.AuthoringSession, request core.SpecMutationRequest, draft core.SpecMutationResult) error {
	operation := session.Guided.PendingOperation
	approval := core.GuidedSpecApproval{RunID: session.ID, Actor: session.Guided.Actor, Policy: core.GuidedSpecApprovalPolicy, GeneratedRevision: draft.Revision, ContentHash: session.Guided.SpecDraft.ContentHash, InputDigest: operation.PreparedSpec.InputDigest}
	request.OperationID, request.Action = "guided-"+operation.Digest+"-approve", "save"
	request.Namespace, request.Reference, request.ExpectedRevision = "", draft.Reference, draft.Revision
	request.Status, request.GuidedApproval = "approved", &approval

	result, err := core.MutateSpec(ctx, s.projectContext.ProjectDir, request)
	if err != nil {
		return err
	}

	if err := s.guidedSpecEffect("approval"); err != nil {
		return err
	}

	session.Guided.SpecApproval = &core.GuidedSpecApprovalReference{GuidedSpecApproval: approval, Reference: result.Reference, Revision: result.Revision}

	return core.SealGuidedSpec(ctx, s.projectContext.ProjectDir, session)
}

func (s service) guidedSpecEffect(phase string) error {
	if s.afterGuidedSpec != nil {
		return s.afterGuidedSpec(phase)
	}

	return nil
}

func (s service) prepareGuidedSpec(ctx context.Context, session core.AuthoringSession) (core.GuidedPreparedSpec, error) {
	bound, ctx, err := s.bindGeneration(ctx, "guided-spec-"+session.Guided.PendingOperation.Digest, string(core.QuizModelActionGenerateSpec))
	if err != nil {
		return core.GuidedPreparedSpec{}, err
	}

	request, digest, err := bound.guidedSpecModelRequest(session)
	if err != nil {
		return core.GuidedPreparedSpec{}, err
	}

	response, err := core.InvokeQuizModel(ctx, bound.quizModel, request)
	if err != nil {
		return core.GuidedPreparedSpec{}, err
	}

	if response.Status != core.QuizModelResponseReady || strings.TrimSpace(response.DraftSpec.Body) == "" || len(response.DraftSpec.Body) > 65536 || strings.TrimSpace(response.DraftSpec.Title) == "" || len(response.DraftSpec.Title) > 512 || strings.ContainsAny(response.DraftSpec.Title, "\r\n") {
		return core.GuidedPreparedSpec{}, errors.New("guided specification requires valid generated content within 64 KiB")
	}

	return core.GuidedPreparedSpec{Title: response.DraftSpec.Title, Body: response.DraftSpec.Body, InputDigest: digest, RequestID: request.ID, ResponseID: response.ID}, nil
}

func (s service) guidedSpecModelRequest(session core.AuthoringSession) (core.QuizModelRequest, string, error) {
	sections, state, entries, checkpoint, err := s.guidedSpecInputs(session)
	if err != nil {
		return core.QuizModelRequest{}, "", err
	}

	payload := s.quizModelPayloadWithEntries(state, core.QuizModelActionGenerateSpec, entries)
	payload.Project = core.QuizModelProjectPayload{ID: session.ProjectID, Name: session.Guided.CaseName}

	coreSections, err := core.QuizModelContextSectionsForPayload(payload)
	if err != nil {
		return core.QuizModelRequest{}, "", err
	}

	digest, err := core.GuidedCheckpointDigest(checkpoint)
	request := core.QuizModelRequest{
		ID: core.HarnessGenerationRequestID(s.ctx, "guided-"+session.ID+"-spec"), Action: core.QuizModelActionGenerateSpec, Strategy: "guided_committed_inputs",
		Scope: core.QuizScope{ProjectID: session.ProjectID, Namespace: session.Namespace, CaseSlug: session.Guided.CaseSlug}, Mode: core.QuizModeDeep,
		RuntimeState: state, Payload: payload, ContextSections: append(coreSections, sections...), OutputContract: core.QuizModelOutputContract{RequiredFields: payload.OutputContract.RequiredFields}, CreatedAt: s.utcNow(),
	}

	return request, digest, err
}

func (s service) guidedSpecInputs(session core.AuthoringSession) ([]core.QuizModelContextSection, core.QuizRuntimeState, []core.QuizEntry, core.GuidedStageCheckpoint, error) {
	var (
		state   core.QuizRuntimeState
		entries []core.QuizEntry
		quiz    core.GuidedStageCheckpoint
	)

	sections, err := s.guidedCommittedSections(session, "dump")
	if err != nil {
		return nil, state, entries, quiz, err
	}

	clarified, err := s.guidedCommittedSections(session, "quiz")
	if err != nil {
		return nil, state, entries, quiz, err
	}

	state, quiz, err = s.guidedCommittedQuiz(session)
	if err != nil {
		return nil, state, entries, quiz, err
	}

	for _, section := range clarified {
		if section.Name == state.RawPath {
			entries, err = core.ParseQuizRawCaptureEntries(section.Content)
		}
	}

	if err != nil {
		return nil, state, entries, quiz, err
	}

	if !state.ClarificationComplete || len(entries) == 0 {
		return nil, state, entries, quiz, fmt.Errorf("committed quiz is not sufficiently answered")
	}

	return append(sections, clarified...), state, entries, quiz, nil
}

func (s service) guidedCommittedQuiz(session core.AuthoringSession) (core.QuizRuntimeState, core.GuidedStageCheckpoint, error) {
	var state core.QuizRuntimeState

	for _, checkpoint := range session.Guided.Checkpoints {
		if checkpoint.Stage != "quiz" {
			continue
		}

		for _, artifact := range checkpoint.Artifacts {
			if !strings.Contains(artifact.Path, "/tmp/quiz-runtime/") {
				continue
			}

			data, err := core.ReadGuidedCommittedArtifact(s.ctx, s.projectContext.ProjectDir, checkpoint, artifact.Path)
			if err != nil {
				return state, checkpoint, err
			}

			err = json.Unmarshal(data, &state)

			return state, checkpoint, err
		}
	}

	return state, core.GuidedStageCheckpoint{}, errors.New("committed quiz runtime is unavailable")
}
