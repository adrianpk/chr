package core

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"
)

func specDocumentChanges(document SpecDocument, request SpecMutationRequest) ([]workflowFileChange, error) {
	status := request.Status
	if status == "" {
		status = document.Status
	}

	if status == "" {
		status = "draft"
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)

	if request.Action == "save" {
		if err := advanceAuthoringMetadata(document.metadata); err != nil {
			return nil, err
		}
	}

	front, err := specDocumentFront(document, request, status, now)
	if err != nil {
		return nil, err
	}

	body := request.Body
	if document.heading {
		body = "\n# " + request.Title + "\n\n" + body
	}

	content := []byte("---\n" + strings.Join(front, "\n") + "\n---\n" + body + document.provenance)
	document.metadata["title"], document.metadata["status"] = request.Title, status
	document.metadata["project_id"], document.metadata["case_slug"] = request.ProjectID, request.CaseSlug

	document.metadata["updated_at"] = now
	if request.GuidedApproval != nil {
		document.metadata["guided_approval"] = request.GuidedApproval
	}

	metadata, err := json.MarshalIndent(document.metadata, "", "  ")
	if err != nil {
		return nil, err
	}

	changes := append([]workflowFileChange(nil), document.images...)
	changes[0].After, changes[1].After = content, append(metadata, '\n')

	return changes, nil
}

func specReceiptChange(digest string, result SpecMutationResult) (workflowFileChange, error) {
	receipt := specMutationReceipt{Schema: "chr-spec-receipt/v1", Digest: digest, Result: result}

	encoded, err := json.Marshal(receipt)
	if err != nil {
		return workflowFileChange{}, err
	}

	receipt.Checksum = workflowHash(encoded)
	encoded, err = json.Marshal(receipt)

	return workflowFileChange{Path: authoringReceiptRef(result.OperationID), After: encoded}, err
}

func (r *fileAuthoringSessionRepository) specReceipt(operationID, digest string) (SpecMutationResult, bool, error) {
	var receipt specMutationReceipt

	encoded, err := r.readAuthoringDocument(authoringReceiptRef(operationID))
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

	if receipt.Schema != "chr-spec-receipt/v1" || receipt.Digest != digest || receipt.Result.OperationID != operationID || workflowHash(encoded) != checksum {
		return receipt.Result, false, ErrWorkflowConflict
	}

	return receipt.Result, true, nil
}

func specDocumentFront(document SpecDocument, request SpecMutationRequest, status, now string) ([]string, error) {
	var front []string
	if document.front != "" {
		front = strings.Split(document.front, "\n")
	}

	for _, field := range []struct{ key, value string }{
		{"title", yamlString(request.Title)},
		{"project_id", yamlString(request.ProjectID)},
		{"case_slug", yamlString(request.CaseSlug)},
		{"status", yamlString(status)},
		{"updated_at", now},
	} {
		front = replaceProjectMetadataField(front, field.key, field.value)
	}

	if approval := request.GuidedApproval; approval != nil {
		for _, field := range []struct{ key, value string }{
			{"approved_by", approval.Actor},
			{"guided_run", approval.RunID},
			{"guided_approval_policy", approval.Policy},
			{"guided_generated_revision", approval.GeneratedRevision},
			{"guided_content_hash", approval.ContentHash},
			{"guided_input_digest", approval.InputDigest},
		} {
			front = replaceProjectMetadataField(front, field.key, yamlString(field.value))
		}
	}

	fields, err := projectDocumentFields(document.front)
	if err != nil {
		return nil, err
	}

	if fields["created_at"] == "" {
		created, err := specMetadataString(document.metadata, "created_at")
		if err != nil {
			return nil, err
		}

		front = replaceProjectMetadataField(front, "created_at", yamlString(created))
	}

	return front, nil
}
