package workflow

import (
	"cmp"
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"syscall"
	"time"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// RoleReservation exposes live owners without their claim credentials.
type RoleReservation struct {
	InvocationID string
	Namespace    string
	AttemptID    string
	Role         string
	LeaseUntil   time.Time
	Owner        ProcessIdentity
}

// IntegrationCapacity is an observation, not permission to launch work.
// All roles consume the same repository slots, including live expired groups.
type IntegrationCapacity struct {
	Maximum         int
	Available       int
	LiveProcesses   []Process
	UnexpiredOwners []RoleReservation
}

// QueryWorkflowIntegrationCapacity observes capacity without pruning dead ledgers.
func (s *Store) QueryWorkflowIntegrationCapacity(ctx context.Context) (result IntegrationCapacity, err error) {
	err = s.withRepositoryLock(ctx, s.requireNoWorkflowTransition, func() error {
		snapshot, readErr := s.readWorkflowIntegrationSnapshot(ctx)
		if readErr != nil {
			return readErr
		}

		result, readErr = s.workflowIntegrationCapacity(ctx, snapshot.Runs)

		return readErr
	})

	return result, integrationRecordError("capacity observations", err)
}

func (s *Store) workflowIntegrationCapacity(ctx context.Context, runs []workflowInvocationFile) (IntegrationCapacity, error) {
	limits, err := s.workflowRepositoryCapacity(ctx, runs)
	if err != nil {
		return IntegrationCapacity{}, err
	}

	common, err := projectcore.ResolveRepositoryDirectory(ctx, s.layout.ProjectDir())
	if err != nil {
		return IntegrationCapacity{}, err
	}

	processes, err := observeWorkflowIntegrationProcesses(common)
	if err != nil {
		return IntegrationCapacity{}, err
	}

	result := IntegrationCapacity{Maximum: limits.Workers, Available: limits.Workers, LiveProcesses: processes}
	slots := map[int]bool{}

	for _, process := range processes {
		if process.CapacitySlot == 0 {
			result.Available = 0
		}

		slots[process.CapacitySlot] = true
	}

	if err := appendWorkflowIntegrationOwners(&result, runs, slots); err != nil {
		return result, err
	}

	if err := observeWorkflowIntegrationSlots(&result, slots, common); err != nil {
		return result, err
	}

	if err := s.appendWorkflowIntegrationClosingOwners(ctx, &result, slots); err != nil {
		return result, err
	}

	result.Available = max(0, result.Available-len(slots))

	return result, nil
}

func appendWorkflowIntegrationOwners(result *IntegrationCapacity, runs []workflowInvocationFile, slots map[int]bool) error {
	now := time.Now().UTC()

	for _, run := range runs {
		for _, attempt := range run.Attempts {
			if err := appendWorkflowIntegrationAttemptOwner(result, run, attempt, slots, now); err != nil {
				return err
			}
		}
	}

	return nil
}

func appendWorkflowIntegrationAttemptOwner(result *IntegrationCapacity, run workflowInvocationFile, attempt Attempt, slots map[int]bool, now time.Time) error {
	if attempt.OwnerProcess == nil || attempt.Outcome != "running" {
		return nil
	}

	alive, err := workflowProcessAlive(*attempt.OwnerProcess)
	if err != nil || !alive {
		return err
	}

	occupyWorkflowIntegrationSlot(result, slots, attempt.CapacitySlot)

	if !attempt.LeaseUntil.After(now) {
		return nil
	}

	role := workflowIntegrationOwnerRole(attempt.Purpose)

	result.UnexpiredOwners = append(result.UnexpiredOwners, RoleReservation{InvocationID: run.ID, Namespace: run.Namespace, AttemptID: attempt.ID, Role: role, LeaseUntil: attempt.LeaseUntil, Owner: *attempt.OwnerProcess})

	return nil
}

func occupyWorkflowIntegrationSlot(result *IntegrationCapacity, slots map[int]bool, slot int) {
	if slot == 0 {
		result.Available = 0

		return
	}

	slots[slot] = true
}

func observeWorkflowIntegrationSlot(path string) (held bool, err error) {
	file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	if err != nil {
		return false, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_SH|syscall.LOCK_NB)
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return true, nil
	}

	if err != nil {
		return false, err
	}

	return false, syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}

func observeWorkflowIntegrationSlots(result *IntegrationCapacity, slots map[int]bool, common string) error {
	for slot := 1; slot <= workflowWorkerSlotLimit; slot++ {
		held, err := observeWorkflowIntegrationSlot(filepath.Join(common, workflowWorkerSlotName(slot)))
		if err != nil {
			return err
		}

		if held {
			slots[slot] = true
		}
	}

	held, err := observeWorkflowIntegrationSlot(filepath.Join(common, "chr-workflow-worker.lock"))
	if held {
		result.Available = 0
	}

	return err
}

