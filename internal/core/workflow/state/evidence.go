package state

// ClosingEvidence binds one set-owned attempt to its exact file image.
type ClosingEvidence struct {
	Ref  string `json:"ref"`
	Hash string `json:"hash"`
}

// EvidenceRef names immutable, hashed evidence beneath one invocation.
type EvidenceRef struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}
