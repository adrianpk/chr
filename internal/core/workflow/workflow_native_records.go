package workflow

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

const (
	workflowNativeCandidateSchema = "chr-integration-candidate/v2"
	workflowNativeReceiptSchema   = "chr-workflow-receipt/v2"
)

// NativeRecordBinding preserves S independently of the published P.
// Versioned references name host evidence, not a worker assertion or credential.
type NativeRecordBinding struct {
	SchemaVersion      int                         `json:"schema_version"`
	EntryID            string                      `json:"entry_id"`
	InvocationID       string                      `json:"invocation_id"`
	Unit               workflowstate.Ref           `json:"unit"`
	CanonicalPR        string                      `json:"canonical_pr"`
	SourceRepository   string                      `json:"source_repository"`
	SourceRef          string                      `json:"source_ref"`
	TargetRef          string                      `json:"target_ref"`
	ObjectFormat       string                      `json:"object_format"`
	OriginalSource     string                      `json:"original_source"`
	PublishedSource    string                      `json:"published_source"`
	ValidatedBase      string                      `json:"validated_base"`
	PreparedResult     string                      `json:"prepared_result"`
	PreparedTree       string                      `json:"prepared_tree"`
	Method             string                      `json:"method"`
	ContractHash       string                      `json:"contract_hash"`
	PolicyHash         string                      `json:"policy_hash"`
	CapabilityHash     string                      `json:"capability_hash"`
	SourceManifestRef  workflowstate.EvidenceRef   `json:"source_manifest_ref"`
	HistoryManifestRef workflowstate.EvidenceRef   `json:"history_manifest_ref"`
	PreparationRef     workflowstate.EvidenceRef   `json:"preparation_ref"`
	CapabilityRef      workflowstate.EvidenceRef   `json:"capability_ref"`
	ValidationRefs     []workflowstate.EvidenceRef `json:"validation_refs"`
	MergerAttemptRefs  []workflowstate.EvidenceRef `json:"merger_attempt_refs,omitempty"`
	Extensions         map[string]json.RawMessage  `json:"extensions,omitempty"`
}

// NativeCandidate adds immutable authorization to the existing candidate.
type NativeCandidate struct {
	NativeRecordBinding
	CandidateHash string `json:"candidate_hash"`
}

// NativeReceipt requires both transaction and original-PR observations.
type NativeReceipt struct {
	NativeRecordBinding
	CandidateHash       string                      `json:"candidate_hash"`
	Mechanism           string                      `json:"mechanism,omitempty"`
	CandidateRef        workflowstate.EvidenceRef   `json:"candidate_ref"`
	TransactionProofRef workflowstate.EvidenceRef   `json:"transaction_proof_ref"`
	GitApplicationRef   *workflowstate.EvidenceRef  `json:"git_application_ref,omitempty"`
	ObservationRefs     []workflowstate.EvidenceRef `json:"observation_refs"`
	ActualMergeResult   string                      `json:"actual_merge_result"`
	ObservationContract string                      `json:"observation_contract,omitempty"`
}

func validateWorkflowEvidenceRef(id string, ref workflowstate.EvidenceRef) error {
	if !workflowstate.ValidRequestKey(id) || filepath.Base(id) != id || strings.ContainsAny(ref.Path, "\\\x00") || !filepath.IsLocal(ref.Path) || filepath.ToSlash(filepath.Clean(ref.Path)) != ref.Path || !validWorkflowNativeFingerprint(ref.Hash) {
		return workflowstate.ErrConflict
	}

	root := "run/" + id + "/"
	if !strings.HasPrefix(ref.Path, root+"evidence/") && !strings.HasPrefix(ref.Path, root+"extension/chr/") {
		return fmt.Errorf("%w: evidence belongs to another invocation", workflowstate.ErrConflict)
	}

	return nil
}

func (b NativeRecordBinding) evidenceRefs() []workflowstate.EvidenceRef {
	refs := []workflowstate.EvidenceRef{b.SourceManifestRef, b.HistoryManifestRef, b.PreparationRef, b.CapabilityRef}
	refs = append(refs, b.ValidationRefs...)

	return append(refs, b.MergerAttemptRefs...)
}

