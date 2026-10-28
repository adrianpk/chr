package state

// IntegrationReplay records retained intent, including empty replays.
// The source/prepared attribution lists retain both parents and author metadata.
type IntegrationReplay struct {
	Source        string `json:"source"`
	Commit        string `json:"commit"`
	EmptyDecision string `json:"empty_decision"`
}

// IntegrationFix attributes a separate prepared commit to an actionable
// finding. Candidate validation and review bind the entire resulting sequence.
type IntegrationFix struct {
	Commit  string        `json:"commit"`
	Finding ReviewFinding `json:"finding"`
}
