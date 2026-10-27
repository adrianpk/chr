package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	"forge.adrianpk.com/chr/internal/harness/pi"
)

type harnessLivePiOptions struct{ root, provider, model, native, bridge, image, auth string }

// This opt-in is the separately authorized real-provider proof. Missing access
// fails after opt-in; deterministic peers and ordinary journey runs cannot
// satisfy it. Only the selected static credential enters the private fixture.
func TestHarnessLivePi(t *testing.T) {
	if os.Getenv("CHR_LIVE_PI") != "1" {
		t.Skip("requires explicit authorization for one real generation and worker task")
	}

	options := harnessLivePiConfiguration(t)
	plan := serviceWorkflowPlanFixture()
	task := &plan.Slices[0].Tasks[0]
	task.Title = "Produce one disposable proof file"
	task.Body = "Create result.txt containing exactly Delivered followed by a newline. Stage only result.txt and commit it with git -c user.name=Fixture -c user.email=fixture@example.com commit -m 'Live Pi proof'. Do not edit other files, push, or contact a remote repository. Return a short result after committing."
	task.Outcome, task.Scope = "One host-validated Git commit containing result.txt", "Disposable local repository only"
	operations, hierarchy, target := serviceWorkflowTaskFixture(t, plan)

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{DevMode: true, DevRoot: options.root})
	if err != nil {
		t.Fatal(err)
	}

	operations.env = env
	operations.generationBinder = service.bindHarnessGeneration
	operations.conversationModel = core.NewConfiguredStructuredModel(false)
	prepareHarnessLivePi(t, operations, options)

	status, err := operations.InspectHarnessSelections(t.Context(), "spec_generation")
	if err != nil || !status.Worker.Readiness.Ready() || !status.Generation.Readiness.Ready() {
		t.Fatal("prepare both live roles and Docker access before any model call", err)
	}

	request := core.WorkflowStartRequest{StartInput: workflowstate.StartInput{
		RequestKey: "live-pi-worker",
		Policy:     workflowstate.ExecutionPolicy{WorkerTimeoutSeconds: 90, MaxWorkerAttempts: 1, CheckpointStepLimit: 1},
		Selection:  workflowstate.PreviewRequest{Selector: workflowstate.Selector{Extent: "ticket", Identity: hierarchy.Tasks[0].Task.Key}, Target: workflowstate.PreviewTarget{Repository: "https://forge.example/fixture/disposable", Commit: target}},
	}}
	generation := harnessLivePiGeneration(t, operations)
	started, result := harnessLivePiWorker(t, operations, request)
	writeHarnessLivePiEvidence(t, options, generation, started, result)
}

func harnessLivePiConfiguration(t *testing.T) harnessLivePiOptions {
	t.Helper()

	options := harnessLivePiOptions{root: os.Getenv("CHR_LIVE_PI_ROOT"), provider: os.Getenv("CHR_LIVE_PI_PROVIDER"), model: os.Getenv("CHR_LIVE_PI_MODEL"), native: os.Getenv("CHR_LIVE_PI_NATIVE"), bridge: os.Getenv("CHR_LIVE_PI_BRIDGE"), image: os.Getenv("CHR_LIVE_PI_IMAGE"), auth: os.Getenv("CHR_LIVE_PI_AUTH")}
	for _, path := range []string{options.root, options.native, options.bridge, options.auth} {
		if !filepath.IsAbs(path) {
			t.Fatal("live proof requires absolute private fixture, executable and auth paths")
		}
	}

	info, err := os.Stat(options.root)
	if err != nil || !info.IsDir() || info.Mode().Perm()&0o077 != 0 || options.image == "" {
		t.Fatal("prepare a private live fixture root and explicit runtime image")
	}

	if err := pi.ValidateGenerationProfile(options.provider, options.model, "off"); err != nil {
		t.Fatal(err)
	}

	return options
}

func prepareHarnessLivePi(t *testing.T, operations service, options harnessLivePiOptions) {
	t.Helper()

	auth := harnessLivePiCredential(t, options)

	directory := filepath.Join(operations.env.ConfigDir(), "live-provider")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}

	harnessSelectionWrite(t, filepath.Join(directory, "auth.json"), string(auth))

	profiles := "profiles:\n"

	for _, role := range []struct{ name, program string }{{"proof-generation", options.native}, {"proof-worker", options.bridge}} {
		profile := harnesscore.HarnessProfile{Name: role.name, Harness: "pi", Command: role.program, ModelProvider: options.provider, Model: options.model, Thinking: "off", CredentialRef: "live-provider/auth.json"}
		if role.name == "proof-worker" {
			profile.RuntimeRef = options.image
		}

		profiles += harnessLivePiProfile(profile)
	}

	harnessSelectionWrite(t, filepath.Join(operations.env.ConfigDir(), "harnesses.yml"), profiles)
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "ticket:\n  key_prefix: PRJ\nharnesses:\n  generation: proof-generation\n  worker: proof-worker\n")
}

func harnessLivePiGeneration(t *testing.T, operations service) core.HarnessRunMetadata {
	t.Helper()

	bound, ctx, err := operations.bindGeneration(t.Context(), "live-pi-generation", "spec_generation")
	if err != nil {
		t.Fatal(err)
	}

	request := harnessJourneyRequest(bound, "live-structured-call")

	output, err := bound.conversationModel.InvokeStructuredModel(ctx, request)
	if err != nil {
		t.Fatal("real Pi generation failed", err)
	}

	if err := validateHarnessLivePiOutput(ctx, request, output); err != nil {
		t.Fatal("real Pi generation did not satisfy the typed result", err)
	}

	records := generationFixtureMetadata(t, operations.projectContext.ProjectDir)
	if len(records) != 1 || records[0].Selection.Selection.Harness != "pi" || records[0].OutputValidation != "valid" || records[0].ProcessOutcome != "completed" {
		t.Fatal("real generation lacks exact selection and caller validation evidence")
	}

	return records[0]
}

