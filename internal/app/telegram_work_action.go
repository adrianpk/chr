package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func (r telegramReceiver) applyWorkAdmission(ctx context.Context, claim core.TelegramActionClaim, config core.TelegramReceiverConfig) error {
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
	if errors.Is(err, core.ErrWorkflowConflict) || errors.Is(err, core.ErrWorkflowOwned) || errors.Is(err, core.ErrWorkflowCapacity) || errors.Is(err, ErrRecordUnsupported) {
		return r.finishWorkActionFailure(ctx, claim, err, config)
	}

	if err != nil {
		return err
	}

	return r.finishWorkAdmission(ctx, claim, result, config)
}

func (r telegramReceiver) workAdmissionCheck(claim core.TelegramActionClaim) core.WorkflowAdmissionCheck {
	return func(ctx context.Context) error {
		if !claim.Action.ExpiresAt.After(time.Now()) {
			return core.ErrWorkflowConflict
		}

		return r.validateWorkAdmissionAuthority(ctx, claim.Request)
	}
}

func telegramRetainedWorkPreview(request core.TelegramOperationRequest) (WorkflowBoundPreview, error) {
	var preview WorkflowBoundPreview
	if err := decodeBoundedOperationJSON(string(request.PreparedInput), &preview, 1048576); err != nil {
		return preview, err
	}

	var selector core.WorkflowSelector
	if err := decodeOperationJSON(string(request.Arguments), &selector); err != nil {
		return preview, err
	}

	if preview.Request.Selector != selector || preview.Request.SnapshotHash != preview.View.Preview.SnapshotHash || preview.Request.Target != preview.View.Target {
		return preview, core.ErrWorkflowConflict
	}

	return preview, nil
}

func telegramWorkPreviewText(preview WorkflowBoundPreview) string {
	var text strings.Builder

	selection := preview.View.Preview
	policy := preview.View.Policy
	fmt.Fprintf(&text, "Native work preview: %s %s\nSnapshot: %s\nTarget commit: %s\nPublication: %s (%s, %s)\nPolicy revision: %s\nSelected tickets: %d; displayed: %d\n", selection.Selector.Extent, selection.Selector.Identity, selection.SnapshotHash, preview.View.Target.Commit, policy.Mode, policy.Method, policy.Source, policy.Revision, selection.Total, len(selection.Members))

	execution := preview.View.ExecutionPolicy
	fmt.Fprintf(&text, "Include later tickets: %t; worker timeout: %ds; maximum worker attempts: %d; validation timeout: %ds (zero uses each contract).\n", execution.IncludeLaterTickets, execution.WorkerTimeoutSeconds, execution.MaxWorkerAttempts, execution.ValidationTimeoutSeconds)

	for _, member := range selection.Members[:min(8, len(selection.Members))] {
		fmt.Fprintf(&text, "%s revision %d contract %s; %s\n", member.Key, member.Revision, member.ContractHash, member.Readiness)
	}

	if selection.Total > 8 {
		fmt.Fprintf(&text, "%d more tickets are bound by the snapshot; inspect native work before starting.\n", selection.Total-8)
	}

	text.WriteString("Start exactly this extent under the displayed policy?")

	return text.String()
}

type telegramWorkAdmissionResult struct {
	Invocation core.TelegramInvocationBinding `json:"invocation"`
	RequestKey string                         `json:"request_key"`
	Selector   core.WorkflowSelector          `json:"selector"`
	Status     string                         `json:"status"`
	Members    int                            `json:"members"`
}

func (r telegramReceiver) finishWorkAdmission(ctx context.Context, claim core.TelegramActionClaim, result AgentWorkResult, config core.TelegramReceiverConfig) error {
	if result.Workflow == nil || result.Workflow.ID == "" {
		return core.ErrWorkflowConflict
	}

	run := result.Workflow
	binding := core.TelegramInvocationBinding{ID: run.ID, Revision: run.Revision}

	facts, err := json.Marshal(telegramWorkAdmissionResult{Invocation: binding, RequestKey: run.RequestKey, Selector: run.Selection.Selector, Status: run.Status, Members: len(run.Members)})
	if err != nil {
		return err
	}

	text := fmt.Sprintf("Native work admitted.\nInvocation: %s\nRequest: %s\nExtent: %s %s\nState: %s; admitted tickets: %d. Admission is not evidence of validation or integration.", run.ID, claim.Request.ID, run.Selection.Selector.Extent, run.Selection.Selector.Identity, run.Status, len(run.Members))

	return r.store.FinishTelegramAction(ctx, claim, core.TelegramActionResult{State: "confirmed", ReceiptRef: "workflow:" + run.ID, Facts: facts, Text: text, ConfigDigest: config.Digest(), Invocation: &binding}, time.Now())
}
