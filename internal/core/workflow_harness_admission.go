package core

import (
	"context"
	"slices"

	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowHarnessResolution binds the prepared host command to its semantic identity.
// This transient outcome does not change retained admission or invocation records.
type WorkflowHarnessResolution struct {
	Binding *harnesscore.Binding
	Command []string
}

// WorkflowHarnessResolver prepares the current explicit choice from the same
// complete configuration image that owns admission. Retained requests skip it.
type WorkflowHarnessResolver func(context.Context, WorkspaceConfig) (WorkflowHarnessResolution, error)

func (s *WorkflowStore) resolveWorkflowHarnessRequest(ctx context.Context, request WorkflowStartRequest) (WorkflowStartRequest, error) {
	retained, found, err := s.retainedWorkflowHarnessRequest(ctx, request)
	if err != nil || found {
		return retained, err
	}

	config, err := LoadWorkspaceConfig(s.layout.ProjectDir())
	if err != nil {
		return request, err
	}

	if request.ResolveHarness != nil {
		resolved, err := request.ResolveHarness(ctx, config)
		if err != nil {
			return request, err
		}

		if request.HarnessBinding != nil && !sameHarnessBinding(request.HarnessBinding, resolved.Binding) {
			return request, workflowstate.ErrConflict
		}

		request.HarnessBinding, request.HarnessCommand = resolved.Binding, resolved.Command
	}

	if err := checkCurrentWorkflowHarness(config, request); err != nil {
		return request, err
	}

	return request, harnesshost.CheckExecutable(request.HarnessBinding, request.HarnessCommand)
}

func (s *WorkflowStore) retainedWorkflowHarnessRequest(ctx context.Context, request WorkflowStartRequest) (WorkflowStartRequest, bool, error) {
	runs, err := s.repositoryWorkflowInvocations(ctx)
	if err != nil {
		return request, false, err
	}

	for _, run := range runs {
		if run.Namespace == s.layout.Namespace() && slices.Contains(run.RequestKeys, request.RequestKey) {
			retained, err := retainWorkflowHarnessRequest(request, run)

			return retained, true, err
		}
	}

	return request, false, nil
}

func retainWorkflowHarnessRequest(request WorkflowStartRequest, run workflowInvocationFile) (WorkflowStartRequest, error) {
	if request.HarnessBinding != nil && !sameHarnessBinding(request.HarnessBinding, run.HarnessBinding) {
		return request, workflowstate.ErrConflict
	}

	if len(request.HarnessCommand) > 0 && !slices.Equal(request.HarnessCommand, run.HarnessCommand) {
		return request, workflowstate.ErrConflict
	}

	request.HarnessBinding = harnesscore.CloneBinding(run.HarnessBinding)
	request.HarnessCommand = slices.Clone(run.HarnessCommand)

	return request, nil
}

func sameHarnessBinding(left, right *harnesscore.Binding) bool {
	if left == nil || right == nil {
		return left == right
	}

	return *left == *right
}

func checkCurrentWorkflowHarness(config WorkspaceConfig, request WorkflowStartRequest) error {
	binding := request.HarnessBinding
	if binding == nil {
		if config.Harnesses.Worker != "" {
			return workflowstate.ErrConflict
		}

		return harnesshost.ValidateBindingCommand(nil, request.HarnessCommand)
	}

	if binding.SelectionRevision != config.Revision() || binding.Profile != config.Harnesses.Worker {
		return workflowstate.ErrConflict
	}

	if request.Policy.ConfigRevision != binding.SelectionRevision {
		return workflowstate.ErrConflict
	}

	return harnesshost.ValidateBindingCommand(binding, request.HarnessCommand)
}

// WorkflowHarnessWorker reads only the invocation's retained host command. A nil
// binding returns nil so the explicit historical v1 execution path stays intact.
func (s *WorkflowStore) WorkflowHarnessWorker(ctx context.Context, id string) (command []string, err error) {
	err = s.withRepositoryLock(ctx, func() error { return s.prepareWorkflowRead(ctx) }, func() error {
		run, _, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		if readErr := harnesshost.CheckExecutable(run.HarnessBinding, run.HarnessCommand); readErr != nil {
			return readErr
		}

		command = slices.Clone(run.HarnessCommand)

		return nil
	})

	return command, err
}

func validateWorkflowInvocationHarness(run workflowInvocationFile) error {
	switch run.Schema {
	case "chr-invocation/v1":
		if run.HarnessBinding != nil || len(run.HarnessCommand) != 0 {
			return workflowstate.ErrConflict
		}
	case "chr-invocation/v2":
		if run.HarnessBinding == nil {
			return workflowstate.ErrConflict
		}
	default:
		return workflowstate.ErrConflict
	}

	return harnesshost.ValidateBindingCommand(run.HarnessBinding, run.HarnessCommand)
}
