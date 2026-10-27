package app

import (
	"forge.adrianpk.com/chr/internal/core"
	harnessruntime "forge.adrianpk.com/chr/internal/core/harness/runtime"
)

type ticketExecutionInput struct {
	Record  core.ExecutionRecord
	Runtime harnessruntime.ValidationResult
	Events  ticketExecutionEvents
}

type ticketLoopSelection struct {
	Tickets       []core.IndexedTicket
	EligibleCount int
}

// ticketLoop retains one run's repository, event owner, progress and result.
// Context stays an explicit stage argument.
type ticketLoop struct {
	Repository core.LoopRunRepository
	Record     core.LoopRunRecord
	Events     core.AgentEventSink
	Result     ticketLoopExecutionResult
}
