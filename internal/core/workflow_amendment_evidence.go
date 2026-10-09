package core

import (
	"encoding/json"
	"fmt"
	"path/filepath"
)

func (s *WorkflowStore) checkWorkflowAmendmentEvidence(set WorkflowRecord, binding WorkflowAmendmentRef) error {
	path, err := s.safePath(binding.Ref)
	if err != nil {
		return err
	}

	data, err := readWorkflowFile(path)
	if err != nil {
		return err
	}

	if workflowHash(data) != binding.Hash {
		return ErrWorkflowConflict
	}

	var receipt workflowSetAmendmentReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return err
	}

	root := filepath.ToSlash(filepath.Join("plan", set.ID, "extension", "chr", fmt.Sprintf("amendment-%04d", binding.Revision)))
	if !workflowAmendmentReceiptMatches(set, binding, receipt, root) {
		return ErrWorkflowConflict
	}

	hashes := map[string]string{root + "/plan.md": receipt.Plan.PlanHash}
	if receipt.ReopenedClosure != "" {
		hashes[receipt.ReopenedClosure] = receipt.ClosureHash
	}

	return s.checkWorkflowPrerequisiteEvidence(hashes)
}

func workflowAmendmentReceiptMatches(set WorkflowRecord, binding WorkflowAmendmentRef, receipt workflowSetAmendmentReceipt, root string) bool {
	return receipt.Schema == "chr-set-amendment/v1" && receipt.SetID == set.ID && receipt.PreviousRevision+1 == binding.Revision && binding.Revision <= set.Revision && binding.Ref == root+"/amendment.json"
}
