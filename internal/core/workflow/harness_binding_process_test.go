package workflow

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func runHarnessBindingHost(t *testing.T, store *Store, id string) {
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
