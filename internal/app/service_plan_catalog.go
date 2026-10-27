package app

import authoringcore "forge.adrianpk.com/chr/internal/core/authoring"

// ListPlanDocuments reads actual documents independently of generation scopes.
func (s service) ListPlanDocuments() ([]authoringcore.PlanDocument, error) {
	return authoringcore.ListPlanDocuments(s.ctx, s.projectContext.ProjectDir)
}
