package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	opsnamespace "forge.adrianpk.com/chr/internal/core/ops/namespace"
)

// ValidateGuidedUseCase checks the existing complete guided session contract.
func ValidateGuidedUseCase(session authoringcore.AuthoringSession) error {
	state := session.Guided
	if state == nil {
		return nil
	}

	if !validGuidedIntent(session) {
		return errors.New("invalid guided use-case schema, scope or progression intent")
	}

	if !boundedGuidedState(state) {
		return errors.New("guided use-case state exceeds its bounds")
	}

	if !guidedStageValid(state.Stage) || !guidedStatusValid(state.Status) {
		return errors.New("invalid guided use-case stage or status")
	}

	if state.Stage != "name" && (state.CaseSlug == "" || state.CaseName == "") {
		return errors.New("guided use case requires a canonical name")
	}

	return nil
}

func guidedStageValid(value string) bool {
	switch value {
	case "name", "dump", "quiz", "specification", "preparation", "execution", "results":
		return true
	default:
		return false
	}
}

func guidedStatusValid(value string) bool {
	switch value {
	case "waiting_input", "running", "blocked", "paused", "completed", "cancelled", "failed":
		return true
	default:
		return false
	}
}

func parseGuidedUseCase(value string) (*authoringcore.GuidedUseCaseState, error) {
	if value == "" {
		return nil, nil
	}

	var state authoringcore.GuidedUseCaseState
	if err := authoringstate.Decode([]byte(opsmetadata.Unquote(value)), &state); err != nil {
		return nil, fmt.Errorf("parse guided state: %w", err)
	}

	return &state, nil
}

func guidedOperation(command authoringcore.GuidedUseCaseCommand) (authoringcore.GuidedUseCaseOperation, error) {
	if err := authoringstate.ValidateSessionID(command.OperationID); err != nil {
		return authoringcore.GuidedUseCaseOperation{}, err
	}

	if len(command.Content) > 65536 || command.ExpectedRevision < 1 {
		return authoringcore.GuidedUseCaseOperation{}, errors.New("guided command requires bounded content and a positive revision")
	}

	encoded, err := json.Marshal(command)

	return authoringcore.GuidedUseCaseOperation{ID: command.OperationID, Action: command.Action, NativeRevision: command.NativeRevision, QuestionID: command.QuestionID, InputDigest: command.InputDigest, ExpectedRevision: command.ExpectedRevision, Digest: contenthash.Sum(encoded), Content: command.Content}, err
}

func renderGuidedState(output *bytes.Buffer, state *authoringcore.GuidedUseCaseState) error {
	if state == nil {
		return nil
	}

	encoded, err := json.Marshal(state)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(output, "guided_state: %s\n", opsmetadata.Quote(string(encoded)))

	return err
}

func validGuidedIntent(session authoringcore.AuthoringSession) bool {
	state := session.Guided

	return state.SchemaVersion == 1 && session.Revision > 0 && opsnamespace.Valid(session.Namespace) && strings.TrimSpace(state.Actor) != "" && len(state.Actor) <= 128 && state.ProgressionIntent
}

func boundedGuidedState(state *authoringcore.GuidedUseCaseState) bool {
	return len(state.Capture) <= 65536 && len(state.CompletedOperations) <= 256 && len(state.Checkpoints) <= 8 && len(state.LaterIdeas) <= 64 && boundedGuidedControls(state.Control)
}
