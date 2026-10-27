package app

import (
	"context"
	"errors"
	"slices"

	"forge.adrianpk.com/chr/internal/core"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
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

	resolved, err := s.resolveHarnessSelection(ctx, config, harnesscore.Worker, "")
	if err != nil {
		return core.WorkflowHarnessResolution{}, err
	}

	selected, err := s.selectedWorkerCommand(ctx, resolved)
	if err != nil {
		return core.WorkflowHarnessResolution{}, err
	}

	executable, err := harnesshost.FreezeHarnessExecutable(selected.Command)
	if err != nil {
		return core.WorkflowHarnessResolution{}, err
	}

	if s.workflowWorker != nil && !slices.Equal(s.workflowWorker, executable.Command) {
		return core.WorkflowHarnessResolution{}, workflowstate.ErrConflict
	}

	binding, err := harnesscore.SealBinding(harnesscore.Binding{
		SchemaVersion: 1, Role: harnesscore.Worker, Harness: resolved.View.Harness, Source: resolved.View.Source,
		SelectionRevision: resolved.View.SelectionRevision, Profile: resolved.View.Profile, ProfileRevision: resolved.View.ProfileRevision,
		ModelProvider: harnessWorkerProvider(resolved.View.ModelProvider), Model: resolved.View.Model, Thinking: resolved.View.Thinking,
		ExecutableRef: executable.Reference, RuntimeRef: selected.RuntimeRef, ProtocolRevision: selected.ProtocolRevision, Capabilities: resolved.View.Capabilities,
	})

	return core.WorkflowHarnessResolution{Binding: &binding, Command: executable.Command}, err
}

func harnessWorkerProvider(provider string) string {
	if provider == "" {
		return "cli_default"
	}

	return provider
}

type workerCommandPlan struct {
	Command          []string
	RuntimeRef       string
	ProtocolRevision string
}

func (s service) selectedWorkerCommand(ctx context.Context, resolved resolvedHarnessSelection) (workerCommandPlan, error) {
	if resolved.View.Harness == codex.Identity {
		worker := (codex.Runner{WorkerProgram: resolved.Executable, WorkerModel: resolved.Profile.Model}).WorkflowWorker()

		return workerCommandPlan{Command: worker, RuntimeRef: "native", ProtocolRevision: codex.WorkerProtocol}, nil
	}

	if resolved.View.Harness != pi.Identity {
		return workerCommandPlan{}, errors.New("selected worker harness adapter is unavailable")
	}

	resources, err := pi.ResolveWorkerResources(ctx, s.env.ConfigDir(), resolved.Profile.CredentialRef, resolved.Profile.ModelConfigRef, resolved.Profile.ModelProvider)
	if err != nil {
		return workerCommandPlan{}, err
	}

	command, err := pi.WorkerCommand(resolved.Executable, pi.WorkerRuntimeOptions{
		Image: resolved.RuntimeImage, Provider: resolved.Profile.ModelProvider, Model: resolved.Profile.Model, Thinking: resolved.Profile.Thinking,
		CredentialFile: resources.CredentialFile, ModelFile: resources.ModelConfigFile, ModelHash: resources.ModelConfigHash, OAuth: resources.OAuth,
	})

	return workerCommandPlan{Command: command, RuntimeRef: resolved.RuntimeImage, ProtocolRevision: pi.ProtocolRevision}, err
}
