package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/urfave/cli/v3"

	"forge.adrianpk.com/chc/internal/core"
)

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	return Run(ctx, args, strings.NewReader(""), stdout, stderr, testAssetsFS())
}

func testAssetsFS() fs.FS {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("resolve test asset path")
	}

	root := filepath.Join(filepath.Dir(file), "..", "..", "cmd", "chc")

	return os.DirFS(root)
}

func testAppDependencies() appDependencies {
	return appDependencies{assetsFS: testAssetsFS()}
}

func TestRootCommandDefaultsToTUIAction(t *testing.T) {
	cmd := newRootCommand(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, testAppDependencies())
	if cmd.Action == nil {
		t.Fatal("root command Action is nil")
	}
}

func TestRootCommandIncludesTUICommand(t *testing.T) {
	cmd := newRootCommand(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, testAppDependencies())

	for _, name := range []string{"acceptance", "dump", "metadata", "plan", "quiz", "spec", "tui", "web", "work"} {
		if rootCommandIncludesCommand(cmd, name) {
			continue
		}

		t.Fatalf("root command does not include %s command", name)
	}
}

func rootCommandIncludesCommand(cmd *cli.Command, name string) bool {
	for _, child := range cmd.Commands {
		if child.Name == name {
			return true
		}
	}

	return false
}

func TestRootCommandRegistersParitySubcommands(t *testing.T) {
	cmd := newRootCommand(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, testAppDependencies())

	for _, path := range [][]string{
		{"metadata", "project", "set"},
		{"metadata", "case", "set"},
		{"dump", "list"},
		{"dump", "new"},
		{"dump", "show"},
		{"dump", "save"},
		{"dump", "append"},
		{"dump", "open"},
		{"gate", "dev-main"},
		{"gate", "woodpecker", "init"},
		{"quiz", "list"},
		{"quiz", "start"},
		{"quiz", "resume"},
		{"quiz", "answer"},
		{"quiz", "skip"},
		{"quiz", "complete"},
		{"quiz", "cancel"},
		{"quiz", "mode"},
		{"acceptance", "feature", "generate"},
		{"spec", "list"},
		{"spec", "generate"},
		{"plan", "list"},
		{"plan", "generate"},
		{"plan", "regenerate"},
		{"plan", "tracking", "create"},
		{"ticket", "board"},
	} {
		if rootCommandIncludesPath(cmd, path...) {
			continue
		}

		t.Fatalf("root command does not include %s", strings.Join(path, " "))
	}
}

func TestWebServeDryRunPrintsResolvedDefaults(t *testing.T) {
	projectDir := newTestWorkspace(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{"web", "serve", "--dir", projectDir, "--dry-run"}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	got := stdout.String()
	for _, want := range []string{
		"web ui dry-run: http://127.0.0.1:7331\n",
		"project: " + projectDir + "\n",
		"bind: 127.0.0.1\n",
		"port: 7331\n",
		"enabled: false\n",
		"open_browser: false\n",
		"require_token: auto\n",
		"token_env: CHC_WEB_TOKEN\n",
		"static_assets: healthz-only\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, got)
		}
	}
}

func TestWebServeDryRunAppliesFlagOverridesAndWarnsForNonLoopback(t *testing.T) {
	projectDir := newTestWorkspace(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"web",
		"serve",
		"--dir",
		projectDir,
		"--host",
		"0.0.0.0",
		"--port",
		"7444",
		"--open-browser",
		"--require-token",
		"false",
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	got := stdout.String()
	for _, want := range []string{
		"warning: web ui is bound to a non-loopback interface\n",
		"web ui dry-run: http://0.0.0.0:7444\n",
		"bind: 0.0.0.0\n",
		"port: 7444\n",
		"open_browser: true\n",
		"require_token: false\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, got)
		}
	}
}

func TestWebServeDryRunRejectsInvalidConfig(t *testing.T) {
	projectDir := newTestWorkspace(t)

	err := run(t.Context(), []string{
		"web",
		"serve",
		"--dir",
		projectDir,
		"--port",
		"70000",
		"--dry-run",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("run returned nil error")
	}

	if err.Error() != "web port must be between 1 and 65535" {
		t.Fatalf("run error = %q, want web port validation", err.Error())
	}
}

func TestWebServeHealthz(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	newWebServeMux().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	if got := recorder.Body.String(); got != "ok\n" {
		t.Fatalf("body = %q, want ok", got)
	}
}

func TestAcceptanceFeatureGenerateOutputGenerated(t *testing.T) {
	var stdout bytes.Buffer

	err := writeAcceptanceFeatureGenerateResult(&stdout, AcceptanceFeatureOperationResult{
		Message: "Acceptance feature generated.",
		Status:  core.AcceptanceFeatureStatusGenerated,
		Feature: ".chc/user/acceptance/blog.feature",
	})
	if err != nil {
		t.Fatalf("writeAcceptanceFeatureGenerateResult returned error: %v", err)
	}

	got := stdout.String()
	for _, want := range []string{
		"Acceptance feature generated.\n",
		"status: generated\n",
		"feature: .chc/user/acceptance/blog.feature\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q in:\n%s", want, got)
		}
	}
}

func TestAcceptanceFeatureGenerateOutputNeedsMoreInfo(t *testing.T) {
	var stdout bytes.Buffer

	err := writeAcceptanceFeatureGenerateResult(&stdout, AcceptanceFeatureOperationResult{
		Message:         "Acceptance feature needs more information.",
		Status:          core.AcceptanceFeatureStatusNeedsMoreInfo,
		RejectionReason: "Missing observable failure path.",
	})
	if err != nil {
		t.Fatalf("writeAcceptanceFeatureGenerateResult returned error: %v", err)
	}

	got := stdout.String()
	for _, want := range []string{
		"status: needs_more_info\n",
		"reason: Missing observable failure path.\n",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q in:\n%s", want, got)
		}
	}
}

func TestAcceptanceFeatureGenerateOutputNotApplicable(t *testing.T) {
	var stdout bytes.Buffer

	err := writeAcceptanceFeatureGenerateResult(&stdout, AcceptanceFeatureOperationResult{
		Message:         "Acceptance feature not applicable.",
		Status:          core.AcceptanceFeatureStatusNotApplicable,
		RejectionReason: "The topic is not observable behavior.",
	})
	if err != nil {
		t.Fatalf("writeAcceptanceFeatureGenerateResult returned error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "status: not_applicable\n") {
		t.Fatalf("output = %q, want not_applicable status", got)
	}

	if !strings.Contains(got, "reason: The topic is not observable behavior.\n") {
		t.Fatalf("output = %q, want reason", got)
	}
}

func TestGateWoodpeckerInitDryRunPrintsConfigWithoutWriting(t *testing.T) {
	projectDir := newGateCLIGitProject(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"gate",
		"woodpecker",
		"init",
		"--dir",
		projectDir,
		"--dry-run",
	}, &stdout, io.Discard)
	if err != nil {
		t.Fatalf("gate woodpecker init returned error: %v", err)
	}

	output := stdout.String()
	for _, expected := range []string{
		"woodpecker: init",
		"action: dry-run",
		"feedback: woodpecker-status",
		"wrote: false",
		"chc-dev-main-gate",
	} {
		if strings.Contains(output, expected) {
			continue
		}

		t.Fatalf("output does not contain %q:\n%s", expected, output)
	}

	_, statErr := os.Stat(filepath.Join(projectDir, ".woodpecker.yml"))
	if !os.IsNotExist(statErr) {
		t.Fatalf("Woodpecker config stat error = %v, want not exist", statErr)
	}
}

func TestGateDevMainWritesReportsWhenValidationPasses(t *testing.T) {
	projectDir := newGateCLIGitProject(t)
	writeGateCLIFile(t, projectDir, "go.mod", "module gatecli.test\n\ngo 1.25\n")
	writeGateCLIFile(t, projectDir, "gatecli_test.go", "package gatecli\n")
	runGateCLIGit(t, projectDir, "add", ".")
	runGateCLIGit(t, projectDir, "commit", "-m", "base")

	baseSHA := gateCLIGitOutput(t, projectDir, "rev-parse", "HEAD")

	writeGateCLIFile(t, projectDir, "README.md", "gate evidence\n")
	runGateCLIGit(t, projectDir, "add", ".")
	runGateCLIGit(t, projectDir, "commit", "-m", "change")

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"gate",
		"dev-main",
		"--dir",
		projectDir,
		"--base",
		baseSHA,
		"--head",
		"HEAD",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("gate dev-main returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "status: passed") {
		t.Fatalf("stdout = %q, want passed status", stdout.String())
	}

	report := readOnlyGateCLIJSONReport(t, projectDir)
	if report.Status != core.GateReportStatusPassed {
		t.Fatalf("Status = %q, want passed", report.Status)
	}

	if len(report.ManualOrExternalChanges) == 0 {
		t.Fatalf("ManualOrExternalChanges = %#v, want manual change context", report.ManualOrExternalChanges)
	}

	if len(report.Failures) != 0 {
		t.Fatalf("Failures = %#v, want no failures for manual changes", report.Failures)
	}

	assertPathExists(t, report.Artifacts.MarkdownPath)
}

