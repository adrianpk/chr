package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type projectDumpMutationReceipt struct {
	Schema, Digest, Checksum string
	Result                   ProjectDumpMutationResult
}

func projectDumpReconciliationDigest(projectDir string, request ProjectDumpReconciliationRequest) (string, error) {
	// Root identity prevents receipts
	// copied between otherwise byte-identical projects from replaying there.
	encoded, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	return contenthash.Sum(append([]byte(projectDir+"\x00reconcile\x00"), encoded...)), nil
}

func projectDumpReceiptChange(digest string, result ProjectDumpMutationResult) (workflowFileChange, error) {
	receipt := projectDumpMutationReceipt{Schema: "chr-project-dump-receipt/v1", Digest: digest, Result: result}

	encoded, err := json.Marshal(receipt)
	if err != nil {
		return workflowFileChange{}, err
	}

	receipt.Checksum = contenthash.Sum(encoded)
	encoded, err = json.Marshal(receipt)

	return workflowFileChange{Path: authoringReceiptRef(result.OperationID), After: encoded}, err
}

func (r *fileAuthoringSessionRepository) projectDumpReceipt(operationID, digest string) (projectDumpReceiptLookup, error) {
	var receipt projectDumpMutationReceipt

	encoded, err := r.readAuthoringDocument(filepath.ToSlash(authoringReceiptRef(operationID)))
	if errors.Is(err, os.ErrNotExist) {
		return projectDumpReceiptLookup{Result: receipt.Result, Found: false}, nil
	}

	if err != nil {
		return projectDumpReceiptLookup{Result: receipt.Result, Found: false}, err
	}

	if err := decodeAuthoringState(encoded, &receipt); err != nil {
		return projectDumpReceiptLookup{Result: receipt.Result, Found: false}, err
	}

	checksum := receipt.Checksum
	receipt.Checksum = ""

	encoded, err = json.Marshal(receipt)
	if err != nil {
		return projectDumpReceiptLookup{Result: receipt.Result, Found: false}, err
	}

	if receipt.Schema != "chr-project-dump-receipt/v1" || receipt.Digest != digest || receipt.Result.OperationID != operationID || contenthash.Sum(encoded) != checksum {
		return projectDumpReceiptLookup{Result: receipt.Result, Found: false}, workflowstate.ErrConflict
	}

	return projectDumpReceiptLookup{Result: receipt.Result, Found: true}, nil
}

func (r *fileAuthoringSessionRepository) projectDumpOperationAvailable(operationID string) error {
	_, err := r.readAuthoringDocument(filepath.ToSlash(authoringRequestRef(operationID)))
	if err == nil {
		return workflowstate.ErrConflict
	}

	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return r.authoringStateCapacity("receipts")
}
