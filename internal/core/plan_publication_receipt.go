package core

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// PlanPublicationResult identifies the retained revision from one committed
// operation. Reference is project-relative; binding paths are namespace-relative.
type PlanPublicationResult struct {
	OperationID string
	Reference   string
	Binding     generationcore.PlanRevisionBinding
}

type planPublicationReceipt struct {
	Schema, Digest, Checksum string
	Result                   PlanPublicationResult
}

// PlanPublicationLookup distinguishes a retained publication from expected absence.
// Result preserves any original partial observation when a lookup fails.
type PlanPublicationLookup struct {
	Result PlanPublicationResult
	Found  bool
}

// LookupPlanPublication recovers any pending authoring transaction and returns
// a committed retry without invoking a model or advancing the current revision.
func LookupPlanPublication(ctx context.Context, projectDir string, operation generationcore.PlanPublicationOperation) (lookup PlanPublicationLookup, err error) {
	err = WithAuthoringMutation(ctx, projectDir, func(context.Context) error {
		r := &fileAuthoringSessionRepository{projectDir: projectDir}
		lookup, err = r.planReceipt(operation)

		return err
	})

	return lookup, err
}

func (r *fileAuthoringSessionRepository) planOperationDigest(operation generationcore.PlanPublicationOperation) (string, error) {
	if err := authoringstate.ValidateSessionID(operation.ID); err != nil {
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

func (r *fileAuthoringSessionRepository) planReceipt(operation generationcore.PlanPublicationOperation) (PlanPublicationLookup, error) {
	var receipt planPublicationReceipt

	digest, err := r.planOperationDigest(operation)
	if err != nil {
		return PlanPublicationLookup{Result: receipt.Result}, err
	}

	encoded, err := r.authoringRead(authoringReceiptRef(operation.ID))
	if errors.Is(err, os.ErrNotExist) {
		return PlanPublicationLookup{Result: receipt.Result}, nil
	}

	if err != nil {
		return PlanPublicationLookup{Result: receipt.Result}, err
	}

	if err := authoringstate.Decode(encoded, &receipt); err != nil {
		return PlanPublicationLookup{Result: receipt.Result}, err
	}

	if err := validatePlanPublicationReceipt(receipt, operation.ID, digest); err != nil {
		return PlanPublicationLookup{Result: receipt.Result}, err
	}

	if err := r.validatePlanReceiptResult(operation.Scope, receipt.Result); err != nil {
		return PlanPublicationLookup{Result: receipt.Result}, err
	}

	return PlanPublicationLookup{Result: receipt.Result, Found: true}, nil
}

func (r *fileAuthoringSessionRepository) completedPlanOperation(plan generationcore.GeneratedPlan) (bool, error) {
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

	lookup, err := r.planReceipt(operation)

	return lookup.Found, err
}

func (r *fileAuthoringSessionRepository) planPublicationReceiptChange(paths GeneratedPlanPaths, plan generationcore.GeneratedPlan, changes []workflowstate.FileChange) ([]workflowstate.FileChange, error) {
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
		Binding generationcore.PlanRevisionBinding `json:"plan_binding"`
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

	return []workflowstate.FileChange{{Path: authoringReceiptRef(plan.Operation.ID), After: encoded}}, err
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
