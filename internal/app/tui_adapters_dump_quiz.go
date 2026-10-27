package app

import (
	"errors"
	"os/exec"

	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	"forge.adrianpk.com/chr/internal/tui"
)

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
	request tui.QuizActionRequest,
) (tui.QuizActionResult, error) {
	operation, ok := e.quizOperation(request.Action)
	if !ok {
		return tui.QuizActionResult{}, errors.New("unknown quiz action")
	}

	message, err := operation(request)
	if err != nil {
		return tui.QuizActionResult{}, err
	}

	quizzes, err := loadTUIQuizzes(e.operations)
	if err != nil {
		return tui.QuizActionResult{}, err
	}

	return tui.QuizActionResult{
		Message:       message,
		Quizzes:       quizzes,
		SelectedScope: request.Scope,
	}, nil
}

func (e tuiQuizActionExecutor) quizOperation(
	action tui.QuizActionID,
) (func(tui.QuizActionRequest) (string, error), bool) {
	operations := map[tui.QuizActionID]func(tui.QuizActionRequest) (string, error){
		tui.QuizActionStart:    e.executeQuizOperationStart,
		tui.QuizActionResume:   e.executeQuizOperationResume,
		tui.QuizActionAnswer:   e.executeQuizOperationAnswer,
		tui.QuizActionSkip:     e.executeQuizOperationSkip,
		tui.QuizActionComplete: e.executeQuizOperationComplete,
		tui.QuizActionCancel:   e.executeQuizOperationCancel,
		tui.QuizActionRefresh:  executeQuizOperationRefresh,
		tui.QuizActionDeepMode: e.executeQuizOperationMode,
		tui.QuizActionFastMode: e.executeQuizOperationMode,
	}

	operation, ok := operations[action]

	return operation, ok
}

func executeQuizOperationRefresh(_ tui.QuizActionRequest) (string, error) {
	return "Quizzes refreshed.", nil
}

func (e tuiQuizActionExecutor) executeQuizOperationStart(request tui.QuizActionRequest) (string, error) {
	result, err := e.operations.StartQuiz(QuizStartRequest{
		Scope:   quizcore.Scope{CaseSlug: request.Scope.CaseSlug},
		Binding: tuiQuizBinding(request.Binding),
		Mode:    quizcore.ModeDeep,
	})
	if err != nil {
		return "", err
	}

	return result.Message, nil
}

func (e tuiQuizActionExecutor) executeQuizOperationResume(request tui.QuizActionRequest) (string, error) {
	result, err := e.operations.ResumeQuiz(QuizResumeRequest{
		Scope:   quizcore.Scope{CaseSlug: request.Scope.CaseSlug},
		Binding: tuiQuizBinding(request.Binding),
	})
	if err != nil {
		return "", err
	}

	return result.Message, nil
}

func (e tuiQuizActionExecutor) executeQuizOperationAnswer(request tui.QuizActionRequest) (string, error) {
	result, err := e.operations.AnswerQuiz(QuizAnswerRequest{
		Scope:   quizcore.Scope{CaseSlug: request.Scope.CaseSlug},
		Binding: tuiQuizBinding(request.Binding),
		Answer:  request.Answer,
	})
	if err != nil {
		return "", err
	}

	return result.Message, nil
}

func (e tuiQuizActionExecutor) executeQuizOperationSkip(request tui.QuizActionRequest) (string, error) {
	result, err := e.operations.SkipQuiz(QuizSkipRequest{
		Scope:   quizcore.Scope{CaseSlug: request.Scope.CaseSlug},
		Binding: tuiQuizBinding(request.Binding),
	})
	if err != nil {
		return "", err
	}

	return result.Message, nil
}

func (e tuiQuizActionExecutor) executeQuizOperationComplete(request tui.QuizActionRequest) (string, error) {
	result, err := e.operations.CompleteQuiz(QuizCompleteRequest{
		Scope:   quizcore.Scope{CaseSlug: request.Scope.CaseSlug},
		Binding: tuiQuizBinding(request.Binding),
	})
	if err != nil {
		return "", err
	}

	return result.Message, nil
}

func (e tuiQuizActionExecutor) executeQuizOperationCancel(request tui.QuizActionRequest) (string, error) {
	result, err := e.operations.CancelQuiz(QuizCancelRequest{
		Scope:   quizcore.Scope{CaseSlug: request.Scope.CaseSlug},
		Binding: tuiQuizBinding(request.Binding),
	})
	if err != nil {
		return "", err
	}

	return result.Message, nil
}

func (e tuiQuizActionExecutor) executeQuizOperationMode(request tui.QuizActionRequest) (string, error) {
	mode := quizcore.Mode(request.Mode)

	result, err := e.operations.SwitchQuizMode(QuizModeRequest{
		Scope:   quizcore.Scope{CaseSlug: request.Scope.CaseSlug},
		Binding: tuiQuizBinding(request.Binding),
		Mode:    mode,
	})
	if err != nil {
		return "", err
	}

	return result.Message, nil
}

