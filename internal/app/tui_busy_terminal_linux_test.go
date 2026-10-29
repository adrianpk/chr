package app

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
)

// A real busy Bubble Tea process must exit on Ctrl+C, release its provider call
// and leave an independent retained workflow's state and attempts unchanged.
func TestTUIBusyTerminal(t *testing.T) {
	operations, hierarchy, _ := workflowCLIFixture(t)
	id := workflowCLIManualStart(t, operations, hierarchy)
	before, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	terminal := startTUITerminal(t, operations, "TestTUIBusyProcess")
	terminal.keys(t, "\x1b[15~")
	terminal.see(t, "No use case selected.")
	terminal.keys(t, ":")
	terminal.see(t, "Command Palette")
	terminal.keys(t, "generate-spec")
	terminal.see(t, "Generate spec")
	terminal.keys(t, "\r")
	terminal.see(t, "Confirm Spec Action")
	terminal.keys(t, "\r")
	terminal.see(t, "BUSY_PROVIDER_STARTED")
	terminal.stop(t)

	after, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("UI shutdown changed retained workflow state", err)
	}
}

// Only the provider response is delayed. The process uses production adapters,
// the managed host, SQLite/filesystem readers and actual terminal input.
func TestTUIBusyProcess(t *testing.T) {
	project := os.Getenv("CHR_GUIDED_TUI_PROJECT")
	if project == "" {
		t.Skip("launched by the owned busy terminal fixture")
	}

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{DevMode: true, DevRoot: os.Getenv("CHR_GUIDED_TUI_ROOT")})
	if err != nil {
		t.Fatal(err)
	}

	operations := newService(t.Context(), env, commandProjectContext{ProjectDir: project, ProjectID: os.Getenv("CHR_GUIDED_TUI_ID")}, testAssetsFS())
	provider := &busyTerminalProvider{started: make(chan struct{}), done: make(chan struct{})}
	operations.quizModel = provider
	operations.generationBinder = testGenerationBinder
	options := newTUIOptions(tuiOptionsData{ctx: t.Context(), env: env, projectContext: operations.projectContext, stdin: os.Stdin, stdout: os.Stdout, assetsFS: testAssetsFS(), operations: operations})
	if err := runManagedTUI(t.Context(), options, operations); err != nil {
		t.Fatal(err)
	}

	select {
	case <-provider.started:
	default:
		t.Fatal("provider call never started")
	}

	select {
	case <-provider.done:
	case <-time.After(2 * time.Second):
		t.Fatal("provider call outlived managed UI shutdown")
	}
}

type busyTerminalProvider struct {
	started chan struct{}
	done    chan struct{}
}

func (p *busyTerminalProvider) Invoke(ctx context.Context, _ quizmodel.Request) (quizmodel.Response, error) {
	defer close(p.done)
	close(p.started)

	if _, err := fmt.Fprintln(os.Stdout, "BUSY_PROVIDER_STARTED"); err != nil {
		return quizmodel.Response{}, err
	}

	<-ctx.Done()

	return quizmodel.Response{}, ctx.Err()
}
