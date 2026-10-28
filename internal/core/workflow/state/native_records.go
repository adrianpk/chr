package state

import (
	"encoding/json"
)

// NativeRecordBinding preserves S independently of the published P.
// Versioned references name host evidence, not a worker assertion or credential.
type NativeRecordBinding struct {
	SchemaVersion      int                        `json:"schema_version"`
	EntryID            string                     `json:"entry_id"`
	InvocationID       string                     `json:"invocation_id"`
	Unit               Ref                        `json:"unit"`
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
	SourceManifestRef  EvidenceRef                `json:"source_manifest_ref"`
	HistoryManifestRef EvidenceRef                `json:"history_manifest_ref"`
	PreparationRef     EvidenceRef                `json:"preparation_ref"`
	CapabilityRef      EvidenceRef                `json:"capability_ref"`
	ValidationRefs     []EvidenceRef              `json:"validation_refs"`
	MergerAttemptRefs  []EvidenceRef              `json:"merger_attempt_refs,omitempty"`
	Extensions         map[string]json.RawMessage `json:"extensions,omitempty"`
}

// NativeCandidate adds immutable authorization to the existing candidate.
type NativeCandidate struct {
	NativeRecordBinding
	CandidateHash string `json:"candidate_hash"`
}

// EvidenceRefs returns the original ordered evidence references of this retained value.
func (b NativeRecordBinding) EvidenceRefs() []EvidenceRef {
	refs := []EvidenceRef{b.SourceManifestRef, b.HistoryManifestRef, b.PreparationRef, b.CapabilityRef}
	refs = append(refs, b.ValidationRefs...)

	return append(refs, b.MergerAttemptRefs...)
}
