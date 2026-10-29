// Package operation defines transport-neutral application write outcomes.
package operation

// Status values retain the existing write contract across adapters.
const (
	OK               = "ok"
	ValidationFailed = "validation_failed"
	Stale            = "stale"
	Blocked          = "blocked"
)

// Result describes observed changes and available recovery without UI routes.
type Result struct {
	Status, Summary, BlockedReason string
	OperationID                    string
	ChangedFiles, Validation       []string
	Reload, Retry                  bool
}
