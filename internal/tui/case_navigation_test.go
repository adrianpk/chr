package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func caseNavigationModel() *Model {
	return NewModel(Options{
		Cases:   []authoringcore.CaseMetadata{{Slug: "one", Name: "First case"}, {Slug: "two", Name: "Second case"}},
		Dumps:   []DumpView{{Title: "First dump", Scope: DumpScope{CaseSlug: "one"}}, {Title: "Second dump", Scope: DumpScope{CaseSlug: "two"}}},
		Quizzes: []QuizView{{Title: "First quiz", Scope: QuizScope{CaseSlug: "one"}}, {Title: "Second quiz", Scope: QuizScope{CaseSlug: "two"}}},
		Specs:   []SpecView{{Title: "First spec", Scope: QuizScope{CaseSlug: "one"}}, {Title: "Second spec", Scope: QuizScope{CaseSlug: "two"}}},
		Plans:   []PlanView{{Title: "First plan", Scope: QuizScope{CaseSlug: "one"}}, {Title: "Second plan", Scope: QuizScope{CaseSlug: "two"}}},
		Tickets: []ticketcore.Indexed{
			{IndexedIdentity: ticketcore.IndexedIdentity{Key: "ONE"}, IndexedClassification: ticketcore.IndexedClassification{UseCase: "one", Status: ticketcore.StatusReadyForDev}},
			{IndexedIdentity: ticketcore.IndexedIdentity{Key: "TWO"}, IndexedClassification: ticketcore.IndexedClassification{UseCase: "two", Status: ticketcore.StatusReadyForDev}},
		},
	})
}

// Selecting a case binds all document tabs and Board without calling an
// executor; returning through Help or Project must retain the same scope.
func TestCaseNavigation(t *testing.T) {
	model := caseNavigationModel()
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	for _, key := range []tea.KeyType{tea.KeyF2, tea.KeyF3, tea.KeyF4, tea.KeyF5, tea.KeyF6} {
		_, command := model.Update(tea.KeyMsg{Type: key})
		if command != nil || model.caseContext.CaseSlug != "two" || model.defaultDumpScope().CaseSlug != "two" {
			t.Fatal("navigation dispatched work or changed case binding")
		}
	}

	assertSecondCaseArtifacts(t, model)

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if model.caseContext.CaseSlug != "two" || model.caseDirty || model.projectDirty {
		t.Fatal("contextual navigation changed documents")
	}
}

func assertSecondCaseArtifacts(t *testing.T, model *Model) {
	t.Helper()

	dump, _ := model.selectedDumpView()
	quiz, _ := model.selectedQuizView()
	spec, _ := model.selectedSpecView()
	plan, _ := model.selectedPlanView()

	ticket, _ := model.selectedTicketView()
	if dump.Title != "Second dump" || quiz.Title != "Second quiz" || spec.Title != "Second spec" || plan.Title != "Second plan" || ticket.Key != "TWO" {
		t.Fatal("a tab selected another case's artifact")
	}
}

// A removed case stays explicitly unavailable; stale rows cannot cause a
// fallback to another case or to project-wide generation.
func TestCaseNavigationRemoved(t *testing.T) {
	model := caseNavigationModel()
	model.bindCaseContext("two", "")

	model.cases = model.cases[:1]
	model.clampSelectedCase()
	model.selectView(ViewSpecs)

	if _, ok := model.selectedSpecView(); ok || model.selectedSpecScope().CaseSlug != "two" || !strings.Contains(model.statusMessage, "unavailable") {
		t.Fatal("removed binding silently selected another scope")
	}

	model.selectView(ViewCases)
	model.selectView(ViewPlans)

	if model.caseContext.CaseSlug != "two" || len(model.visiblePlans()) != 0 {
		t.Fatal("returning through Cases replaced an unavailable binding")
	}

	model.executeCommand(Command{Action: CommandActionClearCase})
	model.selectView(ViewDumps)

	if model.caseContext.CaseSlug != "" || len(model.visibleDumps()) != 2 {
		t.Fatal("explicit context clearing did not recover all scopes")
	}
}

// Opening a retained run binds its observed case; passive observations of the
// same run keep selection and unsent input rather than applying navigation.
func TestCaseNavigationRun(t *testing.T) {
	model := caseNavigationModel()
	model.guided = guidedState{Active: true, Sequence: 1}
	session := &authoringcore.Session{ID: "retained", Guided: &authoringcore.GuidedUseCaseState{CaseSlug: "two"}}
	model.completeGuided(guidedFinishedMsg{Sequence: 1, View: GuidedView{Session: session}})

	model.caseDirty = true
	model.cases[1].Context = "Unsent case context"
	model.selectedSpec = 1
	model.completeGuided(guidedFinishedMsg{Sequence: 1, Request: GuidedRequest{Action: "inspect"}, View: GuidedView{Session: session, Cases: []authoringcore.CaseMetadata{{Slug: "two", Name: "Observed replacement"}}}})

	if model.cases[1].Context != "Unsent case context" || model.selectedSpec != 1 {
		t.Fatal("passive run observation replaced local input or selection")
	}

	model.Update(tea.KeyMsg{Type: tea.KeyF2})

	if model.caseContext.RunID != "retained" || model.caseContext.CaseSlug != "two" || !model.guided.Active {
		t.Fatal("retained run did not bind the shared document view")
	}

	model.selectView(ViewCases)
	model.selectView(ViewPlans)

	if model.caseContext.RunID != "retained" {
		t.Fatal("navigation lost the retained run")
	}
}
