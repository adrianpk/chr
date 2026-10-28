package state

// IntegrationBinding pins an invocation to one guarded queue image.
type IntegrationBinding struct {
	EntryID  string      `json:"entry_id"`
	Unit     Ref         `json:"unit"`
	Revision uint64      `json:"revision"`
	EntryRef EvidenceRef `json:"entry_ref"`
}
