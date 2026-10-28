package workflow

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

	"forge.adrianpk.com/chr/internal/core/contenthash"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ErrWorkflowOwned identifies overlap with retained task or hierarchy ownership.
var ErrWorkflowOwned = errors.New("workflow work is already owned")

// ErrWorkflowCapacity reports occupied repository capacity, independently of a lease.
var ErrWorkflowCapacity = errors.New("workflow waiting for repository capacity")

// Claim is an owner credential. It cannot be reconstructed from status.
type Claim struct {
	InvocationID  string
	AttemptID     string
	Generation    uint64
	Token         string `json:"-"`
	store         *Store
	reservation   *os.File
	capacityGuard *os.File
	locks         []*AgentOperationLock
	execution     *workflowClaimExecution
	purpose       string
}

// ClaimWorkflow acquires one task and a repository-wide worker reservation.
func (s *Store) ClaimWorkflow(ctx context.Context, id string) (claim *Claim, err error) {
	err = s.withLock(ctx, func() error {
		run, owner, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		run, readErr = owner.reconcileWorkflowMembership(ctx, run)
		if readErr != nil {
			return readErr
		}

		claim, readErr = owner.claimWorkflow(ctx, run)

		return readErr
	})

	return claim, err
}

func (s *Store) claimWorkflow(ctx context.Context, run workflowInvocationFile) (claim *Claim, err error) {
	if err := harnesshost.CheckExecutable(run.HarnessBinding, run.HarnessCommand); err != nil {
		return nil, err
	}

	if err := s.prepareWorkflowClaimState(ctx, &run); err != nil {
		return nil, err
	}

	if !workflowInvocationClaimable(run) {
		return nil, fmt.Errorf("%w: invocation %s is %s", ErrWorkflowOwned, run.ID, run.Status)
	}

	record, err := s.observeWorkflowClaim(ctx, run)
	if err != nil {
		return nil, s.retainWorkflowPrerequisiteStop(&run, err)
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

func (s *Store) prepareWorkflowClaimState(ctx context.Context, run *workflowInvocationFile) error {
	restoreWorkflowCapacityWait(run)

	if run.StopIntent != "" {
		return ErrWorkflowOwned
	}

	if err := s.guardWorkflowIntegrationExecutor(ctx, *run); err != nil {
		return err
	}

	return s.reuseWorkflowValidatedTasks(ctx, run)
}

func (s *Store) observeWorkflowClaim(ctx context.Context, run workflowInvocationFile) (Record, error) {
	if err := s.checkWorkflowInputs(ctx, run); err != nil {
		return Record{}, err
	}

	if err := s.checkWorkflowTarget(ctx, run.ExecutionSelection().Target); err != nil {
		return Record{}, err
	}

	selection, err := s.readWorkflowSelection(ctx)
	if err != nil {
		return Record{}, err
	}

	member := run.Members[run.Cursor]
	request := run.ExecutionSelection()
	request.Selector = workflowstate.Selector{Extent: "ticket", Identity: member.Admission.Key}

	request, err = s.workflowLaneSelection(ctx, run, selection.records[member.Admission.Key], request)
	if err != nil {
		return Record{}, err
	}

	preview, err := selection.preview(request)
	if err != nil {
		return Record{}, err
	}

	if err := s.observePrerequisites(ctx, selection, request, &preview); err != nil {
		return Record{}, err
	}

	if err := workflowClaimReadiness(preview, member.Admission.Key); err != nil {
		return Record{}, err
	}

	record := selection.records[member.Admission.Key]
	if err := validateWorkflowClaimableContract(record); err != nil {
		return Record{}, err
	}

	return record, nil
}

func validateWorkflowClaimableContract(record Record) error {
	if !workflowTaskClaimable(record.Task) {
		return ErrWorkflowOwned
	}

	return validateWorkflowClaimContract(record)
}

func workflowClaimReadiness(preview Preview, key string) error {
	if len(preview.Members) == 1 && preview.Members[0].Readiness == "ready" {
		return nil
	}

	if len(preview.Members) == 1 && preview.Members[0].Blocker != nil {
		return workflowPrerequisiteError{blocker: *preview.Members[0].Blocker}
	}

	return fmt.Errorf("workflow task is not ready: %s", key)
}

func newWorkflowClaimToken() (string, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}

	return hex.EncodeToString(token), nil
}

func workflowTaskClaimable(task *Task) bool {
	return task.AttemptID == "" && task.ClaimGeneration != math.MaxUint64
}

func (s *Store) beginWorkflowAttempt(run *workflowInvocationFile, record *Record, claim *Claim) error {
	count := 0

	for _, attempt := range run.Attempts {
		if attempt.TaskKey == record.Task.Key && attempt.Purpose == claim.purpose {
			count++
		}
	}

	limit, err := s.workflowGrantedLimit(*run, "executor")
	if err != nil {
		return err
	}

	if slices.Contains([]string{workflowSliceReviewPurpose, workflowIntegrationValidationPurpose, workflowMergerPurpose}, claim.purpose) {
		limit = workflowstate.PageLimit
	}

	if int64(count) >= limit || len(run.Attempts) >= workflowstate.RecordLimit {
		return fmt.Errorf("workflow worker attempt budget exhausted")
	}

	now := time.Now().UTC()
	claim.AttemptID = fmt.Sprintf("%s-attempt-%d", run.ID, len(run.Attempts)+1)
	claim.Generation = record.Task.ClaimGeneration + 1

	owner, err := workflowProcessIdentity(os.Getpid())
	if err != nil {
		return err
	}

	attempt := Attempt{CapacitySlot: workflowWorkerSlotNumber(claim.reservation.Name()), Purpose: claim.purpose, OwnerProcess: &owner, Replaces: workflowReplacedAttempt(*run, record.Task.Key, claim.purpose), ID: claim.AttemptID, TaskKey: record.Task.Key, Generation: claim.Generation, InputRevision: record.Revision, ContractHash: record.Task.ContractHash, LeaseUntil: now.Add(time.Duration(run.Policy.WorkerTimeoutSeconds) * time.Second), StartedAt: now, Outcome: "running"}
	run.Attempts = append(run.Attempts, attempt)
	run.Owners[claim.AttemptID] = contenthash.Sum([]byte(claim.Token))
	run.Status = "running"

	record.Task.AttemptID, record.Task.ClaimGeneration = claim.AttemptID, claim.Generation
	if claim.purpose == "" {
		record.Task.Phase, record.Task.Status = "running", string(ticketcore.StatusInProgress)
	}

	return nil
}

func (s *Store) commitWorkflowClaim(run workflowInvocationFile, record Record, extra ...workflowstate.FileChange) error {
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

func (s *Store) workflowClaimRecordChanges(run *workflowInvocationFile, record *Record) ([]workflowstate.FileChange, error) {
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

func workflowClaimRecordHash(record Record, files []workflowstate.FileChange) string {
	hash := contenthash.Sum(files[0].After)
	if len(files) > 1 {
		hash += ":" + contenthash.Sum(record.Body)
	}

	return hash
}

// CheckWorkflowClaim rejects expired credentials and any unaccepted file image.
func (s *Store) CheckWorkflowClaim(ctx context.Context, claim *Claim) error {
	if claim == nil || claim.store == nil || claim.execution == nil {
		return ErrWorkflowOwned
	}

	return s.withLock(ctx, func() error {
		_, _, err := s.checkWorkflowClaim(ctx, claim)

		return err
	})
}

func (s *Store) checkWorkflowClaim(ctx context.Context, claim *Claim) (workflowInvocationFile, Record, error) {
	return s.checkWorkflowCredential(ctx, claim, true)
}

func (s *Store) checkWorkflowCredential(ctx context.Context, claim *Claim, lease bool) (workflowInvocationFile, Record, error) {
	run, owner, err := s.locateWorkflowClaim(ctx, claim)
	if err != nil {
		return run, Record{}, err
	}

	if owner.layout.Root() != claim.store.layout.Root() || len(run.Attempts) == 0 {
		return run, Record{}, ErrWorkflowOwned
	}

	if err := owner.checkWorkflowInputs(ctx, run); err != nil {
		return run, Record{}, err
	}

	attempt := run.Attempts[len(run.Attempts)-1]
	if !workflowCredentialValid(run, claim, attempt, lease) {
		return run, Record{}, ErrWorkflowOwned
	}

	record, err := owner.readWorkflowAttemptTask(run, attempt)
	if err != nil {
		return run, record, err
	}

	if record.Task.AttemptID != claim.AttemptID || record.Task.ClaimGeneration != claim.Generation {
		return run, record, workflowstate.ErrConflict
	}

	return run, record, nil
}

// CheckpointWorkflow yields at a quiescent task boundary without claiming success.
func (s *Store) CheckpointWorkflow(ctx context.Context, claim *Claim) error {
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

func workflowAttemptCredentialMatches(run workflowInvocationFile, claim *Claim, attempt Attempt) bool {
	return run.Status == "running" && attempt.Outcome == "running" && attempt.Purpose == claim.purpose && attempt.ID == claim.AttemptID && attempt.Generation == claim.Generation && run.Owners[attempt.ID] == contenthash.Sum([]byte(claim.Token))
}

func workflowReplacedAttempt(run workflowInvocationFile, key, purpose string) string {
	for _, attempt := range slices.Backward(run.Attempts) {
		if attempt.TaskKey == key && attempt.Purpose == purpose && attempt.Outcome != "validated" {
			return attempt.ID
		}
	}

	return ""
}

func (s *Store) readWorkflowAttemptTask(run workflowInvocationFile, attempt Attempt) (Record, error) {
	cursor := run.Cursor
	if attempt.Purpose == workflowSliceValidationPurpose || attempt.Purpose == workflowSliceReviewPurpose || attempt.Purpose == workflowIntegrationValidationPurpose || attempt.Purpose == workflowMergerPurpose {
		cursor--
	}

	if cursor < 0 || cursor >= len(run.Members) || run.Members[cursor].Admission.Key != attempt.TaskKey {
		return Record{}, workflowstate.ErrConflict
	}

	return s.readRecord(run.Members[cursor].Admission.Path)
}

func workflowInvocationClaimable(run workflowInvocationFile) bool {
	return run.StopIntent == "" && !workflowPrerequisitePaused(run) && !workflowFailedTaskPaused(run) && slices.Contains([]string{"pending", "paused", "waiting_capacity"}, run.Status) && run.Cursor < len(run.Members)
}

func workflowCredentialValid(run workflowInvocationFile, claim *Claim, attempt Attempt, lease bool) bool {
	return workflowAttemptCredentialMatches(run, claim, attempt) && (!lease || time.Now().Before(attempt.LeaseUntil))
}
