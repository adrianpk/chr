package core

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// WorkflowNativeSourceManifest retains the ordered original source identities.
type WorkflowNativeSourceManifest struct {
	Schema         string   `json:"schema"`
	Repository     string   `json:"repository"`
	SourceRef      string   `json:"source_ref"`
	OriginalSource string   `json:"original_source"`
	Commits        []string `json:"commits"`
}

// WorkflowNativeHistoryManifest versions the existing host method/history proof.
// Later provenance refinements append evidence rather than rewriting v1 records.
type WorkflowNativeHistoryManifest struct {
	Schema string                   `json:"schema"`
	Git    WorkflowIntegrationProof `json:"git"`
}

func (s *WorkflowStore) readWorkflowNativeEvidence(id string, ref WorkflowEvidenceRef, value any) error {
	if err := validateWorkflowEvidenceRef(id, ref); err != nil {
		return err
	}

	data, err := s.readWorkflowReviewImage(ref.Path, ref.Hash)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, value)
}

func (s *WorkflowStore) checkWorkflowNativeCandidate(candidate WorkflowIntegrationCandidate) error {
	if err := validateWorkflowCandidateSchema(candidate); err != nil || candidate.Native == nil {
		return err
	}

	b := candidate.Native.WorkflowNativeRecordBinding
	if err := s.checkWorkflowNativeRecordEvidence(b); err != nil {
		return err
	}

	var input WorkflowPreparationInput
	if err := s.readWorkflowNativeEvidence(b.InvocationID, b.PreparationRef, &input); err != nil {
		return err
	}

	if !workflowNativeJSONEqual(input, candidate.Input) || !slices.Equal(b.MergerAttemptRefs, candidate.Input.MergerAttemptRefs) {
		return ErrWorkflowConflict
	}

	var capability WorkflowNativeCapability
	if err := s.readWorkflowNativeEvidence(b.InvocationID, b.CapabilityRef, &capability); err != nil {
		return err
	}

	if capability.WorkflowPublicationCapability != candidate.Input.Capability {
		return ErrWorkflowConflict
	}

	return s.checkWorkflowNativeCandidateHistory(candidate, b)
}

