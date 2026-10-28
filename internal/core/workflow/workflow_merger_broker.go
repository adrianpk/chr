package workflow

import (
	"unicode/utf8"

	workflowmerge "forge.adrianpk.com/chr/internal/core/workflow/merge"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (s *Store) workflowMergerCall(input workflowstate.MergerInput) (workflowmerge.Call, error) {
	call := workflowmerge.Call{Protocol: workflowmerge.BrokerProtocol, Input: input, Evidence: map[string]string{}}
	total := 0

	for _, ref := range input.EvidenceRefs() {
		if _, present := call.Evidence[ref.Path]; present {
			continue
		}

		data, err := s.readWorkflowReviewImage(ref.Path, ref.Hash)
		if err != nil {
			return call, err
		}

		total += len(data)
		if total >= workflowstate.FileLimit || !utf8.Valid(data) {
			return call, workflowstate.ErrConflict
		}

		call.Evidence[ref.Path] = string(data)
	}

	return call, nil
}
