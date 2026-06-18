package core

import "time"

// AgentEventReadbackStatus is the latest known work status inferred from events.
type AgentEventReadbackStatus string

const (
	// AgentEventReadbackStatusUnknown means no valid event has been read.
	AgentEventReadbackStatusUnknown AgentEventReadbackStatus = "unknown"
	// AgentEventReadbackStatusRunning means work has started and no terminal event is known.
	AgentEventReadbackStatusRunning AgentEventReadbackStatus = "running"
	// AgentEventReadbackStatusSucceeded means the latest event indicates successful completion.
	AgentEventReadbackStatusSucceeded AgentEventReadbackStatus = "succeeded"
	// AgentEventReadbackStatusBlocked means the latest event indicates blocked work.
	AgentEventReadbackStatusBlocked AgentEventReadbackStatus = "blocked"
	// AgentEventReadbackStatusFailed means the latest event indicates failed work.
	AgentEventReadbackStatusFailed AgentEventReadbackStatus = "failed"
	// AgentEventReadbackStatusStopped means the latest event indicates a stop request or stop result.
	AgentEventReadbackStatusStopped AgentEventReadbackStatus = "stopped"
)

// AgentEventStreamSummary is a degraded-safe latest-state summary.
type AgentEventStreamSummary struct {
	Path              string
	EventCount        int
	MalformedCount    int
	Degraded          bool
	Status            AgentEventReadbackStatus
	LatestTimestamp   time.Time
	LatestLoopID      string
	LatestTicketKey   string
	LatestExecutionID string
	LatestPhase       AgentEventPhase
	LatestLevel       AgentEventLevel
	LatestMessage     string
}

// SummarizeAgentEventStream infers the latest known state from valid events.
func SummarizeAgentEventStream(stream AgentEventStream) AgentEventStreamSummary {
	summary := AgentEventStreamSummary{
		Path:           stream.Path,
		EventCount:     len(stream.Events),
		MalformedCount: len(stream.Malformed),
		Degraded:       len(stream.Malformed) > 0,
		Status:         AgentEventReadbackStatusUnknown,
	}

	for _, event := range stream.Events {
		summary.apply(event)
	}

	return summary
}

func (s *AgentEventStreamSummary) apply(event AgentEvent) {
	if event.LoopID != "" {
		s.LatestLoopID = event.LoopID
	}

	if event.TicketKey != "" {
		s.LatestTicketKey = event.TicketKey
	}

	if event.ExecutionID != "" {
		s.LatestExecutionID = event.ExecutionID
	}

	s.LatestTimestamp = event.Timestamp
	s.LatestPhase = event.Phase
	s.LatestLevel = event.Level
	s.LatestMessage = event.Message
	s.Status = AgentEventReadbackStatusForEvent(event)
}

// AgentEventReadbackStatusForEvent maps one event to a latest-known work status.
func AgentEventReadbackStatusForEvent(event AgentEvent) AgentEventReadbackStatus {
	switch event.Phase {
	case AgentEventPhaseWorkBlocked:
		return AgentEventReadbackStatusBlocked
	case AgentEventPhaseWorkFailed:
		return AgentEventReadbackStatusFailed
	case AgentEventPhaseWorkStopped:
		return AgentEventReadbackStatusStopped
	case AgentEventPhaseLoopFinished:
		return loopFinishedReadbackStatus(event.Level)
	case AgentEventPhaseValidationFinished,
		AgentEventPhaseWorkerFinished:
		if event.Level == AgentEventLevelError {
			return AgentEventReadbackStatusFailed
		}

		return AgentEventReadbackStatusRunning
	case AgentEventPhaseResultApplied,
		AgentEventPhaseTicketStatusUpdated:
		return ticketAppliedReadbackStatus(event.Level)
	default:
		return AgentEventReadbackStatusRunning
	}
}

func loopFinishedReadbackStatus(level AgentEventLevel) AgentEventReadbackStatus {
	switch level {
	case AgentEventLevelInfo:
		return AgentEventReadbackStatusSucceeded
	case AgentEventLevelWarn:
		return AgentEventReadbackStatusBlocked
	default:
		return AgentEventReadbackStatusFailed
	}
}

func ticketAppliedReadbackStatus(level AgentEventLevel) AgentEventReadbackStatus {
	if level == AgentEventLevelWarn {
		return AgentEventReadbackStatusBlocked
	}

	if level == AgentEventLevelError {
		return AgentEventReadbackStatusFailed
	}

	return AgentEventReadbackStatusSucceeded
}
