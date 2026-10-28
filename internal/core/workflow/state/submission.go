package state

import (
	"time"
)

// Submission freezes publication intent before any provider mutation.
// State is awaiting_observation, unpublished or confirmed; responses alone never
// release exclusion or complete tasks.
type Submission struct {
	Native           *NativeSubmission      `json:"native,omitempty"`
	ID               string                 `json:"id"`
	InvocationID     string                 `json:"invocation_id"`
	Unit             Ref                    `json:"unit"`
	Attempt          int                    `json:"attempt"`
	CandidateID      string                 `json:"candidate_id"`
	CandidateHash    string                 `json:"candidate_hash"`
	Guard            PublicationGuard       `json:"guard"`
	PreparedLane     Lane                   `json:"prepared_lane"`
	State            string                 `json:"state"`
	SubmittedAt      time.Time              `json:"submitted_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	Response         string                 `json:"response,omitempty"`
	ResponseCommit   string                 `json:"response_commit,omitempty"`
	ObservedPR       *PullRequestSnapshot   `json:"observed_pr,omitempty"`
	Proof            *IntegrationProof      `json:"proof,omitempty"`
	ExternalDelivery bool                   `json:"external_delivery,omitempty"`
	Unpublished      *SubmissionObservation `json:"unpublished_observation,omitempty"`
}

// SubmissionResponse records an untrusted command response for later
// observation. Acceptance and a result SHA are not delivery proof.
type SubmissionResponse struct {
	Accepted     bool
	ResultCommit string
}

// SubmissionObservation is a trusted server observation of one stable
// submission ID. Unpublished guarantees that operation cannot publish later;
// an open PR, timeout, lease expiry or missing result does not establish this.
type SubmissionObservation struct {
	SubmissionID string `json:"submission_id"`
	Unpublished  bool   `json:"unpublished"`
}
