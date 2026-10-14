package app

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chr/internal/core"
)

func (adapter workWorkflowCLI) preview(cmd *cli.Command, preview WorkflowWorkPreview) error {
	if cmd.Bool("json") {
		return json.NewEncoder(adapter.stdout).Encode(preview)
	}

	selection := preview.Preview
	lines := []string{
		"preview: true",
		fmt.Sprintf("scope: %s %s", selection.Selector.Extent, selection.Selector.Identity),
		"authority: " + selection.Authority,
		fmt.Sprintf("selected: %d", selection.Total),
		"snapshot: " + selection.SnapshotHash,
		fmt.Sprintf("next_offset: %d", selection.NextOffset),
		"set_order: " + selection.SetOrderRule,
	}
	lines = append(lines, workflowPublicationLines(preview.Target, preview.Policy)...)
	lines = append(lines, workflowExcludedLines(selection.Excluded)...)
	lines = append(lines, workflowMemberLines(selection.Members)...)

	return printAgentWorkLines(adapter.stdout, lines)
}

func (adapter workWorkflowCLI) status(ctx context.Context, cmd *cli.Command, operations service, id, action string) error {
	request := WorkflowStatusRequest{Limit: core.WorkflowPageLimit}
	if action == "status" {
		request.Offset, request.Limit = cmd.Int("offset"), cmd.Int("limit")
		request.ExpectedRevision = uint64(cmd.Int("revision"))
	}

	status, err := operations.WorkflowStatus(ctx, id, request)
	if err != nil {
		return err
	}

	if cmd.Bool("json") {
		return json.NewEncoder(adapter.stdout).Encode(struct {
			Action  string `json:"action"`
			Preview bool   `json:"preview"`
			WorkflowWorkStatus
		}{Action: action, Preview: cmd.Bool("preview"), WorkflowWorkStatus: status})
	}

	lines := []string{fmt.Sprintf("action: %s", action), fmt.Sprintf("preview: %t", cmd.Bool("preview"))}

	return printAgentWorkLines(adapter.stdout, append(lines, workflowStatusLines(status)...))
}

func workflowStatusLines(status WorkflowWorkStatus) []string {
	lines := []string{
		"execution: " + status.ID,
		fmt.Sprintf("revision: %d", status.Revision),
		fmt.Sprintf("scope: %s %s", status.Selector.Extent, status.Selector.Identity),
		"authority: " + status.Authority,
		"status: " + status.Status,
		"outcome: " + status.Outcome,
		fmt.Sprintf("task_validation: %d/%d", status.ValidatedTasks, status.Total),
		fmt.Sprintf("confirmed_tickets: %d; integrated_units: %d; remaining: %d", status.ConfirmedTickets, status.IntegratedUnits, status.Remaining),
		fmt.Sprintf("next_offset: %d", status.NextOffset),
		"stop_intent: " + status.StopIntent,
	}
	lines = append(lines, workflowPublicationLines(status.Target, status.Policy)...)
	lines = append(lines, workflowDeliveryLines(status)...)
	lines = append(lines, workflowIntegrationLines(status.Integration)...)
	lines = append(lines, workflowBlockerLines(status.Blocker)...)
	lines = append(lines, workflowExcludedLines(status.Excluded)...)

	lines = append(lines, workflowMemberLines(status.Members)...)
	for _, proof := range status.Evidence {
		lines = append(lines, fmt.Sprintf("evidence: %s; purpose: %s; ticket: %s; outcome: %s; ref: %s; hash: %s", proof.ID, proof.Purpose, proof.TicketKey, proof.Outcome, proof.ResultRef, proof.ResultHash))
	}

	return lines
}

func workflowPublicationLines(target core.WorkflowPreviewTarget, policy core.WorkflowReviewPolicy) []string {
	return []string{
		fmt.Sprintf("target: %s @ %s", target.Repository, target.Commit),
		fmt.Sprintf("publication_policy: %s; method: %s; source: %s; revision: %s", policy.Mode, policy.Method, policy.Source, policy.Revision),
	}
}

func workflowDeliveryLines(status WorkflowWorkStatus) []string {
	lines := []string{
		fmt.Sprintf("candidate: %s @ %s", status.CandidateRef, status.CandidateHash),
		"approval_binding: " + status.ApprovalHash,
		"review: " + status.ReviewState,
		fmt.Sprintf("delivery: %s @ %s", status.DeliveryRef, status.DeliveryHash),
	}
	if status.PullRequest != nil {
		binding := status.PullRequest
		lines = append(lines, fmt.Sprintf("pull_request: %s; id: %s; candidate: %s; target: %s", binding.URL, binding.ID, binding.CandidateHash, binding.TargetCommit))
	}

	return lines
}

func workflowBlockerLines(blocker *core.WorkflowBlocker) []string {
	if blocker == nil {
		return nil
	}

	return []string{fmt.Sprintf("blocker: %s; dependency: %s/%s; next_action_owner: %s", blocker.Reason, blocker.Dependency.Kind, blocker.Dependency.ID, blocker.NextActionOwner)}
}

func workflowExcludedLines(excluded map[string]int) []string {
	keys := make([]string, 0, len(excluded))
	for key := range excluded {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("excluded: %s %d", key, excluded[key]))
	}

	return lines
}

func workflowMemberLines(members []core.WorkflowPreviewMember) []string {
	lines := make([]string, 0, len(members))
	for _, member := range members {
		lines = append(lines, fmt.Sprintf("member: %s; revision: %d; slice_set: %s; slice: %s; order: %d; status: %s; phase: %s; readiness: %s", member.Key, member.Revision, member.SetID, member.SliceID, member.TaskOrder, member.Status, member.Phase, member.Readiness))
		lines = append(lines, workflowBlockerLines(member.Blocker)...)
	}

	return lines
}
