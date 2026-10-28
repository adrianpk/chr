package state

import (
	"time"
)

// NativeSubmission extends the common journal with the exact Git intent.
// Git application alone retains awaiting_observation and cannot issue a receipt.
type NativeSubmission struct {
	NativeRecordBinding
	CandidateHash            string                 `json:"candidate_hash"`
	CandidateRef             EvidenceRef            `json:"candidate_ref"`
	ClaimGeneration          uint64                 `json:"claim_generation"`
	Mechanism                string                 `json:"mechanism"`
	ReviewGrantRef           EvidenceRef            `json:"review_grant_ref"`
	FastForwardProofRef      EvidenceRef            `json:"fast_forward_proof_ref"`
	IntentTime               time.Time              `json:"intent_time"`
	DispatchState            NativeDispatchState    `json:"dispatch_state"`
	DispatchIdentity         *ProcessIdentity       `json:"dispatch_identity"`
	DispatchCompletionRef    *EvidenceRef           `json:"dispatch_completion_ref,omitempty"`
	AdvertisementRef         *EvidenceRef           `json:"advertisement_observation_ref"`
	CommandManifestRef       *EvidenceRef           `json:"command_manifest_ref"`
	TransportResponseRef     *EvidenceRef           `json:"transport_response_ref"`
	TransportResponseHistory []EvidenceRef          `json:"transport_response_history,omitempty"`
	ApplicationState         NativeApplicationState `json:"application_state"`
	ObservationDeadline      time.Time              `json:"observation_deadline"`
	ObservationRefs          []EvidenceRef          `json:"observation_refs"`
	GitObservationRefs       []EvidenceRef          `json:"git_observation_refs,omitempty"`
	ReceiptRef               *EvidenceRef           `json:"receipt_ref"`
	NonApplicationRef        *EvidenceRef           `json:"non_application_ref"`
}
