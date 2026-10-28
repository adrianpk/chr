package state

// NativeReviewInput binds source rewrite, history and publication
// policy to one host-validated candidate and durable queue generation.
type NativeReviewInput struct {
	Candidate       NativeCandidate  `json:"candidate"`
	ClaimGeneration uint64           `json:"claim_generation"`
	Capability      NativeCapability `json:"capability"`
}
