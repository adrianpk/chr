package host

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/harness"
)

const generationRecordLimit = 16 * 1024

// GenerationOperation is an immutable host-owned operation identity.
// The private executable path is retained locally and omitted from projections.
type GenerationOperation struct {
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

// GenerationLaunch retains only private host-resolved launch references.
type GenerationLaunch struct {
	Program   string
	Resources harness.GenerationResources
}

// GenerationPreparation pairs admitted public identity and private launch
// references from one coherent selection. It contains no operation context.
type GenerationPreparation struct {
	Binding harnesscore.GenerationBinding
	Launch  GenerationLaunch
	// Admit checks the actual configuration image after operation-value validation.
	Admit func(harnesscore.Selection) error
}

// GenerationLocation separates an existing host's artifact owner from
// the selected project's configuration and the host-owned execution directory.
// Empty directories use the artifact owner, preserving project-local calls.
type GenerationLocation struct {
	ConfigDir    string
	ExecutionDir string
	ProjectRef   string
}

// GenerationResolver returns a prepared launch and configuration admission.
type GenerationResolver func(context.Context, string) (GenerationPreparation, error)

type generationOperationFile struct {
	Schema       string                       `json:"schema"`
	Operation    GenerationOperation          `json:"operation"`
	Program      string                       `json:"program,omitempty"`
	Resources    *harness.GenerationResources `json:"resources,omitempty"`
	ConfigDir    string                       `json:"config_dir"`
	ExecutionDir string                       `json:"execution_dir"`
	Checksum     string                       `json:"checksum"`
}

// GenerationFreezeInput identifies one host-owned freeze operation and
// its optional routing. Locations retains the original zero-or-one validation.
type GenerationFreezeInput struct {
	Project   string
	ID        string
	Purpose   string
	Locations []GenerationLocation
}

// FreezeGeneration resolves a new operation once. Retrying an existing
// operation reads its artifact before consulting current profiles or settings.
// This artifact retains identity only; callers still own retries and budgets.
func FreezeGeneration(ctx context.Context, input GenerationFreezeInput, resolve GenerationResolver) (GenerationOperation, error) {
	project, id, purpose, locations := input.Project, input.ID, input.Purpose, input.Locations

	var operation GenerationOperation
	if project == "" {
		return operation, errors.New("generation requires bounded operation and purpose identities and a project")
	}

	root, err := filepath.Abs(project)
	if err != nil {
		return operation, err
	}

	if retained, ok := ctx.Value(generationContextKey{}).(GenerationOperation); ok {
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

		operation, err = freezeGenerationLocked(ctx, input, resolve, location)

		return err
	})

	return operation, err
}

func freezeGenerationLocked(ctx context.Context, input GenerationFreezeInput, resolve GenerationResolver, location GenerationLocation) (GenerationOperation, error) {
	project, id, purpose := input.Project, input.ID, input.Purpose

	operation, err := readGenerationOperation(project, id, purpose)
	if !errors.Is(err, os.ErrNotExist) {
		return operation, err
	}

	if resolve == nil {
		return operation, errors.New("generation selection resolver is required for a new operation")
	}

	if location.ConfigDir == project {
		return prepareGenerationOperation(ctx, input, resolve, location)
	}

	err = workspacecore.WithConfigLock(ctx, location.ConfigDir, func() error {
		var prepareErr error

		operation, prepareErr = prepareGenerationOperation(ctx, input, resolve, location)

		return prepareErr
	})

	return operation, err
}

func prepareGenerationOperation(ctx context.Context, input GenerationFreezeInput, resolve GenerationResolver, location GenerationLocation) (GenerationOperation, error) {
	project, id, purpose := input.Project, input.ID, input.Purpose

	var operation GenerationOperation

	prepared, err := resolve(ctx, location.ConfigDir)
	if err != nil {
		return operation, err
	}

	binding, launch := prepared.Binding, prepared.Launch

	operation = GenerationOperation{ID: id, Purpose: purpose, ProjectRef: location.ProjectRef, Binding: binding, program: launch.Program, resources: launch.Resources, project: project, configDir: location.ConfigDir, executionDir: location.ExecutionDir}
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

	return operation, writeGenerationOperation(operation)
}

func generationOperationLocation(root string, locations []GenerationLocation) (GenerationLocation, error) {
	location := GenerationLocation{ConfigDir: root, ExecutionDir: root}
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

func validateGenerationOperation(operation GenerationOperation) error {
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
	return CheckCommandImage(harnesscore.Binding{ExecutableRef: operation.Binding.ExecutableRef}, []string{operation.program})
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

func checkGenerationExecutable(operation GenerationOperation) error {
	if operation.Binding.Selection.Backend == "openai" {
		return nil
	}

	resolved, err := filepath.EvalSymlinks(operation.program)
	if err != nil || resolved != operation.program {
		return errors.New("retained generation executable cannot be resolved")
	}

	hash, err := ExecutableHash(operation.program)
	if err != nil {
		return err
	}

	if hash != operation.Binding.ExecutableRef.Hash {
		return errors.New("retained generation executable changed; restore the selected executable")
	}

	return nil
}