func harnessLivePiWorker(t *testing.T, operations service, request core.WorkflowStartRequest) (core.WorkflowInvocation, core.WorkflowTaskResult) {
	t.Helper()

	started, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal("real worker admission failed", err)
	}

	claim, err := operations.ClaimWorkflow(t.Context(), started.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := claim.Close(); err != nil {
			t.Error(err)
		}
	}()

	result, err := operations.ExecuteWorkflowTask(t.Context(), claim, nil)
	if err != nil || !result.Successful || result.Commit == request.Selection.Target.Commit {
		t.Fatal("real Docker worker did not pass the frozen host validation", result.Failure, err)
	}

	retained := harnessLivePiWorkerResult(t, operations, started.Invocation.ID, request, result)

	return retained, result
}

func writeHarnessLivePiEvidence(t *testing.T, options harnessLivePiOptions, generation core.HarnessRunMetadata, worker core.WorkflowInvocation, result core.WorkflowTaskResult) {
	t.Helper()

	candidate, err := exec.CommandContext(t.Context(), "git", "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}

	outputHash := sha256.Sum256([]byte("Delivered\n"))
	evidence := struct {
		Candidate, Provider, Model, PiVersion, Runtime, WorkerCommit, OutputHash string
		WorkerAttempts, WorkerDeadlineSeconds                                    int
		Generation                                                               core.HarnessRunMetadata
	}{strings.TrimSpace(string(candidate)), options.provider, options.model, "1.3.0", worker.HarnessBinding.RuntimeRef, result.Commit, hex.EncodeToString(outputHash[:]), len(worker.Attempts), worker.Policy.WorkerTimeoutSeconds, generation}

	data, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	harnessSelectionWrite(t, filepath.Join(options.root, "evidence.json"), string(data))
	t.Logf("real Pi proof passed: candidate=%s provider=%s model=%s runtime=%s worker_commit=%s generation_usage=%s", evidence.Candidate, options.provider, options.model, evidence.Runtime, evidence.WorkerCommit, generation.UsageStatus)
}

func harnessLivePiProfile(profile harnesscore.HarnessProfile) string {
	text := fmt.Sprintf("  %s:\n    harness: pi\n    command: %q\n    model_provider: %q\n    model: %q\n    thinking: off\n    credential_ref: %q\n", profile.Name, profile.Command, profile.ModelProvider, profile.Model, profile.CredentialRef)
	if profile.RuntimeRef != "" {
		text += fmt.Sprintf("    runtime_ref: %q\n", profile.RuntimeRef)
	}

	return text
}

func harnessLivePiCredential(t *testing.T, options harnessLivePiOptions) []byte {
	t.Helper()

	info, err := os.Stat(options.auth)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o077 != 0 || info.Size() > 65536 {
		t.Fatal("live credential source must be a private regular file within 64 KiB")
	}

	data, err := os.ReadFile(options.auth)
	if err != nil {
		t.Fatal("selected live authentication source is unavailable")
	}

	var credentials map[string]json.RawMessage
	if err := json.Unmarshal(data, &credentials); err != nil {
		t.Fatal("selected live authentication source is invalid")
	}

	selected, found := credentials[options.provider]
	if !found {
		t.Fatal("selected provider has no explicit live credential")
	}

	auth, err := json.Marshal(map[string]json.RawMessage{options.provider: selected})
	if err != nil {
		t.Fatal(err)
	}

	return auth
}

func validateHarnessLivePiOutput(ctx context.Context, request core.StructuredModelRequest, output []byte) error {
	var value map[string]bool

	validation := json.Unmarshal(output, &value)
	if validation == nil && (len(value) != 1 || !value["ok"]) {
		validation = fmt.Errorf("expected exactly one true ok field")
	}

	return core.ValidateStructuredModelOutput(ctx, request, validation)
}

func harnessLivePiWorkerResult(t *testing.T, operations service, id string, request core.WorkflowStartRequest, result core.WorkflowTaskResult) core.WorkflowInvocation {
	t.Helper()

	retained, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil || len(retained.Attempts) != 1 || retained.DeliveryRef != "" {
		t.Fatal("bounded worker attempt or delivery boundary changed", err)
	}

	checkHarnessLivePiCommit(t, retained, request.Selection.Target.Commit, result.Commit)

	return retained
}

func checkHarnessLivePiCommit(t *testing.T, retained core.WorkflowInvocation, before, after string) {
	t.Helper()

	if retained.HarnessBinding == nil || retained.HarnessBinding.Harness != "pi" || retained.Policy.MaxWorkerAttempts != 1 || retained.Policy.WorkerTimeoutSeconds != 90 {
		t.Fatal("real worker changed its selected harness or bounded policy")
	}

	output, err := exec.CommandContext(t.Context(), "git", "-C", retained.Lane.Worktree, "show", after+":result.txt").Output()
	if err != nil || string(output) != "Delivered\n" {
		t.Fatal("real worker result does not match the approved disposable task", err)
	}

	paths, err := exec.CommandContext(t.Context(), "git", "-C", retained.Lane.Worktree, "diff", "--name-only", before, after).Output()
	if err != nil || strings.TrimSpace(string(paths)) != "result.txt" {
		t.Fatal("real worker changed files outside the one approved output", err)
	}
}
