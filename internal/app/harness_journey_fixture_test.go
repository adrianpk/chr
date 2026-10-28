package app

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	openaiquiz "forge.adrianpk.com/chr/internal/core/quiz/openai"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type harnessJourney struct {
	operations       service
	telegram         *telegramConversationAppFixture
	record           workflowstate.Record
	start            workflowcore.StartRequest
	code, pi         string
	worker, auth     string
	bridge, image    string
	barrier, inspect string
	apiCalls         *atomic.Int64
	active           context.Context
}

func newHarnessJourney(t *testing.T) *harnessJourney {
	t.Helper()

	telegram, operations, record, start, _ := telegramHarnessWorkerFixture(t)
	workerProfiles := prepareServiceHarnessBinding(t, operations)

	profiles, err := harnesscore.LoadProfiles(filepath.Dir(workerProfiles))
	if err != nil {
		t.Fatal(err)
	}

	_, code := harnessGenerationFixture(t, `{"ok":true}`)
	_, pi := piGenerationFixture(t, `{"ok":true}`)

	f := &harnessJourney{operations: operations, telegram: telegram, record: record, start: start, code: code, pi: pi, worker: profiles["writing"].Command, auth: filepath.Join(operations.env.ConfigDir(), "auth.json"), image: "sha256:" + strings.Repeat("a", 64), barrier: t.TempDir()}
	if err := projectcore.NewActiveStore(operations.env).SetActiveProject(projectcore.Active{ProjectID: operations.projectContext.ProjectID}); err != nil {
		t.Fatal(err)
	}

	harnessSelectionWrite(t, f.auth, `{"fixture":{"type":"api_key","key":"fixture-only-key"}}`)

	f.operations.generationBinder = service.bindHarnessGeneration
	f.configureAPI(t)
	f.prepareInspector(t)
	f.writeProfiles(t, f.blockedDriver(t), false)
	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "ticket:\n  key_prefix: PRJ\nunknown_extension:\n  retained: yes\nharnesses:\n  worker: code-worker\n  generation: code-gen\n")

	return f
}

func (f *harnessJourney) writeProfiles(t *testing.T, generation string, includePi bool) {
	t.Helper()

	model := "fixture-model"
	if includePi {
		model = "replacement-model"
	}

	text := fmt.Sprintf("profiles:\n  code-worker:\n    harness: codex\n    command: %s\n    model: fixture-model\n  code-gen:\n    harness: codex\n    command: %s\n    model: %s\n", f.worker, generation, model)
	if includePi {
		text += fmt.Sprintf("  pi-gen:\n    harness: pi\n    command: %s\n    model_provider: fixture\n    model: fixture-model\n    credential_ref: %s\n  pi-worker:\n    harness: pi\n    command: %s\n    model_provider: fixture\n    model: fixture-model\n    credential_ref: %s\n    runtime_ref: fixture-runtime\n", filepath.Join(f.pi, "driver"), f.auth, f.bridge, f.auth)
	}

	harnessSelectionWrite(t, filepath.Join(f.operations.env.ConfigDir(), "harnesses.yml"), text)
}

func (f *harnessJourney) prepareInspector(t *testing.T) {
	t.Helper()

	directory := t.TempDir()
	f.inspect = filepath.Join(directory, "inspected")
	metadata := fmt.Sprintf(`{"id":%q,"os":"linux","architecture":%q,"labels":{"org.chr.harness":"pi","org.chr.pi.version":"1.3.0","org.chr.pi.protocol":"pi-rpc/1.3.0","org.chr.node.version":"22.23.3"}}`, f.image, runtime.GOARCH)
	script := "#!/bin/sh\n[ \"$1 $2\" = 'image inspect' ] || exit 91\n[ \"$5\" = fixture-runtime ] || exit 92\nprintf 'inspect\\n' >> '" + f.inspect + "'\nprintf '%s\\n' '" + metadata + "'\n"
	program := filepath.Join(directory, "docker")
	harnessSelectionWrite(t, program, script)

	if err := os.Chmod(program, 0o700); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", directory+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func (f *harnessJourney) enablePi(t *testing.T) {
	t.Helper()

	if _, err := os.Stat(f.inspect); !os.IsNotExist(err) {
		t.Fatal("Codex-only setup probed Docker or Pi")
	}

	f.bridge = buildPiWorkerFixture(t)
	f.writeProfiles(t, filepath.Join(f.code, "driver"), true)
}

