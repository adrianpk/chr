package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func retainedTelegramHarnessChange(request core.TelegramOperationRequest) (telegramHarnessPrepared, error) {
	var prepared telegramHarnessPrepared
	if err := decodeOperationJSON(string(request.PreparedInput), &prepared); err != nil {
		return prepared, err
	}

	choices, err := telegramHarnessSourceChoices(request.Source)
	if err != nil {
		return prepared, workflowstate.ErrConflict
	}

	arguments, err := json.Marshal(choices)
	if err != nil {
		return prepared, err
	}

	var retained telegramHarnessChoices
	if err := decodeOperationJSON(string(request.Arguments), &retained); err != nil {
		return prepared, err
	}

	retainedArguments, err := json.Marshal(retained)
	if err != nil {
		return prepared, err
	}

	if string(arguments) != string(retainedArguments) || !telegramHarnessPreparedBinding(prepared.Change, request, choices) {
		return prepared, workflowstate.ErrConflict
	}

	return prepared, nil
}

func telegramHarnessSourceChoices(source core.TelegramReceivedUpdate) (telegramHarnessChoices, error) {
	choices, recognized, err := parseTelegramHarness(source)
	if err != nil || !recognized || (choices.Worker == nil && choices.Generation == nil) {
		return choices, workflowstate.ErrConflict
	}

	return choices, nil
}

func telegramHarnessPreparedBinding(change core.HarnessSelectionChangeRequest, request core.TelegramOperationRequest, choices telegramHarnessChoices) bool {
	return request.OperationID == core.TelegramHarnessSelectionOperation && change.Binding == (workflowstate.MutationBinding{RequestID: request.ID, Operator: request.Operator}) && change.ProjectRef == request.Target.ProjectID && len(change.ExpectedRevision) == 64 && sameTelegramHarnessChoice(change.Worker, choices.Worker) && sameTelegramHarnessChoice(change.Generation, choices.Generation)
}

func sameTelegramHarnessChoice(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}

	return *a == *b
}

func (r telegramReceiver) applyHarnessAction(ctx context.Context, claim core.TelegramActionClaim, config core.TelegramReceiverConfig) error {
	if claim.Choice == "cancel" {
		return r.finishHarnessAction(ctx, claim, config, "blocked", nil, "Harness change cancelled. Defaults are unchanged.")
	}

	operations, ok := r.operations.(telegramHarnessService)
	if !ok || claim.Choice != "confirm" {
		return r.finishHarnessActionFailure(ctx, claim, config, ErrRecordUnsupported)
	}

	prepared, err := retainedTelegramHarnessChange(claim.Request)
	if err != nil {
		return r.finishHarnessActionFailure(ctx, claim, config, workflowstate.ErrConflict)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	ctx, expire := context.WithDeadline(ctx, claim.Action.ExpiresAt)
	defer expire()

	result, err := operations.ChangeTelegramHarnessSelections(ctx, claim.Request.Target, prepared.Change, r.harnessPublishAuthority(claim))
	if err != nil {
		return r.finishHarnessActionFailure(ctx, claim, config, err)
	}

	text := fmt.Sprintf("Harness defaults saved for %s.\nWorker: %s\nGeneration: %s\nNew operations only; admitted work keeps its binding.", result.ProjectRef, telegramHarnessSavedRole(result.Worker), telegramHarnessSavedRole(result.Generation))

	return r.finishHarnessAction(ctx, claim, config, "confirmed", &result, text)
}

func (r telegramReceiver) harnessPublishAuthority(claim core.TelegramActionClaim) func(context.Context) error {
	return func(ctx context.Context) error {
		if !claim.Action.ExpiresAt.After(time.Now()) {
			return workflowstate.ErrConflict
		}

		input, err := r.harnessAuthorityInput(claim.Request)
		if err != nil {
			return err
		}

		return r.store.ValidateTelegramHarnessTarget(ctx, input, claim.Request)
	}
}

func (r telegramReceiver) finishHarnessActionFailure(ctx context.Context, claim core.TelegramActionClaim, config core.TelegramReceiverConfig, failure error) error {
	text := "The selected profile is no longer locally prepared. Defaults are unchanged; inspect /harness and issue a new change."
	if errors.Is(failure, workflowstate.ErrConflict) {
		text = "The project, actor, configuration revision or request binding changed. Inspect /harness before issuing a new change."
	} else if !errors.Is(failure, core.ErrHarnessSelectionPreparation) && !errors.Is(failure, ErrRecordUnsupported) {
		// An uncertain transport/owner failure keeps the original action applying.
		return failure
	}

	return r.finishHarnessAction(ctx, claim, config, "blocked", nil, text)
}

func (r telegramReceiver) finishHarnessAction(ctx context.Context, claim core.TelegramActionClaim, config core.TelegramReceiverConfig, state string, snapshot *core.HarnessSelectionSnapshot, text string) error {
	facts, err := json.Marshal(struct {
		State    string                         `json:"state"`
		Snapshot *core.HarnessSelectionSnapshot `json:"snapshot,omitempty"`
	}{state, snapshot})
	if err != nil {
		return err
	}

	receipt := ""
	if snapshot != nil {
		receipt = "harness-selection:" + claim.Request.ID
	}

	return r.store.FinishTelegramAction(ctx, claim, core.TelegramActionResult{Message: telegramOutcomeMessage("39", text), State: state, ReceiptRef: receipt, Facts: facts, Text: text, ConfigDigest: config.Digest()}, time.Now())
}

func telegramHarnessSavedRole(profile string) string {
	if profile == "" {
		return "legacy dispatch"
	}

	return profile
}
