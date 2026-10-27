package execution

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// AgentEventSchemaVersion is the current durable event schema version.
const AgentEventSchemaVersion = 1

// AgentEventPhase defines a package boundary type.
type AgentEventPhase string

const (
	// AgentEventPhaseLoopStarted defines a package boundary constant.
	AgentEventPhaseLoopStarted AgentEventPhase = "loop_started"
	// AgentEventPhaseLoopFinished defines a package boundary constant.
	AgentEventPhaseLoopFinished AgentEventPhase = "loop_finished"
	// AgentEventPhaseTicketSelected defines a package boundary constant.
	AgentEventPhaseTicketSelected AgentEventPhase = "ticket_selected"
	// AgentEventPhaseCycleRehydrated defines a package boundary constant.
	AgentEventPhaseCycleRehydrated AgentEventPhase = "cycle_rehydrated"
	// AgentEventPhaseExecutionCreated defines a package boundary constant.
	AgentEventPhaseExecutionCreated AgentEventPhase = "execution_created"
	// AgentEventPhaseBranchPrepared defines a package boundary constant.
	AgentEventPhaseBranchPrepared AgentEventPhase = "branch_prepared"
	// AgentEventPhaseRuntimePrepared defines a package boundary constant.
	AgentEventPhaseRuntimePrepared AgentEventPhase = "runtime_prepared"
	// AgentEventPhaseWorkerStarted defines a package boundary constant.
	AgentEventPhaseWorkerStarted AgentEventPhase = "worker_started"
	// AgentEventPhaseWorkerFinished defines a package boundary constant.
	AgentEventPhaseWorkerFinished AgentEventPhase = "worker_finished"
	// AgentEventPhaseChangedFilesDetected defines a package boundary constant.
	AgentEventPhaseChangedFilesDetected AgentEventPhase = "changed_files_detected"
	// AgentEventPhaseValidationStarted defines a package boundary constant.
	AgentEventPhaseValidationStarted AgentEventPhase = "validation_started"
	// AgentEventPhaseValidationFinished defines a package boundary constant.
	AgentEventPhaseValidationFinished AgentEventPhase = "validation_finished"
	// AgentEventPhaseResultApplied defines a package boundary constant.
	AgentEventPhaseResultApplied AgentEventPhase = "result_applied"
	// AgentEventPhaseTicketStatusUpdated defines a package boundary constant.
	AgentEventPhaseTicketStatusUpdated AgentEventPhase = "ticket_status_updated"
	// AgentEventPhaseReviewStarted defines a package boundary constant.
	AgentEventPhaseReviewStarted AgentEventPhase = "review_started"
	// AgentEventPhaseReviewFinished defines a package boundary constant.
	AgentEventPhaseReviewFinished AgentEventPhase = "review_finished"
	// AgentEventPhaseRepairStarted defines a package boundary constant.
	AgentEventPhaseRepairStarted AgentEventPhase = "repair_started"
	// AgentEventPhaseRepairFinished defines a package boundary constant.
	AgentEventPhaseRepairFinished AgentEventPhase = "repair_finished"
	// AgentEventPhaseMergeStarted defines a package boundary constant.
	AgentEventPhaseMergeStarted AgentEventPhase = "merge_started"
	// AgentEventPhaseMergeFinished defines a package boundary constant.
	AgentEventPhaseMergeFinished AgentEventPhase = "merge_finished"
	// AgentEventPhaseWorkBlocked defines a package boundary constant.
	AgentEventPhaseWorkBlocked AgentEventPhase = "work_blocked"
	// AgentEventPhaseWorkFailed defines a package boundary constant.
	AgentEventPhaseWorkFailed AgentEventPhase = "work_failed"
	// AgentEventPhaseWorkStopped defines a package boundary constant.
	AgentEventPhaseWorkStopped AgentEventPhase = "work_stopped"
)

// AgentEventLevel defines a package boundary type.
type AgentEventLevel string

const (
	// AgentEventLevelInfo defines a package boundary constant.
	AgentEventLevelInfo AgentEventLevel = "info"
	// AgentEventLevelWarn defines a package boundary constant.
	AgentEventLevelWarn AgentEventLevel = "warn"
	// AgentEventLevelError defines a package boundary constant.
	AgentEventLevelError AgentEventLevel = "error"
)

// AgentEventRefs stores optional structured references attached to an event.
type AgentEventRefs map[string]string

// AgentEvent describes one durable agent-work lifecycle event.
type AgentEvent struct {
	SchemaVersion int             `json:"schema_version"`
	Timestamp     time.Time       `json:"timestamp"`
	ProjectID     string          `json:"project_id"`
	LoopID        string          `json:"loop_id,omitempty"`
	ExecutionID   string          `json:"execution_id,omitempty"`
	TicketKey     string          `json:"ticket_key,omitempty"`
	Phase         AgentEventPhase `json:"phase"`
	Level         AgentEventLevel `json:"level"`
	Message       string          `json:"message"`
	Refs          AgentEventRefs  `json:"refs,omitempty"`
}

// AgentEventSink receives durable agent-work events.
type AgentEventSink interface {
	EmitAgentEvent(ctx context.Context, event AgentEvent) error
}

// ValidateAgentEvent validates the corresponding package boundary value.
func ValidateAgentEvent(event AgentEvent) error {
	if event.SchemaVersion != AgentEventSchemaVersion {
		return fmt.Errorf("agent event schema version is invalid: %d", event.SchemaVersion)
	}

	if event.Timestamp.IsZero() {
		return fmt.Errorf("agent event timestamp is required")
	}

	if strings.TrimSpace(event.ProjectID) == "" {
		return fmt.Errorf("agent event project id is required")
	}

	if !event.Phase.Valid() {
		return fmt.Errorf("agent event phase is invalid: %s", event.Phase)
	}

	if !event.Level.Valid() {
		return fmt.Errorf("agent event level is invalid: %s", event.Level)
	}

	if strings.TrimSpace(event.Message) == "" {
		return fmt.Errorf("agent event message is required")
	}

	return nil
}

// Valid reports whether the value is one of the supported constants.
func (p AgentEventPhase) Valid() bool {
	switch p {
	case AgentEventPhaseLoopStarted,
		AgentEventPhaseLoopFinished,
		AgentEventPhaseTicketSelected,
		AgentEventPhaseCycleRehydrated,
		AgentEventPhaseExecutionCreated,
		AgentEventPhaseBranchPrepared,
		AgentEventPhaseRuntimePrepared,
		AgentEventPhaseWorkerStarted,
		AgentEventPhaseWorkerFinished,
		AgentEventPhaseChangedFilesDetected,
		AgentEventPhaseValidationStarted,
		AgentEventPhaseValidationFinished,
		AgentEventPhaseResultApplied,
		AgentEventPhaseTicketStatusUpdated,
		AgentEventPhaseReviewStarted,
		AgentEventPhaseReviewFinished,
		AgentEventPhaseRepairStarted,
		AgentEventPhaseRepairFinished,
		AgentEventPhaseMergeStarted,
		AgentEventPhaseMergeFinished,
		AgentEventPhaseWorkBlocked,
		AgentEventPhaseWorkFailed,
		AgentEventPhaseWorkStopped:
		return true
	default:
		return false
	}
}

// Valid reports whether the value is one of the supported constants.
func (l AgentEventLevel) Valid() bool {
	switch l {
	case AgentEventLevelInfo,
		AgentEventLevelWarn,
		AgentEventLevelError:
		return true
	default:
		return false
	}
}
