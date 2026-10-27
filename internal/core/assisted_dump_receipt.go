package core

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type assistedDumpMutationReceipt struct {
	Schema, Digest, Checksum string
	Session                  AuthoringSession
	Reference                AuthoringArtifactRef
	Mutation                 ProjectDumpMutationResult
}

func assistedDumpReceiptChange(digest string, result AssistedArtifactSynthesisResult) (workflowFileChange, error) {
	receipt := assistedDumpMutationReceipt{Schema: "chr-assisted-dump-receipt/v1", Digest: digest, Session: result.Session, Reference: result.Layout.Ref, Mutation: *result.Mutation}

	encoded, err := json.Marshal(receipt)
	if err != nil {
		return workflowFileChange{}, err
	}

	receipt.Checksum = contenthash.Sum(encoded)
	encoded, err = json.Marshal(receipt)

	return workflowFileChange{Path: authoringReceiptRef(receipt.Mutation.OperationID), After: encoded}, err
}

func (r *fileAuthoringSessionRepository) assistedDumpReceipt(operationID, digest string) (AssistedArtifactSynthesisResult, bool, error) {
	var receipt assistedDumpMutationReceipt

	encoded, err := r.readAuthoringDocument(filepath.ToSlash(authoringReceiptRef(operationID)))
	if errors.Is(err, os.ErrNotExist) {
		return AssistedArtifactSynthesisResult{}, false, nil
	}

	if err != nil {
		return AssistedArtifactSynthesisResult{}, false, err
	}

	if err := authoringstate.Decode(encoded, &receipt); err != nil {
		return AssistedArtifactSynthesisResult{}, false, err
	}

	if err := validateAssistedDumpReceipt(receipt, operationID, digest); err != nil {
		return AssistedArtifactSynthesisResult{}, false, err
	}

	content, err := r.assistedDumpReceiptContent(receipt.Mutation)
	if err != nil {
		return AssistedArtifactSynthesisResult{}, false, err
	}

	fullPath := filepath.Join(r.projectDir, filepath.FromSlash(receipt.Reference.Path))
	result := AssistedArtifactSynthesisResult{Session: receipt.Session, Layout: AuthoringArtifactLayout{Ref: receipt.Reference, FullPath: fullPath, Dir: filepath.Dir(fullPath)}, Content: string(content), Mutation: &receipt.Mutation}

	return result, true, nil
}

func validateAssistedDumpReceipt(receipt assistedDumpMutationReceipt, operationID, digest string) error {
	checksum := receipt.Checksum
	receipt.Checksum = ""

	encoded, err := json.Marshal(receipt)
	if err != nil {
		return err
	}

	if receipt.Schema != "chr-assisted-dump-receipt/v1" || receipt.Digest != digest || receipt.Mutation.OperationID != operationID || contenthash.Sum(encoded) != checksum {
		return workflowstate.ErrConflict
	}

	if receipt.Reference.Path == "" || receipt.Reference.Path != receipt.Mutation.Destination {
		return workflowstate.ErrConflict
	}

	return ValidateProjectDumpReference(receipt.Reference.Path)
}

func (r *fileAuthoringSessionRepository) assistedDumpReceiptContent(mutation ProjectDumpMutationResult) ([]byte, error) {
	artifact := ProjectDumpArtifact{Reference: mutation.Destination, SourceHash: mutation.SourceHash, MetadataHash: mutation.MetadataHash}

	body, err := r.readAuthoringDocument(projectDumpHistoryRef(artifact) + "/artifact.md")
	if err != nil {
		return nil, err
	}

	metadata, err := r.readAuthoringDocument(projectDumpHistoryRef(artifact) + "/metadata.json")
	if err != nil {
		return nil, err
	}

	if contenthash.Sum(body) != mutation.SourceHash || contenthash.Sum(metadata) != mutation.MetadataHash {
		return nil, workflowstate.ErrConflict
	}

	return body, nil
}
