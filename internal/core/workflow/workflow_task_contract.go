package workflow

import (
	"encoding/json"
	"maps"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// PrepareIndependentWorkflowTask seals an executable description without making
// a backlog ticket ready or admitting it, and leaves the input record unchanged.
func PrepareIndependentWorkflowTask(record Record, contract generationcore.GeneratedWorkflowTask) (Record, error) {
	if !workflowIndependentBacklog(record) || len(contract.DependsOn) != 0 {
		return Record{}, workflowstate.ErrConflict
	}

	if err := generationcore.ValidateWorkflowPlanTask(contract); err != nil {
		return Record{}, err
	}

	encoded, err := json.Marshal(contract)
	if err != nil {
		return Record{}, err
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

func workflowIndependentBacklog(record Record) bool {
	return record.Kind == "ticket" && record.Task != nil && record.Task.SetID == "" && record.Task.SliceID == "" && record.Task.Status == string(ticketcore.StatusBacklog) && record.Task.Phase == "pending" && record.Task.AttemptID == "" && len(record.Task.DependsOn) == 0
}
