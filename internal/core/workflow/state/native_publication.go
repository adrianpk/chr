package state

import (
	"errors"
	"time"
)

// NativePolicy contains current host-owned grants and certification.
// Workers cannot supply it. Ownership names one full ref, never a prefix grant.
type NativePolicy struct {
	Repository                    string     `json:"repository"`
	SourceRef                     string     `json:"source_ref"`
	TargetRef                     string     `json:"target_ref"`
	OwnershipRevision             string     `json:"ownership_revision"`
	SourceRewriteAuthorized       bool       `json:"source_rewrite_authorized"`
	FastForwardTargetAuthorized   bool       `json:"fast_forward_target_authorized"`
	HostCandidateReview           bool       `json:"host_candidate_review"`
	ReviewMode                    ReviewMode `json:"review_mode"`
	CredentialIdentityFingerprint string     `json:"credential_identity_fingerprint"`
	APIIdentityFingerprint        string     `json:"api_identity_fingerprint,omitempty"`
	ObservationContract           string     `json:"observation_contract"`
	CertifiedMethods              []string   `json:"certified_methods"`
}

// NativeGitObservation binds independently fetched refs and history to
// the same intent. It asserts no forge merge metadata or delivery receipt.
type NativeGitObservation struct {
	SchemaVersion     int               `json:"schema_version"`
	SubmissionID      string            `json:"submission_id"`
	CandidateRevision string            `json:"candidate_revision"`
	SourceCommit      string            `json:"source_commit"`
	TargetCommit      string            `json:"target_commit"`
	Condition         string            `json:"condition"`
	Git               *IntegrationProof `json:"git,omitempty"`
}

// NativePublicationBinding retains the original PR/source and exact
// prepared history independently of the rewritten published source head.
type NativePublicationBinding struct {
	ReceiptConfirmed    bool                  `json:"receipt_confirmed,omitempty"`
	GitApplication      *NativeGitObservation `json:"git_application,omitempty"`
	Capability          NativeCapability      `json:"capability"`
	PullRequest         PullRequestSnapshot   `json:"original_pull_request"`
	Git                 IntegrationProof      `json:"prepared_git"`
	Publication         GitPublicationResult  `json:"publication"`
	SourceDirectory     string                `json:"-"`
	ObservationDeadline time.Time             `json:"observation_deadline"`
}

// NativePublicationObservation distinguishes Git proof from canonical
// merged metadata. Only merged supplies both; no outcome issues a receipt.
type NativePublicationObservation struct {
	ObservationContract   string              `json:"observation_contract,omitempty"`
	CanonicalPR           string              `json:"canonical_pr,omitempty"`
	State                 string              `json:"state"`
	Condition             string              `json:"condition,omitempty"`
	PullRequest           PullRequestSnapshot `json:"pull_request"`
	Git                   *IntegrationProof   `json:"git,omitempty"`
	OriginalSourceCommit  string              `json:"original_source_commit"`
	PublishedSourceCommit string              `json:"published_source_commit"`
	ObservedSourceCommit  string              `json:"observed_source_commit,omitempty"`
	LaterSourceCommit     string              `json:"later_source_commit,omitempty"`
	ObservedTargetCommit  string              `json:"observed_target_commit,omitempty"`
}

// GitPublicationResult distinguishes transport application from delivery.
// Unknown means commands may have reached the server and must not be retried.
type GitPublicationResult struct {
	Schema                 string `json:"schema"`
	State                  string `json:"state"`
	CommandsSent           bool   `json:"commands_sent"`
	OriginalSourceCommit   string `json:"original_source_commit"`
	AdvertisedSourceCommit string `json:"advertised_source_commit"`
	AdvertisedTargetCommit string `json:"advertised_target_commit"`
	PublishedSourceCommit  string `json:"published_source_commit,omitempty"`
	ResultCommit           string `json:"result_commit"`
	CandidateRevision      string `json:"candidate_revision"`
}

// GitPublicationAdvertisement is a read-only observation of the actual
// receive-pack server. It is not a claim that an update was sent or applied.
type GitPublicationAdvertisement struct {
	SourceCommit string
	TargetCommit string
	Fingerprint  string
}

// NativeReceipt requires both transaction and original-PR observations.
type NativeReceipt struct {
	NativeRecordBinding
	CandidateHash       string        `json:"candidate_hash"`
	Mechanism           string        `json:"mechanism,omitempty"`
	CandidateRef        EvidenceRef   `json:"candidate_ref"`
	TransactionProofRef EvidenceRef   `json:"transaction_proof_ref"`
	GitApplicationRef   *EvidenceRef  `json:"git_application_ref,omitempty"`
	ObservationRefs     []EvidenceRef `json:"observation_refs"`
	ActualMergeResult   string        `json:"actual_merge_result"`
	ObservationContract string        `json:"observation_contract,omitempty"`
}

// ErrHistoryDivergence means fetched target history excludes the known
// result object. Transport errors and unavailable objects do not prove removal.
var ErrHistoryDivergence = errors.New("canonical target history excludes the integration result")
