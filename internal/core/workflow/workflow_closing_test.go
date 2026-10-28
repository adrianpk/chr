package workflow

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	workflowcloseout "forge.adrianpk.com/chr/internal/core/workflow/closeout"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowClosingProvider struct {
	workflowFrontierProvider
	prs map[string]workflowstate.PullRequestSnapshot
}

func (p workflowClosingProvider) ReadWorkflowPullRequest(ctx context.Context, url string) (workflowstate.PullRequestSnapshot, error) {
	pr, found := p.prs[url]
	if !found {
		return workflowstate.PullRequestSnapshot{}, workflowstate.ErrConflict
	}

	target, err := p.ReadWorkflowBranch(ctx, pr.Repository, pr.BaseBranch)
	pr.BaseCommit = target

	return pr, err
}

func workflowClosingFixture(t *testing.T) (workflowIntegrationFixture, FrontierPorts) {
	t.Helper()

	return workflowClosingExtentFixture(t, "slice_set", workflowstate.ExecutionPolicy{})
}

func workflowClosingExtentFixture(t *testing.T, extent string, policy workflowstate.ExecutionPolicy) (workflowIntegrationFixture, FrontierPorts) {
	t.Helper()

	fixture := workflowDeliveryPolicyFixture(t, "rebase", extent, policy)
	prs := map[string]workflowstate.PullRequestSnapshot{fixture.provider.pr.URL: fixture.provider.pr}
	workflowFinalizeFrontier(t, fixture)
	workflowAdvanceFrontier(t, fixture)
	workflowDeliverNextFrontier(t, &fixture)

	prs[fixture.provider.pr.URL] = fixture.provider.pr
	workflowAdvanceFrontier(t, fixture)

	ports := workflowFrontierPorts(fixture)
	ports.Provider = workflowClosingProvider{workflowFrontierProvider: workflowFrontierProvider{fixture.provider, fixture.remote}, prs: prs}

	return fixture, ports
}

