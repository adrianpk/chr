package authoring

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	modelcore "forge.adrianpk.com/chr/internal/core/model"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ContinueAuthoring binds arguments before provider preparation and journals the
// canonical session, transcript, artifacts and result receipt under one owner.
func (r *fileSessionRepository) ContinueAuthoring(ctx context.Context, request ContinuationRequest, generate Generator, guards ...CommitGuard) (result ContinuationResult, err error) {
	if len(guards) > 1 || (len(guards) == 1 && guards[0] == nil) {
		return result, errors.New("authoring accepts at most one current commit guard")
	}

	normalized, err := normalizeAuthoringRequest(request)
	if err != nil {
		return result, err
	}

	request, digest := normalized.Request, normalized.Digest

	var guard CommitGuard
	if len(guards) == 1 {
		guard = guards[0]
	}

	err = WithMutation(ctx, r.projectDir, func(ctx context.Context) error {
		var applyErr error

		result, applyErr = r.continueAuthoring(ctx, request, digest, generate, guard)

		return applyErr
	})

	return result, err
}

func (r *fileSessionRepository) continueAuthoring(ctx context.Context, request ContinuationRequest, digest string, generate Generator, guard CommitGuard) (ContinuationResult, error) {
	binding, err := r.reserveAuthoringRequest(request, digest)
	if err != nil {
		return ContinuationResult{}, err
	}

	lookup, err := r.readAuthoringReceipt(request, digest)
	retained := lookup.Result

	if err != nil || lookup.Found {
		return retained, err
	}

	if err := validateAuthoringRequest(request); err != nil {
		return retained, err
	}

	snapshot, err := r.guardAuthoringSnapshot(request)
	if err != nil {
		return retained, err
	}

	if err := r.prepareAuthoringSnapshot(ctx, &snapshot, binding); err != nil {
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

func newAuthoringSnapshot(request ContinuationRequest) Snapshot {
	now := time.Now().UTC()

	return Snapshot{Session: Session{ID: request.SessionID, ProjectID: request.ProjectID, Namespace: request.Namespace, Status: SessionStatusActive, Mode: SessionModeAssisted, Target: Target(request.Target), Topic: request.Topic, CreatedAt: now, StartedAt: &now, UpdatedAt: now}, Events: []AssistedTranscriptEvent{}}
}

func guardAuthoringContinuation(snapshot Snapshot, request ContinuationRequest) error {
	if request.Action == "answer" || request.Action == "respond" {
		return guardAuthoringAnswer(snapshot, request)
	}

	if authoringQuestionCount(snapshot) >= 8 && request.Action == "ask" {
		return errors.New("authoring question capacity exhausted")
	}

	if snapshot.Question != nil && request.Action != "ask" {
		return fmt.Errorf("%w: answer or replace the pending question", workflowstate.ErrConflict)
	}

	if len(snapshot.Events) >= 64 || len(snapshot.Session.Provenance) >= 64 {
		return errors.New("authoring transcript capacity exhausted")
	}

	return nil
}

func (r *fileSessionRepository) generateAuthoring(ctx context.Context, binding *authoringRequestBinding, snapshot Snapshot, generate Generator) (Draft, error) {
	if binding.Request.Action == "answer" {
		return Draft{}, nil
	}

	if generate == nil {
		return Draft{}, errors.New("configured authoring provider is required")
	}

	if binding.Attempts >= 5 {
		return Draft{}, ErrProviderExhausted
	}

	if time.Now().Before(binding.RetryAt) {
		return Draft{}, ErrProviderWait
	}

	binding.Attempts++
	binding.RetryAt = time.Now().UTC().Add(61 * time.Second)

	binding.ErrorCode = "interrupted"
	if err := r.saveAuthoringRequest(*binding); err != nil {
		return Draft{}, err
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

		return draft, errors.Join(ErrProviderFailure, callErr, r.saveAuthoringRequest(*binding))
	}

	return draft, nil
}

func validateAuthoringDraft(draft Draft, request ContinuationRequest) error {
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

	if request.Action == "ask" || request.Target == ArtifactKindQuiz {
		return errors.New("authoring provider returned no question")
	}

	if draft.Body == "" {
		return errors.New("authoring provider returned no content")
	}

	return nil
}

func (r *fileSessionRepository) commitAuthoringContinuation(ctx context.Context, binding authoringRequestBinding, snapshot Snapshot, draft Draft) (ContinuationResult, error) {
	changes, err := r.prepareAuthoringContinuation(ctx, binding, snapshot, draft)
	if err != nil {
		return ContinuationResult{}, err
	}

	projected, err := r.readAuthoringSnapshot(binding.Request.SessionID, authoringImageMap(changes))
	if err != nil {
		return ContinuationResult{}, err
	}

	result := ContinuationResult{RequestID: binding.Request.Binding.RequestID, ArgumentDigest: binding.Digest, ReceiptRef: authoringReceiptRef(binding.Request.Binding.RequestID), Session: projected.Session, SessionRef: projected.SessionRef, SourceHash: projected.SourceHash, Question: projected.Question, Prompt: draft.Question}

	receipt, err := authoringReceiptChange(binding, result)
	if err != nil {
		return result, err
	}

	changes = append(changes, receipt)

	if err := ctx.Err(); err != nil {
		return result, err
	}

	if err := r.writeAuthoringContinuationJournal(binding.Request.Target, changes); err != nil {
		return result, err
	}

	return result, r.recoverAuthoringMutation()
}

func authoringImageMap(changes []workflowstate.FileChange) map[string][]byte {
	result := map[string][]byte{}
	for _, change := range changes {
		result[change.Path] = change.After
	}

	return result
}

func authoringReceiptChange(binding authoringRequestBinding, result ContinuationResult) (workflowstate.FileChange, error) {
	receipt := authoringMutationReceipt{Schema: "chr-authoring-receipt/v1", Digest: binding.Digest, Result: result}

	encoded, err := json.Marshal(receipt)
	if err != nil {
		return workflowstate.FileChange{}, err
	}

	receipt.Checksum = contenthash.Sum(encoded)
	encoded, err = json.Marshal(receipt)

	return workflowstate.FileChange{Path: authoringReceiptRef(binding.Request.Binding.RequestID), After: encoded}, err
}

func authoringProviderFailure(err error) string {
	switch {
	case errors.Is(err, modelcore.ErrCapacity):
		return "provider_capacity"
	case errors.Is(err, context.DeadlineExceeded):
		return "provider_timeout"
	case errors.Is(err, context.Canceled):
		return "provider_canceled"
	default:
		return "provider_unavailable"
	}
}

func authoringQuestionCount(snapshot Snapshot) int {
	count := 0

	for _, event := range snapshot.Events {
		if event.Type == AssistedTranscriptEventQuestion {
			count++
		}
	}

	return count
}

func (r *fileSessionRepository) prepareAuthoringSnapshot(ctx context.Context, snapshot *Snapshot, binding authoringRequestBinding) error {
	if binding.Request.Action == "create" {
		snapshot.Session.CreatedAt = binding.CreatedAt
		snapshot.Session.StartedAt = &binding.CreatedAt
	}

	if err := guardAuthoringContinuation(*snapshot, binding.Request); err != nil {
		return err
	}

	return r.prepareAuthoringDumpSnapshot(ctx, snapshot, binding.Request)
}

func guardAuthoringAnswer(snapshot Snapshot, request ContinuationRequest) error {
	if request.Target != ArtifactKind(snapshot.Session.Target) || snapshot.Question == nil || *snapshot.Question != *request.Question || snapshot.Question.Revision != snapshot.Session.Revision {
		return workflowstate.ErrConflict
	}

	return nil
}

func validateGeneratedAuthoring(ctx context.Context, draft Draft, snapshot Snapshot, request ContinuationRequest) error {
	if err := validateAuthoringDraft(draft, request); err != nil {
		return err
	}

	if draft.Question != "" && authoringQuestionCount(snapshot) >= 8 {
		return errors.New("authoring question capacity exhausted")
	}

	return ctx.Err()
}

func (r *fileSessionRepository) writeAuthoringContinuationJournal(target ArtifactKind, changes []workflowstate.FileChange) error {
	if target == ArtifactKindDump {
		return r.writeAuthoringJournalVersion(projectDumpJournalSchema, changes)
	}

	return r.writeAuthoringJournal(changes)
}
