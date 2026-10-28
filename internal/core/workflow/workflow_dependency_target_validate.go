package workflow

import (
	"bytes"
	"context"
	"slices"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func validateWorkflowDependencyTargets(run workflowInvocationFile) error {
	if len(run.DependencyTargets) > workflowstate.PageLimit || len(run.DependencyTargets) > 0 && run.ExecutionTarget == nil {
		return workflowstate.ErrConflict
	}

	previousCursor := 0
	for i, proof := range run.DependencyTargets {
		if !validWorkflowDependencyTarget(run, proof, i) || proof.Cursor < previousCursor || proof.Previous != workflowDependencyPreviousTarget(run, i) {
			return workflowstate.ErrConflict
		}

		previousCursor = proof.Cursor
	}

	return nil
}

func validWorkflowDependencyTarget(run workflowInvocationFile, proof workflowstate.DependencyTarget, i int) bool {
	if !validWorkflowDependencyTargetHeader(run, proof) {
		return false
	}

	if !slices.ContainsFunc(run.Members, func(member workflowstate.Member) bool { return member.Admission.Key == proof.TicketKey }) {
		return false
	}

	return validWorkflowDependencyTargetScope(run, proof) && validWorkflowDependencyTargetEvidence(run, proof, i)
}

func validWorkflowDependencyTargetHeader(run workflowInvocationFile, proof workflowstate.DependencyTarget) bool {
	if proof.Schema != "chr-dependency-target/v1" || proof.InvocationID != run.ID || !workflowstate.ValidBranch(proof.Branch) || proof.Cursor < 0 || proof.Cursor >= len(run.Members) || proof.Cursor > run.Cursor || run.Members[proof.Cursor].Admission.SliceID != "" {
		return false
	}

	return true
}

func validWorkflowDependencyTargetScope(run workflowInvocationFile, proof workflowstate.DependencyTarget) bool {
	return (run.Selection.Selector.Extent == "loose" || run.Selection.Selector.Extent == "ticket") && proof.Target.Repository == run.Selection.Target.Repository && proof.Previous.Repository == proof.Target.Repository && workflowstate.ValidCommitID(proof.Target.Commit) && workflowstate.ValidCommitID(proof.Previous.Commit) && workflowDependencyBranchMatches(run, proof.Branch)
}

func validWorkflowDependencyTargetEvidence(run workflowInvocationFile, proof workflowstate.DependencyTarget, i int) bool {
	if proof.EvidenceRef != workflowDependencyTargetPath(run.ID, i+1) || len(proof.EvidenceHash) != 64 || proof.ConfirmedAt.IsZero() || proof.Blocker.NextActionOwner != "prerequisite_owner" || len(proof.Receipts) == 0 || len(proof.Receipts) > workflowstate.PageLimit {
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

func workflowDependencyPreviousTarget(run workflowInvocationFile, index int) workflowstate.PreviewTarget {
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

func (s *Store) checkWorkflowDependencyTargets(ctx context.Context, run workflowInvocationFile) error {
	for _, proof := range run.DependencyTargets {
		path, err := s.safePath(proof.EvidenceRef)
		if err != nil {
			return err
		}

		data, err := opsfile.Read(path)
		if err != nil || contenthash.Sum(data) != proof.EvidenceHash {
			return workflowstate.ErrConflict
		}

		expected, err := encodeWorkflowDependencyTarget(proof)
		if err != nil || !bytes.Equal(data, expected) {
			return workflowstate.ErrConflict
		}

		if err := s.checkWorkflowPrerequisiteEvidence(proof.Receipts); err != nil {
			return err
		}

		reader := workflowProofReader{store: s}
		if !reader.isAncestor(ctx, proof.Previous.Commit, proof.Target.Commit) {
			return workflowstate.ErrConflict
		}
	}

	return nil
}
