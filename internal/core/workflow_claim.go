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
	if err := s.reuseWorkflowValidatedTasks(ctx, &run); err != nil {
		return nil, err
	}

	if !slices.Contains([]string{"pending", "paused", "waiting_capacity"}, run.Status) || run.Cursor == len(run.Members) {
		return nil, fmt.Errorf("%w: invocation %s is %s", ErrWorkflowOwned, run.ID, run.Status)
	}

	record, err := s.observeWorkflowClaim(ctx, run)
	if err != nil {
		return nil, err
	}

	claim, err = s.reserveWorkflowClaim(ctx, run.ID)
	if err != nil {
		return nil, err
	}

	owned := claim

	defer func() {
		if err != nil {
			err = errors.Join(err, owned.releaseReservation())
		}
	}()

	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return nil, err
	}

	claim.Token = hex.EncodeToString(token)
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

func workflowTaskClaimable(task *WorkflowTask) bool {
	return task.AttemptID == "" && task.ClaimGeneration != math.MaxUint64
}

func (s *WorkflowStore) beginWorkflowAttempt(run *workflowInvocationFile, record *WorkflowRecord, claim *WorkflowClaim) error {
	count := 0

	for _, attempt := range run.Attempts {
		if attempt.TaskKey == record.Task.Key {
			count++
		}
	}

	if count >= run.Policy.MaxWorkerAttempts || len(run.Attempts) >= WorkflowRecordLimit {
		return fmt.Errorf("workflow worker attempt budget exhausted")
	}

	now := time.Now().UTC()
	claim.AttemptID = fmt.Sprintf("%s-attempt-%d", run.ID, len(run.Attempts)+1)
	claim.Generation = record.Task.ClaimGeneration + 1
	attempt := WorkflowAttempt{ID: claim.AttemptID, TaskKey: record.Task.Key, Generation: claim.Generation, InputRevision: record.Revision, ContractHash: record.Task.ContractHash, LeaseUntil: now.Add(time.Duration(run.Policy.WorkerTimeoutSeconds) * time.Second), StartedAt: now, Outcome: "running"}
	run.Attempts = append(run.Attempts, attempt)
	run.Owners[claim.AttemptID] = workflowHash([]byte(claim.Token))
	run.Status = "running"
	record.Task.AttemptID, record.Task.ClaimGeneration = claim.AttemptID, claim.Generation
	record.Task.Phase, record.Task.Status = "running", TicketStatusInProgress

	return nil
}

func (s *WorkflowStore) commitWorkflowClaim(run workflowInvocationFile, record WorkflowRecord, extra ...workflowFileChange) error {
	record.Revision++
	record.UpdatedAt = time.Now().UTC()

	files, err := s.encodeChanges(record)
	if err != nil {
		return err
	}

	if err := guardWorkflowClaimImages(record, files); err != nil {
		return err
	}

	for i := range run.Members {
		if run.Members[i].Admission.Key == record.Task.Key {
			run.Members[i].ExpectedRevision = record.Revision
			run.Members[i].ExpectedSourceHash = workflowClaimRecordHash(record, files)
		}
	}

	run.Revision++
	run.UpdatedAt = record.UpdatedAt

	state, err := workflowInvocationChange(run)
	if err != nil {
		return err
	}

	if err := record.Validate(); err != nil {
		return err
	}

	files = append(files, extra...)

	return s.commitFiles(append(files, state))
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
	if !workflowAttemptOwnerMatches(run, claim, attempt) {
		return run, WorkflowRecord{}, ErrWorkflowOwned
	}

	if run.Cursor >= len(run.Members) {
		return run, WorkflowRecord{}, ErrWorkflowConflict
	}

	member := run.Members[run.Cursor]

	record, err := owner.readRecord(member.Admission.Path)
	if err != nil {
		return run, record, err
	}

	if record.Task.AttemptID != claim.AttemptID || record.Task.ClaimGeneration != claim.Generation {
		return run, record, ErrWorkflowConflict
	}

	return run, record, nil
}

func workflowAttemptOwnerMatches(run workflowInvocationFile, claim *WorkflowClaim, attempt WorkflowAttempt) bool {
	return run.Status == "running" && attempt.ID == claim.AttemptID && attempt.Generation == claim.Generation && run.Owners[attempt.ID] == workflowHash([]byte(claim.Token)) && time.Now().Before(attempt.LeaseUntil)
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

		record.Task.AttemptID, record.Task.Phase, record.Task.Status = "", "pending", TicketStatusReadyForDev
		attempt := &run.Attempts[len(run.Attempts)-1]
		attempt.Outcome, attempt.FinishedAt = "checkpointed", time.Now().UTC()

		run.Status = "paused"
		if err := claim.store.commitWorkflowClaim(run, record); err != nil {
			return err
		}

		return claim.releaseReservation()
	})
}
