package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type quizActionReceipt struct {
	Schema, Digest, Checksum string
	Result                   quizcore.ActionResult
	Revision                 string
}

func (r *fileAuthoringSessionRepository) quizActionDigest(request quizcore.ActionRequest) (string, error) {
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

func quizReceiptChange(digest string, result quizcore.ActionResult) (workflowFileChange, error) {
	receipt := quizActionReceipt{Schema: "chr-quiz-action/v1", Digest: digest, Result: result, Revision: result.State.Revision}

	encoded, err := json.Marshal(receipt)
	if err != nil {
		return workflowFileChange{}, err
	}

	receipt.Checksum = contenthash.Sum(encoded)
	encoded, err = json.Marshal(receipt)

	return workflowFileChange{Path: authoringReceiptRef(result.OperationID), After: encoded}, err
}

func (r *fileAuthoringSessionRepository) quizReceipt(operationID, digest string) (quizReceiptLookup, error) {
	var receipt quizActionReceipt

	encoded, err := r.authoringRead(authoringReceiptRef(operationID))
	if errors.Is(err, os.ErrNotExist) {
		return quizReceiptLookup{Result: receipt.Result}, nil
	}

	if err != nil {
		return quizReceiptLookup{Result: receipt.Result}, err
	}

	if err := authoringstate.Decode(encoded, &receipt); err != nil {
		return quizReceiptLookup{Result: receipt.Result}, err
	}

	checksum := receipt.Checksum
	receipt.Checksum = ""

	encoded, err = json.Marshal(receipt)
	if err != nil {
		return quizReceiptLookup{Result: receipt.Result}, err
	}

	if receipt.Schema != "chr-quiz-action/v1" || receipt.Digest != digest || receipt.Result.OperationID != operationID || contenthash.Sum(encoded) != checksum {
		return quizReceiptLookup{Result: receipt.Result}, workflowstate.ErrConflict
	}

	receipt.Result.State.Revision = receipt.Revision

	return quizReceiptLookup{Result: receipt.Result, Found: true}, nil
}

type quizReceiptLookup struct {
	Result quizcore.ActionResult
	Found  bool
}
