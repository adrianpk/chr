package core

import (
	"bytes"
	"context"
	"slices"
	"strings"
)

func validateWorkflowDependencyTargets(run workflowInvocationFile) error {
	if len(run.DependencyTargets) > WorkflowPageLimit || len(run.DependencyTargets) > 0 && run.ExecutionTarget == nil {
		return ErrWorkflowConflict
	}

	previousCursor := 0
	for i, proof := range run.DependencyTargets {
		if !validWorkflowDependencyTarget(run, proof, i) || proof.Cursor < previousCursor || proof.Previous != workflowDependencyPreviousTarget(run, i) {
			return ErrWorkflowConflict
		}

		previousCursor = proof.Cursor
	}

	return nil
}

func validWorkflowDependencyTarget(run workflowInvocationFile, proof WorkflowDependencyTarget, i int) bool {
	if !validWorkflowDependencyTargetHeader(run, proof) {
		return false
	}

	if !slices.ContainsFunc(run.Members, func(member WorkflowMember) bool { return member.Admission.Key == proof.TicketKey }) {
		return false
	}

	return validWorkflowDependencyTargetScope(run, proof) && validWorkflowDependencyTargetEvidence(run, proof, i)
}

func validWorkflowDependencyTargetHeader(run workflowInvocationFile, proof WorkflowDependencyTarget) bool {
	if proof.Schema != "chr-dependency-target/v1" || proof.InvocationID != run.ID || !validWorkflowPRBranch(proof.Branch) || proof.Cursor < 0 || proof.Cursor >= len(run.Members) || proof.Cursor > run.Cursor || run.Members[proof.Cursor].Admission.SliceID != "" {
		return false
	}

	return true
}

func validWorkflowDependencyTargetScope(run workflowInvocationFile, proof WorkflowDependencyTarget) bool {
	return (run.Selection.Selector.Extent == "loose" || run.Selection.Selector.Extent == "ticket") && proof.Target.Repository == run.Selection.Target.Repository && proof.Previous.Repository == proof.Target.Repository && workflowCommitID(proof.Target.Commit) && workflowCommitID(proof.Previous.Commit) && workflowDependencyBranchMatches(run, proof.Branch)
}

func validWorkflowDependencyTargetEvidence(run workflowInvocationFile, proof WorkflowDependencyTarget, i int) bool {
	if proof.EvidenceRef != workflowDependencyTargetPath(run.ID, i+1) || len(proof.EvidenceHash) != 64 || proof.ConfirmedAt.IsZero() || proof.Blocker.NextActionOwner != "prerequisite_owner" || len(proof.Receipts) == 0 || len(proof.Receipts) > WorkflowPageLimit {
		return false
	}

	for path, hash := range proof.Receipts {
		if !validWorkflowDependencyReceipt(path, hash) {
			return false
		}
	}

	return true
}

func validWorkflowDependencyReceipt(path, hash string) bool {
	return strings.HasPrefix(path, "run/") && strings.Contains(path, "/evidence/") && strings.HasSuffix(path, ".json") && len(hash) == 64
}

func workflowDependencyPreviousTarget(run workflowInvocationFile, index int) WorkflowPreviewTarget {
	cursor := run.DependencyTargets[index].Cursor

	target := run.Selection.Target
	for _, delivery := range run.Deliveries {
		if delivery.Cursor <= cursor {
			target = delivery.NextTarget
		}
	}

	if index > 0 {
		previous := run.DependencyTargets[index-1]
		if previous.Cursor == cursor {
			target = previous.Target
		}
	}

	return target
}

func (s *WorkflowStore) checkWorkflowDependencyTargets(ctx context.Context, run workflowInvocationFile) error {
	for _, proof := range run.DependencyTargets {
		path, err := s.safePath(proof.EvidenceRef)
		if err != nil {
			return err
		}

		data, err := readWorkflowFile(path)
		if err != nil || workflowHash(data) != proof.EvidenceHash {
			return ErrWorkflowConflict
		}

		expected, err := encodeWorkflowDependencyTarget(proof)
		if err != nil || !bytes.Equal(data, expected) {
			return ErrWorkflowConflict
		}

		if err := s.checkWorkflowPrerequisiteEvidence(proof.Receipts); err != nil {
			return err
		}

		reader := workflowProofReader{store: s}
		if !reader.isAncestor(ctx, proof.Previous.Commit, proof.Target.Commit) {
			return ErrWorkflowConflict
		}
	}

	return nil
}