func validateWorkflowNativeRecordBinding(b NativeRecordBinding) error {
	if !validWorkflowNativeRecordOwner(b) {
		return fmt.Errorf("%w: unsupported native record version or owner", workflowstate.ErrConflict)
	}

	if b.PublishedSource != b.PreparedResult || b.PreparedResult == b.OriginalSource || b.PreparedResult == b.ValidatedBase {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowNativeRecordObjects(b); err != nil {
		return err
	}

	if len(b.ValidationRefs) == 0 || len(b.evidenceRefs()) > workflowstate.PageLimit {
		return workflowstate.ErrConflict
	}

	for _, ref := range b.evidenceRefs() {
		if err := validateWorkflowEvidenceRef(b.InvocationID, ref); err != nil {
			return err
		}
	}

	return nil
}

func validateWorkflowNativeRecordObjects(b NativeRecordBinding) error {
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

func validateWorkflowNativeRecordRepository(b NativeRecordBinding) error {
	reference, err := DetectReviewProviderReference(b.CanonicalPR)
	if err != nil || "https://"+reference.Host+"/"+reference.Owner+"/"+reference.Repo != b.SourceRepository {
		return workflowstate.ErrConflict
	}

	request := GitPublication{SourceRef: b.SourceRef, TargetRef: b.TargetRef, Git: IntegrationProof{IntegrationQuery: IntegrationQuery{SourceCommit: b.OriginalSource, ResultCommit: b.PreparedResult}, BaseCommit: b.ValidatedBase}}
	if err := validateWorkflowPublicationRefs(request); err != nil {
		return err
	}

	if b.Method != "merge" && b.Method != "rebase" && b.Method != "squash" {
		return workflowstate.ErrConflict
	}

	return nil
}

// NativeCandidateHash seals every candidate input, including extensions.
func NativeCandidateHash(candidate IntegrationCandidate) (string, error) {
	if candidate.Schema != workflowNativeCandidateSchema || candidate.Native == nil {
		return "", workflowstate.ErrConflict
	}

	native := *candidate.Native
	native.CandidateHash = ""
	candidate.Native = &native

	data, err := json.Marshal(candidate)

	return contenthash.Sum(data), err
}

func validateWorkflowCandidateSchema(candidate IntegrationCandidate) error {
	switch candidate.Schema {
	case "chr-integration-candidate/v1":
		if candidate.Native != nil || candidate.Input.Capability.Mechanism == workspaceconfig.WorkflowPublicationGitAtomicDualRef {
			return workflowstate.ErrConflict
		}

		return nil
	case workflowNativeCandidateSchema:
		return validateWorkflowNativeCandidate(candidate)
	default:
		return fmt.Errorf("%w: unsupported integration candidate schema", workflowstate.ErrConflict)
	}
}

func validateWorkflowNativeCandidate(candidate IntegrationCandidate) error {
	native := candidate.Native
	if native == nil {
		return workflowstate.ErrConflict
	}

	if !validWorkflowNativeCandidateHeader(candidate) {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowNativeRecordBinding(native.NativeRecordBinding); err != nil {
		return err
	}

	if !workflowNativeCandidateOwnerMatches(candidate) {
		return workflowstate.ErrConflict
	}

	if !workflowNativeCandidateSourceMatches(candidate) {
		return workflowstate.ErrConflict
	}

	hash, err := NativeCandidateHash(candidate)
	if err != nil || hash != native.CandidateHash {
		return workflowstate.ErrConflict
	}

	return nil
}

func workflowNativeRecordGitMatches(b NativeRecordBinding, git IntegrationProof) bool {
	expected := [7]string{b.SourceRepository, b.Method, b.OriginalSource, b.ValidatedBase, b.PreparedResult, b.PreparedTree, b.TargetRef}
	actual := [7]string{git.Repository, git.Method, git.SourceCommit, git.BaseCommit, git.ResultCommit, git.PreparedTree, "refs/heads/" + git.TargetBranch}

	return expected == actual
}

func validateWorkflowReceiptSchema(receipt Receipt) error {
	if receipt.Schema == WorkflowSchema && receipt.Native == nil {
		return nil
	}

	if receipt.Schema != workflowNativeReceiptSchema || receipt.Native == nil {
		return fmt.Errorf("unsupported workflow receipt schema")
	}

	return validateWorkflowNativeReceiptSchema(receipt)
}

func validateWorkflowNativeReceiptSchema(receipt Receipt) error {
	native := receipt.Native
	if err := validateWorkflowNativeRecordBinding(native.NativeRecordBinding); err != nil {
		return err
	}

	if !workflowNativeReceiptOwnerMatches(receipt) {
		return workflowstate.ErrConflict
	}

	git := IntegrationProof{IntegrationQuery: IntegrationQuery{Repository: receipt.Repository, Method: receipt.Method, SourceCommit: receipt.SourceCommit, ResultCommit: receipt.ResultCommit, TargetBranch: strings.TrimPrefix(native.TargetRef, "refs/heads/")}, BaseCommit: receipt.BaseCommit, PreparedTree: receipt.PreparedTree}
	if !workflowNativeRecordGitMatches(native.NativeRecordBinding, git) {
		return workflowstate.ErrConflict
	}

	refs := append([]workflowstate.EvidenceRef{native.CandidateRef, native.TransactionProofRef}, native.ObservationRefs...)
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

func validWorkflowNativeRecordOwner(b NativeRecordBinding) bool {
	return b.SchemaVersion == 1 && workflowstate.ValidRequestKey(b.EntryID) && b.Unit.ID != "" && (b.Unit.Kind == "slice" || b.Unit.Kind == "ticket")
}

func validWorkflowNativeCandidateHeader(candidate IntegrationCandidate) bool {
	return candidate.Round >= 1 && validWorkflowLane(candidate.Lane) && candidate.Input.Source.InvocationID == candidate.InvocationID
}

func workflowNativeCandidateOwnerMatches(candidate IntegrationCandidate) bool {
	native := candidate.Native

	return native.InvocationID == candidate.InvocationID && native.Unit == candidate.Input.Source.Unit && native.ContractHash == candidate.Input.Source.Slice.ContractHash && workflowNativeRecordGitMatches(native.NativeRecordBinding, candidate.Input.Git)
}

func workflowNativeCandidateSourceMatches(candidate IntegrationCandidate) bool {
	native, source := candidate.Native, candidate.Input.Source
	expected := [4]string{native.CanonicalPR, native.OriginalSource, native.ValidatedBase, native.Method}
	actual := [4]string{source.PullRequest.URL, source.PullRequest.HeadCommit, source.TargetCommit, source.Policy.Method}

	return expected == actual
}

func workflowNativeReceiptOwnerMatches(receipt Receipt) bool {
	native := receipt.Native
	if native.Mechanism != "" && native.Mechanism != workspaceconfig.WorkflowPublicationGitAtomicDualRef {
		return false
	}

	return native.Unit == receipt.Unit && native.CanonicalPR == receipt.PullRequest && validWorkflowNativeFingerprint(native.CandidateHash) && workflowNativeReceiptResultMatches(receipt) && len(native.ObservationRefs) > 0 && len(native.ObservationRefs) <= workflowstate.PageLimit
}
