package state

// Workflow bounds apply to durable input, not just its derived index.
const (
	WorkflowRecordLimit = 4096
	WorkflowFileLimit   = 1 << 20
	WorkflowPageLimit   = 128
)
