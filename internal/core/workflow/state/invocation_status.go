package state

// InvocationStatus names coordinator progress while retaining its durable spelling.
type InvocationStatus string

// Coordinator states preserve the existing transition and optional-resume values.
const (
	InvocationUnspecified        InvocationStatus = ""
	InvocationPending            InvocationStatus = "pending"
	InvocationRunning            InvocationStatus = "running"
	InvocationWaitingCapacity    InvocationStatus = "waiting_capacity"
	InvocationWaitingReview      InvocationStatus = "waiting_review"
	InvocationWaitingIntegration InvocationStatus = "waiting_integration"
	InvocationBlocked            InvocationStatus = "blocked"
	InvocationCompleted          InvocationStatus = "completed"
	InvocationPaused             InvocationStatus = "paused"
	InvocationCancelled          InvocationStatus = "cancelled"
)
