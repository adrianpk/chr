package authoring

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
)

func parseGuidedUseCase(value string) (*GuidedUseCaseState, error) {
	if value == "" {
		return nil, nil
	}

	var state GuidedUseCaseState
	if err := authoringstate.Decode([]byte(opsmetadata.Unquote(value)), &state); err != nil {
		return nil, fmt.Errorf("parse guided state: %w", err)
	}

	return &state, nil
}

func guidedOperation(command GuidedUseCaseCommand) (GuidedUseCaseOperation, error) {
	if err := authoringstate.ValidateSessionID(command.OperationID); err != nil {
		return GuidedUseCaseOperation{}, err
	}

	if len(command.Content) > 65536 || command.ExpectedRevision < 1 {
		return GuidedUseCaseOperation{}, errors.New("guided command requires bounded content and a positive revision")
	}

	encoded, err := json.Marshal(command)

	return GuidedUseCaseOperation{ID: command.OperationID, Action: command.Action, NativeRevision: command.NativeRevision, QuestionID: command.QuestionID, InputDigest: command.InputDigest, ExpectedRevision: command.ExpectedRevision, Digest: contenthash.Sum(encoded), Content: command.Content}, err
}

func renderGuidedState(output *bytes.Buffer, state *GuidedUseCaseState) error {
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
