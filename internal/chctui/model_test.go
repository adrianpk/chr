package chctui

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"forge.adrianpk.com/chc/internal/chc"
)

func TestModelViewRendersProjectAndNavigation(t *testing.T) {
	model := NewModel(Options{
		ProjectDir: "/workspace/project",
		ProjectID:  "project-1",
	})

	view := model.View()

	for _, want := range []string{
		"Project",
		"Name: project",
		"/workspace/project",
		"> Project",
		"Cases",
		"Dash",
		"Keys:",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestTabMovesPaneFocusWithoutChangingView(t *testing.T) {
	model := NewModel(Options{})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	navigated, ok := updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if navigated.SelectedView() != ViewProject {
		t.Fatalf("SelectedView() = %q, want %q", navigated.SelectedView(), ViewProject)
	}

	if navigated.selectedPane != 1 {
		t.Fatalf("selectedPane = %d, want 1", navigated.selectedPane)
	}

	updated, _ = navigated.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	navigated, ok = updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if navigated.SelectedView() != ViewProject {
		t.Fatalf("SelectedView() = %q, want %q", navigated.SelectedView(), ViewProject)
	}

	if navigated.selectedPane != 0 {
		t.Fatalf("selectedPane = %d, want 0", navigated.selectedPane)
	}
}

func TestPanelSplitsDefaultToEqualAspectRatio(t *testing.T) {
	model := NewModel(Options{})

	leftWidth, rightWidth := model.splitWidths(101)
	if leftWidth != 50 || rightWidth != 50 {
		t.Fatalf("splitWidths(101) = %d, %d, want 50, 50", leftWidth, rightWidth)
	}

	topHeight, bottomHeight := model.splitHeights(41)
	if topHeight != 20 || bottomHeight != 21 {
		t.Fatalf("splitHeights(41) = %d, %d, want 20, 21", topHeight, bottomHeight)
	}
}

func TestMouseDragUpdatesPanelSplits(t *testing.T) {
	model := NewModel(Options{})
	model.width = 100
	model.height = 30

	updated, _ := model.Update(tea.MouseMsg{
		X:      49,
		Y:      10,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	dragging := requireUpdatedModel(t, updated)

	updated, _ = dragging.Update(tea.MouseMsg{
		X:      70,
		Y:      10,
		Action: tea.MouseActionMotion,
		Button: tea.MouseButtonLeft,
	})
	dragged := requireUpdatedModel(t, updated)

	if dragged.verticalSplitPercent != 70 {
		t.Fatalf("verticalSplitPercent = %d, want 70", dragged.verticalSplitPercent)
	}

	updated, _ = dragged.Update(tea.MouseMsg{
		X:      80,
		Y:      14,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	dragging = requireUpdatedModel(t, updated)

	updated, _ = dragging.Update(tea.MouseMsg{
		X:      80,
		Y:      20,
		Action: tea.MouseActionMotion,
		Button: tea.MouseButtonLeft,
	})
	dragged = requireUpdatedModel(t, updated)

	if dragged.horizontalSplitPercent != 70 {
		t.Fatalf("horizontalSplitPercent = %d, want 70", dragged.horizontalSplitPercent)
	}
}

func TestAltMTogglesFocusedPaneMaximize(t *testing.T) {
	model := NewModel(Options{})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}, Alt: true})
	maximized := requireUpdatedModel(t, updated)

	if maximized.maximizedPane != 0 {
		t.Fatalf("maximizedPane = %d, want 0", maximized.maximizedPane)
	}

	view := maximized.renderProjectWorkspace(100, 20)
	if strings.Contains(view, "Location") || strings.Contains(view, "Artifacts") {
		t.Fatalf("maximized project workspace contains secondary panels:\n%s", view)
	}

	updated, _ = maximized.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}, Alt: true})
	restored := requireUpdatedModel(t, updated)

	if restored.maximizedPane != -1 {
		t.Fatalf("maximizedPane = %d, want -1", restored.maximizedPane)
	}
}

func TestLeftAndRightDoNotChangeGlobalTabs(t *testing.T) {
	model := NewModel(Options{})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	navigated := requireUpdatedModel(t, updated)

	if navigated.SelectedView() != ViewProject {
		t.Fatalf("right selected view = %q, want %q", navigated.SelectedView(), ViewProject)
	}

	updated, _ = navigated.Update(tea.KeyMsg{Type: tea.KeyLeft})
	navigated = requireUpdatedModel(t, updated)

	if navigated.SelectedView() != ViewProject {
		t.Fatalf("left selected view = %q, want %q", navigated.SelectedView(), ViewProject)
	}
}

func TestMouseClickSelectsShellTab(t *testing.T) {
	model := NewModel(Options{})
	model.width = 100
	model.height = 30

	updated, _ := model.Update(tea.MouseMsg{
		X:      12,
		Y:      29,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	navigated := requireUpdatedModel(t, updated)

	if navigated.SelectedView() != ViewCases {
		t.Fatalf("SelectedView() = %q, want %q", navigated.SelectedView(), ViewCases)
	}
}

func TestShellFooterAlignsKeysAndLabels(t *testing.T) {
	model := NewModel(Options{})
	footer := model.renderShellFooter(100)
	lines := strings.Split(footer, "\n")

	if len(lines) != 2 {
		t.Fatalf("footer lines = %d, want 2:\n%s", len(lines), footer)
	}

	for index, line := range lines {
		if lipgloss.Width(line) != 100 {
			t.Fatalf("footer line %d width = %d, want 100:\n%s", index, lipgloss.Width(line), footer)
		}
	}
}

func TestModelNavigatesWithFunctionKeys(t *testing.T) {
	model := NewModel(Options{})

	cases := []struct {
		key  tea.KeyType
		want View
	}{
		{key: tea.KeyF1, want: ViewProject},
		{key: tea.KeyF2, want: ViewCases},
		{key: tea.KeyF3, want: ViewDumps},
		{key: tea.KeyF4, want: ViewQuizzes},
		{key: tea.KeyF5, want: ViewSpecs},
		{key: tea.KeyF6, want: ViewPlans},
		{key: tea.KeyF7, want: ViewBoard},
		{key: tea.KeyF8, want: ViewDash},
		{key: tea.KeyF9, want: ViewHelp},
	}

	for _, testCase := range cases {
		updated, _ := model.Update(tea.KeyMsg{Type: testCase.key})
		navigated, ok := updated.(Model)

		if !ok {
			t.Fatalf("Update returned %T, want Model", updated)
		}

		if navigated.SelectedView() != testCase.want {
			t.Fatalf("SelectedView() = %q, want %q", navigated.SelectedView(), testCase.want)
		}

		model = navigated
	}
}

func TestModelEditsAndSavesProjectMetadata(t *testing.T) {
	executor := &fakeProjectMetadataExecutor{}
	model := NewModel(Options{
		ProjectMetadata: chc.ProjectMetadata{
			SchemaVersion: chc.ProjectMetadataSchemaVersion,
			Name:          "Old",
			Description:   "Old description.",
		},
		ProjectExecutor: executor,
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	editing, ok := updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	editing.projectEdit.Input.SetValue("New")

	updated, _ = editing.Update(tea.KeyMsg{Type: tea.KeyEnter})
	changed, ok := updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if !changed.projectDirty {
		t.Fatal("projectDirty = false, want true after project edit")
	}

	updated, _ = changed.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	saved, ok := updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if executor.saved.Name != "New" {
		t.Fatalf("saved.Name = %q, want New", executor.saved.Name)
	}

	if saved.projectDirty {
		t.Fatal("projectDirty = true, want false after save")
	}

	if saved.statusMessage != "" {
		t.Fatalf("statusMessage = %q, want empty status after normal project save", saved.statusMessage)
	}
}

func TestModelRefreshesWorkspaceGraphAfterProjectSave(t *testing.T) {
	projectExecutor := &fakeProjectMetadataExecutor{}
	workspaceExecutor := &fakeWorkspaceRefreshExecutor{
		result: WorkspaceRefreshResult{
			Graph: chc.WorkspaceGraph{
				Project: &chc.WorkspaceNode{
					Name:        "Graph Project",
					Description: "Graph description.",
					Path:        ".chc/project.md",
					SourceHash:  "hash-project",
				},
			},
		},
	}
	model := NewModel(Options{
		ProjectMetadata: chc.ProjectMetadata{
			SchemaVersion: chc.ProjectMetadataSchemaVersion,
			Name:          "Edited Project",
			Context:       "Preserved context.",
		},
		ProjectExecutor:   projectExecutor,
		WorkspaceExecutor: workspaceExecutor,
	})
	model.projectDirty = true

	model.saveProjectMetadata()

	if !workspaceExecutor.called {
		t.Fatal("RefreshWorkspace was not called after project save")
	}

	if model.projectDirty {
		t.Fatal("projectDirty = true, want false after save")
	}

	if model.projectMetadata.Name != "Graph Project" {
		t.Fatalf("project name = %q, want refreshed graph project", model.projectMetadata.Name)
	}

	if model.projectMetadata.Context != "Preserved context." {
		t.Fatalf("project context = %q, want body preserved from saved metadata", model.projectMetadata.Context)
	}
}

func TestModelProjectEditAcceptsSpaces(t *testing.T) {
	model := NewModel(Options{
		ProjectMetadata: chc.ProjectMetadata{
			SchemaVersion: chc.ProjectMetadataSchemaVersion,
			Name:          "Old",
			Description:   "Old description.",
		},
	})

	model.projectEdit = newProjectEditState(projectFieldDescription, "New")
	model.projectEdit.Input.CursorEnd()

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeySpace})
	editing, ok := updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	updated, _ = editing.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("description")})
	editing, ok = updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if editing.projectEdit.Input.Value() != "New description" {
		t.Fatalf("projectEdit value = %q, want space-preserving edit", editing.projectEdit.Input.Value())
	}
}

func TestModelProjectEditInsertsAtCursor(t *testing.T) {
	model := NewModel(Options{
		ProjectMetadata: chc.ProjectMetadata{
			SchemaVersion: chc.ProjectMetadataSchemaVersion,
			Name:          "Old",
			Description:   "Old description.",
		},
	})

	model.projectEdit = newProjectEditState(projectFieldDescription, "New description")
	model.projectEdit.Input.SetCursor(3)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeySpace})
	editing, ok := updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	updated, _ = editing.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("technical")})
	editing, ok = updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if editing.projectEdit.Input.Value() != "New technical description" {
		t.Fatalf("projectEdit value = %q, want insertion at cursor", editing.projectEdit.Input.Value())
	}
}

func TestProjectEditStartsAtBeginning(t *testing.T) {
	edit := newProjectEditState(projectFieldDescription, "abc")

	if edit.Input.Position() != 0 {
		t.Fatalf("cursor position = %d, want 0", edit.Input.Position())
	}
}

