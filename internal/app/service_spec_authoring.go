package app

import (
	"forge.adrianpk.com/chr/internal/core"
)

// CreateSpecRequest selects a native lifecycle without a generation prerequisite.
type CreateSpecRequest struct {
	OperationID, Title, Body, Status, CaseSlug string
}

// SaveSpecRequest retains the exact document and scope observed by its editor.
type SaveSpecRequest struct {
	OperationID, Reference, Revision, Title, Body, Status, CaseSlug string
}

func (s service) CreateSpec(request CreateSpecRequest) (core.SpecMutationResult, error) {
	layout, err := core.ResolveOpsLayout(core.ResolveOpsLayoutOptions{ProjectDir: s.projectContext.ProjectDir})
	if err != nil {
		return core.SpecMutationResult{}, err
	}

	return core.MutateSpec(s.ctx, s.projectContext.ProjectDir, core.SpecMutationRequest{
		Action: "create", Namespace: layout.Namespace(), ProjectID: s.projectContext.ProjectID,
		OperationID: request.OperationID, Title: request.Title, Body: request.Body, Status: request.Status, CaseSlug: request.CaseSlug,
	})
}

func (s service) SaveSpec(request SaveSpecRequest) (core.SpecMutationResult, error) {
	return core.MutateSpec(s.ctx, s.projectContext.ProjectDir, core.SpecMutationRequest{
		Action: "save", ProjectID: s.projectContext.ProjectID, OperationID: request.OperationID,
		Reference: request.Reference, ExpectedRevision: request.Revision,
		Title: request.Title, Body: request.Body, Status: request.Status, CaseSlug: request.CaseSlug,
	})
}
