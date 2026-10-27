package core

import (
	"encoding/json"
	"maps"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// PrepareIndependentWorkflowTask seals an executable description without making
// a backlog ticket ready or admitting it, and leaves the input record unchanged.
func PrepareIndependentWorkflowTask(record WorkflowRecord, contract GeneratedWorkflowTask) (WorkflowRecord, error) {
	if !workflowIndependentBacklog(record) || len(contract.DependsOn) != 0 {
		return WorkflowRecord{}, workflowstate.ErrConflict
	}

	if err := validateWorkflowPlanTask(contract); err != nil {
		return WorkflowRecord{}, err
	}

	encoded, err := json.Marshal(contract)
	if err != nil {
		return WorkflowRecord{}, err
	}

	task := *record.Task
	task.ContractHash = contenthash.Sum(append([]byte("\x00\x00"), encoded...))

	record.Task, record.Domain = &task, maps.Clone(record.Domain)
	if record.Domain == nil {
		record.Domain = map[string]json.RawMessage{}
	}

	record.Domain["execution_contract"] = encoded

	return record, nil
}

func workflowIndependentBacklog(record WorkflowRecord) bool {
	return record.Kind == "ticket" && record.Task != nil && record.Task.SetID == "" && record.Task.SliceID == "" && record.Task.Status == TicketStatusBacklog && record.Task.Phase == "pending" && record.Task.AttemptID == "" && len(record.Task.DependsOn) == 0
}
