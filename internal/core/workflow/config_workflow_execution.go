package workflow

import (
	"context"
	"fmt"
	"slices"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

type retainedExecutionPolicy struct {
	Policy workflowstate.ExecutionPolicy
	Found  bool
}

func (s *Store) retainedWorkflowExecutionPolicy(ctx context.Context, key string, policy workflowstate.ExecutionPolicy, later *bool) (retainedExecutionPolicy, error) {
	runs, err := s.repositoryWorkflowInvocations(ctx)
	if err != nil {
		return retainedExecutionPolicy{Policy: policy}, err
	}

	for _, run := range runs {
		if run.Namespace != s.layout.Namespace() || !slices.Contains(run.RequestKeys, key) {
			continue
		}

		if workflowRequestedRetainedPolicy(policy, run.Policy, later) != run.Policy {
			return retainedExecutionPolicy{Policy: policy}, fmt.Errorf("%w: request key names a different frozen policy", workflowstate.ErrConflict)
		}

		return retainedExecutionPolicy{Policy: run.Policy, Found: true}, nil
	}

	if policy.ConfigRevision != "" {
		return retainedExecutionPolicy{Policy: policy}, fmt.Errorf("%w: a frozen policy must match its retained request", workflowstate.ErrConflict)
	}

	return retainedExecutionPolicy{Policy: policy}, nil
}

func (s *Store) resolveWorkflowExecutionPolicy(ctx context.Context, key string, policy workflowstate.ExecutionPolicy, later *bool) (workflowstate.ExecutionPolicy, error) {
	config, err := workspaceconfig.Load(s.layout.ProjectDir())
	if err != nil {
		return policy, err
	}

	// Retained admission, including pre-version records, keeps its exact policy.
	// Current execution defaults cannot refill budgets. Publication reads policy
	// again, and a caller cannot forge a frozen snapshot for another request.
	retained, err := s.retainedWorkflowExecutionPolicy(ctx, key, policy, later)
	if err != nil || retained.Found {
		return retained.Policy, err
	}

	return configuredWorkflowExecutionPolicy(config, policy, later)
}

// InspectWorkflowExecutionPolicy reads finite admission defaults without
// allocating an invocation, refilling a budget or claiming a worker.
func (s *Store) InspectWorkflowExecutionPolicy(ctx context.Context) (policy workflowstate.ExecutionPolicy, err error) {
	err = s.withWorkflowInspection(ctx, func() error {
		config, readErr := workspaceconfig.Load(s.layout.ProjectDir())
		if readErr != nil {
			return readErr
		}

		policy, readErr = configuredWorkflowExecutionPolicy(config, workflowstate.ExecutionPolicy{}, nil)

		return readErr
	})

	return policy, err
}

func configuredWorkflowExecutionPolicy(config workspaceconfig.Config, policy workflowstate.ExecutionPolicy, later *bool) (workflowstate.ExecutionPolicy, error) {
	var err error

	policy, err = applyWorkflowExecutionLimits(config, policy)
	if err != nil {
		return policy, err
	}

	policy.IncludeLaterTickets = policy.IncludeLaterTickets || config.Work.IncludeLaterTickets
	if later != nil {
		policy.IncludeLaterTickets = *later
	}

	policy.ConfigRevision = config.Revision()

	policy = normalizeWorkflowExecutionPolicy(policy)
	if !validWorkflowExecutionPolicy(policy) {
		return policy, fmt.Errorf("workflow execution policy bounds exceeded")
	}

	return policy, nil
}

func workflowValidationTimeout(contract, configured int) int {
	if configured == 0 {
		return contract
	}

	return min(contract, configured)
}

func (s *Store) workflowClosingValidationTimeout(ctx context.Context, id string, contract int) (int, error) {
	config, err := workspaceconfig.Load(s.layout.ProjectDir())
	if err != nil {
		return 0, err
	}

	seconds := config.Work.ValidationTimeoutSeconds

	if id != "" {
		run, _, err := s.locateWorkflowInvocation(ctx, id)
		if err != nil {
			return 0, err
		}

		seconds = run.Policy.ValidationTimeoutSeconds
	}

	return workflowValidationTimeout(contract, seconds), nil
}

func applyWorkflowExecutionLimits(config workspaceconfig.Config, policy workflowstate.ExecutionPolicy) (workflowstate.ExecutionPolicy, error) {
	integration := config.Work.MaxIntegrationAttempts
	if config.Workflow.Version == 0 && !config.Explicit("work.max_integration_attempts") {
		// Legacy zero already means the finite default. Retain that original
		// request representation rather than changing its durable identity.
		integration = 0
	}

	fields := []struct {
		key   string
		value *int
		limit int
	}{
		{"work.worker_timeout_seconds", &policy.WorkerTimeoutSeconds, config.Work.WorkerTimeoutSeconds},
		{"work.validation_timeout_seconds", &policy.ValidationTimeoutSeconds, config.Work.ValidationTimeoutSeconds},
		{"work.max_worker_attempts", &policy.MaxWorkerAttempts, config.Work.MaxWorkerAttempts},
		{"work.max_integration_attempts", &policy.MaxIntegrationAttempts, integration},
		{"work.checkpoint_step_limit", &policy.CheckpointStepLimit, config.Work.CheckpointStepLimit},
	}
	for _, field := range fields {
		if *field.value == 0 {
			*field.value = field.limit
		} else if config.Explicit(field.key) && *field.value > field.limit {
			return policy, fmt.Errorf("invocation cannot relax explicit %s", field.key)
		}
	}

	return policy, nil
}
