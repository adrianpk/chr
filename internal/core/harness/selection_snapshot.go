package harness

// HarnessSelectionAudit records the last accepted change without private resources.
type HarnessSelectionAudit struct {
	ProjectRef    string `json:"project_ref"`
	ChangedBy     string `json:"changed_by,omitempty"`
	ChangedAt     string `json:"changed_at,omitempty"`
	RequestID     string `json:"request_id,omitempty"`
	RequestDigest string `json:"request_digest,omitempty"`
}

// HarnessSelectionSnapshot binds both role references to one full configuration image.
type HarnessSelectionSnapshot struct {
	HarnessSelectionAudit
	Worker     string `json:"worker,omitempty"`
	Generation string `json:"generation,omitempty"`
	Revision   string `json:"revision"`
}
