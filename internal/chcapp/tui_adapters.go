package chcapp

import (
	"errors"
	"fmt"
	"os/exec"

	"forge.adrianpk.com/chc/internal/chc"
	"forge.adrianpk.com/chc/internal/chctui"
)

type tuiWorkspaceRefreshExecutor struct {
	operations service
}

func newTUIWorkspaceRefreshExecutor(
	operations service,
) tuiWorkspaceRefreshExecutor {
	return tuiWorkspaceRefreshExecutor{
		operations: operations,
	}
}

func (e tuiWorkspaceRefreshExecutor) RefreshWorkspace() (chctui.WorkspaceRefreshResult, error) {
	result, err := e.operations.RefreshWorkspace()
	if err != nil {
		return chctui.WorkspaceRefreshResult{}, err
	}

	return chctui.WorkspaceRefreshResult{
		Message:              result.Message,
		Graph:                result.Graph,
		Index:                result.Index,
		IndexHealth:          result.IndexHealth,
		IndexChanges:         result.IndexChanges,
		IndexCache:           result.IndexCache,
		SemanticDigestStates: result.SemanticDigestStates,
	}, nil
}

type tuiProjectMetadataExecutor struct {
	operations service
}

func newTUIProjectMetadataExecutor(
	operations service,
) tuiProjectMetadataExecutor {
	return tuiProjectMetadataExecutor{
		operations: operations,
	}
}

func (e tuiProjectMetadataExecutor) SaveProjectMetadata(
	metadata chc.ProjectMetadata,
) (chctui.ProjectMetadataResult, error) {
	result, err := e.operations.SaveProjectMetadata(metadata)
	if err != nil {
		return chctui.ProjectMetadataResult{}, err
	}

	return chctui.ProjectMetadataResult{
		Message:  result.Message,
		Metadata: result.Metadata,
	}, nil
}

type tuiCaseMetadataExecutor struct {
	operations service
}

func newTUICaseMetadataExecutor(
	operations service,
) tuiCaseMetadataExecutor {
	return tuiCaseMetadataExecutor{
		operations: operations,
	}
}

func (e tuiCaseMetadataExecutor) SaveCaseMetadata(
	metadata chc.CaseMetadata,
) (chctui.CaseMetadataResult, error) {
	result, err := e.operations.SaveCaseMetadata(metadata)
	if err != nil {
		return chctui.CaseMetadataResult{}, err
	}

	return chctui.CaseMetadataResult{
		Message: result.Message,
		Case:    result.Case,
		Cases:   result.Cases,
	}, nil
}

type tuiArtifactActionExecutor struct {
	operations service
}

func newTUIArtifactActionExecutor(
	operations service,
) tuiArtifactActionExecutor {
	return tuiArtifactActionExecutor{
		operations: operations,
	}
}

func (e tuiArtifactActionExecutor) ExecuteArtifactAction(
	action chctui.ArtifactActionID,
) (chctui.ArtifactActionResult, error) {
	switch action {
	case chctui.ArtifactActionCreateTicket:
		result, err := e.operations.CreateDefaultTicket()
		if err != nil {
			return chctui.ArtifactActionResult{}, err
		}

		return e.artifactActionResult(result)
	case chctui.ArtifactActionCreateAuthoringSession:
		result, err := e.operations.CreateTUIAuthoringSession()
		if err != nil {
			return chctui.ArtifactActionResult{}, err
		}

		return e.artifactActionResult(result)
	case chctui.ArtifactActionCreateContextQuiz:
		return chctui.ArtifactActionResult{}, errors.New(
			"context-aware quiz creation requires provider-backed question generation and is not available from this TUI action yet",
		)
	default:
		return chctui.ArtifactActionResult{}, errors.New("unknown artifact action")
	}
}

func (e tuiArtifactActionExecutor) artifactActionResult(
	result ArtifactOperationResult,
) (chctui.ArtifactActionResult, error) {
	documents, err := chctui.LoadDocumentViews(e.operations.ctx, e.operations.projectContext.ProjectDir)
	if err != nil {
		return chctui.ArtifactActionResult{}, err
	}

	return chctui.ArtifactActionResult{
		Message:           result.Message,
		Tickets:           result.Tickets,
		Documents:         documents,
		AuthoringSessions: result.AuthoringSessions,
	}, nil
}

