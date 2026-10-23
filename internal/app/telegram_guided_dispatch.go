package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func (r telegramReceiver) respondGuided(ctx context.Context, user core.TelegramAllowedUser, config core.TelegramReceiverConfig, update core.TelegramReceivedUpdate) (bool, error) {
	request, found, err := r.store.TelegramOperation(ctx, update, 0)
	if err != nil {
		return true, err
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

func (r telegramReceiver) routeTelegramGuided(ctx context.Context, operations telegramGuidedService, user core.TelegramAllowedUser, config core.TelegramReceiverConfig, update core.TelegramReceivedUpdate) (bool, error) {
	input := core.TelegramConversationInput{Update: update, Config: config, Text: "New guided use case"}

	projects, err := r.operations.ListRegisteredProjects(user.ProjectIDs)
	input.Projects = projects

	if err != nil {
		return true, err
	}

	start := telegramGuidedStart(update.Text)
	if !start {
		present, err := r.hasGuidedConversation(ctx, update.Peer)
		if err != nil || !present {
			return false, err
		}
	}

	request, err := r.reserveTelegramGuided(ctx, operations, user, input, start)
	if errors.Is(err, core.ErrTelegramNotGuided) {
		return false, nil
	}

	if err != nil {
		return true, r.guidedDiagnostic(ctx, update, config, err)
	}

	return true, r.continueTelegramGuided(ctx, request, config)
}

func (r telegramReceiver) hasGuidedConversation(ctx context.Context, peer core.TelegramPeerIdentity) (bool, error) {
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

func (r telegramReceiver) reserveTelegramGuided(ctx context.Context, operations telegramGuidedService, user core.TelegramAllowedUser, input core.TelegramConversationInput, start bool) (core.TelegramOperationRequest, error) {
	var (
		c      core.TelegramConversation
		target core.TelegramOperationTarget
		err    error
	)
	if start {
		c, target, err = r.store.ReserveTelegramGuidedConversation(ctx, input, time.Now())
	} else {
		c, target, err = r.store.ResolveTelegramGuidedTarget(ctx, input)
	}

	if err != nil {
		return core.TelegramOperationRequest{}, err
	}

	owner := telegramGuidedInput{Actor: user.OperatorID, Start: start, Command: core.GuidedUseCaseCommand{GuidedUseCaseScope: core.GuidedUseCaseScope{RunID: c.ID, ProjectID: target.ProjectID, Namespace: target.Namespace}}}
	if !start {
		owner, err = telegramGuidedCommand(c, input.Update, user.OperatorID)
		if err != nil {
			return core.TelegramOperationRequest{}, err
		}

		if _, err := operations.InspectTelegramGuided(ctx, target, owner.Command.RunID); err != nil {
			return core.TelegramOperationRequest{}, err
		}
	}

	data, err := json.Marshal(owner)
	if err != nil {
		return core.TelegramOperationRequest{}, err
	}

	args, err := json.Marshal(telegramAuthoringArguments{PayloadHash: recordDigest(data)})
	if err != nil {
		return core.TelegramOperationRequest{}, err
	}

	return r.store.ReservePreparedTelegramOperation(ctx, core.TelegramOperationRequest{Source: input.Update, Operator: user.OperatorID, Target: target, OperationID: "guided.continue", Arguments: args}, data, time.Now())
}

func (r telegramReceiver) guidedDiagnostic(ctx context.Context, update core.TelegramReceivedUpdate, config core.TelegramReceiverConfig, failure error) error {
	text := "The guided use case changed or is waiting for its current owner. Reopen its current prompt; saved input is unchanged."
	if safe, ok := core.TelegramConversationFailure(failure); ok {
		text = safe
	}

	return r.recordDiagnostic(ctx, update, config, text)
}

func telegramGuidedMessage(text string) bool {
	name, body := telegramCommandWord(strings.TrimSpace(text))
	if !strings.HasPrefix(name, "/") {
		return true
	}

	return body == "" && (name == "/cancel" || name == "/pause" || name == "/resume" || name == "/done")
}

func (r telegramReceiver) validateGuidedRunBinding(ctx context.Context, request core.TelegramOperationRequest, config core.TelegramReceiverConfig, id string) error {
	user, _ := config.AuthorizedUser(request.Source.Peer.ChatID, request.Source.Peer.TopicID, request.Source.Peer.UserID, request.Source.ChatType)

	projects, err := r.operations.ListRegisteredProjects(user.ProjectIDs)
	if err != nil {
		return err
	}

	c, err := r.store.ReadTelegramConversationBinding(ctx, core.TelegramConversationInput{Update: request.Source, Config: config, Projects: projects}, request.Target)
	if err != nil {
		return err
	}

	if c.GuidedRunID != id {
		return core.ErrWorkflowConflict
	}

	bound, err := telegramAuthoringSessionID(c.AuthoringSessionRef)
	if err != nil || bound != id {
		return core.ErrWorkflowConflict
	}

	return nil
}
