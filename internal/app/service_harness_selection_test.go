package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core"
	harnesscore "forge.adrianpk.com/chr/internal/core/harness"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// Explicit workspace references outrank legacy dispatch and refresh per request.
func TestHarnessSelectionResolution(t *testing.T) {
	operations, project, configDir, command := harnessSelectionFixture(t)
	t.Setenv("CHR_SPEC_GENERATION_BACKEND", "openai")

	view, err := operations.ResolveHarnessSelection(t.Context(), harnesscore.Generation, "spec_generation")
	assertHarnessSelection(t, view, err, "openai", "legacy", "", "cli_default")

	harnessSelectionWrite(t, workspacecore.UserPath(project, "config.yml"), "harnesses:\n  generation: writing\n")

	view, err = operations.ResolveHarnessSelection(t.Context(), harnesscore.Generation, "spec_generation")
	assertHarnessSelection(t, view, err, "codex", "project", "writing", "fixture-model")

	if !view.Readiness.Ready() {
		t.Fatal("prepared profile reported unavailable")
	}

	data, err := json.Marshal(view)
	if err != nil || strings.Contains(string(data), configDir) || strings.Contains(string(data), command) {
		t.Fatal("readable selection exposed private paths")
	}

	prior := view.SelectionRevision

	harnessSelectionWrite(t, workspacecore.UserPath(project, "config.yml"), "harnesses:\n  generation: null\n")

	view, err = operations.ResolveHarnessSelection(t.Context(), harnesscore.Generation, "spec_generation")
	assertHarnessSelection(t, view, err, "openai", "legacy", "", "cli_default")

	if view.SelectionRevision == prior {
		t.Fatalf("fresh legacy selection = %#v, %v", view, err)
	}

	if _, err := os.Stat(filepath.Join(configDir, "spawned")); !os.IsNotExist(err) {
		t.Fatal("readiness executed the model driver")
	}
}

// Missing or unsupported explicit preparation never falls through to a default.
func TestHarnessSelectionRejections(t *testing.T) {
	for _, name := range []string{"missing", "pi-writing", "missing-executable"} {
		t.Run(name, func(t *testing.T) {
			operations, project, _, _ := harnessSelectionFixture(t)
			harnessSelectionWrite(t, workspacecore.UserPath(project, "config.yml"), "harnesses:\n  generation: "+name+"\n")

			view, err := operations.ResolveHarnessSelection(t.Context(), harnesscore.Generation, "spec_generation")
			if err == nil || view.Readiness.Ready() {
				t.Fatalf("unprepared explicit selection = %#v, %v", view, err)
			}
		})
	}
}

// Preserve each existing action default while rejecting unknown explicit values.
func TestHarnessSelectionLegacyDispatch(t *testing.T) {
	for _, test := range []struct {
		name, action, variable, value, backend string
		invalid                                bool
	}{
		{"spec default", "spec_generation", "CHR_SPEC_GENERATION_BACKEND", "", "codex", false},
		{"spec API", "spec_generation", "CHR_SPEC_GENERATION_BACKEND", "openai", "openai", false},
		{"question default", "quiz_next_question", "CHR_SPEC_GENERATION_BACKEND", "codex", "openai", false},
		{"contextual Quiz", "contextual_quiz", "CHR_CODEX_COMMAND", "unused", "codex", false},
		{"invalid backend", "spec_generation", "CHR_SPEC_GENERATION_BACKEND", "other", "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(test.variable, test.value)

			backend, err := legacyHarnessBackend(harnesscore.Generation, test.action)
			if backend != test.backend || (err != nil) != test.invalid {
				t.Fatalf("backend = %q, %v", backend, err)
			}
		})
	}
}

func harnessSelectionFixture(t *testing.T) (service, string, string, string) {
	t.Helper()

	project, configDir := t.TempDir(), t.TempDir()

	env, err := core.NewUserEnvironment(core.UserEnvironmentOptions{HomeDir: t.TempDir(), ConfigDir: configDir})
	if err != nil {
		t.Fatal(err)
	}

	command := filepath.Join(configDir, "fixture-driver")
	harnessSelectionWrite(t, command, "#!/bin/sh\ntouch '"+filepath.Join(configDir, "spawned")+"'\n")

	if err := os.Chmod(command, 0o700); err != nil {
		t.Fatal(err)
	}

	body := "profiles:\n  writing:\n    harness: codex\n    command: " + command + "\n    model: fixture-model\n" +
		"  pi-writing:\n    harness: pi\n    command: " + command + "\n    model_provider: example\n    model: exact-model\n" +
		"  missing-executable:\n    harness: codex\n    command: /missing/fixture-driver\n"
	harnessSelectionWrite(t, filepath.Join(configDir, "harnesses.yml"), body)

	return newService(t.Context(), env, commandProjectContext{ProjectDir: project, ProjectID: "fixture"}, nil), project, configDir, command
}

func harnessSelectionWrite(t *testing.T, path, body string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertHarnessSelection(t *testing.T, view harnesscore.HarnessSelection, err error, backend, source, profile, model string) {
	t.Helper()

	if err != nil || view.Backend != backend || view.Source != source || view.Profile != profile || view.Model != model {
		t.Fatalf("selection = %#v, %v", view, err)
	}
}
