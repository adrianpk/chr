// Package work defines application-owned, transport-neutral work observations.
package work

import (
	"time"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// Request selects a read-only queue page; eligibility remains service-owned.
type Request struct {
	Offset   int
	Priority string
	Order    workflowstate.TicketOrder
}

// QueueItem is a display observation, never an executable scheduler input.
type QueueItem struct {
	Key, Title, Scope, Readiness, Reason string
	Revision, Priority, Severity         string
	CreatedAt                            *time.Time
}

// Context preserves one invocation/attempt/ticket association and separates
// host validation from receipt-confirmed delivery.
type Context struct {
	ID, Kind, InvocationID, AttemptID, TicketKey, Title, Scope, Status, Phase, Outcome, Message string
	UpdatedAt, ActivityAt                                                                       time.Time
	Eligible, Processed, Blocked, Failed                                                        int
	Validated, Confirmed, Remaining                                                             int
}

// Event has a stable identity within its original execution context.
type Event struct {
	ID, ContextID, TicketKey, Phase, Message string
	Time                                     time.Time
}

// Observation is the adapter-neutral snapshot for live work inspection.
type Observation struct {
	Warnings                        []string
	Authority                       workspaceconfig.WorkStateAuthority
	Readiness                       executioncore.WorkReadinessReport
	Queue                           []QueueItem
	QueueTotal, Eligible, QueueNext int
	Current, Results                []Context
	Events                          []Event
	Excluded                        map[string]int
	ObservedAt                      time.Time
}
