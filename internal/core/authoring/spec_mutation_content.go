package authoring

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	authoringstate "forge.adrianpk.com/chr/internal/core/authoring/state"
	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func specDocumentChanges(document SpecDocument, request SpecMutationRequest) ([]workflowstate.FileChange, error) {
	status := request.Status
	if status == "" {
		status = document.Status
	}

	if status == "" {
		status = "draft"
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)

	if request.Action == "save" {
		if err := AdvanceArtifactMetadata(document.metadata); err != nil {
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

	changes := append([]workflowstate.FileChange(nil), document.images...)
	changes[0].After, changes[1].After = content, append(metadata, '\n')

	return changes, nil
}

func specReceiptChange(digest string, result SpecMutationResult) (workflowstate.FileChange, error) {
	receipt := specMutationReceipt{Schema: "chr-spec-receipt/v1", Digest: digest, Result: result}

	encoded, err := json.Marshal(receipt)
	if err != nil {
		return workflowstate.FileChange{}, err
	}

	receipt.Checksum = contenthash.Sum(encoded)
	encoded, err = json.Marshal(receipt)

	return workflowstate.FileChange{Path: authoringReceiptRef(result.OperationID), After: encoded}, err
}

type specReceiptLookup struct {
	Result SpecMutationResult
	Found  bool
}

func (r *fileSessionRepository) specReceipt(operationID, digest string) (specReceiptLookup, error) {
	var receipt specMutationReceipt

	encoded, err := r.readAuthoringDocument(authoringReceiptRef(operationID))
	if errors.Is(err, os.ErrNotExist) {
		return specReceiptLookup{Result: receipt.Result}, nil
	}

	if err != nil {
		return specReceiptLookup{Result: receipt.Result}, err
	}

	if err := authoringstate.Decode(encoded, &receipt); err != nil {
		return specReceiptLookup{Result: receipt.Result}, err
	}

	checksum := receipt.Checksum
	receipt.Checksum = ""

	encoded, err = json.Marshal(receipt)
	if err != nil {
		return specReceiptLookup{Result: receipt.Result}, err
	}

	if receipt.Schema != "chr-spec-receipt/v1" || receipt.Digest != digest || receipt.Result.OperationID != operationID || contenthash.Sum(encoded) != checksum {
		return specReceiptLookup{Result: receipt.Result}, workflowstate.ErrConflict
	}

	return specReceiptLookup{Result: receipt.Result, Found: true}, nil
}

func specDocumentFront(document SpecDocument, request SpecMutationRequest, status, now string) ([]string, error) {
	var front []string
	if document.front != "" {
		front = strings.Split(document.front, "\n")
	}

	for _, field := range []struct{ key, value string }{
		{"title", opsmetadata.Quote(request.Title)},
		{"project_id", opsmetadata.Quote(request.ProjectID)},
		{"case_slug", opsmetadata.Quote(request.CaseSlug)},
		{"status", opsmetadata.Quote(status)},
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
			front = replaceProjectMetadataField(front, field.key, opsmetadata.Quote(field.value))
		}
	}

	fields, err := ProjectDocumentFields(document.front)
	if err != nil {
		return nil, err
	}

	if fields["created_at"] == "" {
		created, err := specMetadataString(document.metadata, "created_at")
		if err != nil {
			return nil, err
		}

		front = replaceProjectMetadataField(front, "created_at", opsmetadata.Quote(created))
	}

	return front, nil
}
