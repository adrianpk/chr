package app

import (
	"context"
	"fmt"
	"slices"
	"time"

	"forge.adrianpk.com/chr/internal/app/work"
	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// WorkQueueItem is a display observation, never an executable scheduler input.
type WorkQueueItem = work.QueueItem

// WorkContext keeps invocation and delivery evidence separate.
type WorkContext = work.Context

// WorkEvent retains its original context identity.
type WorkEvent = work.Event

// WorkObservation is the shared adapter-neutral work snapshot.
type WorkObservation = work.Observation

// ObserveWork never starts, claims, recovers or stops work. Bounds apply to the
// source reads and to the queue page, current contexts and retained timeline.
func (s service) ObserveWork(ctx context.Context, offset int) (WorkObservation, error) {
	return s.ObserveWorkPage(ctx, work.Request{Offset: offset})
}

// ObserveWorkPage applies display filters and ordering before bounded pagination.
// Neither a filter nor a queue row grants execution authority.
func (s service) ObserveWorkPage(ctx context.Context, request work.Request) (result WorkObservation, err error) {
	offset := request.Offset

	order, err := validateWorkObservationRequest(request)
	if err != nil {
		return result, err
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	config, err := workspaceconfig.Load(s.projectContext.ProjectDir)
	if err != nil {
		return result, err
	}

	if err = workspaceconfig.ValidateWorkConfig(config.Work); err != nil {
		return result, err
	}

	result.Authority = workspaceconfig.WorkAuthority(config.Work)
	result.ObservedAt = s.utcNow()

	result.Readiness, err = executioncore.CheckWorkReadiness(ctx, executioncore.WorkReadinessOptions{ProjectDir: s.projectContext.ProjectDir, Policy: config.Work.BranchPolicy, Workspace: config.Work.Workspace})
	if err != nil {
		return result, err
	}

	if result.Authority == workspaceconfig.WorkStateAuthorityTickets {
		err = s.observeNativeWork(ctx, config, &result)
	} else {
		err = s.observeLegacyWork(ctx, config, &result)
	}

	if err != nil {
		return result, err
	}

	selectObservedQueue(&result, request.Priority, order)
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

func validateWorkObservationRequest(request work.Request) (workflowstate.TicketOrder, error) {
	if request.Offset < 0 || request.Offset > workflowstate.RecordLimit {
		return "", fmt.Errorf("work queue offset out of bounds")
	}

	order, err := ticketcore.NormalizeLoopOrder(request.Order)
	if err != nil {
		return "", err
	}

	if request.Priority != "" && !slices.Contains([]string{"critical", "high", "normal", "low"}, request.Priority) {
		return "", fmt.Errorf("invalid work queue priority")
	}

	return order, nil
}
