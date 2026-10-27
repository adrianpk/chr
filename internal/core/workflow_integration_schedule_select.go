package core

import (
	"context"
	"fmt"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (scheduler *workflowIntegrationScheduler) selectTasks(ctx context.Context, q WorkflowIntegrationQueueQuery) ([]WorkflowIntegrationSelection, map[string]int, error) {
	roles, err := scheduler.store.workflowIntegrationRoles(ctx, scheduler.snapshot)
	if err != nil {
		return nil, nil, err
	}

	priority := len(roles) > 0
	items := []WorkflowIntegrationSelection{}
	excluded := map[string]int{}

	for _, view := range scheduler.snapshot.Views {
		if !workflowIntegrationQueryMatches(view.Entry, q) {
			continue
		}

		item, reason, err := scheduler.selectTask(ctx, view.Entry)
		if err != nil {
			return nil, nil, integrationRecordError(workflowInvocationPath(view.Entry.InvocationID), err)
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

	return items, excluded, nil
}

func (scheduler *workflowIntegrationScheduler) selectTask(ctx context.Context, entry WorkflowIntegrationEntry) (WorkflowIntegrationSelection, string, error) {
	run, err := workflowIntegrationScheduleRun(scheduler.snapshot.Runs, entry)
	if err != nil {
		return WorkflowIntegrationSelection{}, "", err
	}

	owner, err := scheduler.store.workflowOwnerStore(run)
	if err != nil {
		return WorkflowIntegrationSelection{}, "", err
	}

	selection, err := scheduler.selection(ctx, owner)
	if err != nil {
		return WorkflowIntegrationSelection{}, "", err
	}

	if reason := workflowIntegrationScheduleExclusion(run, entry); reason != "" {
		return WorkflowIntegrationSelection{}, reason, nil
	}

	if err := owner.checkWorkflowInputsPrepared(ctx, run, (*WorkflowStore).requireNoWorkflowTransition); err != nil {
		return WorkflowIntegrationSelection{}, "", err
	}

	task, err := owner.workflowIntegrationScheduledTask(ctx, run, selection)
	if err != nil {
		return WorkflowIntegrationSelection{}, "", err
	}

	if task.Readiness != "ready" {
		return WorkflowIntegrationSelection{}, "task_prerequisite", nil
	}

	return WorkflowIntegrationSelection{Entry: entry, InvocationRevision: run.Revision, Task: task}, "", nil
}

func (scheduler *workflowIntegrationScheduler) selection(ctx context.Context, owner *WorkflowStore) (workflowSelection, error) {
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

func workflowIntegrationScheduleRun(runs []workflowInvocationFile, entry WorkflowIntegrationEntry) (workflowInvocationFile, error) {
	for _, run := range runs {
		if run.ID == entry.InvocationID && run.Namespace == entry.Namespace && run.ProjectID == entry.ProjectID {
			return run, nil
		}
	}

	return workflowInvocationFile{}, workflowstate.ErrConflict
}

func workflowIntegrationScheduleExclusion(run workflowInvocationFile, entry WorkflowIntegrationEntry) string {
	if entry.State != "queued" {
		return entry.State
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

func workflowIntegrationExecutorFrontier(run workflowInvocationFile, entry WorkflowIntegrationEntry) bool {
	return run.Integration != nil && run.Integration.EntryID == entry.ID && run.Cursor < len(run.Members) && workflowMemberOwnsUnit(run.Members[run.Cursor], entry.Unit)
}

func (s *WorkflowStore) workflowIntegrationScheduledTask(ctx context.Context, run workflowInvocationFile, selection workflowSelection) (WorkflowPreviewMember, error) {
	request := run.ExecutionSelection()
	request.Selector = workflowstate.Selector{Extent: "ticket", Identity: run.Members[run.Cursor].Admission.Key}
	request.Offset, request.Limit, request.SnapshotHash = 0, 1, ""

	if run.Lane != nil {
		if _, err := s.observeWorkflowLane(ctx, *run.Lane); err != nil {
			return WorkflowPreviewMember{}, err
		}

		request.CumulativeBranch, request.CumulativeCommit = run.Lane.Branch, run.Lane.HeadCommit
	}

	preview, err := selection.preview(request)
	if err != nil {
		return WorkflowPreviewMember{}, err
	}

	if err := s.observePrerequisites(ctx, selection, request, &preview); err != nil {
		return WorkflowPreviewMember{}, err
	}

	if len(preview.Members) != 1 {
		return WorkflowPreviewMember{}, fmt.Errorf("%w: admitted task is no longer selectable", workflowstate.ErrConflict)
	}

	return preview.Members[0], validateWorkflowClaimContract(selection.records[request.Selector.Identity])
}
