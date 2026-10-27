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
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type guidedSpecInput struct {
	Sections   []quizmodel.ContextSection
	State      quizcore.RuntimeState
	Entries    []quizcore.Entry
	Checkpoint core.GuidedStageCheckpoint
}

type guidedSpecRequest struct {
	Request quizmodel.Request
	Digest  string
}

type guidedCommittedQuizResult struct {
	State      quizcore.RuntimeState
	Checkpoint core.GuidedStageCheckpoint
}

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
	bound, ctx, err := s.bindGeneration(ctx, "guided-spec-"+session.Guided.PendingOperation.Digest, string(quizmodel.ActionGenerateSpec))
	if err != nil {
		return core.GuidedPreparedSpec{}, err
	}

	prepared, err := bound.guidedSpecModelRequest(session)
	if err != nil {
		return core.GuidedPreparedSpec{}, err
	}

	response, err := quizmodel.Invoke(ctx, bound.quizModel, prepared.Request)
	if err != nil {
		return core.GuidedPreparedSpec{}, err
	}

	if response.Status != quizmodel.ResponseReady || strings.TrimSpace(response.DraftSpec.Body) == "" || len(response.DraftSpec.Body) > 65536 || strings.TrimSpace(response.DraftSpec.Title) == "" || len(response.DraftSpec.Title) > 512 || strings.ContainsAny(response.DraftSpec.Title, "\r\n") {
		return core.GuidedPreparedSpec{}, errors.New("guided specification requires valid generated content within 64 KiB")
	}

	return core.GuidedPreparedSpec{Title: response.DraftSpec.Title, Body: response.DraftSpec.Body, InputDigest: prepared.Digest, RequestID: prepared.Request.ID, ResponseID: response.ID}, nil
}

func (s service) guidedSpecModelRequest(session core.AuthoringSession) (guidedSpecRequest, error) {
	inputs, err := s.guidedSpecInputs(session)
	if err != nil {
		return guidedSpecRequest{}, err
	}

	payload := s.quizModelPayloadWithEntries(inputs.State, quizmodel.ActionGenerateSpec, inputs.Entries)
	payload.Project = quizmodel.ProjectPayload{ID: session.ProjectID, Name: session.Guided.CaseName}

	coreSections, err := quizmodel.ContextSectionsForPayload(payload)
	if err != nil {
		return guidedSpecRequest{}, err
	}

	digest, err := core.GuidedCheckpointDigest(inputs.Checkpoint)
	request := quizmodel.Request{
		ID: core.HarnessGenerationRequestID(s.ctx, "guided-"+session.ID+"-spec"), Action: quizmodel.ActionGenerateSpec, Strategy: "guided_committed_inputs",
		Scope: quizcore.Scope{ProjectID: session.ProjectID, Namespace: session.Namespace, CaseSlug: session.Guided.CaseSlug}, Mode: quizcore.ModeDeep,
		RuntimeState: inputs.State, Payload: payload, ContextSections: append(coreSections, inputs.Sections...), OutputContract: quizmodel.OutputContract{RequiredFields: payload.OutputContract.RequiredFields}, CreatedAt: s.utcNow(),
	}

	return guidedSpecRequest{Request: request, Digest: digest}, err
}

func (s service) guidedSpecInputs(session core.AuthoringSession) (guidedSpecInput, error) {
	var (
		state   quizcore.RuntimeState
		entries []quizcore.Entry
		quiz    core.GuidedStageCheckpoint
	)

	sections, err := s.guidedCommittedSections(session, "dump")
	if err != nil {
		return guidedSpecInput{State: state, Entries: entries, Checkpoint: quiz}, err
	}

	clarified, err := s.guidedCommittedSections(session, "quiz")
	if err != nil {
		return guidedSpecInput{State: state, Entries: entries, Checkpoint: quiz}, err
	}

	committed, err := s.guidedCommittedQuiz(session)
	if err != nil {
		return guidedSpecInput{State: committed.State, Entries: entries, Checkpoint: committed.Checkpoint}, err
	}

	state, quiz = committed.State, committed.Checkpoint

	for _, section := range clarified {
		if section.Name == state.RawPath {
			entries, err = quizcore.ParseRawCaptureEntries(section.Content)
		}
	}

	if err != nil {
		return guidedSpecInput{State: state, Entries: entries, Checkpoint: quiz}, err
	}

	if !state.ClarificationComplete || len(entries) == 0 {
		return guidedSpecInput{State: state, Entries: entries, Checkpoint: quiz}, fmt.Errorf("committed quiz is not sufficiently answered")
	}

	return guidedSpecInput{Sections: append(sections, clarified...), State: state, Entries: entries, Checkpoint: quiz}, nil
}

func (s service) guidedCommittedQuiz(session core.AuthoringSession) (guidedCommittedQuizResult, error) {
	var state quizcore.RuntimeState

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
				return guidedCommittedQuizResult{State: state, Checkpoint: checkpoint}, err
			}

			err = json.Unmarshal(data, &state)

			return guidedCommittedQuizResult{State: state, Checkpoint: checkpoint}, err
		}
	}

	return guidedCommittedQuizResult{State: state}, errors.New("committed quiz runtime is unavailable")
}
