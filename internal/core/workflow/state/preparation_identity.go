package state

// IntegrationPreparation binds host identity, bounded retries and
// signature policy to the prepared objects without adding artificial commits.
type IntegrationPreparation struct {
	Identity             string                 `json:"identity"`
	InitialCommitterDate string                 `json:"initial_committer_date"`
	CommitterDate        string                 `json:"committer_date"`
	NoOpCollisions       int                    `json:"no_op_collisions"`
	RequiredSignatures   bool                   `json:"required_signatures"`
	Signing              *CommitSigningProfile  `json:"signing,omitempty"`
	Signatures           []IntegrationSignature `json:"signatures,omitempty"`
}