type tuiTicketActionExecutor struct {
	operations service
}

type tuiSpecActionExecutor struct {
	operations service
}

type tuiPlanActionExecutor struct {
	operations service
}

func newTUISpecActionExecutor(
	operations service,
) tuiSpecActionExecutor {
	return tuiSpecActionExecutor{
		operations: operations,
	}
}

func newTUIPlanActionExecutor(
	operations service,
) tuiPlanActionExecutor {
	return tuiPlanActionExecutor{
		operations: operations,
	}
}

func (e tuiSpecActionExecutor) ExecuteSpecAction(
	request chctui.SpecActionRequest,
) (chctui.SpecActionResult, error) {
	switch request.Action {
	case chctui.SpecActionGenerate:
		_, err := e.operations.GenerateSpec(SpecGenerateRequest{
			Scope: chc.QuizScope{CaseSlug: request.Scope.CaseSlug},
		})
		if err != nil {
			return chctui.SpecActionResult{}, err
		}
	case chctui.SpecActionRefresh:
	default:
		return chctui.SpecActionResult{}, errors.New("unknown spec action")
	}

	documents, err := chctui.LoadDocumentViews(e.operations.ctx, e.operations.projectContext.ProjectDir)
	if err != nil {
		return chctui.SpecActionResult{}, err
	}

	specs, err := loadTUISpecs(e.operations)
	if err != nil {
		return chctui.SpecActionResult{}, err
	}

	message := "Specs refreshed."
	if request.Action == chctui.SpecActionGenerate {
		message = "Spec generated."
	}

	return chctui.SpecActionResult{
		Message:   message,
		Documents: documents,
		Specs:     specs,
		Scope:     request.Scope,
	}, nil
}

func (e tuiPlanActionExecutor) ExecutePlanAction(
	request chctui.PlanActionRequest,
) (chctui.PlanActionResult, error) {
	e.operations.appendPlanGenerationLog("tui_plan_action_started", map[string]string{
		"action":    string(request.Action),
		"case_slug": request.Scope.CaseSlug,
	})

	message := ""

	switch request.Action {
	case chctui.PlanActionGenerate:
		result, err := e.operations.GeneratePlan(PlanGenerateRequest{
			Scope: chc.QuizScope{CaseSlug: request.Scope.CaseSlug},
		})
		if err != nil {
			e.operations.appendPlanGenerationLog("tui_plan_action_failed", map[string]string{
				"action":    string(request.Action),
				"case_slug": request.Scope.CaseSlug,
				"error":     err.Error(),
			})

			return chctui.PlanActionResult{}, err
		}

		message = result.Message
	case chctui.PlanActionRegenerate:
		result, err := e.operations.RegeneratePlan(PlanRegenerateRequest{
			Scope: chc.QuizScope{CaseSlug: request.Scope.CaseSlug},
		})
		if err != nil {
			e.operations.appendPlanGenerationLog("tui_plan_action_failed", map[string]string{
				"action":    string(request.Action),
				"case_slug": request.Scope.CaseSlug,
				"error":     err.Error(),
			})

			return chctui.PlanActionResult{}, err
		}

		message = result.Message
	case chctui.PlanActionRefresh:
	default:
		e.operations.appendPlanGenerationLog("tui_plan_action_failed", map[string]string{
			"action":    string(request.Action),
			"case_slug": request.Scope.CaseSlug,
			"error":     "unknown plan action",
		})

		return chctui.PlanActionResult{}, errors.New("unknown plan action")
	}

	documents, err := chctui.LoadDocumentViews(e.operations.ctx, e.operations.projectContext.ProjectDir)
	if err != nil {
		e.operations.appendPlanGenerationLog("tui_plan_action_failed", map[string]string{
			"action":    string(request.Action),
			"case_slug": request.Scope.CaseSlug,
			"error":     err.Error(),
		})

		return chctui.PlanActionResult{}, err
	}

	plans, err := loadTUIPlans(e.operations)
	if err != nil {
		e.operations.appendPlanGenerationLog("tui_plan_action_failed", map[string]string{
			"action":    string(request.Action),
			"case_slug": request.Scope.CaseSlug,
			"error":     err.Error(),
		})

		return chctui.PlanActionResult{}, err
	}

	if message == "" {
		message = "Plans refreshed."
	}

	e.operations.appendPlanGenerationLog("tui_plan_action_finished", map[string]string{
		"action":     string(request.Action),
		"case_slug":  request.Scope.CaseSlug,
		"plan_count": fmt.Sprintf("%d", len(plans)),
	})

	return chctui.PlanActionResult{
		Message:   message,
		Documents: documents,
		Plans:     plans,
		Scope:     request.Scope,
	}, nil
}

