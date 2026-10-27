package app

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	generationcore "forge.adrianpk.com/chr/internal/core/authoring/generation"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	quizcore "forge.adrianpk.com/chr/internal/core/quiz"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// Both transports exercise the same real authoring retry and publication owner.
func TestPiGenerationAuthoring(t *testing.T) {
	generationBindingAuthoringCase(t, piGenerationFixture)
}

// Pi retains quiz continuation identity across role changes and service recovery.
func TestPiGenerationQuiz(t *testing.T) {
	generationBindingQuizCase(t, piGenerationFixture)
}

// Host intent routing preserves project access checks and selected Pi identity.
func TestPiGenerationConversation(t *testing.T) {
	generationBindingConversationCase(t, piGenerationFixture)
}

// The API-owned plan stages finish before a fresh operation dispatches to Pi.
func TestPiGenerationPlanSwitch(t *testing.T) {
	generationBindingPlanCase(t, piGenerationFixture)
}

// Fast mode keeps buffered questions and does not create extra provider work.
func TestPiGenerationFastQuiz(t *testing.T) {
	question := generationQuestionFixture("Which scope must be retained?")
	output := `{"questions":[` + question + `,` + generationQuestionFixture("How will that scope be validated?") + `]}`
	operations, driver := piGenerationFixture(t, output)
	writeServiceQuizFixture(t, operations.projectContext.ProjectDir)

	started, err := operations.StartQuiz(QuizStartRequest{Mode: quizcore.ModeFast})
	if err != nil || started.State.ActiveQuestion == nil {
		t.Fatal("Pi fast batch failed", err)
	}

	answered, err := operations.AnswerQuiz(QuizAnswerRequest{Answer: "Preserve the bounded operation owner."})
	if err != nil || answered.State.ActiveQuestion == nil || answered.State.ContinuationPending {
		t.Fatal("Pi discarded buffered fast question", err)
	}

	if calls := piGenerationCalls(t, driver); len(calls) != 1 {
		t.Fatal("buffered answer created another Pi call", len(calls))
	}
}

// Existing feature validation rejects bad Pi output before replacing a valid feature.
func TestPiGenerationAcceptance(t *testing.T) {
	output, err := json.Marshal(map[string]string{"status": generationcore.AcceptanceFeatureStatusGenerated, "body": validBlogAcceptanceFeatureBody()})
	if err != nil {
		t.Fatal(err)
	}

	operations, driver := piGenerationFixture(t, string(output))
	project := operations.projectContext.ProjectDir
	writeAcceptanceFeatureServiceFixture(t, project)

	request := AcceptanceFeatureGenerateRequest{OperationID: "pi-feature", Scope: quizcore.Scope{CaseSlug: "blog"}}

	result, err := operations.GenerateAcceptanceFeature(t.Context(), request)
	if err != nil || result.Status != generationcore.AcceptanceFeatureStatusGenerated {
		t.Fatal("Pi feature generation failed", err)
	}

	path := filepath.Join(project, result.Feature)

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	harnessSelectionWrite(t, filepath.Join(driver, "result.json"), `{"status":"generated","body":"Feature: invalid"}`)

	request.OperationID = "pi-feature-invalid"

	invalid, rejected := operations.GenerateAcceptanceFeature(t.Context(), request)
	if rejected == nil && invalid.Status == generationcore.AcceptanceFeatureStatusGenerated {
		t.Fatal("invalid Pi feature was accepted")
	}

	after, err := os.ReadFile(path)
	if err != nil || string(before) != string(after) {
		t.Fatal("rejected Pi feature changed canonical output", err)
	}
}

// Curation remains under the mutation owner, with no fallback after explicit Pi failure.
func TestPiGenerationDump(t *testing.T) {
	operations, driver := piGenerationFixture(t, `{"polished_body":"Retain the prepared scope.","highlights":["Bounded scope."],"warnings":[]}`)
	project := operations.projectContext.ProjectDir
	writeServiceWorkspaceQueryFixture(t, project)

	if _, err := operations.CreateDump(DumpCreateRequest{OperationID: "pi-dump", Title: "Prepared Notes", Content: "raw bounded scope"}); err != nil {
		t.Fatal(err)
	}

	harnessSelectionWrite(t, filepath.Join(driver, "result.json"), `{"polished_body":"","highlights":[],"warnings":[]}`)

	if _, err := operations.CreateDump(DumpCreateRequest{OperationID: "pi-dump-invalid", Title: "Invalid Notes", Content: "raw invalid scope"}); err == nil {
		t.Fatal("invalid Pi curation fell back to deterministic publication")
	}

	checkPiGenerationMetadata(t, project, 1, 1)
}

