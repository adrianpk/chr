package app

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/core"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	"forge.adrianpk.com/chr/internal/tui"
	"forge.adrianpk.com/chr/internal/webui"
)

// An independent process saves through the real Web adapter while the TUI is
// open. Its stale keyboard save retains the choice; explicit reload sees the
// externally persisted default without recreating the model or service.
func TestTUIHarnessSelectionRefresh(t *testing.T) {
	operations := newTestService(t, newTestWorkspace(t))
	prepareWebHarnessProfiles(t, operations)

	options := newTUIOptions(tuiOptionsData{projectContext: operations.projectContext, operations: operations})

	var model tea.Model = tui.NewModel(options)

	model = tuiHarnessStep(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	model = tuiHarnessCycle(t, model, 2)
	tuiHarnessOtherProcess(t, operations)

	model = tuiHarnessStep(t, model, tea.KeyMsg{Type: tea.KeyEnter})

	visible := strings.Join(strings.Fields(model.View()), " ")
	if !strings.Contains(visible, "Choices retained") || !strings.Contains(visible, "Worker choice: alpha") || webHarnessSnapshot(t, operations).Worker != "beta" {
		t.Fatal("stale keyboard save lost input or overwrote the Web save", visible)
	}

	model = tuiHarnessStep(t, model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	if !strings.Contains(model.View(), "Worker default: beta") || !strings.Contains(model.View(), "Worker choice: Keep current") {
		t.Fatal("reload did not observe the independent Web change")
	}

	model = tuiHarnessCycle(t, model, 2)
	tuiHarnessStep(t, model, tea.KeyMsg{Type: tea.KeyEsc})

	if webHarnessSnapshot(t, operations).Worker != "beta" {
		t.Fatal("cancelling unsent input published a change")
	}
}

func tuiHarnessOtherProcess(t *testing.T, operations service) {
	t.Helper()

	program, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	command := exec.CommandContext(t.Context(), program, "-test.run=^TestTUIHarnessSelectionProcess$")

	command.Env = append(os.Environ(), "CHR_TUI_HARNESS_PROJECT="+operations.projectContext.ProjectDir, "CHR_TUI_HARNESS_ID="+operations.projectContext.ProjectID, "CHR_TUI_HARNESS_ROOT="+filepath.Dir(filepath.Dir(operations.env.ConfigDir())))

	output, err := command.Output()
	if err != nil {
		t.Fatal("independent Web change failed", err)
	}

	var snapshot harnesscore.SelectionSnapshot
	if err := json.NewDecoder(strings.NewReader(string(output))).Decode(&snapshot); err != nil || snapshot.Worker != "beta" {
		t.Fatal("independent Web process did not save its choice", err)
	}
}

// This opt-in child builds real service/Web adapters with the parent's private
// temporary environment. It does not share a service instance or cached config.
func TestTUIHarnessSelectionProcess(t *testing.T) {
	project := os.Getenv("CHR_TUI_HARNESS_PROJECT")
	if project == "" {
		t.Skip("owned independent process is launched by the refresh test")
	}

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{DevMode: true, DevRoot: os.Getenv("CHR_TUI_HARNESS_ROOT")})
	if err != nil {
		t.Fatal(err)
	}

	operations := newService(t.Context(), env, commandProjectContext{ProjectDir: project, ProjectID: os.Getenv("CHR_TUI_HARNESS_ID")}, testAssetsFS())
	handler, cookies := webHarnessAuthenticated(t, operations)
	form := webHarnessForm(t, handler, cookies)
	form.Set("worker", "beta")
	webHarnessAssertWrite(t, handler, cookies, form, webui.WriteStatusOK)

	if err := json.NewEncoder(os.Stdout).Encode(webHarnessSnapshot(t, operations)); err != nil {
		t.Fatal(err)
	}
}
