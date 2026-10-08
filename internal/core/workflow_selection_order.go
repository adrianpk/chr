package core

import (
	"fmt"
	"slices"
	"strings"
)

func (selection workflowSelection) orderMembers(members []WorkflowPreviewMember, selector WorkflowSelector) (string, error) {
	if selector.Extent == "loose" || selector.Extent == "ticket" {
		return "not_applicable", nil
	}

	for _, member := range members {
		descriptor := selection.slices[member.SliceID]
		if descriptor.TaskInputHash != WorkflowTaskInputHash(selection.snapshot.Records, member.SliceID) {
			return "", fmt.Errorf("%w: slice %s membership fingerprint changed", ErrWorkflowConflict, member.SliceID)
		}
	}

	rule := selection.setOrderRule(selector)

	slices.SortFunc(members, func(left, right WorkflowPreviewMember) int {
		if left.SetID != right.SetID {
			return selection.compareSets(left.SetID, right.SetID, rule)
		}

		if left.SliceID != right.SliceID {
			return compareRank(selection.slices[left.SliceID].Order, selection.slices[right.SliceID].Order)
		}

		return compareRank(left.TaskOrder, right.TaskOrder)
	})

	return rule, nil
}

func (selection workflowSelection) compareSets(leftID, rightID, rule string) int {
	left, right := selection.sets[leftID], selection.sets[rightID]
	if rule == "use_case_order" {
		return compareRank(left.Set.UseCaseOrder, right.Set.UseCaseOrder)
	}

	if left.CreatedAt.Before(right.CreatedAt) {
		return -1
	}

	if right.CreatedAt.Before(left.CreatedAt) {
		return 1
	}

	return strings.Compare(leftID, rightID)
}

func (selection workflowSelection) setOrderRule(selector WorkflowSelector) string {
	if selector.Extent != "use_case" {
		return "slice_order"
	}

	for _, set := range selection.sets {
		if set.Set.UseCase == selector.Identity && set.Set.UseCaseOrder > 0 {
			return "use_case_order"
		}
	}

	return "created_at_id"
}
