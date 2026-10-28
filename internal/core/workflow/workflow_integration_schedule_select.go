package workflow

import (
	"context"
	"fmt"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (scheduler *workflowIntegrationScheduler) selectTasks(ctx context.Context, q IntegrationQueueQuery) (workflowTaskSelections, error) {
	roles, err := (workflowIntegrationRoleSelector{store: scheduler.store}).roles(ctx, scheduler.snapshot)
	if err != nil {
		return workflowTaskSelections{Items: nil, Excluded: nil}, err
	}

	priority := len(roles) > 0
	items := []IntegrationSelection{}
	excluded := map[string]int{}

	for _, view := range scheduler.snapshot.Views {
		if !workflowIntegrationQueryMatches(view.Entry, q) {
			continue
		}

		taskSelection, err := scheduler.selectTask(ctx, view.Entry)

		item, reason := taskSelection.Item, taskSelection.Reason

		if err != nil {
			return workflowTaskSelections{Items: nil, Excluded: nil}, integrationRecordError(workflowInvocationPath(view.Entry.InvocationID), err)
		}

		if reason != "" {
			excluded[reason]++

			continue
		}

		if priority {
			excluded["integration_priority"]++

			continue
		}

		items = append(items, item)
	}

	return workflowTaskSelections{Items: items, Excluded: excluded}, nil
}

func (scheduler *workflowIntegrationScheduler) selectTask(ctx context.Context, entry IntegrationEntry) (workflowTaskSelection, error) {
	run, err := workflowIntegrationScheduleRun(scheduler.snapshot.Runs, entry)
	if err != nil {
		return workflowTaskSelection{Item: IntegrationSelection{}, Reason: ""}, err
	}

	owner, err := scheduler.store.workflowOwnerStore(run)
	if err != nil {
		return workflowTaskSelection{Item: IntegrationSelection{}, Reason: ""}, err
	}

	selection, err := scheduler.selection(ctx, owner)
	if err != nil {
		return workflowTaskSelection{Item: IntegrationSelection{}, Reason: ""}, err
	}

	if reason := workflowIntegrationScheduleExclusion(run, entry); reason != "" {
		return workflowTaskSelection{Item: IntegrationSelection{}, Reason: reason}, nil
	}

	if err := owner.checkWorkflowInputsPrepared(ctx, run, (*Store).requireNoWorkflowTransition); err != nil {
		return workflowTaskSelection{Item: IntegrationSelection{}, Reason: ""}, err
	}

	task, err := owner.workflowIntegrationScheduledTask(ctx, run, selection)
	if err != nil {
		return workflowTaskSelection{Item: IntegrationSelection{}, Reason: ""}, err
	}

	if task.Readiness != "ready" {
		return workflowTaskSelection{Item: IntegrationSelection{}, Reason: "task_prerequisite"}, nil
	}

	return workflowTaskSelection{Item: IntegrationSelection{Entry: entry, InvocationRevision: run.Revision, Task: task}, Reason: ""}, nil
}

func (scheduler *workflowIntegrationScheduler) selection(ctx context.Context, owner *Store) (workflowSelection, error) {
	key := owner.layout.Root()

	selection, found := scheduler.selections[key]
	if found {
		return selection, nil
	}

	selection, err := owner.readWorkflowSelection(ctx)
	if err == nil {
		scheduler.selections[key] = selection
	}

	return selection, err
}

func workflowIntegrationScheduleRun(runs []workflowInvocationFile, entry IntegrationEntry) (workflowInvocationFile, error) {
	for _, run := range runs {
		if run.ID == entry.InvocationID && run.Namespace == entry.Namespace && run.ProjectID == entry.ProjectID {
			return run, nil
		}
	}

	return workflowInvocationFile{}, workflowstate.ErrConflict
}

func workflowIntegrationScheduleExclusion(run workflowInvocationFile, entry IntegrationEntry) string {
	if entry.State != "queued" {
		return string(entry.State)
	}

	if entry.StopIntent != "" || run.StopIntent != "" {
		return "stop_intent"
	}

	restoreWorkflowCapacityWait(&run)

	if !workflowInvocationClaimable(run) {
		return "invocation_wait"
	}

	if !workflowIntegrationExecutorFrontier(run, entry) {
		return "receipt_barrier"
	}

	return ""
}

func workflowIntegrationExecutorFrontier(run workflowInvocationFile, entry IntegrationEntry) bool {
	return run.Integration != nil && run.Integration.EntryID == entry.ID && run.Cursor < len(run.Members) && workflowMemberOwnsUnit(run.Members[run.Cursor], entry.Unit)
}

func (s *Store) workflowIntegrationScheduledTask(ctx context.Context, run workflowInvocationFile, selection workflowSelection) (workflowstate.PreviewMember, error) {
	request := run.ExecutionSelection()
	request.Selector = workflowstate.Selector{Extent: "ticket", Identity: run.Members[run.Cursor].Admission.Key}
	request.Offset, request.Limit, request.SnapshotHash = 0, 1, ""

	if run.Lane != nil {
		if _, err := s.observeWorkflowLane(ctx, *run.Lane); err != nil {
			return workflowstate.PreviewMember{}, err
		}

		request.CumulativeBranch, request.CumulativeCommit = run.Lane.Branch, run.Lane.HeadCommit
	}

	preview, err := selection.preview(request)
	if err != nil {
		return workflowstate.PreviewMember{}, err
	}

	if err := s.observePrerequisites(ctx, selection, request, &preview); err != nil {
		return workflowstate.PreviewMember{}, err
	}

	if len(preview.Members) != 1 {
		return workflowstate.PreviewMember{}, fmt.Errorf("%w: admitted task is no longer selectable", workflowstate.ErrConflict)
	}

	return preview.Members[0], validateWorkflowClaimContract(selection.records[request.Selector.Identity])
}
