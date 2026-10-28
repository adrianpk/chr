package state

// PreparationQuery describes host-owned combined-result preparation.
type PreparationQuery struct {
	IntegrationQuery
	BaseCommit         string `json:"base_commit"`
	Lane               Lane   `json:"lane"`
	RequiredSignatures bool   `json:"required_signatures,omitempty"`
	SigningProfile     string `json:"signing_profile,omitempty"`
}
