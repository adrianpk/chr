package app

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strings"

	"forge.adrianpk.com/chc/internal/core"
	"forge.adrianpk.com/chc/internal/webui"
)

const (
	webActionCaseUpdate = "case.update"
	webActionDumpCreate = "dump.create"
	webActionDumpSave   = "dump.save"
	webActionDumpAppend = "dump.append"
)

func (p webViewProvider) WriteAction(ctx context.Context, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	if !request.Confirm {
		return webBlockedWrite("Action confirmation is required."), nil
	}

	operations := newService(ctx, p.env, p.projectContext, p.assetsFS)

	switch request.Action {
	case webActionCaseUpdate:
		return p.writeCaseUpdate(operations, request)
	case webActionDumpCreate:
		return p.writeDumpCreate(operations, request)
	case webActionDumpSave:
		return p.writeDumpSave(operations, request)
	case webActionDumpAppend:
		return p.writeDumpAppend(operations, request)
	default:
		return webBlockedWrite("Unsupported web action: " + request.Action), nil
	}
}

func (p webViewProvider) writeCaseUpdate(operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	current, err := core.NewCaseMetadataService(p.projectContext.ProjectDir).LoadPath(
		core.CaseMetadataPath(p.projectContext.ProjectDir, request.ID),
	)
	if err != nil {
		return webui.WriteActionResponse{}, err
	}

	if stale := webStaleWrite(current.SourceHash, request.Revision); stale.Status != "" {
		return stale, nil
	}

	result, err := operations.PatchCaseMetadata(CaseMetadataPatchRequest{
		Slug:        request.ID,
		Name:        fieldPointer(request.Fields, "name"),
		Description: fieldPointer(request.Fields, "description"),
		Context:     fieldPointer(request.Fields, "context"),
	})
	if err != nil {
		return webValidationWrite(err), nil
	}

	return webui.WriteActionResponse{
		Status:       webui.WriteStatusOK,
		Summary:      result.Message,
		ChangedFiles: []string{caseRelativePath(p.projectContext.ProjectDir, result.Case)},
		RedirectURL:  "/cases/detail?id=" + url.QueryEscape(request.ID),
	}, nil
}

func (p webViewProvider) writeDumpCreate(operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	result, err := operations.CreateDump(DumpCreateRequest{
		Title:   request.Fields["title"],
		Content: request.Fields["content"],
		Scope:   DumpScope{CaseSlug: request.Fields["case_slug"]},
	})
	if err != nil {
		return webValidationWrite(err), nil
	}

	return webui.WriteActionResponse{
		Status:      webui.WriteStatusOK,
		Summary:     result.Message,
		RedirectURL: "/dumps",
	}, nil
}

func (p webViewProvider) writeDumpSave(operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	if response, ok, err := p.validateDumpRevision(request); ok || err != nil {
		return response, err
	}

	result, err := operations.SaveDump(DumpSaveRequest{
		RelativePath: request.ID,
		Content:      request.Fields["content"],
	})
	if err != nil {
		return webValidationWrite(err), nil
	}

	return webui.WriteActionResponse{
		Status:       webui.WriteStatusOK,
		Summary:      result.Message,
		ChangedFiles: []string{request.ID},
		RedirectURL:  "/dumps/detail?id=" + url.QueryEscape(request.ID),
	}, nil
}

func (p webViewProvider) writeDumpAppend(operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	if response, ok, err := p.validateDumpRevision(request); ok || err != nil {
		return response, err
	}

	result, err := operations.AppendDump(DumpAppendRequest{
		RelativePath: request.ID,
		Content:      request.Fields["content"],
	})
	if err != nil {
		return webValidationWrite(err), nil
	}

	return webui.WriteActionResponse{
		Status:       webui.WriteStatusOK,
		Summary:      result.Message,
		ChangedFiles: []string{request.ID},
		RedirectURL:  "/dumps/detail?id=" + url.QueryEscape(request.ID),
	}, nil
}

func (p webViewProvider) validateDumpRevision(request webui.WriteActionRequest) (webui.WriteActionResponse, bool, error) {
	content, err := p.readProjectArtifact(request.ID)
	if errors.Is(err, os.ErrNotExist) {
		return webBlockedWrite("Dump does not exist."), true, nil
	}

	if err != nil {
		return webui.WriteActionResponse{}, false, err
	}

	if stale := webStaleWrite(webContentRevision(content), request.Revision); stale.Status != "" {
		return stale, true, nil
	}

	return webui.WriteActionResponse{}, false, nil
}

func webBlockedWrite(reason string) webui.WriteActionResponse {
	return webui.WriteActionResponse{
		Status:        webui.WriteStatusBlocked,
		Summary:       "Write action blocked.",
		BlockedReason: reason,
	}
}

func webValidationWrite(err error) webui.WriteActionResponse {
	return webui.WriteActionResponse{
		Status:     webui.WriteStatusValidation,
		Summary:    "Write action failed validation.",
		Validation: []string{err.Error()},
	}
}

func webStaleWrite(currentRevision, expectedRevision string) webui.WriteActionResponse {
	if strings.TrimSpace(expectedRevision) == "" || currentRevision == expectedRevision {
		return webui.WriteActionResponse{}
	}

	return webui.WriteActionResponse{
		Status:        webui.WriteStatusStale,
		Summary:       "Write action rejected stale data.",
		BlockedReason: "The artifact changed after the page was loaded. Reload before saving.",
	}
}

func fieldPointer(fields map[string]string, name string) *string {
	value, ok := fields[name]
	if !ok {
		return nil
	}

	return &value
}
