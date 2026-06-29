package app

import (
	"context"
	"errors"
	"fmt"
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

	webActionSpecGenerate       = "spec.generate"
	webActionPlanGenerate       = "plan.generate"
	webActionPlanRegenerate     = "plan.regenerate"
	webActionPlanTrackingCreate = "plan.tracking.create"
	webActionQuizStart          = "quiz.start"
	webActionQuizResume         = "quiz.resume"
	webActionWorkPreview        = "work.preview"
	webActionWorkDaemonStop     = "work.daemon.stop"
)

func (p webViewProvider) WriteAction(ctx context.Context, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	if !request.Confirm {
		return webBlockedWrite("Action confirmation is required."), nil
	}

	operations := p.service(ctx)

	if response, handled, err := p.writeDocumentAction(operations, request); handled || err != nil {
		return response, err
	}

	if response, handled, err := p.writeGenerationAction(ctx, operations, request); handled || err != nil {
		return response, err
	}

	if response, handled, err := p.writeWorkAction(operations, request); handled || err != nil {
		return response, err
	}

	return webBlockedWrite("Unsupported web action: " + request.Action), nil
}

func (p webViewProvider) writeDocumentAction(operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, bool, error) {
	switch request.Action {
	case webActionCaseUpdate:
		response, err := p.writeCaseUpdate(operations, request)

		return response, true, err
	case webActionDumpCreate:
		response, err := p.writeDumpCreate(operations, request)

		return response, true, err
	case webActionDumpSave:
		response, err := p.writeDumpSave(operations, request)

		return response, true, err
	case webActionDumpAppend:
		response, err := p.writeDumpAppend(operations, request)

		return response, true, err
	default:
		return webui.WriteActionResponse{}, false, nil
	}
}

func (p webViewProvider) writeGenerationAction(ctx context.Context, operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, bool, error) {
	switch request.Action {
	case webActionSpecGenerate:
		response, err := p.writeSpecGenerate(ctx, operations, request)

		return response, true, err
	case webActionPlanGenerate:
		response, err := p.writePlanGenerate(ctx, operations, request)

		return response, true, err
	case webActionPlanRegenerate:
		response, err := p.writePlanRegenerate(ctx, operations, request)

		return response, true, err
	case webActionPlanTrackingCreate:
		response, err := p.writePlanTrackingCreate(ctx, operations, request)

		return response, true, err
	case webActionQuizStart:
		response, err := p.writeQuizStart(operations, request)

		return response, true, err
	case webActionQuizResume:
		response, err := p.writeQuizResume(operations, request)

		return response, true, err
	default:
		return webui.WriteActionResponse{}, false, nil
	}
}

func (p webViewProvider) writeWorkAction(operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, bool, error) {
	switch request.Action {
	case webActionWorkPreview:
		response, err := p.writeWorkPreview(operations)

		return response, true, err
	case webActionWorkDaemonStop:
		response, err := p.writeWorkDaemonStop(operations)

		return response, true, err
	default:
		return webui.WriteActionResponse{}, false, nil
	}
}

func (p webViewProvider) writeSpecGenerate(ctx context.Context, operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	result, err := operations.GenerateSpec(ctx, SpecGenerateRequest{Scope: webActionScope(request)})
	if err != nil {
		return webValidationWrite(err), nil
	}

	return webui.WriteActionResponse{
		Status:       webui.WriteStatusOK,
		Summary:      result.Message,
		ChangedFiles: compactChangedFiles(result.Spec.RelativePath),
		RedirectURL:  "/specs",
		Operation:    webOperation("spec.generate", "completed", result.Spec.RelativePath, result.Spec.Status),
	}, nil
}

func (p webViewProvider) writePlanGenerate(ctx context.Context, operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	result, err := operations.GeneratePlan(ctx, PlanGenerateRequest{Scope: webActionScope(request)})
	if err != nil {
		return webValidationWrite(err), nil
	}

	return webui.WriteActionResponse{
		Status:       webui.WriteStatusOK,
		Summary:      result.Message,
		ChangedFiles: compactChangedFiles(result.Plan.RelativePath),
		RedirectURL:  "/plans",
		Operation:    webOperation("plan.generate", "completed", result.Plan.RelativePath, result.Plan.Status),
	}, nil
}

