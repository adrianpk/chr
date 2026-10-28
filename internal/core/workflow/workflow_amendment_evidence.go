package workflow

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) checkWorkflowAmendmentEvidence(set workflowstate.Record, binding workflowstate.WorkflowAmendmentRef) error {
	path, err := s.safePath(binding.Ref)
	if err != nil {
		return err
	}

	data, err := opsfile.Read(path)
	if err != nil {
		return err
	}

	if contenthash.Sum(data) != binding.Hash {
		return workflowstate.ErrConflict
	}

	var receipt workflowSetAmendmentReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return err
	}

	root := filepath.ToSlash(filepath.Join("plan", set.ID, "extension", "chr", fmt.Sprintf("amendment-%04d", binding.Revision)))
	if !workflowAmendmentReceiptMatches(set, binding, receipt, root) {
		return workflowstate.ErrConflict
	}

	hashes := map[string]string{root + "/plan.md": receipt.Plan.PlanHash}
	if receipt.ReopenedClosure != "" {
		hashes[receipt.ReopenedClosure] = receipt.ClosureHash
	}

	return s.checkWorkflowPrerequisiteEvidence(hashes)
}

func workflowAmendmentReceiptMatches(set workflowstate.Record, binding workflowstate.WorkflowAmendmentRef, receipt workflowSetAmendmentReceipt, root string) bool {
	return receipt.Schema == "chr-set-amendment/v1" && receipt.SetID == set.ID && receipt.PreviousRevision+1 == binding.Revision && binding.Revision <= set.Revision && binding.Ref == root+"/amendment.json"
}
