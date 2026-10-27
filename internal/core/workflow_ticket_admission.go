package core

import (
	"context"
	"encoding/json"
	"reflect"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowTicketAdmission binds the explicit transition of one independent
// backlog ticket to ready, together with its existing executable contract.
type WorkflowTicketAdmission struct {
	Key          string                   `json:"key"`
	Revision     uint64                   `json:"revision"`
	ContractHash string                   `json:"contract_hash"`
	Validation   workflowstate.Validation `json:"validation"`
}

func validateWorkflowTicketAdmission(guard *WorkflowAdmissionGuard, selection WorkflowPreviewRequest) error {
	ticket := guard.ReadyTicket
	if ticket == nil {
		return nil
	}

	if guard.Selected == nil || ticket.Key != guard.Selected.Key || ticket.Revision != guard.Selected.Revision || !validWorkflowNativeFingerprint(ticket.ContractHash) || selection.Selector.Extent != "ticket" || selection.Selector.Identity != ticket.Key {
		return workflowstate.ErrConflict
	}

	return workflowstate.ValidateGate(ticket.Validation)
}

// PreviewWorkflowTicketAdmission observes a proposed readiness transition. It
// never changes the ticket or admits work; non-backlog tickets need no transition.
func (s *WorkflowStore) PreviewWorkflowTicketAdmission(ctx context.Context, member WorkflowPreviewMember) (result *WorkflowTicketAdmission, err error) {
	err = s.withRepositoryLock(ctx, s.requireNoWorkflowTransition, func() error {
		record, err := s.readRecord(member.Path)
		if err != nil {
			return err
		}

		if record.Task == nil || record.ProjectID != s.projectID || record.Namespace != s.layout.Namespace() || record.Task.Key != member.Key || record.Revision != member.Revision || record.SourceHash != member.SourceHash || record.Task.ContractHash != member.ContractHash {
			return workflowstate.ErrConflict
		}

		if !workflowIndependentBacklog(record) {
			return nil
		}

		result, err = workflowTicketAdmission(record)

		return err
	})

	return result, err
}

func workflowTicketAdmission(record WorkflowRecord) (*WorkflowTicketAdmission, error) {
	if err := validateWorkflowClaimContract(record); err != nil {
		return nil, err
	}

	var contract generationcore.GeneratedWorkflowTask
	if err := json.Unmarshal(record.Domain["execution_contract"], &contract); err != nil {
		return nil, err
	}

	return &WorkflowTicketAdmission{Key: record.Task.Key, Revision: record.Revision, ContractHash: record.Task.ContractHash, Validation: contract.Validation}, nil
}

func (s *WorkflowStore) workflowTicketAdmissionFiles(run *workflowInvocationFile) ([]workflowstate.FileChange, error) {
	if run.Admission == nil || run.Admission.ReadyTicket == nil {
		return nil, nil
	}

	if len(run.Members) != 1 {
		return nil, workflowstate.ErrConflict
	}

	member := run.Members[0]

	record, err := s.readRecord(member.Admission.Path)
	if err != nil {
		return nil, err
	}

	if !workflowIndependentBacklog(record) || !workflowMemberImageMatches(record, member, run.WorkflowIdentity) {
		return nil, workflowstate.ErrConflict
	}

	observed, err := workflowTicketAdmission(record)
	if err != nil {
		return nil, err
	}

	if !reflect.DeepEqual(observed, run.Admission.ReadyTicket) {
		return nil, workflowstate.ErrConflict
	}

	record.Task.Status = TicketStatusReadyForDev

	return s.workflowClaimRecordChanges(run, &record)
}
