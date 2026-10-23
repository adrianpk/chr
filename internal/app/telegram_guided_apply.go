package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func (r telegramReceiver) continueTelegramGuided(ctx context.Context, request core.TelegramOperationRequest, config core.TelegramReceiverConfig) error {
	switch request.State {
	case "confirmed", "blocked":
		return nil
	}

	operations, ok := r.operations.(telegramGuidedService)
	if !ok {
		return ErrRecordUnsupported
	}

	current, err := r.validateAuthoringAuthority(ctx, request, request.State == "applying")
	if err != nil {
		return r.finishGuidedFailure(ctx, request, config, err)
	}

	owner, err := telegramRetainedGuidedInput(request)
	if err != nil {
		return err
	}

	if err := r.validateGuidedRunBinding(ctx, request, current, owner.Command.RunID); err != nil {
		return r.finishGuidedFailure(ctx, request, current, err)
	}

	if err := r.store.BeginTelegramOperation(ctx, request, time.Now()); err != nil {
		return err
	}

	operations = r.boundTelegramGuidedService(operations, request, owner.Command.RunID)

	session, err := applyTelegramGuided(ctx, operations, request.Target, owner)
	if errors.Is(err, core.ErrWorkflowConflict) {
		return r.finishGuidedFailure(ctx, request, current, err)
	}

	if err != nil {
		return err
	}

	observed, err := operations.InspectTelegramGuided(ctx, request.Target, session.ID)
	if err != nil {
		return err
	}

	return r.finishTelegramGuided(ctx, request, observed, current, owner.Document)
}

func (r telegramReceiver) finishTelegramGuided(ctx context.Context, request core.TelegramOperationRequest, observed GuidedUseCaseInspection, current core.TelegramReceiverConfig, kind string) error {
	progress, facts, err := telegramGuidedResult(observed)
	if err != nil {
		return err
	}

	text := telegramGuidedPrompt(observed)
	request.Message = telegramOutcomeMessage("43", text)

	request.Delivery, err = r.guidedDocument(ctx, request, observed, kind)
	if errors.Is(err, core.ErrWorkflowConflict) {
		return r.finishGuidedFailure(ctx, request, current, telegramGuidedRoutingFailure("That original has not been sealed for this run. Send Show current step to continue; saved inputs are unchanged."))
	}

	if err != nil {
		return err
	}

	err = r.store.FinishTelegramGuided(ctx, request, progress, facts, text, current.Digest(), time.Now())
	if errors.Is(err, core.ErrWorkflowConflict) {
		return r.finishOvertakenGuided(ctx, request, progress, facts, current)
	}

	return err
}

func telegramGuidedResult(observed GuidedUseCaseInspection) (core.TelegramGuidedProgress, json.RawMessage, error) {
	progress, err := telegramGuidedProgress(observed)
	if err != nil {
		return progress, nil, err
	}

	session := observed.Session
	facts, err := json.Marshal(map[string]any{"session": session.ID, "revision": session.Revision, "stage": session.Guided.Stage})

	return progress, facts, err
}

func (r telegramReceiver) finishGuidedFailure(ctx context.Context, request core.TelegramOperationRequest, config core.TelegramReceiverConfig, failure error) error {
	text := telegramGuidedFailure(failure)

	facts, err := json.Marshal(map[string]string{"state": "blocked", "message": text})
	if err != nil {
		return err
	}

	request.Message = telegramOutcomeMessage("43", text)

	return r.store.FinishTelegramOperation(ctx, request, "blocked", "", facts, text, config.Digest(), time.Now())
}

func telegramGuidedFailure(failure error) string {
	var routing telegramGuidedRoutingFailure
	if errors.As(failure, &routing) {
		return string(routing)
	}

	if safe, ok := core.TelegramConversationFailure(failure); ok {
		return safe
	}

	return "The use case changed or its current step needs recovery. Saved inputs and work are retained. Send Show current step to read the current prompt."
}

func telegramRetainedGuidedInput(request core.TelegramOperationRequest) (telegramGuidedInput, error) {
	var owner telegramGuidedInput
	if err := decodeOperationJSON(string(request.PreparedInput), &owner); err != nil {
		return owner, err
	}

	var args telegramAuthoringArguments
	if err := decodeOperationJSON(string(request.Arguments), &args); err != nil {
		return owner, err
	}

	if !telegramGuidedOwnerMatches(owner, request, args) {
		return owner, core.ErrWorkflowConflict
	}

	if !telegramGuidedSourceMatches(owner, request) {
		return owner, core.ErrWorkflowConflict
	}

	return owner, nil
}

func telegramGuidedOwnerMatches(owner telegramGuidedInput, request core.TelegramOperationRequest, args telegramAuthoringArguments) bool {
	return args.PayloadHash == recordDigest(request.PreparedInput) && owner.Actor == request.Operator && owner.Command.ProjectID == request.Target.ProjectID && owner.Command.Namespace == request.Target.Namespace
}

func telegramGuidedSourceMatches(owner telegramGuidedInput, request core.TelegramOperationRequest) bool {
	if owner.Start {
		return telegramGuidedStart(request.Source.Text) && owner.Command.RunID == request.Target.ConversationID
	}

	if owner.Retry {
		text := strings.ToLower(strings.TrimSpace(request.Source.Text))

		return (text == "retry" || text == "continue") && owner.Command.OperationID != "" && !owner.Start && !owner.Inspect && !owner.Offer
	}

	return owner.Command.OperationID == request.ID
}

func applyTelegramGuided(ctx context.Context, operations telegramGuidedService, target core.TelegramOperationTarget, input telegramGuidedInput) (core.AuthoringSession, error) {
	session, err := operations.ApplyTelegramGuided(ctx, target, input)
	// A missing prerequisite is a retained stage outcome, not a failed input.
	var prerequisite *core.GuidedPrerequisiteError
	if errors.As(err, &prerequisite) && session.Guided != nil {
		return session, nil
	}

	return session, err
}