func workflowClosingContract(t *testing.T, fixture workflowIntegrationFixture, gate workflowstate.Validation) workflowstate.Record {
	t.Helper()

	set, err := fixture.store.Read(t.Context(), fixture.hierarchy.Set.Path)
	if err != nil {
		t.Fatal(err)
	}

	request := workflowApprovedAmendment(t, fixture.store, set)
	request.Plan.ClosingValidation = gate

	changed, err := fixture.store.AmendWorkflowSet(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	return changed.Set
}

// TestWorkflowClosingRetry proves a failed host gate retries closure alone on
// real canonical Git and two distinct merged PRs, preserving every task byte.
func TestWorkflowClosingRetry(t *testing.T) {
	fixture, ports := workflowClosingFixture(t)
	counter := filepath.Join(t.TempDir(), "attempts")
	gate := workflowstate.Validation{Required: true, TimeoutSeconds: 30, Commands: [][]string{{"sh", "-c", "n=0; [ ! -f \"$1\" ] || read -r n <\"$1\"; n=$((n+1)); printf '%s\\n' \"$n\" >\"$1\"; [ \"$n\" -gt 1 ]", "closing-gate", counter}}}
	set := workflowClosingContract(t, fixture, gate)
	before := workflowClosingTaskBytes(t, fixture)
	run := workflowTestRead(t, fixture.store, fixture.run.ID)

	failed, err := CloseInvocationSet(t.Context(), fixture.store, run.ID, "dev", ports)
	if err == nil || failed.Outcome != "closing_validation_failed" || failed.Set.Set.ClosureRef != "" || len(failed.Set.Set.Closing.Attempts) != 1 {
		t.Fatal("failed gate established closure or lost its budget", err)
	}

	store := NewWorkflowStore(fixture.store.layout, fixture.store.projectID)

	closed, err := CloseInvocationSet(t.Context(), store, run.ID, "dev", ports)
	if err != nil {
		t.Fatal(err)
	}

	checkWorkflowClosingResult(t, fixture, set, closed)
	checkWorkflowClosingTasks(t, fixture, before, run)

	attached, err := CloseSet(t.Context(), store, workflowcloseout.Request{SetPath: closed.Set.Path, ExpectedRevision: closed.Set.Revision, ExpectedSourceHash: closed.Set.SourceHash, Repository: closed.Receipt.Target.Repository, Branch: "dev"}, ports)
	if err != nil || attached.ReceiptRef != closed.ReceiptRef || attached.Set.Revision != closed.Set.Revision {
		t.Fatal("confirmed retry created another closure attempt", err)
	}
}

func checkWorkflowClosingResult(t *testing.T, fixture workflowIntegrationFixture, set workflowstate.Record, closed workflowcloseout.Result) {
	t.Helper()

	if !workflowClosingResultComplete(closed) {
		t.Fatalf("incomplete closure: %+v", closed)
	}

	if closed.Receipt.Plan.Hash != set.Set.PlanHash || closed.Receipt.ContractHash == "" || closed.Receipt.Target.Commit != fixture.target || closed.Receipt.Tree != workflowIntegrationCommand(t, fixture.actor, "rev-parse", "HEAD^{tree}") {
		t.Fatal("receipt lost exact plan, contract or canonical target/tree")
	}
}

func workflowClosingResultComplete(closed workflowcloseout.Result) bool {
	return closed.Outcome == "completed" && closed.Set.Set.Status == "completed" && closed.Set.Set.ClosureRef == closed.ReceiptRef && len(closed.Receipt.Members) == 6 && len(closed.Receipt.Slices) == 2 && len(closed.Set.Set.Closing.Attempts) == 2
}

func workflowClosingTaskBytes(t *testing.T, fixture workflowIntegrationFixture) map[string][]byte {
	t.Helper()

	images := map[string][]byte{}
	for _, task := range fixture.hierarchy.Tasks {
		images[task.Path] = workflowReadTestBytes(t, fixture.store, task.Path)
	}

	return images
}

func checkWorkflowClosingTasks(t *testing.T, fixture workflowIntegrationFixture, images map[string][]byte, before workflowstate.Invocation) {
	t.Helper()

	for path, image := range images {
		if !bytes.Equal(image, workflowReadTestBytes(t, fixture.store, path)) {
			t.Fatal("closure replayed or mutated an integrated task")
		}
	}

	after := workflowTestRead(t, fixture.store, before.ID)
	if after.Status != workflowstate.InvocationCompleted || len(after.Attempts) != len(before.Attempts) || after.RequestHash != before.RequestHash || len(after.Closures) != 1 || after.Cursor != 6 {
		t.Fatal("closure changed task attempts, admission or confirmed progress")
	}
}

// TestWorkflowClosingEmpty requires an explicit approved gate or waiver even
// without tickets, and observes the real canonical target before confirmation.
func TestWorkflowClosingEmpty(t *testing.T) {
	for _, required := range []bool{false, true} {
		t.Run(fmt.Sprint(required), func(t *testing.T) { testWorkflowClosingEmpty(t, required) })
	}
}

func testWorkflowClosingEmpty(t *testing.T, required bool) {
	t.Helper()

	fixture := workflowDeliveryFixture(t, "rebase")
	plan := workflowPlanFixture()
	plan.MaterializationID, plan.Title, plan.UseCase, plan.UseCaseOrder = "empty-set", "Empty set", "", 0
	plan.Slices = nil

	plan.ClosingValidation = workflowstate.Validation{WaiverReason: "Approved explicit empty-set closure"}
	if required {
		plan.ClosingValidation = workflowstate.Validation{Required: true, TimeoutSeconds: 30, Commands: [][]string{{"git", "rev-parse", "HEAD"}}}
	}

	set, err := fixture.store.Create(t.Context(), "Empty set", workflowstate.Record{Kind: "plan", Set: &workflowstate.WorkflowSet{Schema: workflowstate.WorkflowSchema, Status: "ready", PlanRef: plan.PlanRef, PlanHash: plan.PlanHash, ClosingValidation: plan.ClosingValidation}})
	if err != nil {
		t.Fatal(err)
	}

	request := workflowcloseout.Request{SetPath: set.Path, ExpectedRevision: set.Revision, ExpectedSourceHash: set.SourceHash, Repository: fixture.provider.pr.Repository, Branch: "dev"}

	ports := workflowFrontierPorts(fixture)
	ports.Target = workflowClosingDeadlineTarget{ports.Target}

	closed, err := CloseSet(t.Context(), fixture.store, request, ports)
	if err != nil {
		t.Fatal(err)
	}

	checkWorkflowClosingEmpty(t, closed, required)

	if required && (len(closed.Receipt.Validation) != 1 || closed.Receipt.Validation[0].Stdout != fixture.target+"\n") {
		t.Fatal("closing gate used another checkout or target")
	}
}

func checkWorkflowClosingEmpty(t *testing.T, closed workflowcloseout.Result, required bool) {
	t.Helper()

	if closed.Outcome != "completed" || len(closed.Receipt.Members) != 0 || len(closed.Receipt.Slices) != 0 {
		t.Fatal("explicit empty-set decision was lost")
	}

	if !required && closed.Receipt.Waiver == "" {
		t.Fatal("empty-set closure lost its explicit waiver")
	}
}

// TestWorkflowClosingActualMembers rejects closure when frozen admission has
// excluded approved later work. A waiver cannot substitute for its integration.
func TestWorkflowClosingActualMembers(t *testing.T) {
	fixture, ports := workflowClosingFixture(t)

	set, err := fixture.store.Read(t.Context(), fixture.hierarchy.Set.Path)
	if err != nil {
		t.Fatal(err)
	}

	request := workflowApprovedAmendment(t, fixture.store, set)
	workflowAppendPlanSlice(&request.Plan)

	request.Plan.ClosingValidation = workflowstate.Validation{WaiverReason: "Approved closing waiver"}
	if _, err := fixture.store.AmendWorkflowSet(t.Context(), request); err != nil {
		t.Fatal(err)
	}

	if _, err := CloseInvocationSet(t.Context(), fixture.store, fixture.run.ID, "dev", ports); err == nil {
		t.Fatal("frozen exclusion or waiver hid an actual unfinished member")
	}

	current, err := fixture.store.Read(t.Context(), set.Path)
	if err != nil || current.Set.ClosureRef != "" || current.Set.Closing != nil {
		t.Fatal("rejected actual membership created closure evidence", err)
	}
}

// workflowClosingDeadlineTarget rejects admission or confirmation observations
// without the host-owned finite deadline; it delegates actual Git observation.
type workflowClosingDeadlineTarget struct {
	TargetGit
}

func (target workflowClosingDeadlineTarget) ObserveWorkflowTarget(ctx context.Context, query TargetQuery) (workflowstate.PreviewTarget, error) {
	deadline, present := ctx.Deadline()
	if !present || time.Until(deadline) > 3*time.Minute {
		return workflowstate.PreviewTarget{}, fmt.Errorf("closing observation has no finite host deadline")
	}

	return target.TargetGit.ObserveWorkflowTarget(ctx, query)
}
