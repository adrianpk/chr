package core

import (
	"encoding/json"
	"math"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) workflowNativeReceiptProjection(run *workflowInvocationFile, entry WorkflowIntegrationEntry, last WorkflowSubmission) (workflowNativeProjection, error) {
	projection := workflowNativeProjection{entry: entry}
	if last.State != "confirmed" || entry.State == "integrated" || entry.State == "blocked" {
		return projection, nil
	}

	if entry.State == "finalizing" {
		return projection, s.checkWorkflowIntegrationEntryReceipt(*run, entry)
	}

	if entry.Revision == math.MaxUint64 || entry.StateRevision == math.MaxUint64 {
		return projection, workflowIntegrationConflict("receipt projection revision exhausted")
	}

	files, err := s.projectWorkflowNativeReceipt(run, &entry, last)
	if err != nil {
		return projection, err
	}

	entry.Revision++
	entry.StateRevision++
	entry.UpdatedAt = time.Now().UTC()
	projectWorkflowIntegrationPhase(run, entry)

	return workflowNativeProjection{entry: entry, files: files, changed: true}, nil
}

func (s *WorkflowStore) projectWorkflowNativeReceipt(run *workflowInvocationFile, entry *WorkflowIntegrationEntry, last WorkflowSubmission) ([]workflowFileChange, error) {
	if last.State != "confirmed" {
		return nil, nil
	}

	ref := last.Native.ReceiptRef
	if ref == nil {
		return nil, workflowstate.ErrConflict
	}

	record, err := s.readRecord(ref.Path)
	if err != nil || s.checkWorkflowNativeRetainedReceipt(*run, last, record) != nil {
		return nil, errorsOrConflict(err)
	}

	proof, err := s.workflowNativeDeliveryProof(*run, record)
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(proof, "", "  ")
	if err != nil || len(data) >= workflowstate.FileLimit {
		return nil, errorsOrConflict(err)
	}

	data = append(data, '\n')
	run.DeliveryProofRef, run.DeliveryProofHash = workflowDeliveryProofPath(*run), contenthash.Sum(data)
	run.DeliveryRef, run.DeliveryHash = record.Path, record.SourceHash
	entry.ReceiptRef = ref

	if err := s.projectWorkflowNativeReceiptObservation(run, entry, last); err != nil {
		return nil, err
	}

	return []workflowFileChange{{Path: run.DeliveryProofRef, After: data}}, nil
}

func (s *WorkflowStore) workflowNativeDeliveryProof(run workflowInvocationFile, record WorkflowRecord) (WorkflowSliceDeliveryProof, error) {
	if record.Receipt == nil || record.Receipt.Native == nil || !workflowNativeDeliveryReceiptMatches(run, record) {
		return WorkflowSliceDeliveryProof{}, workflowstate.ErrConflict
	}

	receipt := record.Receipt

	refs := receipt.Native.ObservationRefs
	if len(refs) == 0 {
		return WorkflowSliceDeliveryProof{}, workflowstate.ErrConflict
	}

	var observation WorkflowNativePublicationObservation
	if err := s.readWorkflowNativeEvidence(run.ID, refs[len(refs)-1], &observation); err != nil {
		return WorkflowSliceDeliveryProof{}, err
	}

	if err := validateWorkflowNativeReceiptObservation(*receipt, observation); err != nil {
		return WorkflowSliceDeliveryProof{}, err
	}

	return WorkflowSliceDeliveryProof{Schema: "chr-slice-delivery-proof/v1", InvocationID: run.ID, Unit: receipt.Unit, CandidateID: receipt.CandidateID, CandidateRef: run.CandidateRef, CandidateHash: run.CandidateHash, SubmissionID: receipt.SubmissionID, PullRequest: observation.PullRequest, Git: *observation.Git, ConfirmedAt: receipt.ConfirmedAt, NativeReceiptRef: &WorkflowEvidenceRef{Path: record.Path, Hash: record.SourceHash}}, nil
}

func workflowNativeDeliveryReceiptMatches(run workflowInvocationFile, record WorkflowRecord) bool {
	if run.Lane == nil || run.PullRequest == nil || run.Submission == nil || run.Submission.State != "confirmed" {
		return false
	}

	receipt := record.Receipt
	if record.ID != workflowSliceReceiptID(run) || record.Path != workflowSliceReceiptPath(run) || record.ProjectID != run.ProjectID || record.Namespace != run.Namespace {
		return false
	}

	return workflowNativeDeliveryOwnerMatches(run, *receipt)
}

func workflowNativeDeliveryOwnerMatches(run workflowInvocationFile, receipt WorkflowReceipt) bool {
	return receipt.Native.InvocationID == run.ID && receipt.Unit == run.Lane.Unit && receipt.CandidateID == run.PullRequest.CandidateID && receipt.SubmissionID == run.Submission.ID && receipt.PullRequest == run.PullRequest.URL
}

func (s *WorkflowStore) checkWorkflowNativeDeliveryProof(run workflowInvocationFile, proof WorkflowSliceDeliveryProof) error {
	record, err := s.readWorkflowNativeDeliveryRecord(run, *proof.NativeReceiptRef)
	if err != nil {
		return err
	}

	if err := s.checkWorkflowNativeDeliverySource(run, record); err != nil {
		return err
	}

	expected, err := s.workflowNativeDeliveryProof(run, record)
	if err != nil || !workflowNativeJSONEqual(proof, expected) {
		return errorsOrConflict(err)
	}

	return nil
}

func (s *WorkflowStore) readWorkflowNativeDeliveryRecord(run workflowInvocationFile, ref WorkflowEvidenceRef) (WorkflowRecord, error) {
	if err := validateWorkflowEvidenceRef(run.ID, ref); err != nil {
		return WorkflowRecord{}, err
	}

	record, err := s.readRecord(ref.Path)
	if err != nil || record.SourceHash != ref.Hash || record.Path != run.DeliveryRef || record.SourceHash != run.DeliveryHash {
		return record, errorsOrConflict(err)
	}

	if record.Receipt == nil || record.Receipt.Native == nil {
		return record, workflowstate.ErrConflict
	}

	return record, s.checkWorkflowNativeReceipt(record)
}

func (s *WorkflowStore) checkWorkflowNativeDeliverySource(run workflowInvocationFile, record WorkflowRecord) error {
	var candidate WorkflowIntegrationCandidate
	if err := s.readWorkflowNativeEvidence(run.ID, record.Receipt.Native.CandidateRef, &candidate); err != nil {
		return err
	}

	if candidate.Input.Source.CandidateRef != run.CandidateRef || candidate.Input.Source.CandidateHash != run.CandidateHash {
		return workflowstate.ErrConflict
	}

	return nil
}
