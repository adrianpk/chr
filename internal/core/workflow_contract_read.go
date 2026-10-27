package core

import (
	"context"
	"fmt"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ReadContract observes one bounded canonical input under repository exclusion.
func (s *WorkflowStore) ReadContract(ctx context.Context, reference, expectedHash string) (data []byte, err error) {
	err = s.withRepositoryLock(ctx, func() error { return s.prepareWorkflowRead(ctx) }, func() error {
		var readErr error

		data, readErr = s.readContractImage(reference, expectedHash)

		return readErr
	})

	return data, err
}

func (s *WorkflowStore) readContractImage(reference, expectedHash string) ([]byte, error) {
	path, err := s.safePath(reference)
	if err != nil {
		return nil, err
	}

	data, err := readWorkflowFile(path)
	if err != nil {
		return nil, err
	}

	if contenthash.Sum(data) != expectedHash {
		return nil, fmt.Errorf("%w: contract image changed", workflowstate.ErrConflict)
	}

	return data, nil
}

func (s *WorkflowStore) guardMutationContracts(record WorkflowRecord, expected uint64, snapshot WorkflowSnapshot) error {
	if record.Set == nil {
		return nil
	}

	if expected == 0 {
		_, err := s.readContractImage(record.Set.PlanRef, record.Set.PlanHash)

		return err
	}

	return s.guardMutationSlices(record, snapshot)
}

func (s *WorkflowStore) guardMutationSlices(record WorkflowRecord, snapshot WorkflowSnapshot) error {
	previous, err := s.readRecord(record.Path)
	if err != nil {
		return err
	}

	for _, slice := range record.Set.Slices {
		if !newMutationSlice(previous, slice) {
			continue
		}

		if previous.Set.Status != "proposed" || slice.Phase != "pending" {
			return fmt.Errorf("%w: slice contract changes require approved amendment authority", ErrWorkflowOwned)
		}

		if _, err := s.readContractImage(slice.ContractRef, slice.ContractHash); err != nil {
			return err
		}

		members := workflowMutationSetMembers(snapshot, record.ID)
		if slice.TaskInputHash != WorkflowTaskInputHash(members, slice.ID) {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func newMutationSlice(previous WorkflowRecord, next WorkflowSlice) bool {
	for _, old := range previous.Set.Slices {
		if old.ID == next.ID {
			return old.ContractRef != next.ContractRef || old.ContractHash != next.ContractHash || old.TaskInputHash != next.TaskInputHash
		}
	}

	return true
}

func workflowMutationSetMembers(snapshot WorkflowSnapshot, setID string) []WorkflowRecord {
	var members []WorkflowRecord

	for _, member := range snapshot.Records {
		if member.Task != nil && member.Task.SetID == setID {
			members = append(members, member)
		}
	}

	return members
}
