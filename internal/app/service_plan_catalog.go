package app

import "forge.adrianpk.com/chr/internal/core"

// ListPlanDocuments reads actual documents independently of generation scopes.
func (s service) ListPlanDocuments() ([]core.PlanDocument, error) {
	return core.ListPlanDocuments(s.ctx, s.projectContext.ProjectDir)
}
