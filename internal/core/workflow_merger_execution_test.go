package core

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func nativeMergerExecution(t *testing.T, id string) (workflowPreparationFixture, WorkflowMergerRequest, WorkflowMergerHarness) {
	t.Helper()

	f, diagnostic := nativeMergerConflictFixture(t)
	entry := nativeRoleAdmit(t, f.store, f.run.ID)
	entry = nativeRoleTransition(t, f.store, entry, "preparing")
	entry = nativeRoleTransition(t, f.store, entry, "merging")
	run := workflowTestRead(t, f.store, f.run.ID)
	now := time.Now().UTC()
	input := WorkflowMergerInput{ProtocolVersion: 1, ID: id, InvocationID: run.ID, EntryID: entry.ID, Unit: entry.Unit, CandidateRevision: entry.StateRevision, CandidateHash: run.CandidateHash, ClaimGeneration: entry.ClaimGeneration, ScopeHash: entry.FrozenScopeHash, Reason: "preparation_conflict", Repository: entry.Repository, SourceRef: run.Lane.Branch, TargetRef: entry.TargetRef, Source: run.Lane.HeadCommit, Base: f.base, Method: "merge", AllowedPaths: []string{"result.txt"}, ProtectedPaths: []string{"Makefile"}, ValidationCommands: workflowPreparationGate("test -f result.txt").Commands, CapacityReservationID: "pending-capacity", BudgetReservationID: id, RemainingAttempts: 3, WorkerTimeoutSeconds: run.Policy.WorkerTimeoutSeconds, AdmittedAt: now, Deadline: now.Add(15 * time.Second)}
	input.OwnedWorkspaceID = workflowMergerWorkspaceID(input)
	manifest := WorkflowNativeSourceManifest{Schema: "chr-native-source/v1", Repository: input.Repository, SourceRef: input.SourceRef, OriginalSource: input.Source, Commits: strings.Fields(workflowIntegrationCommand(t, f.run.Lane.Worktree, "rev-list", "--reverse", f.run.Lane.BaseCommit+".."+input.Source))}
	input.SourceManifestRef = retainNativeRecordEvidence(t, f.store, run.ID, "execution-source", manifest)
	input.ValidationContractRef = retainNativeRecordEvidence(t, f.store, run.ID, "execution-gate", workflowPreparationGate("test -f result.txt"))
	source := retainNativeRecordEvidence(t, f.store, run.ID, "execution-intent", map[string]string{"intent": "Preserve source behavior"})
	target := retainNativeRecordEvidence(t, f.store, run.ID, "execution-target", map[string]string{"intent": "Preserve target validation"})
	input.IntentRefs, input.TargetChangeRefs = []WorkflowEvidenceRef{source}, []WorkflowEvidenceRef{target}
	input.Requirements = []WorkflowMergerRequirement{{ID: "source-api", Origin: "source", Evidence: source}, {ID: "target-validation", Origin: "target", Evidence: target}}
	input.Conflicts = []WorkflowMergerConflict{{ID: "conflict-1", Path: "result.txt", Kind: "text", Evidence: diagnostic, RequirementIDs: []string{"source-api", "target-validation"}}}

	input, err := SealWorkflowMergerInput(input, now)
	if err != nil {
		t.Fatal(err)
	}

	harness := nativeMergerTestHarness(t, f.store, run.ID)

	return f, WorkflowMergerRequest{WorkflowIntegrationBudgetGuard: nativeBudgetGuard(t, f.store, run.ID), Input: input}, harness
}

func nativeMergerTestHarness(t *testing.T, store *WorkflowStore, id string) WorkflowMergerHarness {
	t.Helper()

	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		t.Fatal(err)
	}

	file, err := os.Open(executable)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = file.Close() }() //nolint:errcheck

	hash, err := workflowMergerFileHash(file)
	if err != nil {
		t.Fatal(err)
	}

	command := []string{executable, "-test.run=^TestNativeMergerChild$"}
	certificate := WorkflowMergerHarnessCertificate{SchemaVersion: 1, Protocol: workflowMergerBrokerProtocol, Command: command, ExecutableHash: hash, CapabilityRevision: workflowHash([]byte("fixture-broker/v1")), BuiltinToolsDisabled: true}

	return WorkflowMergerHarness{Command: command, Certificate: retainNativeRecordEvidence(t, store, id, "execution-harness", certificate)}
}

// TestNativeMergerExecution uses a real driver and Git checkout. Denied model
// requests cannot publish, edit control files or supply new validation commands.
func TestNativeMergerExecution(t *testing.T) {
	f, request, harness := nativeMergerExecution(t, "merger-execution")
	source := workflowIntegrationCommand(t, f.run.Lane.Worktree, "status", "--porcelain")
	head := workflowIntegrationCommand(t, f.run.Lane.Worktree, "rev-parse", "HEAD")

	attempt, err := f.store.ExecuteWorkflowMerger(t.Context(), request, harness, f.provider)
	if err != nil || attempt.Proposal == nil || attempt.Proposal.Outcome != "resolved" || len(attempt.Checks) != 0 {
		t.Fatal("bounded merger failed", attempt.Failure, err)
	}

	if got := workflowIntegrationCommand(t, f.run.Lane.Worktree, "status", "--porcelain"); got != source || workflowIntegrationCommand(t, f.run.Lane.Worktree, "rev-parse", "HEAD") != head {
		t.Fatal("merger changed the worker checkout")
	}

	checkNativeMergerRetained(t, f, attempt)
}

