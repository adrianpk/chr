package app

import (
	"errors"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

func (s service) authoringWorkContract(record core.WorkflowRecord, title string, request AuthoringWorkRequest, source core.AuthoringArtifactObservation) (core.WorkflowRecord, error) {
	config, err := core.LoadWorkspaceConfig(s.projectContext.ProjectDir)
	if err != nil {
		return core.WorkflowRecord{}, err
	}

	if strings.TrimSpace(config.Runtime.ValidationCommand) == "" {
		return core.WorkflowRecord{}, errors.New("native work preparation requires a configured validation command")
	}

	timeout := config.Work.ValidationTimeoutSeconds
	if timeout == 0 {
		timeout = 3600
	}

	// Keep the observed description intact rather than asking the model to
	// reinterpret its acceptance criteria or invent a validation command.
	contract := core.GeneratedWorkflowTask{ID: record.Task.Key, Title: title, Body: source.Body, Outcome: request.Instruction, Scope: source.Ref.Path, Acceptance: []string{source.Body}, Validation: core.WorkflowValidation{Required: true, Commands: [][]string{{"sh", "-c", config.Runtime.ValidationCommand}}, TimeoutSeconds: timeout}}

	return core.PrepareIndependentWorkflowTask(record, contract)
}
