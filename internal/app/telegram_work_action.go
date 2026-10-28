package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r telegramReceiver) applyWorkAdmission(ctx context.Context, claim telegramcore.ActionClaim, config telegramcore.ReceiverConfig) error {
	operations, ok := r.operations.(telegramWorkService)
	if !ok || claim.Choice != "start" {
		return r.finishWorkActionFailure(ctx, claim, ErrRecordUnsupported, config)
	}

	preview, err := telegramRetainedWorkPreview(claim.Request)
	if err != nil {
		return err
	}

	if err := r.validateRecordAuthority(ctx, claim.Request); err != nil {
		return r.finishWorkActionFailure(ctx, claim, err, config)
	}

	binding := WorkflowBoundAdmission{RequestKey: claim.Request.ID, Policy: preview.View.Policy, ExecutionPolicy: preview.View.ExecutionPolicy, Check: r.workAdmissionCheck(claim)}

	result, err := operations.StartTelegramWork(ctx, claim.Request.Target, preview, binding)
	if errors.Is(err, workflowstate.ErrConflict) || errors.Is(err, workflowstate.ErrOwned) || errors.Is(err, workflowcore.ErrWorkflowCapacity) || errors.Is(err, ErrRecordUnsupported) {
		return r.finishWorkActionFailure(ctx, claim, err, config)
	}

	if err != nil {
		return err
	}

	return r.finishWorkAdmission(ctx, claim, result, config)
}

func (r telegramReceiver) workAdmissionCheck(claim telegramcore.ActionClaim) workflowcore.AdmissionCheck {
	return func(ctx context.Context) error {
		if !claim.Action.ExpiresAt.After(time.Now()) {
			return workflowstate.ErrConflict
		}

		return r.validateWorkAdmissionAuthority(ctx, claim.Request)
	}
}

func telegramRetainedWorkPreview(request telegramcore.OperationRequest) (WorkflowBoundPreview, error) {
	var preview WorkflowBoundPreview
	if err := decodeBoundedOperationJSON(string(request.PreparedInput), &preview, 1048576); err != nil {
		return preview, err
	}

	var selector workflowstate.Selector
	if err := decodeOperationJSON(string(request.Arguments), &selector); err != nil {
		return preview, err
	}

	if preview.Request.Selector != selector || preview.Request.SnapshotHash != preview.View.Preview.SnapshotHash || preview.Request.Target != preview.View.Target {
		return preview, workflowstate.ErrConflict
	}

	return preview, nil
}

func telegramWorkPreviewText(preview WorkflowBoundPreview) string {
	var text strings.Builder

	selection, policy, execution := preview.View.Preview, preview.View.Policy, preview.View.ExecutionPolicy
	fmt.Fprintf(&text, "Start %s %s?\n\nSelected tickets: %d. Publication: %s (%s).\nTarget: %s\n", strings.ReplaceAll(selection.Selector.Extent, "_", " "), selection.Selector.Identity, selection.Total, policy.Mode, policy.Method, preview.View.Target.Repository)
	fmt.Fprintf(&text, "Limits: %d worker attempts; %ds per worker; %ds validation (zero uses each ticket contract). Include later tickets: %t.\n", execution.MaxWorkerAttempts, execution.WorkerTimeoutSeconds, execution.ValidationTimeoutSeconds, execution.IncludeLaterTickets)

	for _, member := range selection.Members[:min(8, len(selection.Members))] {
		fmt.Fprintf(&text, "- %s · %s\n", member.Key, member.Readiness)
	}

	if selection.Total > len(selection.Members) || len(selection.Members) > 8 {
		text.WriteString("This is a shortened selection view. Review the complete scope before confirming.\n")
	}

	if ticket := preview.ReadyTicket; ticket != nil {
		fmt.Fprintf(&text, "Confirmation also makes ticket %s ready. Its validation commands:\n", ticket.Key)

		for _, argv := range ticket.Validation.Commands {
			fmt.Fprintf(&text, "%q\n", argv)
		}
	}

	text.WriteString("Confirming admits this exact selection under the displayed policy.")

	return text.String()
}

type telegramWorkAdmissionResult struct {
	Invocation telegramcore.InvocationBinding `json:"invocation"`
	RequestKey string                         `json:"request_key"`
	Selector   workflowstate.Selector         `json:"selector"`
	Status     string                         `json:"status"`
	Members    int                            `json:"members"`
}

func (r telegramReceiver) finishWorkAdmission(ctx context.Context, claim telegramcore.ActionClaim, result AgentWorkResult, config telegramcore.ReceiverConfig) error {
	if result.Workflow == nil || result.Workflow.ID == "" {
		return workflowstate.ErrConflict
	}

	run := result.Workflow
	binding := telegramcore.InvocationBinding{ID: run.ID, Revision: run.Revision}

	facts, err := json.Marshal(telegramWorkAdmissionResult{Invocation: binding, RequestKey: run.RequestKey, Selector: run.Selection.Selector, Status: string(run.Status), Members: len(run.Members)})
	if err != nil {
		return err
	}

	text := fmt.Sprintf("Work admitted: %s %s.\n\nState: %s; selected tickets: %d. Execution, validation and integration are reported separately.\nUse /status %s to inspect this work.", run.Selection.Selector.Extent, run.Selection.Selector.Identity, run.Status, len(run.Members), run.ID)

	return r.store.FinishAction(ctx, claim, telegramcore.ActionResult{Message: telegramOutcomeMessage("32", text), State: telegramcore.OperationConfirmed, ReceiptRef: "workflow:" + run.ID, Facts: facts, Text: text, ConfigDigest: config.Digest(), Invocation: &binding}, time.Now())
}
