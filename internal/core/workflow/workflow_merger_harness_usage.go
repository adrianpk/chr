package workflow

import (
	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func checkWorkflowMergerHarnessUsage(input workflowstate.MergerInput, certificate MergerHarnessCertificate, usage *workflowstate.MergerUsage) error {
	if usage == nil {
		return nil
	}

	if !certificate.ReportsWholeCall || usage.CapabilityRevision != certificate.CapabilityRevision {
		return workflowstate.ErrConflict
	}

	if err := workflowmerge.ValidateWorkflowMergerReportedUsage(usage, input.TokenAllowance); err != nil {
		return err
	}

	if allowance := input.TokenAllowance; allowance != nil && (usage.Input > allowance.Input || usage.Output > allowance.Output) {
		return ErrWorkflowIntegrationBudget
	}

	return nil
}
