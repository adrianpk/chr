package core

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"os"
	"slices"
	"time"
)

// ErrWorkflowOwned identifies overlap with retained task or hierarchy ownership.
var ErrWorkflowOwned = errors.New("workflow work is already owned")

// ErrWorkflowCapacity reports a live serial reservation, independently of a lease.
var ErrWorkflowCapacity = errors.New("workflow waiting for repository capacity")

// WorkflowClaim is an owner credential. It cannot be reconstructed from status.
type WorkflowClaim struct {
	InvocationID string
	AttemptID    string
	Generation   uint64
	Token        string `json:"-"`
	store        *WorkflowStore
	reservation  *os.File
	locks        []*AgentOperationLock
	execution    *workflowClaimExecution
	purpose      string
}

// ClaimWorkflow acquires one task and its existing serial operation exclusion.
func (s *WorkflowStore) ClaimWorkflow(ctx context.Context, id string) (claim *WorkflowClaim, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		claim, readErr = owner.claimWorkflow(ctx, run)

		return readErr
	})

	return claim, err
}

func (s *WorkflowStore) claimWorkflow(ctx context.Context, run workflowInvocationFile) (claim *WorkflowClaim, err error) {
	if run.StopIntent != "" {
		return nil, ErrWorkflowOwned
	}

	if err := s.reuseWorkflowValidatedTasks(ctx, &run); err != nil {
		return nil, err
	}

	if !workflowInvocationClaimable(run) {
		return nil, fmt.Errorf("%w: invocation %s is %s", ErrWorkflowOwned, run.ID, run.Status)
	}

	record, err := s.observeWorkflowClaim(ctx, run)
	if err != nil {
		return nil, err
	}

	claim, err = s.reserveWorkflowClaim(ctx, run)
	if err != nil {
		return nil, err
	}

	owned := claim

	defer func() {
		if err != nil {
			err = errors.Join(err, owned.releaseReservation())
		}
	}()

	claim.Token, err = newWorkflowClaimToken()
	if err != nil {
		return nil, err
	}

	if err := s.beginWorkflowAttempt(&run, &record, claim); err != nil {
		return nil, err
	}

	if err := s.commitWorkflowClaim(run, record); err != nil {
		return nil, err
	}

	return claim, nil
}

func (s *WorkflowStore) observeWorkflowClaim(ctx context.Context, run workflowInvocationFile) (WorkflowRecord, error) {
	if err := s.checkWorkflowInputs(ctx, run); err != nil {
		return WorkflowRecord{}, err
	}

	if err := s.checkWorkflowTarget(ctx, run.Selection.Target); err != nil {
		return WorkflowRecord{}, err
	}

	selection, err := s.readWorkflowSelection(ctx)
	if err != nil {
		return WorkflowRecord{}, err
	}

	member := run.Members[run.Cursor]
	request := run.Selection
	request.Selector = WorkflowSelector{Extent: "ticket", Identity: member.Admission.Key}

	request, err = s.workflowLaneSelection(ctx, run, selection.records[member.Admission.Key], request)
	if err != nil {
		return WorkflowRecord{}, err
	}

	preview, err := selection.preview(request)
	if err != nil {
		return WorkflowRecord{}, err
	}

	if err := s.observePrerequisites(ctx, selection, request, &preview); err != nil {
		return WorkflowRecord{}, err
	}

	if len(preview.Members) != 1 || preview.Members[0].Readiness != "ready" {
		return WorkflowRecord{}, fmt.Errorf("workflow task is not ready: %s", member.Admission.Key)
	}

	record := selection.records[member.Admission.Key]
	if err := validateWorkflowClaimableContract(record); err != nil {
		return WorkflowRecord{}, err
	}

	return record, nil
}

func validateWorkflowClaimableContract(record WorkflowRecord) error {
	if !workflowTaskClaimable(record.Task) {
		return ErrWorkflowOwned
	}

	return validateWorkflowClaimContract(record)
}

func newWorkflowClaimToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}

	return hex.EncodeToString(token), nil
}

