package core

import (
	"fmt"
	"slices"
	"strings"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func (selection workflowSelection) orderMembers(members []WorkflowPreviewMember, selector workflowstate.Selector) (string, error) {
	if err := selection.validateMemberInputs(members); err != nil {
		return "", err
	}

	if selector.Extent == "ticket" {
		return "not_applicable", nil
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

func (selection workflowSelection) setOrderRule(selector workflowstate.Selector) string {
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

func (selection workflowSelection) validateMemberInputs(members []WorkflowPreviewMember) error {
	checked := map[string]bool{}
	for _, member := range members {
		if member.SliceID == "" || checked[member.SliceID] {
			continue
		}

		descriptor := selection.slices[member.SliceID]
		if descriptor.TaskInputHash != WorkflowTaskInputHash(selection.snapshot.Records, member.SliceID) {
			return fmt.Errorf("%w: slice %s membership fingerprint changed", workflowstate.ErrConflict, member.SliceID)
		}

		checked[member.SliceID] = true
	}

	return nil
}
