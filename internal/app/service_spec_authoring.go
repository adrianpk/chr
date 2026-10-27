package app

import (
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
)

// CreateSpecRequest selects a native lifecycle without a generation prerequisite.
type CreateSpecRequest struct {
	OperationID, Title, Body, Status, CaseSlug string
}

// SaveSpecRequest retains the exact document and scope observed by its editor.
type SaveSpecRequest struct {
	OperationID, Reference, Revision, Title, Body, Status, CaseSlug string
}

func (s service) CreateSpec(request CreateSpecRequest) (authoringcore.SpecMutationResult, error) {
	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: s.projectContext.ProjectDir})
	if err != nil {
		return authoringcore.SpecMutationResult{}, err
	}

	return authoringcore.MutateSpec(s.ctx, s.projectContext.ProjectDir, authoringcore.SpecMutationRequest{
		Action: "create", Namespace: layout.Namespace(), ProjectID: s.projectContext.ProjectID,
		OperationID: request.OperationID, Title: request.Title, Body: request.Body, Status: request.Status, CaseSlug: request.CaseSlug,
	})
}

func (s service) SaveSpec(request SaveSpecRequest) (authoringcore.SpecMutationResult, error) {
	return authoringcore.MutateSpec(s.ctx, s.projectContext.ProjectDir, authoringcore.SpecMutationRequest{
		Action: "save", ProjectID: s.projectContext.ProjectID, OperationID: request.OperationID,
		Reference: request.Reference, ExpectedRevision: request.Revision,
		Title: request.Title, Body: request.Body, Status: request.Status, CaseSlug: request.CaseSlug,
	})
}
