package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
	"forge.adrianpk.com/chr/internal/webui"
)

type webWorkflowBinding struct {
	Request  workflowstate.PreviewRequest
	Selected *workflowstate.SelectionGuard
}

func (p webViewProvider) workflowEnabled() bool {
	config, err := workspaceconfig.Load(p.projectContext.ProjectDir)

	return err == nil && workspaceconfig.WorkAuthority(config.Work) == workspaceconfig.WorkStateAuthorityTickets
}

// WorkflowView reads the same native preview and status operations as the CLI.
func (p webViewProvider) WorkflowView(ctx context.Context, request webui.WorkflowViewRequest) (webui.WorkflowView, error) {
	view := webui.WorkflowView{Enabled: p.workflowEnabled()}
	if !view.Enabled && request.Extent != "" {
		view.Failure = "Native workflow controls require tickets authority."

		return view, nil
	}

	operations := p.service(ctx)
	if request.Extent != "" {
		return webWorkflowPreview(ctx, operations, request)
	}

	if request.InvocationID != "" {
		status, err := operations.WorkflowStatus(ctx, request.InvocationID, WorkflowStatusRequest{Offset: request.Offset, Limit: 8, ExpectedRevision: request.Revision})
		if err != nil {
			current, readErr := operations.WorkflowStatus(ctx, request.InvocationID, WorkflowStatusRequest{Limit: 8})
			view.Current = webWorkflowStatusPointer(current)

			return view, errors.Join(err, readErr)
		}

		view.Current = webWorkflowStatusPointer(status)

		return view, nil
	}

	page, err := operations.WorkflowStatusPage(ctx, request.Offset, 16)
	for _, status := range page.Items {
		view.Items = append(view.Items, webWorkflowRun(status))
	}

	if page.NextOffset >= 0 {
		view.NextURL = "/workflows?offset=" + strconv.Itoa(page.NextOffset)
	}

	return view, err
}

func webWorkflowPreview(ctx context.Context, operations service, request webui.WorkflowViewRequest) (webui.WorkflowView, error) {
	if request.Offset != 0 || request.Revision != 0 {
		return webui.WorkflowView{Enabled: true}, fmt.Errorf("admission preview requires offset zero")
	}

	selection := workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: request.Extent, Identity: request.Identity}, Filters: workflowstate.TicketFilters{ProjectAreas: webWorkflowFilter(request.ProjectArea), Severities: webWorkflowFilter(request.Severity), Priorities: webWorkflowFilter(request.Priority)}, Order: workflowstate.TicketOrder(request.Order)}

	bound, err := operations.PreviewSelectedWork(ctx, selection, request.TicketKey)
	if err != nil {
		return webui.WorkflowView{Enabled: true}, err
	}

	binding, err := json.Marshal(webWorkflowBinding{Request: bound.Request, Selected: bound.Selected})
	if err != nil {
		return webui.WorkflowView{Enabled: true}, err
	}

	preview := webWorkflowPreviewView(bound)
	preview.Binding = string(binding)

	return webui.WorkflowView{Enabled: true, Preview: &preview}, nil
}

func (p webViewProvider) writeWorkflowAction(ctx context.Context, operations service, request webui.WriteActionRequest) webui.WriteActionResponse {
	if len(request.Action) > 32 || len(request.Revision) > 20 || len(request.Fields["binding_hash"]) > 64 {
		return webValidationWrite(fmt.Errorf("workflow action input bounds exceeded"))
	}

	if request.Action == "workflow.start" {
		return webWorkflowStart(ctx, operations, request.Fields["binding"])
	}

	revision, err := strconv.ParseUint(request.Revision, 10, 64)
	if err != nil || revision == 0 {
		return webValidationWrite(fmt.Errorf("workflow action requires the displayed invocation revision"))
	}

	control := WorkflowControlRequest{Action: strings.TrimPrefix(request.Action, "workflow."), ID: request.ID, ExpectedRevision: revision, BindingHash: request.Fields["binding_hash"], StepLimit: 1}

	control, err = webWorkflowIntegrationControl(ctx, operations, control, request.Fields)
	if err != nil {
		status, readErr := operations.WorkflowStatus(ctx, request.ID, WorkflowStatusRequest{Limit: 8})

		return webWorkflowResponse(status, errors.Join(err, readErr))
	}

	status, err := operations.ControlWorkExtent(ctx, control)

	return webWorkflowResponse(status, err)
}

func webWorkflowStart(ctx context.Context, operations service, value string) webui.WriteActionResponse {
	binding, err := decodeWebWorkflowBinding(value)
	if err != nil {
		return webValidationWrite(err)
	}

	result, err := operations.StartBoundWorkExtent(ctx, binding.Request, binding.Selected, 1)
	if result.Workflow == nil {
		return webWorkflowResponse(WorkflowWorkStatus{}, err)
	}

	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()

	status, readErr := operations.WorkflowStatus(cleanup, result.Workflow.ID, WorkflowStatusRequest{Limit: 8})

	return webWorkflowResponse(status, errors.Join(err, readErr))
}

func decodeWebWorkflowBinding(value string) (webWorkflowBinding, error) {
	var binding webWorkflowBinding
	if len(value) == 0 || len(value) > 8192 {
		return binding, fmt.Errorf("workflow confirmation binding bounds exceeded")
	}

	decoder := json.NewDecoder(strings.NewReader(value))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&binding); err != nil {
		return binding, fmt.Errorf("invalid workflow confirmation: %w", err)
	}

	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return binding, fmt.Errorf("workflow confirmation must contain one binding")
	}

	return binding, nil
}

func webWorkflowResponse(status WorkflowWorkStatus, err error) webui.WriteActionResponse {
	result := webui.WriteActionResponse{Status: webui.WriteStatusOK, Summary: "Workflow checkpoint retained.", Workflow: webWorkflowStatusPointer(status)}
	if status.ID != "" {
		result.ReloadURL = "/workflows?id=" + url.QueryEscape(status.ID)
		result.Operation = webOperation("workflow", status.Status, status.ID, "Task validation and confirmed integration are separate.")
	}

	if err == nil {
		return result
	}

	result.Status, result.Summary, result.Validation = webui.WriteStatusValidation, "Workflow action failed.", []string{err.Error()}
	if errors.Is(err, workflowstate.ErrConflict) {
		result.Status = webui.WriteStatusStale
	}

	if errors.Is(err, core.ErrWorkflowOwned) {
		result.Status, result.BlockedReason = webui.WriteStatusBlocked, "Another owner controls this workflow."
	}

	return result
}

func webWorkflowStatusPointer(status WorkflowWorkStatus) *webui.WorkflowRunView {
	if status.ID == "" {
		return nil
	}

	view := webWorkflowRun(status)

	return &view
}

func webWorkflowFilter(value string) []string {
	if value == "" {
		return nil
	}

	return []string{value}
}
