package tui

func nativeWorkflowCommands() []Command {
	definitions := []struct {
		id, title string
		action    WorkActionID
	}{
		{"workflow-preview", "Preview bounded work scope", WorkActionWorkflowPreview},
		{"workflow-ticket", "Work native selected ticket", WorkActionWorkflowTicket},
		{"workflow-slice", "Work selected ticket's slice", WorkActionWorkflowSlice},
		{"workflow-set", "Work selected ticket's slice set", WorkActionWorkflowSet},
		{"workflow-case", "Work selected use case", WorkActionWorkflowCase},
		{"workflow-loose", "Work independent native tickets", WorkActionWorkflowLoose},
		{"workflow-status", "Inspect retained workflow status", WorkActionWorkflowStatus},
		{"workflow-next", "Select next retained workflow", WorkActionWorkflowNext},
		{"workflow-page", "Load next retained workflow page", WorkActionWorkflowPage},
		{"workflow-continue", "Continue selected workflow", WorkActionWorkflowContinue},
		{"workflow-pause", "Pause selected workflow", WorkActionWorkflowPause},
		{"workflow-cancel", "Cancel selected workflow", WorkActionWorkflowCancel},
		{"workflow-resume", "Resume selected workflow", WorkActionWorkflowResume},
		{"workflow-approve", "Approve displayed workflow binding", WorkActionWorkflowApprove},
		{"workflow-reconcile", "Reconcile selected workflow owner", WorkActionWorkflowReconcile},
	}

	commands := make([]Command, 0, len(definitions))
	for _, definition := range definitions {
		commands = append(commands, Command{ID: definition.id, Title: definition.title, Description: "Inspect the shared operation before confirmation.", Action: CommandActionPlanWork, Work: definition.action})
	}

	return commands
}
