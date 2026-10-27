package core

import workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"

func validWorkflowIntegrationPolicyBounds(policy workflowstate.ExecutionPolicy) bool {
	return policy.MaxMergerAttempts >= 0 && policy.MaxMergerAttempts <= 10 && policy.MaxDriftRevisions >= 0 && policy.MaxDriftRevisions <= 10
}

func workflowMergerAttemptLimit(policy workflowstate.ExecutionPolicy) int {
	if policy.MaxMergerAttempts == 0 {
		return 3
	}

	return policy.MaxMergerAttempts
}

func workflowDriftRevisionLimit(policy workflowstate.ExecutionPolicy) int {
	if policy.MaxDriftRevisions == 0 {
		return 5
	}

	return policy.MaxDriftRevisions
}

func workflowIntegrationAdmissionLimits(policy workflowstate.ExecutionPolicy) WorkflowIntegrationLimits {
	return WorkflowIntegrationLimits{
		Worker: int64(policy.MaxWorkerAttempts), Review: workflowstate.PageLimit, Closing: workflowstate.PageLimit,
		Merger: int64(workflowMergerAttemptLimit(policy)), Drift: int64(workflowDriftRevisionLimit(policy)),
		Submission: int64(workflowIntegrationAttemptLimit(policy)),
	}
}