func TestRenderProjectInputDoesNotShiftTextAtCursor(t *testing.T) {
	model := NewModel(Options{})
	model.projectEdit = newProjectEditState(projectFieldDescription, "abc")
	model.projectEdit.Input.SetCursor(1)

	value := model.renderProjectInput(20)

	if strings.Contains(value, "_") {
		t.Fatalf("rendered edit value contains manual underscore cursor: %q", value)
	}
}

func TestModelCreatesAndSavesCaseMetadata(t *testing.T) {
	executor := &fakeCaseMetadataExecutor{}
	model := NewModel(Options{
		CaseExecutor: executor,
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyF2})
	casesView := requireUpdatedModel(t, updated)

	updated, _ = casesView.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	editing := requireUpdatedModel(t, updated)

	if !editing.caseEdit.Active {
		t.Fatal("caseEdit.Active = false, want active new case name edit")
	}

	editing.caseEdit.Input.SetValue("Blog Publishing")

	updated, _ = editing.Update(tea.KeyMsg{Type: tea.KeyEnter})
	changed := requireUpdatedModel(t, updated)

	updated, _ = changed.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	saved := requireUpdatedModel(t, updated)

	if executor.saved.Name != "Blog Publishing" {
		t.Fatalf("saved.Name = %q, want Blog Publishing", executor.saved.Name)
	}

	if saved.caseDirty {
		t.Fatal("caseDirty = true, want false after save")
	}

	if len(saved.cases) != 1 || saved.cases[0].Slug != "blog-publishing" {
		t.Fatalf("cases = %#v, want saved blog-publishing case", saved.cases)
	}
}

func TestModelRefreshesWorkspaceGraphAfterCaseSave(t *testing.T) {
	caseExecutor := &fakeCaseMetadataExecutor{}
	workspaceExecutor := &fakeWorkspaceRefreshExecutor{
		result: WorkspaceRefreshResult{
			Graph: chc.WorkspaceGraph{
				Cases: []chc.WorkspaceNode{
					{
						Slug:        "blog-publishing",
						Name:        "Graph Case",
						Description: "Graph case description.",
						Path:        ".chc/cases/blog-publishing.md",
						SourceHash:  "hash-case",
					},
				},
				CaseDocuments: map[string][]chc.WorkspaceNode{},
			},
		},
	}
	model := NewModel(Options{
		Cases: []chc.CaseMetadata{
			{
				SchemaVersion: chc.CaseMetadataSchemaVersion,
				Slug:          "blog-publishing",
				Name:          "Edited Case",
				Context:       "Preserved context.",
			},
		},
		CaseExecutor:      caseExecutor,
		WorkspaceExecutor: workspaceExecutor,
	})
	model.caseDirty = true

	model.saveSelectedCase()

	if !workspaceExecutor.called {
		t.Fatal("RefreshWorkspace was not called after case save")
	}

	if model.caseDirty {
		t.Fatal("caseDirty = true, want false after save")
	}

	if len(model.cases) != 1 || model.cases[0].Name != "Graph Case" {
		t.Fatalf("cases = %#v, want graph case after save", model.cases)
	}

	if model.cases[0].Context != "Preserved context." {
		t.Fatalf("case context = %q, want body preserved from saved metadata", model.cases[0].Context)
	}
}

func TestModelFiltersCaseMetadata(t *testing.T) {
	model := NewModel(Options{
		Cases: []chc.CaseMetadata{
			{SchemaVersion: chc.CaseMetadataSchemaVersion, Slug: "blog", Name: "Blog Publishing"},
			{SchemaVersion: chc.CaseMetadataSchemaVersion, Slug: "runtime", Name: "Runtime Observation"},
		},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyF2})
	casesView := requireUpdatedModel(t, updated)

	updated, _ = casesView.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	filtering := requireUpdatedModel(t, updated)

	updated, _ = filtering.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("runtime")})
	filtered := requireUpdatedModel(t, updated)

	cases := filtered.filteredCases()
	if len(cases) != 1 {
		t.Fatalf("len(filteredCases) = %d, want 1", len(cases))
	}

	if cases[0].Slug != "runtime" {
		t.Fatalf("filtered case = %q, want runtime", cases[0].Slug)
	}
}

func TestCasesUseTabForPanesAndArrowsWithinFocusedPane(t *testing.T) {
	model := NewModel(Options{
		Cases: []chc.CaseMetadata{
			{SchemaVersion: chc.CaseMetadataSchemaVersion, Slug: "blog", Name: "Blog Publishing"},
			{SchemaVersion: chc.CaseMetadataSchemaVersion, Slug: "runtime", Name: "Runtime Observation"},
		},
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyF2})
	casesView := requireUpdatedModel(t, updated)

	updated, _ = casesView.Update(tea.KeyMsg{Type: tea.KeyDown})
	selectedSecond := requireUpdatedModel(t, updated)

	if selectedSecond.selectedCase != 1 {
		t.Fatalf("selectedCase = %d, want 1", selectedSecond.selectedCase)
	}

	updated, _ = selectedSecond.Update(tea.KeyMsg{Type: tea.KeyTab})
	detailPane := requireUpdatedModel(t, updated)

	if detailPane.SelectedView() != ViewCases {
		t.Fatalf("SelectedView() = %q, want %q", detailPane.SelectedView(), ViewCases)
	}

	if detailPane.selectedPane != 1 {
		t.Fatalf("selectedPane = %d, want 1", detailPane.selectedPane)
	}

	updated, _ = detailPane.Update(tea.KeyMsg{Type: tea.KeyDown})
	nextField := requireUpdatedModel(t, updated)

	if nextField.selectedCase != 1 {
		t.Fatalf("selectedCase = %d, want 1", nextField.selectedCase)
	}

	if nextField.selectedCaseField != caseFieldDescription {
		t.Fatalf("selectedCaseField = %d, want %d", nextField.selectedCaseField, caseFieldDescription)
	}
}

func TestModelRefreshesWorkspaceGraphFromCommandPalette(t *testing.T) {
	executor := &fakeWorkspaceRefreshExecutor{
		result: WorkspaceRefreshResult{
			Message: "Workspace refreshed.",
			Graph: chc.WorkspaceGraph{
				Project: &chc.WorkspaceNode{Name: "External"},
			},
		},
	}
	model := NewModel(Options{
		WorkspaceExecutor: executor,
	})

	model.executeCommand(Command{Action: CommandActionRefresh})

	if model.workspaceGraph.Project == nil || model.workspaceGraph.Project.Name != "External" {
		t.Fatalf("workspaceGraph.Project = %#v, want refreshed graph", model.workspaceGraph.Project)
	}

	if model.statusMessage != "Workspace refreshed." {
		t.Fatalf("statusMessage = %q, want refresh status", model.statusMessage)
	}
}

func TestProjectRefreshKeyRebuildsWorkspaceGraph(t *testing.T) {
	executor := &fakeWorkspaceRefreshExecutor{
		result: WorkspaceRefreshResult{
			Message: "Workspace refreshed.",
			Graph: chc.WorkspaceGraph{
				Project: &chc.WorkspaceNode{Name: "External"},
			},
		},
	}
	model := NewModel(Options{
		WorkspaceExecutor: executor,
	})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	refreshed := requireUpdatedModel(t, updated)

	if !executor.called {
		t.Fatal("RefreshWorkspace was not called from Project refresh key")
	}

	if refreshed.projectMetadata.Name != "External" {
		t.Fatalf("project name = %q, want refreshed graph project", refreshed.projectMetadata.Name)
	}
}

func TestCasesRefreshKeyRebuildsWorkspaceGraph(t *testing.T) {
	executor := &fakeWorkspaceRefreshExecutor{
		result: WorkspaceRefreshResult{
			Message: "Workspace refreshed.",
			Graph: chc.WorkspaceGraph{
				Cases: []chc.WorkspaceNode{
					{Slug: "external", Name: "External Case"},
				},
				CaseDocuments: map[string][]chc.WorkspaceNode{},
			},
		},
	}
	model := NewModel(Options{
		WorkspaceExecutor: executor,
	})
	model.cases = []chc.CaseMetadata{{Slug: "blog", Name: "Blog"}}
	model.selected = 1

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	refreshed := requireUpdatedModel(t, updated)

	if !executor.called {
		t.Fatal("RefreshWorkspace was not called from Cases refresh key")
	}

	if len(refreshed.cases) != 1 || refreshed.cases[0].Slug != "external" {
		t.Fatalf("cases = %#v, want refreshed graph case", refreshed.cases)
	}
}

func TestModelInitializesProjectAndCasesFromWorkspaceGraph(t *testing.T) {
	model := NewModel(Options{
		ProjectMetadata: chc.ProjectMetadata{
			Name:        "Fallback",
			Description: "Fallback description.",
		},
		Cases: []chc.CaseMetadata{
			{
				Slug:        "fallback",
				Name:        "Fallback Case",
				Description: "Fallback case description.",
				Context:     "Preserved context.",
			},
		},
		WorkspaceGraph: chc.WorkspaceGraph{
			Project: &chc.WorkspaceNode{
				Name:        "Graph Project",
				Description: "Graph description.",
				Path:        ".chc/project.md",
				SourceHash:  "hash-project",
			},
			Cases: []chc.WorkspaceNode{
				{
					Slug:        "fallback",
					Name:        "Graph Case",
					Description: "Graph case description.",
					Path:        ".chc/cases/fallback.md",
					SourceHash:  "hash-case",
				},
			},
			CaseDocuments: map[string][]chc.WorkspaceNode{},
		},
	})

	if model.projectMetadata.Name != "Graph Project" {
		t.Fatalf("project name = %q, want graph project", model.projectMetadata.Name)
	}

	if len(model.cases) != 1 || model.cases[0].Name != "Graph Case" {
		t.Fatalf("cases = %#v, want graph case", model.cases)
	}

	if model.cases[0].Context != "Preserved context." {
		t.Fatalf("case context = %q, want preserved service-loaded body", model.cases[0].Context)
	}

	if len(model.workspaceIndex.NodesByID) == 0 {
		t.Fatal("workspaceIndex.NodesByID is empty, want index derived from graph")
	}
}