func newTUITicketActionExecutor(
	operations service,
) tuiTicketActionExecutor {
	return tuiTicketActionExecutor{
		operations: operations,
	}
}

func (e tuiTicketActionExecutor) ExecuteTicketAction(
	request chctui.TicketActionRequest,
) (chctui.TicketActionResult, error) {
	switch request.Action {
	case chctui.TicketActionAddComment:
		result, err := e.operations.AddSelectedTicketComment(request.Key)
		if err != nil {
			return chctui.TicketActionResult{}, err
		}

		return ticketActionResult(result), nil
	case chctui.TicketActionMoveStatus:
		result, err := e.operations.MoveSelectedTicketStatus(request.Key, request.Status)
		if err != nil {
			return chctui.TicketActionResult{}, err
		}

		return ticketActionResult(result), nil
	case chctui.TicketActionToggleChecklist:
		return chctui.TicketActionResult{}, errors.New(
			"checklist toggle requires a safe checklist item writer and is not available from this TUI action yet",
		)
	case chctui.TicketActionAddChecklistItem:
		return chctui.TicketActionResult{}, errors.New(
			"checklist item creation requires a safe checklist item writer and is not available from this TUI action yet",
		)
	default:
		return chctui.TicketActionResult{}, errors.New("unknown ticket action")
	}
}

func ticketActionResult(result TicketOperationResult) chctui.TicketActionResult {
	return chctui.TicketActionResult{
		Message: result.Message,
		Tickets: result.Tickets,
	}
}

func loadTUIDumps(operations service) ([]chctui.DumpView, error) {
	result, err := operations.ListDumps()
	if err != nil {
		return nil, err
	}

	dumps := make([]chctui.DumpView, 0, len(result.Dumps))
	for index, dump := range result.Dumps {
		dumps = append(dumps, chctui.DumpView{
			Index:        index,
			NodeID:       dump.NodeID,
			UseCaseTitle: dump.UseCaseTitle,
			Title:        dump.Title,
			Status:       dump.Status,
			Scope:        chctui.DumpScope{CaseSlug: dump.Scope.CaseSlug},
			RelativePath: dump.RelativePath,
			UpdatedAt:    dump.UpdatedAt,
			Preview:      dump.Content,
		})
	}

	return dumps, nil
}

func loadTUIQuizzes(operations service) ([]chctui.QuizView, error) {
	result, err := operations.ListQuizzes()
	if err != nil {
		return nil, err
	}

	quizzes := make([]chctui.QuizView, 0, len(result.Views))
	for index, quiz := range result.Views {
		quizzes = append(quizzes, chctui.QuizView{
			Index:                 index,
			Title:                 quiz.Title,
			Scope:                 chctui.QuizScope{CaseSlug: quiz.Scope.CaseSlug},
			Status:                quiz.Status,
			Mode:                  string(quiz.Mode),
			RawPath:               quiz.RawPath,
			CuratedPath:           quiz.CuratedPath,
			RuntimePath:           quiz.RuntimePath,
			QuestionBudget:        quiz.QuestionBudget,
			QuestionsAsked:        quiz.QuestionsAsked,
			AnswersRecorded:       quiz.AnswersRecorded,
			SkipsRecorded:         quiz.SkipsRecorded,
			CandidateCount:        quiz.CandidateCount,
			BranchReturnReason:    quiz.BranchReturnReason,
			UpdatedAt:             quiz.UpdatedAt,
			ActiveQuestionSummary: quiz.ActiveQuestionSummary,
			TrailSummary:          quiz.TrailSummary,
			CuratedPreview:        quiz.CuratedPreview,
		})
	}

	return quizzes, nil
}

