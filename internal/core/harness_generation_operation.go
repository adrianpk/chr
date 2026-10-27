package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/harness"
)

const harnessGenerationRecordLimit = 16 * 1024

// HarnessGenerationOperation is an immutable host-owned operation identity.
// The private executable path is retained locally and omitted from projections.
type HarnessGenerationOperation struct {
	ID           string                        `json:"id"`
	Purpose      string                        `json:"purpose"`
	ProjectRef   string                        `json:"project_ref,omitempty"`
	Binding      harnesscore.GenerationBinding `json:"binding"`
	program      string
	project      string
	configDir    string
	executionDir string
	resources    harness.GenerationResources
}

// HarnessGenerationLaunch retains only private host-resolved launch references.
type HarnessGenerationLaunch struct {
	Program   string
	Resources harness.GenerationResources
}

// HarnessGenerationLocation separates an existing host's artifact owner from
// the selected project's configuration and the host-owned execution directory.
// Empty directories use the artifact owner, preserving project-local calls.
type HarnessGenerationLocation struct {
	ConfigDir    string
	ExecutionDir string
	ProjectRef   string
}

// HarnessGenerationResolver prepares one complete selection under the config owner.
type HarnessGenerationResolver func(context.Context, WorkspaceConfig) (harnesscore.GenerationBinding, HarnessGenerationLaunch, error)

type harnessGenerationOperationFile struct {
	Schema       string                       `json:"schema"`
	Operation    HarnessGenerationOperation   `json:"operation"`
	Program      string                       `json:"program,omitempty"`
	Resources    *harness.GenerationResources `json:"resources,omitempty"`
	ConfigDir    string                       `json:"config_dir"`
	ExecutionDir string                       `json:"execution_dir"`
	Checksum     string                       `json:"checksum"`
}

// FreezeHarnessGeneration resolves a new operation once. Retrying an existing
// operation reads its artifact before consulting current profiles or settings.
// This artifact retains identity only; callers still own retries and budgets.
func FreezeHarnessGeneration(ctx context.Context, project, id, purpose string, resolve HarnessGenerationResolver, locations ...HarnessGenerationLocation) (HarnessGenerationOperation, error) {
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
			return operation, workflowstate.ErrConflict
		}

		return retained, validateGenerationOperation(retained)
	}

	if !authoringstate.ValidOperationIdentity(id) || !authoringstate.ValidOperationIdentity(purpose) {
		return operation, errors.New("generation requires bounded operation and purpose identities")
	}

	location, err := generationOperationLocation(root, locations)
	if err != nil {
		return operation, err
	}

	err = withWorkspaceConfigLock(ctx, root, func() error {
		var err error

		operation, err = freezeHarnessGenerationLocked(ctx, root, id, purpose, resolve, location)

		return err
	})

	return operation, err
}

func freezeHarnessGenerationLocked(ctx context.Context, project, id, purpose string, resolve HarnessGenerationResolver, location HarnessGenerationLocation) (HarnessGenerationOperation, error) {
	operation, err := readHarnessGenerationOperation(project, id, purpose)
	if !errors.Is(err, os.ErrNotExist) {
		return operation, err
	}

	if resolve == nil {
		return operation, errors.New("generation selection resolver is required for a new operation")
	}

	if location.ConfigDir == project {
		return prepareHarnessGenerationOperation(ctx, project, id, purpose, resolve, location)
	}

	err = withWorkspaceConfigLock(ctx, location.ConfigDir, func() error {
		var prepareErr error

		operation, prepareErr = prepareHarnessGenerationOperation(ctx, project, id, purpose, resolve, location)

		return prepareErr
	})

	return operation, err
}

func prepareHarnessGenerationOperation(ctx context.Context, project, id, purpose string, resolve HarnessGenerationResolver, location HarnessGenerationLocation) (HarnessGenerationOperation, error) {
	var operation HarnessGenerationOperation

	config, err := LoadWorkspaceConfig(location.ConfigDir)
	if err != nil {
		return operation, err
	}

	binding, launch, err := resolve(ctx, config)
	if err != nil {
		return operation, err
	}

	operation = HarnessGenerationOperation{ID: id, Purpose: purpose, ProjectRef: location.ProjectRef, Binding: binding, program: launch.Program, resources: launch.Resources, project: project, configDir: location.ConfigDir, executionDir: location.ExecutionDir}
	if err := validateGenerationOperation(operation); err != nil {
		return operation, err
	}

	if binding.Selection.SelectionRevision != config.Revision() || binding.Selection.Profile != config.Harnesses.Generation || binding.Selection.Action != purpose {
		return operation, workflowstate.ErrConflict
	}

	if err := checkGenerationExecutable(operation); err != nil {
		return operation, err
	}

	return operation, writeHarnessGenerationOperation(operation)
}

func generationOperationLocation(root string, locations []HarnessGenerationLocation) (HarnessGenerationLocation, error) {
	location := HarnessGenerationLocation{ConfigDir: root, ExecutionDir: root}
	if len(locations) > 1 {
		return location, errors.New("generation accepts one host-owned location")
	}

	if len(locations) == 1 {
		location.ProjectRef = locations[0].ProjectRef
		if locations[0].ConfigDir != "" {
			location.ConfigDir = locations[0].ConfigDir
		}

		if locations[0].ExecutionDir != "" {
			location.ExecutionDir = locations[0].ExecutionDir
		}
	}

	var err error

	location.ConfigDir, err = filepath.Abs(location.ConfigDir)
	if err != nil {
		return location, err
	}

	location.ExecutionDir, err = filepath.Abs(location.ExecutionDir)

	return location, err
}

func validateGenerationOperation(operation HarnessGenerationOperation) error {
	if err := harnesscore.ValidateGenerationBinding(operation.Binding); err != nil {
		return err
	}

	if operation.Binding.Selection.Action != operation.Purpose {
		return workflowstate.ErrConflict
	}

	if !filepath.IsAbs(operation.configDir) || !filepath.IsAbs(operation.executionDir) || len(operation.ProjectRef) > 256 {
		return errors.New("generation operation requires retained host locations")
	}

	if operation.Binding.Selection.Backend == "openai" {
		if operation.program != "" || operation.resources != (harness.GenerationResources{}) {
			return workflowstate.ErrConflict
		}

		return nil
	}

	if err := validateGenerationResources(operation.resources, operation.Binding.Selection.Backend); err != nil {
		return err
	}

	// The generic command-image check does no worker-role or live-profile IO.
	return checkHarnessCommandImage(harnesscore.Binding{ExecutableRef: operation.Binding.ExecutableRef}, []string{operation.program})
}

func validateGenerationResources(resources harness.GenerationResources, backend string) error {
	if backend != "pi" && resources != (harness.GenerationResources{}) {
		return errors.New("generation resources require a supported adapter")
	}

	for _, path := range []string{resources.CredentialFile, resources.ModelConfigFile} {
		if !validGenerationResourcePath(path) {
			return errors.New("generation resources require bounded private absolute references")
		}
	}

	if resources.ModelConfigFile == "" {
		if resources.ModelConfigHash != "" {
			return errors.New("generation model resource requires a file reference")
		}

		return nil
	}

	if !harnesscore.SelectionDigest(resources.ModelConfigHash) {
		return errors.New("generation model resource requires identity evidence")
	}

	return nil
}

func validGenerationResourcePath(path string) bool {
	return path == "" || (filepath.IsAbs(path) && len(path) <= 4096 && !strings.ContainsAny(path, "\x00\r\n"))
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
