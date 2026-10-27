package app

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/url"

	"forge.adrianpk.com/chr/internal/core"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/webui"
)

func (p webViewProvider) specArtifactDetail(ctx context.Context, request webui.ArtifactDetailRequest) (webui.ArtifactDetailView, error) {
	operations := p.service(ctx)

	cases, err := core.NewCaseMetadataService(p.projectContext.ProjectDir).LoadBounded(ctx, workflowstate.WorkflowRecordLimit)
	if err != nil {
		return webui.ArtifactDetailView{}, err
	}

	editor := &webui.SpecEditorView{New: request.ID == "", Draft: true, RequestID: rand.Text()}
	for _, item := range cases {
		editor.Cases = append(editor.Cases, webui.TicketCaseView{Slug: item.Slug, Name: item.Name})
	}

	view := webui.ArtifactDetailView{Family: "specs", ID: request.ID, Spec: editor}
	if editor.New {
		return view, nil
	}

	document, err := core.ReadSpecDocument(ctx, operations.projectContext.ProjectDir, request.ID)
	if err != nil {
		return view, err
	}

	editor.Draft, editor.CaseSlug = document.Status == "draft", document.CaseSlug
	view.Title, view.Content, view.Status, view.Revision = document.Title, document.Body, document.Status, document.Revision

	return view, nil
}

func (p webViewProvider) writeSpecAction(operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	if request.Action == webActionSpecGenerate {
		return p.writeSpecGenerate(operations.ctx, operations, request)
	}

	status := "approved"
	if request.Fields["draft"] == "true" {
		status = "draft"
	}

	var (
		result core.SpecMutationResult
		err    error
	)

	switch request.Action {
	case "spec.create":
		result, err = operations.CreateSpec(CreateSpecRequest{OperationID: request.Fields["request_id"], Title: request.Fields["title"], Body: request.Fields["content"], Status: status, CaseSlug: request.Fields["case_slug"]})
	case "spec.save":
		result, err = operations.SaveSpec(SaveSpecRequest{OperationID: request.Fields["request_id"], Reference: request.ID, Revision: request.Revision, Title: request.Fields["title"], Body: request.Fields["content"], Status: status, CaseSlug: request.Fields["case_slug"]})
	default:
		err = fmt.Errorf("unsupported spec editor action")
	}

	if errors.Is(err, workflowstate.ErrConflict) {
		return webui.WriteActionResponse{Status: webui.WriteStatusStale, Summary: "This spec changed. Your edits are retained; reload before saving.", ReloadURL: "/specs/detail?id=" + url.QueryEscape(request.ID)}, nil
	}

	if err != nil {
		return webValidationWrite(err), nil
	}

	return webui.WriteActionResponse{Status: webui.WriteStatusOK, Summary: "Spec saved.", RedirectURL: "/specs", ChangedFiles: compactChangedFiles(result.Reference)}, nil
}