func workflowTaskClaimable(task *WorkflowTask) bool {
	return task.AttemptID == "" && task.ClaimGeneration != math.MaxUint64
}

func (s *WorkflowStore) beginWorkflowAttempt(run *workflowInvocationFile, record *WorkflowRecord, claim *WorkflowClaim) error {
	count := 0

	for _, attempt := range run.Attempts {
		if attempt.TaskKey == record.Task.Key && attempt.Purpose == claim.purpose {
			count++
		}
	}

	limit := run.Policy.MaxWorkerAttempts
	if claim.purpose == workflowSliceReviewPurpose || claim.purpose == workflowIntegrationValidationPurpose {
		limit = WorkflowPageLimit
	}

	if count >= limit || len(run.Attempts) >= WorkflowRecordLimit {
		return fmt.Errorf("workflow worker attempt budget exhausted")
	}

	now := time.Now().UTC()
	claim.AttemptID = fmt.Sprintf("%s-attempt-%d", run.ID, len(run.Attempts)+1)
	claim.Generation = record.Task.ClaimGeneration + 1

	owner, err := workflowProcessIdentity(os.Getpid())
	if err != nil {
		return err
	}

	attempt := WorkflowAttempt{Purpose: claim.purpose, OwnerProcess: &owner, Replaces: workflowReplacedAttempt(*run, record.Task.Key, claim.purpose), ID: claim.AttemptID, TaskKey: record.Task.Key, Generation: claim.Generation, InputRevision: record.Revision, ContractHash: record.Task.ContractHash, LeaseUntil: now.Add(time.Duration(run.Policy.WorkerTimeoutSeconds) * time.Second), StartedAt: now, Outcome: "running"}
	run.Attempts = append(run.Attempts, attempt)
	run.Owners[claim.AttemptID] = workflowHash([]byte(claim.Token))
	run.Status = "running"

	record.Task.AttemptID, record.Task.ClaimGeneration = claim.AttemptID, claim.Generation
	if claim.purpose == "" {
		record.Task.Phase, record.Task.Status = "running", TicketStatusInProgress
	}

	return nil
}

func (s *WorkflowStore) commitWorkflowClaim(run workflowInvocationFile, record WorkflowRecord, extra ...workflowFileChange) error {
	files, err := s.workflowClaimRecordChanges(&run, &record)
	if err != nil {
		return err
	}

	run.Revision++
	run.UpdatedAt = record.UpdatedAt

	state, err := workflowInvocationChange(run)
	if err != nil {
		return err
	}

	files = append(files, extra...)

	return s.commitFiles(append(files, state))
}

func (s *WorkflowStore) workflowClaimRecordChanges(run *workflowInvocationFile, record *WorkflowRecord) ([]workflowFileChange, error) {
	record.Revision++
	record.UpdatedAt = time.Now().UTC()

	files, err := s.encodeChanges(*record)
	if err != nil {
		return nil, err
	}

	if err := guardWorkflowClaimImages(*record, files); err != nil {
		return nil, err
	}

	record.SourceHash = workflowClaimRecordHash(*record, files)

	for i := range run.Members {
		if run.Members[i].Admission.Key == record.Task.Key {
			run.Members[i].ExpectedRevision = record.Revision
			run.Members[i].ExpectedSourceHash = record.SourceHash
		}
	}

	if err := record.Validate(); err != nil {
		return nil, err
	}

	return files, nil
}

func workflowClaimRecordHash(record WorkflowRecord, files []workflowFileChange) string {
	hash := workflowHash(files[0].After)
	if len(files) > 1 {
		hash += ":" + workflowHash(record.Body)
	}

	return hash
}

// CheckWorkflowClaim rejects expired credentials and any unaccepted file image.
func (s *WorkflowStore) CheckWorkflowClaim(ctx context.Context, claim *WorkflowClaim) error {
	if claim == nil || claim.store == nil || claim.execution == nil {
		return ErrWorkflowOwned
	}

	return s.withLock(ctx, func() error {
		_, _, err := s.checkWorkflowClaim(ctx, claim)

		return err
	})
}