func checkNativeMergerRetained(t *testing.T, f workflowPreparationFixture, result WorkflowMergerAttempt) {
	t.Helper()

	run := workflowTestRead(t, f.store, f.run.ID)

	attempt := run.Attempts[len(run.Attempts)-1]
	if attempt.Purpose != workflowMergerPurpose || attempt.Outcome != "checkpointed" || run.Outcome != "merger_proposal_pending" || attempt.Lane.Worktree == run.Lane.Worktree {
		t.Fatal("proposal acquired candidate authority or reused the worker checkout")
	}

	if workflowIntegrationCommand(t, attempt.Lane.Worktree, "rev-parse", "HEAD") != result.Input.Source {
		t.Fatal("driver committed its proposal")
	}

	if workflowIntegrationCommand(t, attempt.Lane.Worktree, "remote") != "" {
		t.Fatal("publication remote was exposed")
	}

	checkNativeMergerBudget(t, f)
	checkNativeMergerProcesses(t, attempt)
	checkNativeMergerFrozenHarness(t, f.store, attempt, result.HarnessRef)
}

func checkNativeMergerFrozenHarness(t *testing.T, store *WorkflowStore, attempt WorkflowAttempt, reference WorkflowEvidenceRef) {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(store.layout.Root(), attempt.InputRef))
	if err != nil {
		t.Fatal(err)
	}

	var input WorkflowTaskInput
	if err := json.Unmarshal(data, &input); err != nil {
		t.Fatal(err)
	}

	if workflowHash(data) != attempt.InputHash || input.MergerHarnessRef == nil || *input.MergerHarnessRef != reference {
		t.Fatal("driver certificate was not frozen before launch")
	}
}

func checkNativeMergerBudget(t *testing.T, f workflowPreparationFixture) {
	t.Helper()

	budget, err := f.store.ReadWorkflowIntegrationBudget(t.Context(), f.run.ID)
	if err != nil || len(budget.Reservations) != 1 || budget.Reservations[0].State != "unknown" || budget.UnitCounters[0].Merger != 1 || budget.TokensReserved != 0 {
		t.Fatal("unmetered attempt was lost or represented as a token grant", err)
	}

	restarted := NewWorkflowStore(f.store.layout, f.store.projectID)

	budget, err = restarted.ReadWorkflowIntegrationBudget(t.Context(), f.run.ID)
	if err != nil || budget.UnitCounters[0].Merger != 1 {
		t.Fatal("restart reset consumed merger calls", err)
	}
}

func checkNativeMergerProcesses(t *testing.T, attempt WorkflowAttempt) {
	t.Helper()

	if len(attempt.Processes) != 1 || workflowAttemptQuiescent(attempt) != nil {
		t.Fatal("driver or local check retained a live process")
	}

	for _, process := range attempt.Processes {
		if process.State != "stopped" || process.StdoutHash == "" || process.StderrHash == "" {
			t.Fatal("host process evidence was not retained")
		}
	}
}

// TestNativeMergerAdmission rejects stale scope and generic worker drivers before
// an attempt, owned checkout or budget reservation can be created.
func TestNativeMergerAdmission(t *testing.T) {
	f, original, harness := nativeMergerExecution(t, "merger-admission")

	before := nativeQueueImages(t, f.store.layout.Root())

	cases := map[string]func(*WorkflowMergerRequest, *WorkflowMergerHarness){
		"candidate": func(r *WorkflowMergerRequest, _ *WorkflowMergerHarness) {
			r.Input.CandidateHash = workflowHash([]byte("stale"))
		},
		"source": func(r *WorkflowMergerRequest, _ *WorkflowMergerHarness) { r.Input.Source = strings.Repeat("f", 40) },
		"base":   func(r *WorkflowMergerRequest, _ *WorkflowMergerHarness) { r.Input.Base = strings.Repeat("f", 40) },
		"scope": func(r *WorkflowMergerRequest, _ *WorkflowMergerHarness) {
			r.Input.ScopeHash = workflowHash([]byte("new-scope"))
		},
		"generation": func(r *WorkflowMergerRequest, _ *WorkflowMergerHarness) { r.Input.ClaimGeneration++ },
		"method":     func(r *WorkflowMergerRequest, _ *WorkflowMergerHarness) { r.Input.Method = "squash" },
		"commands": func(r *WorkflowMergerRequest, _ *WorkflowMergerHarness) {
			r.Input.ValidationCommands = [][]string{{"git", "push"}}
		},
		"harness": func(_ *WorkflowMergerRequest, h *WorkflowMergerHarness) {
			h.Command = []string{"codex", "exec", "--dangerously-bypass-approvals-and-sandbox"}
		},
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			request, err := freezeWorkflowMergerRequest(original)
			if err != nil {
				t.Fatal(err)
			}

			driver := harness
			change(&request, &driver)

			request.Input, err = SealWorkflowMergerInput(request.Input, time.Now().UTC())
			if err != nil {
				t.Fatal(err)
			}

			if _, err := f.store.ExecuteWorkflowMerger(t.Context(), request, driver, f.provider); !errors.Is(err, ErrWorkflowConflict) {
				t.Fatal("invalid admission dispatched", err)
			}
		})
	}

	if !workflowNativeJSONEqual(before, nativeQueueImages(t, f.store.layout.Root())) {
		t.Fatal("rejected admission mutated retained records")
	}
}

