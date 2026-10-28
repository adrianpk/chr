package app

import (
	"context"
	"fmt"
	"path/filepath"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

func (s service) observeNativeWork(ctx context.Context, config workspaceconfig.Config, result *WorkObservation) error {
	store, err := s.workflowStore()
	if err != nil {
		return err
	}

	var target workflowstate.PreviewTarget

	identity, targetErr := s.nativeAgentWorkTarget(config)
	if targetErr == nil {
		target, targetErr = s.previewWorkTarget(config, identity)
	}
	// Queue records remain inspectable when Git readiness is blocked. Missing
	// target proof is still evaluated by the shared prerequisite owner.
	if targetErr != nil {
		target = workflowstate.PreviewTarget{}
		result.Readiness.Status = executioncore.WorkReadinessStatusBlocked
		result.Readiness.Diagnostics = append(result.Readiness.Diagnostics, executioncore.WorkReadinessDiagnostic{Code: "target_unavailable", Severity: executioncore.WorkReadinessSeverityBlocking, Detail: "Integration target proof is unavailable."})
	}

	observation, err := store.InspectWorkflowWork(ctx, target)
	if err != nil {
		return err
	}

	projectNativeWork(result, observation)

	return nil
}

func projectNativeWork(result *WorkObservation, observation workflowcore.WorkObservation) {
	tickets := map[string]ticketcore.Indexed{}
	records := map[string]workflowcore.Record{}

	for _, ticket := range observation.Tickets {
		tickets[ticket.Key] = ticket
	}

	for _, record := range observation.Records {
		if record.Task != nil {
			records[record.Task.Key] = record
		}
	}

	result.Excluded = observation.Excluded
	for _, member := range observation.Queue {
		ticket := tickets[member.Key]

		item := WorkQueueItem{Key: member.Key, Title: ticket.Title, Scope: ticket.UseCase, Readiness: member.Readiness}
		if member.Blocker != nil {
			item.Reason = member.Blocker.Reason
		}

		bindWorkQueueReadiness(result, &item)

		result.Queue = append(result.Queue, item)
	}

	for _, run := range observation.Invocations {
		appendNativeWork(result, run, tickets, records)
	}
}

func appendNativeWork(result *WorkObservation, run workflowcore.Invocation, tickets map[string]ticketcore.Indexed, records map[string]workflowcore.Record) {
	base := WorkContext{ID: run.ID, Kind: "native", InvocationID: run.ID, Status: string(run.Status), Outcome: run.Outcome, UpdatedAt: run.UpdatedAt, Validated: run.Cursor, Confirmed: run.ConfirmedCursor, Remaining: len(run.Members) - run.ConfirmedCursor}
	if len(run.Members) > 0 {
		base.TicketKey = run.Members[min(run.Cursor, len(run.Members)-1)].Admission.Key
	}

	if run.Blocker != nil {
		base.Message = run.Blocker.Reason
	}

	active := false

	for _, attempt := range run.Attempts {
		appendNativeAttemptEvent(result, run.ID, attempt)

		if attempt.Outcome != "running" || !attempt.FinishedAt.IsZero() {
			continue
		}

		current := base
		current.ID, current.AttemptID, current.TicketKey = run.ID+":"+attempt.ID, attempt.ID, attempt.TaskKey
		current.UpdatedAt = run.UpdatedAt
		bindNativeWorkContext(&current, tickets, records)
		bindNativeWorkProcess(&current, attempt)

		result.Current = append(result.Current, current)
		active = true
	}

	bindNativeWorkContext(&base, tickets, records)

	if run.Status == workflowstate.InvocationCompleted || run.Status == workflowstate.InvocationCancelled {
		result.Results = append(result.Results, base)
	} else if !active {
		result.Current = append(result.Current, base)
	}

	result.Events = append(result.Events, WorkEvent{ID: fmt.Sprintf("%s:revision:%d", run.ID, run.Revision), ContextID: run.ID, TicketKey: base.TicketKey, Time: run.UpdatedAt, Phase: string(run.Status), Message: run.Outcome})
}

func bindNativeWorkContext(current *WorkContext, tickets map[string]ticketcore.Indexed, records map[string]workflowcore.Record) {
	ticket := tickets[current.TicketKey]

	current.Title, current.Scope = ticket.Title, ticket.UseCase

	current.Phase = firstNonEmptyString(current.Outcome, current.Status)
	if current.AttemptID == "" {
		return
	}

	if record := records[current.TicketKey]; record.Task != nil {
		current.Phase = record.Task.Phase
		if current.AttemptID != "" && record.Task.AttemptID != current.AttemptID {
			current.Phase = "awaiting reconciliation"
		}
	}

	if current.Phase == "" {
		current.Phase = current.Status
	}
}

func appendNativeAttemptEvent(result *WorkObservation, id string, attempt workflowcore.Attempt) {
	at := attempt.StartedAt
	if !attempt.FinishedAt.IsZero() {
		at = attempt.FinishedAt
	}

	result.Events = append(result.Events, WorkEvent{ID: id + ":" + attempt.ID + ":" + attempt.Outcome, ContextID: id + ":" + attempt.ID, TicketKey: attempt.TaskKey, Time: at, Phase: attempt.Purpose, Message: attempt.Outcome})
}

func bindNativeWorkProcess(current *WorkContext, attempt workflowcore.Attempt) {
	if current.Message != "" {
		return
	}

	for index := len(attempt.Processes) - 1; index >= 0; index-- {
		process := attempt.Processes[index]
		if process.State == workflowcore.ProcessRunning && len(process.Command) > 0 {
			current.Message = "Running " + filepath.Base(process.Command[0])

			return
		}
	}
}