func loadTUISpecs(operations service) ([]chctui.SpecView, error) {
	result, err := operations.ListSpecs()
	if err != nil {
		return nil, err
	}

	specs := make([]chctui.SpecView, 0, len(result.Specs))
	for index, spec := range result.Specs {
		specs = append(specs, chctui.SpecView{
			Index:        index,
			UseCaseTitle: spec.UseCaseTitle,
			Title:        spec.Title,
			Status:       spec.Status,
			Scope:        chctui.QuizScope{CaseSlug: spec.Scope.CaseSlug},
			RelativePath: spec.RelativePath,
			UpdatedAt:    spec.UpdatedAt,
			Content:      spec.Content,
		})
	}

	return specs, nil
}

func loadTUIPlans(operations service) ([]chctui.PlanView, error) {
	result, err := operations.ListPlans()
	if err != nil {
		return nil, err
	}

	plans := make([]chctui.PlanView, 0, len(result.Plans))
	for index, plan := range result.Plans {
		plans = append(plans, chctui.PlanView{
			Index:        index,
			UseCaseTitle: plan.UseCaseTitle,
			Title:        plan.Title,
			Status:       plan.Status,
			Scope:        chctui.QuizScope{CaseSlug: plan.Scope.CaseSlug},
			RelativePath: plan.RelativePath,
			UpdatedAt:    plan.UpdatedAt,
			Content:      plan.Content,
		})
	}

	return plans, nil
}

type tuiDumpActionExecutor struct {
	operations service
}

func newTUIDumpActionExecutor(
	operations service,
) tuiDumpActionExecutor {
	return tuiDumpActionExecutor{
		operations: operations,
	}
}

type tuiQuizActionExecutor struct {
	operations service
}

func newTUIQuizActionExecutor(
	operations service,
) tuiQuizActionExecutor {
	return tuiQuizActionExecutor{
		operations: operations,
	}
}

func (e tuiQuizActionExecutor) ExecuteQuizAction(
	request chctui.QuizActionRequest,
) (chctui.QuizActionResult, error) {
	operation, ok := e.quizOperation(request.Action)
	if !ok {
		return chctui.QuizActionResult{}, errors.New("unknown quiz action")
	}

	message, err := operation(request)
	if err != nil {
		return chctui.QuizActionResult{}, err
	}

	quizzes, err := loadTUIQuizzes(e.operations)
	if err != nil {
		return chctui.QuizActionResult{}, err
	}

	return chctui.QuizActionResult{
		Message:       message,
		Quizzes:       quizzes,
		SelectedScope: request.Scope,
	}, nil
}

func (e tuiQuizActionExecutor) quizOperation(
	action chctui.QuizActionID,
) (func(chctui.QuizActionRequest) (string, error), bool) {
	operations := map[chctui.QuizActionID]func(chctui.QuizActionRequest) (string, error){
		chctui.QuizActionStart:    e.executeQuizOperationStart,
		chctui.QuizActionResume:   e.executeQuizOperationResume,
		chctui.QuizActionAnswer:   e.executeQuizOperationAnswer,
		chctui.QuizActionSkip:     e.executeQuizOperationSkip,
		chctui.QuizActionComplete: e.executeQuizOperationComplete,
		chctui.QuizActionCancel:   e.executeQuizOperationCancel,
		chctui.QuizActionRefresh:  executeQuizOperationRefresh,
		chctui.QuizActionDeepMode: e.executeQuizOperationMode,
		chctui.QuizActionFastMode: e.executeQuizOperationMode,
	}

	operation, ok := operations[action]

	return operation, ok
}

func executeQuizOperationRefresh(_ chctui.QuizActionRequest) (string, error) {
	return "Quizzes refreshed.", nil
}

func (e tuiQuizActionExecutor) executeQuizOperationStart(request chctui.QuizActionRequest) (string, error) {
	result, err := e.operations.StartQuiz(QuizStartRequest{
		Scope: chc.QuizScope{CaseSlug: request.Scope.CaseSlug},
		Mode:  chc.QuizModeDeep,
	})
	if err != nil {
		return "", err
	}

	return result.Message, nil
}

