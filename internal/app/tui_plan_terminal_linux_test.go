package app

import (
	"context"
	"maps"
	"os"
	"path/filepath"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	quizmodel "forge.adrianpk.com/chr/internal/core/quiz/model"
	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
)

// Production PTY input opens a generated spec, recovers a plan saved before a
// provider failure, opens its document/tickets and cancels explicit replacement.
func TestTUIPlanTerminal(t *testing.T) {
	provider, _ := webPlanDeliveryFixture(t)
	operations := provider.service(t.Context())
	terminal := startTUITerminal(t, operations, "TestTUIPlanProcess")
	terminal.see(t, "Named work")
	openTerminalGeneratedSpec(t, terminal)
	terminal.keys(t, "\x1b[15~")
	terminal.see(t, "g prepare plan + tickets")
	terminal.keys(t, "g")
	terminal.see(t, "ticket creation is incomplete")
	terminal.see(t, "y Retry the original operation")
	terminal.keys(t, "y")
	terminal.see(t, "Plan ready with 1 tickets in 1 slices. Work has not started.")
	terminal.see(t, "o Open document")
	terminal.see(t, "t View tickets")
	terminal.keys(t, "o")
	terminal.see(t, "Opened generated document.")
	terminal.keys(t, "\x1bo")
	terminal.see(t, "t View tickets")
	terminal.keys(t, "t")
	terminal.see(t, "Implement foundation")
	terminal.see(t, "Prepared tickets. Work has not started.")
	cancelTerminalPlanReplacement(t, terminal, operations)
	terminal.stop(t)
	checkWebPlanDelivery(t, provider)
	checkTUIPlanIdle(t, operations)
}

func openTerminalGeneratedSpec(t *testing.T, terminal *guidedTerminal) {
	t.Helper()
	terminal.keys(t, "\x1bOS")
	terminal.see(t, "g generate spec")
	terminal.keys(t, "g")
	terminal.see(t, "o Open document")
	terminal.keys(t, "o")
	terminal.see(t, "Opened generated document.")
}

func cancelTerminalPlanReplacement(t *testing.T, terminal *guidedTerminal, operations service) {
	t.Helper()

	before := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))
	terminal.keys(t, "\x1b[15~")
	terminal.see(t, "g prepare plan + tickets")
	terminal.keys(t, "G")
	terminal.see(t, "Confirm Plan Action")
	terminal.keys(t, "\x1b")
	terminal.see(t, "Plan action cancelled.")

	if !maps.Equal(before, workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))) {
		t.Fatal("cancelled replacement changed retained artifacts")
	}
}

func checkTUIPlanIdle(t *testing.T, operations service) {
	t.Helper()

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: operations.projectContext.ProjectDir})
	if err != nil {
		t.Fatal(err)
	}

	snapshot, err := workflowcore.NewWorkflowStore(layout, operations.projectContext.ProjectID).Snapshot(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	for _, record := range snapshot.Records {
		if record.Task != nil && (record.Task.Phase != "pending" || record.Task.AttemptID != "") {
			t.Fatal("manual preparation started execution", record.Task)
		}
	}
}

type terminalPlanModel struct {
	model  planDeliveryTestModel
	failed bool
}

func (m *terminalPlanModel) Invoke(ctx context.Context, request quizmodel.Request) (quizmodel.Response, error) {
	if request.Strategy == quizmodel.StrategyPlanDelivery && !m.failed {
		m.failed, m.model.fail = true, true
		defer func() { m.model.fail = false }()
	}

	return m.model.Invoke(ctx, request)
}

// Only provider responses and runtime binding are deterministic. The production
// root loader, terminal, service, receipts, ticket allocator and host are real.
func TestTUIPlanProcess(t *testing.T) {
	project := os.Getenv("CHR_GUIDED_TUI_PROJECT")
	if project == "" {
		t.Skip("launched by the owned plan terminal fixture")
	}

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{DevMode: true, DevRoot: os.Getenv("CHR_GUIDED_TUI_ROOT")})
	if err != nil {
		t.Fatal(err)
	}

	projectContext := commandProjectContext{ProjectDir: project, ProjectID: os.Getenv("CHR_GUIDED_TUI_ID")}
	factory := commandFactory{deps: testAppDependencies(), io: IO{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr}}

	options, err := factory.loadTUIOptions(t.Context(), env, projectContext)
	if err != nil {
		t.Fatal(err)
	}

	generator := &terminalPlanModel{}
	operations := newService(t.Context(), env, projectContext, testAssetsFS())
	operations.quizModel, operations.now, operations.generationBinder = generator, serviceQuizClock(), testGenerationBinder
	operations.conversationModel = sufficientPlanReadinessProvider()

	options.SpecExecutor, options.PlanExecutor = newTUISpecActionExecutor(operations), newTUIPlanActionExecutor(operations)
	if err := runManagedTUI(t.Context(), options, operations); err != nil {
		t.Fatal(err)
	}

	if generator.model.calls != 2 {
		t.Fatal("terminal did not retry its retained preparation exactly once", generator.model.calls)
	}
}