func (f *harnessJourney) configureAPI(t *testing.T) {
	t.Helper()

	f.apiCalls = &atomic.Int64{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		f.apiCalls.Add(1)

		if _, err := fmt.Fprint(w, `{"output_text":"{\"ok\":true}"}`); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)

	provider := core.NewConfiguredStructuredModel(false)
	provider.API = openaiquiz.Adapter{Host: harnesshost.QuizGeneration{}, APIKey: "fixture-only", BaseURL: server.URL, Client: server.Client()}
	f.operations.conversationModel = provider
}

func (f *harnessJourney) blockedDriver(t *testing.T) string {
	t.Helper()

	program := filepath.Join(f.barrier, "driver")
	text := "#!/bin/sh\ntouch '" + filepath.Join(f.barrier, "started") + "'\nwhile [ ! -f '" + filepath.Join(f.barrier, "release") + "' ]; do sleep 0.01; done\nexec '" + filepath.Join(f.code, "driver") + "' \"$@\"\n"
	harnessSelectionWrite(t, program, text)

	if err := os.Chmod(program, 0o700); err != nil {
		t.Fatal(err)
	}

	return program
}

type harnessJourneyActive struct {
	release  string
	result   <-chan error
	finished <-chan struct{}
}

func (f *harnessJourney) startGeneration(t *testing.T) harnessJourneyActive {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)

	bound, err := f.operations.bindGeneration(ctx, "original-generation", "spec_generation")
	ctx = bound.context(ctx)

	if err != nil {
		cancel()
		t.Fatal(err)
	}

	f.active = ctx

	result, finished := make(chan error, 1), make(chan struct{})
	go func() {
		defer close(finished)

		result <- harnessJourneyInvoke(ctx, bound.service, "original-call")
	}()

	t.Cleanup(func() {
		cancel()

		select {
		case <-finished:
		case <-time.After(5 * time.Second):
			t.Error("owned generation process did not finish")
		}
	})
	harnessJourneyWait(t, filepath.Join(f.barrier, "started"))

	return harnessJourneyActive{filepath.Join(f.barrier, "release"), result, finished}
}

func harnessJourneyWait(t *testing.T, path string) {
	t.Helper()

	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()

	for {
		if _, err := os.Stat(path); err == nil {
			return
		}

		select {
		case <-deadline.C:
			t.Fatal("owned generation did not reach its barrier")
		case <-time.After(5 * time.Millisecond):
		}
	}
}

func (a harnessJourneyActive) finish(t *testing.T) {
	t.Helper()
	harnessSelectionWrite(t, a.release, "release")

	select {
	case err := <-a.result:
		if err != nil {
			t.Fatal("active generation changed during role switches", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("owned generation did not complete after release")
	}
}

func harnessJourneyRequest(operations service, id string) core.StructuredModelRequest {
	return core.StructuredModelRequest{ID: id, Purpose: "spec_generation", ProjectDir: operations.projectContext.ProjectDir, Instructions: "Return the supplied acceptance result as JSON.", Input: []byte(`{"ok":true}`), Schema: map[string]any{"type": "object", "required": []string{"ok"}, "properties": map[string]any{"ok": map[string]any{"type": "boolean"}}, "additionalProperties": false}}
}

func harnessJourneyInvoke(ctx context.Context, operations service, id string) error {
	request := harnessJourneyRequest(operations, id)

	output, err := operations.conversationModel.InvokeStructuredModel(ctx, request)
	if err != nil {
		return err
	}

	if string(output) != `{"ok":true}` {
		return core.ValidateStructuredModelOutput(ctx, request, fmt.Errorf("unexpected structured result"))
	}

	return core.ValidateStructuredModelOutput(ctx, request, nil)
}
