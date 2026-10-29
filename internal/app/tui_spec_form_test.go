package app

import (
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	"forge.adrianpk.com/chr/internal/tui"
)

// Complete forms use canonical mutation/read owners without generation. Errors
// retain entered body, editing keeps scope immutable, and Cancel writes nothing.
func TestTUISpecForm(t *testing.T) {
	for _, slug := range []string{"", "named-work"} {
		t.Run("scope "+slug, func(t *testing.T) {
			provider, generator := newWebQuizFixture(t)
			operations := provider.service(t.Context())
			model := specFormModel(t, operations)
			createTUISpecForm(t, model, slug)

			documents, err := operations.ListSpecDocuments()
			checkSpecDocumentCount(t, documents, err, 1)

			if documents[0].Title != "Manual contract" || documents[0].Body != "SPEC_FORM_BODY" || documents[0].Status != "approved" || documents[0].CaseSlug != slug || len(generator.requests) != 0 {
				t.Fatal(documents, generator.requests)
			}

			editTUISpecForm(t, operations, model, documents[0])
			cancelTUISpecForm(t, operations, model)
		})
	}
}

// A committed Create whose catalog projection fails must retry its receipt,
// rather than producing a second document after the source is repaired.
func TestTUISpecRetry(t *testing.T) {
	provider, _ := newWebQuizFixture(t)
	operations := provider.service(t.Context())

	seed, err := operations.CreateSpec(CreateSpecRequest{OperationID: "retry-seed", Title: "Retained seed", Body: "Seed body.", Status: "draft"})
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(operations.projectContext.ProjectDir, filepath.Dir(seed.Reference), "metadata.json")

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte("invalid metadata"), 0o644); err != nil {
		t.Fatal(err)
	}

	model := specFormModel(t, operations)
	createTUISpecForm(t, model, "named-work")

	if !strings.Contains(model.View(), "Spec saved; catalog refresh failed") {
		t.Fatal("committed partial result was hidden", model.View())
	}

	if err := os.WriteFile(path, before, 0o644); err != nil {
		t.Fatal(err)
	}

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})

	documents, err := operations.ListSpecDocuments()
	checkSpecDocumentCount(t, documents, err, 2)

	if strings.Contains(model.View(), "Create Spec") || !strings.Contains(model.View(), "Spec saved.") {
		t.Fatal("unchanged retry did not reconcile its committed result", model.View())
	}
}

func specFormModel(t *testing.T, operations service) *tui.Model {
	t.Helper()

	cases, err := authoringcore.NewCaseMetadataService(operations.projectContext.ProjectDir).LoadAll()
	if err != nil {
		t.Fatal(err)
	}

	model := tui.NewModel(tui.Options{Cases: cases, SpecExecutor: newTUISpecActionExecutor(operations)})
	model.Update(tea.WindowSizeMsg{Width: 120, Height: 44})
	runGuidedTUIKey(t, model, "f4", false)

	return model
}

func createTUISpecForm(t *testing.T, model *tui.Model, slug string) {
	t.Helper()
	runGuidedTUIKey(t, model, "c", false)

	if slug == "" {
		model.Update(tea.KeyMsg{Type: tea.KeyLeft})
	}

	model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("SPEC_FORM_BODY")})
	model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	if !strings.Contains(model.View(), "validation_failed") {
		t.Fatal("empty title was accepted", model.View())
	}

	model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Manual contract")})
	model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model.Update(tea.KeyMsg{Type: tea.KeyRight})
	model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
}

func editTUISpecForm(t *testing.T, operations service, model *tui.Model, before authoringcore.SpecDocument) {
	t.Helper()
	runGuidedTUIKey(t, model, "e", false)
	model.Update(tea.KeyMsg{Type: tea.KeyRight})
	model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model.Update(tea.KeyMsg{Type: tea.KeyEnd})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" edited")})
	model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model.Update(tea.KeyMsg{Type: tea.KeyEnd})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" EDITED_BODY")})
	model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	after, err := authoringcore.ReadSpecDocument(t.Context(), operations.projectContext.ProjectDir, before.Reference)
	if err != nil || after.Title != "Manual contract edited" || !strings.Contains(after.Body, "EDITED_BODY") || after.CaseSlug != before.CaseSlug || after.Status != "approved" || after.Revision == before.Revision {
		t.Fatal(after, err)
	}
}

func cancelTUISpecForm(t *testing.T, operations service, model *tui.Model) {
	t.Helper()

	before := workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))
	runGuidedTUIKey(t, model, "e", false)
	model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !maps.Equal(before, workflowCLIImage(t, filepath.Join(operations.projectContext.ProjectDir, "ops"))) {
		t.Fatal("cancelled form changed canonical files")
	}
}