func TestGateDevMainWritesReportsWhenValidationFails(t *testing.T) {
	projectDir := newGateCLIGitProject(t)
	writeGateCLIFile(t, projectDir, "go.mod", "module gatecli.test\n\ngo 1.25\n")
	writeGateCLIFile(t, projectDir, "gatecli_test.go", "package gatecli\n")
	runGateCLIGit(t, projectDir, "add", ".")
	runGateCLIGit(t, projectDir, "commit", "-m", "base")

	baseSHA := gateCLIGitOutput(t, projectDir, "rev-parse", "HEAD")

	writeGateCLIFile(t, projectDir, "failing_test.go", `package gatecli

import "testing"

func TestFailure(t *testing.T) {
	t.Fatal("expected failure")
}
`)
	runGateCLIGit(t, projectDir, "add", ".")
	runGateCLIGit(t, projectDir, "commit", "-m", "failing test")

	err := run(t.Context(), []string{
		"gate",
		"dev-main",
		"--dir",
		projectDir,
		"--base",
		baseSHA,
		"--head",
		"HEAD",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("gate dev-main returned nil error")
	}

	if !strings.Contains(err.Error(), "gate validation failed") {
		t.Fatalf("error = %q, want validation failure", err.Error())
	}

	report := readOnlyGateCLIJSONReport(t, projectDir)
	if report.Status != core.GateReportStatusFailed {
		t.Fatalf("Status = %q, want failed", report.Status)
	}

	if report.Validation.Passed {
		t.Fatalf("Validation.Passed = true, want false: %#v", report.Validation)
	}
}

func TestGateDevMainWritesReportsWhenRefResolutionFails(t *testing.T) {
	projectDir := newGateCLIGitProject(t)
	writeGateCLIFile(t, projectDir, "go.mod", "module gatecli.test\n\ngo 1.25\n")
	writeGateCLIFile(t, projectDir, "gatecli_test.go", "package gatecli\n")
	runGateCLIGit(t, projectDir, "add", ".")
	runGateCLIGit(t, projectDir, "commit", "-m", "base")

	err := run(t.Context(), []string{
		"gate",
		"dev-main",
		"--dir",
		projectDir,
		"--base",
		"missing-ref",
		"--head",
		"HEAD",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("gate dev-main returned nil error")
	}

	if !strings.Contains(err.Error(), "resolve base ref") {
		t.Fatalf("error = %q, want base ref context", err.Error())
	}

	report := readOnlyGateCLIJSONReport(t, projectDir)
	if report.Status != core.GateReportStatusFailed {
		t.Fatalf("Status = %q, want failed", report.Status)
	}

	if len(report.Failures) != 1 || !strings.Contains(report.Failures[0], "resolve base ref") {
		t.Fatalf("Failures = %#v, want base ref failure", report.Failures)
	}
}

func TestGateDevMainFailsWhenAgentEvidenceIsMissing(t *testing.T) {
	projectDir := newGateCLIGitProject(t)
	writeGateCLIFile(t, projectDir, "go.mod", "module gatecli.test\n\ngo 1.25\n")
	writeGateCLIFile(t, projectDir, "gatecli_test.go", "package gatecli\n")
	runGateCLIGit(t, projectDir, "add", ".")
	runGateCLIGit(t, projectDir, "commit", "-m", "base")

	baseSHA := gateCLIGitOutput(t, projectDir, "rev-parse", "HEAD")

	writeGateCLIFile(t, projectDir, "README.md", "gate evidence\n")
	writeGateCLIFile(t, projectDir, filepath.Join(core.DirectoryName, core.UserDirectoryName, "runtime", "runs", "CHC-1", "exec-1", "execution-report.md"), `# Execution Report

## Summary

- execution: exec-1
- ticket: CHC-1
- status: succeeded

## Artifacts

- result: .chc/user/runtime/runs/CHC-1/exec-1/result.json
`)
	runGateCLIGit(t, projectDir, "add", ".")
	runGateCLIGit(t, projectDir, "commit", "-m", "change")

	err := run(t.Context(), []string{
		"gate",
		"dev-main",
		"--dir",
		projectDir,
		"--base",
		baseSHA,
		"--head",
		"HEAD",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("gate dev-main returned nil error")
	}

	if !strings.Contains(err.Error(), "missing agent evidence") {
		t.Fatalf("error = %q, want missing evidence", err.Error())
	}

	report := readOnlyGateCLIJSONReport(t, projectDir)
	if report.Status != core.GateReportStatusFailed {
		t.Fatalf("Status = %q, want failed", report.Status)
	}

	if len(report.MissingAgentEvidence) != 1 {
		t.Fatalf("MissingAgentEvidence = %#v, want one entry", report.MissingAgentEvidence)
	}

	if len(report.Failures) != 1 || report.Failures[0] != "missing agent evidence" {
		t.Fatalf("Failures = %#v, want missing agent evidence failure", report.Failures)
	}
}

func TestGateDevMainAllowsMissingAgentEvidenceWhenConfigured(t *testing.T) {
	projectDir := newGateCLIGitProject(t)
	writeGateCLIFile(t, projectDir, "go.mod", "module gatecli.test\n\ngo 1.25\n")
	writeGateCLIFile(t, projectDir, "gatecli_test.go", "package gatecli\n")
	writeGateCLIFile(t, projectDir, ".chc/user/config/gates.yml", `gates:
  dev_main:
    traceability:
      fail_on_missing_agent_evidence: false
`)
	runGateCLIGit(t, projectDir, "add", ".")
	runGateCLIGit(t, projectDir, "commit", "-m", "base")

	baseSHA := gateCLIGitOutput(t, projectDir, "rev-parse", "HEAD")

	writeGateCLIFile(t, projectDir, "README.md", "gate evidence\n")
	writeGateCLIFile(t, projectDir, filepath.Join(core.DirectoryName, core.UserDirectoryName, "runtime", "runs", "CHC-1", "exec-1", "execution-report.md"), `# Execution Report

## Summary

- execution: exec-1
- ticket: CHC-1
- status: succeeded

## Artifacts

- result: .chc/user/runtime/runs/CHC-1/exec-1/result.json
`)
	runGateCLIGit(t, projectDir, "add", ".")
	runGateCLIGit(t, projectDir, "commit", "-m", "change")

	err := run(t.Context(), []string{
		"gate",
		"dev-main",
		"--dir",
		projectDir,
		"--base",
		baseSHA,
		"--head",
		"HEAD",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("gate dev-main returned error: %v", err)
	}

	report := readOnlyGateCLIJSONReport(t, projectDir)
	if report.Status != core.GateReportStatusPassed {
		t.Fatalf("Status = %q, want passed", report.Status)
	}
}

func TestGateDevMainWarnsOnManualChangesWhenConfigured(t *testing.T) {
	projectDir := newGateCLIGitProject(t)
	writeGateCLIFile(t, projectDir, "go.mod", "module gatecli.test\n\ngo 1.25\n")
	writeGateCLIFile(t, projectDir, "gatecli_test.go", "package gatecli\n")
	writeGateCLIFile(t, projectDir, ".chc/user/config/gates.yml", `gates:
  dev_main:
    traceability:
      warn_on_manual_changes: true
    manual_changes:
      list: false
`)
	runGateCLIGit(t, projectDir, "add", ".")
	runGateCLIGit(t, projectDir, "commit", "-m", "base")

	baseSHA := gateCLIGitOutput(t, projectDir, "rev-parse", "HEAD")

	writeGateCLIFile(t, projectDir, "README.md", "gate evidence\n")
	runGateCLIGit(t, projectDir, "add", ".")
	runGateCLIGit(t, projectDir, "commit", "-m", "change")

	err := run(t.Context(), []string{
		"gate",
		"dev-main",
		"--dir",
		projectDir,
		"--base",
		baseSHA,
		"--head",
		"HEAD",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("gate dev-main returned error: %v", err)
	}

	report := readOnlyGateCLIJSONReport(t, projectDir)
	if len(report.Warnings) != 1 {
		t.Fatalf("Warnings = %#v, want manual warning", report.Warnings)
	}

	if len(report.ManualOrExternalChanges) != 0 {
		t.Fatalf("ManualOrExternalChanges = %#v, want hidden manual changes", report.ManualOrExternalChanges)
	}
}

func TestGateDevMainWarnsOnHighRiskManualChangesWhenConfigured(t *testing.T) {
	projectDir := newGateCLIGitProject(t)
	writeGateCLIFile(t, projectDir, "go.mod", "module gatecli.test\n\ngo 1.25\n")
	writeGateCLIFile(t, projectDir, "gatecli_test.go", "package gatecli\n")
	writeGateCLIFile(t, projectDir, ".chc/user/config/gates.yml", `gates:
  dev_main:
    manual_changes:
      warn_on_high_risk_paths: true
    high_risk_paths:
      - internal/runtime/**
`)
	runGateCLIGit(t, projectDir, "add", ".")
	runGateCLIGit(t, projectDir, "commit", "-m", "base")

	baseSHA := gateCLIGitOutput(t, projectDir, "rev-parse", "HEAD")

	writeGateCLIFile(t, projectDir, "internal/runtime/runner.go", "package runtime\n")
	runGateCLIGit(t, projectDir, "add", ".")
	runGateCLIGit(t, projectDir, "commit", "-m", "change")

	err := run(t.Context(), []string{
		"gate",
		"dev-main",
		"--dir",
		projectDir,
		"--base",
		baseSHA,
		"--head",
		"HEAD",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("gate dev-main returned error: %v", err)
	}

	report := readOnlyGateCLIJSONReport(t, projectDir)
	if len(report.HighRiskManualChanges) != 1 {
		t.Fatalf("HighRiskManualChanges = %#v, want one high-risk change", report.HighRiskManualChanges)
	}

	if len(report.Warnings) != 1 {
		t.Fatalf("Warnings = %#v, want high-risk warning", report.Warnings)
	}
}

func rootCommandIncludesPath(cmd *cli.Command, path ...string) bool {
	current := cmd

	for _, name := range path {
		var next *cli.Command

		for _, child := range current.Commands {
			if child.Name == name {
				next = child

				break
			}
		}

		if next == nil {
			return false
		}

		current = next
	}

	return true
}

func TestInitDryRunDoesNotWriteWorkspace(t *testing.T) {
	projectDir := t.TempDir()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"init",
		"--dir",
		projectDir,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("init dry-run returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "create .chc/user/agents") {
		t.Fatalf("stdout = %q, want planned .chc paths", stdout.String())
	}

	assertPathNotExists(t, filepath.Join(projectDir, ".chc"))
}

func TestInitPreservesExistingRootAgents(t *testing.T) {
	projectDir := t.TempDir()
	agentsPath := filepath.Join(projectDir, "AGENTS.md")

	err := os.WriteFile(agentsPath, []byte("custom agents\n"), 0o644)
	if err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{"init", "--dir", projectDir}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("init returned error: %v", err)
	}

	content := readFile(t, agentsPath)
	if content != "custom agents\n" {
		t.Fatalf("AGENTS.md = %q, want preserved custom content", content)
	}

	if !strings.Contains(stdout.String(), "skip existing AGENTS.md") {
		t.Fatalf("stdout = %q, want skipped AGENTS.md", stdout.String())
	}
}

func TestInitStandardProfileInstallsProtocolAssets(t *testing.T) {
	projectDir := t.TempDir()

	err := run(t.Context(), []string{
		"init",
		"--dir",
		projectDir,
		"--profile",
		"standard",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("init standard returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(projectDir, ".chc", "user", "agents", "CONVENTIONS.md"))
	assertPathExists(t, filepath.Join(projectDir, ".chc", "user", "agents", "playbooks", "ticket-workflow.md"))
	assertPathExists(t, filepath.Join(projectDir, "AGENTS.md"))

	documenting := readTestFile(t, filepath.Join(projectDir, ".chc", "user", "agents", "DOCUMENTING.md"))
	for _, unwanted := range []string{
		"ChrysoCoord Artifacts",
		"Use `.chc/` for project-local operational artifacts",
		"generated operational records",
	} {
		if strings.Contains(documenting, unwanted) {
			t.Fatalf("DOCUMENTING.md contains operational artifact rule %q:\n%s", unwanted, documenting)
		}
	}

	architecture := readTestFile(t, filepath.Join(projectDir, ".chc", "user", "agents", "ARCHITECTURE.md"))
	if strings.Contains(architecture, ".chc/         ChrysoCoord project-local state") {
		t.Fatalf("ARCHITECTURE.md exposes .chc as application layout:\n%s", architecture)
	}

	invariants := readTestFile(t, filepath.Join(projectDir, ".chc", "user", "agents", "INVARIANTS.md"))
	if strings.Contains(invariants, "Treat `.chc/` as project-local operational state") {
		t.Fatalf("INVARIANTS.md contains old .chc operational rule:\n%s", invariants)
	}
}

func TestInitRejectsFileTarget(t *testing.T) {
	projectDir := t.TempDir()
	targetPath := filepath.Join(projectDir, "not-a-dir")

	err := os.WriteFile(targetPath, []byte("content\n"), 0o644)
	if err != nil {
		t.Fatalf("write target file: %v", err)
	}

	err = run(t.Context(), []string{"init", "--dir", targetPath}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("init returned nil error")
	}

	if !strings.Contains(err.Error(), "project directory is not a directory") {
		t.Fatalf("init error = %v, want not a directory", err)
	}
}

func TestMetadataProjectSetSavesProjectMetadata(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, core.DirectoryName, ".keep"), "")

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"metadata",
		"project",
		"set",
		"--dir",
		projectDir,
		"--name",
		"CLI Project",
		"--description",
		"Saved through CLI.",
		"--context",
		"CLI context.",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("metadata project set returned error: %v", err)
	}

	metadata, err := core.NewProjectMetadataService(projectDir).Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if metadata.Name != "CLI Project" || metadata.Context != "CLI context." {
		t.Fatalf("metadata = %#v, want CLI project metadata", metadata)
	}

	if !strings.Contains(stdout.String(), "project_metadata:") {
		t.Fatalf("stdout = %q, want project_metadata path", stdout.String())
	}
}

func TestDumpNewAndShowUseServiceCommands(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceWorkspaceQueryFixture(t, projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"dump",
		"new",
		"--dir",
		projectDir,
		"--title",
		"CLI Dump",
		"--content",
		"Dump body from CLI.",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("dump new returned error: %v", err)
	}

	dumpPath := filepath.ToSlash(filepath.Join(core.DirectoryName, core.UserDirectoryName, "dumps", "cli-dump.md"))
	rawDumpPath := filepath.ToSlash(filepath.Join(core.DirectoryName, core.UserDirectoryName, "dumps", "cli-dump-raw.md"))

	if !strings.Contains(stdout.String(), "Created dump "+dumpPath+".") {
		t.Fatalf("dump new stdout = %q, want curated dump path", stdout.String())
	}

	if _, err := os.Stat(filepath.Join(projectDir, filepath.FromSlash(rawDumpPath))); err != nil {
		t.Fatalf("stat raw dump: %v", err)
	}

	stdout.Reset()

	err = run(t.Context(), []string{
		"dump",
		"show",
		"--dir",
		projectDir,
		dumpPath,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("dump show returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "Dump body from CLI.") {
		t.Fatalf("stdout = %q, want dump body", stdout.String())
	}

	rawContent, err := os.ReadFile(filepath.Join(projectDir, filepath.FromSlash(rawDumpPath)))
	if err != nil {
		t.Fatalf("read raw dump: %v", err)
	}

	hasHighlights := strings.Contains(string(rawContent), "## Highlights")
	hasRawBody := strings.Contains(string(rawContent), "Dump body from CLI.")

	if !hasHighlights || !hasRawBody {
		t.Fatalf("raw dump content = %q, want highlights and raw body", string(rawContent))
	}
}

func TestMetadataCaseSetCreatesUseCaseMetadata(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceWorkspaceFile(t, filepath.Join(projectDir, core.DirectoryName, ".keep"), "")

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"metadata",
		"case",
		"set",
		"--dir",
		projectDir,
		"--name",
		"Blog Publishing",
		"--description",
		"Publish articles.",
		"--context",
		"Use-case context.",
		"blog",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("metadata case set returned error: %v", err)
	}

	metadata, err := core.NewCaseMetadataService(projectDir).LoadPath(core.CaseMetadataPath(projectDir, "blog"))
	if err != nil {
		t.Fatalf("LoadPath returned error: %v", err)
	}

	if metadata.Name != "Blog Publishing" || metadata.Context != "Use-case context." {
		t.Fatalf("metadata = %#v, want CLI case metadata", metadata)
	}

	if !strings.Contains(stdout.String(), "case_metadata:") {
		t.Fatalf("stdout = %q, want case_metadata path", stdout.String())
	}
}

func TestSpecAndPlanListCommandsUseServiceViews(t *testing.T) {
	projectDir := t.TempDir()
	writeServiceWorkspaceQueryFixture(t, projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"spec",
		"list",
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("spec list returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), ".chc/user/specs/blog-spec.md") {
		t.Fatalf("spec list stdout = %q, want blog spec path", stdout.String())
	}

	stdout.Reset()

	err = run(t.Context(), []string{
		"plan",
		"list",
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("plan list returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), ".chc/user/plans/project-plan.md") {
		t.Fatalf("plan list stdout = %q, want project plan path", stdout.String())
	}
}

func TestDaemonStatusPrintsStoppedWhenStateIsMissing(t *testing.T) {
	devRoot := t.TempDir()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"daemon",
		"status",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon status returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: stopped",
		"state: missing",
		"data_dir: " + filepath.Join(devRoot, "data", "chc"),
		"state_dir: " + filepath.Join(devRoot, "state", "chc"),
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestDaemonStatusPrintsStoredRunningState(t *testing.T) {
	devRoot := t.TempDir()
	startedAt := time.Now().UTC()
	heartbeatAt := startedAt.Add(time.Minute)

	upsertDaemonState(t, devRoot, core.DaemonState{
		Status:        core.DaemonStatusRunning,
		PID:           1234,
		StartedAt:     &startedAt,
		HeartbeatAt:   &heartbeatAt,
		FailureReason: "",
		UpdatedAt:     heartbeatAt,
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"daemon",
		"status",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon status returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: running",
		"state: stored",
		"pid: 1234",
		"started_at:",
		"heartbeat_at:",
		"updated_at:",
		"data_dir: " + filepath.Join(devRoot, "data", "chc"),
		"state_dir: " + filepath.Join(devRoot, "state", "chc"),
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestDaemonStatusPrintsStoredStoppingState(t *testing.T) {
	devRoot := t.TempDir()
	startedAt := time.Now().UTC()
	stopRequestedAt := startedAt.Add(time.Minute)

	upsertDaemonState(t, devRoot, core.DaemonState{
		Status:          core.DaemonStatusStopping,
		PID:             1234,
		StartedAt:       &startedAt,
		StopRequestedAt: &stopRequestedAt,
		UpdatedAt:       stopRequestedAt,
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"daemon",
		"status",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon status returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: stopping",
		"state: stored",
		"pid: 1234",
		"stop_requested_at:",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestDaemonStatusPrintsStoredStaleState(t *testing.T) {
	devRoot := t.TempDir()
	updatedAt := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	upsertDaemonState(t, devRoot, core.DaemonState{
		Status:        core.DaemonStatusStale,
		PID:           1234,
		FailureReason: "missing process",
		UpdatedAt:     updatedAt,
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"daemon",
		"status",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon status returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: stale",
		"state: stored",
		"pid: 1234",
		"failure_reason: missing process",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestDaemonStatusRecoversStaleRunningState(t *testing.T) {
	devRoot := t.TempDir()
	startedAt := time.Now().UTC().Add(-10 * time.Minute)
	heartbeatAt := startedAt

	upsertDaemonState(t, devRoot, core.DaemonState{
		Status:      core.DaemonStatusRunning,
		PID:         1234,
		StartedAt:   &startedAt,
		HeartbeatAt: &heartbeatAt,
		UpdatedAt:   heartbeatAt,
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"daemon",
		"status",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon status returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: stale",
		"state: stored",
		"pid: 1234",
		"failure_reason: daemon state is stale",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	state := getDaemonState(t, devRoot)
	if state.Status != core.DaemonStatusStale {
		t.Fatalf("daemon state = %#v, want stale", state)
	}
}

func TestDaemonStartDryRunPrintsPlan(t *testing.T) {
	devRoot := t.TempDir()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"daemon",
		"start",
		"--dev-root",
		devRoot,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon start returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: planned",
		"action: start",
		"dry_run: true",
		"allowed: true",
		"stale: false",
		"reason: daemon state is missing",
		"data_dir: " + filepath.Join(devRoot, "data", "chc"),
		"state_dir: " + filepath.Join(devRoot, "state", "chc"),
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestDaemonStartDryRunRejectsFreshRunningState(t *testing.T) {
	devRoot := t.TempDir()
	startedAt := time.Now().UTC()
	heartbeatAt := startedAt

	upsertDaemonState(t, devRoot, core.DaemonState{
		Status:      core.DaemonStatusRunning,
		PID:         1234,
		StartedAt:   &startedAt,
		HeartbeatAt: &heartbeatAt,
		UpdatedAt:   heartbeatAt,
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"daemon",
		"start",
		"--dev-root",
		devRoot,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon start returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: planned",
		"action: start",
		"allowed: false",
		"stale: false",
		"reason: daemon already running with pid 1234",
		"current_status: running",
		"current_pid: 1234",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestAuthorSessionNewListAndShow(t *testing.T) {
	projectDir := newTestWorkspace(t)
	sessionID := createAuthorSessionForTest(t, projectDir)

	assertAuthorSessionListContains(t, projectDir, sessionID)
	assertAuthorSessionShowContains(t, projectDir, sessionID, sessionID)
	assertAuthorSessionShowContains(t, projectDir, "billing-sync", sessionID)
}

func createAuthorSessionForTest(t *testing.T, projectDir string) string {
	t.Helper()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"session",
		"new",
		"--dir",
		projectDir,
		"--target",
		"spec",
		"--topic",
		"Billing sync",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author session new returned error: %v", err)
	}

	return authorSessionIDFromOutput(t, stdout.String())
}

func assertAuthorSessionListContains(t *testing.T, projectDir, sessionID string) {
	t.Helper()

	var listOut bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"session",
		"list",
		"--dir",
		projectDir,
	}, &listOut, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author session list returned error: %v", err)
	}

	if !strings.Contains(listOut.String(), sessionID) {
		t.Fatalf("list output does not contain session id %q:\n%s", sessionID, listOut.String())
	}

	if !strings.Contains(listOut.String(), "id\tstatus\ttarget\tupdated_at\ttopic") {
		t.Fatalf("list output does not contain headers:\n%s", listOut.String())
	}

	if !strings.Contains(listOut.String(), "Z\tBilling sync") {
		t.Fatalf("list output does not contain updated time before topic:\n%s", listOut.String())
	}
}

func assertAuthorSessionShowContains(t *testing.T, projectDir, sessionHandle, expectedSessionID string) {
	t.Helper()

	var showOut bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"session",
		"show",
		"--dir",
		projectDir,
		sessionHandle,
	}, &showOut, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author session show returned error: %v", err)
	}

	for _, want := range []string{
		"session: " + expectedSessionID,
		"mode: markdown_direct",
		"target: spec",
		"topic: Billing sync",
		"updated_at:",
	} {
		if !strings.Contains(showOut.String(), want) {
			t.Fatalf("show output does not contain %q:\n%s", want, showOut.String())
		}
	}
}

func TestAuthorQuizNewDryRunPrintsEditableQuiz(t *testing.T) {
	projectDir := newTestWorkspace(t)
	installFakeAuthoringQuizProvider(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"quiz",
		"new",
		"--dir",
		projectDir,
		"--target",
		"spec",
		"--topic",
		"Billing sync",
		"--question-count",
		"2",
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author quiz new dry-run returned error: %v", err)
	}

	for _, want := range []string{
		"quiz: planned",
		"dry_run: true",
		"mode: context",
		"provider: fake",
		"requested_questions: 2",
		"generated_questions: 3",
		"# Context-Aware Quiz",
		"Answer:",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestAuthorQuizNewWritesSessionAndQuiz(t *testing.T) {
	projectDir := newTestWorkspace(t)
	installFakeAuthoringQuizProvider(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"quiz",
		"new",
		"--dir",
		projectDir,
		"--target",
		"spec",
		"--topic",
		"Billing sync",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author quiz new returned error: %v", err)
	}

	sessionID := authorSessionIDFromOutput(t, stdout.String())
	sessionPath := filepath.Join(projectDir, ".chc", "authoring", "sessions", sessionID, "session.md")
	quizPath := filepath.Join(projectDir, ".chc", "user", "quizzes", sessionID+"-billing-sync.md")

	for _, path := range []string{sessionPath, quizPath} {
		_, err = os.Stat(path)
		if err != nil {
			t.Fatalf("expected path %s: %v", path, err)
		}
	}
}

func TestAuthorPromoteQuizToDraftWritesDraftAndUpdatesSession(t *testing.T) {
	projectDir := newTestWorkspace(t)
	sessionID, quizPath := createAuthorQuizForTest(t, projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"promote",
		"quiz-to-draft",
		"--dir",
		projectDir,
		"--quiz",
		quizPath,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author promote quiz-to-draft returned error: %v", err)
	}

	draftPath := filepath.Join(projectDir, ".chc", "user", "drafts", sessionID+"-billing-sync.md")

	_, err = os.Stat(draftPath)
	if err != nil {
		t.Fatalf("expected draft path %s: %v", draftPath, err)
	}

	var showOut bytes.Buffer

	err = run(t.Context(), []string{
		"author",
		"session",
		"show",
		"--dir",
		projectDir,
		sessionID,
	}, &showOut, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author session show returned error: %v", err)
	}

	for _, want := range []string{
		"outputs: 2",
		"provenance: 2",
	} {
		if !strings.Contains(showOut.String(), want) {
			t.Fatalf("show output does not contain %q:\n%s", want, showOut.String())
		}
	}
}

func TestAuthorPromoteQuizToDraftDryRunPrintsDraft(t *testing.T) {
	projectDir := newTestWorkspace(t)
	sessionID, quizPath := createAuthorQuizForTest(t, projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"promote",
		"quiz-to-draft",
		"--dir",
		projectDir,
		"--session",
		sessionID,
		"--quiz",
		quizPath,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author promote quiz-to-draft dry-run returned error: %v", err)
	}

	for _, want := range []string{
		"promotion: quiz-to-draft",
		"dry_run: true",
		"# Draft",
		"## Curated Answers",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestAuthorAssistedStartResumeAnswerStatusAndFinalize(t *testing.T) {
	projectDir := newTestWorkspace(t)
	sessionID := createAuthorAssistedSessionForTest(t, projectDir)
	question := askAuthorAssistedQuestionForTest(t, projectDir, sessionID)

	answerAuthorAssistedQuestionForTest(t, projectDir, sessionID, question)

	var statusOut bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"assisted",
		"status",
		"--dir",
		projectDir,
		"--session",
		sessionID,
	}, &statusOut, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author assisted status returned error: %v", err)
	}

	for _, want := range []string{"assisted: status", "questions: 1", "answers: 1"} {
		if !strings.Contains(statusOut.String(), want) {
			t.Fatalf("status output missing %q:\n%s", want, statusOut.String())
		}
	}

	var finalizeOut bytes.Buffer

	err = run(t.Context(), []string{
		"author",
		"assisted",
		"finalize",
		"--dir",
		projectDir,
		"--session",
		sessionID,
		"--fake-summary",
		"Billing sync needs a durable spec.",
		"--target-kind",
		"spec",
		"--target-name",
		"billing sync spec",
	}, &finalizeOut, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author assisted finalize returned error: %v", err)
	}

	for _, want := range []string{"assisted: finalized", "curated-summary.md", ".chc/user/specs/"} {
		if !strings.Contains(finalizeOut.String(), want) {
			t.Fatalf("finalize output missing %q:\n%s", want, finalizeOut.String())
		}
	}
}

func TestDaemonStartDryRunAllowsStaleRunningState(t *testing.T) {
	devRoot := t.TempDir()
	startedAt := time.Now().UTC().Add(-10 * time.Minute)
	heartbeatAt := startedAt

	upsertDaemonState(t, devRoot, core.DaemonState{
		Status:      core.DaemonStatusRunning,
		PID:         1234,
		StartedAt:   &startedAt,
		HeartbeatAt: &heartbeatAt,
		UpdatedAt:   heartbeatAt,
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"daemon",
		"start",
		"--dev-root",
		devRoot,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon start returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: planned",
		"action: start",
		"allowed: true",
		"stale: true",
		"reason: daemon state is stale",
		"current_status: running",
		"current_pid: 1234",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestDaemonStopDryRunPrintsPlan(t *testing.T) {
	devRoot := t.TempDir()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"daemon",
		"stop",
		"--dev-root",
		devRoot,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon stop returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: planned",
		"action: stop",
		"dry_run: true",
		"data_dir: " + filepath.Join(devRoot, "data", "chc"),
		"state_dir: " + filepath.Join(devRoot, "state", "chc"),
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestDaemonStartRunsForegroundLoop(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	err := os.MkdirAll(filepath.Join(projectDir, ".chc", "user", "tickets"), 0o755)
	if err != nil {
		t.Fatalf("create tickets directory: %v", err)
	}

	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"daemon",
		"start",
		"--dev-root",
		devRoot,
		"--max-iterations",
		"1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon start returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: stopped",
		"iterations: 1",
		"stop_requested: false",
		"projects_loaded: 1",
		"projects_ready: 1",
		"projects_skipped: 0",
		"projects_refreshed: 1",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	state := getDaemonState(t, devRoot)
	if state.Status != core.DaemonStatusStopped {
		t.Fatalf("daemon state = %#v, want stopped", state)
	}
}

func TestDaemonStartRejectsInvalidMaxIterations(t *testing.T) {
	err := run(t.Context(), []string{
		"daemon",
		"start",
		"--dev-root",
		t.TempDir(),
		"--max-iterations",
		"0",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("daemon start returned nil error")
	}

	if !strings.Contains(err.Error(), "max-iterations must be greater than zero") {
		t.Fatalf("daemon start error = %v, want max-iterations validation", err)
	}
}

func TestDaemonStartRejectsFreshRunningState(t *testing.T) {
	devRoot := t.TempDir()
	startedAt := time.Now().UTC()
	heartbeatAt := startedAt

	upsertDaemonState(t, devRoot, core.DaemonState{
		Status:      core.DaemonStatusRunning,
		PID:         1234,
		StartedAt:   &startedAt,
		HeartbeatAt: &heartbeatAt,
		UpdatedAt:   heartbeatAt,
	})

	err := run(t.Context(), []string{
		"daemon",
		"start",
		"--dev-root",
		devRoot,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("daemon start returned nil error")
	}

	if !strings.Contains(err.Error(), "daemon start blocked: daemon already running with pid 1234") {
		t.Fatalf("daemon start error = %v, want blocked running daemon", err)
	}
}

func TestDaemonStopRequestsRunningDaemonStop(t *testing.T) {
	devRoot := t.TempDir()
	startedAt := time.Now().UTC()
	heartbeatAt := startedAt

	upsertDaemonState(t, devRoot, core.DaemonState{
		Status:      core.DaemonStatusRunning,
		PID:         1234,
		StartedAt:   &startedAt,
		HeartbeatAt: &heartbeatAt,
		UpdatedAt:   heartbeatAt,
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{"daemon", "stop", "--dev-root", devRoot}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon stop returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: stopping",
		"action: stop",
		"stop_requested_at:",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	state := getDaemonState(t, devRoot)
	if state.Status != core.DaemonStatusStopping || state.StopRequestedAt == nil {
		t.Fatalf("daemon state = %#v, want stopping with stop request", state)
	}
}

func TestDaemonStopReportsStoppedWhenStateIsMissing(t *testing.T) {
	devRoot := t.TempDir()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{"daemon", "stop", "--dev-root", devRoot}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("daemon stop returned error: %v", err)
	}

	for _, want := range []string{
		"daemon: stopped",
		"action: stop",
		"state: missing",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestDaemonStopRequestedHelperReadsStopRequest(t *testing.T) {
	devRoot := t.TempDir()
	startedAt := time.Now().UTC()
	stopRequestedAt := startedAt.Add(time.Minute)

	upsertDaemonState(t, devRoot, core.DaemonState{
		Status:          core.DaemonStatusStopping,
		PID:             1234,
		StartedAt:       &startedAt,
		StopRequestedAt: &stopRequestedAt,
		UpdatedAt:       stopRequestedAt,
	})

	repository := newDaemonStateRepositoryForTest(t, devRoot)

	stopRequested, err := daemonStopRequested(t.Context(), repository)
	if err != nil {
		t.Fatalf("daemonStopRequested returned error: %v", err)
	}

	if !stopRequested {
		t.Fatal("daemonStopRequested = false, want true")
	}
}

func TestTicketNewCreatesTicket(t *testing.T) {
	dir := newTestWorkspace(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"new",
		"--dir",
		dir,
		"--priority",
		"high",
		"--tag",
		"uxui",
		"--synopsis",
		"Do the implementation work.",
		"Do something",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	wantPath := filepath.Join(dir, ".chc", "user", "tickets", "CHC-1-Do-something.md")
	if !strings.Contains(stdout.String(), "ticket: "+wantPath) {
		t.Fatalf("stdout = %q, want ticket path %q", stdout.String(), wantPath)
	}

	content, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	for _, want := range []string{
		`key: "CHC-1"`,
		`title: "Do something"`,
		`synopsis: "Do the implementation work."`,
		`status: "backlog"`,
		`priority: "high"`,
		`tags: ["uxui"]`,
	} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("content does not contain %q:\n%s", want, string(content))
		}
	}
}

func TestTicketNewUsesConfigKeyPrefix(t *testing.T) {
	dir := t.TempDir()

	configPath := filepath.Join(dir, ".chc", "user", "config.yml")

	err := os.MkdirAll(filepath.Dir(configPath), 0o755)
	if err != nil {
		t.Fatalf("create config dir: %v", err)
	}

	err = os.WriteFile(configPath, []byte("tickets:\n  key_prefix: PRJ\n"), 0o644)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	err = run(t.Context(), []string{
		"ticket",
		"new",
		"--dir",
		dir,
		"--synopsis",
		"Do the implementation work.",
		"Do something",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(dir, ".chc", "user", "tickets", "PRJ-1-Do-something.md"))
}

func TestTicketNewUsesActiveProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		devRoot,
		"example",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project select returned error: %v", err)
	}

	err = run(t.Context(), []string{
		"ticket",
		"new",
		"--dev-root",
		devRoot,
		"--synopsis",
		"Create the active project ticket.",
		"Active project ticket",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket new returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(projectDir, ".chc", "user", "tickets", "CHC-1-Active-project-ticket.md"))
}

func TestRuntimeInitCreatesTemplateWithExplicitDir(t *testing.T) {
	projectDir := newTestWorkspace(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"runtime",
		"init",
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("runtime init returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(projectDir, ".chc", "user", "runtime", "agent", "Dockerfile"))

	if !strings.Contains(stdout.String(), "create .chc/user/runtime/agent/Dockerfile") {
		t.Fatalf("stdout = %q, want runtime Dockerfile create", stdout.String())
	}
}

func TestRuntimeInitDryRunDoesNotCreateTemplate(t *testing.T) {
	projectDir := newTestWorkspace(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"runtime",
		"init",
		"--dir",
		projectDir,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("runtime init returned error: %v", err)
	}

	assertPathNotExists(t, filepath.Join(projectDir, ".chc", "user", "runtime", "agent", "Dockerfile"))

	if !strings.Contains(stdout.String(), "create .chc/user/runtime/agent/Dockerfile") {
		t.Fatalf("stdout = %q, want runtime Dockerfile create", stdout.String())
	}
}

func TestRuntimeInitSkipsExistingTemplate(t *testing.T) {
	projectDir := newTestWorkspace(t)

	dockerfilePath := filepath.Join(projectDir, ".chc", "user", "runtime", "agent", "Dockerfile")

	err := os.MkdirAll(filepath.Dir(dockerfilePath), 0o755)
	if err != nil {
		t.Fatalf("create runtime template directory: %v", err)
	}

	err = os.WriteFile(dockerfilePath, []byte("FROM custom\n"), 0o644)
	if err != nil {
		t.Fatalf("write runtime template: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"runtime",
		"init",
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("runtime init returned error: %v", err)
	}

	content, err := os.ReadFile(dockerfilePath)
	if err != nil {
		t.Fatalf("read runtime template: %v", err)
	}

	if string(content) != "FROM custom\n" {
		t.Fatalf("runtime template content = %q, want custom content unchanged", string(content))
	}

	if !strings.Contains(stdout.String(), "skip existing .chc/user/runtime/agent/Dockerfile") {
		t.Fatalf("stdout = %q, want runtime Dockerfile skip", stdout.String())
	}
}

func TestRuntimeInitUsesActiveProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	selectProject(t, devRoot, "example")

	err := run(t.Context(), []string{
		"runtime",
		"init",
		"--dev-root",
		devRoot,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("runtime init returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(projectDir, ".chc", "user", "runtime", "agent", "Dockerfile"))
}

func TestRuntimeInitFallsBackToCurrentDirectory(t *testing.T) {
	projectDir := newTestWorkspace(t)

	runInDirectory(t, projectDir, func() {
		err := run(t.Context(), []string{
			"runtime",
			"init",
			"--dev-root",
			t.TempDir(),
		}, &bytes.Buffer{}, &bytes.Buffer{})
		if err != nil {
			t.Fatalf("runtime init returned error: %v", err)
		}
	})

	assertPathExists(t, filepath.Join(projectDir, ".chc", "user", "runtime", "agent", "Dockerfile"))
}

func TestRuntimeValidatePrintsRuntimePlan(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeRuntimeTemplate(t, projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"runtime",
		"validate",
		"--dir",
		projectDir,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("runtime validate returned error: %v", err)
	}

	for _, want := range []string{
		"dry_run: true",
		"workspace_mode: project",
		"project_checkout_dir: .",
		"execution_workspace_dir: .",
		"image: chc-agent-runtime:local",
		"dockerfile: .chc/user/runtime/agent/Dockerfile",
		"workdir: /workspace",
		"artifact_dir: .chc/user/runtime/runs",
		"mount: . -> /workspace read_only=false",
		"checks:",
		"ok    execution_workspace .",
		"ok    workspace    .chc",
		"ok    dockerfile   .chc/user/runtime/agent/Dockerfile",
		"ok    artifact_dir .chc/user/runtime/runs",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestRuntimeValidateFailsWhenRuntimeTemplateIsMissing(t *testing.T) {
	projectDir := newTestWorkspace(t)

	err := os.MkdirAll(filepath.Join(projectDir, ".chc", "user", "runtime", "runs"), 0o755)
	if err != nil {
		t.Fatalf("create runtime runs dir: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"runtime",
		"validate",
		"--dir",
		projectDir,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("runtime validate returned nil error")
	}

	if !strings.Contains(stdout.String(), "fail  dockerfile") {
		t.Fatalf("stdout = %q, want failed Dockerfile check", stdout.String())
	}
}

func TestRuntimeValidateUsesActiveProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeRuntimeTemplate(t, projectDir)

	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	selectProject(t, devRoot, "example")

	err := run(t.Context(), []string{
		"runtime",
		"validate",
		"--dev-root",
		devRoot,
		"--dry-run",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("runtime validate returned error: %v", err)
	}
}

func TestNotifyStatusPrintsNotificationReadiness(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeCLIWorkspaceConfig(t, projectDir, `notifications:
  telegram_enabled: true
  telegram_bot_token_env: CHC_TELEGRAM_BOT_TOKEN
  telegram_chat_id: "123456"
`)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"notify",
		"status",
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("notify status returned error: %v", err)
	}

	for _, want := range []string{
		"notify: status",
		"project_id:",
		"project_dir: " + projectDir,
		"telegram_enabled: true",
		"telegram_status: ready",
		"telegram_bot_token_env: CHC_TELEGRAM_BOT_TOKEN",
		"telegram_chat_id: 123456",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestNotifyStatusReportsInvalidTelegramConfig(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeCLIWorkspaceConfig(t, projectDir, `notifications:
  telegram_enabled: true
  telegram_chat_id: "123456"
`)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"notify",
		"status",
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("notify status returned error: %v", err)
	}

	for _, want := range []string{
		"telegram_enabled: true",
		"telegram_status: invalid",
		"telegram_error: telegram bot token environment variable is required",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestNotifyTelegramTestDryRunPrintsPlan(t *testing.T) {
	projectDir := newTestWorkspace(t)
	writeCLIWorkspaceConfig(t, projectDir, `notifications:
  telegram_enabled: true
  telegram_bot_token_env: CHC_TELEGRAM_BOT_TOKEN
  telegram_chat_id: "123456"
`)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"notify",
		"telegram",
		"test",
		"--dir",
		projectDir,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("notify telegram test dry-run returned error: %v", err)
	}

	for _, want := range []string{
		"telegram_test: planned",
		"dry_run: true",
		"bot_token_env: CHC_TELEGRAM_BOT_TOKEN",
		"chat_id: 123456",
		"message: ChrysoCoord Telegram notification test",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestNotifyTelegramTestRequiresDryRun(t *testing.T) {
	projectDir := newTestWorkspace(t)

	err := run(t.Context(), []string{
		"notify",
		"telegram",
		"test",
		"--dir",
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("notify telegram test returned nil error")
	}

	if err.Error() != "telegram bot token is required" {
		t.Fatalf("notify telegram test error = %v", err)
	}
}

func TestNotifyTelegramTestSendsAndRecordsAudit(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	t.Setenv("CHC_TELEGRAM_BOT_TOKEN", "token-1")

	writeTelegramNotifyConfig(t, projectDir)

	gotPayload := installTestTelegramSender(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"notify",
		"telegram",
		"test",
		"--dir",
		projectDir,
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("notify telegram test returned error: %v", err)
	}

	assertTelegramNotifyOutput(t, stdout.String())
	assertTelegramPayload(t, gotPayload)

	records := listNotificationEventsForTest(t, devRoot, projectID(projectDir))
	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}

	if records[0].Event.Status != core.NotificationStatusDelivered {
		t.Fatalf("notification status = %q, want delivered", records[0].Event.Status)
	}
}

func TestNotifyTelegramPollDryRunParsesFixture(t *testing.T) {
	devRoot := t.TempDir()
	fixturePath := writeTelegramUpdateFixture(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"notify",
		"telegram",
		"poll",
		"--fixture",
		fixturePath,
		"--dry-run",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("notify telegram poll dry-run returned error: %v", err)
	}

	for _, want := range []string{"telegram_poll: parsed", "dry_run: true", "records: 2", "accepted", "rejected"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	records := listRemoteCommandRecordsForTest(t, devRoot)
	if len(records) != 0 {
		t.Fatalf("remote command records = %d, want dry-run persistence skipped", len(records))
	}
}

func TestNotifyTelegramPollPersistsFixtureRecords(t *testing.T) {
	devRoot := t.TempDir()
	fixturePath := writeTelegramUpdateFixture(t)

	err := run(t.Context(), []string{
		"notify",
		"telegram",
		"poll",
		"--fixture",
		fixturePath,
		"--dev-root",
		devRoot,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("notify telegram poll returned error: %v", err)
	}

	records := listRemoteCommandRecordsForTest(t, devRoot)
	if len(records) != 2 {
		t.Fatalf("remote command records = %d, want 2", len(records))
	}

	if records[0].Status != core.RemoteCommandRecordAccepted || records[1].Status != core.RemoteCommandRecordRejected {
		t.Fatalf("records = %#v, want accepted then rejected", records)
	}
}

type testTelegramPayload struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func writeTelegramNotifyConfig(t *testing.T, projectDir string) {
	t.Helper()

	writeCLIWorkspaceConfig(t, projectDir, `notifications:
  telegram_enabled: true
  telegram_bot_token_env: CHC_TELEGRAM_BOT_TOKEN
  telegram_chat_id: "123456"
`)
}

func writeTelegramUpdateFixture(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "updates.json")
	content := `{"ok":true,"result":[
{"update_id":1,"message":{"message_id":10,"from":{"id":100,"username":"ops"},"chat":{"id":200},"date":1768000000,"text":"@example status"}},
{"update_id":2,"message":{"message_id":11,"from":{"id":101,"username":"ops"},"chat":{"id":201},"date":1768000001,"text":"merge dev"}}
]}`

	err := os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write telegram update fixture: %v", err)
	}

	return path
}

func installTestTelegramSender(t *testing.T) *testTelegramPayload {
	t.Helper()

	gotPayload := &testTelegramPayload{}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		err := json.NewDecoder(request.Body).Decode(gotPayload)
		if err != nil {
			t.Fatalf("decode telegram payload: %v", err)
		}

		_, err = writer.Write([]byte(`{"ok":true}`))
		if err != nil {
			t.Fatalf("write telegram response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	previousSender := notifyTelegramSender
	notifyTelegramSender = core.TelegramSender{
		Client:  server.Client(),
		BaseURL: server.URL,
	}

	t.Cleanup(func() {
		notifyTelegramSender = previousSender
	})

	return gotPayload
}

func assertTelegramNotifyOutput(t *testing.T, output string) {
	t.Helper()

	for _, want := range []string{
		"telegram_test: sent",
		"dry_run: false",
		"notification_id: telegram-test-",
		"status_code: 200",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, output)
		}
	}
}

func assertTelegramPayload(t *testing.T, gotPayload *testTelegramPayload) {
	t.Helper()

	if gotPayload.ChatID != "123456" {
		t.Fatalf("telegram chat_id = %q, want 123456", gotPayload.ChatID)
	}

	if !strings.Contains(gotPayload.Text, "ChrysoCoord Telegram notification test") {
		t.Fatalf("telegram text = %q, want test message", gotPayload.Text)
	}
}

func TestTicketNewFallsBackToCurrentDirectory(t *testing.T) {
	projectDir := newTestWorkspace(t)

	runInDirectory(t, projectDir, func() {
		err := run(t.Context(), []string{
			"ticket",
			"new",
			"--dev-root",
			t.TempDir(),
			"--synopsis",
			"Create the current project ticket.",
			"Current project ticket",
		}, &bytes.Buffer{}, &bytes.Buffer{})
		if err != nil {
			t.Fatalf("ticket new returned error: %v", err)
		}
	})

	assertPathExists(t, filepath.Join(projectDir, ".chc", "user", "tickets", "CHC-1-Current-project-ticket.md"))
}

func TestProjectContextResolverRejectsMissingContext(t *testing.T) {
	env, err := chcTestUserEnvironment(t)
	if err != nil {
		t.Fatalf("chcTestUserEnvironment returned error: %v", err)
	}

	_, err = newCommandProjectContextResolver(env, testAssetsFS()).Resolve(t.Context(), commandProjectContextOptions{})
	if err == nil {
		t.Fatal("Resolve returned nil error")
	}
}

func TestTicketNewRejectsMissingTitle(t *testing.T) {
	err := run(t.Context(), []string{"ticket", "new"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("run returned nil error")
	}
}

func TestTicketNewRejectsMissingSynopsis(t *testing.T) {
	dir := newTestWorkspace(t)

	err := run(t.Context(), []string{"ticket", "new", "--dir", dir, "Do something"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("run returned nil error")
	}
}

func TestTicketListPrintsIndexedTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicketShowReviewTicket())

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"list",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket list returned error: %v", err)
	}

	want := "CHC-1\tDo something\topen\thigh\tnormal\ttask\tonboarding\tCHC-1-Do-something.md"
	if !strings.Contains(stdout.String(), want) {
		t.Fatalf("stdout = %q, want ticket row %q", stdout.String(), want)
	}
}

func TestTicketBoardPrintsBoardColumns(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Backlog.md", core.Ticket{
		Key:       "CHC-1",
		Title:     "Backlog work",
		Synopsis:  "Prepare backlog work.",
		Status:    core.TicketStatusBacklog,
		Priority:  "normal",
		Severity:  "normal",
		Type:      "task",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})
	writeTicketMarkdownOnly(t, projectDir, "CHC-2-Review.md", core.Ticket{
		Key:       "CHC-2",
		Title:     "Review work",
		Synopsis:  "Review implemented work.",
		Status:    core.TicketStatusReadyForReview,
		Priority:  "high",
		Severity:  "normal",
		Type:      "feature",
		CreatedAt: time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC),
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"board",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket board returned error: %v", err)
	}

	for _, want := range []string{
		"total: 2",
		"normal: 2",
		"attention: 0",
		"column: backlog\tBacklog\t1",
		"CHC-1\tbacklog\tnormal\ttask\tBacklog work",
		"column: ready_for_review\tReady For Review\t1",
		"CHC-2\tready_for_review\thigh\tfeature\tReview work",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestTicketLoopDryRunPrintsEligibleTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	writeTicketLoopDryRunTickets(t, projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"loop",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--dry-run",
		"--priority",
		"high",
		"--project-area",
		"frontend",
		"--tag",
		"uxui",
		"--order",
		"priority",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket loop returned error: %v", err)
	}

	for _, want := range []string{
		"loop: planned",
		"dry_run: true",
		"order: priority",
		"eligible: 1",
		"ticket: CHC-1\tReady frontend\tready_for_dev\thigh\tcritical\tfeature\tonboarding\tCHC-1-Ready.md",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	if strings.Contains(stdout.String(), "CHC-2") || strings.Contains(stdout.String(), "CHC-3") {
		t.Fatalf("stdout contains ineligible tickets:\n%s", stdout.String())
	}
}

func writeTicketLoopDryRunTickets(t *testing.T, projectDir string) {
	t.Helper()

	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", testLoopTicket(
		"CHC-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))
	writeTicketMarkdownOnly(t, projectDir, "CHC-2-Ready.md", testLoopTicket(
		"CHC-2", "Ready backend", "ready_for_dev", "normal", "normal", "bug", "maintenance", "backend", []string{"api"},
	))
	writeTicketMarkdownOnly(t, projectDir, "CHC-3-Open.md", testLoopTicket(
		"CHC-3", "Open frontend", "open", "critical", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))
}

func testLoopTicket(
	key, title, status, priority, severity, ticketType, useCase, projectArea string,
	tags []string,
) core.Ticket {
	return core.Ticket{
		Key:         key,
		Title:       title,
		Status:      status,
		Priority:    priority,
		Severity:    severity,
		Type:        ticketType,
		UseCase:     useCase,
		ProjectArea: projectArea,
		Tags:        tags,
		Body:        testLoopTicketBody(),
		CreatedAt:   time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	}
}

func testLoopTicketBody() string {
	rendered, err := testCompleteEvidenceChecklistBlock()
	if err != nil {
		panic(err)
	}

	return rendered + `
# Description

Loop behavior.

# Checklist

# Acceptance

# Comments

# Blocker

# Validation
	`
}

func testReviewRepairTicket() core.Ticket {
	ticket := testLoopTicket(
		"CHC-1", "Review repair", core.TicketStatusReadyForDev, "high", "critical", "bug", "review", "backend", []string{"repair"},
	)
	ticket.Branch = "feat/chc-1-review-repair"
	ticket.ReviewState = "changes_requested"
	ticket.ReviewResult = core.TicketReviewCommandResultNeedsChanges
	ticket.ReviewRound = 1
	ticket.ReviewCommentsURL = "https://forge.example.com/adrianpk/chc/pulls/274"
	ticket.ReviewRefs = []core.TicketReviewRef{{
		Type:     "pull_request",
		Provider: "forgejo",
		URL:      "https://forge.example.com/adrianpk/chc/pulls/274",
		ID:       "274",
		Status:   "open",
	}}

	return ticket
}

func withLoopTicketCreatedAt(ticket core.Ticket, createdAt time.Time) core.Ticket {
	ticket.CreatedAt = createdAt
	ticket.UpdatedAt = createdAt

	return ticket
}

func TestTicketLoopExecutesEligibleTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	executor := installFakeTicketStartHarnessExecutor(t, nil)

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", testLoopTicket(
		"CHC-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"loop",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket loop returned error: %v", err)
	}

	for _, want := range []string{
		"loop: loop-",
		"status: succeeded",
		"eligible: 1",
		"processed: 1",
		"skipped: 0",
		"blocked: 0",
		"failed: 0",
		"execution: exec-",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	if !strings.Contains(strings.Join(executor.commands, ","), "docker build,docker run,docker run") {
		t.Fatalf("commands = %#v, want build, worker run, and validation run", executor.commands)
	}

	records := listTicketExecutions(t, devRoot, projectID(projectDir), "CHC-1")
	if len(records) != 1 {
		t.Fatalf("execution records = %d, want 1", len(records))
	}

	loopRuns := listProjectLoopRuns(t, devRoot, projectID(projectDir))
	if len(loopRuns) != 1 {
		t.Fatalf("loop runs = %d, want 1", len(loopRuns))
	}

	if loopRuns[0].Status != core.LoopRunStatusSucceeded || loopRuns[0].ProcessedCount != 1 {
		t.Fatalf("loop run = %#v, want succeeded with one processed ticket", loopRuns[0])
	}
}

func TestTicketLoopHonorsMaxIterations(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, nil)

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", withLoopTicketCreatedAt(testLoopTicket(
		"CHC-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	), time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)))
	writeTicketMarkdownOnly(t, projectDir, "CHC-2-Ready.md", withLoopTicketCreatedAt(testLoopTicket(
		"CHC-2", "Ready backend", "ready_for_dev", "normal", "normal", "bug", "maintenance", "backend", []string{"api"},
	), time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC)))

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"loop",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--max-iterations",
		"1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket loop returned error: %v", err)
	}

	for _, want := range []string{
		"status: succeeded",
		"eligible: 2",
		"processed: 1",
		"skipped: 1",
		"execution: exec-",
		"CHC-1",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	if strings.Contains(stdout.String(), "CHC-2") {
		t.Fatalf("stdout contains execution beyond max iterations:\n%s", stdout.String())
	}

	records := listTicketExecutions(t, devRoot, projectID(projectDir), "CHC-2")
	if len(records) != 0 {
		t.Fatalf("CHC-2 execution records = %d, want 0", len(records))
	}
}

func TestTicketLoopCompletesWithNoEligibleTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	executor := installFakeTicketStartHarnessExecutor(t, nil)

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Open.md", testLoopTicket(
		"CHC-1", "Open frontend", "open", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"loop",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket loop returned error: %v", err)
	}

	for _, want := range []string{
		"status: succeeded",
		"eligible: 0",
		"processed: 0",
		"skipped: 0",
		"blocked: 0",
		"failed: 0",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	if len(executor.commands) != 0 {
		t.Fatalf("commands = %#v, want no ticket execution", executor.commands)
	}
}

func TestWorkTicketRunsSelectedTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, nil)

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", testLoopTicket(
		"CHC-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"work",
		"ticket",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("work ticket returned error: %v", err)
	}

	for _, want := range []string{
		"work: ticket",
		"ticket: CHC-1",
		"execution: exec-",
		"status: succeeded",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	records := listTicketExecutions(t, devRoot, projectID(projectDir), "CHC-1")
	if len(records) != 1 {
		t.Fatalf("execution records = %d, want 1", len(records))
	}
}

func TestWorkDoctorReportsDefaultGitReadiness(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"work",
		"doctor",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("work doctor returned error: %v", err)
	}

	for _, want := range []string{
		"work: doctor",
		"status: blocked",
		"mode: integration_branch",
		"integration_branch: dev",
		"canonical_remote: origin",
		"diagnostic: git_unavailable\tblocking\tgit state is unavailable:",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestWorkDoctorHonorsNoGitEnforcementConfig(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	writeCLIWorkspaceConfig(t, projectDir, `work:
  mode: no_git_enforcement
`)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"work",
		"doctor",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("work doctor returned error: %v", err)
	}

	for _, want := range []string{
		"work: doctor",
		"status: ready",
		"mode: no_git_enforcement",
		"diagnostics: 0",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestWorkDoctorReportsWorktreeWorkspaceMode(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	writeCLIWorkspaceConfig(t, projectDir, `work:
  mode: no_git_enforcement
  workspace_mode: worktree
`)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"work",
		"doctor",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("work doctor returned error: %v", err)
	}

	for _, want := range []string{
		"work: doctor",
		"status: ready",
		"mode: no_git_enforcement",
		"workspace_mode: worktree",
		"user_checkout_state: unknown",
		"ticket_worktree_state: unknown",
		"diagnostics: 0",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestWorkDoctorFixSafeIgnoresRuntimeArtifacts(t *testing.T) {
	projectDir := newGateCLIGitProject(t)
	devRoot := t.TempDir()

	runGateCLIGit(t, projectDir, "switch", "-c", "dev")
	runGateCLIGit(t, projectDir, "add", ".chc/.keep")
	runGateCLIGit(t, projectDir, "commit", "-m", "base")
	writeCLIWorkspaceConfig(t, projectDir, `work:
  require_remote_tracking: false
`)
	runGateCLIGit(t, projectDir, "add", ".chc/user/config.yml")
	runGateCLIGit(t, projectDir, "commit", "-m", "work config")
	writeGateCLIFile(t, projectDir, ".chc/user/runtime/runs/run-1/events.jsonl", "{}\n")
	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"work",
		"doctor",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--fix-safe",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("work doctor --fix-safe returned error: %v", err)
	}

	for _, want := range []string{
		"work: doctor",
		"fix_safe: true",
		"repairs_applied: 1",
		"repair_applied: ignore_runtime_artifacts\texcluded .chc/user/runtime/ in local Git info exclude",
		"status: ready",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	exclude := readFile(t, filepath.Join(projectDir, ".git", "info", "exclude"))
	if !strings.Contains(exclude, ".chc/user/runtime/") {
		t.Fatalf("exclude = %q, want runtime ignore", exclude)
	}
}

func TestWorkLogsPrintsExecutionLogPaths(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	runDir := filepath.Join(projectDir, ".chc", "user", "runtime", "runs", "CHC-1", "exec-1")
	record := cliExecutionRecord(projectDir, runDir)

	addProject(t, devRoot, "example", projectDir)
	createCLIExecutionRecord(t, devRoot, record)
	writeCLIExecutionArtifacts(t, record)

	err := os.WriteFile(core.AgentEventStreamPath(record.Artifacts.RunDir), []byte("{}\n"), 0o644)
	if err != nil {
		t.Fatalf("write events stream: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"work",
		"logs",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"exec-1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("work logs returned error: %v", err)
	}

	for _, want := range []string{
		"work: logs",
		"execution: exec-1",
		"ticket: CHC-1",
		"status: succeeded",
		"validation_command: make verify",
		"run_dir_path: .chc/user/runtime/runs/CHC-1/exec-1",
		"run_dir_exists: true",
		"runtime_log_path: .chc/user/runtime/runs/CHC-1/exec-1/runtime.log",
		"runtime_log_exists: true",
		"validation_log_path: .chc/user/runtime/runs/CHC-1/exec-1/validation.log",
		"validation_log_exists: true",
		"execution_report_path: .chc/user/runtime/runs/CHC-1/exec-1/execution-report.md",
		"execution_report_exists: true",
		"events_path: .chc/user/runtime/runs/CHC-1/exec-1/events.jsonl",
		"events_exists: true",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestWorkLogsReportsMissingArtifacts(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	runDir := filepath.Join(projectDir, ".chc", "user", "runtime", "runs", "CHC-1", "exec-1")
	record := cliExecutionRecord(projectDir, runDir)

	addProject(t, devRoot, "example", projectDir)
	createCLIExecutionRecord(t, devRoot, record)

	err := os.MkdirAll(record.Artifacts.RunDir, 0o755)
	if err != nil {
		t.Fatalf("create run dir: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"work",
		"logs",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"exec-1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("work logs returned error: %v", err)
	}

	for _, want := range []string{
		"run_dir_exists: true",
		"runtime_log_exists: false",
		"validation_log_exists: false",
		"result_exists: false",
		"execution_report_exists: false",
		"events_exists: false",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestWorkLogsRejectsMissingExecution(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"work",
		"logs",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"exec-missing",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("work logs returned nil error")
	}

	if !strings.Contains(err.Error(), "execution not found: exec-missing") {
		t.Fatalf("error = %v, want missing execution", err)
	}
}

func TestWorkEventsSummarizesDegradedEventStream(t *testing.T) {
	projectDir := newTestWorkspace(t)
	streamPath := filepath.Join(projectDir, ".chc", "user", "runtime", "runs", "loop-1", "events.jsonl")
	event := core.AgentEvent{
		SchemaVersion: core.AgentEventSchemaVersion,
		Timestamp:     time.Date(2026, 6, 18, 12, 30, 0, 0, time.UTC),
		ProjectID:     projectID(projectDir),
		LoopID:        "loop-1",
		TicketKey:     "CHC-1",
		ExecutionID:   "exec-1",
		Phase:         core.AgentEventPhaseWorkBlocked,
		Level:         core.AgentEventLevelWarn,
		Message:       "CHC-1 blocked",
	}

	writeCLIAgentEventStream(t, streamPath, marshalCLIAgentEventLine(t, event), "{not-json")

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"work",
		"events",
		"--dir",
		projectDir,
		".chc/user/runtime/runs/loop-1/events.jsonl",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("work events returned error: %v", err)
	}

	for _, want := range []string{
		"work: events",
		"path: .chc/user/runtime/runs/loop-1/events.jsonl",
		"events: 1",
		"malformed: 1",
		"degraded: true",
		"status: blocked",
		"latest_time: 2026-06-18T12:30:00Z",
		"latest_loop: loop-1",
		"latest_ticket: CHC-1",
		"latest_execution: exec-1",
		"latest_phase: work_blocked",
		"latest_level: warn",
		"latest_message: CHC-1 blocked",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestWorkEventsSummarizesMissingEventStream(t *testing.T) {
	projectDir := newTestWorkspace(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"work",
		"events",
		"--dir",
		projectDir,
		".chc/user/runtime/runs/missing/events.jsonl",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("work events returned error: %v", err)
	}

	for _, want := range []string{
		"path: .chc/user/runtime/runs/missing/events.jsonl",
		"events: 0",
		"malformed: 0",
		"degraded: false",
		"status: unknown",
		"latest_time: none",
		"latest_loop: none",
		"latest_ticket: none",
		"latest_execution: none",
		"latest_phase: none",
		"latest_level: none",
		"latest_message: none",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestWorkRunAllExecutesReadyTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, nil)

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", withLoopTicketCreatedAt(testLoopTicket(
		"CHC-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	), time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)))
	writeTicketMarkdownOnly(t, projectDir, "CHC-2-Ready.md", withLoopTicketCreatedAt(testLoopTicket(
		"CHC-2", "Ready backend", "ready_for_dev", "normal", "normal", "bug", "maintenance", "backend", []string{"api"},
	), time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC)))

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"work",
		"run-all",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("work run-all returned error: %v", err)
	}

	for _, want := range []string{
		"work: run-all",
		"loop: loop-",
		"status: blocked",
		"eligible: 2",
		"processed: 1",
		"skipped: 1",
		"blocked: 1",
		"execution: exec-",
		"CHC-1",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	if strings.Contains(stdout.String(), "CHC-2") {
		t.Fatalf("stdout contains second ticket execution:\n%s", stdout.String())
	}

	chc1Records := listTicketExecutions(t, devRoot, projectID(projectDir), "CHC-1")
	if len(chc1Records) != 1 {
		t.Fatalf("CHC-1 execution records = %d, want 1", len(chc1Records))
	}

	chc2Records := listTicketExecutions(t, devRoot, projectID(projectDir), "CHC-2")
	if len(chc2Records) != 0 {
		t.Fatalf("CHC-2 execution records = %d, want 0", len(chc2Records))
	}
}

func TestWorkRunAllDryRunPrintsQueue(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", testLoopTicket(
		"CHC-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"work",
		"run-all",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("work run-all --dry-run returned error: %v", err)
	}

	for _, want := range []string{
		"work: run-all",
		"dry_run: true",
		"order: created_at",
		"eligible: 1",
		"cycle_ticket: CHC-1",
		"remaining_eligible: 0",
		"ticket: CHC-1\tReady frontend\tready_for_dev",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	records := listTicketExecutions(t, devRoot, projectID(projectDir), "CHC-1")
	if len(records) != 0 {
		t.Fatalf("execution records = %d, want 0", len(records))
	}
}

func TestWorkRunAllDryRunAppliesFiltersAndOrder(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", withLoopTicketCreatedAt(testLoopTicket(
		"CHC-1", "Ready normal", "ready_for_dev", "normal", "normal", "feature", "onboarding", "frontend", []string{"uxui"},
	), time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)))
	writeTicketMarkdownOnly(t, projectDir, "CHC-2-Ready.md", withLoopTicketCreatedAt(testLoopTicket(
		"CHC-2", "Ready high", "ready_for_dev", "high", "critical", "bug", "maintenance", "backend", []string{"api"},
	), time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC)))

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"work",
		"run-all",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--dry-run",
		"--priority",
		"high",
		"--order",
		"priority",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("work run-all --dry-run returned error: %v", err)
	}

	output := stdout.String()
	for _, want := range []string{
		"work: run-all",
		"dry_run: true",
		"order: priority",
		"eligible: 1",
		"ticket: CHC-2\tReady high\tready_for_dev",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, output)
		}
	}

	if strings.Contains(output, "CHC-1") {
		t.Fatalf("stdout contains filtered ticket CHC-1:\n%s", output)
	}
}

func TestWorkRunAllCompletesWithNoEligibleTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, nil)

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Open.md", testLoopTicket(
		"CHC-1", "Open frontend", "open", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"work",
		"run-all",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("work run-all returned error: %v", err)
	}

	for _, want := range []string{
		"work: run-all",
		"status: succeeded",
		"eligible: 0",
		"processed: 0",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestTicketLoopDispatchesConfiguredFinishedNotification(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	t.Setenv("CHC_TELEGRAM_BOT_TOKEN", "token-1")
	installFakeTicketStartHarnessExecutor(t, nil)

	gotPayload := installTestTelegramSender(t)

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTelegramNotifyConfig(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", testLoopTicket(
		"CHC-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))

	err := run(t.Context(), []string{
		"ticket",
		"loop",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket loop returned error: %v", err)
	}

	records := listNotificationEventsForTest(t, devRoot, projectID(projectDir))
	assertNotificationEvent(t, records, 0, core.NotificationEventLoopFinished, core.NotificationStatusDelivered)

	if gotPayload.ChatID != "123456" {
		t.Fatalf("telegram chat id = %q, want configured chat id", gotPayload.ChatID)
	}

	if !strings.Contains(gotPayload.Text, "Ticket loop loop-") {
		t.Fatalf("telegram text = %q, want loop notification", gotPayload.Text)
	}
}

func TestTicketLoopStopsOnExecutionFailure(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	executor := installFakeTicketStartHarnessExecutor(t, nil)
	executor.failWorker = true

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", withLoopTicketCreatedAt(testLoopTicket(
		"CHC-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	), time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)))
	writeTicketMarkdownOnly(t, projectDir, "CHC-2-Ready.md", withLoopTicketCreatedAt(testLoopTicket(
		"CHC-2", "Ready backend", "ready_for_dev", "normal", "normal", "bug", "maintenance", "backend", []string{"api"},
	), time.Date(2026, 5, 18, 13, 0, 0, 0, time.UTC)))

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"loop",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket loop returned nil error")
	}

	for _, want := range []string{
		"status: failed",
		"eligible: 2",
		"processed: 0",
		"skipped: 1",
		"blocked: 0",
		"failed: 1",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	if strings.Contains(stdout.String(), "CHC-2") {
		t.Fatalf("stdout contains execution after failure:\n%s", stdout.String())
	}

	records := listTicketExecutions(t, devRoot, projectID(projectDir), "CHC-2")
	if len(records) != 0 {
		t.Fatalf("CHC-2 execution records = %d, want 0", len(records))
	}

	loopRuns := listProjectLoopRuns(t, devRoot, projectID(projectDir))
	if len(loopRuns) != 1 {
		t.Fatalf("loop runs = %d, want 1", len(loopRuns))
	}

	if loopRuns[0].Status != core.LoopRunStatusFailed || loopRuns[0].FailedCount != 1 {
		t.Fatalf("loop run = %#v, want failed with one failed ticket", loopRuns[0])
	}
}

func TestTicketLoopRecordsFailedNotification(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	executor := installFakeTicketStartHarnessExecutor(t, nil)
	executor.failWorker = true

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Ready.md", testLoopTicket(
		"CHC-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"},
	))

	err := run(t.Context(), []string{
		"ticket",
		"loop",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket loop returned nil error")
	}

	notificationRecords := listNotificationEventsForTest(t, devRoot, projectID(projectDir))
	assertNotificationEvent(t, notificationRecords, 0, core.NotificationEventLoopFailed, core.NotificationStatusSkipped)
}

func TestTicketLoopRecordsBlockedNotifications(t *testing.T) {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	record := core.LoopRunRecord{
		ID:             "loop-1",
		ProjectID:      "project-1",
		Status:         core.LoopRunStatusBlocked,
		EligibleCount:  1,
		ProcessedCount: 1,
		BlockedCount:   1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	event, ok := loopRunNotificationEvent(commandProjectContext{ProjectID: "project-1"}, record, now)
	if !ok {
		t.Fatal("loopRunNotificationEvent returned false")
	}

	if event.Type != core.NotificationEventLoopBlocked {
		t.Fatalf("event type = %q, want loop_blocked", event.Type)
	}
}

func TestTicketShowPrintsIndexedTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", core.Ticket{
		Key:         "CHC-1",
		Title:       "Do something",
		Status:      "open",
		Priority:    "high",
		Severity:    "normal",
		Type:        "task",
		UseCase:     "onboarding",
		ProjectArea: "uxui",
		Branch:      "feat/slice-1-ticket-review-metadata-contract",
		ReviewRefs: []core.TicketReviewRef{
			{
				Type:     "pull_request",
				Provider: "forgejo",
				URL:      "https://forge.example.com/owner/repo/pulls/123",
				ID:       "123",
				Status:   "open",
			},
		},
		ReviewState:       "changes_requested",
		ReviewRound:       1,
		ReviewResult:      "needs_changes",
		ReviewCommentsURL: "https://forge.example.com/owner/repo/pulls/123",
		CreatedAt:         time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt:         time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"show",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket show returned error: %v", err)
	}

	for _, want := range wantTicketShowReviewOutput() {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestTicketShowPrintsEvidenceStrategySummary(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicket())
	appendTicketEvidenceStrategy(t, ticketPath)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"show",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket show returned error: %v", err)
	}

	for _, want := range wantTicketShowEvidenceOutput() {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func testTicketShowReviewTicket() core.Ticket {
	return core.Ticket{
		Key:         "CHC-1",
		Title:       "Do something",
		Status:      "open",
		Priority:    "high",
		Severity:    "normal",
		Type:        "task",
		UseCase:     "onboarding",
		ProjectArea: "uxui",
		Branch:      "feat/slice-1-ticket-review-metadata-contract",
		ReviewRefs: []core.TicketReviewRef{
			{
				Type:     "pull_request",
				Provider: "forgejo",
				URL:      "https://forge.example.com/owner/repo/pulls/123",
				ID:       "123",
				Status:   "open",
			},
		},
		ReviewState:       "changes_requested",
		ReviewRound:       1,
		ReviewResult:      "needs_changes",
		ReviewCommentsURL: "https://forge.example.com/owner/repo/pulls/123",
		CreatedAt:         time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt:         time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	}
}

func wantTicketShowReviewOutput() []string {
	return []string{
		"key: CHC-1",
		"title: Do something",
		"status: open",
		"priority: high",
		"severity: normal",
		"type: task",
		"use_case: onboarding",
		"project_area: uxui",
		"branch: feat/slice-1-ticket-review-metadata-contract",
		"review_state: changes_requested",
		"review_round: 1",
		"review_result: needs_changes",
		"review_comments_url: https://forge.example.com/owner/repo/pulls/123",
		"path: CHC-1-Do-something.md",
	}
}

func wantTicketShowEvidenceOutput() []string {
	return []string{
		"evidence: complete",
		"evidence_strategy: complete done",
		"evidence_gherkin: required done",
		"evidence_tests: mixed done",
		"evidence_validation: passed done",
		"evidence_summary: complete done",
	}
}

func TestTicketStartDryRunPrintsPreflight(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicket())

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"start",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--dry-run",
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket start returned error: %v", err)
	}

	for _, want := range []string{
		"execution: exec-",
		"ticket: CHC-1",
		"ticket_path: " + filepath.Join(projectDir, ".chc", "user", "tickets", "CHC-1-Do-something.md"),
		"status: pending",
		"dry_run: true",
		"run_artifact_dir: .chc/user/runtime/runs/CHC-1/exec-",
		"result_path: .chc/user/runtime/runs/CHC-1/exec-",
		"ok    dockerfile   .chc/user/runtime/agent/Dockerfile",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	records := listTicketExecutions(t, devRoot, projectID(projectDir), "CHC-1")
	if len(records) != 0 {
		t.Fatalf("execution records = %d, want 0", len(records))
	}
}

func TestTicketStartCreatesPendingExecution(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	executor := installFakeTicketStartHarnessExecutor(t, nil)

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicket())

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"start",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket start returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "dry_run: false") {
		t.Fatalf("stdout = %q, want dry_run false", stdout.String())
	}

	records := listTicketExecutions(t, devRoot, projectID(projectDir), "CHC-1")
	if len(records) != 1 {
		t.Fatalf("execution records = %d, want 1", len(records))
	}

	record := records[0]
	assertTicketStartSucceededExecution(t, record, projectDir)

	if !strings.Contains(strings.Join(executor.commands, ","), "docker build,docker run") {
		t.Fatalf("commands = %#v, want build and run", executor.commands)
	}

	logContent, err := os.ReadFile(record.Artifacts.LogPath)
	if err != nil {
		t.Fatalf("read execution log: %v", err)
	}

	if !strings.Contains(string(logContent), "stdout: docker build") {
		t.Fatalf("log content = %q, want harness output", string(logContent))
	}
}

func TestTicketStartPrintsHarnessFailure(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	installFakeTicketStartHarnessExecutor(t, fmt.Errorf("docker unavailable"))

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicket())

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"start",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket start returned nil error")
	}

	for _, want := range []string{
		"status: failed",
		"failure_kind: harness",
		"failure_reason: harness: docker unavailable",
		"diagnostic_log_path: " + filepath.Join(projectDir, ".chc", "user", "runtime", "runs", "CHC-1"),
		"diagnostic_result_path: " + filepath.Join(projectDir, ".chc", "user", "runtime", "runs", "CHC-1"),
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestTicketStartPrintsWorkerFailure(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	executor := installFakeTicketStartHarnessExecutor(t, nil)
	executor.failWorker = true

	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicket())

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"start",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket start returned nil error")
	}

	for _, want := range []string{
		"status: failed",
		"failure_kind: worker",
		"failure_reason: worker: codex failed",
		"diagnostic_log_path: " + filepath.Join(projectDir, ".chc", "user", "runtime", "runs", "CHC-1"),
		"diagnostic_result_path: " + filepath.Join(projectDir, ".chc", "user", "runtime", "runs", "CHC-1"),
		"diagnostic_validation_path: " + filepath.Join(projectDir, ".chc", "user", "runtime", "runs", "CHC-1"),
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func assertTicketStartSucceededExecution(t *testing.T, record core.ExecutionRecord, projectDir string) {
	t.Helper()

	if record.Status != core.ExecutionStatusSucceeded {
		t.Fatalf("Status = %q, want succeeded", record.Status)
	}

	if record.TicketPath != filepath.Join(projectDir, ".chc", "user", "tickets", "CHC-1-Do-something.md") {
		t.Fatalf("TicketPath = %q, want ticket path", record.TicketPath)
	}

	if !strings.HasPrefix(record.ID, "exec-") {
		t.Fatalf("ID = %q, want exec prefix", record.ID)
	}

	if record.Artifacts.RunDir != filepath.Join(projectDir, ".chc", "user", "runtime", "runs", "CHC-1", record.ID) {
		t.Fatalf("Artifacts.RunDir = %q, want execution run dir", record.Artifacts.RunDir)
	}

	if record.Artifacts.ValidationPath != filepath.Join(record.Artifacts.RunDir, "validation.log") {
		t.Fatalf("Artifacts.ValidationPath = %q, want validation log path", record.Artifacts.ValidationPath)
	}

	contextPath := filepath.Join(record.Artifacts.RunDir, "context.md")

	contextContent, err := os.ReadFile(contextPath)
	if err != nil {
		t.Fatalf("read worker context: %v", err)
	}

	if !strings.Contains(string(contextContent), "- ticket_path: /workspace/.chc/user/tickets/CHC-1-Do-something.md") {
		t.Fatalf("worker context = %q, want container ticket path", string(contextContent))
	}
}

func TestTicketStartFailsWhenRuntimeTemplateIsMissing(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	err := os.MkdirAll(filepath.Join(projectDir, ".chc", "user", "runtime", "runs"), 0o755)
	if err != nil {
		t.Fatalf("create runtime runs dir: %v", err)
	}

	addProject(t, devRoot, "example", projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicket())

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"ticket",
		"start",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket start returned nil error")
	}

	if !strings.Contains(stdout.String(), "fail  dockerfile") {
		t.Fatalf("stdout = %q, want failed Dockerfile check", stdout.String())
	}

	if !strings.Contains(stdout.String(), "failure_kind: runtime") {
		t.Fatalf("stdout = %q, want runtime failure kind", stdout.String())
	}

	records := listTicketExecutions(t, devRoot, projectID(projectDir), "CHC-1")
	if len(records) != 0 {
		t.Fatalf("execution records = %d, want 0", len(records))
	}
}

func TestTicketApplyResultAppliesExecutionResult(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	ticket := testTicketWithReviewRef()
	ticket.Body = testApplyResultTicketBody(t)
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", ticket)

	addProject(t, devRoot, "example", projectDir)

	record := createApplyResultExecution(t, devRoot, projectDir, core.ExecutionStatusSucceeded, "")
	writeApplyResultFile(t, record, core.ExecutionResultStatusSucceeded, "ready for review")

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"apply-result",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		record.ID,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket apply-result returned error: %v", err)
	}

	assertTicketApplyResultOutput(t, stdout.String(), record)

	content := readFile(t, ticketPath)
	assertTicketApplyResultContent(t, content)

	assertPathExists(t, filepath.Join(record.Artifacts.RunDir, "result-application.md"))
	assertPathExists(t, filepath.Join(record.Artifacts.RunDir, "execution-report.md"))
	assertApplyResultExecutionReportEvidence(t, readFile(t, filepath.Join(record.Artifacts.RunDir, "execution-report.md")))

	tickets := listProjectTickets(t, devRoot, projectDir)
	if len(tickets) != 1 || tickets[0].Status != "reviewing" {
		t.Fatalf("indexed tickets = %#v, want reviewing CHC-1", tickets)
	}
}

func TestTicketApplyResultDiscoversWorktreeChangedFiles(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	ticket := testTicketWithReviewRef()
	ticket.Body = testApplyResultTicketBody(t)
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", ticket)
	worktreeDir := newApplyResultGitRepo(t)
	writeGateCLIFile(t, projectDir, "README.md", "dirty user checkout\n")
	writeGateCLIFile(t, worktreeDir, filepath.Join("internal", "core", "worktree.go"), "package core\n")

	addProject(t, devRoot, "example", projectDir)

	record := testApplyResultExecutionRecord(projectDir, core.ExecutionStatusSucceeded, "")
	record.Workspace = core.ExecutionWorkspaceRef{
		Mode:                  core.WorkWorkspaceModeWorktree,
		ProjectCheckoutDir:    projectDir,
		ExecutionWorkspaceDir: worktreeDir,
		BaseRef:               "origin/dev",
		BaseSHA:               "abc123",
		TicketBranch:          "chc/CHC-1",
	}
	runDir := filepath.Join(worktreeDir, ".chc", "user", "runtime", "runs", "CHC-1", "exec-apply")
	record.Artifacts = core.ExecutionArtifacts{
		RunDir:         runDir,
		ResultPath:     filepath.Join(runDir, "result.json"),
		LogPath:        filepath.Join(runDir, "runtime.log"),
		ValidationPath: filepath.Join(runDir, "validation.log"),
	}
	createApplyResultExecutionRecord(t, devRoot, record)
	writeApplyResultFile(t, record, core.ExecutionResultStatusSucceeded, "ready for review")

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"apply-result",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		record.ID,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket apply-result returned error: %v", err)
	}

	content := readFile(t, ticketPath)
	if !strings.Contains(content, "internal/core/worktree.go") {
		t.Fatalf("ticket content missing worktree changed file:\n%s", content)
	}

	if strings.Contains(content, "README.md") {
		t.Fatalf("ticket content includes user checkout changed file:\n%s", content)
	}
}

func TestTicketApplyResultBlocksMissingEvidenceStrategy(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicketWithReviewRef())

	addProject(t, devRoot, "example", projectDir)

	record := createApplyResultExecution(t, devRoot, projectDir, core.ExecutionStatusSucceeded, "")
	writeApplyResultFile(t, record, core.ExecutionResultStatusSucceeded, "ready for review")

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"apply-result",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		record.ID,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket apply-result returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "status: blocked") {
		t.Fatalf("stdout = %q, want blocked status", stdout.String())
	}

	content := readFile(t, ticketPath)
	for _, want := range []string{
		`status: "blocked"`,
		"- kind: validation",
		"evidence_strategy_missing",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("ticket content does not contain %q:\n%s", want, content)
		}
	}
}

func TestTicketApplyResultAllowsMissingOptionalEvidenceStrategy(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	writeCLIWorkspaceConfig(t, projectDir, "evidence:\n  strategy:\n    ticket_checklist: optional\n")
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicketWithReviewRef())

	addProject(t, devRoot, "example", projectDir)

	record := createApplyResultExecution(t, devRoot, projectDir, core.ExecutionStatusSucceeded, "")
	writeApplyResultFile(t, record, core.ExecutionResultStatusSucceeded, "ready for review")

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"apply-result",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		record.ID,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket apply-result returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "status: reviewing") {
		t.Fatalf("stdout = %q, want reviewing status", stdout.String())
	}
}

func TestTicketApplyResultBlocksActiveApplicableGherkinNotApplicable(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	writeCLIAcceptanceFeature(t, projectDir, ".chc/user/acceptance/onboarding.feature")

	ticket := testTicketWithReviewRef()
	ticket.Body = testApplyResultTicketBody(t)
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", ticket)

	addProject(t, devRoot, "example", projectDir)

	record := createApplyResultExecution(t, devRoot, projectDir, core.ExecutionStatusSucceeded, "")
	writeApplyResultFile(t, record, core.ExecutionResultStatusSucceeded, "ready for review")

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"apply-result",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		record.ID,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket apply-result returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "status: blocked") {
		t.Fatalf("stdout = %q, want blocked status", stdout.String())
	}

	content := readFile(t, ticketPath)
	if !strings.Contains(content, "gherkin_required_missing") {
		t.Fatalf("ticket content missing Gherkin blocker:\n%s", content)
	}
}

func testApplyResultTicketBody(t *testing.T) string {
	t.Helper()

	return testApplyResultEvidenceChecklistBlock(t) + `
# Description

Apply behavior.

# Checklist

# Acceptance

# Comments

# Blocker

# Validation
`
}

func testApplyResultEvidenceChecklistBlock(t *testing.T) string {
	t.Helper()

	rendered, err := testCompleteEvidenceChecklistBlock()
	if err != nil {
		t.Fatalf("RenderEvidenceStrategyChecklist returned error: %v", err)
	}

	return rendered
}

func testCompleteEvidenceChecklistBlock() (string, error) {
	checklist := core.EvidenceStrategyChecklist{
		Found: true,
		Items: map[string]core.EvidenceStrategyChecklistItem{
			core.EvidenceStrategyChecklistLabelStrategy: {
				Label:   core.EvidenceStrategyChecklistLabelStrategy,
				Checked: true,
				Status:  "complete",
				Detail:  "change_kind=bugfix; risk=medium; test_timing=before",
			},
			core.EvidenceStrategyChecklistLabelGherkin: {
				Label:   core.EvidenceStrategyChecklistLabelGherkin,
				Checked: true,
				Status:  "not_applicable",
				Detail:  "reason=internal behavior",
			},
			core.EvidenceStrategyChecklistLabelTests: {
				Label:   core.EvidenceStrategyChecklistLabelTests,
				Checked: true,
				Status:  "unit",
				Detail:  "reason=state transition",
			},
			core.EvidenceStrategyChecklistLabelValidation: {
				Label:   core.EvidenceStrategyChecklistLabelValidation,
				Checked: true,
				Status:  "passed",
				Detail:  "commands=make verify",
			},
			core.EvidenceStrategyChecklistLabelSummary: {
				Label:   core.EvidenceStrategyChecklistLabelSummary,
				Checked: true,
				Status:  "complete",
			},
		},
	}

	return core.RenderEvidenceStrategyChecklist(checklist)
}

func assertApplyResultExecutionReportEvidence(t *testing.T, content string) {
	t.Helper()

	for _, want := range []string{
		"## Evidence Strategy",
		"### Result Summary",
		"- strategy: planned",
		"- produced: evidence produced",
		"- validation: validation passed",
		"### Ticket Checklist",
		"- status: present",
		"- Evidence strategy: checked=true; status=complete; change_kind=bugfix",
		"- Tests: checked=true; status=unit; reason=state transition",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("execution report missing %q:\n%s", want, content)
		}
	}
}

func testTicketWithReviewRef() core.Ticket {
	ticket := testTicket()
	ticket.Branch = "feat/slice-2-pr-reference-workflow"
	ticket.ReviewRefs = []core.TicketReviewRef{
		{
			Type:     "pull_request",
			Provider: "forgejo",
			URL:      "https://forge.example.com/adrianpk/chc/pulls/274",
			ID:       "274",
			Status:   "open",
		},
	}

	return ticket
}

func testTicketReadyForReview() core.Ticket {
	ticket := testTicketWithReviewRef()
	ticket.Status = core.TicketStatusReadyForReview

	return ticket
}

func testTicketReadyForMerge() core.Ticket {
	ticket := testTicketWithReviewRef()
	ticket.Status = core.TicketStatusReadyForMerge
	ticket.ReviewState = core.TicketReviewCommandResultPass
	ticket.ReviewResult = core.TicketReviewCommandResultPass
	ticket.ReviewRound = 1

	return ticket
}

func setupCLIReviewRemoteEvidence(
	t *testing.T,
	devRoot string,
	projectDir string,
) (*fakeCLIReviewRemoteRunner, func()) {
	t.Helper()

	return setupCLIReviewRemoteBase(t, devRoot, projectDir, "dev")
}

func setupCLIReviewRemoteBase(t *testing.T, devRoot, projectDir, base string) (*fakeCLIReviewRemoteRunner, func()) {
	t.Helper()

	runner := newFakeCLIReviewRemoteRunner()
	runner.outputs["tea api /repos/adrianpk/chc/pulls/274"] = []byte(`{"state":"open","base":{"ref":"` + base + `"},"head":{"ref":"feat/slice-2-pr-reference-workflow","sha":"abc123"},"mergeable":true}`)
	runner.outputs["tea api /repos/adrianpk/chc/pulls/274.diff"] = []byte("diff --git a/file.go b/file.go")
	runner.outputs["tea api /repos/adrianpk/chc/issues/274/comments"] = []byte(`[]`)
	runner.outputs["tea api /repos/adrianpk/chc/commits/abc123/statuses"] = []byte(`[]`)

	addProject(t, devRoot, "example", projectDir)

	return runner, replaceReviewRemoteCommandRunner(runner)
}

func assertTicketApplyResultOutput(t *testing.T, stdout string, record core.ExecutionRecord) {
	t.Helper()

	for _, want := range []string{
		"execution: exec-apply",
		"ticket: CHC-1",
		"status: reviewing",
		"result_status: succeeded",
		"execution_report_path: " + filepath.Join(record.Artifacts.RunDir, "execution-report.md"),
		"dry_run: false",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout)
		}
	}
}

func assertTicketApplyResultContent(t *testing.T, content string) {
	t.Helper()

	for _, want := range []string{
		`status: "reviewing"`,
		"Execution `exec-apply` finished with status `succeeded`.",
		"Summary: ready for review",
		`branch: "feat/slice-2-pr-reference-workflow"`,
		`url: "https://forge.example.com/adrianpk/chc/pulls/274"`,
		"## ",
		"- command: make verify",
		"- status: succeeded",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("ticket content does not contain %q:\n%s", want, content)
		}
	}
}

func TestTicketApplyResultRecordsBlockedNotification(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicket())

	addProject(t, devRoot, "example", projectDir)

	record := createApplyResultExecution(t, devRoot, projectDir, core.ExecutionStatusFailed, "codex failed")
	writeApplyResultFile(t, record, core.ExecutionResultStatusFailed, "blocked by compile error")

	err := run(t.Context(), []string{
		"ticket",
		"apply-result",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		record.ID,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket apply-result returned error: %v", err)
	}

	records := listNotificationEventsForTest(t, devRoot, projectID(projectDir))
	assertNotificationEvent(t, records, 0, core.NotificationEventTicketBlocked, core.NotificationStatusSkipped)

	if records[0].Event.References.ExecutionID != record.ID {
		t.Fatalf("execution reference = %q, want %q", records[0].Event.References.ExecutionID, record.ID)
	}
}

func TestTicketApplyResultDryRunDoesNotMutateTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicket())

	addProject(t, devRoot, "example", projectDir)

	record := createApplyResultExecution(t, devRoot, projectDir, core.ExecutionStatusSucceeded, "")
	writeApplyResultFile(t, record, core.ExecutionResultStatusSucceeded, "ready for review")

	before := readFile(t, ticketPath)

	err := run(t.Context(), []string{
		"ticket",
		"apply-result",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--dry-run",
		record.ID,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket apply-result returned error: %v", err)
	}

	after := readFile(t, ticketPath)
	if after != before {
		t.Fatalf("dry-run mutated ticket:\n%s", after)
	}

	assertPathNotExists(t, filepath.Join(record.Artifacts.RunDir, "result-application.md"))
	assertPathNotExists(t, filepath.Join(record.Artifacts.RunDir, "execution-report.md"))
}

func TestReportListListsProjectReports(t *testing.T) {
	projectDir := newTestWorkspace(t)

	writeCLIReport(t, projectDir, filepath.Join(
		".chc",
		"user",
		"runtime",
		"runs",
		"CHC-1",
		"exec-1",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: exec-1\n- ticket: CHC-1\n- status: succeeded\n")
	writeCLIReport(t, projectDir, filepath.Join(
		".chc",
		"user",
		"runtime",
		"runs",
		"loop-1",
		"loop-report.md",
	), "# Loop Report\n\n## Summary\n\n- loop: loop-1\n- status: blocked\n")

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"report",
		"list",
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("report list returned error: %v", err)
	}

	for _, want := range []string{
		"kind: execution path: .chc/user/runtime/runs/CHC-1/exec-1/execution-report.md ticket: CHC-1 execution: exec-1 loop: none status: succeeded",
		"kind: loop path: .chc/user/runtime/runs/loop-1/loop-report.md ticket: none execution: none loop: loop-1 status: blocked",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestReportListFiltersReports(t *testing.T) {
	projectDir := newTestWorkspace(t)

	writeCLIReport(t, projectDir, filepath.Join(
		".chc",
		"user",
		"runtime",
		"runs",
		"CHC-1",
		"exec-1",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: exec-1\n- ticket: CHC-1\n- status: succeeded\n")
	writeCLIReport(t, projectDir, filepath.Join(
		".chc",
		"user",
		"runtime",
		"runs",
		"CHC-2",
		"exec-2",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: exec-2\n- ticket: CHC-2\n- status: blocked\n")

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"report",
		"list",
		"--dir",
		projectDir,
		"--kind",
		"execution",
		"--ticket",
		"CHC-2",
		"--status",
		"blocked",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("report list returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "execution: exec-2") {
		t.Fatalf("stdout does not contain exec-2:\n%s", stdout.String())
	}

	if strings.Contains(stdout.String(), "exec-1") {
		t.Fatalf("stdout contains filtered report:\n%s", stdout.String())
	}
}

func TestReportShowPrintsReportBody(t *testing.T) {
	projectDir := newTestWorkspace(t)

	writeCLIReport(t, projectDir, filepath.Join(
		".chc",
		"user",
		"runtime",
		"runs",
		"CHC-1",
		"exec-1",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: exec-1\n- ticket: CHC-1\n- status: succeeded\n")

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"report",
		"show",
		"--dir",
		projectDir,
		"--execution",
		"exec-1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("report show returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "# Execution Report") {
		t.Fatalf("stdout does not contain report body:\n%s", stdout.String())
	}
}

func TestReportShowPrintsReportMetadata(t *testing.T) {
	projectDir := newTestWorkspace(t)

	writeCLIReport(t, projectDir, filepath.Join(
		".chc",
		"user",
		"runtime",
		"runs",
		"loop-1",
		"loop-report.md",
	), "# Loop Report\n\n## Summary\n\n- loop: loop-1\n- status: blocked\n")

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"report",
		"show",
		"--dir",
		projectDir,
		"--loop",
		"loop-1",
		"--metadata",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("report show returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "kind: loop path: .chc/user/runtime/runs/loop-1/loop-report.md") {
		t.Fatalf("stdout does not contain report metadata:\n%s", stdout.String())
	}
}

func TestReportShowRejectsAmbiguousLookup(t *testing.T) {
	projectDir := newTestWorkspace(t)

	writeCLIReport(t, projectDir, filepath.Join(
		".chc",
		"user",
		"runtime",
		"runs",
		"CHC-1",
		"exec-1",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: exec-1\n- ticket: CHC-1\n")
	writeCLIReport(t, projectDir, filepath.Join(
		".chc",
		"user",
		"runtime",
		"runs",
		"CHC-1",
		"exec-2",
		"execution-report.md",
	), "# Execution Report\n\n## Summary\n\n- execution: exec-2\n- ticket: CHC-1\n")

	err := run(t.Context(), []string{
		"report",
		"show",
		"--dir",
		projectDir,
		"--ticket",
		"CHC-1",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("report show returned nil error")
	}

	if !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("error = %v, want ambiguous lookup", err)
	}
}

func TestTicketShowRejectsMissingTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"ticket",
		"show",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket show returned nil error")
	}
}

func TestTicketCommentAppendsComment(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", core.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"comment",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--author",
		"adrian",
		"CHC-1",
		"Looks ready.",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket comment returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "ticket: "+ticketPath) {
		t.Fatalf("stdout = %q, want ticket path", stdout.String())
	}

	updatedContent, err := os.ReadFile(ticketPath)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	for _, want := range []string{
		"# Comments",
		"adrian",
		"Looks ready.",
	} {
		if !strings.Contains(string(updatedContent), want) {
			t.Fatalf("updated content does not contain %q:\n%s", want, string(updatedContent))
		}
	}
}

func writeIndexedTicketFile(t *testing.T, devRoot, projectDir string) string {
	t.Helper()

	ticketPath := filepath.Join(projectDir, ".chc", "user", "tickets", "CHC-1-Do-something.md")

	content, err := core.NewTicketRenderer().RenderTicket(core.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	err = os.MkdirAll(filepath.Dir(ticketPath), 0o755)
	if err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	err = os.WriteFile(ticketPath, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write ticket: %v", err)
	}

	indexTicket(t, devRoot, projectDir, core.IndexedTicket{
		Key:          "CHC-1",
		Title:        "Do something",
		Status:       "open",
		RelativePath: filepath.Join(".chc", "user", "tickets", "CHC-1-Do-something.md"),
		Path:         ticketPath,
	})

	return ticketPath
}

func TestTicketCommentRejectsMissingTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"ticket",
		"comment",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
		"Looks ready.",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket comment returned nil error")
	}
}

func TestTicketStatusUpdatesTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", core.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
		"ready",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket status returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "ticket: CHC-1 ready") {
		t.Fatalf("stdout = %q, want status result", stdout.String())
	}

	updatedContent, err := os.ReadFile(ticketPath)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	if !strings.Contains(string(updatedContent), `status: "ready"`) {
		t.Fatalf("updated content does not contain ready status:\n%s", string(updatedContent))
	}
}

func TestTicketReviewRefAttachesBranchAndPullRequest(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	ticket := testTicket()
	ticket.Body = testApplyResultTicketBody(t)
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", ticket)

	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"review-ref",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--branch",
		"feat/slice-2-pr-reference-workflow",
		"--provider",
		"forgejo",
		"--url",
		"https://forge.example.com/adrianpk/chc/pulls/274",
		"--id",
		"274",
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket review-ref returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "ticket: CHC-1 review_ref") {
		t.Fatalf("stdout = %q, want review-ref result", stdout.String())
	}

	err = run(t.Context(), []string{
		"ticket",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
		core.TicketStatusReadyForReview,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket status returned error: %v", err)
	}

	content := readFile(t, ticketPath)
	for _, want := range []string{
		`branch: "feat/slice-2-pr-reference-workflow"`,
		`review_refs:`,
		`url: "https://forge.example.com/adrianpk/chc/pulls/274"`,
		`status: "ready_for_review"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("ticket content does not contain %q:\n%s", want, content)
		}
	}
}

func TestTicketStatusBlocksReadyForReviewWithoutEvidenceStrategy(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicketWithReviewRef())

	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"ticket",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
		core.TicketStatusReadyForReview,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket status returned nil error")
	}

	if !strings.Contains(err.Error(), "evidence_strategy_missing") {
		t.Fatalf("err = %v, want missing evidence strategy", err)
	}

	content := readFile(t, ticketPath)
	if strings.Contains(content, `status: "ready_for_review"`) {
		t.Fatalf("ticket content moved to ready_for_review:\n%s", content)
	}
}

func TestTicketStatusAllowsReadyForReviewWithOptionalEvidenceStrategy(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	writeCLIWorkspaceConfig(t, projectDir, "evidence:\n  strategy:\n    ticket_checklist: optional\n")

	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicketWithReviewRef())

	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"ticket",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
		core.TicketStatusReadyForReview,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket status returned error: %v", err)
	}

	content := readFile(t, ticketPath)
	if !strings.Contains(content, `status: "ready_for_review"`) {
		t.Fatalf("ticket content does not contain ready_for_review:\n%s", content)
	}
}

func TestTicketReviewModeStatusPrintsDefaultMode(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"review-mode",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--dry-run",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket review-mode status returned error: %v", err)
	}

	for _, want := range []string{
		"review_mode: status",
		"mode: human_review",
		"mode_status: ready",
		"dry_run: true",
		"requires_human_review: true",
		"clean_pass_remote_comment: false",
		"non_blocking_comment_limit: 3",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestTicketReviewModeStatusPrintsConfiguredModes(t *testing.T) {
	tests := []struct {
		name string
		mode string
		want string
	}{
		{name: "agentic", mode: "agentic_review", want: "requires_remote_review: true"},
		{name: "human", mode: "human_review", want: "requires_human_review: true"},
		{name: "trusted", mode: "trusted_auto_merge", want: "trusted_auto_merge: true"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			projectDir := newTestWorkspace(t)
			devRoot := t.TempDir()

			writeCLIWorkspaceConfig(t, projectDir, `review:
  mode: `+test.mode+`
  non_blocking_comment_limit: 1
`)
			addProject(t, devRoot, "example", projectDir)

			var stdout bytes.Buffer

			err := run(t.Context(), []string{
				"ticket",
				"review-mode",
				"status",
				"--dev-root",
				devRoot,
				"--dir",
				projectDir,
			}, &stdout, &bytes.Buffer{})
			if err != nil {
				t.Fatalf("ticket review-mode status returned error: %v", err)
			}

			for _, want := range []string{
				"mode: " + test.mode,
				"mode_status: ready",
				test.want,
				"non_blocking_comment_limit: 1",
			} {
				if !strings.Contains(stdout.String(), want) {
					t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
				}
			}
		})
	}
}

func TestTicketReviewModeStatusPrintsInvalidMode(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	writeCLIWorkspaceConfig(t, projectDir, `review:
  mode: unknown
`)
	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"review-mode",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket review-mode status returned error: %v", err)
	}

	for _, want := range []string{
		"mode: unknown",
		"mode_status: invalid",
		"error: invalid review workflow mode: unknown",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestTicketReviewPassUpdatesTicketWithoutRemoteComment(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicketReadyForReview())
	runner, restoreRunner := setupCLIReviewRemoteEvidence(t, devRoot, projectDir)

	defer restoreRunner()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"review",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--result",
		"pass",
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket review returned error: %v", err)
	}

	for _, want := range []string{
		"ticket_review: result",
		"ticket: CHC-1",
		"result: pass",
		"review_state: pass",
		"status: ready_for_merge",
		"remote_comments: 0",
		"dry_run: false",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	content := readFile(t, ticketPath)
	for _, want := range []string{
		`status: "ready_for_merge"`,
		`review_state: "pass"`,
		`review_result: "pass"`,
		`review_comments_url: "https://forge.example.com/adrianpk/chc/pulls/274"`,
		`review_round: 1`,
		`blocking_findings: []`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("ticket content does not contain %q:\n%s", want, content)
		}
	}

	if len(runner.inputs) != 0 {
		t.Fatalf("remote inputs = %#v, want none", runner.inputs)
	}
}

func TestTicketReviewNeedsChangesWritesRemoteFinding(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicketReadyForReview())
	runner, restoreRunner := setupCLIReviewRemoteEvidence(t, devRoot, projectDir)

	defer restoreRunner()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"review",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--result",
		"needs_changes",
		"--finding",
		"Handle nil input.",
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket review returned error: %v", err)
	}

	for _, want := range []string{
		"result: needs_changes",
		"review_state: changes_requested",
		"status: ready_for_dev",
		"remote_comments: 1",
		"finding: Handle nil input.",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	content := readFile(t, ticketPath)
	for _, want := range []string{
		`status: "ready_for_dev"`,
		`review_state: "changes_requested"`,
		`review_result: "needs_changes"`,
		`blocking_findings: ["Handle nil input."]`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("ticket content does not contain %q:\n%s", want, content)
		}
	}

	command := "tea api -X POST /repos/adrianpk/chc/issues/274/comments -d @-"
	input := string(runner.inputs[command])

	if !strings.Contains(input, "Handle nil input.") {
		t.Fatalf("remote comment payload = %q", input)
	}
}

func TestTicketReviewDryRunDoesNotMutateTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicketReadyForReview())
	runner, restoreRunner := setupCLIReviewRemoteEvidence(t, devRoot, projectDir)

	defer restoreRunner()

	before := readFile(t, ticketPath)

	err := run(t.Context(), []string{
		"ticket",
		"review",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--dry-run",
		"--result",
		"needs_changes",
		"--finding",
		"Handle nil input.",
		"CHC-1",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket review dry-run returned error: %v", err)
	}

	after := readFile(t, ticketPath)
	if after != before {
		t.Fatalf("dry-run mutated ticket:\n%s", after)
	}

	if len(runner.inputs) != 0 {
		t.Fatalf("remote inputs = %#v, want none", runner.inputs)
	}
}

func TestTicketMergeDryRunDoesNotMutateTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicketReadyForMerge())
	runner, restoreRunner := setupCLIReviewRemoteEvidence(t, devRoot, projectDir)

	defer restoreRunner()

	before := readFile(t, ticketPath)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"merge",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--dry-run",
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket merge dry-run returned error: %v", err)
	}

	for _, want := range []string{
		"ticket_merge: result",
		"ticket: CHC-1",
		"allowed: true",
		"status: ready_for_merge",
		"merge_state: pending",
		"merge_strategy: rebase",
		"dry_run: true",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	after := readFile(t, ticketPath)
	if after != before {
		t.Fatalf("dry-run mutated ticket:\n%s", after)
	}

	if len(runner.inputs) != 0 {
		t.Fatalf("remote inputs = %#v, want none", runner.inputs)
	}
}

func TestTicketMergeAppliesSuccessfulMerge(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicketReadyForMerge())
	runner, restoreRunner := setupCLIReviewRemoteEvidence(t, devRoot, projectDir)

	defer restoreRunner()

	mergeCommand := "tea api -X PUT /repos/adrianpk/chc/pulls/274/merge -d @-"
	runner.inputOutputs[mergeCommand] = []byte(`{"sha":"abc123"}`)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"merge",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket merge returned error: %v", err)
	}

	for _, want := range []string{
		"ticket_merge: result",
		"ticket: CHC-1",
		"allowed: true",
		"status: done",
		"merge_state: merged",
		"merge_result: merged",
		"merge_commit: abc123",
		"dry_run: false",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	content := readFile(t, ticketPath)
	for _, want := range []string{
		`status: "done"`,
		`merge_state: "merged"`,
		`merge_result: "merged"`,
		`merge_actor: "merger"`,
		`merge_commit: "abc123"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("ticket content does not contain %q:\n%s", want, content)
		}
	}

	input := string(runner.inputs[mergeCommand])
	if !strings.Contains(input, `"Do":"rebase"`) {
		t.Fatalf("merge payload = %q, want rebase strategy", input)
	}
}

func TestTicketMergeUsesConfiguredIntegrationBranch(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	writeCLIWorkspaceConfig(t, projectDir, `work:
  integration_branch: develop
  canonical_remote: upstream
`)
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicketReadyForMerge())

	runner, restoreRunner := setupCLIReviewRemoteBase(t, devRoot, projectDir, "develop")

	defer restoreRunner()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"merge",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--dry-run",
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket merge returned error: %v", err)
	}

	for _, want := range []string{
		"ticket_merge: result",
		"allowed: true",
		"merge_state: pending",
		"dry_run: true",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	if len(runner.inputs) != 0 {
		t.Fatalf("remote inputs = %#v, want none", runner.inputs)
	}
}

func TestTicketMergeRecordsProviderFailure(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	ticketPath := writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicketReadyForMerge())
	runner, restoreRunner := setupCLIReviewRemoteEvidence(t, devRoot, projectDir)

	defer restoreRunner()

	mergeCommand := "tea api -X PUT /repos/adrianpk/chc/pulls/274/merge -d @-"
	runner.inputErrs[mergeCommand] = errors.New("provider failed")

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"merge",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket merge returned error: %v", err)
	}

	for _, want := range []string{
		"allowed: false",
		"status: ready_for_merge",
		"merge_state: blocked",
		"merge_result: blocked",
		"blocker: provider_unreadable kind=operational detail=provider failed",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}

	content := readFile(t, ticketPath)
	for _, want := range []string{
		`status: "ready_for_merge"`,
		`merge_state: "blocked"`,
		`merge_result: "blocked"`,
		`merge_blocker: "provider failed"`,
		`merged_at: ""`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("ticket content does not contain %q:\n%s", want, content)
		}
	}
}

func TestTicketReviewModeDiagnosticsPrintsProviderStatus(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	writeCLIWorkspaceConfig(t, projectDir, `review:
  forgejo_cli: tea
  forgejo_token_env: CHC_FORGEJO_TOKEN
`)
	addProject(t, devRoot, "example", projectDir)

	restoreRunner := replaceReviewProviderDiagnosticRunner(fakeCLIReviewProviderDiagnosticRunner{
		paths: map[string]string{"tea": "/usr/bin/tea"},
		env:   map[string]string{"CHC_FORGEJO_TOKEN": "secret"},
	})
	defer restoreRunner()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"review-mode",
		"diagnostics",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--url",
		"https://forge.adrianpk.com/adrianpk/chc/pulls/275",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket review-mode diagnostics returned error: %v", err)
	}

	for _, want := range []string{
		"review_provider_diagnostics: status",
		"provider: forgejo",
		"provider_status: ready",
		"host: forge.adrianpk.com",
		"owner: adrianpk",
		"repo: chc",
		"number: 275",
		"check.cli_available: ok reason=found detail=/usr/bin/tea",
		"check.token_available: ok reason=found detail=CHC_FORGEJO_TOKEN",
		"check.pr_readable: ok reason=readable",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestTicketReviewModeDiagnosticsPrintsMissingTooling(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	restoreRunner := replaceReviewProviderDiagnosticRunner(fakeCLIReviewProviderDiagnosticRunner{})
	defer restoreRunner()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"ticket",
		"review-mode",
		"diagnostics",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"--url",
		"https://github.com/example/repo/pull/34",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("ticket review-mode diagnostics returned error: %v", err)
	}

	for _, want := range []string{
		"provider: github",
		"provider_status: blocked",
		"check.cli_available: fail reason=missing_cli detail=gh",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, stdout.String())
		}
	}
}

func TestTicketStatusRejectsInvalidStatus(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	writeIndexedTicketFile(t, devRoot, projectDir)
	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"ticket",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
		"unknown",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket status returned nil error")
	}
}

func TestTicketStatusRejectsReadyForReviewWithoutReviewRef(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", testTicket())
	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"ticket",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
		core.TicketStatusReadyForReview,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket status returned nil error")
	}
}

func TestTicketStatusRejectsMissingTicket(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"ticket",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
		"CHC-1",
		"ready",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket status returned nil error")
	}
}

func TestIndexRebuildIndexesTicketFiles(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", core.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})
	indexTicket(t, devRoot, projectDir, core.IndexedTicket{
		Key:          "CHC-9",
		Title:        "Old",
		Status:       "open",
		RelativePath: "CHC-9-Old.md",
		Path:         filepath.Join(projectDir, ".chc", "user", "tickets", "CHC-9-Old.md"),
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"index",
		"rebuild",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("index rebuild returned error: %v", err)
	}

	for _, want := range []string{
		"indexed: 1",
		"parse_errors: 0",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}

	tickets := listProjectTickets(t, devRoot, projectDir)
	if len(tickets) != 1 {
		t.Fatalf("len(tickets) = %d, want 1", len(tickets))
	}

	if tickets[0].Key != "CHC-1" || tickets[0].Title != "Do something" {
		t.Fatalf("ticket = %#v, want rebuilt CHC-1", tickets[0])
	}
}

func TestIndexRebuildReportsParseErrors(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	badTicketPath := filepath.Join(projectDir, ".chc", "user", "tickets", "CHC-1-Bad.md")

	err := os.MkdirAll(filepath.Dir(badTicketPath), 0o755)
	if err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	err = os.WriteFile(badTicketPath, []byte("# Bad\n"), 0o644)
	if err != nil {
		t.Fatalf("write bad ticket: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"index",
		"rebuild",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("index rebuild returned error: %v", err)
	}

	for _, want := range []string{
		"indexed: 0",
		"parse_errors: 1",
		"parse_error: CHC-1-Bad.md",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestIndexStatusReportsCleanIndex(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", core.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	err := run(t.Context(), []string{
		"index",
		"rebuild",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("index rebuild returned error: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"index",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("index status returned error: %v", err)
	}

	for _, want := range []string{
		"missing: 0",
		"stale: 0",
		"orphaned: 0",
		"invalid: 0",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestIndexStatusReportsMissingTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", core.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"index",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("index status returned nil error")
	}

	for _, want := range []string{
		"missing: 1",
		"missing_ticket: CHC-1-Do-something.md",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestIndexStatusReportsStaleTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", core.Ticket{
		Key:       "CHC-1",
		Title:     "Do something",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	})

	err := run(t.Context(), []string{
		"index",
		"rebuild",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("index rebuild returned error: %v", err)
	}

	writeTicketMarkdownOnly(t, projectDir, "CHC-1-Do-something.md", core.Ticket{
		Key:       "CHC-1",
		Title:     "Do something else",
		Status:    "open",
		CreatedAt: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 5, 18, 12, 30, 0, 0, time.UTC),
	})

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"index",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("index status returned nil error")
	}

	for _, want := range []string{
		"stale: 1",
		"stale_ticket: CHC-1-Do-something.md",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestIndexStatusReportsInvalidTickets(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	badTicketPath := filepath.Join(projectDir, ".chc", "user", "tickets", "CHC-1-Bad.md")

	err := os.MkdirAll(filepath.Dir(badTicketPath), 0o755)
	if err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	err = os.WriteFile(badTicketPath, []byte("# Bad\n"), 0o644)
	if err != nil {
		t.Fatalf("write bad ticket: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"index",
		"status",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("index status returned nil error")
	}

	for _, want := range []string{
		"invalid: 1",
		"invalid_ticket: CHC-1-Bad.md",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestMaintenanceDBStatusReportsPendingMigrationsWhenDatabaseIsMissing(t *testing.T) {
	devRoot := t.TempDir()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"maintenance",
		"db",
		"status",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance db status returned error: %v", err)
	}

	for _, want := range []string{
		"database_exists: false",
		"schema_migrations_exists: false",
		"applied: 0",
		"pending: 11",
		"unknown: 0",
		"pending_migration: 20260518000001 create_projects",
		"pending_migration: 20260601000009 add_ticket_review_metadata",
		"pending_migration: 20260602000010 add_ticket_merge_metadata",
		"pending_migration: 20260619000011 add_execution_workspace_metadata",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestMaintenanceDBStatusReportsAppliedMigrations(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	err := run(t.Context(), []string{
		"project",
		"add",
		"--dev-root",
		devRoot,
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project add returned error: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"maintenance",
		"db",
		"status",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance db status returned error: %v", err)
	}

	for _, want := range []string{
		"database_exists: true",
		"schema_migrations_exists: true",
		"available: 11",
		"applied: 11",
		"pending: 0",
		"unknown: 0",
		"applied_migration: 20260518000001 create_projects",
		"applied_migration: 20260601000009 add_ticket_review_metadata",
		"applied_migration: 20260602000010 add_ticket_merge_metadata",
		"applied_migration: 20260619000011 add_execution_workspace_metadata",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestMaintenanceRegistryStatusReportsRegisteredProjectDiagnostics(t *testing.T) {
	projectDir := newTestWorkspace(t)
	noWorkspaceDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	addProject(t, devRoot, "no-workspace", noWorkspaceDir)

	err := os.RemoveAll(filepath.Join(noWorkspaceDir, ".chc"))
	if err != nil {
		t.Fatalf("RemoveAll returned error: %v", err)
	}

	err = run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		devRoot,
		"example",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project select returned error: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"maintenance",
		"registry",
		"status",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance registry status returned error: %v", err)
	}

	for _, want := range []string{
		"projects: 2",
		"project: example path: " + projectDir + " enabled: true active: true path_exists: true workspace_exists: true status: ok",
		"project: no-workspace path: " + noWorkspaceDir + " enabled: true active: false path_exists: true workspace_exists: false status: missing_workspace",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestMaintenanceRegistryStatusReportsMissingProjectPath(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "missing", projectDir)

	err := os.RemoveAll(projectDir)
	if err != nil {
		t.Fatalf("RemoveAll returned error: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"maintenance",
		"registry",
		"status",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance registry status returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "status: missing_path") {
		t.Fatalf("stdout = %q, want missing path status", stdout.String())
	}
}

func TestMaintenanceRegistryDisableDisablesProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"maintenance",
		"registry",
		"disable",
		"--dev-root",
		devRoot,
		"example",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance registry disable returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "disabled: example path: "+projectDir) {
		t.Fatalf("stdout = %q, want disabled project", stdout.String())
	}

	var statusStdout bytes.Buffer

	err = run(t.Context(), []string{
		"maintenance",
		"registry",
		"status",
		"--dev-root",
		devRoot,
	}, &statusStdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance registry status returned error: %v", err)
	}

	if !strings.Contains(statusStdout.String(), "enabled: false active: false path_exists: true workspace_exists: true status: disabled") {
		t.Fatalf("stdout = %q, want disabled status", statusStdout.String())
	}
}

func TestMaintenanceRegistryDisableDryRunDoesNotMutateProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"maintenance",
		"registry",
		"disable",
		"--dev-root",
		devRoot,
		"--dry-run",
		"example",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance registry disable returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "would_disable: example path: "+projectDir) {
		t.Fatalf("stdout = %q, want dry-run disable", stdout.String())
	}

	var statusStdout bytes.Buffer

	err = run(t.Context(), []string{
		"maintenance",
		"registry",
		"status",
		"--dev-root",
		devRoot,
	}, &statusStdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance registry status returned error: %v", err)
	}

	if !strings.Contains(statusStdout.String(), "enabled: true active: false path_exists: true workspace_exists: true status: ok") {
		t.Fatalf("stdout = %q, want unchanged status", statusStdout.String())
	}
}

func TestMaintenanceRegistrySetPathUpdatesProjectPath(t *testing.T) {
	projectDir := newTestWorkspace(t)
	movedDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"maintenance",
		"registry",
		"set-path",
		"--dev-root",
		devRoot,
		"example",
		movedDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance registry set-path returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "updated_path: example from: "+projectDir+" to: "+movedDir) {
		t.Fatalf("stdout = %q, want updated path", stdout.String())
	}

	var statusStdout bytes.Buffer

	err = run(t.Context(), []string{
		"maintenance",
		"registry",
		"status",
		"--dev-root",
		devRoot,
	}, &statusStdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance registry status returned error: %v", err)
	}

	if !strings.Contains(statusStdout.String(), "project: example path: "+movedDir) {
		t.Fatalf("stdout = %q, want moved path", statusStdout.String())
	}
}

func TestMaintenanceRegistrySetPathDryRunDoesNotMutateProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	movedDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"maintenance",
		"registry",
		"set-path",
		"--dev-root",
		devRoot,
		"--dry-run",
		"example",
		movedDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance registry set-path returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "would_update_path: example from: "+projectDir+" to: "+movedDir) {
		t.Fatalf("stdout = %q, want dry-run path update", stdout.String())
	}

	var statusStdout bytes.Buffer

	err = run(t.Context(), []string{
		"maintenance",
		"registry",
		"status",
		"--dev-root",
		devRoot,
	}, &statusStdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance registry status returned error: %v", err)
	}

	if !strings.Contains(statusStdout.String(), "project: example path: "+projectDir) {
		t.Fatalf("stdout = %q, want original path", statusStdout.String())
	}
}

func TestMaintenanceExecutionArtifactsReportsMissingArtifacts(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	runDir := filepath.Join(projectDir, ".chc", "user", "runtime", "runs", "exec-1")
	record := cliExecutionRecord(projectDir, runDir)

	err := os.MkdirAll(record.Artifacts.RunDir, 0o755)
	if err != nil {
		t.Fatalf("create run dir: %v", err)
	}

	createCLIExecutionRecord(t, devRoot, record)

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"maintenance",
		"execution",
		"artifacts",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance execution artifacts returned error: %v", err)
	}

	for _, want := range []string{
		"executions: 1",
		"drifted: 1",
		"missing_artifacts: 4",
		"execution: exec-1 ticket: CHC-1 status: succeeded artifact_status: missing_artifacts",
		"artifact: exec-1 name: run_dir exists: true path: " + record.Artifacts.RunDir,
		"artifact: exec-1 name: result exists: false path: " + record.Artifacts.ResultPath,
		"artifact: exec-1 name: report exists: false path: " + core.ExecutionReportPath(record),
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestMaintenanceExecutionArtifactsReportsCompleteArtifacts(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	runDir := filepath.Join(projectDir, ".chc", "user", "runtime", "runs", "exec-1")
	record := cliExecutionRecord(projectDir, runDir)
	writeCLIExecutionArtifacts(t, record)
	createCLIExecutionRecord(t, devRoot, record)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"maintenance",
		"execution",
		"artifacts",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance execution artifacts returned error: %v", err)
	}

	for _, want := range []string{
		"executions: 1",
		"drifted: 0",
		"missing_artifacts: 0",
		"execution: exec-1 ticket: CHC-1 status: succeeded artifact_status: ok",
		"artifact: exec-1 name: validation exists: true path: " + record.Artifacts.ValidationPath,
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestMaintenanceExecutionCleanupPreviewReportsUnreferencedArtifactDirs(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	artifactDir := filepath.Join(projectDir, ".chc", "user", "runtime", "runs")
	referencedDir := filepath.Join(artifactDir, "CHC-1", "exec-1")
	unreferencedDir := filepath.Join(artifactDir, "CHC-2", "exec-2")
	record := cliExecutionRecord(projectDir, referencedDir)

	for _, dir := range []string{referencedDir, unreferencedDir} {
		err := os.MkdirAll(dir, 0o755)
		if err != nil {
			t.Fatalf("create runtime artifact dir %q: %v", dir, err)
		}
	}

	createCLIExecutionRecord(t, devRoot, record)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"maintenance",
		"execution",
		"cleanup-preview",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance execution cleanup-preview returned error: %v", err)
	}

	for _, want := range []string{
		"artifact_dir: " + artifactDir,
		"referenced_run_dirs: 1",
		"cleanup_candidates: 1",
		"mode: preview",
		"candidate: " + unreferencedDir,
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
}

func TestMaintenanceExecutionCleanupPreviewDoesNotDeleteCandidates(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	unreferencedDir := filepath.Join(projectDir, ".chc", "user", "runtime", "runs", "CHC-2", "exec-2")

	err := os.MkdirAll(unreferencedDir, 0o755)
	if err != nil {
		t.Fatalf("create runtime artifact dir: %v", err)
	}

	err = run(t.Context(), []string{
		"maintenance",
		"execution",
		"cleanup-preview",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance execution cleanup-preview returned error: %v", err)
	}

	assertPathExists(t, unreferencedDir)
}

func TestMaintenanceBundleWritesDiagnosticBundle(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()
	bundlePath := filepath.Join(projectDir, ".chc", "maintenance", "diagnostic-bundle.md")

	addProject(t, devRoot, "example", projectDir)
	writeCLIWorkspaceConfig(t, projectDir, `version: 1
notifications:
  telegram:
    bot_token_env: SUPER_SECRET_TOKEN_VALUE
`)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"maintenance",
		"bundle",
		"--dev-root",
		devRoot,
		"--dir",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("maintenance bundle returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "bundle_path: "+bundlePath) {
		t.Fatalf("stdout = %q, want bundle path", stdout.String())
	}

	content := readFile(t, bundlePath)

	for _, want := range []string{
		"# Diagnostic Bundle",
		"project_dir: " + projectDir,
		"database_exists: true",
		"example path=" + projectDir,
		"workspace config contents omitted",
		"environment values omitted",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("bundle content does not contain %q:\n%s", want, content)
		}
	}

	if strings.Contains(content, "SUPER_SECRET_TOKEN_VALUE") {
		t.Fatalf("bundle content contains secret-like config value:\n%s", content)
	}
}

func TestProjectAddRegistersProject(t *testing.T) {
	projectDir := t.TempDir()
	devRoot := t.TempDir()

	err := os.Mkdir(filepath.Join(projectDir, ".chc"), 0o755)
	if err != nil {
		t.Fatalf("create project workspace: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"project",
		"add",
		"--dev-root",
		devRoot,
		"--name",
		"example",
		projectDir,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "project: example "+projectDir) {
		t.Fatalf("stdout = %q, want registered project", stdout.String())
	}

	assertPathExists(t, filepath.Join(devRoot, "data", "chc", "core.db"))
}

func TestProjectAddRejectsProjectWithoutWorkspace(t *testing.T) {
	err := run(t.Context(), []string{
		"project",
		"add",
		"--dev-root",
		t.TempDir(),
		t.TempDir(),
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("run returned nil error")
	}
}

func TestProjectListPrintsRegisteredProjects(t *testing.T) {
	projectDir := t.TempDir()
	devRoot := t.TempDir()

	err := os.Mkdir(filepath.Join(projectDir, ".chc"), 0o755)
	if err != nil {
		t.Fatalf("create project workspace: %v", err)
	}

	err = run(t.Context(), []string{
		"project",
		"add",
		"--dev-root",
		devRoot,
		"--name",
		"example",
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project add returned error: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"project",
		"list",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project list returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "example\t"+projectDir+"\ttrue\tproject-") {
		t.Fatalf("stdout = %q, want listed project", stdout.String())
	}
}

func TestProjectSelectByNameStoresCurrentProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	var selectStdout bytes.Buffer

	err := run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		devRoot,
		"example",
	}, &selectStdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project select returned error: %v", err)
	}

	if !strings.Contains(selectStdout.String(), "project: example "+projectDir) {
		t.Fatalf("stdout = %q, want selected project", selectStdout.String())
	}

	var currentStdout bytes.Buffer

	err = run(t.Context(), []string{
		"project",
		"current",
		"--dev-root",
		devRoot,
	}, &currentStdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project current returned error: %v", err)
	}

	if !strings.Contains(currentStdout.String(), "project: example "+projectDir) {
		t.Fatalf("stdout = %q, want current project", currentStdout.String())
	}

	assertPathExists(t, filepath.Join(devRoot, "config", "chc", "active-project"))
}

func TestProjectSelectByPathStoresCurrentProject(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	err := run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		devRoot,
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project select returned error: %v", err)
	}

	var stdout bytes.Buffer

	err = run(t.Context(), []string{
		"project",
		"current",
		"--dev-root",
		devRoot,
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project current returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "project: example "+projectDir) {
		t.Fatalf("stdout = %q, want current project", stdout.String())
	}
}

func TestProjectCurrentRejectsMissingSelection(t *testing.T) {
	err := run(t.Context(), []string{
		"project",
		"current",
		"--dev-root",
		t.TempDir(),
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("project current returned nil error")
	}
}

func TestProjectSelectRejectsUnknownProject(t *testing.T) {
	err := run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		t.TempDir(),
		"missing",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("project select returned nil error")
	}
}

func TestProjectSelectRejectsMissingRegisteredPath(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	err := os.RemoveAll(projectDir)
	if err != nil {
		t.Fatalf("remove project dir: %v", err)
	}

	err = run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		devRoot,
		"example",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("project select returned nil error")
	}
}

func TestProjectSelectRejectsRegisteredPathWithoutWorkspace(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)

	err := os.RemoveAll(filepath.Join(projectDir, ".chc"))
	if err != nil {
		t.Fatalf("remove project workspace: %v", err)
	}

	err = run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		devRoot,
		"example",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("project select returned nil error")
	}
}

func TestTicketNewRejectsMissingActiveProjectPath(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	selectProject(t, devRoot, "example")

	err := os.RemoveAll(projectDir)
	if err != nil {
		t.Fatalf("remove project dir: %v", err)
	}

	err = run(t.Context(), []string{
		"ticket",
		"new",
		"--dev-root",
		devRoot,
		"Do something",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket new returned nil error")
	}
}

func TestTicketNewRejectsActiveProjectWithoutWorkspace(t *testing.T) {
	projectDir := newTestWorkspace(t)
	devRoot := t.TempDir()

	addProject(t, devRoot, "example", projectDir)
	selectProject(t, devRoot, "example")

	err := os.RemoveAll(filepath.Join(projectDir, ".chc"))
	if err != nil {
		t.Fatalf("remove project workspace: %v", err)
	}

	err = run(t.Context(), []string{
		"ticket",
		"new",
		"--dev-root",
		devRoot,
		"Do something",
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("ticket new returned nil error")
	}
}

func addProject(t *testing.T, devRoot, name, projectDir string) {
	t.Helper()

	err := run(t.Context(), []string{
		"project",
		"add",
		"--dev-root",
		devRoot,
		"--name",
		name,
		projectDir,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project add returned error: %v", err)
	}
}

func selectProject(t *testing.T, devRoot, selector string) {
	t.Helper()

	err := run(t.Context(), []string{
		"project",
		"select",
		"--dev-root",
		devRoot,
		selector,
	}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("project select returned error: %v", err)
	}
}

func createCLIExecutionRecord(t *testing.T, devRoot string, record core.ExecutionRecord) {
	t.Helper()

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := core.NewExecutionRepository(t.Context(), env, testAssetsFS())
	if err != nil {
		t.Fatalf("NewExecutionRepository returned error: %v", err)
	}

	t.Cleanup(func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close execution repository: %v", err)
		}
	})

	err = repository.CreateExecution(t.Context(), record)
	if err != nil {
		t.Fatalf("CreateExecution returned error: %v", err)
	}
}

func cliExecutionRecord(projectDir, runDir string) core.ExecutionRecord {
	now := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	finishedAt := now.Add(time.Minute)

	return core.ExecutionRecord{
		ID:         "exec-1",
		ProjectID:  projectID(projectDir),
		TicketKey:  "CHC-1",
		TicketPath: filepath.Join(projectDir, ".chc", "user", "tickets", "CHC-1.md"),
		Status:     core.ExecutionStatusSucceeded,
		Runtime: core.ExecutionRuntimeRef{
			Image:             "chc-agent-runtime:local",
			DockerfilePath:    filepath.Join(projectDir, ".chc", "user", "runtime", "agent", "Dockerfile"),
			ContainerWorkdir:  "/workspace",
			ValidationCommand: "make verify",
		},
		Artifacts: core.ExecutionArtifacts{
			RunDir:         runDir,
			ResultPath:     filepath.Join(runDir, "result.json"),
			LogPath:        filepath.Join(runDir, "runtime.log"),
			ValidationPath: filepath.Join(runDir, "validation.log"),
		},
		CreatedAt:  now,
		StartedAt:  &now,
		FinishedAt: &finishedAt,
		UpdatedAt:  now.Add(time.Minute),
	}
}

func writeCLIExecutionArtifacts(t *testing.T, record core.ExecutionRecord) {
	t.Helper()

	err := os.MkdirAll(record.Artifacts.RunDir, 0o755)
	if err != nil {
		t.Fatalf("create run dir: %v", err)
	}

	for _, path := range []string{
		record.Artifacts.ResultPath,
		record.Artifacts.LogPath,
		record.Artifacts.ValidationPath,
		core.ExecutionReportPath(record),
	} {
		err = os.WriteFile(path, []byte("artifact\n"), 0o644)
		if err != nil {
			t.Fatalf("write artifact %q: %v", path, err)
		}
	}
}

func writeCLIAgentEventStream(t *testing.T, path string, lines ...string) {
	t.Helper()

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create event stream dir: %v", err)
	}

	err = os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
	if err != nil {
		t.Fatalf("write event stream: %v", err)
	}
}

func marshalCLIAgentEventLine(t *testing.T, event core.AgentEvent) string {
	t.Helper()

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	return string(data)
}

func upsertDaemonState(t *testing.T, devRoot string, state core.DaemonState) {
	t.Helper()

	repository := newDaemonStateRepositoryForTest(t, devRoot)

	err := repository.UpsertDaemonState(t.Context(), state)
	if err != nil {
		t.Fatalf("UpsertDaemonState returned error: %v", err)
	}
}

func getDaemonState(t *testing.T, devRoot string) core.DaemonState {
	t.Helper()

	repository := newDaemonStateRepositoryForTest(t, devRoot)

	state, err := repository.GetDaemonState(t.Context())
	if err != nil {
		t.Fatalf("GetDaemonState returned error: %v", err)
	}

	return state
}

func newDaemonStateRepositoryForTest(t *testing.T, devRoot string) core.DaemonStateRepository {
	t.Helper()

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := core.NewDaemonStateRepository(t.Context(), env, testAssetsFS())
	if err != nil {
		t.Fatalf("NewDaemonStateRepository returned error: %v", err)
	}

	t.Cleanup(func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close daemon state repository: %v", err)
		}
	})

	return repository
}

func indexTicket(t *testing.T, devRoot, projectDir string, ticket core.IndexedTicket) {
	t.Helper()

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	index, err := core.NewTicketIndex(t.Context(), env, testAssetsFS())
	if err != nil {
		t.Fatalf("NewTicketIndex returned error: %v", err)
	}

	defer func() {
		err := index.Close()
		if err != nil {
			t.Fatalf("close ticket index: %v", err)
		}
	}()

	now := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)

	ticket.ProjectID = projectID(projectDir)
	ticket.FileMTime = now
	ticket.FileSize = 100
	ticket.ContentHash = "hash"
	ticket.IndexedAt = now

	err = index.UpsertTicket(t.Context(), ticket)
	if err != nil {
		t.Fatalf("UpsertTicket returned error: %v", err)
	}
}

func listProjectTickets(t *testing.T, devRoot, projectDir string) []core.IndexedTicket {
	t.Helper()

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	index, err := core.NewTicketIndex(t.Context(), env, testAssetsFS())
	if err != nil {
		t.Fatalf("NewTicketIndex returned error: %v", err)
	}

	defer func() {
		err := index.Close()
		if err != nil {
			t.Fatalf("close ticket index: %v", err)
		}
	}()

	tickets, err := index.ListProjectTickets(t.Context(), projectID(projectDir))
	if err != nil {
		t.Fatalf("ListProjectTickets returned error: %v", err)
	}

	return tickets
}

func listTicketExecutions(t *testing.T, devRoot, projectID, ticketKey string) []core.ExecutionRecord {
	t.Helper()

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := core.NewExecutionRepository(t.Context(), env, testAssetsFS())
	if err != nil {
		t.Fatalf("NewExecutionRepository returned error: %v", err)
	}

	defer func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close execution repository: %v", err)
		}
	}()

	records, err := repository.ListTicketExecutions(t.Context(), projectID, ticketKey)
	if err != nil {
		t.Fatalf("ListTicketExecutions returned error: %v", err)
	}

	return records
}

func listProjectLoopRuns(t *testing.T, devRoot, projectID string) []core.LoopRunRecord {
	t.Helper()

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := core.NewLoopRunRepository(t.Context(), env, testAssetsFS())
	if err != nil {
		t.Fatalf("NewLoopRunRepository returned error: %v", err)
	}

	defer func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close loop run repository: %v", err)
		}
	}()

	records, err := repository.ListProjectLoopRuns(t.Context(), projectID)
	if err != nil {
		t.Fatalf("ListProjectLoopRuns returned error: %v", err)
	}

	return records
}

func createApplyResultExecution(
	t *testing.T,
	devRoot string,
	projectDir string,
	status core.ExecutionStatus,
	failureReason string,
) core.ExecutionRecord {
	t.Helper()

	record := testApplyResultExecutionRecord(projectDir, status, failureReason)

	createApplyResultExecutionRecord(t, devRoot, record)

	return record
}

func createApplyResultExecutionRecord(t *testing.T, devRoot string, record core.ExecutionRecord) {
	t.Helper()

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := core.NewExecutionRepository(t.Context(), env, testAssetsFS())
	if err != nil {
		t.Fatalf("NewExecutionRepository returned error: %v", err)
	}

	defer func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close execution repository: %v", err)
		}
	}()

	err = repository.CreateExecution(t.Context(), record)
	if err != nil {
		t.Fatalf("CreateExecution returned error: %v", err)
	}
}

func testApplyResultExecutionRecord(
	projectDir string,
	status core.ExecutionStatus,
	failureReason string,
) core.ExecutionRecord {
	runDir := filepath.Join(projectDir, ".chc", "user", "runtime", "runs", "CHC-1", "exec-apply")
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	return core.ExecutionRecord{
		ID:            "exec-apply",
		ProjectID:     projectID(projectDir),
		TicketKey:     "CHC-1",
		TicketPath:    filepath.Join(projectDir, ".chc", "user", "tickets", "CHC-1-Do-something.md"),
		Status:        status,
		FailureReason: failureReason,
		Runtime: core.ExecutionRuntimeRef{
			Image:             "chc-runtime:latest",
			DockerfilePath:    filepath.Join(projectDir, ".chc", "user", "runtime", "agent", "Dockerfile"),
			ContainerWorkdir:  "/workspace",
			ValidationCommand: "make verify",
		},
		Artifacts: core.ExecutionArtifacts{
			RunDir:         runDir,
			ResultPath:     filepath.Join(runDir, "result.json"),
			LogPath:        filepath.Join(runDir, "runtime.log"),
			ValidationPath: filepath.Join(runDir, "validation.log"),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func writeApplyResultFile(
	t *testing.T,
	record core.ExecutionRecord,
	status core.ExecutionResultStatus,
	summary string,
) {
	t.Helper()

	err := os.MkdirAll(record.Artifacts.RunDir, 0o755)
	if err != nil {
		t.Fatalf("create execution run dir: %v", err)
	}

	content := fmt.Sprintf(
		`{"execution_id":%q,"ticket_key":%q,"status":%q,"summary":%q,"evidence_summary":{"strategy":"planned","produced":"evidence produced","validation":"validation passed"}}`,
		record.ID,
		record.TicketKey,
		status,
		summary,
	)

	err = os.WriteFile(record.Artifacts.ResultPath, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write result file: %v", err)
	}
}

func newApplyResultGitRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	cmd := exec.Command("git", "-C", dir, "init")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git init failed: %v\n%s", err, string(output))
	}

	return dir
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	return string(content)
}

func testTicket() core.Ticket {
	return core.Ticket{
		Key:         "CHC-1",
		Title:       "Do something",
		Status:      "open",
		Priority:    "high",
		Severity:    "normal",
		Type:        "task",
		UseCase:     "onboarding",
		ProjectArea: "uxui",
		CreatedAt:   time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
	}
}

type fakeTicketStartHarnessExecutor struct {
	commands       []string
	err            error
	failWorker     bool
	resultStatuses map[string]core.ExecutionResultStatus
}

func (e *fakeTicketStartHarnessExecutor) Run(
	_ context.Context,
	command core.HarnessCommand,
	stdout io.Writer,
	stderr io.Writer,
) error {
	name := command.Program + " " + command.Args[0]
	e.commands = append(e.commands, name)

	_, err := fmt.Fprintln(stdout, "stdout: "+name)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(stderr, "stderr: "+name)
	if err != nil {
		return err
	}

	if e.failWorker && isFakeTicketStartWorkerCommand(command) {
		return core.NewExecutionFailure(core.ExecutionFailureHarness, fmt.Errorf("codex failed"))
	}

	err = e.writeFakeTicketStartResult(command)
	if err != nil {
		return err
	}

	return e.err
}

func isFakeTicketStartWorkerCommand(command core.HarnessCommand) bool {
	return command.Args[0] == "run" && harnessEnvValue(command, "CHC_RESULT_PATH") != ""
}

func (e *fakeTicketStartHarnessExecutor) writeFakeTicketStartResult(command core.HarnessCommand) error {
	if command.Args[0] != "run" {
		return nil
	}

	resultPath := harnessEnvValue(command, "CHC_RESULT_PATH")
	if resultPath == "" {
		return nil
	}

	status := core.ExecutionResultStatusSucceeded
	ticketKey := harnessEnvValue(command, "CHC_TICKET_KEY")

	configured, ok := e.resultStatuses[ticketKey]
	if ok {
		status = configured
	}

	content := `{"execution_id":"` + harnessEnvValue(command, "CHC_EXECUTION_ID") +
		`","ticket_key":"` + ticketKey +
		`","status":"` + string(status) +
		`","summary":"done","evidence_summary":{"strategy":"planned","produced":"evidence produced","validation":"validation passed"}}`

	return os.WriteFile(containerPathToProjectPath(command, resultPath), []byte(content), 0o644)
}

func harnessEnvValue(command core.HarnessCommand, name string) string {
	prefix := name + "="
	for _, arg := range command.Args {
		if strings.HasPrefix(arg, prefix) {
			return strings.TrimPrefix(arg, prefix)
		}
	}

	return ""
}

func containerPathToProjectPath(command core.HarnessCommand, path string) string {
	for index, arg := range command.Args {
		if arg == "--volume" && index+1 < len(command.Args) {
			parts := strings.SplitN(command.Args[index+1], ":", 2)
			if len(parts) == 2 && strings.HasPrefix(path, parts[1]) {
				return filepath.Join(parts[0], strings.TrimPrefix(path, parts[1]))
			}
		}
	}

	return path
}

func installFakeTicketStartHarnessExecutor(t *testing.T, err error) *fakeTicketStartHarnessExecutor {
	t.Helper()

	executor := &fakeTicketStartHarnessExecutor{err: err}
	previous := newTicketStartHarnessExecutor
	newTicketStartHarnessExecutor = func() core.HarnessExecutor {
		return executor
	}

	t.Cleanup(func() {
		newTicketStartHarnessExecutor = previous
	})

	return executor
}

type fakeTicketStartWorktreePreparer struct {
	results []core.WorktreePrepareResult
}

func (p *fakeTicketStartWorktreePreparer) Prepare(_ context.Context, opts core.WorktreePlanOptions) (core.WorktreePrepareResult, error) {
	plan, err := core.NewWorktreePlan(opts)
	if err != nil {
		return core.WorktreePrepareResult{}, err
	}

	plan.BaseSHA = "base-sha"

	if err := copyTicketStartWorktreeFixture(opts.ProjectDir, plan.WorktreePath); err != nil {
		return core.WorktreePrepareResult{}, err
	}

	result := core.WorktreePrepareResult{
		Plan:          plan,
		CreatedBranch: true,
		AddedWorktree: true,
	}
	p.results = append(p.results, result)

	return result, nil
}

func copyTicketStartWorktreeFixture(projectDir, worktreeDir string) error {
	if err := os.MkdirAll(worktreeDir, 0o755); err != nil {
		return err
	}

	if output, err := exec.Command("git", "-C", worktreeDir, "init").CombinedOutput(); err != nil {
		return fmt.Errorf("git init worktree fixture: %w: %s", err, strings.TrimSpace(string(output)))
	}

	ticketDir := filepath.Join(projectDir, ".chc", "user", "tickets")

	err := filepath.WalkDir(ticketDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			return nil
		}

		relative, err := filepath.Rel(projectDir, path)
		if err != nil {
			return err
		}

		return copyTicketStartWorktreeFixtureFile(projectDir, worktreeDir, relative)
	})
	if err != nil {
		return err
	}

	return copyTicketStartWorktreeFixtureFile(
		projectDir,
		worktreeDir,
		filepath.Join(".chc", "user", "runtime", "agent", "Dockerfile"),
	)
}

func copyTicketStartWorktreeFixtureFile(projectDir, worktreeDir, relativePath string) error {
	source := filepath.Join(projectDir, relativePath)
	target := filepath.Join(worktreeDir, relativePath)

	content, err := os.ReadFile(source)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}

	return os.WriteFile(target, content, 0o644)
}

func installFakeTicketStartWorktreePreparer(t *testing.T) *fakeTicketStartWorktreePreparer {
	t.Helper()

	preparer := &fakeTicketStartWorktreePreparer{}
	previous := newTicketStartWorktreePreparer
	newTicketStartWorktreePreparer = func() ticketStartWorktreePreparer {
		return preparer
	}

	t.Cleanup(func() {
		newTicketStartWorktreePreparer = previous
	})

	return preparer
}

func writeTicketMarkdownOnly(t *testing.T, projectDir, relativePath string, ticket core.Ticket) string {
	t.Helper()

	ticketPath := filepath.Join(projectDir, ".chc", "user", "tickets", filepath.FromSlash(relativePath))

	content, err := core.NewTicketRenderer().RenderTicket(ticket)
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	err = os.MkdirAll(filepath.Dir(ticketPath), 0o755)
	if err != nil {
		t.Fatalf("create ticket dir: %v", err)
	}

	err = os.WriteFile(ticketPath, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write ticket: %v", err)
	}

	return ticketPath
}

func appendTicketEvidenceStrategy(t *testing.T, ticketPath string) {
	t.Helper()

	content, err := os.ReadFile(ticketPath)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	content = append(content, []byte(testTicketEvidenceStrategyBlock())...)

	err = os.WriteFile(ticketPath, content, 0o644)
	if err != nil {
		t.Fatalf("write ticket: %v", err)
	}
}

func testTicketEvidenceStrategyBlock() string {
	return `
<!-- chc:evidence-strategy:v1 -->
- [x] Evidence strategy: complete; change_kind=cli_behavior; risk=medium
- [x] Gherkin: required; path=.chc/user/acceptance/cli.feature
- [x] Tests: mixed; unit and CLI coverage
- [x] Validation: passed; make verify
- [x] Evidence summary: complete; ready for review
<!-- /chc:evidence-strategy -->
`
}

func newTestWorkspace(t *testing.T) string {
	t.Helper()

	projectDir := t.TempDir()

	err := os.Mkdir(filepath.Join(projectDir, ".chc"), 0o755)
	if err != nil {
		t.Fatalf("create project workspace: %v", err)
	}

	return projectDir
}

func writeCLIWorkspaceConfig(t *testing.T, projectDir, content string) {
	t.Helper()

	configPath := filepath.Join(projectDir, ".chc", "user", "config.yml")

	err := os.MkdirAll(filepath.Dir(configPath), 0o755)
	if err != nil {
		t.Fatalf("create workspace config directory: %v", err)
	}

	err = os.WriteFile(configPath, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write workspace config: %v", err)
	}
}

func writeCLIAcceptanceFeature(t *testing.T, projectDir, relativePath string) {
	t.Helper()

	path := filepath.Join(projectDir, relativePath)

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create acceptance feature directory: %v", err)
	}

	err = os.WriteFile(path, []byte("Feature: Onboarding\n"), 0o644)
	if err != nil {
		t.Fatalf("write acceptance feature: %v", err)
	}
}

func writeCLIReport(t *testing.T, projectDir, relativePath, content string) {
	t.Helper()

	path := filepath.Join(projectDir, relativePath)

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create report directory: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write report: %v", err)
	}
}

