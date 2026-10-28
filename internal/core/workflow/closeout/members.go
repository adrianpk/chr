package closeout

import (
	"fmt"
	"slices"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Members requires integrated actual members and preserves deterministic key order.
func Members(records []workflowstate.Record, id string) ([]workflowstate.ClosingMember, error) {
	members := []workflowstate.ClosingMember{}

	for _, record := range records {
		if record.Task == nil || record.Task.SetID != id {
			continue
		}

		if record.Task.Phase != "integrated" || record.Task.Status != "done" || record.Task.AttemptID != "" {
			return nil, fmt.Errorf("%w: actual member %s requires integration", workflowstate.ErrConflict, record.Task.Key)
		}

		members = append(members, workflowstate.ClosingMember{Identity: record.Identity, Key: record.Task.Key, Path: record.Path, SourceHash: record.SourceHash, SliceID: record.Task.SliceID, Order: record.Task.Order})
	}

	slices.SortFunc(members, func(a, b workflowstate.ClosingMember) int { return strings.Compare(a.Key, b.Key) })

	return members, nil
}
