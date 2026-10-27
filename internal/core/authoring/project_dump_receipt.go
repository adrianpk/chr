package authoring

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
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

func projectDumpReceiptChange(digest string, result ProjectDumpMutationResult) (workflowstate.FileChange, error) {
	receipt := projectDumpMutationReceipt{Schema: "chr-project-dump-receipt/v1", Digest: digest, Result: result}

	encoded, err := json.Marshal(receipt)
	if err != nil {
		return workflowstate.FileChange{}, err
	}

	receipt.Checksum = contenthash.Sum(encoded)
	encoded, err = json.Marshal(receipt)

	return workflowstate.FileChange{Path: authoringReceiptRef(result.OperationID), After: encoded}, err
}

func (r *fileSessionRepository) projectDumpReceipt(operationID, digest string) (projectDumpReceiptLookup, error) {
	var receipt projectDumpMutationReceipt

	encoded, err := r.readAuthoringDocument(filepath.ToSlash(authoringReceiptRef(operationID)))
	if errors.Is(err, os.ErrNotExist) {
		return projectDumpReceiptLookup{Result: receipt.Result, Found: false}, nil
	}

	if err != nil {
		return projectDumpReceiptLookup{Result: receipt.Result, Found: false}, err
	}

	if err := authoringstate.Decode(encoded, &receipt); err != nil {
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

func (r *fileSessionRepository) projectDumpOperationAvailable(operationID string) error {
	_, err := r.readAuthoringDocument(filepath.ToSlash(authoringRequestRef(operationID)))
	if err == nil {
		return workflowstate.ErrConflict
	}

	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return r.authoringStateCapacity("receipts")
}
