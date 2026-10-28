package state

// IntegrationProvenance supplements the original history proof without
// rewriting historical receipts. Source includes every original reachable commit.
type IntegrationProvenance struct {
	Schema            string                   `json:"schema"`
	ObjectFormat      string                   `json:"object_format"`
	AttributionPolicy string                   `json:"attribution_policy"`
	CombinedDiffHash  string                   `json:"combined_diff_hash"`
	Source            []IntegrationAttribution `json:"source"`
	Prepared          []IntegrationAttribution `json:"prepared"`
	Replays           []IntegrationReplay      `json:"replays,omitempty"`
	IntegrationFixes  []IntegrationFix         `json:"integration_fixes,omitempty"`
	Repairs           []IntegrationTreeRepair  `json:"repairs,omitempty"`
	Preparation       *IntegrationPreparation  `json:"preparation,omitempty"`
}