// Unsupported retained protocols fail before launching even an otherwise ready profile.
func TestPiGenerationProtocolIdentity(t *testing.T) {
	operations, driver := piGenerationFixture(t, `{}`)

	config, err := core.LoadWorkspaceConfig(operations.projectContext.ProjectDir)
	if err != nil {
		t.Fatal(err)
	}

	prepared, err := operations.prepareGenerationBinding(t.Context(), config, "conversation")
	if err != nil {
		t.Fatal(err)
	}

	prepared.Binding.ProtocolRevision = "pi-rpc/future"

	prepared.Binding, err = harnesscore.SealGenerationBinding(prepared.Binding)
	if err != nil {
		t.Fatal(err)
	}

	op, err := core.FreezeHarnessGeneration(t.Context(), harnesshost.GenerationFreezeInput{Project: operations.projectContext.ProjectDir, ID: "unsupported-protocol", Purpose: "conversation"}, func(context.Context, core.WorkspaceConfig) (harnesshost.GenerationPreparation, error) {
		return harnesshost.GenerationPreparation{Binding: prepared.Binding, Launch: prepared.Launch}, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := operations.bindGeneration(harnesshost.WithGeneration(t.Context(), op), "ignored", "conversation"); err == nil || errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("unsupported Pi protocol did not reach its dispatch fence", err)
	}

	if _, err := os.Stat(filepath.Join(driver, "version-called")); !os.IsNotExist(err) {
		t.Fatal("unsupported protocol spawned Pi", err)
	}
}

// The real Spec owner validates Pi output and preserves an earlier artifact on rejection.
func TestPiGenerationSpec(t *testing.T) {
	operations, driver := piGenerationFixture(t, `{"title":"Bounded Spec","body":"Preserve the selected operation scope.","status":"draft"}`)
	project := operations.projectContext.ProjectDir
	writeServiceQuizFixture(t, project)

	request := SpecGenerateRequest{OperationID: "pi-spec", Scope: quizcore.Scope{CaseSlug: "blog"}}

	result, err := operations.GenerateSpec(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(project, result.Spec.RelativePath)

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	harnessSelectionWrite(t, filepath.Join(driver, "result.json"), `{"title":"Invalid","body":"","status":"draft"}`)

	request.OperationID = "pi-invalid"
	if _, err := operations.GenerateSpec(t.Context(), request); err == nil {
		t.Fatal("invalid Pi output replaced a valid Spec")
	}

	after, err := os.ReadFile(path)
	if err != nil || string(before) != string(after) {
		t.Fatal("rejected Pi output changed canonical artifact", err)
	}

	checkPiGenerationMetadata(t, project, 1, 1)
}

func piGenerationFixture(t *testing.T, output string) (service, string) {
	t.Helper()

	operations, project, configDir, _ := harnessSelectionFixture(t)
	operations.assetsFS = testAssetsFS()

	dir := t.TempDir()

	peer, err := filepath.Abs("../harness/pi/testdata/rpc-peer.py")
	if err != nil {
		t.Fatal(err)
	}

	python, err := exec.LookPath("python3")
	if err != nil {
		t.Fatal(err)
	}

	quote := func(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'" }
	driver := filepath.Join(dir, "driver")
	harnessSelectionWrite(t, driver, "#!/bin/sh\nexec "+quote(python)+" "+quote(peer)+" "+quote(dir)+" \"$@\"\n")

	if err := os.Chmod(driver, 0o700); err != nil {
		t.Fatal(err)
	}

	harnessSelectionWrite(t, filepath.Join(dir, "result.json"), output)
	harnessSelectionWrite(t, filepath.Join(configDir, "auth.json"), `{"fixture":{"type":"api_key","key":"fixture-only-key"},"unrelated":{"type":"api_key","key":"unused-fixture-key"}}`)

	profiles := "profiles:\n  writing:\n    harness: pi\n    command: " + driver + "\n    model_provider: fixture\n    model: fixture-model\n    thinking: high\n    credential_ref: auth.json\n  other:\n    harness: pi\n    command: " + driver + "\n    model_provider: fixture\n    model: other-model\n    thinking: low\n    credential_ref: auth.json\n"
	harnessSelectionWrite(t, filepath.Join(configDir, "harnesses.yml"), profiles)
	harnessSelectionWrite(t, workspacecore.UserPath(project, "config.yml"), "harnesses:\n  generation: writing\n")

	return operations, dir
}

func checkPiGenerationMetadata(t *testing.T, project string, valid, invalid int) {
	t.Helper()

	for _, record := range generationFixtureMetadata(t, project) {
		if !validPiGenerationMetadata(record) {
			t.Fatal("Pi lost host run evidence", record)
		}

		switch record.OutputValidation {
		case "valid":
			valid--
		case "invalid":
			invalid--
		default:
			t.Fatal("Pi parser outcome was not retained", record)
		}
	}

	if valid != 0 || invalid != 0 {
		t.Fatal("unexpected Pi validation counts", valid, invalid)
	}
}

func validPiGenerationMetadata(record harnesshost.RunMetadata) bool {
	return record.Selection.Selection.Harness == "pi" && record.ProcessOutcome == "completed" && record.ObservedModel == "fixture/fixture-model" && record.Usage != nil && record.UsageStatus == "known"
}

func piGenerationCalls(t *testing.T, driver string) []map[string]any {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(driver, "calls", "*.json"))
	if err != nil {
		t.Fatal(err)
	}

	var calls []map[string]any

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		var call map[string]any
		if err := json.Unmarshal(data, &call); err != nil {
			t.Fatal(err)
		}

		calls = append(calls, call)
	}

	return calls
}
