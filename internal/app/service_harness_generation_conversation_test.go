package app

import (
	"errors"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	projectcore "forge.adrianpk.com/chr/internal/core/project"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// The host retains an intent's binding outside the selected project's mutable
// routing. Switching projects or roles cannot rebind an interrupted request.
func TestHarnessGenerationBindingConversation(t *testing.T) {
	generationBindingConversationCase(t, harnessGenerationFixture)
}

func generationBindingConversationCase(t *testing.T, fixture func(*testing.T, string) (service, string)) {
	output := `{"operation_id":"status","arguments":{"project":"","conversation":"","namespace":"","text":"","action":"","record_json":""},"unresolved":[],"explanation":"Read the current status."}`
	projectService, _ := fixture(t, output)
	project := projectService.projectContext.ProjectDir
	second := registerGenerationProjects(t, projectService)
	global := newService(t.Context(), projectService.env, commandProjectContext{}, testAssetsFS())

	input := ConversationIntentContext{Message: "Read current status", SelectedProject: "fixture", Projects: []ConversationIntentProject{{ID: "fixture", Name: "First"}, {ID: "second", Name: "Second"}}, Operations: []ConversationIntentOperation{{ID: "status", Description: "Read status", Arguments: "No arguments"}}}
	if _, err := global.ResolveConversationIntent(t.Context(), "source", input); err != nil {
		t.Fatal("selected project could not resolve conversation intent", err)
	}

	switchGenerationFixture(t, projectService, "other")

	input.SelectedProject = "second"

	restarted := newService(t.Context(), projectService.env, commandProjectContext{}, testAssetsFS())
	if _, err := restarted.ResolveConversationIntent(t.Context(), "source", input); err != nil {
		t.Fatal("host retry changed its retained project/profile", err)
	}

	if _, err := restarted.ResolveConversationIntent(t.Context(), "independent", input); err != nil {
		t.Fatal("independent conversation missed the newly selected project", err)
	}

	checkGenerationFixtureModels(t, project, map[string]string{"intent-source": "fixture-model"})
	checkGenerationFixtureModels(t, second, map[string]string{"intent-independent": "other-model"})

	input.Projects = input.Projects[1:]
	if _, err := restarted.ResolveConversationIntent(t.Context(), "source", input); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("retained generation bypassed current allowed projects", err)
	}
}

func registerGenerationProjects(t *testing.T, operations service) string {
	t.Helper()

	registry, err := projectcore.NewRegistry(t.Context(), operations.env, testAssetsFS())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := registry.Close(); err != nil {
			t.Error(err)
		}
	})

	second := t.TempDir()
	for _, project := range []projectcore.Project{
		{ID: "fixture", Name: "First", Path: operations.projectContext.ProjectDir, Enabled: true, RegisteredAt: time.Now().UTC()},
		{ID: "second", Name: "Second", Path: second, Enabled: true, RegisteredAt: time.Now().UTC()},
	} {
		if err := registry.AddProject(t.Context(), project); err != nil {
			t.Fatal(err)
		}
	}

	harnessSelectionWrite(t, core.UserPath(second, "config.yml"), "harnesses:\n  generation: other\n")

	return second
}
