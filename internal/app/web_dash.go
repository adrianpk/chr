package app

import (
	"context"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/webui"
)

func (p webViewProvider) DashView(ctx context.Context) (webui.DashView, error) {
	return p.DashPageView(ctx, webui.DashRequest{})
}

func (p webViewProvider) DashPageView(ctx context.Context, request webui.DashRequest) (webui.DashView, error) {
	observation, err := p.service(ctx).ObserveWork(ctx, request.QueueOffset)
	if err != nil {
		return webui.DashView{}, err
	}

	return webDashObservation(observation, request.QueueOffset), nil
}

func webDashObservation(observed WorkObservation, offset int) webui.DashView {
	view := webui.DashView{Authority: string(observed.Authority), Warnings: observed.Warnings}

	view.Readiness = webui.WorkReadinessView{Status: string(observed.Readiness.Status), WorkspaceMode: string(observed.Readiness.Workspace.Mode), CurrentBranch: observed.Readiness.CurrentBranch, IntegrationBranch: observed.Readiness.Policy.IntegrationBranch}
	for _, diagnostic := range observed.Readiness.Diagnostics {
		view.Readiness.Diagnostics = append(view.Readiness.Diagnostics, webui.WorkReadinessDiagnosticView{Code: string(diagnostic.Code), Severity: string(diagnostic.Severity), Detail: diagnostic.Detail})
	}

	webDashQueue(&view, observed, offset)

	for _, current := range observed.Current {
		view.Current = append(view.Current, webWorkContext(current))
	}

	for _, result := range observed.Results {
		view.Results = append(view.Results, webWorkContext(result))
	}

	for _, event := range observed.Events {
		view.RecentEvents = append(view.RecentEvents, webui.AgentEventView{ID: event.ID, ContextID: event.ContextID, Time: event.Time.UTC().Format(time.RFC3339), TicketKey: event.TicketKey, Phase: workDisplayLabel(event.Phase), Message: event.Message})
	}

	view.Revision = webStableRevision(view)
	view.ObservedAt = observed.ObservedAt.UTC().Format(time.RFC3339)

	return view
}

func webWorkContext(current WorkContext) webui.CurrentWorkView {
	view := webui.CurrentWorkView{ID: current.ID, Kind: current.Kind, Title: firstNonEmptyString(current.Title, current.TicketKey, workDisplayLabel(current.Kind)+" work"), TicketKey: current.TicketKey, Scope: current.Scope, Status: workDisplayLabel(current.Status), Phase: workDisplayLabel(current.Phase), Message: current.Message, UpdatedAt: current.UpdatedAt.UTC().Format(time.RFC3339), Validated: current.Validated, Confirmed: current.Confirmed, Remaining: current.Remaining, Eligible: current.Eligible, Processed: current.Processed, Blocked: current.Blocked, Failed: current.Failed}
	if !current.ActivityAt.IsZero() {
		view.ActivityAt = current.ActivityAt.UTC().Format(time.RFC3339)
	}

	if current.Kind == "native" {
		view.URL = "/workflows?id=" + url.QueryEscape(current.InvocationID)
	} else if current.TicketKey != "" {
		view.URL = "/board/ticket?key=" + url.QueryEscape(current.TicketKey)
	}

	if current.Kind == "execution" && current.Status == "succeeded" {
		view.Status = "Harness succeeded"
	}

	return view
}

func workDisplayLabel(value string) string { return strings.ReplaceAll(value, "_", " ") }

func webDashQueue(view *webui.DashView, observed WorkObservation, offset int) {
	view.Queue = webui.AgentQueueView{EligibleCount: observed.Eligible, TotalCount: observed.QueueTotal}
	for _, item := range observed.Queue {
		view.Queue.Items = append(view.Queue.Items, webui.WorkQueueItemView{Key: item.Key, Title: item.Title, Scope: item.Scope, Readiness: item.Readiness, Reason: item.Reason, URL: "/board/ticket?key=" + url.QueryEscape(item.Key)})
	}

	if observed.QueueNext >= 0 {
		view.Queue.NextURL = "/dash?queue_offset=" + strconv.Itoa(observed.QueueNext)
	}

	if offset > 0 {
		view.Queue.PreviousURL = "/dash?queue_offset=" + strconv.Itoa(max(0, offset-128))
	}

	for _, reason := range slices.Sorted(maps.Keys(observed.Excluded)) {
		if count := observed.Excluded[reason]; count > 0 {
			view.Exclusions = append(view.Exclusions, fmt.Sprintf("%d excluded: %s", count, workDisplayLabel(reason)))
		}
	}
}