func authorSessionIDFromOutput(t *testing.T, output string) string {
	t.Helper()

	for _, line := range strings.Split(output, "\n") {
		value, ok := strings.CutPrefix(line, "session: created ")
		if ok {
			return value
		}

		value, ok = strings.CutPrefix(line, "id: ")
		if ok {
			return value
		}

		value, ok = strings.CutPrefix(line, "session: ")
		if ok && (strings.HasPrefix(value, "auth-") || strings.HasPrefix(value, "author-")) {
			return value
		}
	}

	t.Fatalf("output does not contain session id:\n%s", output)

	return ""
}

func createAuthorQuizForTest(t *testing.T, projectDir string) (string, string) {
	t.Helper()

	installFakeAuthoringQuizProvider(t)

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"quiz",
		"new",
		"--dir",
		projectDir,
		"--target",
		"spec",
		"--topic",
		"Billing sync",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author quiz new returned error: %v", err)
	}

	quizPath := authorQuizPathFromOutput(t, stdout.String())
	if !filepath.IsAbs(quizPath) {
		quizPath = filepath.Join(projectDir, quizPath)
	}

	return authorSessionIDFromOutput(t, stdout.String()), quizPath
}

func installFakeAuthoringQuizProvider(t *testing.T) {
	t.Helper()

	previous := newAuthoringQuizQuestionProvider
	newAuthoringQuizQuestionProvider = func() core.AuthoringQuizQuestionProvider {
		return &core.FakeAuthoringQuizQuestionProvider{
			Result: core.AuthoringQuizGenerationResult{
				Provider: "fake",
				Strategy: core.AuthoringQuizGenerationStrategyDeep,
				Questions: []core.AuthoringQuizQuestion{
					{
						ID:                    "Q1",
						Text:                  "Which user workflow should the specification pin down first?",
						Tags:                  []string{"workflow", "specification"},
						FollowUpDepth:         1,
						CompletionRecommended: false,
					},
					{
						ID:                    "Q2",
						Text:                  "What project files or docs should be treated as authoritative context?",
						Tags:                  []string{"context", "documents"},
						FollowUpDepth:         1,
						CompletionRecommended: false,
					},
					{
						ID:                    "Q3",
						Text:                  "What validation evidence should prove the generated spec is useful?",
						Tags:                  []string{"validation", "evidence"},
						FollowUpDepth:         0,
						CompletionRecommended: true,
					},
				},
			},
		}
	}

	t.Cleanup(func() {
		newAuthoringQuizQuestionProvider = previous
	})
}

