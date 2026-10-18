package app

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/webui"
)

const (
	webActionCaseUpdate = "case.update"
	webActionDumpCreate = "dump.create"
	webActionDumpSave   = "dump.save"
	webActionDumpAppend = "dump.append"

	webActionTicketUpdate       = "ticket.update"
	webActionTicketComment      = "ticket.comment"
	webActionTicketCreate       = "ticket.create"
	webActionTicketMove         = "ticket.move"
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

	if response, handled, err := p.writeTicketAction(operations, request); handled || err != nil {
		return response, err
	}

	if response, handled, err := p.writeWorkAction(ctx, operations, request); handled || err != nil {
		return response, err
	}

	return webBlockedWrite("Unsupported web action: " + request.Action), nil
}

func (p webViewProvider) writeDocumentAction(operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, bool, error) {
	if request.Surface == "quiz" {
		response, err := p.writeQuizAction(operations, request)

		return response, true, err
	}

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

func (p webViewProvider) writeWorkAction(ctx context.Context, operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, bool, error) {
	switch request.Action {
	case webActionWorkPreview:
		response, err := p.writeWorkPreview(operations)

		return response, true, err
	case webActionWorkDaemonStop:
		response, err := p.writeWorkDaemonStop(operations)

		return response, true, err
	default:
		if strings.HasPrefix(request.Action, "workflow.") {
			return p.writeWorkflowAction(ctx, operations, request), true, nil
		}

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
		RedirectURL: quizFocusURL(result.State.CuratedPath),
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
		RedirectURL: quizFocusURL(result.State.CuratedPath),
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
		stale.ReloadURL = "/cases/detail?id=" + url.QueryEscape(request.ID)
		stale.DiscardURL = "/cases"

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
	if request.Fields["request_id"] == "" {
		return webui.WriteActionResponse{Status: webui.WriteStatusValidation, Summary: "Reopen New before submitting this capture."}, nil
	}

	result, err := operations.CreateDump(DumpCreateRequest{
		OperationID: dumpRequestID(request.Fields["request_id"]), ExpectedRevision: request.Revision,
		Title:   request.Fields["title"],
		Content: request.Fields["content"],
		Scope:   DumpScope{CaseSlug: request.Fields["case_slug"]},
	})
	if err != nil {
		return webDumpWriteFailure(err, request.ID), nil
	}

	return webui.WriteActionResponse{
		Status:          webui.WriteStatusOK,
		Summary:         result.Message,
		Operation:       webDumpOperation("dump.capture", result),
		ProjectionStale: result.ProjectionStale,
		RedirectURL:     "/dumps",
	}, nil
}

func (p webViewProvider) writeDumpSave(operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	result, err := operations.SaveDump(DumpSaveRequest{
		OperationID: dumpRequestID(request.Fields["request_id"]), ExpectedRevision: request.Revision,
		RelativePath: request.ID,
		Content:      request.Fields["content"],
	})
	if err != nil {
		return webDumpWriteFailure(err, request.ID), nil
	}

	return webui.WriteActionResponse{
		Status:          webui.WriteStatusOK,
		Summary:         result.Message,
		ChangedFiles:    []string{result.RelativePath},
		Operation:       webDumpOperation("dump.save", result),
		ProjectionStale: result.ProjectionStale,
		RedirectURL:     "/dumps/detail?id=" + url.QueryEscape(result.RelativePath),
	}, nil
}

func (p webViewProvider) writeDumpAppend(operations service, request webui.WriteActionRequest) (webui.WriteActionResponse, error) {
	result, err := operations.AppendDump(DumpAppendRequest{
		OperationID: dumpRequestID(request.Fields["request_id"]), ExpectedRevision: request.Revision,
		RelativePath: request.ID,
		Content:      request.Fields["content"],
	})
	if err != nil {
		return webDumpWriteFailure(err, request.ID), nil
	}

	return webui.WriteActionResponse{
		Status:          webui.WriteStatusOK,
		Summary:         result.Message,
		ChangedFiles:    []string{result.RelativePath},
		Operation:       webDumpOperation("dump.append", result),
		ProjectionStale: result.ProjectionStale,
		RedirectURL:     "/dumps/detail?id=" + url.QueryEscape(result.RelativePath),
	}, nil
}

func webDumpOperation(kind string, result DumpOperationResult) webui.OperationStateView {
	detail := ""
	if result.ProjectionStale {
		detail = "Catalog refresh is pending."
	}

	return webOperation(kind, "committed", result.Mutation.OperationID, detail)
}

func webDumpWriteFailure(err error, reference string) webui.WriteActionResponse {
	if errors.Is(err, core.ErrWorkflowConflict) {
		reload := "/dumps"
		if reference != "" {
			reload += "/detail?id=" + url.QueryEscape(reference)
		}

		return webui.WriteActionResponse{Status: webui.WriteStatusStale, Summary: "The dump changed after it was loaded. Reload before saving.", ReloadURL: reload, DiscardURL: "/dumps"}
	}

	if errors.Is(err, core.ErrProjectDumpReconciliation) {
		return webBlockedWrite(err.Error())
	}

	return webValidationWrite(err)
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