func (e tuiQuizActionExecutor) executeQuizOperationResume(request chctui.QuizActionRequest) (string, error) {
	result, err := e.operations.ResumeQuiz(QuizResumeRequest{
		Scope: chc.QuizScope{CaseSlug: request.Scope.CaseSlug},
	})
	if err != nil {
		return "", err
	}

	return result.Message, nil
}

func (e tuiQuizActionExecutor) executeQuizOperationAnswer(request chctui.QuizActionRequest) (string, error) {
	result, err := e.operations.AnswerQuiz(QuizAnswerRequest{
		Scope:  chc.QuizScope{CaseSlug: request.Scope.CaseSlug},
		Answer: request.Answer,
	})
	if err != nil {
		return "", err
	}

	return result.Message, nil
}

func (e tuiQuizActionExecutor) executeQuizOperationSkip(request chctui.QuizActionRequest) (string, error) {
	result, err := e.operations.SkipQuiz(QuizSkipRequest{
		Scope: chc.QuizScope{CaseSlug: request.Scope.CaseSlug},
	})
	if err != nil {
		return "", err
	}

	return result.Message, nil
}

func (e tuiQuizActionExecutor) executeQuizOperationComplete(request chctui.QuizActionRequest) (string, error) {
	result, err := e.operations.CompleteQuiz(QuizCompleteRequest{
		Scope: chc.QuizScope{CaseSlug: request.Scope.CaseSlug},
	})
	if err != nil {
		return "", err
	}

	return result.Message, nil
}

func (e tuiQuizActionExecutor) executeQuizOperationCancel(request chctui.QuizActionRequest) (string, error) {
	result, err := e.operations.CancelQuiz(QuizCancelRequest{
		Scope: chc.QuizScope{CaseSlug: request.Scope.CaseSlug},
	})
	if err != nil {
		return "", err
	}

	return result.Message, nil
}

func (e tuiQuizActionExecutor) executeQuizOperationMode(request chctui.QuizActionRequest) (string, error) {
	mode := chc.QuizMode(request.Mode)

	result, err := e.operations.SwitchQuizMode(QuizModeRequest{
		Scope: chc.QuizScope{CaseSlug: request.Scope.CaseSlug},
		Mode:  mode,
	})
	if err != nil {
		return "", err
	}

	return result.Message, nil
}

func (e tuiDumpActionExecutor) ExecuteDumpAction(
	request chctui.DumpActionRequest,
) (chctui.DumpActionResult, error) {
	switch request.Action {
	case chctui.DumpActionCreate:
		return e.executeCreateDump(request)
	case chctui.DumpActionAppend:
		return e.executeAppendDump(request)
	case chctui.DumpActionEdit:
		return e.executeEditDump(request)
	case chctui.DumpActionSave:
		return e.executeSaveDump(request)
	case chctui.DumpActionExternalEdit:
		return e.executeExternalEditDump(request)
	case chctui.DumpActionRefresh:
		return e.executeRefreshDumps(request)
	default:
		return chctui.DumpActionResult{}, errors.New("unknown dump action")
	}
}

func (e tuiDumpActionExecutor) executeCreateDump(
	request chctui.DumpActionRequest,
) (chctui.DumpActionResult, error) {
	result, err := e.operations.CreateDump(DumpCreateRequest{
		Title:   request.Title,
		Content: request.Content,
		Scope:   DumpScope{CaseSlug: request.Scope.CaseSlug},
	})
	if err != nil {
		return chctui.DumpActionResult{}, err
	}

	return dumpActionResult(result, e.operations)
}

func (e tuiDumpActionExecutor) executeAppendDump(
	request chctui.DumpActionRequest,
) (chctui.DumpActionResult, error) {
	result, err := e.operations.AppendDump(DumpAppendRequest{
		RelativePath: request.Path,
		Content:      request.Content,
	})
	if err != nil {
		return chctui.DumpActionResult{}, err
	}

	return dumpActionResult(result, e.operations)
}