func (e tuiDumpActionExecutor) ExecuteDumpAction(
	request tui.DumpActionRequest,
) (tui.DumpActionResult, error) {
	switch request.Action {
	case tui.DumpActionCreate:
		return e.executeCreateDump(request)
	case tui.DumpActionAppend:
		return e.executeAppendDump(request)
	case tui.DumpActionEdit:
		return e.executeEditDump(request)
	case tui.DumpActionSave:
		return e.executeSaveDump(request)
	case tui.DumpActionExternalEdit:
		return e.executeExternalEditDump(request)
	case tui.DumpActionRefresh:
		return e.executeRefreshDumps(request)
	default:
		return tui.DumpActionResult{}, errors.New("unknown dump action")
	}
}

func (e tuiDumpActionExecutor) executeCreateDump(
	request tui.DumpActionRequest,
) (tui.DumpActionResult, error) {
	result, err := e.operations.CreateDump(DumpCreateRequest{
		OperationID: dumpRequestID(request.OperationID), ExpectedRevision: request.Revision,
		Title:   request.Title,
		Content: request.Content,
		Scope:   DumpScope{CaseSlug: request.Scope.CaseSlug},
	})
	if err != nil {
		return tui.DumpActionResult{}, err
	}

	return dumpActionResult(result, e.operations)
}

func (e tuiDumpActionExecutor) executeAppendDump(
	request tui.DumpActionRequest,
) (tui.DumpActionResult, error) {
	result, err := e.operations.AppendDump(DumpAppendRequest{
		OperationID: dumpRequestID(request.OperationID), ExpectedRevision: request.Revision,
		RelativePath: request.Path,
		Content:      request.Content,
	})
	if err != nil {
		return tui.DumpActionResult{}, err
	}

	return dumpActionResult(result, e.operations)
}

func (e tuiDumpActionExecutor) executeEditDump(
	request tui.DumpActionRequest,
) (tui.DumpActionResult, error) {
	result, err := e.operations.ReadDump(DumpReadRequest{RelativePath: request.Path})
	if err != nil {
		return tui.DumpActionResult{}, err
	}

	return tui.DumpActionResult{
		Message: result.Message,
		Content: result.Content, Revision: result.Revision,
		SelectedPath: result.RelativePath,
	}, nil
}

func (e tuiDumpActionExecutor) executeSaveDump(
	request tui.DumpActionRequest,
) (tui.DumpActionResult, error) {
	result, err := e.operations.SaveDump(DumpSaveRequest{
		OperationID: dumpRequestID(request.OperationID), ExpectedRevision: request.Revision,
		RelativePath: request.Path,
		Content:      request.Content,
	})
	if err != nil {
		return tui.DumpActionResult{}, err
	}

	actionResult, err := dumpActionResult(result, e.operations)
	if err != nil {
		return tui.DumpActionResult{}, err
	}

	actionResult.SelectedPath = result.RelativePath

	return actionResult, nil
}

func (e tuiDumpActionExecutor) executeExternalEditDump(
	request tui.DumpActionRequest,
) (tui.DumpActionResult, error) {
	result, err := e.operations.OpenDump(DumpOpenRequest{RelativePath: request.Path})
	if err != nil {
		return tui.DumpActionResult{}, err
	}

	return dumpActionResult(result, e.operations)
}

func (e tuiDumpActionExecutor) executeRefreshDumps(
	request tui.DumpActionRequest,
) (tui.DumpActionResult, error) {
	dumps, err := loadTUIDumps(e.operations)
	if err != nil {
		return tui.DumpActionResult{}, err
	}

	return tui.DumpActionResult{
		Message:      "Dumps refreshed.",
		Dumps:        dumps,
		SelectedPath: request.Path,
	}, nil
}

func dumpActionResult(result DumpOperationResult, operations service) (tui.DumpActionResult, error) {
	dumps, err := loadTUIDumps(operations)
	if err != nil && !result.Committed {
		return tui.DumpActionResult{}, err
	}

	if err != nil && !result.ProjectionStale {
		result.ProjectionStale = true
		result.Message += " Catalog refresh is pending."
	}

	actionResult := tui.DumpActionResult{
		Committed: result.Committed, ProjectionStale: result.ProjectionStale,
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

func tuiQuizBinding(binding tui.QuizActionBinding) quizcore.ActionBinding {
	return quizcore.ActionBinding{SessionID: binding.SessionID, QuestionID: binding.QuestionID, ExpectedRevision: binding.ExpectedRevision, OperationID: binding.OperationID}
}
