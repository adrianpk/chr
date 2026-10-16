package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// WorkflowMutationBinding identifies a caller-neutral, immutable mutation payload.
// Routing identities belong to adapters and must never enter domain receipts.
type WorkflowMutationBinding struct {
	RequestID string `json:"request_id"`
	Operator  string `json:"operator"`
}

// WorkflowMutationRequest creates or replaces a record through its existing owner.
// A zero expected revision means creation; replacement requires its exact source image.
type WorkflowMutationRequest struct {
	Binding          WorkflowMutationBinding `json:"binding"`
	Title            string                  `json:"title"`
	Record           WorkflowRecord          `json:"record"`
	ExpectedRevision uint64                  `json:"expected_revision"`
}

// WorkflowMutationResult is the original confirmed record and its durable receipt.
type WorkflowMutationResult struct {
	RequestID      string         `json:"request_id"`
	ArgumentDigest string         `json:"argument_digest"`
	ReceiptRef     string         `json:"receipt_ref"`
	Record         WorkflowRecord `json:"record"`
}

type workflowMutationReceipt struct {
	State            string                 `json:"state"`
	Operation        string                 `json:"operation"`
	ExpectedRevision uint64                 `json:"expected_revision"`
	BeforeSourceHash string                 `json:"before_source_hash"`
	Schema           string                 `json:"schema"`
	ProjectID        string                 `json:"project_id"`
	Namespace        string                 `json:"namespace"`
	Operator         string                 `json:"operator"`
	Result           WorkflowMutationResult `json:"result"`
	Checksum         string                 `json:"checksum"`
}

