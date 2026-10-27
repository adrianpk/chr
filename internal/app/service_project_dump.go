package app

import authoringcore "forge.adrianpk.com/chr/internal/core/authoring"

// ReadProjectDump observes the project-level destination across ops namespaces.
func (s service) ReadProjectDump() (authoringcore.ProjectDumpObservation, error) {
	return authoringcore.ObserveProjectDump(s.ctx, s.projectContext.ProjectDir)
}

// AdoptProjectDump binds an unambiguous existing dump without altering its components.
func (s service) AdoptProjectDump(expectedRevision string) (authoringcore.ProjectMetadata, error) {
	return authoringcore.NewProjectMetadataService(s.projectContext.ProjectDir).AdoptProjectDump(s.ctx, expectedRevision)
}

// PreviewProjectDumpReconciliation returns the observed non-destructive composition.
func (s service) PreviewProjectDumpReconciliation() (authoringcore.ProjectDumpReconciliationPreview, error) {
	return authoringcore.PreviewProjectDumpReconciliation(s.ctx, s.projectContext.ProjectDir)
}

// ReconcileProjectDump explicitly publishes the confirmed preview through the owner.
func (s service) ReconcileProjectDump(request authoringcore.ProjectDumpReconciliationRequest) (DumpOperationResult, error) {
	result, err := authoringcore.ReconcileProjectDump(s.ctx, s.projectContext.ProjectDir, request)
	if err != nil {
		return DumpOperationResult{}, err
	}

	return s.finishDumpMutation(result, "Reconciled"), nil
}
