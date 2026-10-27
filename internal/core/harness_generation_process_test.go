package core

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
	"forge.adrianpk.com/chr/internal/harness"
)

// A new host process recovers the original operation with no resolver, even
// after the replacement profile and live configuration become unavailable.
func TestHarnessGenerationBindingRestart(t *testing.T) {
	project, operation := generationBindingFixture(t)
	writeHarnessSelectionTestFile(t, workspacecore.UserPath(project, "config.yml"), "harnesses:\n  unknown: invalid\n")

	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestHarnessGenerationBindingHostProcess$", "-test.v")

	cmd.Env = append(os.Environ(), "CHR_GENERATION_FIXTURE_PROJECT="+project)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal("independent host did not retain operation", string(output), err)
	}

	request := generationBindingRequest(project)
	request.ID = "after-restart"

	record := generationBindingReadMetadata(t, request)
	if record.OperationID != operation.ID || record.Selection != operation.Binding || record.ProcessOutcome != "completed" {
		t.Fatal("independent host changed binding", record)
	}
}

// This is the owned independent-host fixture, with no original resolver memory.
func TestHarnessGenerationBindingHostProcess(t *testing.T) {
	project := os.Getenv("CHR_GENERATION_FIXTURE_PROJECT")
	if project == "" {
		return
	}

	operation, err := FreezeHarnessGeneration(t.Context(), harnesshost.HarnessGenerationFreezeInput{Project: project, ID: "operation-fixture", Purpose: "plan_generation"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	request := generationBindingRequest(project)
	request.ID = "after-restart"

	result, err := harnesshost.RunAgent(harnesshost.WithHarnessGeneration(t.Context(), operation), generationBindingProcessRunner{}, request)
	if err != nil || string(result.Output) != `{"ok":true}` {
		t.Fatal("retained driver did not complete", err)
	}
}

type generationBindingProcessRunner struct{}

func (generationBindingProcessRunner) RunAgent(ctx context.Context, request harness.AgentRunRequest) (harness.AgentRunResult, error) {
	cmd := exec.CommandContext(ctx, request.Program)
	cmd.Stdin = bytes.NewBufferString(request.Prompt)
	output, err := cmd.Output()

	return harness.AgentRunResult{Output: output}, err
}
