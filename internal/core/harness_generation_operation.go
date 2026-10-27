package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
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

// HarnessGenerationPreparation pairs admitted public identity and private launch
// references from one coherent selection. It contains no operation context.
type HarnessGenerationPreparation struct {
	Binding harnesscore.GenerationBinding
	Launch  HarnessGenerationLaunch
	// Admit checks the actual configuration image after operation-value validation.
	Admit func(harnesscore.Selection) error
}

// HarnessGenerationLocation separates an existing host's artifact owner from
// the selected project's configuration and the host-owned execution directory.
// Empty directories use the artifact owner, preserving project-local calls.
type HarnessGenerationLocation struct {
	ConfigDir    string
	ExecutionDir string
	ProjectRef   string
}

// harnessGenerationOperationResolver returns a prepared launch and configuration admission.
type harnessGenerationOperationResolver func(context.Context, string) (HarnessGenerationPreparation, error)

type harnessGenerationOperationFile struct {
	Schema       string                       `json:"schema"`
	Operation    HarnessGenerationOperation   `json:"operation"`
	Program      string                       `json:"program,omitempty"`
	Resources    *harness.GenerationResources `json:"resources,omitempty"`
	ConfigDir    string                       `json:"config_dir"`
	ExecutionDir string                       `json:"execution_dir"`
	Checksum     string                       `json:"checksum"`
}

// HarnessGenerationFreezeInput identifies one host-owned freeze operation and
// its optional routing. Locations retains the original zero-or-one validation.
type HarnessGenerationFreezeInput struct {
	Project   string
	ID        string
	Purpose   string
	Locations []HarnessGenerationLocation
}

// FreezeHarnessGeneration resolves a new operation once. Retrying an existing
// operation reads its artifact before consulting current profiles or settings.
// This artifact retains identity only; callers still own retries and budgets.
func freezeHarnessGenerationOperation(ctx context.Context, input HarnessGenerationFreezeInput, resolve harnessGenerationOperationResolver) (HarnessGenerationOperation, error) {
	project, id, purpose, locations := input.Project, input.ID, input.Purpose, input.Locations

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

	input.Project = root

	err = workspacecore.WithConfigLock(ctx, root, func() error {
		var err error

		operation, err = freezeHarnessGenerationLocked(ctx, input, resolve, location)

		return err
	})

	return operation, err
}

func freezeHarnessGenerationLocked(ctx context.Context, input HarnessGenerationFreezeInput, resolve harnessGenerationOperationResolver, location HarnessGenerationLocation) (HarnessGenerationOperation, error) {
	project, id, purpose := input.Project, input.ID, input.Purpose

	operation, err := readHarnessGenerationOperation(project, id, purpose)
	if !errors.Is(err, os.ErrNotExist) {
		return operation, err
	}

	if resolve == nil {
		return operation, errors.New("generation selection resolver is required for a new operation")
	}

	if location.ConfigDir == project {
		return prepareHarnessGenerationOperation(ctx, input, resolve, location)
	}

	err = workspacecore.WithConfigLock(ctx, location.ConfigDir, func() error {
		var prepareErr error

		operation, prepareErr = prepareHarnessGenerationOperation(ctx, input, resolve, location)

		return prepareErr
	})

	return operation, err
}

func prepareHarnessGenerationOperation(ctx context.Context, input HarnessGenerationFreezeInput, resolve harnessGenerationOperationResolver, location HarnessGenerationLocation) (HarnessGenerationOperation, error) {
	project, id, purpose := input.Project, input.ID, input.Purpose

	var operation HarnessGenerationOperation

	prepared, err := resolve(ctx, location.ConfigDir)
	if err != nil {
		return operation, err
	}

	binding, launch := prepared.Binding, prepared.Launch

	operation = HarnessGenerationOperation{ID: id, Purpose: purpose, ProjectRef: location.ProjectRef, Binding: binding, program: launch.Program, resources: launch.Resources, project: project, configDir: location.ConfigDir, executionDir: location.ExecutionDir}
	if err := validateGenerationOperation(operation); err != nil {
		return operation, err
	}

	if prepared.Admit == nil {
		return operation, errors.New("generation configuration admission is required")
	}

	if err := prepared.Admit(binding.Selection); err != nil {
		return operation, err
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
	return harnesshost.CheckCommandImage(harnesscore.Binding{ExecutableRef: operation.Binding.ExecutableRef}, []string{operation.program})
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

	hash, err := harnesshost.ExecutableHash(operation.program)
	if err != nil {
		return err
	}

	if hash != operation.Binding.ExecutableRef.Hash {
		return errors.New("retained generation executable changed; restore the selected executable")
	}

	return nil
}
