package workflow

// IntegrationState is the retained phase of one invocation-owned integration entry.
// The underlying strings are its existing JSON and journal representation.
type IntegrationState string

// Integration phases preserve the existing queue and transition spellings.
const (
	IntegrationQueued         IntegrationState = "queued"
	IntegrationPreparing      IntegrationState = "preparing"
	IntegrationMerging        IntegrationState = "merging"
	IntegrationValidating     IntegrationState = "validating"
	IntegrationReviewing      IntegrationState = "reviewing"
	IntegrationWaitingReview  IntegrationState = "waiting_review"
	IntegrationReadyToPublish IntegrationState = "ready_to_publish"
	IntegrationPublishing     IntegrationState = "publishing"
	IntegrationObserving      IntegrationState = "observing"
	IntegrationFinalizing     IntegrationState = "finalizing"
	IntegrationIntegrated     IntegrationState = "integrated"
	IntegrationBlocked        IntegrationState = "blocked"
	IntegrationPaused         IntegrationState = "paused"
	IntegrationCancelled      IntegrationState = "cancelled"
)
