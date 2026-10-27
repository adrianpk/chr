package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// An explicit new-project instruction can answer the project-choice diagnostic,
// which intentionally carries no authority-bearing conversation or action reply.
func (r telegramReceiver) respondGuidedStart(ctx context.Context, user telegramcore.AllowedUser, config telegramcore.ReceiverConfig, update telegramcore.ReceivedUpdate) (bool, error) {
	if update.Kind != "message" || !telegramGuidedStart(update.Text) {
		return false, nil
	}

	return r.respondGuided(ctx, user, config, update)
}

func (r telegramReceiver) respondGuided(ctx context.Context, user telegramcore.AllowedUser, config telegramcore.ReceiverConfig, update telegramcore.ReceivedUpdate) (bool, error) {
	operationLookup, err := r.store.Operation(ctx, update, 0)
	request, found := operationLookup.Request, operationLookup.Found

	if err != nil {
		return true, err
	}

	if found && request.OperationID == "guided.step" {
		return true, r.offerGuidedControls(ctx, request, config)
	}

	if found && request.OperationID == "guided.continue" {
		return true, r.continueTelegramGuided(ctx, request, config)
	}

	if update.Kind != "message" || !telegramGuidedMessage(update.Text) {
		return false, nil
	}

	operations, supported := r.operations.(telegramGuidedService)
	if !supported {
		return false, nil
	}

	return r.routeTelegramGuided(ctx, operations, user, config, update)
}

func (r telegramReceiver) routeTelegramGuided(ctx context.Context, operations telegramGuidedService, user telegramcore.AllowedUser, config telegramcore.ReceiverConfig, update telegramcore.ReceivedUpdate) (bool, error) {
	guidedRouting, err := r.resolveTelegramGuidedInput(ctx, user, config, update)
	input, start := guidedRouting.Input, guidedRouting.Start

	if err != nil {
		return true, r.guidedDiagnostic(ctx, update, config, err)
	}

	projects := input.Projects

	if !start {
		present, err := r.hasGuidedConversation(ctx, update.Peer)
		if err != nil || !present {
			return false, err
		}
	}

	request, err := r.reserveTelegramGuided(ctx, operations, user, input, start)
	if errors.Is(err, telegramcore.ErrNotGuided) {
		return false, nil
	}

	if err != nil {
		if start {
			err = telegramGuidedStartFailure(err, projects)
		}

		return true, r.guidedDiagnostic(ctx, update, config, err)
	}

	if request.OperationID == "guided.step" {
		return true, r.offerGuidedControls(ctx, request, config)
	}

	return true, r.continueTelegramGuided(ctx, request, config)
}

func (r telegramReceiver) hasGuidedConversation(ctx context.Context, peer telegramcore.PeerIdentity) (bool, error) {
	sessions, err := r.store.Conversations(ctx, peer)
	if err != nil {
		return false, err
	}

	for _, c := range sessions {
		if c.GuidedRunID != "" {
			return true, nil
		}
	}

	return false, nil
}

func (r telegramReceiver) reserveTelegramGuided(ctx context.Context, operations telegramGuidedService, user telegramcore.AllowedUser, input telegramcore.ConversationInput, start bool) (telegramcore.OperationRequest, error) {
	var (
		c      telegramcore.Conversation
		target telegramcore.OperationTarget
		err    error
	)

	if start {
		bindingResult, bindingResultError := r.store.ReserveGuidedConversation(ctx, input, time.Now())
		c, target, err = bindingResult.Conversation, bindingResult.Target, bindingResultError
	} else {
		bindingResult2, bindingResult2Error := r.store.ResolveGuidedTarget(ctx, input)
		c, target, err = bindingResult2.Conversation, bindingResult2.Target, bindingResult2Error
	}

	if err != nil {
		return telegramcore.OperationRequest{}, err
	}

	owner, err := telegramGuidedOwner(c, input, start, user.OperatorID)
	if err != nil {
		return telegramcore.OperationRequest{}, err
	}

	if !start {
		owner, err = prepareTelegramGuidedInput(ctx, operations, target, owner)
		if err != nil {
			return telegramcore.OperationRequest{}, err
		}
	}

	data, err := json.Marshal(owner)
	if err != nil {
		return telegramcore.OperationRequest{}, err
	}

	args, err := json.Marshal(telegramAuthoringArguments{PayloadHash: recordDigest(data)})
	if err != nil {
		return telegramcore.OperationRequest{}, err
	}

	operation := "guided.continue"
	if owner.Offer {
		operation = "guided.step"
	}

	return r.store.ReservePreparedOperation(ctx, telegramcore.OperationRequest{Source: input.Update, Operator: user.OperatorID, Target: target, OperationID: operation, Arguments: args}, data, time.Now())
}

func (r telegramReceiver) guidedDiagnostic(ctx context.Context, update telegramcore.ReceivedUpdate, config telegramcore.ReceiverConfig, failure error) error {
	text := "The guided use case changed or is waiting for its current owner. Reopen its current prompt; saved input is unchanged."

	var routing telegramGuidedRoutingFailure
	if errors.As(failure, &routing) {
		text = string(routing)
	}

	if safe, ok := telegramcore.ConversationFailure(failure); ok {
		text = safe
	}

	return r.recordDiagnostic(ctx, update, config, text)
}

func telegramGuidedMessage(text string) bool {
	word := telegramCommandWord(strings.TrimSpace(text))

	name, body := word.Command, word.Body
	if !strings.HasPrefix(name, "/") {
		return true
	}

	return body == "" && (name == "/cancel" || name == "/pause" || name == "/resume" || name == "/done")
}

func (r telegramReceiver) validateGuidedRunBinding(ctx context.Context, request telegramcore.OperationRequest, config telegramcore.ReceiverConfig, id string) error {
	user, _ := config.AuthorizedUser(request.Source.Peer.ChatID, request.Source.Peer.TopicID, request.Source.Peer.UserID, request.Source.ChatType)

	projects, err := r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return err
	}

	c, err := r.store.ReadConversationBinding(ctx, telegramcore.ConversationInput{Update: request.Source, Config: config, Projects: projects}, request.Target)
	if err != nil {
		return err
	}

	if c.GuidedRunID != id {
		return workflowstate.ErrConflict
	}

	bound, err := telegramAuthoringSessionID(c.AuthoringSessionRef)
	if err != nil || bound != id {
		return workflowstate.ErrConflict
	}

	return nil
}
