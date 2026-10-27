package core

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// WorkflowDependencyTarget binds a resumed independent cursor to canonical
// descendant history and the immutable prerequisite receipts observed there.
type WorkflowDependencyTarget struct {
	Schema       string                `json:"schema"`
	InvocationID string                `json:"invocation_id"`
	Cursor       int                   `json:"cursor"`
	TicketKey    string                `json:"ticket_key"`
	Previous     WorkflowPreviewTarget `json:"previous_target"`
	Target       WorkflowPreviewTarget `json:"target"`
	Branch       string                `json:"branch"`
	Blocker      WorkflowBlocker       `json:"blocker"`
	Receipts     map[string]string     `json:"receipt_hashes"`
	ConfirmedAt  time.Time             `json:"confirmed_at"`
	EvidenceRef  string                `json:"evidence_ref,omitempty"`
	EvidenceHash string                `json:"evidence_hash,omitempty"`
}

func (s *WorkflowStore) resumeWorkflowDependencyTarget(ctx context.Context, run *workflowInvocationFile, ports WorkflowResumePorts) ([]workflowFileChange, error) {
	if !workflowDependencyTargetReady(*run, ports) {
		return nil, ErrWorkflowObservationRequired
	}

	target, err := s.observeWorkflowDependencyTarget(ctx, *run, ports)
	if err != nil {
		return nil, err
	}

	next := *run

	next.ExecutionTarget = &target
	if _, err := s.observeWorkflowClaim(ctx, next); err != nil {
		return nil, err
	}

	receipts, err := s.workflowDependencyTargetReceipts(ctx, next)
	if err != nil {
		return nil, err
	}

	proof := WorkflowDependencyTarget{Schema: "chr-dependency-target/v1", InvocationID: run.ID, Cursor: run.Cursor, TicketKey: run.Members[run.Cursor].Admission.Key, Previous: run.ExecutionSelection().Target, Target: target, Branch: ports.TargetBranch, Blocker: *run.Blocker, Receipts: receipts, ConfirmedAt: time.Now().UTC()}

	data, err := encodeWorkflowDependencyTarget(proof)
	if err != nil {
		return nil, err
	}

	proof.EvidenceRef = workflowDependencyTargetPath(run.ID, len(run.DependencyTargets)+1)
	proof.EvidenceHash = contenthash.Sum(data)
	next.DependencyTargets = append(slices.Clone(run.DependencyTargets), proof)
	next.Blocker = nil
	*run = next

	return []workflowFileChange{{Path: proof.EvidenceRef, After: data}}, nil
}

func workflowDependencyTargetReady(run workflowInvocationFile, ports WorkflowResumePorts) bool {
	if ports.Target == nil || ports.Provider == nil || !validWorkflowPRBranch(ports.TargetBranch) || run.Lane != nil || len(run.DependencyTargets) >= workflowstate.PageLimit {
		return false
	}

	return workflowIndependentCursor(run) && workflowDependencyBranchMatches(run, ports.TargetBranch)
}

func workflowIndependentCursor(run workflowInvocationFile) bool {
	return run.Cursor < len(run.Members) && run.Members[run.Cursor].Admission.SliceID == "" && (run.Selection.Selector.Extent == "loose" || run.Selection.Selector.Extent == "ticket")
}

func workflowDependencyBranchMatches(run workflowInvocationFile, branch string) bool {
	if len(run.Deliveries) > 0 && run.Deliveries[len(run.Deliveries)-1].PullRequest.Request.BaseBranch != branch {
		return false
	}

	return len(run.DependencyTargets) == 0 || run.DependencyTargets[0].Branch == branch
}

func (s *WorkflowStore) observeWorkflowDependencyTarget(ctx context.Context, run workflowInvocationFile, ports WorkflowResumePorts) (WorkflowPreviewTarget, error) {
	query := WorkflowTargetQuery{Repository: run.Selection.Target.Repository, Branch: ports.TargetBranch, Directory: s.layout.ProjectDir()}

	target, err := ports.Target.ObserveWorkflowTarget(ctx, query)
	if err != nil {
		return target, err
	}

	observed, err := ports.Provider.ReadWorkflowBranch(ctx, query.Repository, query.Branch)
	if err != nil || target.Repository != query.Repository || !workflowCommitID(target.Commit) || observed != target.Commit {
		return target, errorsOrConflict(err)
	}

	reader := workflowProofReader{store: s}
	if !reader.isAncestor(ctx, run.ExecutionSelection().Target.Commit, target.Commit) {
		return target, workflowstate.ErrConflict
	}

	return target, s.checkWorkflowTarget(ctx, target)
}

func (s *WorkflowStore) workflowDependencyTargetReceipts(ctx context.Context, run workflowInvocationFile) (map[string]string, error) {
	selection, err := s.readWorkflowSelection(ctx)
	if err != nil {
		return nil, err
	}

	dependencies, err := newWorkflowDependencies(selection)
	if err != nil {
		return nil, err
	}

	key := run.Members[run.Cursor].Admission.Key

	refs := dependencies.refs[WorkflowRef{Kind: "ticket", ID: key}]
	if len(refs) == 0 || len(refs) > workflowstate.PageLimit {
		return nil, workflowstate.ErrConflict
	}

	reader := workflowProofReader{store: s, selection: selection, request: run.ExecutionSelection()}
	images := map[string]string{}

	for _, record := range selection.snapshot.Records {
		if record.Receipt != nil {
			images[record.Path] = record.SourceHash
		}
	}

	receipts := map[string]string{}

	for _, ref := range refs {
		path := reader.deliveryRef(ref)
		if reader.integrationProofReason(ctx, ref) != "" || len(images[path]) != 64 {
			return nil, workflowstate.ErrConflict
		}

		receipts[path] = images[path]
	}

	return receipts, nil
}

func workflowDependencyTargetPath(id string, sequence int) string {
	return fmt.Sprintf("run/%s/extension/chr/dependency-target/%04d.json", id, sequence)
}

func encodeWorkflowDependencyTarget(proof WorkflowDependencyTarget) ([]byte, error) {
	proof.EvidenceRef, proof.EvidenceHash = "", ""

	data, err := json.MarshalIndent(proof, "", "  ")
	if err != nil || len(data) >= workflowstate.FileLimit {
		return nil, fmt.Errorf("workflow dependency target exceeds file bounds: %w", errorsOrConflict(err))
	}

	return append(data, '\n'), nil
}
