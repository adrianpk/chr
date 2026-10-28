package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// MutationRequest creates or replaces a record through its existing owner.
// A zero expected revision means creation; replacement requires its exact source image.
type MutationRequest struct {
	Binding          workflowstate.MutationBinding `json:"binding"`
	Title            string                        `json:"title"`
	Record           workflowstate.Record          `json:"record"`
	ExpectedRevision uint64                        `json:"expected_revision"`
}

// MutationResult is the original confirmed record and its durable receipt.
type MutationResult struct {
	RequestID      string               `json:"request_id"`
	ArgumentDigest string               `json:"argument_digest"`
	ReceiptRef     string               `json:"receipt_ref"`
	Record         workflowstate.Record `json:"record"`
}

// MutationLookup distinguishes a confirmed replay from a request without a receipt.
type MutationLookup struct {
	Result MutationResult
	Found  bool
}

type workflowNormalizedMutation struct {
	Request MutationRequest
	Digest  string
}

type workflowPreparedMutation struct {
	Record  workflowstate.Record
	Changes []workflowstate.FileChange
}

type workflowMutationReceipt struct {
	State            workflowOperationState `json:"state"`
	Operation        string                 `json:"operation"`
	ExpectedRevision uint64                 `json:"expected_revision"`
	BeforeSourceHash string                 `json:"before_source_hash"`
	Schema           string                 `json:"schema"`
	ProjectID        string                 `json:"project_id"`
	Namespace        string                 `json:"namespace"`
	Operator         string                 `json:"operator"`
	Result           MutationResult         `json:"result"`
	Checksum         string                 `json:"checksum"`
}

// ApplyRecordMutation journals the record and its result receipt under repository
// exclusion. Recovery completes the same reserved images before observing replay.
func (s *Store) ApplyRecordMutation(ctx context.Context, request MutationRequest) (result MutationResult, err error) {
	normalized, err := s.normalizeMutation(request)
	if err != nil {
		return result, err
	}

	request, digest := normalized.Request, normalized.Digest

	err = s.withLock(ctx, func() error {
		retained, readErr := s.readMutationReceipt(request, digest)
		if readErr != nil || retained.Found {
			result = retained.Result

			return readErr
		}

		if err := s.reserveMutationBinding(request, digest); err != nil {
			return err
		}

		var applyErr error

		result, applyErr = s.applyNewMutation(ctx, request, digest)

		return applyErr
	})

	return result, err
}

func (s *Store) normalizeMutation(request MutationRequest) (workflowNormalizedMutation, error) {
	request.Binding.RequestID = strings.TrimSpace(request.Binding.RequestID)
	request.Binding.Operator = strings.TrimSpace(request.Binding.Operator)

	request.Title = strings.TrimSpace(request.Title)
	if strings.TrimSpace(s.projectID) == "" {
		return workflowNormalizedMutation{Request: request}, errors.New("mutation owner requires a project identity")
	}

	if err := request.Binding.Validate(); err != nil {
		return workflowNormalizedMutation{Request: request}, err
	}

	if request.Record.Metadata == nil {
		request.Record.Metadata = map[string]json.RawMessage{}
	}

	if request.Record.Domain == nil {
		request.Record.Domain = map[string]json.RawMessage{}
	}

	data, err := json.Marshal(struct {
		ProjectID, Namespace string
		Request              MutationRequest
	}{s.projectID, s.layout.Namespace(), request})
	if err != nil {
		return workflowNormalizedMutation{Request: request}, err
	}

	if len(data) > 64*1024 {
		return workflowNormalizedMutation{Request: request}, errors.New("record mutation arguments exceed 64 KiB")
	}

	return workflowNormalizedMutation{Request: request, Digest: contenthash.Sum(data)}, nil
}

