package app

import (
	"context"
	"errors"
	"slices"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/harness/codex"
)

func (s service) prepareWorkflowHarness(ctx context.Context, config core.WorkspaceConfig) (*core.HarnessBinding, []string, error) {
	if _, err := config.HarnessSelections(s.projectContext.ProjectID); err != nil {
		return nil, nil, err
	}

	if config.Harnesses.Worker == "" {
		return nil, nil, nil
	}

	resolved, err := s.resolveHarnessSelection(ctx, config, core.HarnessWorker, "")
	if err != nil {
		return nil, nil, err
	}

	if resolved.View.Harness != codex.Identity {
		return nil, nil, errors.New("selected worker harness adapter is unavailable")
	}

	runner := codex.Runner{WorkerProgram: resolved.Executable, WorkerModel: resolved.Profile.Model}

	executable, command, err := core.FreezeHarnessExecutable(runner.WorkflowWorker())
	if err != nil {
		return nil, nil, err
	}

	if s.workflowWorker != nil && !slices.Equal(s.workflowWorker, command) {
		return nil, nil, core.ErrWorkflowConflict
	}

	binding, err := core.SealHarnessBinding(core.HarnessBinding{
		SchemaVersion: 1, Role: core.HarnessWorker, Harness: resolved.View.Harness, Source: resolved.View.Source,
		SelectionRevision: resolved.View.SelectionRevision, Profile: resolved.View.Profile, ProfileRevision: resolved.View.ProfileRevision,
		ModelProvider: harnessWorkerProvider(resolved.View.ModelProvider), Model: resolved.View.Model, Thinking: resolved.View.Thinking,
		ExecutableRef: executable, RuntimeRef: "native", ProtocolRevision: codex.WorkerProtocol, Capabilities: resolved.View.Capabilities,
	})

	return &binding, command, err
}

func harnessWorkerProvider(provider string) string {
	if provider == "" {
		return "cli_default"
	}

	return provider
}
