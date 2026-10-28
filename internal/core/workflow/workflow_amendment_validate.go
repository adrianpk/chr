package workflow

import (
	"fmt"
	"path/filepath"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validateWorkflowAmendmentRefs(refs []workflowstate.WorkflowAmendmentRef) error {
	if len(refs) > workflowstate.RecordLimit {
		return fmt.Errorf("workflow amendment history bound exceeded")
	}

	previous := uint64(1)

	seen := map[string]bool{}
	for _, binding := range refs {
		if binding.Revision <= previous || !filepath.IsLocal(binding.Ref) || filepath.Clean(binding.Ref) != binding.Ref || filepath.Base(binding.Ref) != "amendment.json" || len(binding.Hash) != 64 || seen[binding.Ref] {
			return fmt.Errorf("invalid workflow set amendment binding")
		}

		previous = binding.Revision
		seen[binding.Ref] = true
	}

	return nil
}

func validateWorkflowClosureRefs(current string, history []string) error {
	if len(history) > workflowstate.RecordLimit {
		return fmt.Errorf("workflow closure history bound exceeded")
	}

	seen := map[string]bool{}

	refs := append([]string(nil), history...)
	if current != "" {
		refs = append(refs, current)
	}

	for _, ref := range refs {
		if !filepath.IsLocal(ref) || filepath.Clean(ref) != ref || seen[ref] {
			return fmt.Errorf("invalid or duplicate workflow closure reference")
		}

		seen[ref] = true
	}

	return nil
}
