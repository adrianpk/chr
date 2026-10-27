package state

// Ref is a ticket-key or slice-ID reference in the containing scope.
type Ref struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

// Validation is an ordered argv gate with one shared deadline.
type Validation struct {
	Required       bool       `json:"required"`
	Commands       [][]string `json:"commands,omitempty"`
	TimeoutSeconds int        `json:"timeout_seconds"`
	WaiverReason   string     `json:"waiver_reason,omitempty"`
}
