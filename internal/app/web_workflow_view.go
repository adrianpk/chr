package app

import (
	"fmt"
	"net/url"
	"slices"
	"strconv"

	"forge.adrianpk.com/chr/internal/core"
	"forge.adrianpk.com/chr/internal/webui"
)

func webWorkflowPreviewView(bound WorkflowBoundPreview) webui.WorkflowPreviewView {
	view := bound.View

	fields := []webui.WorkflowFieldView{
		{Label: "Extent", Value: view.Preview.Selector.Extent + " " + view.Preview.Selector.Identity},
		{Label: "Authority", Value: view.Preview.Authority},
		{Label: "Selected tickets", Value: strconv.Itoa(view.Preview.Total)},
		{Label: "Hierarchy order", Value: view.Preview.SetOrderRule},
		{Label: "Target", Value: view.Target.Repository + " " + view.Target.Commit},
		{Label: "Review / merge policy", Value: string(view.Policy.Mode) + " / " + view.Policy.Method + "; source " + view.Policy.Source},
		{Label: "Selection snapshot", Value: view.Preview.SnapshotHash},
		{Label: "Execution bound", Value: "At most one shared checkpoint. Continue from retained status."},
	}
	if bound.Selected != nil {
		fields = append(fields, webui.WorkflowFieldView{Label: "Selected ticket revision", Value: fmt.Sprintf("%s %d", bound.Selected.Key, bound.Selected.Revision)})
	}

	fields = append(fields, webWorkflowExclusions(view.Preview.Excluded)...)

	members := webWorkflowMembers(view.Preview.Members[:min(8, len(view.Preview.Members))])
	if view.Preview.Total > len(members) {
		fields = append(fields, webui.WorkflowFieldView{Label: "Member page", Value: fmt.Sprintf("First %d of %d. The snapshot binds the complete extent.", len(members), view.Preview.Total)})
	}

	return webui.WorkflowPreviewView{Fields: fields, Members: members}
}

func webWorkflowRun(status WorkflowWorkStatus) webui.WorkflowRunView {
	view := webui.WorkflowRunView{ID: status.ID, Revision: strconv.FormatUint(status.Revision, 10), Status: status.Status, ApprovalHash: status.ApprovalHash, CanApprove: status.ReviewState == "waiting_approval" && status.ApprovalHash != ""}

	view.Actions = []string{"continue", "pause", "cancel", "resume", "reconcile"}
	if view.CanApprove {
		view.Actions = append(view.Actions, "approve")
	}

	view.Fields = []webui.WorkflowFieldView{
		{Label: "Extent", Value: status.Selector.Extent + " " + status.Selector.Identity},
		{Label: "Authority", Value: status.Authority},
		{Label: "State / outcome / stop intent", Value: status.Status + " / " + status.Outcome + " / " + status.StopIntent},
		{Label: "Progress", Value: fmt.Sprintf("Host-validated tasks %d/%d; confirmed tickets %d; integrated units %d; remaining %d", status.ValidatedTasks, status.Total, status.ConfirmedTickets, status.IntegratedUnits, status.Remaining)},
		{Label: "Selected ticket / slice", Value: status.SelectedTicket + " / " + status.SelectedSlice},
		{Label: "Hierarchy order", Value: status.SetOrderRule},
		{Label: "Review / merge policy", Value: string(status.Policy.Mode) + " / " + status.Policy.Method + "; source " + status.Policy.Source},
		{Label: "Frozen limits", Value: fmt.Sprintf("Worker attempts %d; integration attempts %d; checkpoint steps %d", status.FrozenPolicy.MaxWorkerAttempts, status.FrozenPolicy.MaxIntegrationAttempts, workflowDisplayStepLimit(status.FrozenPolicy.CheckpointStepLimit))},
		{Label: "Target", Value: status.Target.Repository + " " + status.Target.Commit},
		{Label: "Next action owner", Value: status.NextActionOwner},
		{Label: "Capacity resume phase", Value: status.CapacityResumeStatus},
	}
	view.Fields = append(view.Fields, webWorkflowExclusions(status.Excluded)...)
	view.Fields = append(view.Fields, webWorkflowBindings(status)...)
	view.Fields = append(view.Fields, webWorkflowIntegrationFields(status.Integration)...)
	view.Integration = webWorkflowIntegrationControls(status)

	view.Members = webWorkflowMembers(status.Members)
	for _, evidence := range status.Evidence[max(0, len(status.Evidence)-8):] {
		view.Evidence = append(view.Evidence, webui.WorkflowFieldView{Label: evidence.ID, Value: fmt.Sprintf("%s %s %s; slot %d; %s %s", evidence.Purpose, evidence.TicketKey, evidence.Outcome, evidence.Slot, evidence.ResultRef, evidence.ResultHash)})
	}

	if status.NextOffset >= 0 {
		view.NextURL = fmt.Sprintf("/workflows?id=%s&offset=%d&revision=%d", url.QueryEscape(status.ID), status.NextOffset, status.Revision)
	}

	return view
}

