package state

// PublicationCapability is a trusted provider contract for one exact PR
// and method. A source-only merge API does not satisfy AtomicSourceBase.
type PublicationCapability struct {
	Repository         string `json:"repository"`
	PullRequestURL     string `json:"pull_request_url"`
	Method             string `json:"method"`
	Mechanism          string `json:"mechanism"`
	Revision           string `json:"revision"`
	AtomicSourceBase   bool   `json:"atomic_source_base"`
	ExactResult        bool   `json:"exact_result"`
	RequiredSignatures bool   `json:"required_signatures,omitempty"`
	SigningProfile     string `json:"signing_profile,omitempty"`
}

// PreparationInput freezes source proof, current policy, protection and
// the configured method's combined result before running the host slice gate.
type PreparationInput struct {
	Source            ReviewInput             `json:"source"`
	Capability        PublicationCapability   `json:"capability"`
	Git               IntegrationProof        `json:"git"`
	Native            *NativePreparationInput `json:"native,omitempty"`
	MergerAttemptRefs []EvidenceRef           `json:"merger_attempt_refs,omitempty"`
}

// IntegrationCandidate is retained host validation of combined content.
// It is neither a provider response nor proof of delivery to the target.
type IntegrationCandidate struct {
	Native       *NativeCandidate `json:"native,omitempty"`
	Schema       string           `json:"schema"`
	InvocationID string           `json:"invocation_id"`
	Round        int              `json:"round"`
	Input        PreparationInput `json:"input"`
	Lane         Lane             `json:"lane"`
	Validation   TaskResult       `json:"validation"`
}

// PublicationGuard contains the exact protected result a publisher may
// submit after journaling intent. Providers must never substitute another method.
type PublicationGuard struct {
	PullRequest     PullRequestSnapshot   `json:"pull_request"`
	Capability      PublicationCapability `json:"capability"`
	Git             IntegrationProof      `json:"git"`
	PreparationRef  string                `json:"preparation_ref"`
	PreparationHash string                `json:"preparation_hash"`
	ReviewHash      string                `json:"review_hash"`
	Policy          ReviewPolicy          `json:"policy"`
}
