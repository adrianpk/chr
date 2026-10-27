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

// PlanPublicationOperation binds retries to caller-owned intent, independently
// of model output, timestamps, or the newly observed target after a commit.
type PlanPublicationOperation struct {
	TicketDelivery     bool `json:",omitempty"`
	ID                 string
	Scope              QuizScope
	ExpectedRevision   string
	Replace            bool
	TrackerReplaceHash string
}

// PlanPublicationResult identifies the retained revision from one committed
// operation. Reference is project-relative; binding paths are namespace-relative.
type PlanPublicationResult struct {
	OperationID string
	Reference   string
	Binding     PlanRevisionBinding
}

type planPublicationReceipt struct {
	Schema, Digest, Checksum string
	Result                   PlanPublicationResult
}

// LookupPlanPublication recovers any pending authoring transaction and returns
// a committed retry without invoking a model or advancing the current revision.
func LookupPlanPublication(ctx context.Context, projectDir string, operation PlanPublicationOperation) (result PlanPublicationResult, found bool, err error) {
	err = WithAuthoringMutation(ctx, projectDir, func(context.Context) error {
		r := &fileAuthoringSessionRepository{projectDir: projectDir}
		result, found, err = r.planReceipt(operation)

		return err
	})

	return result, found, err
}

func (r *fileAuthoringSessionRepository) planOperationDigest(operation PlanPublicationOperation) (string, error) {
	if err := validateAuthoringSessionID(operation.ID); err != nil {
		return "", err
	}

	paths, err := GeneratedPlanLayout(r.projectDir, operation.Scope)
	if err != nil {
		return "", err
	}

	root, err := filepath.Abs(r.projectDir)
	if err != nil {
		return "", err
	}

	encoded, err := json.Marshal(operation)

	return contenthash.Sum(append([]byte(root+"\x00plan\x00"+paths.RelativePath+"\x00"), encoded...)), err
}

func (r *fileAuthoringSessionRepository) planReceipt(operation PlanPublicationOperation) (PlanPublicationResult, bool, error) {
	var receipt planPublicationReceipt

	digest, err := r.planOperationDigest(operation)
	if err != nil {
		return receipt.Result, false, err
	}

	encoded, err := r.authoringRead(authoringReceiptRef(operation.ID))
	if errors.Is(err, os.ErrNotExist) {
		return receipt.Result, false, nil
	}

	if err != nil {
		return receipt.Result, false, err
	}

	if err := decodeAuthoringState(encoded, &receipt); err != nil {
		return receipt.Result, false, err
	}

	if err := validatePlanPublicationReceipt(receipt, operation.ID, digest); err != nil {
		return receipt.Result, false, err
	}

	if err := r.validatePlanReceiptResult(operation.Scope, receipt.Result); err != nil {
		return receipt.Result, false, err
	}

	return receipt.Result, true, nil
}

func (r *fileAuthoringSessionRepository) completedPlanOperation(plan GeneratedPlan) (bool, error) {
	if plan.Operation == nil {
		return false, nil
	}

	operation := *plan.Operation
	if operation.Scope != plan.Scope || operation.Replace == plan.CreateOnly || operation.TrackerReplaceHash != plan.TrackerReplaceHash {
		return false, workflowstate.ErrConflict
	}

	if operation.ExpectedRevision != "" && operation.ExpectedRevision != plan.ExpectedRevision {
		return false, workflowstate.ErrConflict
	}

	_, found, err := r.planReceipt(operation)

	return found, err
}

func (r *fileAuthoringSessionRepository) planPublicationReceiptChange(paths GeneratedPlanPaths, plan GeneratedPlan, changes []workflowFileChange) ([]workflowFileChange, error) {
	if plan.Operation == nil {
		return nil, nil
	}

	if err := r.projectDumpOperationAvailable(plan.Operation.ID); err != nil {
		return nil, err
	}

	digest, err := r.planOperationDigest(*plan.Operation)
	if err != nil {
		return nil, err
	}

	var metadata struct {
		Binding PlanRevisionBinding `json:"plan_binding"`
	}

	for _, change := range changes {
		if change.Path == filepath.ToSlash(filepath.Join(filepath.Dir(paths.RelativePath), "metadata.json")) {
			if err := json.Unmarshal(change.After, &metadata); err != nil {
				return nil, err
			}
		}
	}

	receipt := planPublicationReceipt{Schema: "chr-plan-publication/v1", Digest: digest, Result: PlanPublicationResult{OperationID: plan.Operation.ID, Reference: paths.RelativePath, Binding: metadata.Binding}}

	encoded, err := json.Marshal(receipt)
	if err != nil {
		return nil, err
	}

	receipt.Checksum = contenthash.Sum(encoded)
	encoded, err = json.Marshal(receipt)

	return []workflowFileChange{{Path: authoringReceiptRef(plan.Operation.ID), After: encoded}}, err
}

func validatePlanPublicationReceipt(receipt planPublicationReceipt, operationID, digest string) error {
	checksum := receipt.Checksum
	receipt.Checksum = ""

	encoded, err := json.Marshal(receipt)
	if err != nil {
		return err
	}

	if receipt.Schema != "chr-plan-publication/v1" || receipt.Digest != digest || receipt.Result.OperationID != operationID || checksum != contenthash.Sum(encoded) {
		return workflowstate.ErrConflict
	}

	return nil
}
