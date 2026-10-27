package core

import (
	"context"

	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// HarnessGenerationResolver prepares one selection from the actual configuration image.
type HarnessGenerationResolver func(context.Context, WorkspaceConfig) (HarnessGenerationPreparation, error)

// FreezeHarnessGeneration composes configuration loading and admission with the
// host's retained operation. Recovery reads the retained image before loading current settings.
func FreezeHarnessGeneration(ctx context.Context, input HarnessGenerationFreezeInput, resolve HarnessGenerationResolver) (HarnessGenerationOperation, error) {
	if resolve == nil {
		return freezeHarnessGenerationOperation(ctx, input, nil)
	}

	prepare := func(ctx context.Context, configDir string) (HarnessGenerationPreparation, error) {
		config, err := LoadWorkspaceConfig(configDir)
		if err != nil {
			return HarnessGenerationPreparation{}, err
		}

		prepared, err := resolve(ctx, config)
		if err != nil {
			return prepared, err
		}

		prepared.Admit = func(selection harnesscore.Selection) error {
			if selection.SelectionRevision != config.Revision() || selection.Profile != config.Harnesses.Generation || selection.Action != input.Purpose {
				return workflowstate.ErrConflict
			}

			return nil
		}

		return prepared, nil
	}

	return freezeHarnessGenerationOperation(ctx, input, prepare)
}
