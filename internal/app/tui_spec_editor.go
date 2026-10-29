package app

import (
	"path/filepath"

	"forge.adrianpk.com/chr/internal/app/operation"
	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/tui"
)

func (e tuiSpecActionExecutor) executeSpecEditor(request tui.SpecActionRequest) (tui.SpecActionResult, error) {
	if request.Action == tui.SpecActionEdit {
		document, err := authoringcore.ReadSpecDocument(e.operations.ctx, e.operations.projectContext.ProjectDir, request.Path)
		view := tui.SpecView{Title: document.Title, Content: document.Body, Status: document.Status, RelativePath: document.Reference, Revision: document.Revision, Scope: tui.QuizScope{CaseSlug: document.CaseSlug}}

		return tui.SpecActionResult{Document: &view}, err
	}

	var (
		saved authoringcore.SpecMutationResult
		err   error
	)
	if request.Action == tui.SpecActionCreate {
		saved, err = e.operations.CreateSpec(CreateSpecRequest{OperationID: request.OperationID, Title: request.Title, Body: request.Content, Status: request.Status, CaseSlug: request.Scope.CaseSlug})
	} else {
		saved, err = e.operations.SaveSpec(SaveSpecRequest{OperationID: request.OperationID, Reference: request.Path, Revision: request.Revision, Title: request.Title, Body: request.Content, Status: request.Status, CaseSlug: request.Scope.CaseSlug})
	}

	if err != nil {
		return tui.SpecActionResult{}, err
	}

	result := tui.SpecActionResult{Committed: true, Message: "Spec saved.", SelectedPath: saved.Reference, Scope: request.Scope}
	result.Outcome = operation.Result{ChangedFiles: compactChangedFiles(saved.Reference, filepath.ToSlash(filepath.Join(filepath.Dir(saved.Reference), "metadata.json")))}
	result.Specs, err = loadTUISpecs(e.operations)

	return result, err
}