func (s *WorkflowStore) checkWorkflowClaim(ctx context.Context, claim *WorkflowClaim) (workflowInvocationFile, WorkflowRecord, error) {
	return s.checkWorkflowCredential(ctx, claim, true)
}

func (s *WorkflowStore) checkWorkflowCredential(ctx context.Context, claim *WorkflowClaim, lease bool) (workflowInvocationFile, WorkflowRecord, error) {
	run, owner, err := s.locateWorkflowInvocation(ctx, claim.InvocationID)
	if err != nil {
		return run, WorkflowRecord{}, err
	}

	if owner.layout.Root() != claim.store.layout.Root() || len(run.Attempts) == 0 {
		return run, WorkflowRecord{}, ErrWorkflowOwned
	}

	if err := owner.checkWorkflowInputs(ctx, run); err != nil {
		return run, WorkflowRecord{}, err
	}

	attempt := run.Attempts[len(run.Attempts)-1]
	if !workflowCredentialValid(run, claim, attempt, lease) {
		return run, WorkflowRecord{}, ErrWorkflowOwned
	}

	record, err := owner.readWorkflowAttemptTask(run, attempt)
	if err != nil {
		return run, record, err
	}

	if record.Task.AttemptID != claim.AttemptID || record.Task.ClaimGeneration != claim.Generation {
		return run, record, ErrWorkflowConflict
	}

	return run, record, nil
}

// CheckpointWorkflow yields at a quiescent task boundary without claiming success.
func (s *WorkflowStore) CheckpointWorkflow(ctx context.Context, claim *WorkflowClaim) error {
	if claim == nil || claim.store == nil || claim.execution == nil {
		return ErrWorkflowOwned
	}

	claim.execution.Lock()
	defer claim.execution.Unlock()

	if !workflowClaimQuiescent(claim) {
		return ErrWorkflowOwned
	}

	return s.withLock(ctx, func() error {
		run, record, err := s.checkWorkflowClaim(ctx, claim)
		if err != nil {
			return err
		}

		if err := claim.store.checkpointWorkflowAttempt(run, record, "checkpointed", nil); err != nil {
			return err
		}

		return claim.releaseReservation()
	})
}

func workflowAttemptCredentialMatches(run workflowInvocationFile, claim *WorkflowClaim, attempt WorkflowAttempt) bool {
	return run.Status == "running" && attempt.Outcome == "running" && attempt.Purpose == claim.purpose && attempt.ID == claim.AttemptID && attempt.Generation == claim.Generation && run.Owners[attempt.ID] == workflowHash([]byte(claim.Token))
}

func workflowReplacedAttempt(run workflowInvocationFile, key, purpose string) string {
	for _, attempt := range slices.Backward(run.Attempts) {
		if attempt.TaskKey == key && attempt.Purpose == purpose && attempt.Outcome != "validated" {
			return attempt.ID
		}
	}

	return ""
}

func (s *WorkflowStore) readWorkflowAttemptTask(run workflowInvocationFile, attempt WorkflowAttempt) (WorkflowRecord, error) {
	cursor := run.Cursor
	if attempt.Purpose == workflowSliceValidationPurpose || attempt.Purpose == workflowSliceReviewPurpose || attempt.Purpose == workflowIntegrationValidationPurpose {
		cursor--
	}

	if cursor < 0 || cursor >= len(run.Members) || run.Members[cursor].Admission.Key != attempt.TaskKey {
		return WorkflowRecord{}, ErrWorkflowConflict
	}

	return s.readRecord(run.Members[cursor].Admission.Path)
}

func workflowInvocationClaimable(run workflowInvocationFile) bool {
	return run.StopIntent == "" && slices.Contains([]string{"pending", "paused", "waiting_capacity"}, run.Status) && run.Cursor < len(run.Members)
}

func workflowCredentialValid(run workflowInvocationFile, claim *WorkflowClaim, attempt WorkflowAttempt, lease bool) bool {
	return workflowAttemptCredentialMatches(run, claim, attempt) && (!lease || time.Now().Before(attempt.LeaseUntil))
}
