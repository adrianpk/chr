package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ContinueAuthoring binds arguments before provider preparation and journals the
// canonical session, transcript, artifacts and result receipt under one owner.
func (r *fileAuthoringSessionRepository) ContinueAuthoring(ctx context.Context, request AuthoringContinuationRequest, generate AuthoringGenerator, guards ...AuthoringCommitGuard) (result AuthoringContinuationResult, err error) {
	if len(guards) > 1 || (len(guards) == 1 && guards[0] == nil) {
		return result, errors.New("authoring accepts at most one current commit guard")
	}

	request, digest, err := normalizeAuthoringRequest(request)
	if err != nil {
		return result, err
	}

	var guard AuthoringCommitGuard
	if len(guards) == 1 {
		guard = guards[0]
	}

	err = WithAuthoringMutation(ctx, r.projectDir, func(ctx context.Context) error {
		var applyErr error

		result, applyErr = r.continueAuthoring(ctx, request, digest, generate, guard)

		return applyErr
	})

	return result, err
}

func (r *fileAuthoringSessionRepository) continueAuthoring(ctx context.Context, request AuthoringContinuationRequest, digest string, generate AuthoringGenerator, guard AuthoringCommitGuard) (AuthoringContinuationResult, error) {
	binding, err := r.reserveAuthoringRequest(request, digest)
	if err != nil {
		return AuthoringContinuationResult{}, err
	}

	retained, found, err := r.readAuthoringReceipt(request, digest)
	if err != nil || found {
		return retained, err
	}

	if err := validateAuthoringRequest(request); err != nil {
		return retained, err
	}

	snapshot, err := r.guardAuthoringSnapshot(request)
	if err != nil {
		return retained, err
	}

	if err := prepareAuthoringSnapshot(&snapshot, binding); err != nil {
		return retained, err
	}

	draft, err := r.generateAuthoring(ctx, &binding, snapshot, generate)
	if err != nil {
		return retained, err
	}

	if guard != nil {
		if err := guard(ctx); err != nil {
			return retained, err
		}
	}

	return r.commitAuthoringContinuation(ctx, binding, snapshot, draft)
}

func newAuthoringSnapshot(request AuthoringContinuationRequest) AuthoringSnapshot {
	now := time.Now().UTC()

	return AuthoringSnapshot{Session: AuthoringSession{ID: request.SessionID, ProjectID: request.ProjectID, Namespace: request.Namespace, Status: AuthoringSessionStatusActive, Mode: AuthoringSessionModeAssisted, Target: AuthoringTarget(request.Target), Topic: request.Topic, CreatedAt: now, StartedAt: &now, UpdatedAt: now}, Events: []AssistedTranscriptEvent{}}
}

func guardAuthoringContinuation(snapshot AuthoringSnapshot, request AuthoringContinuationRequest) error {
	if request.Action == "answer" || request.Action == "respond" {
		return guardAuthoringAnswer(snapshot, request)
	}

	if authoringQuestionCount(snapshot) >= 8 && request.Action == "ask" {
		return errors.New("authoring question capacity exhausted")
	}

	if snapshot.Question != nil && request.Action != "ask" {
		return fmt.Errorf("%w: answer or replace the pending question", ErrWorkflowConflict)
	}

	if len(snapshot.Events) >= 64 || len(snapshot.Session.Provenance) >= 64 {
		return errors.New("authoring transcript capacity exhausted")
	}

	return nil
}

func (r *fileAuthoringSessionRepository) generateAuthoring(ctx context.Context, binding *authoringRequestBinding, snapshot AuthoringSnapshot, generate AuthoringGenerator) (AuthoringDraft, error) {
	if binding.Request.Action == "answer" {
		return AuthoringDraft{}, nil
	}

	if generate == nil {
		return AuthoringDraft{}, errors.New("configured authoring provider is required")
	}

	if binding.Attempts >= 5 {
		return AuthoringDraft{}, ErrAuthoringProviderExhausted
	}

	if time.Now().Before(binding.RetryAt) {
		return AuthoringDraft{}, ErrAuthoringProviderWait
	}

	binding.Attempts++
	binding.RetryAt = time.Now().UTC().Add(61 * time.Second)

	binding.ErrorCode = "interrupted"
	if err := r.saveAuthoringRequest(*binding); err != nil {
		return AuthoringDraft{}, err
	}

	callCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	draft, callErr := generate(callCtx, snapshot, binding.Request)
	if callErr == nil {
		callErr = validateGeneratedAuthoring(callCtx, draft, snapshot, binding.Request)
	}

	if callErr != nil {
		binding.ErrorCode = authoringProviderFailure(callErr)
		binding.RetryAt = time.Now().UTC().Add(time.Second * time.Duration(1<<(binding.Attempts-1)))

		return draft, errors.Join(ErrAuthoringProviderFailure, callErr, r.saveAuthoringRequest(*binding))
	}

	return draft, nil
}

