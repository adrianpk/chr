package app

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	"forge.adrianpk.com/chr/internal/webui"
)

func (j *webJourney) planAndBind(t *testing.T) {
	t.Helper()
	webJourneyPost(t, j.handler, webPlanForm(webPlanDetail(t, j.provider)))

	operations := j.provider.service(t.Context())
	scope := operations.quizScope(quizcore.Scope{CaseSlug: "named-work"})

	binding, err := core.ObservePlanRevision(t.Context(), operations.projectContext.ProjectDir, scope)
	if err != nil {
		t.Fatal(err)
	}

	j.plan = filepath.ToSlash(filepath.Join("ops/default", binding.Reference))
	plan := journeyApprovedPlan(t, operations, binding)

	before, err := j.provider.BoardView(t.Context(), webui.BoardRequest{})
	if err != nil || before.TotalCount != 0 {
		t.Fatal("plan generation allocated tickets", before, err)
	}

	serviceWorkflowDeliveryGit(t, operations.projectContext.ProjectDir, "init", "-q")
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "ticket:\n  key_prefix: PRJ\nwork:\n  state_authority: tickets\n  mode: no_git_enforcement\n")

	approved, err := operations.MaterializePlanTracking(t.Context(), PlanTrackingRequest{Scope: scope, Workflow: &plan})
	if err != nil || approved.Workflow == nil || len(approved.Workflow.Tasks) != 1 {
		t.Fatal(approved, err)
	}

	j.ticket = approved.Workflow.Tasks[0].Task.Key

	retry, err := operations.MaterializePlanTracking(t.Context(), PlanTrackingRequest{Scope: scope, Workflow: &plan})
	if err != nil || retry.Workflow.Tasks[0].Task.Key != j.ticket {
		t.Fatal("approval retry rebound a ticket", retry, err)
	}
}

func journeyApprovedPlan(t *testing.T, operations service, binding core.PlanRevisionBinding) core.GeneratedWorkflowPlan {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(operations.projectContext.ProjectDir, "ops/default", binding.OrderRef))
	if err != nil {
		t.Fatal(err)
	}

	var order core.PlanOrderDocument
	if err := json.Unmarshal(data, &order); err != nil {
		t.Fatal(err)
	}

	if len(order.Slices) != 1 || len(order.Slices[0].Tasks) != 1 || order.Slices[0].Tasks[0].TicketKey != "" {
		t.Fatal("unexpected or preallocated planned order", order)
	}

	slice, task := order.Slices[0], order.Slices[0].Tasks[0]
	gate := core.WorkflowValidation{Required: true, Commands: [][]string{{"go", "test", "./..."}}, TimeoutSeconds: 60}

	return core.GeneratedWorkflowPlan{MaterializationID: "journey-approval", Title: "Journey delivery", PlanRef: binding.BodyRef, PlanHash: binding.BodyHash, PlanRevision: &binding, UseCase: "named-work", ClosingValidation: gate, Slices: []core.GeneratedWorkflowSlice{{ID: slice.ID, Order: slice.Order, Title: slice.Title, Label: "Slice 1", Contract: "Deliver the selected greeting", Validation: gate, Tasks: []core.GeneratedWorkflowTask{{ID: task.ID, Order: task.Order, Title: task.Title, Label: "T1.1", Body: "Print hello once", Outcome: "Greeting printed", Scope: "Selected greeting", Acceptance: []string{"stdout equals hello"}, Validation: gate, Priority: "normal", Severity: "normal", Type: "task"}}}}}
}

func (j *webJourney) boardAndObserve(t *testing.T) {
	t.Helper()

	root := j.provider.projectContext.ProjectDir
	before := journeyMapFiles(t, root)

	after := j.editBoard(t)

	observed, err := j.provider.service(t.Context()).ObserveWork(t.Context(), 0)
	if err != nil || len(observed.Queue) != 1 || observed.Queue[0].Key != j.ticket || observed.Queue[0].Title != after.Title || len(observed.Current) != 0 {
		t.Fatal(observed, err)
	}

	if !reflect.DeepEqual(before, journeyMapFiles(t, root)) {
		t.Fatal("Board progress or Dash observation changed plan/order bytes")
	}

	if !strings.Contains(webPlanGet(j.handler, "/dash").Body.String(), after.Title) {
		t.Fatal("Dash did not observe the edited Board identity")
	}
}

func (j *webJourney) editBoard(t *testing.T) webui.TicketDetailView {
	t.Helper()

	detail, err := j.provider.TicketDetailView(t.Context(), webui.TicketDetailRequest{Key: j.ticket})
	if err != nil || !detail.NativeWorkflow {
		t.Fatal(detail, err)
	}

	webJourneyPost(t, j.handler, url.Values{"action": {"ticket.update"}, "id": {j.ticket}, "revision": {detail.Revision}, "confirm": {"true"}, "request_id": {detail.RequestID}, "owner": {"Journey maintainer"}, "priority": {"high"}, "status": {"ready_for_dev"}})

	after, err := j.provider.TicketDetailView(t.Context(), webui.TicketDetailRequest{Key: j.ticket})
	if err != nil || after.Fields["owner"] != "Journey maintainer" || after.Fields["priority"] != "high" || after.UseCase != "named-work" || after.Status != "ready_for_dev" {
		t.Fatal(after, err)
	}

	return after
}

func journeyMapFiles(t *testing.T, root string) map[string]string {
	t.Helper()

	images := map[string]string{}

	for _, family := range []string{"plan", "tracker"} {
		err := filepath.WalkDir(filepath.Join(root, "ops/default", family), func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return err
			}

			data, err := os.ReadFile(path)
			images[path] = string(data)

			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	return images
}