type authorAssistedQuestionOutput struct {
	TurnID     string
	TurnIndex  int
	QuestionID string
	Question   string
}

func createAuthorAssistedSessionForTest(t *testing.T, projectDir string) string {
	t.Helper()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"assisted",
		"start",
		"--dir",
		projectDir,
		"--target",
		"spec",
		"--topic",
		"Billing sync",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author assisted start returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "assisted: started") {
		t.Fatalf("start output missing assisted marker:\n%s", stdout.String())
	}

	return authorSessionIDFromOutput(t, stdout.String())
}

func askAuthorAssistedQuestionForTest(
	t *testing.T,
	projectDir string,
	sessionID string,
) authorAssistedQuestionOutput {
	t.Helper()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"assisted",
		"resume",
		"--dir",
		projectDir,
		"--session",
		sessionID,
		"--fake-question",
		"What should the billing sync spec include?",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author assisted resume returned error: %v", err)
	}

	return authorAssistedQuestionFromOutput(t, stdout.String())
}

func answerAuthorAssistedQuestionForTest(
	t *testing.T,
	projectDir string,
	sessionID string,
	question authorAssistedQuestionOutput,
) {
	t.Helper()

	var stdout bytes.Buffer

	err := run(t.Context(), []string{
		"author",
		"assisted",
		"answer",
		"--dir",
		projectDir,
		"--session",
		sessionID,
		"--turn",
		question.TurnID,
		"--turn-index",
		fmt.Sprintf("%d", question.TurnIndex),
		"--question-id",
		question.QuestionID,
		"--question",
		question.Question,
		"--answer",
		"It should include retry behavior.",
	}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("author assisted answer returned error: %v", err)
	}

	if !strings.Contains(stdout.String(), "assisted: answer") {
		t.Fatalf("answer output missing assisted marker:\n%s", stdout.String())
	}
}

