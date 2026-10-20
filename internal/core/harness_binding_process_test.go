package core

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// This owned host subprocess resumes and retries a serialized v2 invocation
// without the original service or profile resolver in memory.
func TestHarnessBindingHostProcess(t *testing.T) {
	project := os.Getenv("CHR_BINDING_TEST_PROJECT")
	if project == "" {
		return
	}

	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: project})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, os.Getenv("CHR_BINDING_TEST_PROJECT_ID"))
	id := os.Getenv("CHR_BINDING_TEST_INVOCATION")

	run, err := store.ResumeWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	retry := WorkflowStartRequest{RequestKey: run.RequestKey, Selection: run.Selection, Policy: run.Policy, ReviewRestriction: run.ReviewRestriction, Admission: run.Admission, HarnessBinding: run.HarnessBinding}

	result, err := store.StartWorkflow(t.Context(), retry)
	if err != nil || !result.Attached || result.Invocation.HarnessBinding == nil {
		t.Fatal("resumed process did not retain its binding", err)
	}

	command, err := store.WorkflowHarnessWorker(t.Context(), id)
	if err != nil || len(command) == 0 {
		t.Fatal("resumed process lost the executable", err)
	}
}

func runHarnessBindingHost(t *testing.T, store *WorkflowStore, id string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestHarnessBindingHostProcess$")

	command.Env = append(os.Environ(), "CHR_BINDING_TEST_PROJECT="+store.layout.ProjectDir(), "CHR_BINDING_TEST_PROJECT_ID="+store.projectID, "CHR_BINDING_TEST_INVOCATION="+id)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatal("independent host recovery failed", string(output), err)
	}
}

// The fixed v1-only consumer keeps the old schema gate ahead of its worker
// dispatch. A v2 file must fail there, before its worker dispatch effect.
func TestHarnessBindingOldReader(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, workflowTaskCommand(t, "gate"), 30)
	writeWorkspaceConfig(t, store.layout.ProjectDir(), "harnesses:\n  worker: alpha\n")

	request.ResolveHarness = harnessBindingTestResolver(t, workflowTaskCommand(t, "worker"))
	started := workflowTestStart(t, store, request)
	marker := filepath.Join(t.TempDir(), "spawned")
	command := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestHarnessBindingOldReaderProcess$")

	command.Env = append(os.Environ(), "CHR_BINDING_V1_FILE="+filepath.Join(store.layout.Root(), workflowInvocationPath(started.Invocation.ID)), "CHR_BINDING_V1_MARKER="+marker)

	output, err := command.CombinedOutput()
	if err == nil || !bytes.Contains(output, []byte("unsupported workflow invocation schema")) {
		t.Fatal("v1-only consumer accepted mandatory v2", string(output), err)
	}

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("v1-only consumer started its worker", err)
	}
}

// This fixture deliberately retains the original v1 schema guard rather than
// using the current multi-version validator. The marker is its dispatch effect.
func TestHarnessBindingOldReaderProcess(t *testing.T) {
	path := os.Getenv("CHR_BINDING_V1_FILE")
	if path == "" {
		return
	}

	data, err := readWorkflowFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var header struct {
		Schema string `json:"schema"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		t.Fatal(err)
	}

	if header.Schema != "chr-invocation/v1" {
		t.Fatal("unsupported workflow invocation schema")
	}

	if err := os.WriteFile(os.Getenv("CHR_BINDING_V1_MARKER"), []byte("started"), 0o600); err != nil {
		t.Fatal(err)
	}
}
