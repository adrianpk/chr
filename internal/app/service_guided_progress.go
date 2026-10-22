package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"

	"forge.adrianpk.com/chr/internal/core"
)

func (s service) applyGuidedCommand(ctx context.Context, command core.GuidedUseCaseCommand) (core.AuthoringSession, error) {
	switch command.Action {
	case "finish":
		return core.FinishGuidedUseCaseDump(ctx, s.projectContext.ProjectDir, command)
	case "quiz", "answer":
		return s.ConductGuidedQuiz(ctx, command)
	case "spec":
		return s.GenerateGuidedSpec(ctx, command)
	case "prepare":
		return s.PrepareGuidedWork(ctx, command)
	case "admit":
		return s.AdmitGuidedWork(ctx, command)
	default:
		return core.MutateGuidedUseCase(ctx, s.projectContext.ProjectDir, command)
	}
}

func (s service) advanceGuidedUseCase(ctx context.Context, session core.AuthoringSession) (core.AuthoringSession, error) {
	for range 6 {
		command, ready := guidedProgressCommand(session)
		if !ready {
			return session, nil
		}

		var err error

		session, err = s.applyGuidedCommand(ctx, command)
		if err != nil {
			return session, err
		}
	}

	return session, errors.New("guided authoring progression reached its bounded step limit")
}

func guidedProgressCommand(session core.AuthoringSession) (core.GuidedUseCaseCommand, bool) {
	command := core.GuidedUseCaseCommand{GuidedUseCaseScope: core.GuidedUseCaseScope{RunID: session.ID, ProjectID: session.ProjectID, Namespace: session.Namespace}, ExpectedRevision: session.Revision}
	if !guidedDispatchAllowed(session.Guided.Status) {
		return command, false
	}

	if pending := session.Guided.PendingOperation; pending != nil {
		command.OperationID, command.ExpectedRevision = pending.ID, pending.ExpectedRevision
		command.Action, command.Content = pending.Action, pending.Content
		command.QuestionID, command.InputDigest = pending.QuestionID, pending.InputDigest

		return command, true
	}

	var ready bool

	command.Action, ready = guidedAutomaticAction(session)
	if !ready {
		return command, false
	}

	digest := sha256.Sum256([]byte(session.ID + "\x00" + session.Namespace + "\x00" + command.Action + "\x00" + strconv.FormatInt(session.Revision, 10)))
	command.OperationID = "auto-" + command.Action + "-" + hex.EncodeToString(digest[:])

	return command, true
}

func guidedDispatchAllowed(status string) bool {
	return status != "paused" && status != "cancelled" && status != "completed" && status != "failed"
}

func guidedAutomaticAction(session core.AuthoringSession) (string, bool) {
	switch session.Guided.Stage {
	case "quiz":
		if session.Guided.Quiz != nil {
			return "", false
		}

		return "quiz", true
	case "specification":
		return "spec", true
	case "preparation":
		if _, err := core.GuidedPreparationCheckpoint(session); err == nil {
			return "admit", true
		}

		return "prepare", true
	default:
		return "", false
	}
}
