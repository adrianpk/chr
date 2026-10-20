package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

// An owned endpoint and the actual pinned image exercise admission, retained
// selection, container tools, Git validation and host-only completion. No paid
// provider or account is used; opt-in keeps ordinary short tests Docker-free.
func TestPiWorkerContainer(t *testing.T) {
	image := os.Getenv("CHR_TEST_PI_RUNTIME_IMAGE")
	if image == "" {
		t.Skip("set CHR_TEST_PI_RUNTIME_IMAGE to the prepared Pi image")
	}

	bridge := buildPiWorkerFixture(t)
	operations, hierarchy, target := serviceWorkflowTaskFixture(t)
	endpoint := piWorkerEndpoint(t, false)
	configurePiWorkerFixture(t, operations, bridge, image, endpoint.URL)
	assertPiBridgeGenerationRejection(t, operations)

	request := core.WorkflowStartRequest{RequestKey: "pi-worker", Selection: core.WorkflowPreviewRequest{Selector: core.WorkflowSelector{Extent: "ticket", Identity: hierarchy.Tasks[0].Task.Key}, Target: core.WorkflowPreviewTarget{Repository: "https://forge.example/test/repo", Commit: target}}}

	restarted, started := recoverPiWorkerFixture(t, operations, request)

	claim, err := restarted.ClaimWorkflow(t.Context(), started.Invocation.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := claim.Close(); err != nil {
			t.Error(err)
		}
	}()

	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()

	result, err := restarted.ExecuteWorkflowTask(ctx, claim, nil)
	if err != nil || !result.Successful || result.Commit == target || endpoint.calls(t) != 2 {
		t.Fatalf("container worker: successful=%t, failure=%s, calls=%d, error=%v, worker=%+v", result.Successful, result.Failure, endpoint.calls(t), err, result.Worker)
	}
}

func assertPiBridgeGenerationRejection(t *testing.T, operations service) {
	t.Helper()

	profiles, err := core.LoadHarnessProfiles(operations.env.ConfigDir())
	if err != nil {
		t.Fatal(err)
	}

	readiness, _ := piProfileReadiness(t.Context(), profiles["worker"], core.HarnessGeneration, operations.env.ConfigDir())
	if readiness.Ready() || !strings.Contains(readiness.Problem, "native Pi executable") {
		t.Fatal("worker bridge was advertised as a generation executable")
	}
}

func buildPiWorkerFixture(t *testing.T) string {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "chr")
	command := exec.CommandContext(t.Context(), "go", "build", "-o", binary, "../../cmd/chr")

	command.Env = append(os.Environ(), "CGO_ENABLED=0")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build owned bridge: %v: %s", err, output)
	}

	return binary
}

func configurePiWorkerFixture(t *testing.T, operations service, bridge, image, endpoint string) {
	t.Helper()
	configurePiWorkerProfile(t, operations, bridge, image, endpoint, "worker-fixture-only", 1)
}

func configurePiWorkerProfile(t *testing.T, operations service, bridge, image, endpoint, key string, capacity int) {
	t.Helper()

	private := t.TempDir()
	if err := os.Chmod(private, 0o700); err != nil {
		t.Fatal(err)
	}

	auth, models := filepath.Join(private, "auth.json"), filepath.Join(t.TempDir(), "models.json")
	harnessSelectionWrite(t, auth, fmt.Sprintf(`{"chr-fixture":{"type":"api_key","key":%q}}`, key))
	harnessSelectionWrite(t, models, fmt.Sprintf(`{"providers":{"chr-fixture":{"api":"openai-completions","baseUrl":%q,"models":[{"id":"model","contextWindow":32768,"maxTokens":2048}]}}}`, endpoint+"/v1"))
	harnessSelectionWrite(t, filepath.Join(operations.env.ConfigDir(), "harnesses.yml"), fmt.Sprintf("profiles:\n  worker:\n    harness: pi\n    command: %s\n    model_provider: chr-fixture\n    model: model\n    thinking: off\n    credential_ref: %s\n    model_config_ref: %s\n    runtime_ref: %s\n", bridge, auth, models, image))
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, fmt.Sprintf("ticket:\n  key_prefix: PRJ\nwork:\n  max_active_workers: %d\nharnesses:\n  worker: worker\n", capacity))
}

type piWorkerModelFixture struct{ URL, Name string }

