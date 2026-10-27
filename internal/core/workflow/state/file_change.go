package state

// FileChange retains the original journal image, including its optimistic before hash.
// Operation owners decide validation and atomically commit complete change sets.
type FileChange struct {
	Path   string `json:"path"`
	Before string `json:"before"`
	After  []byte `json:"after"`
}
