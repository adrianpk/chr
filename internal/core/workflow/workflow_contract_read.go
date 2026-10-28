package workflow

import (
	"context"
	"fmt"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// ReadContract observes one bounded canonical input under repository exclusion.
func (s *Store) ReadContract(ctx context.Context, reference, expectedHash string) (data []byte, err error) {
	err = s.withRepositoryLock(ctx, func() error { return s.prepareWorkflowRead(ctx) }, func() error {
		var readErr error

		data, readErr = s.readContractImage(reference, expectedHash)

		return readErr
	})

	return data, err
}

func (s *Store) readContractImage(reference, expectedHash string) ([]byte, error) {
	path, err := s.safePath(reference)
	if err != nil {
		return nil, err
	}

	data, err := opsfile.Read(path)
	if err != nil {
		return nil, err
	}

	if contenthash.Sum(data) != expectedHash {
		return nil, fmt.Errorf("%w: contract image changed", workflowstate.ErrConflict)
	}

	return data, nil
}

func (s *Store) guardMutationContracts(record workflowstate.Record, expected uint64, snapshot workflowstate.Snapshot) error {
	if record.Set == nil {
		return nil
	}

	if expected == 0 {
		_, err := s.readContractImage(record.Set.PlanRef, record.Set.PlanHash)

		return err
	}

	return s.guardMutationSlices(record, snapshot)
}

func (s *Store) guardMutationSlices(record workflowstate.Record, snapshot workflowstate.Snapshot) error {
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
		if slice.TaskInputHash != workflowstate.TaskInputHash(members, slice.ID) {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func newMutationSlice(previous workflowstate.Record, next workflowstate.WorkflowSlice) bool {
	for _, old := range previous.Set.Slices {
		if old.ID == next.ID {
			return old.ContractRef != next.ContractRef || old.ContractHash != next.ContractHash || old.TaskInputHash != next.TaskInputHash
		}
	}

	return true
}

func workflowMutationSetMembers(snapshot workflowstate.Snapshot, setID string) []workflowstate.Record {
	var members []workflowstate.Record

	for _, member := range snapshot.Records {
		if member.Task != nil && member.Task.SetID == setID {
			members = append(members, member)
		}
	}

	return members
}