func TestModelRefreshUpdatesWorkspaceIndex(t *testing.T) {
	initialGraph := chc.WorkspaceGraph{
		ProjectWide: []chc.WorkspaceNode{
			{
				Kind:         chc.WorkspaceNodeSpec,
				RelativePath: ".chc/specs/old.md",
				Title:        "Old",
				SourceHash:   "old",
			},
		},
		CaseDocuments: map[string][]chc.WorkspaceNode{},
	}
	refreshedGraph := chc.WorkspaceGraph{
		ProjectWide: []chc.WorkspaceNode{
			{
				Kind:         chc.WorkspaceNodeSpec,
				RelativePath: ".chc/specs/new.md",
				Title:        "New",
				SourceHash:   "new",
			},
		},
		CaseDocuments: map[string][]chc.WorkspaceNode{},
	}
	executor := &fakeWorkspaceRefreshExecutor{
		result: WorkspaceRefreshResult{
			Graph: refreshedGraph,
			Index: chc.NewWorkspaceIndexFromGraph(refreshedGraph),
			SemanticDigestStates: []chc.WorkspaceSemanticDigestState{
				{NodeID: "spec:.chc/specs/new.md", Status: chc.WorkspaceSemanticDigestPending},
			},
		},
	}
	model := NewModel(Options{
		WorkspaceGraph:    initialGraph,
		WorkspaceExecutor: executor,
	})

	model.refreshWorkspace()

	if _, ok := model.workspaceIndex.NodesByID["spec:.chc/specs/new.md"]; !ok {
		t.Fatalf("workspaceIndex.NodesByID = %#v, want refreshed spec", model.workspaceIndex.NodesByID)
	}

	if len(model.semanticDigestStates) != 1 {
		t.Fatalf("semanticDigestStates = %#v, want one state", model.semanticDigestStates)
	}
}

func TestModelCountsCaseArtifactsFromWorkspaceGraph(t *testing.T) {
	model := NewModel(Options{
		WorkspaceGraph: chc.WorkspaceGraph{
			Cases: []chc.WorkspaceNode{
				{Slug: "blog", Name: "Blog"},
			},
			CaseDocuments: map[string][]chc.WorkspaceNode{
				"blog": {
					{Kind: chc.WorkspaceNodeDump, RelativePath: ".chc/dumps/blog.md", CaseSlug: "blog"},
					{Kind: chc.WorkspaceNodeSpec, RelativePath: ".chc/specs/blog.md", CaseSlug: "blog"},
					{Kind: chc.WorkspaceNodeTicket, RelativePath: ".chc/tickets/CHC-1.md", CaseSlug: "blog"},
				},
			},
			UnresolvedCaseReferences: []chc.WorkspaceCaseReference{
				{
					CaseSlug: "blog",
					References: []chc.WorkspaceNode{
						{Kind: chc.WorkspaceNodePlan, CaseSlug: "blog"},
					},
				},
			},
		},
	})

	lines := model.renderCaseArtifactLines()
	rendered := strings.Join(lines, "\n")

	for _, want := range []string{
		"Dumps             1",
		"Specs             1",
		"Board tickets     1",
		"Unresolved refs   1",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("artifact lines missing %q:\n%s", want, rendered)
		}
	}
}

func TestModelShowsUnresolvedCaseReferencesInCaseList(t *testing.T) {
	model := NewModel(Options{
		WorkspaceGraph: chc.WorkspaceGraph{
			UnresolvedCaseReferences: []chc.WorkspaceCaseReference{
				{
					CaseSlug: "missing",
					References: []chc.WorkspaceNode{
						{Kind: chc.WorkspaceNodePlan, CaseSlug: "missing"},
						{Kind: chc.WorkspaceNodeSpec, CaseSlug: "missing"},
					},
				},
			},
		},
	})

	rendered := strings.Join(model.renderCaseListLines(80, 10), "\n")

	for _, want := range []string{
		"Unresolved references",
		"missing (2)",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("case list missing %q:\n%s", want, rendered)
		}
	}
}

func TestModelFiltersCasesWithWorkspaceIndexSuggestions(t *testing.T) {
	index := chc.WorkspaceIndex{
		NodesByID: map[string]chc.WorkspaceNode{
			"case:blog": {
				Kind:        chc.WorkspaceNodeCase,
				Slug:        "blog",
				Name:        "Blog Publishing",
				Description: "Publish articles.",
			},
			"case:runtime": {
				Kind:        chc.WorkspaceNodeCase,
				Slug:        "runtime",
				Name:        "Runtime Operations",
				Description: "Observe runtime work.",
			},
		},
		CasesBySlug: map[string]string{
			"blog":    "case:blog",
			"runtime": "case:runtime",
		},
	}
	model := NewModel(Options{WorkspaceIndex: index})
	model.caseFilter = "publishing"

	cases := model.filteredCases()

	if len(cases) != 1 || cases[0].Slug != "blog" {
		t.Fatalf("filtered cases = %#v, want blog", cases)
	}
}

func TestModelShowsWorkspaceIndexHealthInDash(t *testing.T) {
	model := NewModel(Options{
		WorkspaceGraph: chc.WorkspaceGraph{
			ProjectWide: []chc.WorkspaceNode{
				{Kind: chc.WorkspaceNodeSpec, RelativePath: ".chc/specs/one.md"},
			},
			CaseDocuments: map[string][]chc.WorkspaceNode{},
		},
		SemanticDigestStates: []chc.WorkspaceSemanticDigestState{
			{NodeID: "spec:.chc/specs/one.md", Status: chc.WorkspaceSemanticDigestPending},
		},
	})

	rendered := model.renderWorkspaceIndexHealth()

	for _, want := range []string{
		"Workspace Index",
		"Nodes: 1",
		"Semantic pending: 1",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("workspace index health missing %q:\n%s", want, rendered)
		}
	}
}

func TestModelRefreshReportsWorkspaceDiagnostics(t *testing.T) {
	executor := &fakeWorkspaceRefreshExecutor{
		result: WorkspaceRefreshResult{
			Graph: chc.WorkspaceGraph{
				Diagnostics: []chc.WorkspaceGraphDiagnostic{
					{Path: ".chc/cases/broken.md", Kind: chc.WorkspaceNodeCase, Message: "invalid"},
				},
			},
		},
	}
	model := NewModel(Options{
		WorkspaceExecutor: executor,
	})

	model.executeCommand(Command{Action: CommandActionRefresh})

	if model.statusMessage != "Workspace refreshed with 1 diagnostics." {
		t.Fatalf("statusMessage = %q, want diagnostic refresh status", model.statusMessage)
	}
}

func TestModelRefreshBlocksActiveEdits(t *testing.T) {
	executor := &fakeWorkspaceRefreshExecutor{}
	model := NewModel(Options{
		WorkspaceExecutor: executor,
	})
	model.projectEdit = newProjectEditState(projectFieldName, "Draft")

	model.executeCommand(Command{Action: CommandActionRefresh})

	if executor.called {
		t.Fatal("RefreshWorkspace called with active edit, want blocked refresh")
	}

	if model.statusMessage != "Finish or cancel the active project edit before refreshing." {
		t.Fatalf("statusMessage = %q, want active edit block", model.statusMessage)
	}
}

