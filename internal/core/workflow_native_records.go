package core

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

const (
	workflowNativeCandidateSchema = "chr-integration-candidate/v2"
	workflowNativeReceiptSchema   = "chr-workflow-receipt/v2"
)

// WorkflowEvidenceRef names immutable, hashed evidence beneath one invocation.
type WorkflowEvidenceRef struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

// WorkflowNativeRecordBinding preserves S independently of the published P.
// Versioned references name host evidence, not a worker assertion or credential.
type WorkflowNativeRecordBinding struct {
	SchemaVersion      int                        `json:"schema_version"`
	EntryID            string                     `json:"entry_id"`
	InvocationID       string                     `json:"invocation_id"`
	Unit               WorkflowRef                `json:"unit"`
	CanonicalPR        string                     `json:"canonical_pr"`
	SourceRepository   string                     `json:"source_repository"`
	SourceRef          string                     `json:"source_ref"`
	TargetRef          string                     `json:"target_ref"`
	ObjectFormat       string                     `json:"object_format"`
	OriginalSource     string                     `json:"original_source"`
	PublishedSource    string                     `json:"published_source"`
	ValidatedBase      string                     `json:"validated_base"`
	PreparedResult     string                     `json:"prepared_result"`
	PreparedTree       string                     `json:"prepared_tree"`
	Method             string                     `json:"method"`
	ContractHash       string                     `json:"contract_hash"`
	PolicyHash         string                     `json:"policy_hash"`
	CapabilityHash     string                     `json:"capability_hash"`
	SourceManifestRef  WorkflowEvidenceRef        `json:"source_manifest_ref"`
	HistoryManifestRef WorkflowEvidenceRef        `json:"history_manifest_ref"`
	PreparationRef     WorkflowEvidenceRef        `json:"preparation_ref"`
	CapabilityRef      WorkflowEvidenceRef        `json:"capability_ref"`
	ValidationRefs     []WorkflowEvidenceRef      `json:"validation_refs"`
	MergerAttemptRefs  []WorkflowEvidenceRef      `json:"merger_attempt_refs,omitempty"`
	Extensions         map[string]json.RawMessage `json:"extensions,omitempty"`
}

// WorkflowNativeCandidate adds immutable authorization to the existing candidate.
type WorkflowNativeCandidate struct {
	WorkflowNativeRecordBinding
	CandidateHash string `json:"candidate_hash"`
}

// WorkflowNativeReceipt requires both transaction and original-PR observations.
type WorkflowNativeReceipt struct {
	WorkflowNativeRecordBinding
	CandidateHash       string                `json:"candidate_hash"`
	Mechanism           string                `json:"mechanism,omitempty"`
	CandidateRef        WorkflowEvidenceRef   `json:"candidate_ref"`
	TransactionProofRef WorkflowEvidenceRef   `json:"transaction_proof_ref"`
	GitApplicationRef   *WorkflowEvidenceRef  `json:"git_application_ref,omitempty"`
	ObservationRefs     []WorkflowEvidenceRef `json:"observation_refs"`
	ActualMergeResult   string                `json:"actual_merge_result"`
	ObservationContract string                `json:"observation_contract,omitempty"`
}

func validateWorkflowEvidenceRef(id string, ref WorkflowEvidenceRef) error {
	if !validWorkflowRequestKey(id) || filepath.Base(id) != id || strings.ContainsAny(ref.Path, "\\\x00") || !filepath.IsLocal(ref.Path) || filepath.ToSlash(filepath.Clean(ref.Path)) != ref.Path || !validWorkflowNativeFingerprint(ref.Hash) {
		return workflowstate.ErrConflict
	}

	root := "run/" + id + "/"
	if !strings.HasPrefix(ref.Path, root+"evidence/") && !strings.HasPrefix(ref.Path, root+"extension/chr/") {
		return fmt.Errorf("%w: evidence belongs to another invocation", workflowstate.ErrConflict)
	}

	return nil
}

func (b WorkflowNativeRecordBinding) evidenceRefs() []WorkflowEvidenceRef {
	refs := []WorkflowEvidenceRef{b.SourceManifestRef, b.HistoryManifestRef, b.PreparationRef, b.CapabilityRef}
	refs = append(refs, b.ValidationRefs...)

	return append(refs, b.MergerAttemptRefs...)
}

