package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowMutationReservation struct {
	Schema             string                        `json:"schema"`
	State              string                        `json:"state"`
	ProjectID          string                        `json:"project_id"`
	Namespace          string                        `json:"namespace"`
	Binding            workflowstate.MutationBinding `json:"binding"`
	Operation          string                        `json:"operation"`
	ArgumentDigest     string                        `json:"argument_digest"`
	ExpectedRevision   uint64                        `json:"expected_revision"`
	ObservedSourceHash string                        `json:"observed_source_hash"`
	ReceiptRef         string                        `json:"receipt_ref"`
}

// Reserve accepted identities before preparing any effect. A failed preparation
// can retry its original payload, but cannot exchange it for another operation.
func (s *Store) reserveMutationBinding(request MutationRequest, digest string) error {
	reservation := workflowMutationReservation{Schema: "chr-record-binding/v1", State: "reserved", ProjectID: s.projectID, Namespace: s.layout.Namespace(), Binding: request.Binding, Operation: workflowMutationOperation(request), ArgumentDigest: digest, ExpectedRevision: request.ExpectedRevision, ObservedSourceHash: request.Record.SourceHash, ReceiptRef: workflowstate.MutationReceiptPath(request.Binding.RequestID)}

	path, err := s.safePath(filepath.Join("run", "operation-bindings", contenthash.Sum([]byte(request.Binding.RequestID))+".json"))
	if err != nil {
		return err
	}

	data, err := opsfile.Read(path)
	if err == nil {
		return checkWorkflowMutationReservation(data, reservation)
	}

	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err := s.checkMutationCapacity(); err != nil {
		return err
	}

	encoded, err := json.Marshal(reservation)
	if err != nil {
		return err
	}

	return opsfile.WriteAtomic(path, encoded)
}

func checkWorkflowMutationReservation(data []byte, expected workflowMutationReservation) error {
	var retained workflowMutationReservation
	if err := json.Unmarshal(data, &retained); err != nil {
		return err
	}
	// Every field is scalar; exact comparison also rejects a newer schema or state.
	if retained != expected {
		return fmt.Errorf("%w: mutation identity is bound to another payload", workflowstate.ErrConflict)
	}

	return nil
}