func TestModelOpensAndFiltersCommandPalette(t *testing.T) {
	model := NewModel(Options{})

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})

	palette, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if !palette.commandPalette.Open {
		t.Fatal("command palette is closed, want open")
	}

	updated, _ = palette.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("dash")})

	palette, ok = updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	commands := palette.filteredCommands()
	if len(commands) == 0 {
		t.Fatal("filteredCommands returned no commands")
	}

	if commands[0].View != ViewDash {
		t.Fatalf("first filtered command view = %q, want %q", commands[0].View, ViewDash)
	}

	view := palette.View()
	for _, want := range []string{
		"Command Palette",
		"Filter: dash",
		"Go to Dash",
		"enter runs command",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelExecutesNavigationCommand(t *testing.T) {
	model := NewModel(Options{})
	model.commandPalette.Open = true
	model.commandPalette.Filter = "dash"

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	navigated, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if navigated.commandPalette.Open {
		t.Fatal("command palette is open after command execution")
	}

	if navigated.SelectedView() != ViewDash {
		t.Fatalf("SelectedView() = %q, want %q", navigated.SelectedView(), ViewDash)
	}

	if navigated.statusMessage != "Navigated to Dash." {
		t.Fatalf("statusMessage = %q, want navigation status", navigated.statusMessage)
	}
}

func TestModelReportsSelectedSourcePathCommand(t *testing.T) {
	model := NewModel(Options{
		Documents: []DocumentView{
			{Index: 0, Family: "spec", Title: "Spec", RelativePath: ".prj/specs/specs/tui.md"},
		},
	})
	model.selected = 4
	model.commandPalette.Open = true
	model.commandPalette.Filter = "source"

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	navigated, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if navigated.statusMessage != "Selected source path: .prj/specs/specs/tui.md" {
		t.Fatalf("statusMessage = %q, want selected source path", navigated.statusMessage)
	}
}

func TestModelPlansArtifactActionCommand(t *testing.T) {
	model := NewModel(Options{
		WorkspaceIndex: chc.WorkspaceIndex{
			NodesByID: map[string]chc.WorkspaceNode{
				"case:blog": {
					Kind: chc.WorkspaceNodeCase,
					Slug: "blog",
					Name: "Blog Publishing",
				},
			},
			CasesBySlug: map[string]string{"blog": "case:blog"},
		},
	})
	model.commandPalette.Open = true
	model.commandPalette.Filter = "create-ticket"

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	planned, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if planned.commandPalette.Open {
		t.Fatal("command palette is open after artifact action planning")
	}

	if !planned.artifactAction.Pending {
		t.Fatal("artifact action is not pending")
	}

	if planned.artifactAction.Action != ArtifactActionCreateTicket {
		t.Fatalf("artifact action = %q, want %q", planned.artifactAction.Action, ArtifactActionCreateTicket)
	}

	view := planned.View()
	for _, want := range []string{
		"Confirm Action",
		"Create ticket",
		"Title: TUI created ticket",
		"Use case suggestions:",
		"blog - Blog Publishing",
		"enter confirms action",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestDefaultCommandRegistryDoesNotExposeDraftCreation(t *testing.T) {
	for _, command := range DefaultCommandRegistry() {
		if command.ID == "create-draft" || strings.Contains(strings.ToLower(command.Title), "draft") {
			t.Fatalf("command registry exposes draft creation command: %#v", command)
		}
	}
}

func TestDefaultCommandRegistryExposesDumpWorkflows(t *testing.T) {
	commandsByID := map[string]Command{}
	for _, command := range DefaultCommandRegistry() {
		commandsByID[command.ID] = command
	}

	for _, id := range []string{"create-dump", "edit-dump"} {
		command, ok := commandsByID[id]
		if !ok {
			t.Fatalf("command %q not found", id)
		}

		if command.Action != CommandActionPlanDump {
			t.Fatalf("command %q action = %q, want %q", id, command.Action, CommandActionPlanDump)
		}
	}
}

func TestModelPlansCreateDumpCommandWithCaseScope(t *testing.T) {
	model := NewModel(Options{
		Cases: []chc.CaseMetadata{
			{Slug: "blog", Name: "Blog"},
		},
		WorkspaceIndex: chc.WorkspaceIndex{
			NodesByID: map[string]chc.WorkspaceNode{
				"case:blog": {
					Kind: chc.WorkspaceNodeCase,
					Slug: "blog",
					Name: "Blog",
				},
			},
			CasesBySlug: map[string]string{"blog": "case:blog"},
		},
	})
	model.cases = []chc.CaseMetadata{{Slug: "blog", Name: "Blog"}}
	model.selected = 1
	model.commandPalette.Open = true
	model.commandPalette.Filter = "create-dump"

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	planned, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if !planned.dumpAction.Pending {
		t.Fatal("dump action is not pending")
	}

	if planned.dumpAction.Request.Scope.CaseSlug != "blog" {
		t.Fatalf("dump scope = %#v, want blog", planned.dumpAction.Request.Scope)
	}

	view := planned.renderDumpAction()
	for _, want := range []string{
		"Confirm Action",
		"Create dump",
		"Title: blog dump",
		"Scope: blog",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelConfirmsDumpAction(t *testing.T) {
	executor := &fakeDumpActionExecutor{
		result: DumpActionResult{
			Message: "Created dump .chc/dumps/project-dump.md.",
			Dumps: []DumpView{
				{Title: "Project dump", RelativePath: ".chc/dumps/project-dump.md"},
			},
		},
	}
	model := NewModel(Options{DumpExecutor: executor})
	model.dumpAction = DumpActionState{
		Pending: true,
		Request: DumpActionRequest{
			Action: DumpActionCreate,
			Title:  "Project dump",
		},
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	confirmed, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if executor.request.Action != DumpActionCreate {
		t.Fatalf("dump action = %q, want %q", executor.request.Action, DumpActionCreate)
	}

	if confirmed.dumpAction.Pending {
		t.Fatal("dump action is pending after confirmation")
	}

	if confirmed.statusMessage != executor.result.Message {
		t.Fatalf("statusMessage = %q, want %q", confirmed.statusMessage, executor.result.Message)
	}

	dump, ok := confirmed.selectedDumpView()
	if !ok {
		t.Fatal("selectedDumpView returned no dump")
	}

	if dump.Title != "Project dump" {
		t.Fatalf("selected dump = %q, want Project dump", dump.Title)
	}
}

func TestModelRendersEmptyDumpState(t *testing.T) {
	model := NewModel(Options{})
	model.selected = 2

	view := model.View()
	for _, want := range []string{
		"No dumps found.",
		"Use the command palette to create one.",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelPlansEditDumpCommand(t *testing.T) {
	model := NewModel(Options{
		Dumps: []DumpView{
			{Title: "Project dump", RelativePath: ".chc/dumps/project-dump.md"},
		},
	})
	model.selected = 2
	model.commandPalette.Open = true
	model.commandPalette.Filter = "edit-dump"

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	planned, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if !planned.dumpAction.Pending {
		t.Fatal("dump action is not pending")
	}

	if planned.dumpAction.Request.Action != DumpActionEdit {
		t.Fatalf("dump action = %q, want %q", planned.dumpAction.Request.Action, DumpActionEdit)
	}

	if planned.dumpAction.Request.Path != ".chc/dumps/project-dump.md" {
		t.Fatalf("dump path = %q", planned.dumpAction.Request.Path)
	}
}

func TestModelConfirmsAppendDumpAction(t *testing.T) {
	executor := &fakeDumpActionExecutor{
		result: DumpActionResult{
			Message: "Appended dump .chc/dumps/project-dump.md.",
			Dumps: []DumpView{
				{Title: "Project dump", RelativePath: ".chc/dumps/project-dump.md"},
			},
		},
	}
	model := NewModel(Options{DumpExecutor: executor})
	model.dumpAction = DumpActionState{
		Pending: true,
		Request: DumpActionRequest{
			Action:  DumpActionAppend,
			Path:    ".chc/dumps/project-dump.md",
			Content: "Appended dump note from the TUI.",
		},
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	confirmed, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if executor.request.Action != DumpActionAppend {
		t.Fatalf("dump action = %q, want %q", executor.request.Action, DumpActionAppend)
	}

	if confirmed.statusMessage != executor.result.Message {
		t.Fatalf("statusMessage = %q, want %q", confirmed.statusMessage, executor.result.Message)
	}
}

func TestModelConfirmsEditDumpActionStartsInternalEditor(t *testing.T) {
	executor := &fakeDumpActionExecutor{
		result: DumpActionResult{
			Message:      "Editing dump .chc/dumps/project-dump.md.",
			Content:      "# Project dump\n\nExisting Markdown.",
			SelectedPath: ".chc/dumps/project-dump.md",
		},
	}
	model := NewModel(Options{DumpExecutor: executor})
	model.dumpAction = DumpActionState{
		Pending: true,
		Request: DumpActionRequest{
			Action: DumpActionEdit,
			Path:   ".chc/dumps/project-dump.md",
		},
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	confirmed, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if executor.request.Action != DumpActionEdit {
		t.Fatalf("dump action = %q, want %q", executor.request.Action, DumpActionEdit)
	}

	if !confirmed.dumpEdit.Active {
		t.Fatal("dumpEdit.Active = false, want active edit")
	}

	if confirmed.dumpEdit.Path != ".chc/dumps/project-dump.md" {
		t.Fatalf("dumpEdit.Path = %q", confirmed.dumpEdit.Path)
	}

	if confirmed.dumpEdit.Input.Value() != "# Project dump\n\nExisting Markdown." {
		t.Fatalf("dumpEdit value = %q", confirmed.dumpEdit.Input.Value())
	}

	if confirmed.statusMessage != "Editing dump. Press ctrl+s to save or esc to cancel." {
		t.Fatalf("statusMessage = %q", confirmed.statusMessage)
	}
}

func TestModelSavesInternalDumpEdit(t *testing.T) {
	executor := &fakeDumpActionExecutor{
		result: DumpActionResult{
			Message: "Saved dump .chc/dumps/project-dump.md.",
			Dumps: []DumpView{
				{Title: "Project dump", RelativePath: ".chc/dumps/project-dump.md"},
			},
			SelectedPath: ".chc/dumps/project-dump.md",
		},
	}
	model := NewModel(Options{DumpExecutor: executor})
	model.startDumpEdit(".chc/dumps/project-dump.md", "# Project dump\n\nChanged Markdown.")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	saved, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if executor.request.Action != DumpActionSave {
		t.Fatalf("dump action = %q, want %q", executor.request.Action, DumpActionSave)
	}

	if executor.request.Content != "# Project dump\n\nChanged Markdown." {
		t.Fatalf("saved content = %q", executor.request.Content)
	}

	if saved.dumpEdit.Active {
		t.Fatal("dumpEdit.Active = true after save")
	}
}

func TestModelRendersDumpEditPanelWithHints(t *testing.T) {
	model := NewModel(Options{
		Dumps: []DumpView{
			{
				Title:        "Project dump",
				Status:       "raw",
				RelativePath: ".chc/dumps/project-dump.md",
			},
		},
	})
	model.selected = 2
	model.selectedPane = 1
	model.width = 100
	model.height = 24
	model.startDumpEdit(".chc/dumps/project-dump.md", "# Project dump\n\nChanged Markdown.")

	view := model.View()

	for _, want := range []string{
		"Edit Dump",
		"Editing Markdown. ctrl+s save | esc cancel |",
		"alt+e external editor",
		"# Project dump",
		"Changed Markdown.",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelCancelsInternalDumpEdit(t *testing.T) {
	model := NewModel(Options{
		Dumps: []DumpView{
			{Title: "Project dump", RelativePath: ".chc/dumps/project-dump.md"},
		},
	})
	model.startDumpEdit(".chc/dumps/project-dump.md", "# Project dump")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	cancelled, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if cancelled.dumpEdit.Active {
		t.Fatal("dumpEdit.Active = true after cancel")
	}

	dump, ok := cancelled.selectedDumpView()
	if !ok {
		t.Fatal("selectedDumpView returned no dump")
	}

	if dump.RelativePath != ".chc/dumps/project-dump.md" {
		t.Fatalf("selected dump path = %q", dump.RelativePath)
	}
}

func TestModelAltELaunchesExternalDumpEditor(t *testing.T) {
	executor := &fakeDumpActionExecutor{
		result: DumpActionResult{
			Message:                  "Editing dump .chc/dumps/project-dump.md.",
			EditorCommand:            exec.Command("true"),
			EditorCommandDescription: "nvim project-dump.md",
		},
	}
	model := NewModel(Options{
		DumpExecutor: executor,
		Dumps: []DumpView{
			{Title: "Project dump", RelativePath: ".chc/dumps/project-dump.md"},
		},
	})
	model.selected = 2

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}, Alt: true})

	confirmed, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if cmd == nil {
		t.Fatal("cmd = nil, want external editor command")
	}

	if executor.request.Action != DumpActionExternalEdit {
		t.Fatalf("dump action = %q, want %q", executor.request.Action, DumpActionExternalEdit)
	}

	wantStatus := "External editor opened. chc will resume when the editor process exits."
	if confirmed.statusMessage != wantStatus {
		t.Fatalf("statusMessage = %q", confirmed.statusMessage)
	}
}

func TestModelReportsExternalEditorFailureContext(t *testing.T) {
	model := NewModel(Options{})

	updated, _ := model.Update(dumpEditorFinishedMsg{
		err:        errors.New("exit status 135"),
		path:       ".chc/dumps/project-dump.md",
		command:    "nvim project-dump.md",
		workingDir: "/workspace/.chc/dumps",
	})

	failed, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	for _, want := range []string{
		"Edit failed:",
		"command=nvim project-dump.md",
		"dir=/workspace/.chc/dumps",
		"exit status 135",
	} {
		if !strings.Contains(failed.statusMessage, want) {
			t.Fatalf("statusMessage missing %q: %q", want, failed.statusMessage)
		}
	}
}

func TestModelRefreshesDumpsAfterEditorFinishes(t *testing.T) {
	executor := &fakeDumpActionExecutor{
		result: DumpActionResult{
			Message: "Edited dump .chc/dumps/second.md.",
			Dumps: []DumpView{
				{Title: "First", RelativePath: ".chc/dumps/first.md"},
				{Title: "Second", RelativePath: ".chc/dumps/second.md"},
			},
			SelectedPath: ".chc/dumps/second.md",
		},
	}
	model := NewModel(Options{
		DumpExecutor: executor,
		Dumps: []DumpView{
			{Title: "Second", RelativePath: ".chc/dumps/second.md"},
		},
	})

	updated, _ := model.Update(dumpEditorFinishedMsg{path: ".chc/dumps/second.md"})

	refreshed, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if executor.request.Action != DumpActionRefresh {
		t.Fatalf("dump action = %q, want %q", executor.request.Action, DumpActionRefresh)
	}

	dump, ok := refreshed.selectedDumpView()
	if !ok {
		t.Fatal("selectedDumpView returned no dump")
	}

	if dump.RelativePath != ".chc/dumps/second.md" {
		t.Fatalf("selected dump path = %q", dump.RelativePath)
	}
}

func TestModelCancelsArtifactAction(t *testing.T) {
	model := NewModel(Options{})
	model.artifactAction = NewArtifactActionState(ArtifactActionCreateTicket)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	cancelled, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if cancelled.artifactAction.Pending {
		t.Fatal("artifact action is pending after cancellation")
	}

	if cancelled.statusMessage != "Action cancelled." {
		t.Fatalf("statusMessage = %q, want cancellation status", cancelled.statusMessage)
	}
}

func TestModelConfirmsArtifactAction(t *testing.T) {
	executor := &fakeArtifactActionExecutor{
		result: ArtifactActionResult{
			Message: "Created ticket CHC-1 at .chc/tickets/CHC-1-tui-created-ticket.md",
			Tickets: []chc.IndexedTicket{
				{Key: "CHC-1", Title: "TUI created ticket", Status: "open"},
			},
			Documents: []DocumentView{
				{Index: 0, Title: "TUI Plan"},
			},
			AuthoringSessions: []chc.AuthoringSession{
				{
					ID:        "author-20260522-100000",
					ProjectID: "project-1",
					Status:    chc.AuthoringSessionStatusDraft,
					Mode:      chc.AuthoringSessionModeTUI,
					Target:    chc.AuthoringTargetSpec,
					Topic:     "TUI artifact creation",
				},
			},
		},
	}
	model := NewModel(Options{ActionExecutor: executor})
	model.artifactAction = NewArtifactActionState(ArtifactActionCreateTicket)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	confirmed, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if executor.action != ArtifactActionCreateTicket {
		t.Fatalf("executor action = %q, want %q", executor.action, ArtifactActionCreateTicket)
	}

	if confirmed.artifactAction.Pending {
		t.Fatal("artifact action is pending after confirmation")
	}

	if confirmed.statusMessage != executor.result.Message {
		t.Fatalf("statusMessage = %q, want %q", confirmed.statusMessage, executor.result.Message)
	}

	assertConfirmedArtifactState(t, confirmed)
}

func TestModelReportsArtifactActionError(t *testing.T) {
	executor := &fakeArtifactActionExecutor{err: errors.New("write failed")}
	model := NewModel(Options{ActionExecutor: executor})
	model.artifactAction = NewArtifactActionState(ArtifactActionCreateTicket)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	failed, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if failed.artifactAction.Pending {
		t.Fatal("artifact action is pending after failure")
	}

	if failed.statusMessage != "Action failed: write failed" {
		t.Fatalf("statusMessage = %q, want action failure", failed.statusMessage)
	}
}

func TestModelPlansTicketActionCommand(t *testing.T) {
	model := NewModel(Options{
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "First", Status: "open"},
		},
	})
	model.commandPalette.Open = true
	model.commandPalette.Filter = "ticket-status-ready"

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	planned, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if planned.commandPalette.Open {
		t.Fatal("command palette is open after ticket action planning")
	}

	if !planned.ticketAction.Pending {
		t.Fatal("ticket action is not pending")
	}

	if planned.ticketAction.Request.Key != "CHC-1" {
		t.Fatalf("ticket action key = %q, want CHC-1", planned.ticketAction.Request.Key)
	}

	if planned.ticketAction.Request.Status != "ready" {
		t.Fatalf("ticket action status = %q, want ready", planned.ticketAction.Request.Status)
	}

	view := planned.View()
	for _, want := range []string{
		"Confirm Ticket Action",
		"Move ticket status",
		"Ticket: CHC-1",
		"New status: ready",
		"enter confirms ticket action",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelCancelsTicketAction(t *testing.T) {
	model := NewModel(Options{
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "First", Status: "open"},
		},
	})

	ticket, ok := model.selectedTicketView()
	if !ok {
		t.Fatal("selectedTicketView returned no ticket")
	}

	model.ticketAction = NewTicketActionState(TicketActionAddComment, ticket, "")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	cancelled, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if cancelled.ticketAction.Pending {
		t.Fatal("ticket action is pending after cancellation")
	}

	if cancelled.statusMessage != "Ticket action cancelled." {
		t.Fatalf("statusMessage = %q, want ticket cancellation status", cancelled.statusMessage)
	}
}

func TestModelConfirmsTicketAction(t *testing.T) {
	executor := &fakeTicketActionExecutor{
		result: TicketActionResult{
			Message: "Moved ticket CHC-1 to ready.",
			Tickets: []chc.IndexedTicket{
				{Key: "CHC-1", Title: "First", Status: "ready"},
			},
		},
	}
	model := NewModel(Options{
		TicketExecutor: executor,
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "First", Status: "open"},
		},
	})

	ticket, ok := model.selectedTicketView()
	if !ok {
		t.Fatal("selectedTicketView returned no ticket")
	}

	model.ticketAction = NewTicketActionState(TicketActionMoveStatus, ticket, "ready")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	confirmed, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if executor.request.Action != TicketActionMoveStatus {
		t.Fatalf("executor action = %q, want %q", executor.request.Action, TicketActionMoveStatus)
	}

	if executor.request.Key != "CHC-1" || executor.request.Status != "ready" {
		t.Fatalf("executor request = %#v, want CHC-1 ready", executor.request)
	}

	if confirmed.ticketAction.Pending {
		t.Fatal("ticket action is pending after confirmation")
	}

	if confirmed.statusMessage != executor.result.Message {
		t.Fatalf("statusMessage = %q, want %q", confirmed.statusMessage, executor.result.Message)
	}

	if len(confirmed.tickets) != 1 || confirmed.tickets[0].Status != "ready" {
		t.Fatalf("tickets = %#v, want reloaded ready ticket", confirmed.tickets)
	}
}

func TestModelReportsTicketActionError(t *testing.T) {
	executor := &fakeTicketActionExecutor{err: errors.New("invalid ticket status")}
	model := NewModel(Options{
		TicketExecutor: executor,
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "First", Status: "open"},
		},
	})

	ticket, ok := model.selectedTicketView()
	if !ok {
		t.Fatal("selectedTicketView returned no ticket")
	}

	model.ticketAction = NewTicketActionState(TicketActionMoveStatus, ticket, "unknown")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	failed, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if failed.ticketAction.Pending {
		t.Fatal("ticket action is pending after failure")
	}

	if failed.statusMessage != "Ticket action failed: invalid ticket status" {
		t.Fatalf("statusMessage = %q, want ticket action failure", failed.statusMessage)
	}
}

func TestModelPlansWorkActionCommand(t *testing.T) {
	model := NewModel(Options{
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "First", Status: "ready"},
		},
	})
	model.commandPalette.Open = true
	model.commandPalette.Filter = "work-start-ticket"

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	planned, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if planned.commandPalette.Open {
		t.Fatal("command palette is open after work action planning")
	}

	if !planned.workAction.Pending {
		t.Fatal("work action is not pending")
	}

	if planned.workAction.Request.TicketKey != "CHC-1" {
		t.Fatalf("work action ticket key = %q, want CHC-1", planned.workAction.Request.TicketKey)
	}

	view := planned.View()
	for _, want := range []string{
		"Confirm Work Action",
		"Start selected ticket",
		"Ticket: CHC-1",
		"enter confirms work action",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelCancelsWorkAction(t *testing.T) {
	model := NewModel(Options{
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "First", Status: "ready"},
		},
	})

	ticket, ok := model.selectedTicketView()
	if !ok {
		t.Fatal("selectedTicketView returned no ticket")
	}

	model.workAction = NewWorkActionState(WorkActionStartTicket, ticket)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	cancelled, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if cancelled.workAction.Pending {
		t.Fatal("work action is pending after cancellation")
	}

	if cancelled.statusMessage != "Work action cancelled." {
		t.Fatalf("statusMessage = %q, want work cancellation status", cancelled.statusMessage)
	}
}

func TestModelConfirmsWorkAction(t *testing.T) {
	executor := &fakeWorkActionExecutor{
		result: WorkActionResult{
			Message: "Execution exec-1 for ticket CHC-1 finished with status succeeded.",
			Tickets: []chc.IndexedTicket{
				{Key: "CHC-1", Title: "First", Status: "reviewing"},
			},
			Observability: ObservabilityView{
				LatestExecution: ExecutionView{ID: "exec-1", Status: "succeeded"},
			},
		},
	}
	model := NewModel(Options{
		WorkExecutor: executor,
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "First", Status: "ready"},
		},
	})

	ticket, ok := model.selectedTicketView()
	if !ok {
		t.Fatal("selectedTicketView returned no ticket")
	}

	model.workAction = NewWorkActionState(WorkActionStartTicket, ticket)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	confirmed, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if executor.request.Action != WorkActionStartTicket || executor.request.TicketKey != "CHC-1" {
		t.Fatalf("executor request = %#v, want start CHC-1", executor.request)
	}

	if confirmed.workAction.Pending {
		t.Fatal("work action is pending after confirmation")
	}

	if confirmed.statusMessage != executor.result.Message {
		t.Fatalf("statusMessage = %q, want %q", confirmed.statusMessage, executor.result.Message)
	}

	if len(confirmed.tickets) != 1 || confirmed.tickets[0].Status != "reviewing" {
		t.Fatalf("tickets = %#v, want reloaded reviewing ticket", confirmed.tickets)
	}

	if confirmed.observability.LatestExecution.ID != "exec-1" {
		t.Fatalf("LatestExecution.ID = %q, want exec-1", confirmed.observability.LatestExecution.ID)
	}
}

