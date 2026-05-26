package chcapp

import (
	"errors"
	"os/exec"

	"forge.adrianpk.com/chc/internal/chc"
	"forge.adrianpk.com/chc/internal/chctui"
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
		Message:      "Dumps refreshed.",
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
