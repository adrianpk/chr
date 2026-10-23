package app

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func (r telegramReceiver) offerGuidedControls(ctx context.Context, request core.TelegramOperationRequest, config core.TelegramReceiverConfig) error {
	input, err := telegramRetainedGuidedInput(request)
	if err != nil {
		return err
	}

	if !input.Offer || request.OperationID != "guided.step" || strings.ToLower(strings.TrimSpace(request.Source.Text)) != "show controls" {
		return core.ErrWorkflowConflict
	}

	current, err := r.validateAuthoringAuthority(ctx, request, false)
	if err != nil {
		return r.finishGuidedFailure(ctx, request, config, err)
	}

	user, _ := current.AuthorizedUser(request.Source.Peer.ChatID, request.Source.Peer.TopicID, request.Source.Peer.UserID, request.Source.ChatType)

	projects, err := r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return err
	}

	_, err = r.store.OfferTelegramAction(ctx, core.TelegramActionOffer{
		Input: core.TelegramConversationInput{Update: request.Source, Config: current, Projects: projects}, RequestID: request.ID, Choices: input.Choices,
		Message: telegramOutcomeMessage("43", "Choose an action for "+input.Name+". Saved inputs and existing work limits are retained. Send Show current step to return to the input prompt."), ExpiresAt: time.Now().Add(30 * time.Minute),
	}, time.Now())

	return err
}

func (r telegramReceiver) applyGuidedAction(ctx context.Context, claim core.TelegramActionClaim, config core.TelegramReceiverConfig) error {
	operations, ok := r.operations.(telegramGuidedService)
	if !ok {
		return ErrRecordUnsupported
	}

	input, current, err := r.prepareGuidedAction(ctx, claim)
	if errors.Is(err, core.ErrWorkflowConflict) {
		return r.finishGuidedActionFailure(ctx, claim, config, err)
	}

	if err != nil {
		return err
	}

	input.Offer, input.Command.Action = false, claim.Choice

	operations = r.boundTelegramGuidedService(operations, claim.Request, input.Command.RunID)

	session, err := applyTelegramGuided(ctx, operations, claim.Request.Target, input)
	if errors.Is(err, core.ErrWorkflowConflict) {
		return r.finishGuidedActionFailure(ctx, claim, current, err)
	}

	if err != nil {
		return err
	}

	observed, err := operations.InspectTelegramGuided(ctx, claim.Request.Target, session.ID)
	if err != nil {
		return err
	}

	return r.finishGuidedAction(ctx, claim, current, observed)
}

func (r telegramReceiver) finishGuidedAction(ctx context.Context, claim core.TelegramActionClaim, config core.TelegramReceiverConfig, observed GuidedUseCaseInspection) error {
	progress, facts, err := telegramGuidedResult(observed)
	if err != nil {
		return err
	}

	text := telegramGuidedPrompt(observed)

	delivery, err := r.guidedDocument(ctx, claim.Request, observed, "")
	if err != nil {
		return err
	}

	return r.store.FinishTelegramAction(ctx, claim, core.TelegramActionResult{State: "confirmed", Delivery: delivery, Guided: &progress, ReceiptRef: progress.SessionRef, Facts: facts, Text: text, Message: telegramOutcomeMessage("43", text), ConfigDigest: config.Digest()}, time.Now())
}

func (r telegramReceiver) finishGuidedActionFailure(ctx context.Context, claim core.TelegramActionClaim, config core.TelegramReceiverConfig, failure error) error {
	text := telegramGuidedFailure(failure)

	return r.store.FinishTelegramAction(ctx, claim, core.TelegramActionResult{State: "blocked", Facts: []byte(`{"state":"blocked"}`), Text: text, Message: telegramOutcomeMessage("43", text), ConfigDigest: config.Digest()}, time.Now())
}

func (r telegramReceiver) prepareGuidedAction(ctx context.Context, claim core.TelegramActionClaim) (telegramGuidedInput, core.TelegramReceiverConfig, error) {
	var input telegramGuidedInput

	current, err := r.validateAuthoringAuthority(ctx, claim.Request, claim.Request.State == "applying")
	if err != nil {
		return input, current, err
	}

	input, err = telegramRetainedGuidedInput(claim.Request)
	if err != nil {
		return input, current, err
	}

	if !input.Offer || !slices.Contains(input.Choices, claim.Choice) || input.Command.Action != "" {
		return input, current, core.ErrWorkflowConflict
	}

	if err := r.validateGuidedRunBinding(ctx, claim.Request, current, input.Command.RunID); err != nil {
		return input, current, err
	}

	err = r.store.BeginTelegramOperation(ctx, claim.Request, time.Now())

	return input, current, err
}