func TestModelReportsWorkActionError(t *testing.T) {
	executor := &fakeWorkActionExecutor{err: errors.New("runtime validation failed")}
	model := NewModel(Options{
		WorkExecutor: executor,
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "First", Status: "ready"},
		},
	})

	ticket, ok := model.selectedTicketView()
	if !ok {
		t.Fatal("selectedTicketView returned no ticket")
	}

	model.workAction = NewWorkActionState(WorkActionStartTicket, ticket)

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	failed, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if failed.workAction.Pending {
		t.Fatal("work action is pending after failure")
	}

	if failed.statusMessage != "Work action failed: runtime validation failed" {
		t.Fatalf("statusMessage = %q, want work action failure", failed.statusMessage)
	}
}

func TestModelRendersTicketBoardAndDetail(t *testing.T) {
	projectDir := t.TempDir()
	firstPath := writeTicketViewFile(t, projectDir, "open.md", "- [x] Done\n- [ ] Pending\n\n# Comments\n\n## 2026-05-22T10:00:00Z chc\n")
	secondPath := writeTicketViewFile(t, projectDir, "ready.md", "- [ ] Ready\n")

	model := NewModel(Options{
		Tickets: []chc.IndexedTicket{
			{
				Path:         firstPath,
				RelativePath: "open.md",
				Key:          "CHC-1",
				Title:        "Fix publishing",
				Status:       "open",
				Priority:     "high",
				Severity:     "critical",
				Type:         "bug",
				UseCase:      "cli-blog",
				ProjectArea:  "indexing",
			},
			{
				Path:         secondPath,
				RelativePath: "ready.md",
				Key:          "CHC-2",
				Title:        "Add completion",
				Status:       "ready",
				Priority:     "normal",
				Severity:     "normal",
				Type:         "feature",
				UseCase:      "cli-blog",
				ProjectArea:  "tasks",
			},
		},
	})
	model.selected = 6

	view := model.View()

	for _, want := range []string{
		"Board",
		"open",
		"ready",
		"CHC-1",
		"Fix publishing",
		"Selected Ticket",
		"Checklist: 1 done / 2 total",
		"Comments: 1",
		"Path: open.md",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelSelectsTickets(t *testing.T) {
	model := NewModel(Options{
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "First", Status: "open"},
			{Key: "CHC-2", Title: "Second", Status: "ready"},
		},
	})
	model.selected = 6

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	navigated, ok := updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	ticket, ok := navigated.selectedTicketView()
	if !ok {
		t.Fatal("selectedTicketView returned no ticket")
	}

	if ticket.Key != "CHC-2" {
		t.Fatalf("selected ticket = %q, want CHC-2", ticket.Key)
	}
}

