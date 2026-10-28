package state

// TaskValidation binds host validation to a cumulative lane commit.
// Evidence files are immutable inputs; status text alone is never proof.
type TaskValidation struct {
	Schema       string `json:"schema"`
	Key          string `json:"ticket_key"`
	ContractHash string `json:"execution_contract_hash"`
	Branch       string `json:"branch"`
	Commit       string `json:"commit"`
	Tree         string `json:"tree"`
	Successful   bool   `json:"successful"`
}
