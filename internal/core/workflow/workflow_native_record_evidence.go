package workflow

import (
	"encoding/json"
	"slices"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowintegrate "forge.adrianpk.com/chr/internal/core/workflow/integrate"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// NativeSourceManifest retains the ordered original source identities.
type NativeSourceManifest struct {
	Schema         string   `json:"schema"`
	Repository     string   `json:"repository"`
	SourceRef      string   `json:"source_ref"`
	OriginalSource string   `json:"original_source"`
	Commits        []string `json:"commits"`
}

// NativeHistoryManifest versions the existing host method/history proof.
// Later provenance refinements append evidence rather than rewriting v1 records.
type NativeHistoryManifest struct {
	Schema string                         `json:"schema"`
	Git    workflowstate.IntegrationProof `json:"git"`
}

func (s *Store) readWorkflowNativeEvidence(id string, ref workflowstate.EvidenceRef, value any) error {
	if err := workflowstate.CheckEvidenceRef(id, ref); err != nil {
		return err
	}

	data, err := s.readWorkflowReviewImage(ref.Path, ref.Hash)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, value)
}

func (s *Store) checkWorkflowNativeCandidate(candidate workflowstate.IntegrationCandidate) error {
	if err := validateWorkflowCandidateSchema(candidate); err != nil || candidate.Native == nil {
		return err
	}

	b := candidate.Native.NativeRecordBinding
	if err := s.checkWorkflowNativeRecordEvidence(b); err != nil {
		return err
	}

	var input workflowstate.PreparationInput
	if err := s.readWorkflowNativeEvidence(b.InvocationID, b.PreparationRef, &input); err != nil {
		return err
	}

	if !workflowNativeJSONEqual(input, candidate.Input) || !slices.Equal(b.MergerAttemptRefs, candidate.Input.MergerAttemptRefs) {
		return workflowstate.ErrConflict
	}

	var capability workflowstate.NativeCapability
	if err := s.readWorkflowNativeEvidence(b.InvocationID, b.CapabilityRef, &capability); err != nil {
		return err
	}

	if capability.PublicationCapability != candidate.Input.Capability {
		return workflowstate.ErrConflict
	}

	return s.checkWorkflowNativeCandidateHistory(candidate, b)
}

