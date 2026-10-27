package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type quizActionReceipt struct {
	Schema, Digest, Checksum string
	Result                   QuizActionResult
	Revision                 string
}

func (r *fileAuthoringSessionRepository) quizActionDigest(request QuizActionRequest) (string, error) {
	root, err := filepath.Abs(r.projectDir)
	if err != nil {
		return "", err
	}

	paths, err := QuizArtifactLayout(r.projectDir, request.Scope)
	if err != nil {
		return "", err
	}

	encoded, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	return contenthash.Sum(append([]byte(root+"\x00quiz\x00"+paths.RawRelativePath+"\x00"), encoded...)), nil
}

func quizReceiptChange(digest string, result QuizActionResult) (workflowFileChange, error) {
	receipt := quizActionReceipt{Schema: "chr-quiz-action/v1", Digest: digest, Result: result, Revision: result.State.Revision}

	encoded, err := json.Marshal(receipt)
	if err != nil {
		return workflowFileChange{}, err
	}

	receipt.Checksum = contenthash.Sum(encoded)
	encoded, err = json.Marshal(receipt)

	return workflowFileChange{Path: authoringReceiptRef(result.OperationID), After: encoded}, err
}

func (r *fileAuthoringSessionRepository) quizReceipt(operationID, digest string) (QuizActionResult, bool, error) {
	var receipt quizActionReceipt

	encoded, err := r.authoringRead(authoringReceiptRef(operationID))
	if errors.Is(err, os.ErrNotExist) {
		return receipt.Result, false, nil
	}

	if err != nil {
		return receipt.Result, false, err
	}

	if err := decodeAuthoringState(encoded, &receipt); err != nil {
		return receipt.Result, false, err
	}

	checksum := receipt.Checksum
	receipt.Checksum = ""

	encoded, err = json.Marshal(receipt)
	if err != nil {
		return receipt.Result, false, err
	}

	if receipt.Schema != "chr-quiz-action/v1" || receipt.Digest != digest || receipt.Result.OperationID != operationID || contenthash.Sum(encoded) != checksum {
		return receipt.Result, false, workflowstate.ErrConflict
	}

	receipt.Result.State.Revision = receipt.Revision

	return receipt.Result, true, nil
}
