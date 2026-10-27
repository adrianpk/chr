package app

import (
	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	harnessruntime "forge.adrianpk.com/chr/internal/core/harness/runtime"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

type ticketExecutionInput struct {
	Record  executioncore.ExecutionRecord
	Runtime harnessruntime.ValidationResult
	Events  ticketExecutionEvents
}

type ticketLoopSelection struct {
	Tickets       []ticketcore.Indexed
	EligibleCount int
}

// ticketLoop retains one run's repository, event owner, progress and result.
// Context stays an explicit stage argument.
type ticketLoop struct {
	Repository executioncore.LoopRunRepository
	Record     executioncore.LoopRunRecord
	Events     executioncore.AgentEventSink
	Result     ticketLoopExecutionResult
}