func (s *Store) appendWorkflowIntegrationClosingOwners(ctx context.Context, result *IntegrationCapacity, slots map[int]bool) error {
	layouts, err := s.workflowRepositoryLayouts(ctx)
	if err != nil {
		return err
	}

	for _, layout := range layouts {
		owner := NewWorkflowStore(layout, s.projectID)

		snapshot, err := owner.readSnapshot(ctx)
		if err != nil {
			return err
		}

		for _, record := range snapshot.Records {
			if record.Set == nil || record.Set.Closing == nil {
				continue
			}

			if err := owner.appendWorkflowIntegrationClosingOwner(result, slots, record); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Store) appendWorkflowIntegrationClosingOwner(result *IntegrationCapacity, slots map[int]bool, set Record) error {
	for i := range set.Set.Closing.Attempts {
		state, err := s.readWorkflowClosingAttempt(set, i)
		if err != nil {
			return err
		}

		if state.Outcome != "running" {
			continue
		}

		alive, err := workflowProcessAlive(state.Owner)
		if err != nil {
			return err
		}

		if !alive {
			continue
		}

		occupyWorkflowIntegrationSlot(result, slots, state.CapacitySlot)

		if state.LeaseUntil.After(time.Now().UTC()) {
			result.UnexpiredOwners = append(result.UnexpiredOwners, RoleReservation{InvocationID: state.InvocationID, Namespace: set.Namespace, AttemptID: workflowClosingAttemptID(set, state.Generation), Role: "closing", LeaseUntil: state.LeaseUntil, Owner: state.Owner})
		}
	}

	return nil
}

func observeWorkflowIntegrationProcesses(common string) ([]Process, error) {
	dir := filepath.Join(common, "chr-workflow-processes")

	files, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	if err != nil || len(files) > workflowstate.RecordLimit {
		return nil, errorsOrConflict(err)
	}

	processes := []Process{}

	for _, file := range files {
		if !file.Type().IsRegular() {
			return nil, workflowstate.ErrConflict
		}

		process, alive, err := observeWorkflowCapacityProcess(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}

		if alive {
			processes = append(processes, process)
		}
	}

	return processes, nil
}

// IntegrationClosureStatus reports retained receipt/fanout and gate proof.
type IntegrationClosureStatus struct {
	CanClose         bool
	ReceiptsVerified bool
	GateVerified     bool
	Blockers         []string
}

// QueryWorkflowIntegrationSetClosure never runs a gate or changes set status.
func (s *Store) QueryWorkflowIntegrationSetClosure(ctx context.Context, setPath string) (status IntegrationClosureStatus, err error) {
	err = s.withRepositoryLock(ctx, s.requireNoWorkflowTransition, func() error {
		snapshot, readErr := s.readWorkflowIntegrationSnapshot(ctx)
		if readErr != nil {
			return readErr
		}

		set, readErr := s.readRecord(setPath)
		if readErr != nil {
			return readErr
		}

		if set.Set == nil {
			return workflowstate.ErrConflict
		}

		status, readErr = s.workflowIntegrationClosureStatus(ctx, snapshot, set)

		return readErr
	})

	return status, integrationRecordError(setPath, err)
}

func (s *Store) workflowIntegrationClosureStatus(ctx context.Context, snapshot workflowIntegrationSnapshot, set Record) (IntegrationClosureStatus, error) {
	status := IntegrationClosureStatus{}
	if err := s.checkWorkflowIntegrationSetReceipts(ctx, set); err != nil {
		status.Blockers = append(status.Blockers, "unit_receipts_or_fanout_required")

		return status, err
	}

	status.ReceiptsVerified = true

	for _, journal := range snapshot.Journals {
		if journal.Attempts[len(journal.Attempts)-1].State == "awaiting_observation" {
			status.Blockers = append(status.Blockers, "publication_requires_observation")
		}
	}

	capacity, err := s.workflowIntegrationCapacity(ctx, snapshot.Runs)
	if err != nil {
		return status, err
	}

	if len(capacity.LiveProcesses) > 0 || len(capacity.UnexpiredOwners) > 0 {
		status.Blockers = append(status.Blockers, "live_process_requires_reconciliation")
	}

	if set.Set.Status == "completed" {
		if _, err := s.readWorkflowClosedSet(set); err != nil {
			return status, err
		}

		status.GateVerified = true
	} else {
		status.Blockers = append(status.Blockers, "final_gate_required")
	}

	status.CanClose = len(status.Blockers) == 0 && status.GateVerified

	return status, nil
}

func (s *Store) checkWorkflowIntegrationSetReceipts(ctx context.Context, set Record) error {
	if _, err := s.workflowClosingMembers(ctx, set.ID); err != nil {
		return err
	}

	for _, descriptor := range set.Set.Slices {
		if descriptor.Phase != "integrated" || descriptor.DeliveryRef == "" {
			return workflowstate.ErrConflict
		}

		receipt, err := s.readRecord(descriptor.DeliveryRef)
		if err != nil {
			return err
		}

		proof, _, err := s.workflowClosingDeliveryProof(ctx, receipt)
		if err != nil {
			return err
		}

		if err := s.workflowClosingSliceMembers(ctx, descriptor, receipt, proof); err != nil {
			return err
		}
	}

	return nil
}

// QueryWorkflowPublicationBarriers returns current uncertain journal heads.
func (s *Store) QueryWorkflowPublicationBarriers(ctx context.Context, targetKey string) (result []PublicationBarrier, err error) {
	err = s.withRepositoryLock(ctx, s.requireNoWorkflowTransition, func() error {
		journals, readErr := s.workflowIntegrationJournals(ctx)
		if readErr != nil {
			return readErr
		}

		for _, journal := range journals {
			last := journal.Attempts[len(journal.Attempts)-1]
			if last.State == "awaiting_observation" && (targetKey == "" || journal.TargetKey == targetKey) {
				result = append(result, workflowPublicationBarrier(journal, last))
			}
		}

		return nil
	})
	if err != nil {
		result = nil
	}

	slices.SortFunc(result, func(a, b PublicationBarrier) int {
		return cmp.Compare(a.Submission.Attempt, b.Submission.Attempt)
	})

	return result, integrationRecordError("publication barriers", err)
}

func workflowIntegrationOwnerRole(purpose string) string {
	switch purpose {
	case "":
		return "executor"
	case workflowSliceReviewPurpose:
		return "reviewer"
	default:
		return purpose
	}
}