func validateWorkflowNativeRecordBinding(b WorkflowNativeRecordBinding) error {
	if !validWorkflowNativeRecordOwner(b) {
		return fmt.Errorf("%w: unsupported native record version or owner", workflowstate.ErrConflict)
	}

	if b.PublishedSource != b.PreparedResult || b.PreparedResult == b.OriginalSource || b.PreparedResult == b.ValidatedBase {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowNativeRecordObjects(b); err != nil {
		return err
	}

	if len(b.ValidationRefs) == 0 || len(b.evidenceRefs()) > WorkflowPageLimit {
		return workflowstate.ErrConflict
	}

	for _, ref := range b.evidenceRefs() {
		if err := validateWorkflowEvidenceRef(b.InvocationID, ref); err != nil {
			return err
		}
	}

	return nil
}

func validateWorkflowNativeRecordObjects(b WorkflowNativeRecordBinding) error {
	size := 40
	if b.ObjectFormat == "sha256" {
		size = 64
	} else if b.ObjectFormat != "sha1" {
		return workflowstate.ErrConflict
	}

	for _, object := range []string{b.OriginalSource, b.ValidatedBase, b.PreparedResult, b.PreparedTree} {
		if !workflowCommitID(object) || len(object) != size {
			return workflowstate.ErrConflict
		}
	}

	for _, hash := range []string{b.ContractHash, b.PolicyHash, b.CapabilityHash} {
		if !validWorkflowNativeFingerprint(hash) {
			return workflowstate.ErrConflict
		}
	}

	return validateWorkflowNativeRecordRepository(b)
}

func validateWorkflowNativeRecordRepository(b WorkflowNativeRecordBinding) error {
	reference, err := DetectReviewProviderReference(b.CanonicalPR)
	if err != nil || "https://"+reference.Host+"/"+reference.Owner+"/"+reference.Repo != b.SourceRepository {
		return workflowstate.ErrConflict
	}

	request := WorkflowGitPublication{SourceRef: b.SourceRef, TargetRef: b.TargetRef, Git: WorkflowIntegrationProof{WorkflowIntegrationQuery: WorkflowIntegrationQuery{SourceCommit: b.OriginalSource, ResultCommit: b.PreparedResult}, BaseCommit: b.ValidatedBase}}
	if err := validateWorkflowPublicationRefs(request); err != nil {
		return err
	}

	if b.Method != "merge" && b.Method != "rebase" && b.Method != "squash" {
		return workflowstate.ErrConflict
	}

	return nil
}

// WorkflowNativeCandidateHash seals every candidate input, including extensions.
func WorkflowNativeCandidateHash(candidate WorkflowIntegrationCandidate) (string, error) {
	if candidate.Schema != workflowNativeCandidateSchema || candidate.Native == nil {
		return "", workflowstate.ErrConflict
	}

	native := *candidate.Native
	native.CandidateHash = ""
	candidate.Native = &native

	data, err := json.Marshal(candidate)

	return contenthash.Sum(data), err
}

func validateWorkflowCandidateSchema(candidate WorkflowIntegrationCandidate) error {
	switch candidate.Schema {
	case "chr-integration-candidate/v1":
		if candidate.Native != nil || candidate.Input.Capability.Mechanism == WorkflowPublicationGitAtomicDualRef {
			return workflowstate.ErrConflict
		}

		return nil
	case workflowNativeCandidateSchema:
		return validateWorkflowNativeCandidate(candidate)
	default:
		return fmt.Errorf("%w: unsupported integration candidate schema", workflowstate.ErrConflict)
	}
}

func validateWorkflowNativeCandidate(candidate WorkflowIntegrationCandidate) error {
	native := candidate.Native
	if native == nil {
		return workflowstate.ErrConflict
	}

	if !validWorkflowNativeCandidateHeader(candidate) {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowNativeRecordBinding(native.WorkflowNativeRecordBinding); err != nil {
		return err
	}

	if !workflowNativeCandidateOwnerMatches(candidate) {
		return workflowstate.ErrConflict
	}

	if !workflowNativeCandidateSourceMatches(candidate) {
		return workflowstate.ErrConflict
	}

	hash, err := WorkflowNativeCandidateHash(candidate)
	if err != nil || hash != native.CandidateHash {
		return workflowstate.ErrConflict
	}

	return nil
}

func workflowNativeRecordGitMatches(b WorkflowNativeRecordBinding, git WorkflowIntegrationProof) bool {
	expected := [7]string{b.SourceRepository, b.Method, b.OriginalSource, b.ValidatedBase, b.PreparedResult, b.PreparedTree, b.TargetRef}
	actual := [7]string{git.Repository, git.Method, git.SourceCommit, git.BaseCommit, git.ResultCommit, git.PreparedTree, "refs/heads/" + git.TargetBranch}

	return expected == actual
}

func validateWorkflowReceiptSchema(receipt WorkflowReceipt) error {
	if receipt.Schema == WorkflowSchema && receipt.Native == nil {
		return nil
	}

	if receipt.Schema != workflowNativeReceiptSchema || receipt.Native == nil {
		return fmt.Errorf("unsupported workflow receipt schema")
	}

	return validateWorkflowNativeReceiptSchema(receipt)
}

func validateWorkflowNativeReceiptSchema(receipt WorkflowReceipt) error {
	native := receipt.Native
	if err := validateWorkflowNativeRecordBinding(native.WorkflowNativeRecordBinding); err != nil {
		return err
	}

	if !workflowNativeReceiptOwnerMatches(receipt) {
		return workflowstate.ErrConflict
	}

	git := WorkflowIntegrationProof{WorkflowIntegrationQuery: WorkflowIntegrationQuery{Repository: receipt.Repository, Method: receipt.Method, SourceCommit: receipt.SourceCommit, ResultCommit: receipt.ResultCommit, TargetBranch: strings.TrimPrefix(native.TargetRef, "refs/heads/")}, BaseCommit: receipt.BaseCommit, PreparedTree: receipt.PreparedTree}
	if !workflowNativeRecordGitMatches(native.WorkflowNativeRecordBinding, git) {
		return workflowstate.ErrConflict
	}

	refs := append([]WorkflowEvidenceRef{native.CandidateRef, native.TransactionProofRef}, native.ObservationRefs...)
	if native.GitApplicationRef != nil {
		refs = append(refs, *native.GitApplicationRef)
	}

	for _, ref := range refs {
		if err := validateWorkflowEvidenceRef(native.InvocationID, ref); err != nil {
			return err
		}
	}

	return nil
}

func validWorkflowNativeRecordOwner(b WorkflowNativeRecordBinding) bool {
	return b.SchemaVersion == 1 && validWorkflowRequestKey(b.EntryID) && b.Unit.ID != "" && (b.Unit.Kind == "slice" || b.Unit.Kind == "ticket")
}

func validWorkflowNativeCandidateHeader(candidate WorkflowIntegrationCandidate) bool {
	return candidate.Round >= 1 && validWorkflowLane(candidate.Lane) && candidate.Input.Source.InvocationID == candidate.InvocationID
}

func workflowNativeCandidateOwnerMatches(candidate WorkflowIntegrationCandidate) bool {
	native := candidate.Native

	return native.InvocationID == candidate.InvocationID && native.Unit == candidate.Input.Source.Unit && native.ContractHash == candidate.Input.Source.Slice.ContractHash && workflowNativeRecordGitMatches(native.WorkflowNativeRecordBinding, candidate.Input.Git)
}

func workflowNativeCandidateSourceMatches(candidate WorkflowIntegrationCandidate) bool {
	native, source := candidate.Native, candidate.Input.Source
	expected := [4]string{native.CanonicalPR, native.OriginalSource, native.ValidatedBase, native.Method}
	actual := [4]string{source.PullRequest.URL, source.PullRequest.HeadCommit, source.TargetCommit, source.Policy.Method}

	return expected == actual
}

func workflowNativeReceiptOwnerMatches(receipt WorkflowReceipt) bool {
	native := receipt.Native
	if native.Mechanism != "" && native.Mechanism != WorkflowPublicationGitAtomicDualRef {
		return false
	}

	return native.Unit == receipt.Unit && native.CanonicalPR == receipt.PullRequest && validWorkflowNativeFingerprint(native.CandidateHash) && workflowNativeReceiptResultMatches(receipt) && len(native.ObservationRefs) > 0 && len(native.ObservationRefs) <= WorkflowPageLimit
}
