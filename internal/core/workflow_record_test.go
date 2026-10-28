package core

import (
	"time"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowTestTask(id string) workflowstate.Record {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	return workflowstate.Record{
		Identity: workflowstate.Identity{ID: id, ProjectID: "project-1", Namespace: "default", Revision: 1, CreatedAt: now, UpdatedAt: now},
		Kind:     "ticket", Path: "ticket/" + id + "/metadata.json", Body: []byte("# Original body\n"),
		Task: &workflowstate.Task{Schema: workflowstate.WorkflowSchema, Key: id, Status: string(ticketcore.StatusReadyForDev), Phase: "pending", ContractHash: "contract-1"},
	}
}
