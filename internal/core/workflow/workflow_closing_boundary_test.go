package workflow

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowcloseout "forge.adrianpk.com/chr/internal/core/workflow/closeout"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowClosingTargetHook struct {
	TargetGit
	calls int
	hook  func()
}

func (target *workflowClosingTargetHook) ObserveWorkflowTarget(ctx context.Context, query TargetQuery) (workflowstate.PreviewTarget, error) {
	target.calls++
	if target.calls == 2 {
		target.hook()
	}

	return target.TargetGit.ObserveWorkflowTarget(ctx, query)
}

// TestWorkflowClosingChanges prevents confirmation when canonical target or
// actual member bytes change after the frozen gate input was accepted.
func TestWorkflowClosingChanges(t *testing.T) {
	for _, change := range []string{"target", "member"} {
		t.Run(change, func(t *testing.T) {
			fixture, ports := workflowClosingFixture(t)
			gate := workflowstate.Validation{Required: true, TimeoutSeconds: 30, Commands: [][]string{{"git", "rev-parse", "HEAD"}}}

			if change == "target" {
				ports.Target = &workflowClosingTargetHook{TargetGit: ports.Target, hook: func() {
					workflowIntegrationCommit(t, fixture.actor, "movement.txt", "Concurrent closing target movement")
					workflowIntegrationCommand(t, fixture.actor, "push", "origin", "dev")
				}}
			} else {
				path := filepath.Join(fixture.store.layout.Root(), filepath.Dir(fixture.hierarchy.Tasks[0].Path), "artifact.md")
				gate.Commands = [][]string{{"sh", "-c", "printf '\\nChanged member during validation\\n' >>\"$1\"", "closing-gate", path}}
			}

			workflowClosingContract(t, fixture, gate)

			result, err := CloseInvocationSet(t.Context(), fixture.store, fixture.run.ID, "dev", ports)
			if !errors.Is(err, workflowstate.ErrConflict) || result.Outcome != "closing_validation_failed" || result.Set.Set.ClosureRef != "" {
				t.Fatal("changed closing input or target established completion", err)
			}

			if len(result.Set.Set.Closing.Attempts) != 1 || result.Set.Set.Status == "closing" {
				t.Fatal("changed input lost its failed attempt or retained a live closing owner")
			}
		})
	}
}

// TestWorkflowClosingUseCase closes two empty sets in declared order through
// one frozen invocation. Unsliced work stays excluded and no task worker runs.
func TestWorkflowClosingUseCase(t *testing.T) {
	fixture := workflowDeliveryFixture(t, "rebase")
	sets := []workflowstate.Record{}

	for _, order := range []int{2, 1} {
		plan := workflowPlanFixture()

		set, err := fixture.store.Create(t.Context(), "Ordered empty set", workflowstate.Record{Kind: "plan", Set: &workflowstate.WorkflowSet{Schema: workflowstate.WorkflowSchema, Status: "ready", UseCase: "empty-case", UseCaseOrder: order, PlanRef: plan.PlanRef, PlanHash: plan.PlanHash, ClosingValidation: workflowstate.Validation{WaiverReason: "Approved ordered empty-set closure"}}})
		if err != nil {
			t.Fatal(err)
		}

		sets = append(sets, set)
	}

	start := StartRequest{StartInput: workflowstate.StartInput{
		RequestKey: "empty-case",
		Selection:  workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "use_case", Identity: "empty-case"}, Target: workflowstate.PreviewTarget{Repository: fixture.provider.pr.Repository, Commit: fixture.target}},
	}}
	run := workflowTestStart(t, fixture.store, start)
	ports := ContinuationPorts{Publication: PublicationPorts{Provider: workflowFrontierPorts(fixture).Provider, Observer: fixture.git}, Target: fixture.git}

	first := workflowContinueOne(t, fixture.store, run.Invocation.ID, ports)
	if len(first.Invocation.Closures) != 1 || first.Invocation.Closures[0].SetID != sets[1].ID || first.Invocation.Status == workflowstate.InvocationCompleted {
		t.Fatal("empty set order or closing frontier was lost")
	}

	second := workflowContinueOne(t, fixture.store, run.Invocation.ID, ports)
	checkWorkflowClosingUseCase(t, second.Invocation, sets[0].ID, run.Invocation.RequestHash)
}

// TestWorkflowClosingReopen preserves a real confirmed receipt and the terminal
// invocation while an explicit changed-plan amendment opens another set cycle.
func TestWorkflowClosingReopen(t *testing.T) {
	fixture, ports := workflowClosingFixture(t)
	workflowClosingContract(t, fixture, workflowstate.Validation{WaiverReason: "Approved closure waiver"})

	closed, err := CloseInvocationSet(t.Context(), fixture.store, fixture.run.ID, "dev", ports)
	if err != nil {
		t.Fatal(err)
	}

	before := workflowReadTestBytes(t, fixture.store, workflowInvocationPath(fixture.run.ID))
	receipt := workflowReadTestBytes(t, fixture.store, closed.ReceiptRef)
	request := workflowApprovedAmendment(t, fixture.store, closed.Set)
	request.Reopen = true
	request.Plan.PlanRef = "plan/reopening.md"

	request.Plan.PlanHash = contenthash.Sum([]byte("approved reopening\n"))
	if err := os.WriteFile(filepath.Join(fixture.store.layout.Root(), request.Plan.PlanRef), []byte("approved reopening\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	changed, err := fixture.store.AmendWorkflowSet(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	checkWorkflowClosingReopen(t, fixture, changed.Set, closed, before, receipt)
}

func checkWorkflowClosingReopen(t *testing.T, fixture workflowIntegrationFixture, reopened workflowstate.Record, closed workflowcloseout.Result, before, receipt []byte) {
	t.Helper()

	if reopened.Set.Status != "ready" || reopened.Set.ClosureRef != "" || len(reopened.Set.ClosureHistory) != 1 || reopened.Set.ClosureHistory[0] != closed.ReceiptRef {
		t.Fatal("explicit reopening lost prior confirmed closure")
	}

	if !bytes.Equal(before, workflowReadTestBytes(t, fixture.store, workflowInvocationPath(fixture.run.ID))) || !bytes.Equal(receipt, workflowReadTestBytes(t, fixture.store, closed.ReceiptRef)) {
		t.Fatal("reopening rewrote terminal invocation or receipt")
	}
}

func checkWorkflowClosingUseCase(t *testing.T, second workflowstate.Invocation, setID, requestHash string) {
	t.Helper()

	if second.Status != workflowstate.InvocationCompleted || len(second.Closures) != 2 || second.Closures[1].SetID != setID || len(second.Attempts) != 0 || second.RequestHash != requestHash {
		t.Fatal("empty use-case continuation changed admission or executed a task")
	}
}