func (p webViewProvider) writePlanRegenerate(ctx context.Context, operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	result, err := operations.RegeneratePlan(ctx, PlanRegenerateRequest{Scope: webActionScope(request)})
	if err != nil {
		return webValidationWrite(err), nil
	}

	return webui.WriteActionResponse{
		Status:       webui.WriteStatusOK,
		Summary:      result.Message,
		ChangedFiles: compactChangedFiles(result.Plan.RelativePath),
		RedirectURL:  "/plans",
		Operation:    webOperation("plan.regenerate", "completed", result.Plan.RelativePath, result.Plan.Status),
	}, nil
}

func (p webViewProvider) writePlanTrackingCreate(ctx context.Context, operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	result, err := operations.MaterializePlanTracking(ctx, PlanTrackingRequest{Scope: webActionScope(request)})
	if err != nil {
		return webValidationWrite(err), nil
	}

	files := append(compactChangedFiles(result.TrackingPath), ticketPaths(result.Tickets)...)

	return webui.WriteActionResponse{
		Status:       webui.WriteStatusOK,
		Summary:      result.Message,
		ChangedFiles: files,
		RedirectURL:  "/plans",
		Operation:    webOperation("plan.tracking.create", "completed", result.TrackingPath, result.Plan.Status),
	}, nil
}

func (p webViewProvider) writeQuizStart(operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	result, err := operations.StartQuiz(QuizStartRequest{
		Scope: webActionScope(request),
		Mode:  webQuizMode(request.Fields["mode"]),
	})
	if err != nil {
		return webValidationWrite(err), nil
	}

	return webui.WriteActionResponse{
		Status:      webui.WriteStatusOK,
		Summary:     result.Message,
		RedirectURL: "/quizzes",
		Operation:   webOperation("quiz.start", string(result.State.Status), quizOperationID(result.State), result.State.BranchReturnReason),
	}, nil
}

func (p webViewProvider) writeQuizResume(operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	result, err := operations.ResumeQuiz(QuizResumeRequest{Scope: webActionScope(request)})
	if err != nil {
		return webValidationWrite(err), nil
	}

	return webui.WriteActionResponse{
		Status:      webui.WriteStatusOK,
		Summary:     result.Message,
		RedirectURL: "/quizzes",
		Operation:   webOperation("quiz.resume", string(result.State.Status), quizOperationID(result.State), result.State.BranchReturnReason),
	}, nil
}

func (p webViewProvider) writeWorkPreview(operations service) (webui.WriteActionResponse, error) {
	result, err := operations.PreviewAgentWorkLoop(AgentWorkSelectionOptions{})
	if err != nil {
		return webValidationWrite(err), nil
	}

	return webui.WriteActionResponse{
		Status:      webui.WriteStatusOK,
		Summary:     result.Message,
		RedirectURL: "/dash",
		Operation: webOperation(
			"work.preview",
			"completed",
			fmt.Sprintf("%d eligible", result.EligibleCount),
			fmt.Sprintf("%d selected", len(result.SelectedQueue)),
		),
	}, nil
}

func (p webViewProvider) writeWorkDaemonStop(operations service) (webui.WriteActionResponse, error) {
	result, err := operations.RequestDaemonStop()
	if err != nil {
		return webValidationWrite(err), nil
	}

	return webui.WriteActionResponse{
		Status:      webui.WriteStatusOK,
		Summary:     result.Message,
		RedirectURL: "/dash",
		Operation:   webOperation("work.daemon.stop", "completed", "daemon", result.Message),
	}, nil
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

func webActionScope(request webui.WriteActionRequest) core.QuizScope {
	return core.QuizScope{CaseSlug: strings.TrimSpace(request.Fields["case_slug"])}
}

func webQuizMode(value string) core.QuizMode {
	mode := core.QuizMode(strings.TrimSpace(value))
	if mode.Valid() {
		return mode
	}

	return core.QuizModeDeep
}

func webOperation(kind, status, id, detail string) webui.OperationStateView {
	return webui.OperationStateView{
		ID:     id,
		Kind:   kind,
		Status: status,
		Detail: detail,
	}
}

func quizOperationID(state core.QuizRuntimeState) string {
	return firstNonEmptyString(state.CuratedPath, state.RawPath, state.CaseSlug, state.ProjectID)
}

func compactChangedFiles(paths ...string) []string {
	files := []string{}

	for _, path := range paths {
		if strings.TrimSpace(path) != "" {
			files = append(files, path)
		}
	}

	return files
}

func ticketPaths(tickets []core.IndexedTicket) []string {
	paths := make([]string, 0, len(tickets))
	for _, ticket := range tickets {
		if strings.TrimSpace(ticket.RelativePath) != "" {
			paths = append(paths, ticket.RelativePath)
		}
	}

	return paths
}
