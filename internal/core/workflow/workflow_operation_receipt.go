package workflow

import (
	"encoding/json"
	"errors"
	"os"

	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowOperationState string

const (
	workflowOperationReserved  workflowOperationState = "reserved"
	workflowOperationConfirmed workflowOperationState = "confirmed"
)

type workflowOperationReceiptRead[T any] struct {
	Receipt T
	Found   bool
}

// Presence is established by the contained file and successful decode only.
// Each operation must validate its own checksum, identity and revision guards.
func readWorkflowOperationReceipt[T any](store *Store, requestID string) (workflowOperationReceiptRead[T], error) {
	path, err := store.safePath(workflowstate.MutationReceiptPath(requestID))
	if err != nil {
		return workflowOperationReceiptRead[T]{}, err
	}

	data, err := opsfile.Read(path)
	if errors.Is(err, os.ErrNotExist) {
		return workflowOperationReceiptRead[T]{}, nil
	}

	if err != nil {
		return workflowOperationReceiptRead[T]{}, err
	}

	var receipt T
	if err := json.Unmarshal(data, &receipt); err != nil {
		return workflowOperationReceiptRead[T]{}, err
	}

	return workflowOperationReceiptRead[T]{Receipt: receipt, Found: true}, nil
}
