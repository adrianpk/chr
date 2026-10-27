package state

// WorkflowClosingEvidence binds one set-owned attempt to its exact file image.
type WorkflowClosingEvidence struct {
	Ref  string `json:"ref"`
	Hash string `json:"hash"`
}

// WorkflowEvidenceRef names immutable, hashed evidence beneath one invocation.
type WorkflowEvidenceRef struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}
