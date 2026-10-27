package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
)

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
