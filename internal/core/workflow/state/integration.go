package state

// IntegrationQuery binds host Git verification to canonical provider facts.
type IntegrationQuery struct {
	Repository      string `json:"repository"`
	TargetBranch    string `json:"target_branch"`
	Method          string `json:"method"`
	SourceCommit    string `json:"source_commit"`
	SourceTree      string `json:"source_tree"`
	OriginCommit    string `json:"origin_commit"`
	ResultCommit    string `json:"result_commit"`
	SourceDirectory string `json:"-"`
}

// IntegrationCommit retains ordered source/result tree and parent proof.
type IntegrationCommit struct {
	Source string `json:"source"`
	Commit string `json:"commit"`
	Tree   string `json:"tree"`
	Parent string `json:"parent"`
}

// IntegrationProof contains host observations of an actual remote target.
type IntegrationProof struct {
	IntegrationQuery
	BaseCommit     string                 `json:"base_commit"`
	PreparedTree   string                 `json:"prepared_tree"`
	VerifiedTree   string                 `json:"verified_tree"`
	ObservedTarget string                 `json:"observed_target"`
	History        []IntegrationCommit    `json:"history"`
	HistoryHash    string                 `json:"history_hash"`
	Provenance     *IntegrationProvenance `json:"provenance,omitempty"`
}
