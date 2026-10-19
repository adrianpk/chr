package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const harnessGenerationRecordLimit = 16 * 1024

// HarnessGenerationOperation is an immutable host-owned operation identity.
// The private executable path is retained locally and omitted from projections.
type HarnessGenerationOperation struct {
	ID      string                   `json:"id"`
	Purpose string                   `json:"purpose"`
	Binding HarnessGenerationBinding `json:"binding"`
	program string
	project string
}

// HarnessGenerationResolver prepares one complete selection under the config owner.
type HarnessGenerationResolver func(context.Context, WorkspaceConfig) (HarnessGenerationBinding, string, error)

type harnessGenerationOperationFile struct {
	Schema    string                     `json:"schema"`
	Operation HarnessGenerationOperation `json:"operation"`
	Program   string                     `json:"program,omitempty"`
	Checksum  string                     `json:"checksum"`
}

// FreezeHarnessGeneration resolves a new operation once. Retrying an existing
// operation reads its artifact before consulting current profiles or settings.
// This artifact retains identity only; callers still own retries and budgets.
func FreezeHarnessGeneration(ctx context.Context, project, id, purpose string, resolve HarnessGenerationResolver) (HarnessGenerationOperation, error) {
	var operation HarnessGenerationOperation
	if project == "" {
		return operation, errors.New("generation requires bounded operation and purpose identities and a project")
	}

	root, err := filepath.Abs(project)
	if err != nil {
		return operation, err
	}

	if retained, ok := ctx.Value(harnessGenerationContextKey{}).(HarnessGenerationOperation); ok {
		if retained.project != root {
			return operation, ErrWorkflowConflict
		}

		return retained, validateGenerationOperation(retained)
	}

	if !validGenerationOperationIdentity(id) || !validGenerationOperationIdentity(purpose) {
		return operation, errors.New("generation requires bounded operation and purpose identities")
	}

	err = withWorkspaceConfigLock(ctx, root, func() error {
		var err error

		operation, err = freezeHarnessGenerationLocked(ctx, root, id, purpose, resolve)

		return err
	})

	return operation, err
}

func freezeHarnessGenerationLocked(ctx context.Context, project, id, purpose string, resolve HarnessGenerationResolver) (HarnessGenerationOperation, error) {
	operation, err := readHarnessGenerationOperation(project, id, purpose)
	if !errors.Is(err, os.ErrNotExist) {
		return operation, err
	}

	if resolve == nil {
		return operation, errors.New("generation selection resolver is required for a new operation")
	}

	config, err := LoadWorkspaceConfig(project)
	if err != nil {
		return operation, err
	}

	binding, program, err := resolve(ctx, config)
	if err != nil {
		return operation, err
	}

	operation = HarnessGenerationOperation{ID: id, Purpose: purpose, Binding: binding, program: program, project: project}
	if err := validateGenerationOperation(operation); err != nil {
		return operation, err
	}

	if binding.Selection.SelectionRevision != config.Revision() || binding.Selection.Profile != config.Harnesses.Generation || binding.Selection.Action != purpose {
		return operation, ErrWorkflowConflict
	}

	if err := checkGenerationExecutable(operation); err != nil {
		return operation, err
	}

	return operation, writeHarnessGenerationOperation(operation)
}

func validGenerationOperationIdentity(value string) bool {
	return strings.TrimSpace(value) != "" && len(value) <= 256 && !strings.ContainsAny(value, "\x00\r\n")
}

func validateGenerationOperation(operation HarnessGenerationOperation) error {
	if err := ValidateHarnessGenerationBinding(operation.Binding); err != nil {
		return err
	}

	if operation.Binding.Selection.Action != operation.Purpose {
		return ErrWorkflowConflict
	}

	if operation.Binding.Selection.Backend == "openai" {
		if operation.program != "" {
			return ErrWorkflowConflict
		}

		return nil
	}

	// The generic command-image check does no worker-role or live-profile IO.
	return checkHarnessCommandImage(HarnessBinding{ExecutableRef: operation.Binding.ExecutableRef}, []string{operation.program})
}

func checkGenerationExecutable(operation HarnessGenerationOperation) error {
	if operation.Binding.Selection.Backend == "openai" {
		return nil
	}

	resolved, err := filepath.EvalSymlinks(operation.program)
	if err != nil || resolved != operation.program {
		return errors.New("retained generation executable cannot be resolved")
	}

	hash, err := harnessExecutableHash(operation.program)
	if err != nil {
		return err
	}

	if hash != operation.Binding.ExecutableRef.Hash {
		return errors.New("retained generation executable changed; restore the selected executable")
	}

	return nil
}
