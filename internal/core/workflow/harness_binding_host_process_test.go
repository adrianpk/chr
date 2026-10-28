package workflow

import (
	"encoding/json"
	"os"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// This owned host subprocess resumes and retries a serialized v2 invocation
// without the original service or profile resolver in memory.
func TestHarnessBindingHostProcess(t *testing.T) {
	project := os.Getenv("CHR_BINDING_TEST_PROJECT")
	if project == "" {
		return
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: project})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, os.Getenv("CHR_BINDING_TEST_PROJECT_ID"))
	id := os.Getenv("CHR_BINDING_TEST_INVOCATION")

	run, err := store.ResumeWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	retry := StartRequest{StartInput: workflowstate.StartInput{
		RequestKey:        run.RequestKey,
		Selection:         run.Selection,
		Policy:            run.Policy,
		ReviewRestriction: run.ReviewRestriction,
		Admission:         run.Admission,
		HarnessBinding:    run.HarnessBinding,
	}}

	result, err := store.StartWorkflow(t.Context(), retry)
	if err != nil || !result.Attached || result.Invocation.HarnessBinding == nil {
		t.Fatal("resumed process did not retain its binding", err)
	}

	command, err := store.WorkflowHarnessWorker(t.Context(), id)
	if err != nil || len(command) == 0 {
		t.Fatal("resumed process lost the executable", err)
	}
}

// This fixture deliberately retains the original v1 schema guard rather than
// using the current multi-version validator. The marker is its dispatch effect.
func TestHarnessBindingOldReaderProcess(t *testing.T) {
	path := os.Getenv("CHR_BINDING_V1_FILE")
	if path == "" {
		return
	}

	data, err := opsfile.Read(path)
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