func (s *WorkflowStore) checkWorkflowNativeCandidateHistory(candidate WorkflowIntegrationCandidate, b WorkflowNativeRecordBinding) error {
	var history WorkflowNativeHistoryManifest
	if err := s.readWorkflowNativeEvidence(b.InvocationID, b.HistoryManifestRef, &history); err != nil {
		return err
	}

	if !workflowNativeJSONEqual(history.Git, candidate.Input.Git) {
		return ErrWorkflowConflict
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowNativeRecordEvidence(b WorkflowNativeRecordBinding) error {
	if err := validateWorkflowNativeRecordBinding(b); err != nil {
		return err
	}

	for _, ref := range b.evidenceRefs() {
		if _, err := s.readWorkflowReviewImage(ref.Path, ref.Hash); err != nil {
			return err
		}
	}

	var history WorkflowNativeHistoryManifest
	if err := s.readWorkflowNativeEvidence(b.InvocationID, b.HistoryManifestRef, &history); err != nil {
		return err
	}

	if history.Schema != "chr-native-history/v1" || !workflowNativeRecordGitMatches(b, history.Git) {
		return ErrWorkflowConflict
	}

	if err := validateWorkflowIntegrationProof(history.Git.WorkflowIntegrationQuery, history.Git); err != nil {
		return err
	}

	if err := s.checkWorkflowNativeSourceManifest(b, history.Git); err != nil {
		return err
	}

	return s.checkWorkflowNativeCapabilityEvidence(b)
}

func (s *WorkflowStore) checkWorkflowNativeSourceManifest(b WorkflowNativeRecordBinding, git WorkflowIntegrationProof) error {
	var source WorkflowNativeSourceManifest
	if err := s.readWorkflowNativeEvidence(b.InvocationID, b.SourceManifestRef, &source); err != nil {
		return err
	}

	if !validWorkflowNativeSourceHeader(b, source) {
		return ErrWorkflowConflict
	}

	if source.Commits[len(source.Commits)-1] != b.OriginalSource {
		return ErrWorkflowConflict
	}

	return validateWorkflowNativeSourceCommits(b, source.Commits, git)
}

func (s *WorkflowStore) checkWorkflowNativeCapabilityEvidence(b WorkflowNativeRecordBinding) error {
	var capability WorkflowNativeCapability
	if err := s.readWorkflowNativeEvidence(b.InvocationID, b.CapabilityRef, &capability); err != nil {
		return err
	}

	if err := validateWorkflowNativeCapability(capability); err != nil {
		return err
	}

	if !validWorkflowNativeCapabilityRecord(capability) {
		return ErrWorkflowConflict
	}

	expected := [10]string{b.SourceRepository, b.CanonicalPR, b.SourceRef, b.TargetRef, b.Method, b.OriginalSource, b.ValidatedBase, b.PreparedResult, b.CapabilityHash, b.PolicyHash}

	actual := [10]string{capability.Repository, capability.PullRequestURL, capability.SourceRef, capability.TargetRef, capability.Method, capability.OriginalSourceCommit, capability.ExpectedBaseCommit, capability.PreparedResultCommit, capability.Revision, capability.PolicyFingerprint}
	if expected != actual {
		return ErrWorkflowConflict
	}

	return s.checkWorkflowNativeValidationEvidence(b)
}

func (s *WorkflowStore) checkWorkflowNativeValidationEvidence(b WorkflowNativeRecordBinding) error {
	for _, ref := range b.ValidationRefs {
		var result WorkflowTaskResult
		if err := s.readWorkflowNativeEvidence(b.InvocationID, ref, &result); err != nil {
			return err
		}

		if result.Schema != "chr-integration-validation/v1" || !result.Successful || result.Outcome != "integration_validated" || result.Commit != b.PreparedResult || result.Tree != b.PreparedTree || result.FinishedAt.IsZero() {
			return ErrWorkflowConflict
		}

		if err := s.checkWorkflowNativeValidationInput(b, result); err != nil {
			return err
		}
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowNativeValidationInput(b WorkflowNativeRecordBinding, result WorkflowTaskResult) error {
	var input WorkflowTaskInput
	if err := s.readWorkflowNativeEvidence(b.InvocationID, WorkflowEvidenceRef{Path: result.InputRef, Hash: result.InputHash}, &input); err != nil {
		return err
	}

	if input.Preparation == nil || input.AttemptID != result.AttemptID || input.Generation != result.Generation || input.ContractHash != b.ContractHash || !workflowNativeRecordGitMatches(b, input.Preparation.Git) {
		return ErrWorkflowConflict
	}

	var preparation WorkflowPreparationInput
	if err := s.readWorkflowNativeEvidence(b.InvocationID, b.PreparationRef, &preparation); err != nil {
		return err
	}

	if !workflowNativeJSONEqual(input.Preparation, preparation) {
		return ErrWorkflowConflict
	}

	return validateWorkflowPreparationGate(input, result)
}

func (s *WorkflowStore) checkWorkflowNativeReceipt(record WorkflowRecord) error {
	receipt := record.Receipt
	if err := validateWorkflowReceiptSchema(*receipt); err != nil || receipt.Native == nil {
		return err
	}

	native := receipt.Native
	if !strings.HasPrefix(record.Path, "run/"+native.InvocationID+"/evidence/") || record.Revision != 1 {
		return ErrWorkflowConflict
	}

	var candidate WorkflowIntegrationCandidate
	if err := s.readWorkflowNativeEvidence(native.InvocationID, native.CandidateRef, &candidate); err != nil {
		return err
	}

	if err := s.checkWorkflowNativeCandidate(candidate); err != nil {
		return err
	}

	if candidate.Native == nil {
		return ErrWorkflowConflict
	}

	if !workflowNativeJSONEqual(candidate.Native.WorkflowNativeRecordBinding, native.WorkflowNativeRecordBinding) || candidate.Native.CandidateHash != native.CandidateHash {
		return ErrWorkflowConflict
	}

	return s.checkWorkflowNativeReceiptObservations(*receipt)
}

func (s *WorkflowStore) checkWorkflowNativeReceiptObservations(receipt WorkflowReceipt) error {
	native := receipt.Native

	var transaction WorkflowGitPublicationResult
	if err := s.readWorkflowNativeEvidence(native.InvocationID, native.TransactionProofRef, &transaction); err != nil {
		return err
	}

	if !validWorkflowNativeReceiptTransaction(*native, transaction) {
		return ErrWorkflowConflict
	}

	if err := s.checkWorkflowNativeTransactionRevision(*native, transaction); err != nil {
		return err
	}

	verified := false

	for _, ref := range native.ObservationRefs {
		var observation WorkflowNativePublicationObservation
		if err := s.readWorkflowNativeEvidence(native.InvocationID, ref, &observation); err != nil {
			return err
		}

		if observation.State == "merged" {
			if err := validateWorkflowNativeReceiptObservation(receipt, observation); err != nil {
				return err
			}

			verified = true
		}
	}

	if !verified {
		return fmt.Errorf("%w: receipt requires canonical merged PR proof", ErrWorkflowConflict)
	}

	return nil
}

func (s *WorkflowStore) checkWorkflowNativeTransactionRevision(native WorkflowNativeReceipt, transaction WorkflowGitPublicationResult) error {
	var capability WorkflowNativeCapability
	if err := s.readWorkflowNativeEvidence(native.InvocationID, native.CapabilityRef, &capability); err != nil {
		return err
	}

	var history WorkflowNativeHistoryManifest
	if err := s.readWorkflowNativeEvidence(native.InvocationID, native.HistoryManifestRef, &history); err != nil {
		return err
	}

	request := WorkflowGitPublication{Repository: native.SourceRepository, SourceRepository: native.SourceRepository, SourceRef: native.SourceRef, TargetRef: native.TargetRef, SourceRewriteAuthorized: true, OwnershipRevision: capability.OwnershipRevision, Git: history.Git}

	revision, err := WorkflowGitPublicationRevision(request)
	if err != nil || revision != transaction.CandidateRevision {
		return ErrWorkflowConflict
	}

	return nil
}

func validateWorkflowNativeReceiptObservation(receipt WorkflowReceipt, observation WorkflowNativePublicationObservation) error {
	native := receipt.Native
	if !validWorkflowNativeReceiptObservationHeader(receipt, observation) {
		return ErrWorkflowConflict
	}

	git := *observation.Git
	if !workflowNativeRecordGitMatches(native.WorkflowNativeRecordBinding, git) || !workflowReceiptGitMatches(receipt, git) {
		return ErrWorkflowConflict
	}

	return validateWorkflowIntegrationProof(git.WorkflowIntegrationQuery, git)
}

func workflowNativeJSONEqual(a, b any) bool {
	first, firstErr := json.Marshal(a)
	second, secondErr := json.Marshal(b)

	return firstErr == nil && secondErr == nil && workflowHash(first) == workflowHash(second)
}

func validWorkflowNativeSourceHeader(b WorkflowNativeRecordBinding, source WorkflowNativeSourceManifest) bool {
	expected := [4]string{"chr-native-source/v1", b.SourceRepository, b.SourceRef, b.OriginalSource}
	actual := [4]string{source.Schema, source.Repository, source.SourceRef, source.OriginalSource}

	return expected == actual && len(source.Commits) > 0 && len(source.Commits) <= WorkflowRecordLimit
}

func validWorkflowNativeCapabilityRecord(capability WorkflowNativeCapability) bool {
	for _, fingerprint := range []string{capability.ServerCapabilityFingerprint, capability.CredentialIdentityFingerprint} {
		if !validWorkflowNativeFingerprint(fingerprint) {
			return false
		}
	}

	return len(capability.OwnershipRevision) > 0 && len(capability.OwnershipRevision) <= 256 && !strings.ContainsAny(capability.OwnershipRevision, "\x00\r\n") && slices.Contains([]string{"forgejo_same_pr_git_auto_merge/v1", "github_same_pr_git_merge/v1"}, capability.ObservationContract)
}

func validateWorkflowNativeSourceCommits(b WorkflowNativeRecordBinding, commits []string, git WorkflowIntegrationProof) error {
	seen := map[string]bool{}
	for _, commit := range commits {
		if !workflowCommitID(commit) || len(commit) != len(b.OriginalSource) || seen[commit] {
			return ErrWorkflowConflict
		}

		seen[commit] = true
	}

	return validateWorkflowNativeSourceMapping(b, commits, git)
}

func validateWorkflowNativeSourceMapping(b WorkflowNativeRecordBinding, commits []string, git WorkflowIntegrationProof) error {
	if git.Provenance != nil {
		originals := make([]string, 0, len(git.Provenance.Source))
		for _, entry := range git.Provenance.Source {
			originals = append(originals, entry.Commit)
		}

		if !slices.Equal(commits, originals) {
			return ErrWorkflowConflict
		}
	}

	if b.Method != "rebase" {
		return nil
	}

	originals := make([]string, 0, len(git.History))
	for _, step := range git.History {
		originals = append(originals, step.Source)
	}

	if !slices.Equal(commits, originals) {
		return ErrWorkflowConflict
	}

	return nil
}

func validWorkflowNativeReceiptTransaction(native WorkflowNativeReceipt, transaction WorkflowGitPublicationResult) bool {
	expected := [7]string{"chr-git-publication-result/v1", "applied", native.OriginalSource, native.OriginalSource, native.ValidatedBase, native.PreparedResult, native.PreparedResult}
	actual := [7]string{transaction.Schema, transaction.State, transaction.OriginalSourceCommit, transaction.AdvertisedSourceCommit, transaction.AdvertisedTargetCommit, transaction.PublishedSourceCommit, transaction.ResultCommit}

	return expected == actual && transaction.CommandsSent && validWorkflowNativeFingerprint(transaction.CandidateRevision)
}

func validWorkflowNativeReceiptObservationHeader(receipt WorkflowReceipt, observation WorkflowNativePublicationObservation) bool {
	native, pr := receipt.Native, observation.PullRequest
	expected := [9]string{native.CanonicalPR, receipt.Repository, native.ActualMergeResult, native.OriginalSource, native.PublishedSource, "closed", native.SourceRepository, native.SourceRef, native.TargetRef}
	actual := [9]string{pr.URL, pr.Repository, pr.ResultCommit, observation.OriginalSourceCommit, observation.PublishedSourceCommit, pr.State, pr.HeadRepository, "refs/heads/" + pr.HeadBranch, "refs/heads/" + pr.BaseBranch}

	return expected == actual && pr.Merged && observation.Git != nil
}
