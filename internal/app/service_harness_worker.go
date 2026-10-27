package app

import (
	"context"
	"errors"
	"slices"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/harness/codex"
	"forge.adrianpk.com/chr/internal/harness/pi"
)

func (s service) prepareWorkflowHarness(ctx context.Context, config core.WorkspaceConfig) (core.WorkflowHarnessResolution, error) {
	if _, err := config.HarnessSelections(s.projectContext.ProjectID); err != nil {
		return core.WorkflowHarnessResolution{}, err
	}

	if config.Harnesses.Worker == "" {
		return core.WorkflowHarnessResolution{}, nil
	}

	resolved, err := s.resolveHarnessSelection(ctx, config, core.HarnessWorker, "")
	if err != nil {
		return core.WorkflowHarnessResolution{}, err
	}

	worker, runtimeRef, protocol, err := s.selectedWorkerCommand(ctx, resolved)
	if err != nil {
		return core.WorkflowHarnessResolution{}, err
	}

	executable, command, err := core.FreezeHarnessExecutable(worker)
	if err != nil {
		return core.WorkflowHarnessResolution{}, err
	}

	if s.workflowWorker != nil && !slices.Equal(s.workflowWorker, command) {
		return core.WorkflowHarnessResolution{}, workflowstate.ErrConflict
	}

	binding, err := core.SealHarnessBinding(core.HarnessBinding{
		SchemaVersion: 1, Role: core.HarnessWorker, Harness: resolved.View.Harness, Source: resolved.View.Source,
		SelectionRevision: resolved.View.SelectionRevision, Profile: resolved.View.Profile, ProfileRevision: resolved.View.ProfileRevision,
		ModelProvider: harnessWorkerProvider(resolved.View.ModelProvider), Model: resolved.View.Model, Thinking: resolved.View.Thinking,
		ExecutableRef: executable, RuntimeRef: runtimeRef, ProtocolRevision: protocol, Capabilities: resolved.View.Capabilities,
	})

	return core.WorkflowHarnessResolution{Binding: &binding, Command: command}, err
}

func harnessWorkerProvider(provider string) string {
	if provider == "" {
		return "cli_default"
	}

	return provider
}

func (s service) selectedWorkerCommand(ctx context.Context, resolved resolvedHarnessSelection) ([]string, string, string, error) {
	if resolved.View.Harness == codex.Identity {
		worker := (codex.Runner{WorkerProgram: resolved.Executable, WorkerModel: resolved.Profile.Model}).WorkflowWorker()

		return worker, "native", codex.WorkerProtocol, nil
	}

	if resolved.View.Harness != pi.Identity {
		return nil, "", "", errors.New("selected worker harness adapter is unavailable")
	}

	resources, err := pi.ResolveWorkerResources(ctx, s.env.ConfigDir(), resolved.Profile.CredentialRef, resolved.Profile.ModelConfigRef, resolved.Profile.ModelProvider)
	if err != nil {
		return nil, "", "", err
	}

	command, err := pi.WorkerCommand(resolved.Executable, pi.WorkerRuntimeOptions{
		Image: resolved.RuntimeImage, Provider: resolved.Profile.ModelProvider, Model: resolved.Profile.Model, Thinking: resolved.Profile.Thinking,
		CredentialFile: resources.CredentialFile, ModelFile: resources.ModelConfigFile, ModelHash: resources.ModelConfigHash, OAuth: resources.OAuth,
	})

	return command, resolved.RuntimeImage, pi.ProtocolRevision, err
}
