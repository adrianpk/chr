package state

// SliceTaskProof binds ordered membership to host-accepted task evidence.
type SliceTaskProof struct {
	Key          string `json:"ticket_key"`
	ContractHash string `json:"contract_hash"`
	EvidenceRef  string `json:"evidence_ref"`
	EvidenceHash string `json:"evidence_hash"`
	InputRef     string `json:"input_ref"`
	InputHash    string `json:"input_hash"`
}

// SliceInput binds an integration contract and its task proofs.
// Independent ticket inputs omit set and aggregate-contract references.
type SliceInput struct {
	Unit         Ref              `json:"unit"`
	SetID        string           `json:"set_id,omitempty"`
	ContractRef  string           `json:"contract_ref,omitempty"`
	ContractHash string           `json:"contract_hash"`
	Tasks        []SliceTaskProof `json:"tasks"`
}
