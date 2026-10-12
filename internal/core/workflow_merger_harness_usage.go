package core

func checkWorkflowMergerHarnessUsage(input WorkflowMergerInput, certificate WorkflowMergerHarnessCertificate, usage *WorkflowMergerUsage) error {
	if usage == nil {
		return nil
	}

	if !certificate.ReportsWholeCall || usage.CapabilityRevision != certificate.CapabilityRevision {
		return ErrWorkflowConflict
	}

	if err := validateWorkflowMergerReportedUsage(usage, input.TokenAllowance); err != nil {
		return err
	}

	if allowance := input.TokenAllowance; allowance != nil && (usage.Input > allowance.Input || usage.Output > allowance.Output) {
		return ErrWorkflowIntegrationBudget
	}

	return nil
}
