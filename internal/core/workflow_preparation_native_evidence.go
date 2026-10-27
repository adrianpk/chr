package core

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *WorkflowStore) sealWorkflowNativePreparation(run workflowInvocationFile, candidate *WorkflowIntegrationCandidate) error {
	input := candidate.Input
	if input.Native == nil {
		return nil
	}

	binding := workflowNativePreparationRecord(*candidate)

	files, err := workflowNativePreparationEvidence(candidate, &binding)
	if err != nil {
		return err
	}

	candidate.Schema = workflowNativeCandidateSchema
	candidate.Native = &WorkflowNativeCandidate{WorkflowNativeRecordBinding: binding}

	candidate.Native.CandidateHash, err = WorkflowNativeCandidateHash(*candidate)
	if err != nil {
		return err
	}

	if err := validateWorkflowNativeCandidate(*candidate); err != nil {
		return err
	}

	if err := s.checkWorkflowNativePreparationFence(run, input.Native); err != nil {
		return err
	}

	// Immutable evidence can survive a crash without granting authority. The
	// candidate, attempt and queue pointers are journaled together afterward.
	return s.commitFiles(files)
}

func workflowNativePreparationRecord(candidate WorkflowIntegrationCandidate) WorkflowNativeRecordBinding {
	input, proof := candidate.Input, candidate.Input.Git
	capability := input.Native.Capability

	format := "sha1"
	if len(proof.ResultCommit) == 64 {
		format = "sha256"
	}

	return WorkflowNativeRecordBinding{
		SchemaVersion: 1, EntryID: input.Native.Entry.EntryID, InvocationID: candidate.InvocationID, Unit: input.Source.Unit,
		CanonicalPR: input.Source.PullRequest.URL, SourceRepository: proof.Repository, SourceRef: capability.SourceRef, TargetRef: capability.TargetRef,
		ObjectFormat: format, OriginalSource: proof.SourceCommit, PublishedSource: proof.ResultCommit, ValidatedBase: proof.BaseCommit,
		PreparedResult: proof.ResultCommit, PreparedTree: proof.PreparedTree, Method: proof.Method,
		ContractHash: input.Source.Slice.ContractHash, PolicyHash: capability.PolicyFingerprint, CapabilityHash: capability.Revision,
		MergerAttemptRefs: input.MergerAttemptRefs,
	}
}

func workflowNativePreparationEvidence(candidate *WorkflowIntegrationCandidate, b *WorkflowNativeRecordBinding) ([]workflowFileChange, error) {
	proof := candidate.Input.Git
	if proof.Provenance == nil || len(proof.Provenance.Source) == 0 {
		return nil, fmt.Errorf("%w: native validation requires the original source manifest", workflowstate.ErrConflict)
	}

	commits := make([]string, 0, len(proof.Provenance.Source))
	for _, source := range proof.Provenance.Source {
		commits = append(commits, source.Commit)
	}

	b.ValidationRefs = make([]WorkflowEvidenceRef, 1)
	values := []struct {
		ref   *WorkflowEvidenceRef
		value any
	}{
		{&b.SourceManifestRef, WorkflowNativeSourceManifest{"chr-native-source/v1", b.SourceRepository, b.SourceRef, b.OriginalSource, commits}},
		{&b.HistoryManifestRef, WorkflowNativeHistoryManifest{"chr-native-history/v1", proof}},
		{&b.PreparationRef, candidate.Input},
		{&b.CapabilityRef, candidate.Input.Native.Capability},
		{&b.ValidationRefs[0], candidate.Validation},
	}
	files := make([]workflowFileChange, 0, len(values))

	for _, value := range values {
		file, ref, err := workflowNativePreparationEvidenceFile(b.InvocationID, value.value)
		if err != nil {
			return nil, err
		}

		*value.ref = ref

		files = append(files, file)
	}

	return files, nil
}

func workflowNativePreparationEvidenceFile(id string, value any) (workflowFileChange, WorkflowEvidenceRef, error) {
	data, err := json.Marshal(value)
	if err != nil || len(data) >= WorkflowFileLimit {
		return workflowFileChange{}, WorkflowEvidenceRef{}, workflowstate.ErrConflict
	}

	hash := contenthash.Sum(data)
	ref := WorkflowEvidenceRef{Path: filepath.ToSlash(filepath.Join("run", id, "extension", "chr", "preparation", "evidence", hash+".json")), Hash: hash}

	return workflowFileChange{Path: ref.Path, After: data}, ref, nil
}

func (s *WorkflowStore) workflowNativePreparedEntryChange(run *workflowInvocationFile, candidate WorkflowIntegrationCandidate) (workflowFileChange, error) {
	if err := s.checkWorkflowNativePreparationFence(*run, candidate.Input.Native); err != nil {
		return workflowFileChange{}, err
	}

	entry, err := s.readWorkflowIntegrationEntry(*run, *run.Integration)
	if err != nil {
		return workflowFileChange{}, err
	}

	request := WorkflowIntegrationTransition{EntryRevision: entry.Revision, ClaimGeneration: entry.ClaimGeneration, State: "reviewing", CandidateRef: &WorkflowEvidenceRef{Path: run.Preparation.ResultRef, Hash: run.Preparation.ResultHash}}
	if entry.CandidateRef != nil {
		request.CandidateHash = entry.CandidateRef.Hash
	}

	if err := guardWorkflowIntegrationTransition(entry, request); err != nil {
		return workflowFileChange{}, err
	}

	updateWorkflowIntegrationEntry(&entry, request)

	file, err := workflowIntegrationEntryChange(entry, run.Integration.EntryRef.Hash)
	if err != nil {
		return file, err
	}

	updateWorkflowIntegrationBinding(run, WorkflowIntegrationBinding{EntryID: entry.ID, Unit: entry.Unit, Revision: entry.Revision, EntryRef: WorkflowEvidenceRef{Path: file.Path, Hash: contenthash.Sum(file.After)}})

	return file, nil
}