func authorAssistedQuestionFromOutput(t *testing.T, output string) authorAssistedQuestionOutput {
	t.Helper()

	fields := authorOutputFields(output)

	turnIndex, err := parseAuthorOutputInt(fields["turn_index"])
	if err != nil {
		t.Fatalf("parse turn_index %q: %v", fields["turn_index"], err)
	}

	parsed := authorAssistedQuestionOutput{
		TurnID:     fields["turn"],
		TurnIndex:  turnIndex,
		QuestionID: fields["question_id"],
		Question:   fields["question"],
	}

	if parsed.TurnID == "" || parsed.TurnIndex == 0 || parsed.QuestionID == "" || parsed.Question == "" {
		t.Fatalf("invalid assisted question output:\n%s", output)
	}

	return parsed
}

func authorOutputFields(output string) map[string]string {
	fields := map[string]string{}

	for _, line := range strings.Split(output, "\n") {
		key, value, ok := strings.Cut(line, ": ")
		if ok {
			fields[key] = value
		}
	}

	return fields
}

func parseAuthorOutputInt(value string) (int, error) {
	var parsed int

	_, err := fmt.Sscanf(value, "%d", &parsed)
	if err != nil {
		return 0, err
	}

	return parsed, nil
}

func authorQuizPathFromOutput(t *testing.T, output string) string {
	t.Helper()

	var path string

	for _, line := range strings.Split(output, "\n") {
		value, ok := strings.CutPrefix(line, "path: ")
		if ok && strings.Contains(value, ".chc/user/quizzes/") {
			path = value
		}
	}

	if path != "" {
		return path
	}

	t.Fatalf("output does not contain quiz path:\n%s", output)

	return ""
}

