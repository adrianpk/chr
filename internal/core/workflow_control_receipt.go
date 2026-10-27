package core

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowControlReceipt struct {
	Schema    string                `json:"schema"`
	State     string                `json:"state"`
	ProjectID string                `json:"project_id"`
	Namespace string                `json:"namespace"`
	Operator  string                `json:"operator"`
	Result    WorkflowControlResult `json:"result"`
	Checksum  string                `json:"checksum"`
}

type workflowControlReservation struct {
	Schema           string                  `json:"schema"`
	State            string                  `json:"state"`
	ProjectID        string                  `json:"project_id"`
	Namespace        string                  `json:"namespace"`
	Binding          WorkflowMutationBinding `json:"binding"`
	Action           string                  `json:"action"`
	InvocationID     string                  `json:"invocation_id"`
	ArgumentDigest   string                  `json:"argument_digest"`
	ExpectedRevision uint64                  `json:"expected_revision"`
	ReceiptRef       string                  `json:"receipt_ref"`
}

type workflowControlCommit struct {
	receipt workflowControlReceipt
	check   WorkflowAdmissionCheck
}

func (commit *workflowControlCommit) files(ctx context.Context, revision uint64) ([]workflowFileChange, error) {
	if commit == nil {
		return nil, nil
	}

	if err := workflowCheckAuthorization(ctx, commit.check); err != nil {
		return nil, err
	}

	commit.receipt.Result.Revision = revision
	receipt := commit.receipt

	data, err := json.Marshal(receipt)
	if err != nil {
		return nil, err
	}

	receipt.Checksum = contenthash.Sum(data)
	data, err = json.Marshal(receipt)

	return []workflowFileChange{{Path: receipt.Result.ReceiptRef, After: data}}, err
}

func (s *WorkflowStore) readWorkflowControlReceipt(request WorkflowControlMutationRequest, digest string) (WorkflowControlResult, bool, error) {
	path, err := s.safePath(workflowMutationReceiptPath(request.Binding.RequestID))
	if err != nil {
		return WorkflowControlResult{}, false, err
	}

	data, err := readWorkflowFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return WorkflowControlResult{}, false, nil
	}

	if err != nil {
		return WorkflowControlResult{}, false, err
	}

	var receipt workflowControlReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return WorkflowControlResult{}, false, err
	}

	if err := s.validateWorkflowControlReceipt(receipt, request, digest); err != nil {
		return WorkflowControlResult{}, false, err
	}

	return receipt.Result, true, nil
}

func (s *WorkflowStore) validateWorkflowControlReceipt(receipt workflowControlReceipt, request WorkflowControlMutationRequest, digest string) error {
	checksum := receipt.Checksum
	receipt.Checksum = ""

	encoded, err := json.Marshal(receipt)
	if err != nil {
		return err
	}

	expected := WorkflowControlResult{RequestID: request.Binding.RequestID, Action: request.Action, InvocationID: request.ID, ArgumentDigest: digest, ReceiptRef: workflowMutationReceiptPath(request.Binding.RequestID), PriorRevision: request.ExpectedRevision, Revision: request.ExpectedRevision + 1}
	if receipt.Schema != "chr-workflow-control/v1" || receipt.State != "confirmed" || receipt.ProjectID != s.projectID || receipt.Namespace != s.layout.Namespace() || receipt.Operator != request.Binding.Operator || checksum != contenthash.Sum(encoded) || receipt.Result != expected {
		return workflowstate.ErrConflict
	}

	return nil
}

func (s *WorkflowStore) reserveWorkflowControlBinding(request WorkflowControlMutationRequest, digest string) error {
	reservation := workflowControlReservation{Schema: "chr-workflow-control-binding/v1", State: "reserved", ProjectID: s.projectID, Namespace: s.layout.Namespace(), Binding: request.Binding, Action: request.Action, InvocationID: request.ID, ArgumentDigest: digest, ExpectedRevision: request.ExpectedRevision, ReceiptRef: workflowMutationReceiptPath(request.Binding.RequestID)}

	path, err := s.safePath(filepath.Join("run", "operation-bindings", contenthash.Sum([]byte(request.Binding.RequestID))+".json"))
	if err != nil {
		return err
	}

	data, err := readWorkflowFile(path)
	if err == nil {
		var prior workflowControlReservation
		if err := json.Unmarshal(data, &prior); err != nil {
			return err
		}

		if prior != reservation {
			return workflowstate.ErrConflict
		}

		return nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err := s.checkMutationCapacity(); err != nil {
		return err
	}

	encoded, err := json.Marshal(reservation)
	if err != nil {
		return err
	}

	return atomicWorkflowFile(path, encoded)
}

func (s *WorkflowStore) retainWorkflowControlNoop(ctx context.Context, run workflowInvocationFile, control *workflowControlCommit) error {
	if control == nil {
		return nil
	}

	files, err := control.files(ctx, run.Revision+1)
	if err != nil {
		return err
	}

	return s.commitWorkflowPullRequestFiles(&run, files, workflowFileChange{})
}

func (s *WorkflowStore) completeWorkflowControlRetry(ctx context.Context, run workflowInvocationFile, control *workflowControlCommit, failure error) error {
	if failure != nil {
		return failure
	}

	return s.retainWorkflowControlNoop(ctx, run, control)
}

func (s *WorkflowStore) commitWorkflowControlBudget(ctx context.Context, run *workflowInvocationFile, budget *WorkflowIntegrationBudget, control *workflowControlCommit) error {
	files, err := control.files(ctx, run.Revision+1)
	if err != nil {
		return err
	}

	return s.commitWorkflowIntegrationBudget(ctx, run, budget, files...)
}
