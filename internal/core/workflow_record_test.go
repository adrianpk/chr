package core

import (
	"time"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowTestTask(id string) workflowcore.Record {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	return workflowcore.Record{
		Identity: workflowstate.Identity{ID: id, ProjectID: "project-1", Namespace: "default", Revision: 1, CreatedAt: now, UpdatedAt: now},
		Kind:     "ticket", Path: "ticket/" + id + "/metadata.json", Body: []byte("# Original body\n"),
		Task: &workflowcore.Task{Schema: workflowcore.WorkflowSchema, Key: id, Status: string(ticketcore.StatusReadyForDev), Phase: "pending", ContractHash: "contract-1"},
	}
}
