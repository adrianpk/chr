package workflow

import (
	"context"
	"slices"

	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

type workflowHarnessRequestLookup struct {
	Request StartRequest
	Found   bool
}

// HarnessResolution binds the prepared host command to its semantic identity.
// This transient outcome does not change retained admission or invocation records.
type HarnessResolution struct {
	Binding *harnesscore.Binding
	Command []string
}

// HarnessResolver prepares the current explicit choice from the same
// complete configuration image that owns admission. Retained requests skip it.
type HarnessResolver func(context.Context, workspaceconfig.Config) (HarnessResolution, error)

// workflowHarnessResolver resolves current host configuration and executable bindings under Store-owned admission exclusion.
type workflowHarnessResolver struct{ store *Store }

func (host workflowHarnessResolver) resolve(ctx context.Context, request StartRequest) (StartRequest, error) {
	harnessRequest, err := host.store.retainedWorkflowHarnessRequest(ctx, request)

	if err != nil || harnessRequest.Found {
		return harnessRequest.Request, err
	}

	config, err := workspaceconfig.Load(host.store.layout.ProjectDir())
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

func (s *Store) retainedWorkflowHarnessRequest(ctx context.Context, request StartRequest) (workflowHarnessRequestLookup, error) {
	runs, err := s.repositoryWorkflowInvocations(ctx)
	if err != nil {
		return workflowHarnessRequestLookup{Request: request, Found: false}, err
	}

	for _, run := range runs {
		if run.Namespace == s.layout.Namespace() && slices.Contains(run.RequestKeys, request.RequestKey) {
			retained, err := retainWorkflowHarnessRequest(request, run)

			return workflowHarnessRequestLookup{Request: retained, Found: true}, err
		}
	}

	return workflowHarnessRequestLookup{Request: request, Found: false}, nil
}

func retainWorkflowHarnessRequest(request StartRequest, run workflowInvocationFile) (StartRequest, error) {
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

func checkCurrentWorkflowHarness(config workspaceconfig.Config, request StartRequest) error {
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
func (s *Store) WorkflowHarnessWorker(ctx context.Context, id string) (command []string, err error) {
	err = s.withRepositoryLock(ctx, func() error { return s.prepareWorkflowRead(ctx) }, func() error {
		invocationLocation, readErr := s.locateWorkflowInvocation(ctx, id)
		if readErr != nil {
			return readErr
		}

		if readErr := harnesshost.CheckExecutable(invocationLocation.Run.HarnessBinding, invocationLocation.Run.HarnessCommand); readErr != nil {
			return readErr
		}

		command = slices.Clone(invocationLocation.Run.HarnessCommand)

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
