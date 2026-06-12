package app

import (
	"os/exec"

	"forge.adrianpk.com/chc/internal/tui"
)

type tuiSourceOpenExecutor struct {
	operations service
}

func newTUISourceOpenExecutor(operations service) tuiSourceOpenExecutor {
	return tuiSourceOpenExecutor{operations: operations}
}

func (e tuiSourceOpenExecutor) OpenSource(path string) (tui.SourceOpenResult, error) {
	result, err := e.operations.OpenSource(path)
	if err != nil {
		return tui.SourceOpenResult{}, err
	}

	if result.EditorCommand == nil {
		return tui.SourceOpenResult{}, nil
	}

	command := exec.CommandContext(
		e.operations.ctx,
		result.EditorCommand.Program,
		result.EditorCommand.Args...,
	)
	command.Dir = result.EditorCommand.Dir

	return tui.SourceOpenResult{
		EditorCommand:            command,
		EditorCommandDescription: result.EditorCommand.Description,
	}, nil
}