func nativeMergerChildRequest(encoder *json.Encoder, decoder *json.Decoder, request workflowMergerToolRequest) workflowMergerToolResponse {
	if err := encoder.Encode(request); err != nil {
		os.Exit(71)
	}

	var response workflowMergerToolResponse
	if err := decoder.Decode(&response); err != nil {
		os.Exit(72)
	}

	return response
}

// TestNativeMergerChild is a trusted model-driver fixture. Model messages can use
// only the broker; the driver performs no implicit repository or shell tools.
func TestNativeMergerChild(t *testing.T) {
	if !strings.Contains(strings.Join(os.Args, " "), "-test.run=^TestNativeMergerChild$") {
		return
	}

	decoder, encoder := json.NewDecoder(os.Stdin), json.NewEncoder(os.Stdout)

	var call workflowMergerCall
	if err := decoder.Decode(&call); err != nil {
		os.Exit(70)
	}

	if call.Input.ID == "merger-timeout" {
		if err := os.WriteFile("driver-started", []byte("owned driver"), 0o600); err != nil {
			os.Exit(77)
		}

		time.Sleep(30 * time.Second)
		os.Exit(77)
	}

	for _, request := range []workflowMergerToolRequest{{Operation: "push"}, {Operation: "merge"}, {Operation: "approve"}, {Operation: "receipt"}, {Operation: "write", Path: ".git/config", Content: "remote"}, {Operation: "write", Path: "ops/default/ticket/ready/scope.md", Content: "changed scope"}, {Operation: "check", Check: new(int), Content: "git push"}} {
		if response := nativeMergerChildRequest(encoder, decoder, request); response.Error == "" {
			os.Exit(73)
		}
	}

	if response := nativeMergerChildRequest(encoder, decoder, workflowMergerToolRequest{Operation: "write", Path: "result.txt", Content: "Preserved source and target behavior\n"}); response.Error != "" {
		os.Exit(74)
	}

	if response := nativeMergerChildRequest(encoder, decoder, workflowMergerToolRequest{Operation: "check", Check: new(int)}); response.Error == "" {
		os.Exit(75)
	}

	proposal := nativeMergerResolved(call.Input)
	proposal.ChangedPaths[0], proposal.ResolutionNotes[0].Path = "result.txt", "result.txt"

	data, err := json.Marshal(proposal)
	if err != nil {
		os.Exit(76)
	}

	nativeMergerChildDescendant(call.Input.ID)

	_ = nativeMergerChildRequest(encoder, decoder, workflowMergerToolRequest{Operation: "result", Result: data, Usage: nativeMergerChildMeter(call.Input)})

	os.Exit(0)
}

func nativeMergerChildDescendant(id string) {
	if id != "merger-orphan" {
		return
	}

	child := exec.Command("sleep", "2") //nolint:contextcheck
	if err := child.Start(); err != nil {
		os.Exit(78)
	}

	if err := os.WriteFile("child.pid", []byte("owned child"), 0o600); err != nil {
		os.Exit(79)
	}
}

func nativeMergerConflictFixture(t *testing.T) (workflowPreparationFixture, WorkflowEvidenceRef) {
	t.Helper()

	f := newWorkflowPreparationFixture(t, "merge", workflowPreparationGate("test -f result.txt"))
	f.base = workflowIntegrationCommit(t, f.actor, "result.txt", "Independent target behavior")
	workflowIntegrationCommand(t, f.actor, "push", "origin", "dev")

	f.provider.pr.BaseCommit, f.provider.target = f.base, f.base
	workflowIntegrationCommand(t, f.actor, "fetch", "origin", f.run.Lane.HeadCommit)

	command := exec.Command("git", "-C", f.actor, "merge-tree", "--write-tree", f.base, f.run.Lane.HeadCommit) //nolint:contextcheck

	data, err := command.CombinedOutput()
	if err == nil || !strings.Contains(string(data), "CONFLICT") {
		t.Fatal("fixture did not demonstrate a deterministic conflict", err)
	}

	ref := retainNativeRecordEvidence(t, f.store, f.run.ID, "execution-conflict", map[string]string{"diagnostic": string(data)})

	return f, ref
}

func nativeMergerChildMeter(input WorkflowMergerInput) *WorkflowMergerUsage {
	if input.ID != "merger-metered" {
		return nil
	}

	return &WorkflowMergerUsage{Input: 20, Output: 10, WholeCall: true, CapabilityRevision: input.TokenAllowance.CapabilityRevision}
}
