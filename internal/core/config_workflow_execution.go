package core

import (
	"context"
	"fmt"
	"slices"
)

func (s *WorkflowStore) retainedWorkflowExecutionPolicy(ctx context.Context, key string, policy WorkflowExecutionPolicy, later *bool) (WorkflowExecutionPolicy, bool, error) {
	runs, err := s.repositoryWorkflowInvocations(ctx)
	if err != nil {
		return policy, false, err
	}

	for _, run := range runs {
		if run.Namespace != s.layout.Namespace() || !slices.Contains(run.RequestKeys, key) {
			continue
		}

		if workflowRequestedRetainedPolicy(policy, run.Policy, later) != run.Policy {
			return policy, false, fmt.Errorf("%w: request key names a different frozen policy", ErrWorkflowConflict)
		}

		return run.Policy, true, nil
	}

	if policy.ConfigRevision != "" {
		return policy, false, fmt.Errorf("%w: a frozen policy must match its retained request", ErrWorkflowConflict)
	}

	return policy, false, nil
}

func (s *WorkflowStore) resolveWorkflowExecutionPolicy(ctx context.Context, key string, policy WorkflowExecutionPolicy, later *bool) (WorkflowExecutionPolicy, error) {
	config, err := LoadWorkspaceConfig(s.layout.ProjectDir())
	if err != nil {
		return policy, err
	}

	// Retained admission, including pre-version records, keeps its exact policy.
	// Current execution defaults cannot refill budgets. Publication reads policy
	// again, and a caller cannot forge a frozen snapshot for another request.
	retained, found, err := s.retainedWorkflowExecutionPolicy(ctx, key, policy, later)
	if err != nil || found {
		return retained, err
	}

	policy, err = applyWorkflowExecutionLimits(config, policy)
	if err != nil {
		return policy, err
	}

	policy.IncludeLaterTickets = policy.IncludeLaterTickets || config.Work.IncludeLaterTickets
	if later != nil {
		policy.IncludeLaterTickets = *later
	}

	policy.ConfigRevision = workflowHash(nil)
	if config.workflowRevision != "" {
		policy.ConfigRevision = config.workflowRevision
	}

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

func (s *WorkflowStore) workflowClosingValidationTimeout(ctx context.Context, id string, contract int) (int, error) {
	config, err := LoadWorkspaceConfig(s.layout.ProjectDir())
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

func applyWorkflowExecutionLimits(config WorkspaceConfig, policy WorkflowExecutionPolicy) (WorkflowExecutionPolicy, error) {
	integration := config.Work.MaxIntegrationAttempts
	if config.Workflow.Version == 0 && !config.workflowExplicit["work.max_integration_attempts"] {
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
		} else if config.workflowExplicit[field.key] && *field.value > field.limit {
			return policy, fmt.Errorf("invocation cannot relax explicit %s", field.key)
		}
	}

	return policy, nil
}
