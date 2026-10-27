package app

import (
	"context"
	"fmt"
	"slices"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkQueueItem is a display observation, never an executable scheduler input.
type WorkQueueItem struct {
	Key, Title, Scope, Readiness, Reason string
}

// WorkContext preserves one invocation/attempt/ticket association and separates
// host validation from receipt-confirmed delivery.
type WorkContext struct {
	ID, Kind, InvocationID, AttemptID, TicketKey, Title, Scope, Status, Phase, Outcome, Message string
	UpdatedAt, ActivityAt                                                                       time.Time
	Eligible, Processed, Blocked, Failed                                                        int
	Validated, Confirmed, Remaining                                                             int
}

// WorkEvent has a stable identity within its original execution context.
type WorkEvent struct {
	ID, ContextID, TicketKey, Phase, Message string
	Time                                     time.Time
}

// WorkObservation is the adapter-neutral snapshot for live work inspection.
type WorkObservation struct {
	Warnings                        []string
	Authority                       core.WorkStateAuthority
	Readiness                       executioncore.WorkReadinessReport
	Queue                           []WorkQueueItem
	QueueTotal, Eligible, QueueNext int
	Current, Results                []WorkContext
	Events                          []WorkEvent
	Excluded                        map[string]int
	ObservedAt                      time.Time
}

// ObserveWork never starts, claims, recovers or stops work. Bounds apply to the
// source reads and to the queue page, current contexts and retained timeline.
func (s service) ObserveWork(ctx context.Context, offset int) (result WorkObservation, err error) {
	if offset < 0 || offset > workflowstate.RecordLimit {
		return result, fmt.Errorf("work queue offset out of bounds")
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	config, err := core.LoadWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return result, err
	}

	if err = core.ValidateWorkConfig(config.Work); err != nil {
		return result, err
	}

	result.Authority = core.WorkAuthority(config.Work)
	result.ObservedAt = s.utcNow()

	result.Readiness, err = executioncore.CheckWorkReadiness(ctx, executioncore.WorkReadinessOptions{ProjectDir: s.projectContext.ProjectDir, Policy: config.Work.BranchPolicy, Workspace: config.Work.Workspace})
	if err != nil {
		return result, err
	}

	if result.Authority == core.WorkStateAuthorityTickets {
		err = s.observeNativeWork(ctx, config, &result)
	} else {
		err = s.observeLegacyWork(ctx, config, &result)
	}

	if err != nil {
		return result, err
	}

	finishWorkObservation(&result, offset)

	return result, nil
}

func finishWorkObservation(result *WorkObservation, offset int) {
	result.Events = dedupeWorkEvents(result.Events)
	slices.Sort(result.Warnings)

	result.Warnings = slices.Compact(result.Warnings)
	result.QueueTotal = len(result.Queue)
	result.QueueNext = -1

	end := min(len(result.Queue), offset+workflowstate.PageLimit)
	if end < len(result.Queue) {
		result.QueueNext = end
	}

	result.Queue = result.Queue[min(offset, end):end]
	slices.SortStableFunc(result.Current, func(a, b WorkContext) int { return b.UpdatedAt.Compare(a.UpdatedAt) })
	slices.SortStableFunc(result.Results, func(a, b WorkContext) int { return b.UpdatedAt.Compare(a.UpdatedAt) })
	slices.SortStableFunc(result.Events, func(a, b WorkEvent) int { return b.Time.Compare(a.Time) })

	result.Current = result.Current[:min(len(result.Current), workflowstate.PageLimit)]
	result.Results = result.Results[:min(len(result.Results), 16)]
	result.Events = result.Events[:min(len(result.Events), workflowstate.PageLimit)]
}

func dedupeWorkEvents(events []WorkEvent) []WorkEvent {
	seen := map[string]bool{}

	retained := make([]WorkEvent, 0, len(events))
	for _, event := range events {
		if !seen[event.ID] {
			seen[event.ID] = true
			retained = append(retained, event)
		}
	}

	return retained
}

func bindWorkQueueReadiness(result *WorkObservation, item *WorkQueueItem) {
	if item.Readiness != "ready" {
		return
	}

	if result.Readiness.Status != executioncore.WorkReadinessStatusReady {
		item.Readiness = "blocked"
		item.Reason = "Workspace readiness is blocked."

		return
	}

	result.Eligible++
}