func (e tuiDumpActionExecutor) executeEditDump(
	request chctui.DumpActionRequest,
) (chctui.DumpActionResult, error) {
	result, err := e.operations.ReadDump(DumpReadRequest{RelativePath: request.Path})
	if err != nil {
		return chctui.DumpActionResult{}, err
	}

	return chctui.DumpActionResult{
		Message:      result.Message,
		Content:      result.Content,
		SelectedPath: request.Path,
	}, nil
}

func (e tuiDumpActionExecutor) executeSaveDump(
	request chctui.DumpActionRequest,
) (chctui.DumpActionResult, error) {
	result, err := e.operations.SaveDump(DumpSaveRequest{
		RelativePath: request.Path,
		Content:      request.Content,
	})
	if err != nil {
		return chctui.DumpActionResult{}, err
	}

	actionResult, err := dumpActionResult(result, e.operations)
	if err != nil {
		return chctui.DumpActionResult{}, err
	}

	actionResult.SelectedPath = request.Path

	return actionResult, nil
}

func (e tuiDumpActionExecutor) executeExternalEditDump(
	request chctui.DumpActionRequest,
) (chctui.DumpActionResult, error) {
	result, err := e.operations.OpenDump(DumpOpenRequest{RelativePath: request.Path})
	if err != nil {
		return chctui.DumpActionResult{}, err
	}

	return dumpActionResult(result, e.operations)
}

func (e tuiDumpActionExecutor) executeRefreshDumps(
	request chctui.DumpActionRequest,
) (chctui.DumpActionResult, error) {
	dumps, err := loadTUIDumps(e.operations)
	if err != nil {
		return chctui.DumpActionResult{}, err
	}

	return chctui.DumpActionResult{
		Message:      "Edited dump " + request.Path + ".",
		Dumps:        dumps,
		SelectedPath: request.Path,
	}, nil
}

func dumpActionResult(result DumpOperationResult, operations service) (chctui.DumpActionResult, error) {
	dumps, err := loadTUIDumps(operations)
	if err != nil {
		return chctui.DumpActionResult{}, err
	}

	actionResult := chctui.DumpActionResult{
		Message: result.Message,
		Dumps:   dumps,
	}

	if result.EditorCommand != nil {
		actionResult.EditorCommand = exec.CommandContext(
			operations.ctx,
			result.EditorCommand.Program,
			result.EditorCommand.Args...,
		)
		actionResult.EditorCommand.Dir = result.EditorCommand.Dir
		actionResult.EditorCommandDescription = result.EditorCommand.Description
	}

	return actionResult, nil
}

type tuiWorkActionExecutor struct {
	operations service
}

func newTUIWorkActionExecutor(
	operations service,
) tuiWorkActionExecutor {
	return tuiWorkActionExecutor{
		operations: operations,
	}
}

func (e tuiWorkActionExecutor) ExecuteWorkAction(
	request chctui.WorkActionRequest,
) (chctui.WorkActionResult, error) {
	switch request.Action {
	case chctui.WorkActionStartTicket:
		result, err := e.operations.StartSelectedTicket(request.TicketKey)
		if err != nil {
			return chctui.WorkActionResult{}, err
		}

		return e.workActionResult(result)
	case chctui.WorkActionStartReadyLoop:
		result, err := e.operations.StartReadyTicketLoop()
		if err != nil {
			return chctui.WorkActionResult{}, err
		}

		return e.workActionResult(result)
	case chctui.WorkActionRequestStop:
		result, err := e.operations.RequestDaemonStop()
		if err != nil {
			return chctui.WorkActionResult{}, err
		}

		return e.workActionResult(result)
	default:
		return chctui.WorkActionResult{}, errors.New("unknown work action")
	}
}

func (e tuiWorkActionExecutor) workActionResult(
	result WorkOperationResult,
) (chctui.WorkActionResult, error) {
	observability, err := loadTUIObservability(e.operations.ctx, e.operations.env, e.operations.projectContext)
	if err != nil {
		return chctui.WorkActionResult{}, err
	}

	return chctui.WorkActionResult{
		Message:       result.Message,
		Tickets:       result.Tickets,
		Observability: observability,
	}, nil
}
