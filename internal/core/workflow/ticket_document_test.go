package workflow

import (
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func ticketEditorBinding(id string) workflowstate.MutationBinding {
	return workflowstate.MutationBinding{RequestID: id, Operator: "editor"}
}
