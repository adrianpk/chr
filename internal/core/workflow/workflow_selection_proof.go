package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowProofReader struct {
	store        *Store
	selection    workflowSelection
	request      workflowstate.PreviewRequest
	observations map[workflowProofKey]string
}

type workflowProofKey struct {
	ref     workflowstate.Ref
	sliceID string
}

func (reader *workflowProofReader) prerequisiteReason(ctx context.Context, task workflowstate.Record, ref workflowstate.Ref) string {
	key := workflowProofKey{ref: ref, sliceID: task.Task.SliceID}
	if reason, found := reader.observations[key]; found {
		return reason
	}

	if len(reader.observations) >= workflowstate.PageLimit {
		return "proof_observation_bound"
	}

	record := reader.selection.records[ref.ID]
	sameSlice := ref.Kind == "ticket" && task.Task.SliceID != "" && record.Task != nil && record.Task.SliceID == task.Task.SliceID

	var reason string
	if sameSlice {
		reason = reader.taskProofReason(ctx, record)
	} else {
		reason = reader.integrationProofReason(ctx, ref)
	}

	reader.observations[key] = reason

	return reason
}

func (reader *workflowProofReader) taskProofReason(ctx context.Context, record workflowstate.Record) string {
	if !slices.Contains([]string{"validated", "waiting_slice", "integrated"}, record.Task.Phase) {
		return "task_validation_required"
	}

	if reader.request.CumulativeBranch == "" || reader.request.CumulativeCommit == "" {
		return "cumulative_lane_required"
	}

	for _, relative := range record.Task.Evidence {
		data, err := reader.readEvidence(relative)
		if err != nil {
			continue
		}

		var proof workflowstate.TaskValidation
		if err := json.Unmarshal(data, &proof); err != nil {
			continue
		}

		if reader.validTaskProof(ctx, record, proof) {
			return ""
		}
	}

	return "task_evidence_invalid"
}

func (reader *workflowProofReader) validTaskProof(ctx context.Context, record workflowstate.Record, proof workflowstate.TaskValidation) bool {
	matches := proof.Schema == "chr-task-validation/v1" && proof.Successful && proof.Key == record.Task.Key && proof.ContractHash == record.Task.ContractHash && proof.Branch == reader.request.CumulativeBranch

	return matches && reader.commitHasTree(ctx, proof.Commit, proof.Tree) && reader.isAncestor(ctx, proof.Commit, reader.request.CumulativeCommit) && reader.branchAtCommit(ctx)
}

func (reader *workflowProofReader) readEvidence(relative string) ([]byte, error) {
	path, err := reader.store.safePath(relative)
	if err != nil {
		return nil, err
	}

	data, err := opsfile.Read(path)
	if err == nil && contenthash.Sum(data) != reader.selection.evidenceHashes[relative] {
		return nil, workflowstate.ErrConflict
	}

	return data, err
}

func (reader *workflowProofReader) integrationProofReason(ctx context.Context, ref workflowstate.Ref) string {
	delivery := reader.deliveryRef(ref)
	if delivery == "" {
		return "integration_evidence_required"
	}

	for _, record := range reader.selection.snapshot.Records {
		if record.Path != delivery || record.Receipt == nil {
			continue
		}

		receipt := record.Receipt

		expected := reader.integrationUnit(ref)
		if receipt.Unit != expected || receipt.Repository != reader.request.Target.Repository || reader.request.Target.Commit == "" {
			return "target_delivery_unproven"
		}

		if reader.validReceiptTarget(ctx, receipt) {
			return ""
		}
	}

	return "target_delivery_unproven"
}

func (reader *workflowProofReader) validReceiptTarget(ctx context.Context, receipt *workflowstate.Receipt) bool {
	return reader.commitHasTree(ctx, receipt.ResultCommit, receipt.VerifiedTree) && reader.isAncestor(ctx, receipt.ResultCommit, receipt.ObservedTarget) && reader.isAncestor(ctx, receipt.ObservedTarget, reader.request.Target.Commit)
}

func (reader *workflowProofReader) integrationUnit(ref workflowstate.Ref) workflowstate.Ref {
	task := reader.selection.records[ref.ID].Task
	if ref.Kind == "ticket" && task != nil && task.SliceID != "" {
		return workflowstate.Ref{Kind: "slice", ID: task.SliceID}
	}

	return ref
}

func (reader *workflowProofReader) deliveryRef(ref workflowstate.Ref) string {
	if ref.Kind == "slice" {
		return reader.selection.slices[ref.ID].DeliveryRef
	}

	task := reader.selection.records[ref.ID].Task
	if task == nil {
		return ""
	}

	if task.SliceID != "" {
		return reader.selection.slices[task.SliceID].DeliveryRef
	}

	return task.DeliveryRef
}

func (reader *workflowProofReader) gitOutput(ctx context.Context, args ...string) (string, error) {
	output, err := exec.CommandContext(ctx, "git", append([]string{"-C", reader.store.layout.ProjectDir()}, args...)...).Output()

	return strings.TrimSpace(string(output)), err
}

func (reader *workflowProofReader) commitHasTree(ctx context.Context, commit, tree string) bool {
	if !workflowstate.ValidCommitID(commit) || !workflowstate.ValidCommitID(tree) {
		return false
	}

	actual, err := reader.gitOutput(ctx, "rev-parse", "--verify", commit+"^{tree}")

	return err == nil && actual == tree
}

func (reader *workflowProofReader) isAncestor(ctx context.Context, commit, target string) bool {
	if !workflowstate.ValidCommitID(commit) || !workflowstate.ValidCommitID(target) {
		return false
	}

	_, err := reader.gitOutput(ctx, "merge-base", "--is-ancestor", commit, target)

	return err == nil
}

func (reader *workflowProofReader) branchAtCommit(ctx context.Context) bool {
	if !strings.HasPrefix(reader.request.CumulativeBranch, "refs/heads/") {
		return false
	}

	actual, err := reader.gitOutput(ctx, "rev-parse", "--verify", reader.request.CumulativeBranch+"^{commit}")

	return err == nil && actual == reader.request.CumulativeCommit
}

func (host workflowWorkObserver) prerequisites(ctx context.Context, selection workflowSelection, request workflowstate.PreviewRequest, preview *Preview) error {
	dependencies, err := newWorkflowDependencies(selection)
	if err != nil {
		return fmt.Errorf("workflow prerequisites: %w", err)
	}

	if err := dependencies.validateMembers(preview.Members); err != nil {
		return err
	}

	reader := workflowProofReader{store: host.store, selection: selection, request: request, observations: map[workflowProofKey]string{}}

	for i := range preview.Members {
		member := &preview.Members[i]
		record := selection.records[member.Key]

		member.Readiness = workflowTaskReadiness(record.Task)

		for _, ref := range dependencies.refs[workflowstate.Ref{Kind: "ticket", ID: member.Key}] {
			reason := reader.prerequisiteReason(ctx, record, ref)
			if reason == "" {
				continue
			}

			if preview.admitted[ref] {
				reason = "awaiting_selected_prerequisite: " + reason
			} else {
				reason = "outside_selection: " + reason
			}

			member.Readiness = "blocked"
			member.Blocker = &workflowstate.Blocker{Dependency: ref, Reason: reason, NextActionOwner: "prerequisite_owner"}

			break
		}
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	return nil
}

func workflowTaskReadiness(task *workflowstate.Task) string {
	if task.Phase == "pending" && task.Status == string(ticketcore.StatusReadyForDev) {
		return "ready"
	}

	return task.Phase
}
