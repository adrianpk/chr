package app

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

func (r telegramReceiver) continueTelegramGuided(ctx context.Context, request core.TelegramOperationRequest, config core.TelegramReceiverConfig) error {
	if request.State == "confirmed" || request.State == "blocked" {
		return nil
	}

	operations, ok := r.operations.(telegramGuidedService)
	if !ok {
		return ErrRecordUnsupported
	}

	current, err := r.validateAuthoringAuthority(ctx, request, request.State == "applying")
	if err != nil {
		return r.finishAuthoringFailure(ctx, request, config, err)
	}

	owner, err := telegramRetainedGuidedInput(request)
	if err != nil {
		return err
	}

	if err := r.validateGuidedRunBinding(ctx, request, current, owner.Command.RunID); err != nil {
		return r.finishAuthoringFailure(ctx, request, current, err)
	}

	if err := r.store.BeginTelegramOperation(ctx, request, time.Now()); err != nil {
		return err
	}

	session, err := operations.ApplyTelegramGuided(ctx, request.Target, owner)
	if errors.Is(err, core.ErrWorkflowConflict) {
		return r.finishAuthoringFailure(ctx, request, current, err)
	}

	if err != nil {
		return err
	}

	return r.finishTelegramGuided(ctx, request, session, current)
}

func (r telegramReceiver) finishTelegramGuided(ctx context.Context, request core.TelegramOperationRequest, session core.AuthoringSession, current core.TelegramReceiverConfig) error {
	progress, err := telegramGuidedProgress(session, core.DirectoryName+"/authoring/sessions/"+session.ID+"/session.md")
	if err != nil {
		return err
	}

	facts, err := json.Marshal(map[string]any{"session": session.ID, "revision": session.Revision, "stage": session.Guided.Stage})
	if err != nil {
		return err
	}

	return r.store.FinishTelegramGuided(ctx, request, progress, facts, telegramGuidedPrompt(session), current.Digest(), time.Now())
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

	return owner.Command.OperationID == request.ID
}

func telegramGuidedPrompt(session core.AuthoringSession) string {
	text := "Guided use case: " + session.Topic + "\nStage: " + session.Guided.Stage + " · " + session.Guided.Status
	switch session.Guided.Stage {
	case "name":
		return text + "\n\nWhat is the use-case name? Reply here."
	case "dump":
		return text + "\n\nAdd everything known about this use case. Reply Finish capture when the context is complete."
	default:
		return text + "\n\nSaved input is retained. The shared workflow owns the next step."
	}
}