func TestTicketFilterMatchesMetadata(t *testing.T) {
	model := NewModel(Options{
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "Fix publishing", Status: "open"},
			{Key: "CHC-2", Title: "Add completion", Status: "ready"},
		},
	})
	model.ticketFilter = "completion"

	tickets := model.filteredTickets()
	if len(tickets) != 1 {
		t.Fatalf("len(filteredTickets) = %d, want 1", len(tickets))
	}

	if tickets[0].Key != "CHC-2" {
		t.Fatalf("filtered ticket = %q, want CHC-2", tickets[0].Key)
	}
}

func TestModelRendersDocumentsAndDetail(t *testing.T) {
	model := NewModel(Options{
		Documents: []DocumentView{
			{
				Index:        0,
				Family:       "spec",
				Status:       "accepted",
				Title:        "TUI Navigation",
				RelativePath: ".prj/specs/specs/tui-navigation.md",
				Preview:      "Navigation details.",
			},
		},
	})
	model.selected = 4

	view := model.View()

	for _, want := range []string{
		"Specs",
		"spec",
		"accepted",
		"TUI Navigation",
		"Selected Document",
		"Path: .prj/specs/specs/tui-navigation.md",
		"Navigation details.",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestLoadDocumentViewsDiscoversProjectMarkdown(t *testing.T) {
	projectDir := t.TempDir()
	writeTicketViewFile(t, filepath.Join(projectDir, ".chc", "dumps"), "context.md", "# Project Context Dump\n\nStatus: raw\n\nNotes.")
	writeTicketViewFile(t, filepath.Join(projectDir, ".prj", "specs", "specs"), "feature.md", "# Feature Spec\n\nStatus: accepted\n\nDetails.")
	writeTicketViewFile(t, filepath.Join(projectDir, ".prj", "trackers"), "feature-tracker.md", "# Feature Tracker\n\nStatus: planned\n")
	writeTicketViewFile(t, filepath.Join(projectDir, ".prj", "drafts"), "legacy-draft.md", "# Legacy Draft\n\nStatus: draft\n")
	writeTicketViewFile(t, filepath.Join(projectDir, ".chc", "drafts"), "runtime-draft.md", "# Runtime Draft\n\nStatus: draft\n")

	documents, err := LoadDocumentViews(t.Context(), projectDir)
	if err != nil {
		t.Fatalf("LoadDocumentViews returned error: %v", err)
	}

	if len(documents) != 3 {
		t.Fatalf("len(documents) = %d, want 3", len(documents))
	}

	if documents[0].Title != "Project Context Dump" {
		t.Fatalf("first document title = %q, want Project Context Dump", documents[0].Title)
	}

	if documents[0].Status != "raw" {
		t.Fatalf("first document status = %q, want raw", documents[0].Status)
	}

	for _, document := range documents {
		if document.Family == "draft" {
			t.Fatalf("document family = draft for %#v, want no active draft documents", document)
		}
	}
}

func TestModelRendersAuthoringSessionsAndDetail(t *testing.T) {
	updated := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	model := NewModel(Options{
		AuthoringSessions: []chc.AuthoringSession{
			{
				ID:        "auth-1",
				ProjectID: "project-1",
				Status:    chc.AuthoringSessionStatusActive,
				Mode:      chc.AuthoringSessionModeAssisted,
				Target:    chc.AuthoringTargetSpec,
				Topic:     "Billing sync",
				UpdatedAt: updated,
				Inputs: []chc.AuthoringArtifactRef{
					{Path: ".prj/dumps/billing.md"},
				},
				Outputs: []chc.AuthoringArtifactRef{
					{Path: ".chc/specs/auth-1-billing-sync.md"},
				},
			},
		},
	})
	model.views = []View{ViewAuthoring}
	model.selected = 0

	view := model.View()

	for _, want := range []string{
		"Authoring",
		"active",
		"spec",
		"Billing sync",
		"Selected Authoring Session",
		"ID: auth-1",
		"Transcript: .chc/authoring/sessions/auth-1/transcript.md",
		"Inputs: .prj/dumps/billing.md",
		"Outputs: .chc/specs/auth-1-billing-sync.md",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelRendersAgentObservability(t *testing.T) {
	model := NewModel(Options{
		Observability: ObservabilityView{
			DaemonStatus:    "running",
			DaemonPID:       1234,
			DaemonUpdatedAt: "2026-05-22T10:00:00Z",
			DaemonHeartbeat: "2026-05-22T10:00:00Z",
			LatestLoop: LoopRunView{
				ID:             "loop-1",
				Status:         "blocked",
				UpdatedAt:      "2026-05-22T10:01:00Z",
				EligibleCount:  3,
				ProcessedCount: 1,
				BlockedCount:   1,
				FailedCount:    0,
				FailureReason:  "blocked ticket",
			},
			LatestExecution: ExecutionView{
				ID:             "exec-1",
				TicketKey:      "CHC-1",
				Status:         "failed",
				UpdatedAt:      "2026-05-22T10:02:00Z",
				FailureReason:  "validation failed",
				RunDir:         ".chc/runtime/runs/CHC-1/exec-1",
				ResultPath:     ".chc/runtime/runs/CHC-1/exec-1/result.json",
				LogPath:        ".chc/runtime/runs/CHC-1/exec-1/runtime.log",
				ValidationPath: ".chc/runtime/runs/CHC-1/exec-1/validation.log",
			},
		},
	})
	model.selected = 7

	view := model.View()

	for _, want := range []string{
		"Agent",
		"Daemon: running",
		"PID: 1234",
		"Latest Loop",
		"ID: loop-1",
		"Counts: 3 eligible / 1 processed / 1 blocked / 0 failed",
		"Latest Execution",
		"Ticket: CHC-1",
		"Validation: .chc/runtime/runs/CHC-1/exec-1/validation.log",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelRendersRuntimeObservability(t *testing.T) {
	model := NewModel(Options{
		Observability: ObservabilityView{
			RuntimeImage:             "chc-agent-runtime:local",
			RuntimeDockerfile:        ".chc/runtime/agent/Dockerfile",
			RuntimeArtifactDir:       ".chc/runtime/runs",
			RuntimeValidationCommand: "make verify",
			RuntimeChecks: []RuntimeCheckView{
				{Name: "workspace", Status: "ok", Path: ".chc", Message: "ok"},
				{Name: "dockerfile", Status: "fail", Path: ".chc/runtime/agent/Dockerfile", Message: "missing"},
			},
		},
	})
	model.views = []View{ViewRuntime}
	model.selected = 0

	view := model.View()

	for _, want := range []string{
		"Runtime",
		"Image: chc-agent-runtime:local",
		"Dockerfile: .chc/runtime/agent/Dockerfile",
		"Artifact Dir: .chc/runtime/runs",
		"Validation: make verify",
		"workspace",
		"dockerfile",
		"fail",
		"missing",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestNewObservabilityViewSelectsLatestRecords(t *testing.T) {
	projectDir := t.TempDir()
	older := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	newer := older.Add(time.Hour)

	view := NewObservabilityView(
		projectDir,
		chc.NewStoppedDaemonState(newer),
		[]chc.ExecutionRecord{
			{ID: "exec-old", Status: chc.ExecutionStatusSucceeded, UpdatedAt: older},
			{ID: "exec-new", Status: chc.ExecutionStatusFailed, UpdatedAt: newer},
		},
		[]chc.LoopRunRecord{
			{ID: "loop-old", Status: chc.LoopRunStatusSucceeded, UpdatedAt: older},
			{ID: "loop-new", Status: chc.LoopRunStatusFailed, UpdatedAt: newer},
		},
		chc.RuntimeValidationResult{
			Plan: chc.RuntimePlan{
				Image:             "image",
				DockerfilePath:    filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile"),
				ArtifactDir:       filepath.Join(projectDir, ".chc", "runtime", "runs"),
				ValidationCommand: "make verify",
			},
			Checks: []chc.RuntimeValidationCheck{
				{Name: "dockerfile", Path: filepath.Join(projectDir, ".chc", "runtime", "agent", "Dockerfile"), OK: true, Message: "ok"},
			},
		},
	)

	if view.LatestExecution.ID != "exec-new" {
		t.Fatalf("LatestExecution.ID = %q, want exec-new", view.LatestExecution.ID)
	}

	if view.LatestLoop.ID != "loop-new" {
		t.Fatalf("LatestLoop.ID = %q, want loop-new", view.LatestLoop.ID)
	}

	if view.RuntimeDockerfile != ".chc/runtime/agent/Dockerfile" {
		t.Fatalf("RuntimeDockerfile = %q, want relative runtime path", view.RuntimeDockerfile)
	}
}

func TestModelSelectsDocuments(t *testing.T) {
	model := NewModel(Options{
		Documents: []DocumentView{
			{Index: 0, Family: "spec", Title: "First"},
			{Index: 1, Family: "spec", Title: "Second"},
		},
	})
	model.selected = 4

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	navigated, ok := updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	document, ok := navigated.selectedDocumentView()
	if !ok {
		t.Fatal("selectedDocumentView returned no document")
	}

	if document.Title != "Second" {
		t.Fatalf("selected document = %q, want Second", document.Title)
	}
}

func TestModelRendersScopedDumps(t *testing.T) {
	updated := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	model := NewModel(Options{
		Dumps: []DumpView{
			{
				Title:        "Project Context",
				Status:       "raw",
				Scope:        DumpScope{},
				RelativePath: ".chc/dumps/project-context.md",
				UpdatedAt:    updated,
				Preview:      "Project-wide notes.",
			},
			{
				Title:        "Blog Context",
				Status:       "curated",
				Scope:        DumpScope{CaseSlug: "blog"},
				RelativePath: ".chc/dumps/blog-context.md",
				UpdatedAt:    updated,
				Preview:      "Use-case notes.",
			},
		},
	})
	model.selected = 2

	view := model.View()
	for _, want := range []string{
		"Dumps",
		"Scope        Status      Title",
		"project",
		"Project Context",
		"blog",
		"Blog Context",
		"Selected Dump",
		"Scope: project",
		"Path: .chc/dumps/project-context.md",
		"Project-wide notes.",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelSelectsDumpsIndependentlyFromDocuments(t *testing.T) {
	model := NewModel(Options{
		Dumps: []DumpView{
			{Title: "Project Context", RelativePath: ".chc/dumps/project-context.md"},
			{Title: "Blog Context", Scope: DumpScope{CaseSlug: "blog"}, RelativePath: ".chc/dumps/blog-context.md"},
		},
		Documents: []DocumentView{
			{Index: 0, Family: "spec", Title: "Spec", RelativePath: ".chc/specs/spec.md"},
		},
	})
	model.selected = 2

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})

	navigated, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	dump, ok := navigated.selectedDumpView()
	if !ok {
		t.Fatal("selectedDumpView returned no dump")
	}

	if dump.Title != "Blog Context" {
		t.Fatalf("selected dump = %q, want Blog Context", dump.Title)
	}

	if navigated.selectedDocument != 0 {
		t.Fatalf("selectedDocument = %d, want unchanged 0", navigated.selectedDocument)
	}
}

func TestModelRendersScopedQuizzes(t *testing.T) {
	updated := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	model := NewModel(Options{
		Quizzes: []QuizView{
			{
				Title:          "Project",
				Status:         "not_started",
				Mode:           "deep",
				Scope:          QuizScope{},
				RawPath:        ".chc/quizzes/project-quiz-raw.md",
				CuratedPath:    ".chc/quizzes/project-quiz.md",
				RuntimePath:    ".chc/tmp/quiz-runtime/project.json",
				QuestionBudget: 50,
			},
			{
				Title:                 "Blog Publishing",
				Status:                "active",
				Mode:                  "fast",
				Scope:                 QuizScope{CaseSlug: "blog"},
				RawPath:               ".chc/quizzes/blog-quiz-raw.md",
				CuratedPath:           ".chc/quizzes/blog-quiz.md",
				RuntimePath:           ".chc/tmp/quiz-runtime/blog.json",
				QuestionBudget:        50,
				QuestionsAsked:        2,
				AnswersRecorded:       1,
				CandidateCount:        2,
				BranchReturnReason:    "max_depth_reached",
				UpdatedAt:             updated,
				ActiveQuestionSummary: "What should the workflow preserve?",
				TrailSummary:          "workflow",
				CuratedPreview:        "Preserve editorial review.",
			},
		},
	})
	model.selected = 3
	model.selectedQuiz = 1

	view := model.View()
	for _, want := range []string{
		"Quizzes",
		"Blog Publishing",
		"Question 2",
		"Question:",
		"What should the workflow preserve?",
		"Enter answers",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelHidesTechnicalQuizDetails(t *testing.T) {
	model := NewModel(Options{
		Quizzes: []QuizView{
			{
				Title:                 "Blog Publishing",
				Status:                "active",
				Mode:                  "fast",
				Scope:                 QuizScope{CaseSlug: "blog"},
				RawPath:               ".chc/quizzes/blog-quiz-raw.md",
				CuratedPath:           ".chc/quizzes/blog-quiz.md",
				RuntimePath:           ".chc/tmp/quiz-runtime/blog.json",
				QuestionBudget:        50,
				QuestionsAsked:        2,
				CandidateCount:        2,
				BranchReturnReason:    "max_depth_reached",
				ActiveQuestionSummary: "What should the workflow preserve?",
				TrailSummary:          "workflow",
				CuratedPreview:        "Preserve editorial review.",
			},
		},
	})
	model.selected = 3

	view := model.View()
	for _, unwanted := range []string{
		"Mode:",
		"Questions:",
		"Candidates:",
		"Return:",
		"Raw:",
		"Curated:",
		"Runtime:",
		"Trail:",
		"Curated preview:",
	} {
		if strings.Contains(view, unwanted) {
			t.Fatalf("View() includes technical detail %q:\n%s", unwanted, view)
		}
	}
}

func TestModelRendersCuratedQuizBodyWithoutActiveQuestion(t *testing.T) {
	model := NewModel(Options{
		Quizzes: []QuizView{
			{
				Title:          "Blog Publishing",
				Status:         "done",
				Mode:           "deep",
				Scope:          QuizScope{CaseSlug: "blog"},
				QuestionBudget: 50,
				QuestionsAsked: 2,
				CuratedPreview: "## Decisions\n- Preserve editorial review.\n## Open Questions\n- None recorded.",
			},
		},
	})
	model.selected = 3

	view := model.View()
	for _, want := range []string{
		"Done",
		"Notes:",
		"Preserve editorial review.",
		"Open Questions",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelSelectsQuizzesIndependentlyFromDocuments(t *testing.T) {
	model := NewModel(Options{
		Quizzes: []QuizView{
			{Title: "Project", Scope: QuizScope{}},
			{Title: "Blog", Scope: QuizScope{CaseSlug: "blog"}},
		},
		Documents: []DocumentView{
			{Index: 0, Family: "spec", Title: "Spec", RelativePath: ".chc/specs/spec.md"},
		},
	})
	model.selected = 3

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})

	navigated, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	quiz, ok := navigated.selectedQuizView()
	if !ok {
		t.Fatal("selectedQuizView returned no quiz")
	}

	if quiz.Title != "Blog" {
		t.Fatalf("selected quiz = %q, want Blog", quiz.Title)
	}

	if navigated.selectedDocument != 0 {
		t.Fatalf("selectedDocument = %d, want unchanged 0", navigated.selectedDocument)
	}
}

func TestModelConfirmsQuizStartThroughExecutor(t *testing.T) {
	executor := &fakeQuizActionExecutor{
		result: QuizActionResult{
			Message: "Quiz started.",
			Quizzes: []QuizView{
				{Title: "Project", Scope: QuizScope{}, Status: "not_started"},
				{Title: "Blog", Scope: QuizScope{CaseSlug: "blog"}, Status: "active"},
			},
			SelectedScope: QuizScope{CaseSlug: "blog"},
		},
	}
	model := NewModel(Options{
		QuizExecutor: executor,
		Quizzes: []QuizView{
			{Title: "Project", Scope: QuizScope{}, Status: "not_started"},
			{Title: "Blog", Scope: QuizScope{CaseSlug: "blog"}, Status: "not_started"},
		},
	})
	model.selected = 3
	model.selectedQuiz = 1
	model.planQuizAction(QuizActionStart)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	busy := requireUpdatedModel(t, updated)

	if !busy.quizBusy {
		t.Fatal("quiz action is not marked busy")
	}

	if busy.statusMessage != "Loading question..." {
		t.Fatalf("statusMessage = %q, want Loading question...", busy.statusMessage)
	}

	updated, _ = busy.Update(cmd())
	confirmed := requireUpdatedModel(t, updated)

	if executor.request.Action != QuizActionStart {
		t.Fatalf("action = %q, want start", executor.request.Action)
	}

	if executor.request.Scope.CaseSlug != "blog" {
		t.Fatalf("scope = %#v, want blog", executor.request.Scope)
	}

	if confirmed.statusMessage != "Quiz started." {
		t.Fatalf("statusMessage = %q, want Quiz started.", confirmed.statusMessage)
	}

	quiz, ok := confirmed.selectedQuizView()
	if !ok || quiz.Status != "active" || quiz.Scope.CaseSlug != "blog" {
		t.Fatalf("selected quiz = %#v, want active blog", quiz)
	}
}

func TestModelStartsQuizFromQuizzesKey(t *testing.T) {
	executor := &fakeQuizActionExecutor{
		result: QuizActionResult{
			Message: "Quiz started.",
			Quizzes: []QuizView{
				{Title: "Project", Scope: QuizScope{}, Status: "active", ActiveQuestionSummary: "What should be clarified?"},
			},
			SelectedScope: QuizScope{},
		},
	}
	model := NewModel(Options{
		QuizExecutor: executor,
		Quizzes: []QuizView{
			{Title: "Project", Scope: QuizScope{}, Status: "not_started"},
		},
	})
	model.selected = 3

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	busy := requireUpdatedModel(t, updated)

	if !busy.quizBusy {
		t.Fatal("quiz action is not marked busy")
	}

	if busy.statusMessage != "Loading question..." {
		t.Fatalf("statusMessage = %q, want Loading question...", busy.statusMessage)
	}

	updated, _ = busy.Update(cmd())
	confirmed := requireUpdatedModel(t, updated)

	if executor.request.Action != QuizActionStart {
		t.Fatalf("action = %q, want start", executor.request.Action)
	}

	if confirmed.statusMessage != "Answer the question. Enter submits; shift+enter inserts a new line." {
		t.Fatalf("statusMessage = %q, want answer hint", confirmed.statusMessage)
	}

	quiz, ok := confirmed.selectedQuizView()
	if !ok || quiz.ActiveQuestionSummary != "What should be clarified?" {
		t.Fatalf("selected quiz = %#v, want generated question", quiz)
	}

	if !confirmed.quizAnswer.Active {
		t.Fatal("quiz answer editor is not active")
	}

	if confirmed.selectedPane != 1 {
		t.Fatalf("selectedPane = %d, want 1", confirmed.selectedPane)
	}

	if confirmed.maximizedPane != 1 {
		t.Fatalf("maximizedPane = %d, want 1", confirmed.maximizedPane)
	}
}

func TestModelSubmitsQuizAnswerWithEnterThroughExecutor(t *testing.T) {
	executor := &fakeQuizActionExecutor{
		result: QuizActionResult{
			Message: "Quiz answer recorded.",
			Quizzes: []QuizView{
				{Title: "Blog", Scope: QuizScope{CaseSlug: "blog"}, Status: "active", QuestionsAsked: 2},
			},
			SelectedScope: QuizScope{CaseSlug: "blog"},
		},
	}
	model := NewModel(Options{
		QuizExecutor: executor,
		Quizzes: []QuizView{
			{
				Title:                 "Blog",
				Scope:                 QuizScope{CaseSlug: "blog"},
				Status:                "active",
				ActiveQuestionSummary: "What should be clarified?",
			},
		},
	})
	model.selected = 3
	model.startQuizAnswer()
	model.quizAnswer.Input.SetValue("Clarify editorial ownership.")

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	busy := requireUpdatedModel(t, updated)

	if !busy.quizBusy {
		t.Fatal("quiz action is not marked busy")
	}

	if busy.statusMessage != "Generating next question..." {
		t.Fatalf("statusMessage = %q, want Generating next question...", busy.statusMessage)
	}

	updated, _ = busy.Update(cmd())
	confirmed := requireUpdatedModel(t, updated)

	if executor.request.Action != QuizActionAnswer {
		t.Fatalf("action = %q, want answer", executor.request.Action)
	}

	if executor.request.Answer != "Clarify editorial ownership." {
		t.Fatalf("answer = %q", executor.request.Answer)
	}

	if confirmed.quizAnswer.Active {
		t.Fatal("quiz answer editor is still active")
	}

	if confirmed.statusMessage != "Quiz answer recorded." {
		t.Fatalf("statusMessage = %q, want Quiz answer recorded.", confirmed.statusMessage)
	}
}

func TestModelInsertsQuizAnswerLineWithShiftEnter(t *testing.T) {
	model := NewModel(Options{
		Quizzes: []QuizView{
			{
				Title:                 "Blog",
				Scope:                 QuizScope{CaseSlug: "blog"},
				Status:                "active",
				ActiveQuestionSummary: "What should be clarified?",
			},
		},
	})
	model.selected = 3
	model.startQuizAnswer()
	model.quizAnswer.Input.SetValue("First line")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})
	confirmed := requireUpdatedModel(t, updated)

	if !strings.Contains(confirmed.quizAnswer.Input.Value(), "\n") {
		t.Fatalf("answer value = %q, want inserted newline", confirmed.quizAnswer.Input.Value())
	}
}

func TestModelSkipsQuizAnswerWithAltS(t *testing.T) {
	executor := &fakeQuizActionExecutor{
		result: QuizActionResult{
			Message: "Quiz question skipped.",
			Quizzes: []QuizView{
				{Title: "Blog", Scope: QuizScope{CaseSlug: "blog"}, Status: "active", QuestionsAsked: 2},
			},
			SelectedScope: QuizScope{CaseSlug: "blog"},
		},
	}
	model := NewModel(Options{
		QuizExecutor: executor,
		Quizzes: []QuizView{
			{
				Title:                 "Blog",
				Scope:                 QuizScope{CaseSlug: "blog"},
				Status:                "active",
				ActiveQuestionSummary: "What should be clarified?",
			},
		},
	})
	model.selected = 3
	model.startQuizAnswer()

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s"), Alt: true})
	busy := requireUpdatedModel(t, updated)

	if !busy.quizBusy {
		t.Fatal("quiz action is not marked busy")
	}

	updated, _ = busy.Update(cmd())
	confirmed := requireUpdatedModel(t, updated)

	if executor.request.Action != QuizActionSkip {
		t.Fatalf("action = %q, want skip", executor.request.Action)
	}

	if confirmed.quizAnswer.Active {
		t.Fatal("quiz answer editor is still active")
	}
}

func TestModelCompletesQuizAnswerWithAltD(t *testing.T) {
	executor := &fakeQuizActionExecutor{
		result: QuizActionResult{
			Message: "Quiz completed.",
			Quizzes: []QuizView{
				{Title: "Blog", Scope: QuizScope{CaseSlug: "blog"}, Status: "done", QuestionsAsked: 2},
			},
			SelectedScope: QuizScope{CaseSlug: "blog"},
		},
	}
	model := NewModel(Options{
		QuizExecutor: executor,
		Quizzes: []QuizView{
			{
				Title:                 "Blog",
				Scope:                 QuizScope{CaseSlug: "blog"},
				Status:                "active",
				ActiveQuestionSummary: "What should be clarified?",
			},
		},
	})
	model.selected = 3
	model.startQuizAnswer()

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d"), Alt: true})
	busy := requireUpdatedModel(t, updated)

	if !busy.quizBusy {
		t.Fatal("quiz action is not marked busy")
	}

	updated, _ = busy.Update(cmd())
	confirmed := requireUpdatedModel(t, updated)

	if executor.request.Action != QuizActionComplete {
		t.Fatalf("action = %q, want complete", executor.request.Action)
	}

	if confirmed.quizAnswer.Active {
		t.Fatal("quiz answer editor is still active")
	}
}

func TestModelConfirmsQuizFastModeThroughExecutor(t *testing.T) {
	executor := &fakeQuizActionExecutor{
		result: QuizActionResult{
			Message: "Quiz mode switched.",
			Quizzes: []QuizView{
				{Title: "Blog", Scope: QuizScope{CaseSlug: "blog"}, Status: "active", Mode: "fast"},
			},
			SelectedScope: QuizScope{CaseSlug: "blog"},
		},
	}
	model := NewModel(Options{
		QuizExecutor: executor,
		Quizzes: []QuizView{
			{Title: "Blog", Scope: QuizScope{CaseSlug: "blog"}, Status: "active", Mode: "deep"},
		},
	})
	model.selected = 3
	model.planQuizAction(QuizActionFastMode)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	busy := requireUpdatedModel(t, updated)

	if !busy.quizBusy {
		t.Fatal("quiz action is not marked busy")
	}

	updated, _ = busy.Update(cmd())
	confirmed := requireUpdatedModel(t, updated)

	if executor.request.Action != QuizActionFastMode {
		t.Fatalf("action = %q, want fast mode", executor.request.Action)
	}

	if executor.request.Mode != "fast" {
		t.Fatalf("mode = %q, want fast", executor.request.Mode)
	}

	quiz, ok := confirmed.selectedQuizView()
	if !ok || quiz.Mode != "fast" {
		t.Fatalf("selected quiz = %#v, want fast mode", quiz)
	}
}

func TestModelRecordsWindowSize(t *testing.T) {
	model := NewModel(Options{})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	resized, ok := updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if resized.width != 120 {
		t.Fatalf("width = %d, want 120", resized.width)
	}

	if resized.height != 40 {
		t.Fatalf("height = %d, want 40", resized.height)
	}
}

func TestModelViewUsesWindowRealEstate(t *testing.T) {
	model := NewModel(Options{
		ProjectDir: "/workspace/project",
		ProjectID:  "project-1",
		Tickets: []chc.IndexedTicket{
			{Key: "CHC-1", Title: "First", Status: "open"},
		},
	})

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 20})

	resized, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	view := resized.View()

	lines := strings.Split(strings.TrimSuffix(view, "\n"), "\n")
	if len(lines) != 20 {
		t.Fatalf("View() rendered %d lines, want 20:\n%s", len(lines), view)
	}

	for index, line := range lines {
		if lipgloss.Width(line) != 80 {
			t.Fatalf("line %d width = %d, want 80: %q", index, lipgloss.Width(line), line)
		}
	}

	for _, want := range []string{
		"project",
		"F1",
		"Project",
		"F10",
		"Exit",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func writeTicketViewFile(t *testing.T, projectDir, name, body string) string {
	t.Helper()

	err := os.MkdirAll(projectDir, 0o755)
	if err != nil {
		t.Fatalf("create ticket view directory: %v", err)
	}

	path := filepath.Join(projectDir, name)

	err = os.WriteFile(path, []byte(body), 0o644)
	if err != nil {
		t.Fatalf("write ticket view file: %v", err)
	}

	return path
}

func requireUpdatedModel(t *testing.T, updated tea.Model) Model {
	t.Helper()

	model, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	return model
}

type fakeArtifactActionExecutor struct {
	action ArtifactActionID
	result ArtifactActionResult
	err    error
}

func (e *fakeArtifactActionExecutor) ExecuteArtifactAction(action ArtifactActionID) (ArtifactActionResult, error) {
	e.action = action

	return e.result, e.err
}

type fakeDumpActionExecutor struct {
	request DumpActionRequest
	result  DumpActionResult
	err     error
}

func (e *fakeDumpActionExecutor) ExecuteDumpAction(request DumpActionRequest) (DumpActionResult, error) {
	e.request = request

	return e.result, e.err
}

type fakeQuizActionExecutor struct {
	request QuizActionRequest
	result  QuizActionResult
	err     error
}

func (e *fakeQuizActionExecutor) ExecuteQuizAction(request QuizActionRequest) (QuizActionResult, error) {
	e.request = request

	return e.result, e.err
}

type fakeProjectMetadataExecutor struct {
	saved chc.ProjectMetadata
}

func (e *fakeProjectMetadataExecutor) SaveProjectMetadata(
	metadata chc.ProjectMetadata,
) (ProjectMetadataResult, error) {
	e.saved = metadata

	return ProjectMetadataResult{
		Message:  "Project metadata saved.",
		Metadata: metadata,
	}, nil
}

type fakeCaseMetadataExecutor struct {
	saved chc.CaseMetadata
}

func (e *fakeCaseMetadataExecutor) SaveCaseMetadata(
	metadata chc.CaseMetadata,
) (CaseMetadataResult, error) {
	e.saved = metadata

	e.saved.SchemaVersion = chc.CaseMetadataSchemaVersion
	if e.saved.Slug == "" {
		e.saved.Slug = "blog-publishing"
	}

	e.saved.Exists = true

	return CaseMetadataResult{
		Message: "Case metadata saved.",
		Case:    e.saved,
		Cases:   []chc.CaseMetadata{e.saved},
	}, nil
}

type fakeTicketActionExecutor struct {
	request TicketActionRequest
	result  TicketActionResult
	err     error
}

func (e *fakeTicketActionExecutor) ExecuteTicketAction(request TicketActionRequest) (TicketActionResult, error) {
	e.request = request

	return e.result, e.err
}

type fakeWorkActionExecutor struct {
	request WorkActionRequest
	result  WorkActionResult
	err     error
}

type fakeWorkspaceRefreshExecutor struct {
	called bool
	result WorkspaceRefreshResult
	err    error
}

func (e *fakeWorkspaceRefreshExecutor) RefreshWorkspace() (WorkspaceRefreshResult, error) {
	e.called = true

	return e.result, e.err
}

func (e *fakeWorkActionExecutor) ExecuteWorkAction(request WorkActionRequest) (WorkActionResult, error) {
	e.request = request

	return e.result, e.err
}

func assertConfirmedArtifactState(t *testing.T, model Model) {
	t.Helper()

	if len(model.tickets) != 1 || model.tickets[0].Key != "CHC-1" {
		t.Fatalf("tickets = %#v, want reloaded CHC-1 ticket", model.tickets)
	}

	if len(model.documents) != 1 || model.documents[0].Title != "TUI Plan" {
		t.Fatalf("documents = %#v, want reloaded TUI Plan document", model.documents)
	}

	if len(model.authoringSessions) != 1 || model.authoringSessions[0].ID != "author-20260522-100000" {
		t.Fatalf("authoringSessions = %#v, want reloaded authoring session", model.authoringSessions)
	}
}
