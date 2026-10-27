package core

import workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"

func checkWorkflowMergerHarnessUsage(input WorkflowMergerInput, certificate WorkflowMergerHarnessCertificate, usage *WorkflowMergerUsage) error {
	if usage == nil {
		return nil
	}

	if !certificate.ReportsWholeCall || usage.CapabilityRevision != certificate.CapabilityRevision {
		return workflowstate.ErrConflict
	}

	if err := validateWorkflowMergerReportedUsage(usage, input.TokenAllowance); err != nil {
		return err
	}

	if allowance := input.TokenAllowance; allowance != nil && (usage.Input > allowance.Input || usage.Output > allowance.Output) {
		return ErrWorkflowIntegrationBudget
	}

	return nil
}