func writeRuntimeTemplate(t *testing.T, projectDir string) {
	t.Helper()

	dockerfilePath := filepath.Join(projectDir, ".chc", "user", "runtime", "agent", "Dockerfile")

	err := os.MkdirAll(filepath.Dir(dockerfilePath), 0o755)
	if err != nil {
		t.Fatalf("create runtime template directory: %v", err)
	}

	err = os.WriteFile(dockerfilePath, []byte("FROM ubuntu:24.04\n"), 0o644)
	if err != nil {
		t.Fatalf("write runtime template: %v", err)
	}

	err = os.MkdirAll(filepath.Join(projectDir, ".chc", "user", "runtime", "runs"), 0o755)
	if err != nil {
		t.Fatalf("create runtime runs directory: %v", err)
	}
}

func listNotificationEventsForTest(
	t *testing.T,
	devRoot string,
	projectID string,
) []core.NotificationAuditRecord {
	t.Helper()

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := core.NewNotificationEventRepository(t.Context(), env, testAssetsFS())
	if err != nil {
		t.Fatalf("NewNotificationEventRepository returned error: %v", err)
	}

	t.Cleanup(func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close notification repository: %v", err)
		}
	})

	records, err := repository.ListProjectNotificationEvents(t.Context(), projectID)
	if err != nil {
		t.Fatalf("ListProjectNotificationEvents returned error: %v", err)
	}

	return records
}

