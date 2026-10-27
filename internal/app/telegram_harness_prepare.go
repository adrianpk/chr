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

type telegramHarnessPrepared struct {
	Change    core.HarnessSelectionChangeRequest `json:"change"`
	Preview   string                             `json:"preview"`
	ExpiresAt time.Time                          `json:"expires_at"`
}

func (r telegramReceiver) prepareHarnessChange(ctx context.Context, operations telegramHarnessService, user core.TelegramAllowedUser, input core.TelegramConversationInput, target core.TelegramOperationTarget, choices telegramHarnessChoices) error {
	request, found, err := r.store.TelegramOperation(ctx, input.Update, 0)
	if err != nil {
		return err
	}

	if !found {
		request, err = r.reserveHarnessChange(ctx, operations, user, input, target, choices)
		if err != nil {
			return r.recordDiagnostic(ctx, input.Update, input.Config, "The selection could not be prepared. Choose configured, locally ready profiles for the selected project and retry. Defaults are unchanged.")
		}
	}

	if request.OperationID != core.TelegramHarnessSelectionOperation {
		return workflowstate.ErrConflict
	}

	if request.State != "ready" {
		return nil
	}

	return r.offerHarnessChange(ctx, request)
}

func (r telegramReceiver) reserveHarnessChange(ctx context.Context, operations telegramHarnessService, user core.TelegramAllowedUser, input core.TelegramConversationInput, target core.TelegramOperationTarget, choices telegramHarnessChoices) (core.TelegramOperationRequest, error) {
	status, err := operations.InspectTelegramHarnessSelections(ctx, target)
	if err != nil {
		return core.TelegramOperationRequest{}, err
	}

	prepared, err := prepareTelegramHarnessChange(status.Selections, core.WorkflowMutationBinding{RequestID: core.TelegramOperationID(input.Update, 0), Operator: user.OperatorID}, choices)
	if err != nil {
		return core.TelegramOperationRequest{}, err
	}

	arguments, err := json.Marshal(choices)
	if err != nil {
		return core.TelegramOperationRequest{}, err
	}

	data, err := json.Marshal(prepared)
	if err != nil {
		return core.TelegramOperationRequest{}, err
	}

	return r.store.ReserveTelegramHarnessOperation(ctx, input, core.TelegramOperationRequest{Source: input.Update, Operator: user.OperatorID, Target: target, OperationID: core.TelegramHarnessSelectionOperation, Arguments: arguments}, data, time.Now())
}

func prepareTelegramHarnessChange(status HarnessSelectionStatus, binding core.WorkflowMutationBinding, choices telegramHarnessChoices) (telegramHarnessPrepared, error) {
	worker, err := telegramHarnessChosenLabel(status, choices.Worker, false)
	if err != nil {
		return telegramHarnessPrepared{}, err
	}

	generation, err := telegramHarnessChosenLabel(status, choices.Generation, true)
	if err != nil {
		return telegramHarnessPrepared{}, err
	}

	change := core.HarnessSelectionChangeRequest{Binding: binding, ProjectRef: status.Snapshot.ProjectRef, ExpectedRevision: status.Snapshot.Revision, Worker: choices.Worker, Generation: choices.Generation}
	preview := fmt.Sprintf("Change harness defaults for %s?\nWorker profile: %s\nGeneration profile: %s\nNew operations only. Running, paused and resumed work keeps its binding. No model call or private setup is performed.", change.ProjectRef, worker, generation)

	return telegramHarnessPrepared{Change: change, Preview: core.BoundTelegramReceiverText(preview), ExpiresAt: time.Now().Add(10 * time.Minute)}, nil
}

func telegramHarnessChosenLabel(status HarnessSelectionStatus, profile *string, generation bool) (string, error) {
	if profile == nil {
		return "unchanged", nil
	}

	if *profile == "" {
		return "restore legacy dispatch", nil
	}

	for _, choice := range status.Profiles {
		if choice.Name != *profile {
			continue
		}

		readiness := choice.Worker
		if generation {
			readiness = choice.Generation
		}

		if !readiness.Ready() {
			return "", ErrRecordUnsupported
		}

		return telegramHarnessLabel(choice.Name, choice.Harness, choice.ModelProvider, choice.Model), nil
	}

	return "", ErrRecordUnsupported
}

func (r telegramReceiver) offerHarnessChange(ctx context.Context, request core.TelegramOperationRequest) error {
	prepared, err := retainedTelegramHarnessChange(request)
	if err != nil {
		return err
	}

	input, err := r.harnessAuthorityInput(request)
	if err != nil {
		return err
	}

	if err := r.store.ValidateTelegramHarnessTarget(ctx, input, request); err != nil {
		return r.finishHarnessPreparationFailure(ctx, request, input.Config, err)
	}

	if !prepared.ExpiresAt.After(time.Now()) {
		return r.finishHarnessPreparationFailure(ctx, request, input.Config, workflowstate.ErrConflict)
	}

	return r.offerConfirmationDocument(ctx, request, core.TelegramActionOffer{Input: input, RequestID: request.ID, Choices: []string{"confirm", "cancel"}, ExpiresAt: prepared.ExpiresAt}, "39", prepared.Preview, "Only new operations use these defaults. Admitted work keeps its original binding.")
}

func (r telegramReceiver) finishHarnessPreparationFailure(ctx context.Context, request core.TelegramOperationRequest, config core.TelegramReceiverConfig, failure error) error {
	if !errors.Is(failure, workflowstate.ErrConflict) {
		return failure
	}

	text := "The selected project or confirmation binding changed. Inspect /harness before issuing a new change."

	return r.store.FinishTelegramOperation(ctx, request, "blocked", "", json.RawMessage(`{"state":"blocked"}`), text, config.Digest(), time.Now())
}

func (r telegramReceiver) harnessAuthorityInput(request core.TelegramOperationRequest) (core.TelegramConversationInput, error) {
	input, err := r.preparedWorkAuthorityInput(request)
	// Harness changes must use peer selection rather than an explicit session override.
	input.Conversation = ""

	return input, err
}
