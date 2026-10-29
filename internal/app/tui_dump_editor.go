package app

import (
	"os/exec"

	"forge.adrianpk.com/chr/internal/tui"
)

func (e tuiDumpActionExecutor) OpenDumpDraftEditor(content string) (tui.DumpDraftEditor, error) {
	session, err := e.operations.OpenDumpDraftEditor(content)
	if err != nil {
		return tui.DumpDraftEditor{Outcome: operationResult("", err)}, err
	}

	command := exec.CommandContext(e.operations.ctx, session.Command.Program, session.Command.Args...)
	command.Dir = session.Command.Dir

	return tui.DumpDraftEditor{Command: command, Finish: func(runErr error) (tui.DumpActionResult, error) {
		content, finishErr := session.Finish(runErr)

		return tui.DumpActionResult{Content: content, Outcome: operationResult("External edits retained; Ctrl+S saves the dump.", finishErr)}, finishErr
	}}, nil
}