func piWorkerEndpoint(t *testing.T, isolation bool) piWorkerModelFixture {
	t.Helper()

	image := os.Getenv("CHR_TEST_PI_RUNTIME_IMAGE")

	fixture, err := filepath.Abs("testdata/pi-worker-model.mjs")
	if err != nil {
		t.Fatal(err)
	}

	name := "chr-pi-model-" + filepath.Base(t.TempDir()) + "-" + fmt.Sprint(time.Now().UnixNano())
	dockerPiFixture(t, "run", "--detach", "--rm", "--pull=never", "--name", name, "--env", fmt.Sprintf("CHR_FIXTURE_ISOLATION=%d", map[bool]int{false: 0, true: 1}[isolation]), "--mount", "type=bind,src="+fixture+",dst=/fixture.mjs,readonly", image, "node", "/fixture.mjs")
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), 10*time.Second)
		defer cancel()

		if output, err := exec.CommandContext(ctx, "docker", "rm", "--force", name).CombinedOutput(); err != nil {
			t.Errorf("stop owned model fixture: %v: %s", err, output)
		}
	})

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	for {
		command := exec.CommandContext(ctx, "docker", "exec", name, "node", "-e", `fetch("http://127.0.0.1:4317/health").then(r=>{if(!r.ok)process.exit(1)}).catch(()=>process.exit(1))`)
		if command.Run() == nil {
			break
		}

		if ctx.Err() != nil {
			t.Fatal("owned model endpoint did not become ready")
		}
	}

	ip := strings.TrimSpace(dockerPiFixture(t, "inspect", "--format", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", name))

	return piWorkerModelFixture{URL: "http://" + ip + ":4317", Name: name}
}

func (fixture piWorkerModelFixture) calls(t *testing.T) int {
	t.Helper()

	logs := dockerPiFixture(t, "logs", "--tail", "16", fixture.Name)
	calls := 0

	for _, line := range strings.Split(strings.TrimSpace(logs), "\n") {
		var record struct {
			Calls   int
			Invalid bool
		}
		if json.Unmarshal([]byte(line), &record) != nil || record.Invalid {
			t.Fatal("model fixture observed invalid private context or tools")
		}

		if record.Calls != 0 {
			calls = record.Calls
		}
	}

	return calls
}

func (fixture piWorkerModelFixture) wait(t *testing.T, workers int) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	for {
		script := `fetch("http://127.0.0.1:4317/health").then(r=>r.json()).then(x=>console.log(x.waiting))`

		output, err := exec.CommandContext(ctx, "docker", "exec", fixture.Name, "node", "-e", script).Output()
		if err == nil && strings.TrimSpace(string(output)) == fmt.Sprint(workers) {
			return
		}

		select {
		case <-ctx.Done():
			t.Fatal("concurrent Pi calls did not reach the endpoint barrier")
		case <-time.After(20 * time.Millisecond):
		}
	}
}

func (fixture piWorkerModelFixture) release(t *testing.T) {
	t.Helper()
	dockerPiFixture(t, "exec", fixture.Name, "node", "-e", `fetch("http://127.0.0.1:4317/release",{method:"POST"}).then(r=>{if(!r.ok)process.exit(1)})`)
}

func dockerPiFixture(t *testing.T, args ...string) string {
	t.Helper()

	command := exec.CommandContext(t.Context(), "docker", args...)

	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("owned Docker fixture: %v: %s", err, output)
	}

	return string(output)
}

func recoverPiWorkerFixture(t *testing.T, operations service, request core.WorkflowStartRequest) (service, core.WorkflowStartResult) {
	t.Helper()

	started, err := operations.StartWorkflow(t.Context(), request)
	if err != nil {
		t.Fatal(err)
	}

	if started.Invocation.HarnessBinding == nil || !strings.HasPrefix(started.Invocation.HarnessBinding.RuntimeRef, "sha256:") {
		t.Fatal("worker did not freeze the immutable runtime")
	}

	// A recovered operation must not need the now-unavailable current profile.
	harnessSelectionWrite(t, filepath.Join(operations.env.ConfigDir(), "harnesses.yml"), "profiles:\n  worker:\n    harness: pi\n    command: /unavailable\n    model_provider: changed\n    model: changed\n")

	restarted := newService(t.Context(), operations.env, operations.projectContext, testAssetsFS())

	replay, err := restarted.StartWorkflow(t.Context(), request)
	if err != nil || !replay.Attached || *replay.Invocation.HarnessBinding != *started.Invocation.HarnessBinding {
		t.Fatal("worker rebound after recovery", err)
	}

	return restarted, started
}
