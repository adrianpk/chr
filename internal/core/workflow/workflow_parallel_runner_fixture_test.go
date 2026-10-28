package workflow

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	opscore "forge.adrianpk.com/chr/internal/core/ops"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type workflowParallelRunnerSpec struct {
	Directory string
	Namespace string
	ID        string
	Started   string
	Release   string
	Result    string
	Mode      string
}

type workflowParallelRunnerResult struct {
	Successful bool
	Capacity   bool
	Fenced     bool
	Failure    string
}

type workflowParallelRunner struct {
	command *exec.Cmd
	spec    workflowParallelRunnerSpec
	output  bytes.Buffer
}

// TestWorkflowParallelRunner is an owned host process. Its actual source worker
// holds inherited reservations and a separate process group behind file barriers.
func TestWorkflowParallelRunner(t *testing.T) {
	data := os.Getenv("CHR_PARALLEL_RUNNER")
	if data == "" {
		return
	}

	var spec workflowParallelRunnerSpec
	if err := json.Unmarshal([]byte(data), &spec); err != nil {
		t.Fatal(err)
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: spec.Directory, Namespace: spec.Namespace})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, "project-1")
	claim, err := workflowParallelRunnerClaim(t, store, spec)

	result := workflowstate.TaskResult{}
	if err == nil {
		result, err = ExecuteTask(t.Context(), claim.store, claim, workflowParallelWorker(spec))
	}

	observed := workflowParallelRunnerResult{Successful: result.Successful, Capacity: errors.Is(err, ErrWorkflowCapacity), Fenced: errors.Is(err, workflowstate.ErrOwned)}
	if err != nil {
		observed.Failure = err.Error()
	}

	encoded, err := json.Marshal(observed)
	if err != nil {
		t.Fatal(err)
	}

	if err := opsfile.WriteAtomic(spec.Result, encoded); err != nil {
		t.Fatal(err)
	}
}

func workflowParallelRunnerClaim(t *testing.T, store *Store, spec workflowParallelRunnerSpec) (*Claim, error) {
	t.Helper()

	if spec.Mode == "native" {
		return store.ClaimNextWorkflowIntegration(t.Context(), IntegrationScheduleQuery{InvocationID: spec.ID, Limit: 1})
	}

	return store.ClaimWorkflow(t.Context(), spec.ID)
}

func workflowParallelWorker(spec workflowParallelRunnerSpec) []string {
	script := `printf started > "$1"; while [ ! -f "$2" ]; do sleep 0.02; done
case "$3" in fail) exit 7;; marker) printf '{}' > ops/default/coordinator.json;; esac
printf result > result.txt
git add result.txt ops/default/coordinator.json
git -c user.name=Test -c user.email=test@example.com commit -qm 'Parallel worker result'`

	return []string{"sh", "-c", script, "worker", spec.Started, spec.Release, spec.Mode}
}

func workflowParallelRunnerFixture(t *testing.T, store *Store, id, mode string) *workflowParallelRunner {
	t.Helper()

	dir := t.TempDir()
	fixture := &workflowParallelRunner{spec: workflowParallelRunnerSpec{Directory: store.layout.ProjectDir(), Namespace: store.layout.Namespace(), ID: id, Mode: mode, Started: filepath.Join(dir, "started"), Release: filepath.Join(dir, "release"), Result: filepath.Join(dir, "result.json")}}

	data, err := json.Marshal(fixture.spec)
	if err != nil {
		t.Fatal(err)
	}

	fixture.command = exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestWorkflowParallelRunner$")

	fixture.command.Env = append(os.Environ(), "CHR_PARALLEL_RUNNER="+string(data))
	fixture.command.Stdout, fixture.command.Stderr = &fixture.output, &fixture.output

	t.Cleanup(func() {
		if fixture.command.Process != nil && fixture.command.ProcessState == nil {
			finishWorkflowParallelRunner(t, fixture)
		}
	})

	if err := fixture.command.Start(); err != nil {
		t.Fatal(err)
	}

	return fixture
}

func finishWorkflowParallelRunner(t *testing.T, fixture *workflowParallelRunner) workflowParallelRunnerResult {
	t.Helper()

	if err := os.WriteFile(fixture.spec.Release, []byte("finish"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := fixture.command.Wait(); err != nil {
		t.Fatalf("owned runner: %v: %s", err, fixture.output.String())
	}

	data, err := os.ReadFile(fixture.spec.Result)
	if err != nil {
		t.Fatal(err)
	}

	var result workflowParallelRunnerResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	return result
}

func workflowParallelRunnerLanes(t *testing.T, capacity int) ([3]*Store, [3]workflowstate.Invocation) {
	t.Helper()

	store, first, request := workflowTicketFixture(t, "ticket", workflowPreparationGate("test -f result.txt"))
	other := workflowParallelPublicationStore(t, store)
	second := workflowLooseCreate(t, other, first, "PRJ-91", "normal", nil)
	third := workflowLooseCreate(t, store, first, "PRJ-92", "normal", nil)
	workflowSelectionGit(t, store, "add", "ops/other")

	request.Selection.Target.Commit, _ = workflowSelectionCommit(t, store)
	writeWorkspaceConfig(t, store.layout.ProjectDir(), "work:\n  max_active_workers: "+strconv.Itoa(capacity)+"\n")

	linked := workflowCapacityLinked(t, store, capacity)
	linked.projectID = store.projectID
	stores := [3]*Store{store, linked, store}
	runs := [3]workflowstate.Invocation{}

	for i, record := range [3]workflowstate.Record{first, second, third} {
		request.RequestKey, request.Selection.Selector.Identity = record.Task.Key, record.Task.Key
		runs[i] = workflowTestStart(t, stores[i], request).Invocation
	}

	return stores, runs
}
