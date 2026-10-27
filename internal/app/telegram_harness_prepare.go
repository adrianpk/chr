package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

type telegramHarnessPrepared struct {
	Change    workspaceconfig.HarnessSelectionChangeRequest `json:"change"`
	Preview   string                                        `json:"preview"`
	ExpiresAt time.Time                                     `json:"expires_at"`
}

func (r telegramReceiver) prepareHarnessChange(ctx context.Context, operations telegramHarnessService, change telegramHarnessChange) error {
	operationLookup, err := r.store.Operation(ctx, change.Input.Update, 0)
	request, found := operationLookup.Request, operationLookup.Found

	if err != nil {
		return err
	}

	if !found {
		request, err = r.reserveHarnessChange(ctx, operations, telegramHarnessChange{User: change.User, Input: change.Input, Target: change.Target, Choices: change.Choices})
		if err != nil {
			return r.recordDiagnostic(ctx, change.Input.Update, change.Input.Config, "The selection could not be prepared. Choose configured, locally ready profiles for the selected project and retry. Defaults are unchanged.")
		}
	}

	if request.OperationID != telegramcore.HarnessSelectionOperation {
		return workflowstate.ErrConflict
	}

	if request.State != telegramcore.OperationReady {
		return nil
	}

	return r.offerHarnessChange(ctx, request)
}

func (r telegramReceiver) reserveHarnessChange(ctx context.Context, operations telegramHarnessService, change telegramHarnessChange) (telegramcore.OperationRequest, error) {
	status, err := operations.InspectTelegramHarnessSelections(ctx, change.Target)
	if err != nil {
		return telegramcore.OperationRequest{}, err
	}

	prepared, err := prepareTelegramHarnessChange(status.Selections, workflowstate.MutationBinding{RequestID: telegramcore.OperationID(change.Input.Update, 0), Operator: change.User.OperatorID}, change.Choices)
	if err != nil {
		return telegramcore.OperationRequest{}, err
	}

	arguments, err := json.Marshal(change.Choices)
	if err != nil {
		return telegramcore.OperationRequest{}, err
	}

	data, err := json.Marshal(prepared)
	if err != nil {
		return telegramcore.OperationRequest{}, err
	}

	return r.store.ReserveHarnessOperation(ctx, change.Input, telegramcore.OperationRequest{Source: change.Input.Update, Operator: change.User.OperatorID, Target: change.Target, OperationID: telegramcore.HarnessSelectionOperation, Arguments: arguments}, data, time.Now())
}

func prepareTelegramHarnessChange(status HarnessSelectionStatus, binding workflowstate.MutationBinding, choices telegramHarnessChoices) (telegramHarnessPrepared, error) {
	worker, err := telegramHarnessChosenLabel(status, choices.Worker, false)
	if err != nil {
		return telegramHarnessPrepared{}, err
	}

	generation, err := telegramHarnessChosenLabel(status, choices.Generation, true)
	if err != nil {
		return telegramHarnessPrepared{}, err
	}

	change := workspaceconfig.HarnessSelectionChangeRequest{Binding: binding, ProjectRef: status.Snapshot.ProjectRef, ExpectedRevision: status.Snapshot.Revision, Worker: choices.Worker, Generation: choices.Generation}
	preview := fmt.Sprintf("Change harness defaults for %s?\nWorker profile: %s\nGeneration profile: %s\nNew operations only. Running, paused and resumed work keeps its binding. No model call or private setup is performed.", change.ProjectRef, worker, generation)

	return telegramHarnessPrepared{Change: change, Preview: telegramcore.BoundReceiverText(preview), ExpiresAt: time.Now().Add(10 * time.Minute)}, nil
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

func (r telegramReceiver) offerHarnessChange(ctx context.Context, request telegramcore.OperationRequest) error {
	prepared, err := retainedTelegramHarnessChange(request)
	if err != nil {
		return err
	}

	input, err := r.harnessAuthorityInput(request)
	if err != nil {
		return err
	}

	if err := r.store.ValidateHarnessTarget(ctx, input, request); err != nil {
		return r.finishHarnessPreparationFailure(ctx, request, input.Config, err)
	}

	if !prepared.ExpiresAt.After(time.Now()) {
		return r.finishHarnessPreparationFailure(ctx, request, input.Config, workflowstate.ErrConflict)
	}

	return r.offerConfirmationDocument(ctx, request, telegramcore.ActionOffer{Input: input, RequestID: request.ID, Choices: []string{"confirm", "cancel"}, ExpiresAt: prepared.ExpiresAt}, telegramConfirmationDocument{Family: "39", Text: prepared.Preview, Complete: "Only new operations use these defaults. Admitted work keeps its original binding."})
}

func (r telegramReceiver) finishHarnessPreparationFailure(ctx context.Context, request telegramcore.OperationRequest, config telegramcore.ReceiverConfig, failure error) error {
	if !errors.Is(failure, workflowstate.ErrConflict) {
		return failure
	}

	text := "The selected project or confirmation binding changed. Inspect /harness before issuing a new change."

	return r.store.FinishOperation(ctx, request, telegramcore.OperationCompletion{State: telegramcore.OperationBlocked, ReceiptRef: "", Facts: json.RawMessage(`{"state":"blocked"}`), Text: text, ConfigDigest: config.Digest(), At: time.Now()})
}

func (r telegramReceiver) harnessAuthorityInput(request telegramcore.OperationRequest) (telegramcore.ConversationInput, error) {
	input, err := r.preparedWorkAuthorityInput(request)
	// Harness changes must use peer selection rather than an explicit session override.
	input.Conversation = ""

	return input, err
}

type telegramHarnessChange struct {
	User    telegramcore.AllowedUser
	Input   telegramcore.ConversationInput
	Target  telegramcore.OperationTarget
	Choices telegramHarnessChoices
}
