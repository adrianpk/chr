package app

import "forge.adrianpk.com/chr/internal/core"

// ReadProjectDump observes the project-level destination across ops namespaces.
func (s service) ReadProjectDump() (core.ProjectDumpObservation, error) {
	return core.ObserveProjectDump(s.ctx, s.projectContext.ProjectDir)
}

// AdoptProjectDump binds an unambiguous existing dump without altering its components.
func (s service) AdoptProjectDump(expectedRevision string) (core.ProjectMetadata, error) {
	return core.NewProjectMetadataService(s.projectContext.ProjectDir).AdoptProjectDump(s.ctx, expectedRevision)
}

// PreviewProjectDumpReconciliation returns the observed non-destructive composition.
func (s service) PreviewProjectDumpReconciliation() (core.ProjectDumpReconciliationPreview, error) {
	return core.PreviewProjectDumpReconciliation(s.ctx, s.projectContext.ProjectDir)
}

// ReconcileProjectDump explicitly publishes the confirmed preview through the owner.
func (s service) ReconcileProjectDump(request core.ProjectDumpReconciliationRequest) (DumpOperationResult, error) {
	result, err := core.ReconcileProjectDump(s.ctx, s.projectContext.ProjectDir, request)
	if err != nil {
		return DumpOperationResult{}, err
	}

	return s.finishDumpMutation(result, "Reconciled"), nil
}