func webWorkflowBindings(status WorkflowWorkStatus) []webui.WorkflowFieldView {
	fields := []webui.WorkflowFieldView{}
	if status.Blocker != nil {
		fields = append(fields, webui.WorkflowFieldView{Label: "Blocker", Value: fmt.Sprintf("%s %s; %s; owner %s", status.Blocker.Dependency.Kind, status.Blocker.Dependency.ID, status.Blocker.Reason, status.Blocker.NextActionOwner)})
	}

	if status.CandidateHash != "" {
		fields = append(fields, webui.WorkflowFieldView{Label: "Candidate", Value: status.CandidateRef + " " + status.CandidateHash})
	}

	if status.PullRequest != nil {
		fields = append(fields, webui.WorkflowFieldView{Label: "PR", Value: status.PullRequest.URL + " " + status.PullRequest.State + " " + status.PullRequest.SourceCommit})
	}

	if status.ApprovalHash != "" {
		fields = append(fields, webui.WorkflowFieldView{Label: "Approval binding", Value: status.ApprovalHash + " " + status.ReviewState})
	}

	if status.DeliveryHash != "" {
		fields = append(fields, webui.WorkflowFieldView{Label: "Delivery receipt", Value: status.DeliveryRef + " " + status.DeliveryHash})
	}

	if status.Continuation != nil {
		fields = append(fields, webui.WorkflowFieldView{Label: "Continuation", Value: fmt.Sprintf("%d %s; evidence %s %s", status.Continuation.Sequence, status.Continuation.Reason, status.Continuation.EvidenceRef, status.Continuation.EvidenceHash)})
	}

	return fields
}

func webWorkflowMembers(members []core.WorkflowPreviewMember) []webui.WorkflowFieldView {
	fields := make([]webui.WorkflowFieldView, 0, len(members))
	for _, member := range members {
		value := fmt.Sprintf("revision %d; set %s; slice %s; order %d; phase %s; readiness %s", member.Revision, member.SetID, member.SliceID, member.TaskOrder, member.Phase, member.Readiness)
		if member.Blocker != nil {
			value += "; blocker " + member.Blocker.Reason + "; owner " + member.Blocker.NextActionOwner
		}

		fields = append(fields, webui.WorkflowFieldView{Label: member.Key, Value: value})
	}

	return fields
}

func webWorkflowExclusions(excluded map[string]int) []webui.WorkflowFieldView {
	keys := make([]string, 0, len(excluded))
	for key := range excluded {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	fields := make([]webui.WorkflowFieldView, 0, len(keys))
	for _, key := range keys {
		fields = append(fields, webui.WorkflowFieldView{Label: "Excluded " + key, Value: strconv.Itoa(excluded[key])})
	}

	return fields
}