func (s *Store) checkWorkflowNativeCandidateHistory(candidate workflowstate.IntegrationCandidate, b workflowstate.NativeRecordBinding) error {
	var history NativeHistoryManifest
	if err := s.readWorkflowNativeEvidence(b.InvocationID, b.HistoryManifestRef, &history); err != nil {
		return err
	}

	if !workflowNativeJSONEqual(history.Git, candidate.Input.Git) {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *Store) checkWorkflowNativeRecordEvidence(b workflowstate.NativeRecordBinding) error {
	if err := validateWorkflowNativeRecordBinding(b); err != nil {
		return err
	}

	for _, ref := range b.EvidenceRefs() {
		if _, err := s.readWorkflowReviewImage(ref.Path, ref.Hash); err != nil {
			return err
		}
	}

	var history NativeHistoryManifest
	if err := s.readWorkflowNativeEvidence(b.InvocationID, b.HistoryManifestRef, &history); err != nil {
		return err
	}

	if history.Schema != "chr-native-history/v1" || !workflowNativeRecordGitMatches(b, history.Git) {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowIntegrationProof(history.Git.IntegrationQuery, history.Git); err != nil {
		return err
	}

	if err := s.checkWorkflowNativeSourceManifest(b, history.Git); err != nil {
		return err
	}

	return s.checkWorkflowNativeCapabilityEvidence(b)
}

func (s *Store) checkWorkflowNativeSourceManifest(b workflowstate.NativeRecordBinding, git workflowstate.IntegrationProof) error {
	var source NativeSourceManifest
	if err := s.readWorkflowNativeEvidence(b.InvocationID, b.SourceManifestRef, &source); err != nil {
		return err
	}

	if !validWorkflowNativeSourceHeader(b, source) {
		return workflowstate.ErrConflict
	}

	if source.Commits[len(source.Commits)-1] != b.OriginalSource {
		return workflowstate.ErrConflict
	}

	return validateWorkflowNativeSourceCommits(b, source.Commits, git)
}

func (s *Store) checkWorkflowNativeCapabilityEvidence(b workflowstate.NativeRecordBinding) error {
	var capability workflowstate.NativeCapability
	if err := s.readWorkflowNativeEvidence(b.InvocationID, b.CapabilityRef, &capability); err != nil {
		return err
	}

	if err := validateWorkflowNativeCapability(capability); err != nil {
		return err
	}

	if !validWorkflowNativeCapabilityRecord(capability) {
		return workflowstate.ErrConflict
	}

	expected := [10]string{b.SourceRepository, b.CanonicalPR, b.SourceRef, b.TargetRef, b.Method, b.OriginalSource, b.ValidatedBase, b.PreparedResult, b.CapabilityHash, b.PolicyHash}

	actual := [10]string{capability.Repository, capability.PullRequestURL, capability.SourceRef, capability.TargetRef, capability.Method, capability.OriginalSourceCommit, capability.ExpectedBaseCommit, capability.PreparedResultCommit, capability.Revision, capability.PolicyFingerprint}
	if expected != actual {
		return workflowstate.ErrConflict
	}

	return s.checkWorkflowNativeValidationEvidence(b)
}

func (s *Store) checkWorkflowNativeValidationEvidence(b workflowstate.NativeRecordBinding) error {
	for _, ref := range b.ValidationRefs {
		var result workflowstate.TaskResult
		if err := s.readWorkflowNativeEvidence(b.InvocationID, ref, &result); err != nil {
			return err
		}

		if result.Schema != "chr-integration-validation/v1" || !result.Successful || result.Outcome != "integration_validated" || result.Commit != b.PreparedResult || result.Tree != b.PreparedTree || result.FinishedAt.IsZero() {
			return workflowstate.ErrConflict
		}

		if err := s.checkWorkflowNativeValidationInput(b, result); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) checkWorkflowNativeValidationInput(b workflowstate.NativeRecordBinding, result workflowstate.TaskResult) error {
	var input TaskInput
	if err := s.readWorkflowNativeEvidence(b.InvocationID, workflowstate.EvidenceRef{Path: result.InputRef, Hash: result.InputHash}, &input); err != nil {
		return err
	}

	if input.Preparation == nil || input.AttemptID != result.AttemptID || input.Generation != result.Generation || input.ContractHash != b.ContractHash || !workflowNativeRecordGitMatches(b, input.Preparation.Git) {
		return workflowstate.ErrConflict
	}

	var preparation workflowstate.PreparationInput
	if err := s.readWorkflowNativeEvidence(b.InvocationID, b.PreparationRef, &preparation); err != nil {
		return err
	}

	if !workflowNativeJSONEqual(input.Preparation, preparation) {
		return workflowstate.ErrConflict
	}

	return workflowintegrate.CheckGate(input.Contract.Validation, result)
}

func (s *Store) checkWorkflowNativeReceipt(record Record) error {
	receipt := record.Receipt
	if err := validateWorkflowReceiptSchema(*receipt); err != nil || receipt.Native == nil {
		return err
	}

	native := receipt.Native
	if !strings.HasPrefix(record.Path, "run/"+native.InvocationID+"/evidence/") || record.Revision != 1 {
		return workflowstate.ErrConflict
	}

	var candidate workflowstate.IntegrationCandidate
	if err := s.readWorkflowNativeEvidence(native.InvocationID, native.CandidateRef, &candidate); err != nil {
		return err
	}

	if err := s.checkWorkflowNativeCandidate(candidate); err != nil {
		return err
	}

	if !workflowNativeReceiptCandidateMatches(*receipt, candidate) {
		return workflowstate.ErrConflict
	}

	return s.checkWorkflowNativeReceiptObservations(*receipt)
}

func (s *Store) checkWorkflowNativeReceiptObservations(receipt Receipt) error {
	native := receipt.Native

	var transaction GitPublicationResult
	if err := s.readWorkflowNativeEvidence(native.InvocationID, native.TransactionProofRef, &transaction); err != nil {
		return err
	}

	if err := s.checkWorkflowNativeReceiptApplication(receipt, transaction); err != nil {
		return err
	}

	if err := s.checkWorkflowNativeTransactionRevision(*native, transaction); err != nil {
		return err
	}

	for i, ref := range native.ObservationRefs {
		var observation NativePublicationObservation
		if err := s.readWorkflowNativeEvidence(native.InvocationID, ref, &observation); err != nil {
			return err
		}

		if i == len(native.ObservationRefs)-1 && observation.State != "merged" {
			return workflowstate.ErrConflict
		}

		if observation.State == "merged" {
			if err := validateWorkflowNativeReceiptObservation(receipt, observation); err != nil {
				return err
			}
		}
	}

	return nil
}

func workflowNativeReceiptCandidateMatches(receipt Receipt, candidate workflowstate.IntegrationCandidate) bool {
	if candidate.Native == nil || receipt.CandidateID != candidate.Input.Source.CandidateID {
		return false
	}

	return workflowNativeJSONEqual(candidate.Native.NativeRecordBinding, receipt.Native.NativeRecordBinding) && candidate.Native.CandidateHash == receipt.Native.CandidateHash
}

func (s *Store) checkWorkflowNativeTransactionRevision(native NativeReceipt, transaction GitPublicationResult) error {
	var capability workflowstate.NativeCapability
	if err := s.readWorkflowNativeEvidence(native.InvocationID, native.CapabilityRef, &capability); err != nil {
		return err
	}

	if native.ObservationContract != "" && native.ObservationContract != capability.ObservationContract {
		return workflowstate.ErrConflict
	}

	var history NativeHistoryManifest
	if err := s.readWorkflowNativeEvidence(native.InvocationID, native.HistoryManifestRef, &history); err != nil {
		return err
	}

	request := GitPublication{Repository: native.SourceRepository, SourceRepository: native.SourceRepository, SourceRef: native.SourceRef, TargetRef: native.TargetRef, SourceRewriteAuthorized: true, OwnershipRevision: capability.OwnershipRevision, Git: history.Git}

	revision, err := GitPublicationRevision(request)
	if err != nil || revision != transaction.CandidateRevision {
		return workflowstate.ErrConflict
	}

	return nil
}

func validateWorkflowNativeReceiptObservation(receipt Receipt, observation NativePublicationObservation) error {
	native := receipt.Native
	if !workflowNativeObservationTargetValid(observation) || !validWorkflowNativeReceiptObservationHeader(receipt, observation) {
		return workflowstate.ErrConflict
	}

	git := *observation.Git
	if native.ActualMergeResult != receipt.ResultCommit && (native.ObservationContract != observation.ObservationContract || !workflowNativeSourceResultObserved(observation)) {
		return workflowstate.ErrConflict
	}

	if !workflowNativeRecordGitMatches(native.NativeRecordBinding, git) || !workflowReceiptGitMatches(receipt, git) {
		return workflowstate.ErrConflict
	}

	return validateWorkflowIntegrationProof(git.IntegrationQuery, git)
}

func workflowNativeJSONEqual(a, b any) bool {
	first, firstErr := json.Marshal(a)
	second, secondErr := json.Marshal(b)

	return firstErr == nil && secondErr == nil && contenthash.Sum(first) == contenthash.Sum(second)
}

func validWorkflowNativeSourceHeader(b workflowstate.NativeRecordBinding, source NativeSourceManifest) bool {
	expected := [4]string{"chr-native-source/v1", b.SourceRepository, b.SourceRef, b.OriginalSource}
	actual := [4]string{source.Schema, source.Repository, source.SourceRef, source.OriginalSource}

	return expected == actual && len(source.Commits) > 0 && len(source.Commits) <= workflowstate.RecordLimit
}

func validWorkflowNativeCapabilityRecord(capability workflowstate.NativeCapability) bool {
	for _, fingerprint := range []string{capability.ServerCapabilityFingerprint, capability.CredentialIdentityFingerprint} {
		if !workflowstate.ValidFingerprint(fingerprint) {
			return false
		}
	}

	return len(capability.OwnershipRevision) > 0 && len(capability.OwnershipRevision) <= 256 && !strings.ContainsAny(capability.OwnershipRevision, "\x00\r\n") && slices.Contains([]string{"forgejo_same_pr_git_auto_merge/v1", "github_same_pr_git_merge/v1", workflowNativeSourceResultContract}, capability.ObservationContract)
}

func validateWorkflowNativeSourceCommits(b workflowstate.NativeRecordBinding, commits []string, git workflowstate.IntegrationProof) error {
	seen := map[string]bool{}
	for _, commit := range commits {
		if !workflowstate.ValidCommitID(commit) || len(commit) != len(b.OriginalSource) || seen[commit] {
			return workflowstate.ErrConflict
		}

		seen[commit] = true
	}

	return validateWorkflowNativeSourceMapping(b, commits, git)
}

func validateWorkflowNativeSourceMapping(b workflowstate.NativeRecordBinding, commits []string, git workflowstate.IntegrationProof) error {
	if git.Provenance != nil {
		originals := make([]string, 0, len(git.Provenance.Source))
		for _, entry := range git.Provenance.Source {
			originals = append(originals, entry.Commit)
		}

		if !slices.Equal(commits, originals) {
			return workflowstate.ErrConflict
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
		return workflowstate.ErrConflict
	}

	return nil
}

func validWorkflowNativeReceiptTransaction(native NativeReceipt, transaction GitPublicationResult) bool {
	expected := [7]string{"chr-git-publication-result/v1", "applied", native.OriginalSource, native.OriginalSource, native.ValidatedBase, native.PreparedResult, native.PreparedResult}
	actual := [7]string{transaction.Schema, transaction.State, transaction.OriginalSourceCommit, transaction.AdvertisedSourceCommit, transaction.AdvertisedTargetCommit, transaction.PublishedSourceCommit, transaction.ResultCommit}

	return expected == actual && transaction.CommandsSent && workflowstate.ValidFingerprint(transaction.CandidateRevision)
}

func validWorkflowNativeReceiptObservationHeader(receipt Receipt, observation NativePublicationObservation) bool {
	native, pr := receipt.Native, observation.PullRequest
	if observation.CanonicalPR != "" && observation.CanonicalPR != native.CanonicalPR {
		return false
	}

	expected := [9]string{native.CanonicalPR, receipt.Repository, native.ActualMergeResult, native.OriginalSource, native.PublishedSource, "closed", native.SourceRepository, native.SourceRef, native.TargetRef}
	actual := [9]string{pr.URL, pr.Repository, pr.ResultCommit, observation.OriginalSourceCommit, observation.PublishedSourceCommit, pr.State, pr.HeadRepository, "refs/heads/" + pr.HeadBranch, "refs/heads/" + pr.BaseBranch}

	return expected == actual && pr.Merged && observation.Git != nil
}