// ApplyRecordMutation journals the record and its result receipt under repository
// exclusion. Recovery completes the same reserved images before observing replay.
func (s *WorkflowStore) ApplyRecordMutation(ctx context.Context, request WorkflowMutationRequest) (result WorkflowMutationResult, err error) {
	request, digest, err := s.normalizeMutation(request)
	if err != nil {
		return result, err
	}

	err = s.withLock(ctx, func() error {
		retained, found, readErr := s.readMutationReceipt(request, digest)
		if readErr != nil || found {
			result = retained

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

func (s *WorkflowStore) normalizeMutation(request WorkflowMutationRequest) (WorkflowMutationRequest, string, error) {
	request.Binding.RequestID = strings.TrimSpace(request.Binding.RequestID)
	request.Binding.Operator = strings.TrimSpace(request.Binding.Operator)

	request.Title = strings.TrimSpace(request.Title)
	if strings.TrimSpace(s.projectID) == "" {
		return request, "", errors.New("mutation owner requires a project identity")
	}

	if err := validateWorkflowMutationBinding(request.Binding); err != nil {
		return request, "", err
	}

	if request.Record.Metadata == nil {
		request.Record.Metadata = map[string]json.RawMessage{}
	}

	if request.Record.Domain == nil {
		request.Record.Domain = map[string]json.RawMessage{}
	}

	data, err := json.Marshal(struct {
		ProjectID, Namespace string
		Request              WorkflowMutationRequest
	}{s.projectID, s.layout.Namespace(), request})
	if err != nil {
		return request, "", err
	}

	if len(data) > 64*1024 {
		return request, "", errors.New("record mutation arguments exceed 64 KiB")
	}

	return request, workflowHash(data), nil
}

func validateWorkflowMutationBinding(binding WorkflowMutationBinding) error {
	if !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`).MatchString(binding.RequestID) || binding.Operator == "" || len(binding.Operator) > 128 || strings.ContainsAny(binding.Operator, "\x00\r\n") {
		return errors.New("opaque mutation request and operator identities are required")
	}

	return nil
}

func workflowMutationReceiptPath(requestID string) string {
	return filepath.ToSlash(filepath.Join("run", "operation-receipts", workflowHash([]byte(requestID))+".json"))
}

func (s *WorkflowStore) readMutationReceipt(request WorkflowMutationRequest, digest string) (WorkflowMutationResult, bool, error) {
	path, err := s.safePath(workflowMutationReceiptPath(request.Binding.RequestID))
	if err != nil {
		return WorkflowMutationResult{}, false, err
	}

	data, err := readWorkflowFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return WorkflowMutationResult{}, false, nil
	}

	if err != nil {
		return WorkflowMutationResult{}, false, err
	}

	var receipt workflowMutationReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return WorkflowMutationResult{}, false, err
	}

	if err := s.validateMutationReceipt(receipt, request, digest); err != nil {
		return WorkflowMutationResult{}, false, err
	}

	return receipt.Result, true, nil
}

func (s *WorkflowStore) validateMutationReceipt(receipt workflowMutationReceipt, request WorkflowMutationRequest, digest string) error {
	checksum := receipt.Checksum
	receipt.Checksum = ""

	encoded, err := json.Marshal(receipt)
	if err != nil {
		return err
	}

	if receipt.Schema != "chr-record-mutation/v1" || receipt.ProjectID != s.projectID || receipt.Namespace != s.layout.Namespace() || checksum != workflowHash(encoded) {
		return fmt.Errorf("%w: invalid record mutation receipt", ErrWorkflowConflict)
	}

	return validateWorkflowMutationResult(receipt, request, digest)
}

func validateWorkflowMutationResult(receipt workflowMutationReceipt, request WorkflowMutationRequest, digest string) error {
	result := receipt.Result
	if receipt.State != "confirmed" || receipt.Operation != workflowMutationOperation(request) || receipt.ExpectedRevision != request.ExpectedRevision || receipt.BeforeSourceHash != request.Record.SourceHash {
		return fmt.Errorf("%w: mutation receipt guard changed", ErrWorkflowConflict)
	}

	if result.RequestID != request.Binding.RequestID || result.ArgumentDigest != digest || receipt.Operator != request.Binding.Operator || result.ReceiptRef != workflowMutationReceiptPath(request.Binding.RequestID) {
		return fmt.Errorf("%w: mutation request payload changed", ErrWorkflowConflict)
	}

	return result.Record.Validate()
}

func (s *WorkflowStore) mutationReceiptChange(request WorkflowMutationRequest, result WorkflowMutationResult) (workflowFileChange, error) {
	receipt := workflowMutationReceipt{State: "confirmed", Operation: workflowMutationOperation(request), ExpectedRevision: request.ExpectedRevision, BeforeSourceHash: request.Record.SourceHash, Schema: "chr-record-mutation/v1", ProjectID: s.projectID, Namespace: s.layout.Namespace(), Operator: request.Binding.Operator, Result: result}

	data, err := json.Marshal(receipt)
	if err != nil {
		return workflowFileChange{}, err
	}

	receipt.Checksum = workflowHash(data)
	data, err = json.Marshal(receipt)

	return workflowFileChange{Path: result.ReceiptRef, After: data}, err
}

func (s *WorkflowStore) applyNewMutation(ctx context.Context, request WorkflowMutationRequest, digest string) (result WorkflowMutationResult, err error) {
	guard := ""
	if request.ExpectedRevision == 0 {
		guard, err = s.acquireWorkflowAllocation()
		if err != nil {
			return result, err
		}
		defer func() { err = errors.Join(err, s.releaseUnusedMutationAllocation(guard)) }()
	}

	record, changes, err := s.prepareMutation(ctx, request)
	if err != nil {
		return result, err
	}

	result = WorkflowMutationResult{RequestID: request.Binding.RequestID, ArgumentDigest: digest, ReceiptRef: workflowMutationReceiptPath(request.Binding.RequestID), Record: record}

	receipt, err := s.mutationReceiptChange(request, result)
	if err != nil {
		return result, err
	}

	coordinator, err := s.workflowCoordinatorFiles(ctx)
	if err != nil {
		return result, err
	}

	changes = append(changes, coordinator...)

	return result, s.commitAllocatedFiles(append(changes, receipt), guard)
}

func (s *WorkflowStore) checkMutationCapacity() error {
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

func (s *WorkflowStore) releaseUnusedMutationAllocation(guard string) error {
	path, err := s.safePath(workflowJournalPath)
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

func (s *WorkflowStore) prepareMutation(ctx context.Context, request WorkflowMutationRequest) (WorkflowRecord, []workflowFileChange, error) {
	record, err := s.prepareMutationRecord(request)
	if err != nil {
		return record, nil, err
	}

	if err := s.guardWrite(&record, request.ExpectedRevision); err != nil {
		return record, nil, err
	}

	snapshot, err := s.readSnapshot(ctx)
	if err != nil {
		return record, nil, err
	}

	if err := s.guardMutationContracts(record, request.ExpectedRevision, snapshot); err != nil {
		return record, nil, err
	}

	if err := replaceWorkflowSnapshot(&snapshot, record); err != nil {
		return record, nil, err
	}

	files, err := s.encodeChanges(record)
	if err != nil {
		return record, nil, err
	}

	if err := guardMutationCreation(files, request.ExpectedRevision); err != nil {
		return record, nil, err
	}

	files, err = s.workflowLooseWriteChanges(ctx, record, request.ExpectedRevision, files)
	if err != nil {
		return record, nil, err
	}

	record.SourceHash = workflowClaimRecordHash(record, files)

	return record, files, nil
}

func mergeMutationMetadata(base, extra map[string]json.RawMessage) map[string]json.RawMessage {
	for key, value := range extra {
		if _, reserved := base[key]; !reserved {
			base[key] = value
		}
	}

	return base
}

func guardMutationCreation(files []workflowFileChange, expected uint64) error {
	if expected != 0 {
		return nil
	}

	for _, file := range files {
		if file.Before != "" {
			return ErrWorkflowConflict
		}
	}

	return nil
}

func workflowMutationOperation(request WorkflowMutationRequest) string {
	if request.ExpectedRevision == 0 {
		return "record.create"
	}

	return "record.update"
}

func (s *WorkflowStore) prepareMutationRecord(request WorkflowMutationRequest) (WorkflowRecord, error) {
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
		allocated, err := s.materializationEnvelope(record.Kind, request.Title, nextOpsArtifactNumber(s.layout.Root()), time.Now().UTC())
		if err != nil {
			return record, err
		}

		record.WorkflowIdentity, record.Path = allocated.WorkflowIdentity, allocated.Path
		record.Metadata = mergeMutationMetadata(allocated.Metadata, record.Metadata)
	}

	return record, nil
}

func (s *WorkflowStore) validateMutationScope(record WorkflowRecord) error {
	if record.ProjectID != "" && record.ProjectID != s.projectID {
		return fmt.Errorf("%w: foreign mutation project", ErrWorkflowConflict)
	}

	if record.Namespace != "" && record.Namespace != s.layout.Namespace() {
		return fmt.Errorf("%w: foreign mutation namespace", ErrWorkflowConflict)
	}

	return nil
}