func listRemoteCommandRecordsForTest(t *testing.T, devRoot string) []core.RemoteCommandRecord {
	t.Helper()

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: devRoot,
	})
	if err != nil {
		t.Fatalf("NewUserEnvironment returned error: %v", err)
	}

	repository, err := core.NewRemoteCommandRecordRepository(t.Context(), env, testAssetsFS())
	if err != nil {
		t.Fatalf("NewRemoteCommandRecordRepository returned error: %v", err)
	}

	t.Cleanup(func() {
		err := repository.Close()
		if err != nil {
			t.Fatalf("close remote command repository: %v", err)
		}
	})

	records, err := repository.ListRemoteCommandRecords(t.Context())
	if err != nil {
		t.Fatalf("ListRemoteCommandRecords returned error: %v", err)
	}

	return records
}

func assertNotificationEvent(
	t *testing.T,
	records []core.NotificationAuditRecord,
	index int,
	eventType core.NotificationEventType,
	status core.NotificationStatus,
) {
	t.Helper()

	if len(records) <= index {
		t.Fatalf("notification records = %d, want index %d", len(records), index)
	}

	record := records[index]
	if record.Event.Type != eventType {
		t.Fatalf("notification type = %q, want %q", record.Event.Type, eventType)
	}

	if record.Event.Status != status {
		t.Fatalf("notification status = %q, want %q", record.Event.Status, status)
	}

	if record.Event.ProjectID == "" {
		t.Fatal("notification project id is empty")
	}
}