func (s *Store) readMutationReceipt(request MutationRequest, digest string) (MutationLookup, error) {
	observed, err := readWorkflowOperationReceipt[workflowMutationReceipt](s, request.Binding.RequestID)
	if err != nil || !observed.Found {
		return MutationLookup{}, err
	}

	if err := s.validateMutationReceipt(observed.Receipt, request, digest); err != nil {
		return MutationLookup{}, err
	}

	return MutationLookup{Result: observed.Receipt.Result, Found: true}, nil
}

func (s *Store) validateMutationReceipt(receipt workflowMutationReceipt, request MutationRequest, digest string) error {
	checksum := receipt.Checksum
	receipt.Checksum = ""

	encoded, err := json.Marshal(receipt)
	if err != nil {
		return err
	}

	if receipt.Schema != "chr-record-mutation/v1" || receipt.ProjectID != s.projectID || receipt.Namespace != s.layout.Namespace() || checksum != contenthash.Sum(encoded) {
		return fmt.Errorf("%w: invalid record mutation receipt", workflowstate.ErrConflict)
	}

	return validateWorkflowMutationResult(receipt, request, digest)
}

func validateWorkflowMutationResult(receipt workflowMutationReceipt, request MutationRequest, digest string) error {
	result := receipt.Result
	if receipt.State != workflowOperationConfirmed || receipt.Operation != workflowMutationOperation(request) || receipt.ExpectedRevision != request.ExpectedRevision || receipt.BeforeSourceHash != request.Record.SourceHash {
		return fmt.Errorf("%w: mutation receipt guard changed", workflowstate.ErrConflict)
	}

	if result.RequestID != request.Binding.RequestID || result.ArgumentDigest != digest || receipt.Operator != request.Binding.Operator || result.ReceiptRef != workflowstate.MutationReceiptPath(request.Binding.RequestID) {
		return fmt.Errorf("%w: mutation request payload changed", workflowstate.ErrConflict)
	}

	return validateWorkflowRecord(result.Record)
}

func (s *Store) mutationReceiptChange(request MutationRequest, result MutationResult) (workflowstate.FileChange, error) {
	receipt := workflowMutationReceipt{State: workflowOperationConfirmed, Operation: workflowMutationOperation(request), ExpectedRevision: request.ExpectedRevision, BeforeSourceHash: request.Record.SourceHash, Schema: "chr-record-mutation/v1", ProjectID: s.projectID, Namespace: s.layout.Namespace(), Operator: request.Binding.Operator, Result: result}

	data, err := json.Marshal(receipt)
	if err != nil {
		return workflowstate.FileChange{}, err
	}

	receipt.Checksum = contenthash.Sum(data)
	data, err = json.Marshal(receipt)

	return workflowstate.FileChange{Path: result.ReceiptRef, After: data}, err
}

func (s *Store) applyNewMutation(ctx context.Context, request MutationRequest, digest string) (result MutationResult, err error) {
	guard := ""
	if request.ExpectedRevision == 0 {
		guard, err = s.acquireWorkflowAllocation()
		if err != nil {
			return result, err
		}
		defer func() { err = errors.Join(err, s.releaseUnusedMutationAllocation(guard)) }()
	}

	prepared, err := s.prepareMutation(ctx, request)
	if err != nil {
		return result, err
	}

	result = MutationResult{RequestID: request.Binding.RequestID, ArgumentDigest: digest, ReceiptRef: workflowstate.MutationReceiptPath(request.Binding.RequestID), Record: prepared.Record}

	receipt, err := s.mutationReceiptChange(request, result)
	if err != nil {
		return result, err
	}

	coordinator, err := s.workflowCoordinatorFiles(ctx)
	if err != nil {
		return result, err
	}

	changes := append(prepared.Changes, coordinator...)

	if err := ctx.Err(); err != nil {
		return result, err
	}

	return result, s.commitAllocatedFiles(append(changes, receipt), guard)
}

