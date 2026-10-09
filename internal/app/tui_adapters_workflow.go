package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/tui"
)

func (e tuiWorkActionExecutor) executeNativeWorkAction(request tui.WorkActionRequest) (tui.WorkActionResult, error) {
	copy := *request.Workflow
	request.Workflow = &copy

	if request.Workflow.Selection.Selector.Extent != "" {
		return e.nativeWorkExtentResult(request)
	}

	if request.Action == tui.WorkActionWorkflowStatus || request.Action == tui.WorkActionWorkflowPage {
		return e.nativeWorkStatusResult(request)
	}

	if request.Workflow.Preview {
		return e.nativeWorkControlPreview(request)
	}

	native := request.Workflow
	control := WorkflowControlRequest{Action: strings.TrimPrefix(string(request.Action), "workflow_"), ID: native.ID, ExpectedRevision: native.ExpectedRevision, BindingHash: native.BindingHash, StepLimit: 1}
	status, err := e.operations.ControlWorkExtent(e.operations.ctx, control)

	return nativeWorkStatusResult(status), err
}

func (e tuiWorkActionExecutor) nativeWorkExtentResult(request tui.WorkActionRequest) (tui.WorkActionResult, error) {
	native := request.Workflow
	if native.Preview {
		bound, err := e.operations.PreviewSelectedWork(e.operations.ctx, native.Selection, request.TicketKey)
		if err != nil {
			return tui.WorkActionResult{}, err
		}

		native.Selection, native.Selected, native.Preview = bound.Request, bound.Selected, false
		confirmation := tui.WorkActionState{Pending: true, Request: request, Title: "Confirm native workflow extent", Details: nativeWorkPreviewLines(bound)}

		return tui.WorkActionResult{Confirmation: &confirmation}, nil
	}

	result, err := e.operations.StartBoundWorkExtent(e.operations.ctx, native.Selection, native.Selected, 1)
	if result.Workflow == nil {
		return tui.WorkActionResult{}, err
	}

	cleanup, cancel := context.WithTimeout(context.WithoutCancel(e.operations.ctx), 10*time.Second)
	defer cancel()

	status, readErr := e.operations.WorkflowStatus(cleanup, result.Workflow.ID, WorkflowStatusRequest{Limit: 8})
	view := nativeWorkStatusResult(status)
	view.Tickets, view.Message = result.Tickets, result.Message

	return view, errors.Join(err, readErr)
}

func (e tuiWorkActionExecutor) nativeWorkStatusResult(request tui.WorkActionRequest) (tui.WorkActionResult, error) {
	if request.Workflow.ID != "" {
		status, err := e.operations.WorkflowStatus(e.operations.ctx, request.Workflow.ID, WorkflowStatusRequest{Limit: 8})

		return nativeWorkStatusResult(status), err
	}

	page, err := e.operations.WorkflowStatusPage(e.operations.ctx, request.Workflow.Offset, 16)
	view := tuiWorkflowPage(page)

	return tui.WorkActionResult{WorkflowPage: &view, Message: fmt.Sprintf("Retained workflows: %d; next offset: %d.", len(page.Items), page.NextOffset)}, err
}

func (e tuiWorkActionExecutor) nativeWorkControlPreview(request tui.WorkActionRequest) (tui.WorkActionResult, error) {
	native := request.Workflow

	status, err := e.operations.WorkflowStatus(e.operations.ctx, native.ID, WorkflowStatusRequest{Limit: 8, ExpectedRevision: native.ExpectedRevision})
	if err != nil {
		current, readErr := e.operations.WorkflowStatus(e.operations.ctx, native.ID, WorkflowStatusRequest{Limit: 8})

		return nativeWorkStatusResult(current), errors.Join(err, readErr)
	}

	if request.Action == tui.WorkActionWorkflowApprove && (native.BindingHash == "" || native.BindingHash != status.ApprovalHash) {
		return nativeWorkStatusResult(status), fmt.Errorf("approval requires the displayed exact binding")
	}

	// Copy the frozen request before changing interaction state. A later refresh
	// cannot silently replace the revision or approval hash the user confirmed.
	copy := *native
	copy.Preview = false
	request.Workflow = &copy
	confirmation := tui.WorkActionState{Pending: true, Request: request, Title: "Confirm " + strings.TrimPrefix(string(request.Action), "workflow_"), Details: tuiWorkflowView(status).Lines}

	return tui.WorkActionResult{Workflow: nativeWorkStatusResult(status).Workflow, Confirmation: &confirmation}, nil
}

func nativeWorkStatusResult(status WorkflowWorkStatus) tui.WorkActionResult {
	if status.ID == "" {
		return tui.WorkActionResult{}
	}

	view := tuiWorkflowView(status)

	return tui.WorkActionResult{Workflow: &view, Message: fmt.Sprintf("Workflow %s: %s.", status.ID, status.Status)}
}

func loadTUIWorkflowPage(ctx context.Context, project commandProjectContext) (tui.WorkflowPageView, error) {
	operations := service{ctx: ctx, projectContext: project}
	page, err := operations.WorkflowStatusPage(ctx, 0, 16)

	return tuiWorkflowPage(page), err
}

func tuiWorkflowPage(page WorkflowPage) tui.WorkflowPageView {
	view := tui.WorkflowPageView{Offset: page.Offset, Items: []tui.WorkflowView{}, NextOffset: page.NextOffset}
	for _, status := range page.Items {
		view.Items = append(view.Items, tuiWorkflowView(status))
	}

	return view
}