func validateAuthoringDraft(draft AuthoringDraft, request AuthoringContinuationRequest) error {
	if len(draft.Body) > 49152 || len(draft.Question) > 4096 || len(draft.Rationale) > 2048 {
		return errors.New("authoring provider output exceeds its bound")
	}

	if draft.Question != "" {
		if draft.Body != "" {
			return errors.New("authoring cannot generate an artifact and ask a question together")
		}

		return nil
	}

	if draft.Rationale != "" {
		return errors.New("authoring rationale requires a question")
	}

	if request.Action == "ask" || request.Target == AuthoringArtifactKindQuiz {
		return errors.New("authoring provider returned no question")
	}

	if draft.Body == "" {
		return errors.New("authoring provider returned no content")
	}

	return nil
}

func (r *fileAuthoringSessionRepository) commitAuthoringContinuation(ctx context.Context, binding authoringRequestBinding, snapshot AuthoringSnapshot, draft AuthoringDraft) (AuthoringContinuationResult, error) {
	changes, err := r.prepareAuthoringContinuation(binding, snapshot, draft)
	if err != nil {
		return AuthoringContinuationResult{}, err
	}

	projected, err := r.readAuthoringSnapshot(binding.Request.SessionID, authoringImageMap(changes))
	if err != nil {
		return AuthoringContinuationResult{}, err
	}

	result := AuthoringContinuationResult{RequestID: binding.Request.Binding.RequestID, ArgumentDigest: binding.Digest, ReceiptRef: authoringReceiptRef(binding.Request.Binding.RequestID), Session: projected.Session, SessionRef: projected.SessionRef, SourceHash: projected.SourceHash, Question: projected.Question, Prompt: draft.Question}

	receipt, err := authoringReceiptChange(binding, result)
	if err != nil {
		return result, err
	}

	changes = append(changes, receipt)

	if err := ctx.Err(); err != nil {
		return result, err
	}

	if err := r.writeAuthoringJournal(changes); err != nil {
		return result, err
	}

	return result, r.recoverAuthoringMutation()
}

func authoringImageMap(changes []workflowFileChange) map[string][]byte {
	result := map[string][]byte{}
	for _, change := range changes {
		result[change.Path] = change.After
	}

	return result
}

func authoringReceiptChange(binding authoringRequestBinding, result AuthoringContinuationResult) (workflowFileChange, error) {
	receipt := authoringMutationReceipt{Schema: "chr-authoring-receipt/v1", Digest: binding.Digest, Result: result}

	encoded, err := json.Marshal(receipt)
	if err != nil {
		return workflowFileChange{}, err
	}

	receipt.Checksum = workflowHash(encoded)
	encoded, err = json.Marshal(receipt)

	return workflowFileChange{Path: authoringReceiptRef(binding.Request.Binding.RequestID), After: encoded}, err
}

func authoringProviderFailure(err error) string {
	switch {
	case errors.Is(err, ErrStructuredModelCapacity):
		return "provider_capacity"
	case errors.Is(err, context.DeadlineExceeded):
		return "provider_timeout"
	case errors.Is(err, context.Canceled):
		return "provider_canceled"
	default:
		return "provider_unavailable"
	}
}

func authoringQuestionCount(snapshot AuthoringSnapshot) int {
	count := 0

	for _, event := range snapshot.Events {
		if event.Type == AssistedTranscriptEventQuestion {
			count++
		}
	}

	return count
}

func prepareAuthoringSnapshot(snapshot *AuthoringSnapshot, binding authoringRequestBinding) error {
	if binding.Request.Action == "create" {
		snapshot.Session.CreatedAt = binding.CreatedAt
		snapshot.Session.StartedAt = &binding.CreatedAt
	}

	return guardAuthoringContinuation(*snapshot, binding.Request)
}

func guardAuthoringAnswer(snapshot AuthoringSnapshot, request AuthoringContinuationRequest) error {
	if request.Target != AuthoringArtifactKind(snapshot.Session.Target) || snapshot.Question == nil || *snapshot.Question != *request.Question || snapshot.Question.Revision != snapshot.Session.Revision {
		return ErrWorkflowConflict
	}

	return nil
}

func validateGeneratedAuthoring(ctx context.Context, draft AuthoringDraft, snapshot AuthoringSnapshot, request AuthoringContinuationRequest) error {
	if err := validateAuthoringDraft(draft, request); err != nil {
		return err
	}

	if draft.Question != "" && authoringQuestionCount(snapshot) >= 8 {
		return errors.New("authoring question capacity exhausted")
	}

	return ctx.Err()
}