func chcTestUserEnvironment(t *testing.T) (core.UserEnvironment, error) {
	t.Helper()

	return core.NewUserEnvironment(core.UserEnvironmentOptions{
		DevMode: true,
		DevRoot: t.TempDir(),
	})
}

func runInDirectory(t *testing.T, dir string, runFunc func()) {
	t.Helper()

	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	err = os.Chdir(dir)
	if err != nil {
		t.Fatalf("change working directory: %v", err)
	}

	t.Cleanup(func() {
		err := os.Chdir(previousDir)
		if err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})

	runFunc()
}

func newGateCLIGitProject(t *testing.T) string {
	t.Helper()

	projectDir := t.TempDir()
	writeGateCLIFile(t, projectDir, filepath.Join(core.DirectoryName, ".keep"), "")
	runGateCLIGit(t, projectDir, "init")

	return projectDir
}

func writeGateCLIFile(t *testing.T, projectDir, relativePath, content string) {
	t.Helper()

	path := filepath.Join(projectDir, filepath.FromSlash(relativePath))

	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		t.Fatalf("create gate CLI file dir: %v", err)
	}

	err = os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("write gate CLI file: %v", err)
	}
}

func runGateCLIGit(t *testing.T, projectDir string, args ...string) {
	t.Helper()

	cmdArgs := append([]string{
		"-C", projectDir,
		"-c", "user.name=ChrysoCoord Test",
		"-c", "user.email=chc@example.test",
	}, args...)
	cmd := exec.Command("git", cmdArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
}

func gateCLIGitOutput(t *testing.T, projectDir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", append([]string{"-C", projectDir}, args...)...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}

	return strings.TrimSpace(string(output))
}

func readOnlyGateCLIJSONReport(t *testing.T, projectDir string) core.GateReport {
	t.Helper()

	matches, err := filepath.Glob(filepath.Join(projectDir, ".chc", "user", "reports", "gates", "*.json"))
	if err != nil {
		t.Fatalf("glob gate reports: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("JSON report count = %d, want 1: %#v", len(matches), matches)
	}

	content := readFile(t, matches[0])

	var report core.GateReport

	err = json.Unmarshal([]byte(content), &report)
	if err != nil {
		t.Fatalf("decode gate JSON report: %v", err)
	}

	return report
}

func assertPathExists(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat path %s: %v", path, err)
	}
}

func assertPathNotExists(t *testing.T, path string) {
	t.Helper()

	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		t.Fatalf("path %s exists or returned unexpected stat error: %v", path, err)
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}

	return string(content)
}

type fakeCLIReviewProviderDiagnosticRunner struct {
	paths  map[string]string
	env    map[string]string
	runErr error
}

func (runner fakeCLIReviewProviderDiagnosticRunner) LookPath(file string) (string, error) {
	if path, ok := runner.paths[file]; ok {
		return path, nil
	}

	return "", errors.New("not found")
}

func (runner fakeCLIReviewProviderDiagnosticRunner) LookupEnv(key string) string {
	return runner.env[key]
}

func (runner fakeCLIReviewProviderDiagnosticRunner) Run(_ context.Context, _ string, _ ...string) error {
	return runner.runErr
}

func replaceReviewProviderDiagnosticRunner(runner core.ReviewProviderDiagnosticRunner) func() {
	previous := reviewProviderDiagnosticRunner
	reviewProviderDiagnosticRunner = runner

	return func() {
		reviewProviderDiagnosticRunner = previous
	}
}

type fakeCLIReviewRemoteRunner struct {
	outputs      map[string][]byte
	outputErrs   map[string]error
	inputs       map[string][]byte
	inputErrs    map[string]error
	inputOutputs map[string][]byte
}

func newFakeCLIReviewRemoteRunner() *fakeCLIReviewRemoteRunner {
	return &fakeCLIReviewRemoteRunner{
		outputs:      map[string][]byte{},
		outputErrs:   map[string]error{},
		inputs:       map[string][]byte{},
		inputErrs:    map[string]error{},
		inputOutputs: map[string][]byte{},
	}
}

func (runner *fakeCLIReviewRemoteRunner) Output(_ context.Context, command string, args ...string) ([]byte, error) {
	key := cliReviewRemoteCommandKey(command, args...)
	if err := runner.outputErrs[key]; err != nil {
		return nil, err
	}

	output, ok := runner.outputs[key]
	if !ok {
		return nil, errors.New("missing output for " + key)
	}

	return output, nil
}

func (runner *fakeCLIReviewRemoteRunner) Input(
	_ context.Context,
	command string,
	input []byte,
	args ...string,
) ([]byte, error) {
	key := cliReviewRemoteCommandKey(command, args...)
	if err := runner.inputErrs[key]; err != nil {
		return nil, err
	}

	runner.inputs[key] = input

	if output, ok := runner.inputOutputs[key]; ok {
		return output, nil
	}

	return []byte(`{}`), nil
}

func replaceReviewRemoteCommandRunner(runner core.ReviewRemoteCommandRunner) func() {
	previous := reviewRemoteCommandRunner
	reviewRemoteCommandRunner = runner

	return func() {
		reviewRemoteCommandRunner = previous
	}
}

func cliReviewRemoteCommandKey(command string, args ...string) string {
	return strings.Join(append([]string{command}, args...), " ")
}