func (s *Store) checkMutationCapacity() error {
	path, err := s.safePath(filepath.Join("run", "operation-bindings"))
	if err != nil {
		return err
	}

	dir, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	entries, readErr := dir.ReadDir(4097)

	closeErr := dir.Close()
	if readErr != nil && !errors.Is(readErr, io.EOF) || closeErr != nil {
		return errors.Join(readErr, closeErr)
	}

	if len(entries) >= 4096 {
		return errors.New("record mutation receipt capacity exhausted")
	}

	return nil
}

func (s *Store) releaseUnusedMutationAllocation(guard string) error {
	path, err := s.safePath(workflowstate.JournalPath)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return s.releaseWorkflowAllocation(guard)
}

func (s *Store) prepareMutation(ctx context.Context, request MutationRequest) (workflowPreparedMutation, error) {
	record, err := s.prepareMutationRecord(request)
	if err != nil {
		return workflowPreparedMutation{Record: record}, err
	}

	if err := s.guardWrite(&record, request.ExpectedRevision); err != nil {
		return workflowPreparedMutation{Record: record}, err
	}

	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return workflowPreparedMutation{Record: record}, err
	}

	if err := s.guardMutationContracts(record, request.ExpectedRevision, snapshot); err != nil {
		return workflowPreparedMutation{Record: record}, err
	}

	if err := replaceWorkflowSnapshot(&snapshot, record); err != nil {
		return workflowPreparedMutation{Record: record}, err
	}

	files, err := s.encodeChanges(record)
	if err != nil {
		return workflowPreparedMutation{Record: record}, err
	}

	if err := guardMutationCreation(files, request.ExpectedRevision); err != nil {
		return workflowPreparedMutation{Record: record}, err
	}

	files, err = s.workflowLooseWriteChanges(ctx, record, request.ExpectedRevision, files)
	if err != nil {
		return workflowPreparedMutation{Record: record}, err
	}

	record.SourceHash = workflowClaimRecordHash(record, files)

	return workflowPreparedMutation{Record: record, Changes: files}, nil
}

func mergeMutationMetadata(base, extra map[string]json.RawMessage) map[string]json.RawMessage {
	for key, value := range extra {
		if _, reserved := base[key]; !reserved {
			base[key] = value
		}
	}

	return base
}

func guardMutationCreation(files []workflowstate.FileChange, expected uint64) error {
	if expected != 0 {
		return nil
	}

	for _, file := range files {
		if file.Before != "" {
			return workflowstate.ErrConflict
		}
	}

	return nil
}

func workflowMutationOperation(request MutationRequest) string {
	if request.ExpectedRevision == 0 {
		return "record.create"
	}

	return "record.update"
}

func (s *Store) prepareMutationRecord(request MutationRequest) (workflowstate.Record, error) {
	record := request.Record
	if err := s.validateMutationScope(record); err != nil {
		return record, err
	}

	if record.Kind != "ticket" && record.Kind != "plan" {
		return record, errors.New("unsupported record mutation kind")
	}

	if request.ExpectedRevision == 0 && request.Title == "" {
		return record, errors.New("creation requires a record title")
	}

	if request.ExpectedRevision == 0 {
		allocated, err := s.materializationEnvelope(record.Kind, request.Title, opscore.NextArtifactNumber(s.layout.Root()), time.Now().UTC())
		if err != nil {
			return record, err
		}

		record.Identity, record.Path = allocated.Identity, allocated.Path
		record.Metadata = mergeMutationMetadata(allocated.Metadata, record.Metadata)
	}

	return record, nil
}

func (s *Store) validateMutationScope(record workflowstate.Record) error {
	if record.ProjectID != "" && record.ProjectID != s.projectID {
		return fmt.Errorf("%w: foreign mutation project", workflowstate.ErrConflict)
	}

	if record.Namespace != "" && record.Namespace != s.layout.Namespace() {
		return fmt.Errorf("%w: foreign mutation namespace", workflowstate.ErrConflict)
	}

	return nil
}
