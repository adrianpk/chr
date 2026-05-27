package chcapp

import (
	"os/exec"

	"forge.adrianpk.com/chc/internal/chctui"
)

type tuiSourceOpenExecutor struct {
	operations service
}

func newTUISourceOpenExecutor(operations service) tuiSourceOpenExecutor {
	return tuiSourceOpenExecutor{operations: operations}
}

func (e tuiSourceOpenExecutor) OpenSource(path string) (chctui.SourceOpenResult, error) {
	result, err := e.operations.OpenSource(path)
	if err != nil {
		return chctui.SourceOpenResult{}, err
	}

	if result.EditorCommand == nil {
		return chctui.SourceOpenResult{}, nil
	}

	command := exec.CommandContext(
		e.operations.ctx,
		result.EditorCommand.Program,
		result.EditorCommand.Args...,
	)
	command.Dir = result.EditorCommand.Dir

	return chctui.SourceOpenResult{
		EditorCommand:            command,
		EditorCommandDescription: result.EditorCommand.Description,
	}, nil
}
