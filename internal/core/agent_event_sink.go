package core

import (
	"context"
	"errors"
)

// AgentEventFanoutSink forwards one event to multiple sinks.
type AgentEventFanoutSink struct {
	Sinks []AgentEventSink
}

// NewAgentEventFanoutSink constructs the corresponding package service or value.
func NewAgentEventFanoutSink(sinks ...AgentEventSink) AgentEventFanoutSink {
	filtered := make([]AgentEventSink, 0, len(sinks))
	for _, sink := range sinks {
		if sink != nil {
			filtered = append(filtered, sink)
		}
	}

	return AgentEventFanoutSink{Sinks: filtered}
}

// EmitAgentEvent forwards the requested event to every configured sink.
func (s AgentEventFanoutSink) EmitAgentEvent(ctx context.Context, event AgentEvent) error {
	var emitErr error

	for _, sink := range s.Sinks {
		err := sink.EmitAgentEvent(ctx, event)
		if err != nil {
			emitErr = errors.Join(emitErr, err)
		}
	}

	return emitErr
}
