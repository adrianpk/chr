package app

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (r telegramReceiver) offerGuidedControls(ctx context.Context, request telegramcore.OperationRequest, config telegramcore.ReceiverConfig) error {
	input, err := telegramRetainedGuidedInput(request)
	if err != nil {
		return err
	}

	if !input.Offer || request.OperationID != "guided.step" || strings.ToLower(strings.TrimSpace(request.Source.Text)) != "show controls" {
		return workflowstate.ErrConflict
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

	_, err = r.store.OfferAction(ctx, telegramcore.ActionOffer{
		Input: telegramcore.ConversationInput{Update: request.Source, Config: current, Projects: projects}, RequestID: request.ID, Choices: input.Choices,
		Message: telegramOutcomeMessage("43", "Choose an action for "+input.Name+". Saved inputs and existing work limits are retained. Send Show current step to return to the input prompt."), ExpiresAt: time.Now().Add(30 * time.Minute),
	}, time.Now())

	return err
}

func (r telegramReceiver) applyGuidedAction(ctx context.Context, claim telegramcore.ActionClaim, config telegramcore.ReceiverConfig) error {
	operations, ok := r.operations.(telegramGuidedService)
	if !ok {
		return ErrRecordUnsupported
	}

	preparedGuided, err := r.prepareGuidedAction(ctx, claim)
	input, current := preparedGuided.Input, preparedGuided.Config

	if errors.Is(err, workflowstate.ErrConflict) {
		return r.finishGuidedActionFailure(ctx, claim, config, err)
	}

	if err != nil {
		return err
	}

	input.Offer, input.Command.Action = false, claim.Choice

	operations = r.boundTelegramGuidedService(operations, claim.Request, input.Command.RunID)

	session, err := applyTelegramGuided(ctx, operations, claim.Request.Target, input)
	if errors.Is(err, workflowstate.ErrConflict) {
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

func (r telegramReceiver) finishGuidedAction(ctx context.Context, claim telegramcore.ActionClaim, config telegramcore.ReceiverConfig, observed GuidedUseCaseInspection) error {
	guidedOutcome, err := telegramGuidedResult(observed)
	progress, facts := guidedOutcome.Progress, guidedOutcome.Facts

	if err != nil {
		return err
	}

	text := telegramGuidedPrompt(observed)

	delivery, err := r.guidedDocument(ctx, claim.Request, observed, "")
	if err != nil {
		return err
	}

	result := telegramcore.ActionResult{State: telegramcore.OperationConfirmed, Delivery: delivery, Guided: &progress, ReceiptRef: progress.SessionRef, Facts: facts, Text: text, Message: telegramOutcomeMessage("43", text), ConfigDigest: config.Digest()}

	err = r.store.FinishAction(ctx, claim, result, time.Now())
	if errors.Is(err, workflowstate.ErrConflict) {
		return r.finishOvertakenGuidedAction(ctx, claim, result)
	}

	return err
}

func (r telegramReceiver) finishGuidedActionFailure(ctx context.Context, claim telegramcore.ActionClaim, config telegramcore.ReceiverConfig, failure error) error {
	text := telegramGuidedFailure(failure)

	return r.store.FinishAction(ctx, claim, telegramcore.ActionResult{State: telegramcore.OperationBlocked, Facts: []byte(`{"state":"blocked"}`), Text: text, Message: telegramOutcomeMessage("43", text), ConfigDigest: config.Digest()}, time.Now())
}

func (r telegramReceiver) prepareGuidedAction(ctx context.Context, claim telegramcore.ActionClaim) (telegramGuidedPreparation, error) {
	var input telegramGuidedInput

	current, err := r.validateAuthoringAuthority(ctx, claim.Request, claim.Request.State == telegramcore.OperationApplying)
	if err != nil {
		return telegramGuidedPreparation{Input: input, Config: current}, err
	}

	input, err = telegramRetainedGuidedInput(claim.Request)
	if err != nil {
		return telegramGuidedPreparation{Input: input, Config: current}, err
	}

	if !input.Offer || !slices.Contains(input.Choices, claim.Choice) || input.Command.Action != "" {
		return telegramGuidedPreparation{Input: input, Config: current}, workflowstate.ErrConflict
	}

	if err := r.validateGuidedRunBinding(ctx, claim.Request, current, input.Command.RunID); err != nil {
		return telegramGuidedPreparation{Input: input, Config: current}, err
	}

	err = r.store.BeginOperation(ctx, claim.Request, time.Now())

	return telegramGuidedPreparation{Input: input, Config: current}, err
}
