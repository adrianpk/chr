package tui

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

	"forge.adrianpk.com/chr/internal/core"
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

func TestMouseClickSelectsDumpTextPane(t *testing.T) {
	model := NewModel(Options{
		Dumps: []DumpView{
			{Title: "Project Context", RelativePath: ".chr/default/dumps/project-context.md", Preview: "# Project Context"},
		},
	})
	model.selected = 2
	model.width = 100
	model.height = 30

	updated, _ := model.Update(tea.MouseMsg{
		X:      70,
		Y:      10,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	clicked := requireUpdatedModel(t, updated)

	if clicked.selectedPane != 1 {
		t.Fatalf("selectedPane = %d, want text pane", clicked.selectedPane)
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
		ProjectMetadata: core.ProjectMetadata{
			SchemaVersion: core.ProjectMetadataSchemaVersion,
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
			Graph: core.WorkspaceGraph{
				Project: &core.WorkspaceNode{
					Name:        "Graph Project",
					Description: "Graph description.",
					Path:        ".chr/default/project.md",
					SourceHash:  "hash-project",
				},
			},
		},
	}
	model := NewModel(Options{
		ProjectMetadata: core.ProjectMetadata{
			SchemaVersion: core.ProjectMetadataSchemaVersion,
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
		ProjectMetadata: core.ProjectMetadata{
			SchemaVersion: core.ProjectMetadataSchemaVersion,
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
		ProjectMetadata: core.ProjectMetadata{
			SchemaVersion: core.ProjectMetadataSchemaVersion,
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
			Graph: core.WorkspaceGraph{
				Cases: []core.WorkspaceNode{
					{
						Slug:        "blog-publishing",
						Name:        "Graph Case",
						Description: "Graph case description.",
						Path:        ".chr/default/cases/blog-publishing.md",
						SourceHash:  "hash-case",
					},
				},
				CaseDocuments: map[string][]core.WorkspaceNode{},
			},
		},
	}
	model := NewModel(Options{
		Cases: []core.CaseMetadata{
			{
				SchemaVersion: core.CaseMetadataSchemaVersion,
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
		Cases: []core.CaseMetadata{
			{SchemaVersion: core.CaseMetadataSchemaVersion, Slug: "blog", Name: "Blog Publishing"},
			{SchemaVersion: core.CaseMetadataSchemaVersion, Slug: "runtime", Name: "Runtime Observation"},
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
		Cases: []core.CaseMetadata{
			{SchemaVersion: core.CaseMetadataSchemaVersion, Slug: "blog", Name: "Blog Publishing"},
			{SchemaVersion: core.CaseMetadataSchemaVersion, Slug: "runtime", Name: "Runtime Observation"},
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
			Graph: core.WorkspaceGraph{
				Project: &core.WorkspaceNode{Name: "External"},
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
			Graph: core.WorkspaceGraph{
				Project: &core.WorkspaceNode{Name: "External"},
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
			Graph: core.WorkspaceGraph{
				Cases: []core.WorkspaceNode{
					{Slug: "external", Name: "External Case"},
				},
				CaseDocuments: map[string][]core.WorkspaceNode{},
			},
		},
	}
	model := NewModel(Options{
		WorkspaceExecutor: executor,
	})
	model.cases = []core.CaseMetadata{{Slug: "blog", Name: "Blog"}}
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
		ProjectMetadata: core.ProjectMetadata{
			Name:        "Fallback",
			Description: "Fallback description.",
		},
		Cases: []core.CaseMetadata{
			{
				Slug:        "fallback",
				Name:        "Fallback Case",
				Description: "Fallback case description.",
				Context:     "Preserved context.",
			},
		},
		WorkspaceGraph: core.WorkspaceGraph{
			Project: &core.WorkspaceNode{
				Name:        "Graph Project",
				Description: "Graph description.",
				Path:        ".chr/default/project.md",
				SourceHash:  "hash-project",
			},
			Cases: []core.WorkspaceNode{
				{
					Slug:        "fallback",
					Name:        "Graph Case",
					Description: "Graph case description.",
					Path:        ".chr/default/cases/fallback.md",
					SourceHash:  "hash-case",
				},
			},
			CaseDocuments: map[string][]core.WorkspaceNode{},
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
	initialGraph := core.WorkspaceGraph{
		ProjectWide: []core.WorkspaceNode{
			{
				Kind:         core.WorkspaceNodeSpec,
				RelativePath: ".chr/default/specs/old.md",
				Title:        "Old",
				SourceHash:   "old",
			},
		},
		CaseDocuments: map[string][]core.WorkspaceNode{},
	}
	refreshedGraph := core.WorkspaceGraph{
		ProjectWide: []core.WorkspaceNode{
			{
				Kind:         core.WorkspaceNodeSpec,
				RelativePath: ".chr/default/specs/new.md",
				Title:        "New",
				SourceHash:   "new",
			},
		},
		CaseDocuments: map[string][]core.WorkspaceNode{},
	}
	executor := &fakeWorkspaceRefreshExecutor{
		result: WorkspaceRefreshResult{
			Graph: refreshedGraph,
			Index: core.NewWorkspaceIndexFromGraph(refreshedGraph),
			SemanticDigestStates: []core.WorkspaceSemanticDigestState{
				{NodeID: "spec:.chr/default/specs/new.md", Status: core.WorkspaceSemanticDigestPending},
			},
		},
	}
	model := NewModel(Options{
		WorkspaceGraph:    initialGraph,
		WorkspaceExecutor: executor,
	})

	model.refreshWorkspace()

	if _, ok := model.workspaceIndex.NodesByID["spec:.chr/default/specs/new.md"]; !ok {
		t.Fatalf("workspaceIndex.NodesByID = %#v, want refreshed spec", model.workspaceIndex.NodesByID)
	}

	if len(model.semanticDigestStates) != 1 {
		t.Fatalf("semanticDigestStates = %#v, want one state", model.semanticDigestStates)
	}
}

func TestModelCountsCaseArtifactsFromWorkspaceGraph(t *testing.T) {
	model := NewModel(Options{
		WorkspaceGraph: core.WorkspaceGraph{
			Cases: []core.WorkspaceNode{
				{Slug: "blog", Name: "Blog"},
			},
			CaseDocuments: map[string][]core.WorkspaceNode{
				"blog": {
					{Kind: core.WorkspaceNodeDump, RelativePath: ".chr/default/dumps/blog.md", CaseSlug: "blog"},
					{Kind: core.WorkspaceNodeSpec, RelativePath: ".chr/default/specs/blog.md", CaseSlug: "blog"},
					{Kind: core.WorkspaceNodeTicket, RelativePath: ".chr/default/tickets/CHR-1.md", CaseSlug: "blog"},
				},
			},
			UnresolvedCaseReferences: []core.WorkspaceCaseReference{
				{
					CaseSlug: "blog",
					References: []core.WorkspaceNode{
						{Kind: core.WorkspaceNodePlan, CaseSlug: "blog"},
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
		WorkspaceGraph: core.WorkspaceGraph{
			UnresolvedCaseReferences: []core.WorkspaceCaseReference{
				{
					CaseSlug: "missing",
					References: []core.WorkspaceNode{
						{Kind: core.WorkspaceNodePlan, CaseSlug: "missing"},
						{Kind: core.WorkspaceNodeSpec, CaseSlug: "missing"},
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
	index := core.WorkspaceIndex{
		NodesByID: map[string]core.WorkspaceNode{
			"case:blog": {
				Kind:        core.WorkspaceNodeCase,
				Slug:        "blog",
				Name:        "Blog Publishing",
				Description: "Publish articles.",
			},
			"case:runtime": {
				Kind:        core.WorkspaceNodeCase,
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
		WorkspaceGraph: core.WorkspaceGraph{
			ProjectWide: []core.WorkspaceNode{
				{Kind: core.WorkspaceNodeSpec, RelativePath: ".chr/default/specs/one.md"},
			},
			CaseDocuments: map[string][]core.WorkspaceNode{},
		},
		SemanticDigestStates: []core.WorkspaceSemanticDigestState{
			{NodeID: "spec:.chr/default/specs/one.md", Status: core.WorkspaceSemanticDigestPending},
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
			Graph: core.WorkspaceGraph{
				Diagnostics: []core.WorkspaceGraphDiagnostic{
					{Path: ".chr/default/cases/broken.md", Kind: core.WorkspaceNodeCase, Message: "invalid"},
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
		Specs: []SpecView{
			{Index: 0, Title: "Spec", RelativePath: ".chr/default/specs/tui.md"},
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

	if navigated.statusMessage != "Selected source path: .chr/default/specs/tui.md" {
		t.Fatalf("statusMessage = %q, want selected source path", navigated.statusMessage)
	}
}

func TestModelPlansArtifactActionCommand(t *testing.T) {
	model := NewModel(Options{
		WorkspaceIndex: core.WorkspaceIndex{
			NodesByID: map[string]core.WorkspaceNode{
				"case:blog": {
					Kind: core.WorkspaceNodeCase,
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

func TestSpecGenerationCommandsOnlyAppearInSpecs(t *testing.T) {
	model := NewModel(Options{})
	model.selected = 3
	model.commandPalette.Filter = "generate"

	if commands := model.filteredCommands(); len(commands) != 0 {
		t.Fatalf("quiz commands include spec generation actions: %#v", commands)
	}

	model.selected = 4

	commands := model.filteredCommands()
	if len(commands) != 2 {
		t.Fatalf("spec commands = %#v, want two generate commands", commands)
	}

	if commands[0].ID != "generate-spec" || commands[1].ID != "generate-acceptance-feature" {
		t.Fatalf("spec commands = %#v, want spec and acceptance feature generation", commands)
	}
}

func TestPlanCommandsOnlyAppearInPlans(t *testing.T) {
	model := NewModel(Options{})
	model.selected = 4
	model.commandPalette.Filter = "regenerate-plan"

	if commands := model.filteredCommands(); len(commands) != 0 {
		t.Fatalf("spec commands include plan actions: %#v", commands)
	}

	model.selected = 5

	commands := model.filteredCommands()
	if len(commands) != 1 || commands[0].ID != "regenerate-plan" || commands[0].Plan != PlanActionRegenerate {
		t.Fatalf("plan commands = %#v, want regenerate-plan", commands)
	}

	commandsByID := map[string]Command{}
	for _, command := range DefaultCommandRegistry() {
		commandsByID[command.ID] = command
	}

	for id, action := range map[string]PlanActionID{
		"generate-plan":   PlanActionGenerate,
		"regenerate-plan": PlanActionRegenerate,
		"refresh-plans":   PlanActionRefresh,
	} {
		command, ok := commandsByID[id]
		if !ok {
			t.Fatalf("command %q not found", id)
		}

		if command.Action != CommandActionPlanPlan || command.Plan != action {
			t.Fatalf("command %q = %#v, want plan action %q", id, command, action)
		}
	}
}

func TestDefaultCommandRegistryExposesDumpWorkflows(t *testing.T) {
	commandsByID := map[string]Command{}
	for _, command := range DefaultCommandRegistry() {
		commandsByID[command.ID] = command
	}

	for _, id := range []string{"create-dump", "edit-dump", "refresh-dumps"} {
		command, ok := commandsByID[id]
		if !ok {
			t.Fatalf("command %q not found", id)
		}

		if command.Action != CommandActionPlanDump {
			t.Fatalf("command %q action = %q, want %q", id, command.Action, CommandActionPlanDump)
		}
	}
}

func TestModelRefreshesDumpsFromDumpsKey(t *testing.T) {
	executor := &fakeDumpActionExecutor{
		result: DumpActionResult{
			Message: "Dumps refreshed.",
			Dumps: []DumpView{
				{Title: "First", RelativePath: ".chr/default/dumps/first.md"},
				{Title: "Second", RelativePath: ".chr/default/dumps/second.md"},
			},
			SelectedPath: ".chr/default/dumps/second.md",
		},
	}
	model := NewModel(Options{
		DumpExecutor: executor,
		Dumps: []DumpView{
			{Title: "Second", RelativePath: ".chr/default/dumps/second.md"},
		},
	})
	model.selected = 2

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	refreshed := requireUpdatedModel(t, updated)

	if cmd != nil {
		t.Fatal("cmd is non-nil, want synchronous dump refresh")
	}

	if executor.request.Action != DumpActionRefresh {
		t.Fatalf("dump action = %q, want %q", executor.request.Action, DumpActionRefresh)
	}

	if refreshed.statusMessage != "Dumps refreshed." {
		t.Fatalf("statusMessage = %q, want Dumps refreshed.", refreshed.statusMessage)
	}

	dump, ok := refreshed.selectedDumpView()
	if !ok {
		t.Fatal("selectedDumpView returned no dump")
	}

	if dump.RelativePath != ".chr/default/dumps/second.md" {
		t.Fatalf("selected dump path = %q", dump.RelativePath)
	}
}

func TestModelRefreshesEmptyDumpsFromDumpsKey(t *testing.T) {
	executor := &fakeDumpActionExecutor{
		result: DumpActionResult{
			Message: "Dumps refreshed.",
			Dumps:   []DumpView{},
		},
	}
	model := NewModel(Options{DumpExecutor: executor})
	model.selected = 2

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	refreshed := requireUpdatedModel(t, updated)

	if cmd != nil {
		t.Fatal("cmd is non-nil, want synchronous dump refresh")
	}

	if executor.request.Action != DumpActionRefresh {
		t.Fatalf("dump action = %q, want %q", executor.request.Action, DumpActionRefresh)
	}

	if executor.request.Path != "" {
		t.Fatalf("dump refresh path = %q, want empty path", executor.request.Path)
	}

	if refreshed.statusMessage != "Dumps refreshed." {
		t.Fatalf("statusMessage = %q, want Dumps refreshed.", refreshed.statusMessage)
	}
}

func TestModelRendersDocumentRefreshHints(t *testing.T) {
	model := NewModel(Options{})

	if hint := model.dumpActionHint(); !strings.Contains(hint, "r refresh") {
		t.Fatalf("dumpActionHint() = %q, want refresh hint", hint)
	}

	if hint := model.quizActionHint(QuizView{}); !strings.Contains(hint, "r refresh") {
		t.Fatalf("quizActionHint() = %q, want refresh hint", hint)
	}

	if hint := model.specActionHint(); !strings.Contains(hint, "r refresh") {
		t.Fatalf("specActionHint() = %q, want refresh hint", hint)
	}

	if hint := model.planActionHint(); !strings.Contains(hint, "r refresh") {
		t.Fatalf("planActionHint() = %q, want refresh hint", hint)
	}

	model.selected = 4
	if keys := model.renderKeys(); !strings.Contains(keys, "r refreshes") {
		t.Fatalf("spec renderKeys() = %q, want refresh key", keys)
	}

	model.selected = 5
	if keys := model.renderKeys(); !strings.Contains(keys, "r refreshes") {
		t.Fatalf("plan renderKeys() = %q, want refresh key", keys)
	}
}

func TestModelPlansCreateDumpCommandWithCaseScope(t *testing.T) {
	model := NewModel(Options{
		Cases: []core.CaseMetadata{
			{Slug: "blog", Name: "Blog"},
		},
		WorkspaceIndex: core.WorkspaceIndex{
			NodesByID: map[string]core.WorkspaceNode{
				"case:blog": {
					Kind: core.WorkspaceNodeCase,
					Slug: "blog",
					Name: "Blog",
				},
			},
			CasesBySlug: map[string]string{"blog": "case:blog"},
		},
	})
	model.cases = []core.CaseMetadata{{Slug: "blog", Name: "Blog"}}
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
		"Use Case: blog",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelConfirmsDumpAction(t *testing.T) {
	executor := &fakeDumpActionExecutor{
		result: DumpActionResult{
			Message: "Created dump .chr/default/dumps/project-dump.md.",
			Dumps: []DumpView{
				{Title: "Project dump", RelativePath: ".chr/default/dumps/project-dump.md"},
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
		"c create",
		"r refresh",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelPlansEditDumpCommand(t *testing.T) {
	model := NewModel(Options{
		Dumps: []DumpView{
			{Title: "Project dump", RelativePath: ".chr/default/dumps/project-dump.md"},
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

	if planned.dumpAction.Request.Path != ".chr/default/dumps/project-dump.md" {
		t.Fatalf("dump path = %q", planned.dumpAction.Request.Path)
	}
}

func TestModelEditsSelectedDumpDirectlyWithEnter(t *testing.T) {
	executor := &fakeDumpActionExecutor{
		result: DumpActionResult{
			Message:      "Editing dump .chr/default/dumps/project-dump.md.",
			Content:      "# Project dump\n\nExisting Markdown.",
			SelectedPath: ".chr/default/dumps/project-dump.md",
		},
	}
	model := NewModel(Options{
		DumpExecutor: executor,
		Dumps: []DumpView{
			{Title: "Project dump", RelativePath: ".chr/default/dumps/project-dump.md"},
		},
	})
	model.selected = 2
	model.selectedPane = 1

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	confirmed := requireUpdatedModel(t, updated)

	if executor.request.Action != DumpActionEdit {
		t.Fatalf("dump action = %q, want %q", executor.request.Action, DumpActionEdit)
	}

	if !confirmed.dumpEdit.Active {
		t.Fatal("dumpEdit.Active = false, want active edit")
	}

	if confirmed.dumpAction.Pending {
		t.Fatal("dump action is pending after direct edit")
	}

	if confirmed.selectedPane != 1 {
		t.Fatalf("selectedPane = %d, want 1", confirmed.selectedPane)
	}

	if confirmed.maximizedPane != 1 {
		t.Fatalf("maximizedPane = %d, want 1", confirmed.maximizedPane)
	}
}

func TestModelConfirmsAppendDumpAction(t *testing.T) {
	executor := &fakeDumpActionExecutor{
		result: DumpActionResult{
			Message: "Appended dump .chr/default/dumps/project-dump.md.",
			Dumps: []DumpView{
				{Title: "Project dump", RelativePath: ".chr/default/dumps/project-dump.md"},
			},
		},
	}
	model := NewModel(Options{DumpExecutor: executor})
	model.dumpAction = DumpActionState{
		Pending: true,
		Request: DumpActionRequest{
			Action:  DumpActionAppend,
			Path:    ".chr/default/dumps/project-dump.md",
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
			Message:      "Editing dump .chr/default/dumps/project-dump.md.",
			Content:      "# Project dump\n\nExisting Markdown.",
			SelectedPath: ".chr/default/dumps/project-dump.md",
		},
	}
	model := NewModel(Options{DumpExecutor: executor})
	model.dumpAction = DumpActionState{
		Pending: true,
		Request: DumpActionRequest{
			Action: DumpActionEdit,
			Path:   ".chr/default/dumps/project-dump.md",
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

	if confirmed.dumpEdit.Path != ".chr/default/dumps/project-dump.md" {
		t.Fatalf("dumpEdit.Path = %q", confirmed.dumpEdit.Path)
	}

	if confirmed.dumpEdit.Input.Value() != "# Project dump\n\nExisting Markdown." {
		t.Fatalf("dumpEdit value = %q", confirmed.dumpEdit.Input.Value())
	}

	if confirmed.statusMessage != "Editing .chr/default/dumps/project-dump.md" {
		t.Fatalf("statusMessage = %q", confirmed.statusMessage)
	}
}

func TestModelSavesInternalDumpEdit(t *testing.T) {
	executor := &fakeDumpActionExecutor{
		result: DumpActionResult{
			Message: "Saved dump .chr/default/dumps/project-dump.md.",
			Dumps: []DumpView{
				{Title: "Project dump", RelativePath: ".chr/default/dumps/project-dump.md"},
			},
			SelectedPath: ".chr/default/dumps/project-dump.md",
		},
	}
	model := NewModel(Options{DumpExecutor: executor})
	model.startDumpEdit(".chr/default/dumps/project-dump.md", "# Project dump\n\nChanged Markdown.", "")

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

func TestModelRendersDumpEditPanelWithMinimalHint(t *testing.T) {
	model := NewModel(Options{
		Dumps: []DumpView{
			{
				Title:        "Project dump",
				Status:       "raw",
				RelativePath: ".chr/default/dumps/project-dump.md",
			},
		},
	})
	model.selected = 2
	model.selectedPane = 1
	model.width = 100
	model.height = 24
	model.startDumpEdit(".chr/default/dumps/project-dump.md", "# Project dump\n\nChanged Markdown.", "")

	view := model.View()

	for _, want := range []string{
		"Edit",
		"ctrl+s save  esc cancel  alt+e external",
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
			{Title: "Project dump", RelativePath: ".chr/default/dumps/project-dump.md"},
		},
	})
	model.startDumpEdit(".chr/default/dumps/project-dump.md", "# Project dump", "")

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

	if dump.RelativePath != ".chr/default/dumps/project-dump.md" {
		t.Fatalf("selected dump path = %q", dump.RelativePath)
	}
}

func TestModelAltELaunchesExternalDumpEditor(t *testing.T) {
	executor := &fakeDumpActionExecutor{
		result: DumpActionResult{
			Message:                  "Editing dump .chr/default/dumps/project-dump.md.",
			EditorCommand:            exec.Command("true"),
			EditorCommandDescription: "nvim project-dump.md",
		},
	}
	model := NewModel(Options{
		DumpExecutor: executor,
		Dumps: []DumpView{
			{Title: "Project dump", RelativePath: ".chr/default/dumps/project-dump.md"},
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

	wantStatus := "External editor opened. chr will resume when the editor process exits."
	if confirmed.statusMessage != wantStatus {
		t.Fatalf("statusMessage = %q", confirmed.statusMessage)
	}
}

func TestModelAltEOpensSelectedTicketSource(t *testing.T) {
	executor := &fakeSourceOpenExecutor{
		result: SourceOpenResult{
			EditorCommand:            exec.Command("true"),
			EditorCommandDescription: "nvim ticket.md",
		},
	}
	model := NewModel(Options{
		SourceExecutor: executor,
		Tickets: []core.IndexedTicket{
			{
				Key:          "CHR-1",
				Title:        "Preview command",
				Status:       core.TicketStatusBacklog,
				RelativePath: "local-preview/CHR-1-preview-command.md",
			},
		},
	})
	model.selected = 6

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}, Alt: true})

	confirmed, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if cmd == nil {
		t.Fatal("cmd = nil, want external editor command")
	}

	if executor.path != ".chr/default/tickets/local-preview/CHR-1-preview-command.md" {
		t.Fatalf("source path = %q", executor.path)
	}

	wantStatus := "External editor opened. chr will resume when the editor process exits."
	if confirmed.statusMessage != wantStatus {
		t.Fatalf("statusMessage = %q", confirmed.statusMessage)
	}
}

func TestModelCtrlEOpensSelectedTicketSource(t *testing.T) {
	executor := &fakeSourceOpenExecutor{
		result: SourceOpenResult{
			EditorCommand:            exec.Command("true"),
			EditorCommandDescription: "nvim ticket.md",
		},
	}
	model := NewModel(Options{
		SourceExecutor: executor,
		Tickets: []core.IndexedTicket{
			{
				Key:          "CHR-1",
				Title:        "Preview command",
				Status:       core.TicketStatusBacklog,
				RelativePath: "local-preview/CHR-1-preview-command.md",
			},
		},
	})
	model.selected = 6

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlE})

	if cmd == nil {
		t.Fatal("cmd = nil, want external editor command")
	}

	if executor.path != ".chr/default/tickets/local-preview/CHR-1-preview-command.md" {
		t.Fatalf("source path = %q", executor.path)
	}
}

func TestModelEOpensSelectedSourceWhenDetailPaneFocused(t *testing.T) {
	executor := &fakeSourceOpenExecutor{
		result: SourceOpenResult{
			EditorCommand:            exec.Command("true"),
			EditorCommandDescription: "nvim ticket.md",
		},
	}
	model := NewModel(Options{
		SourceExecutor: executor,
		Tickets: []core.IndexedTicket{
			{
				Key:          "CHR-1",
				Title:        "Preview command",
				Status:       core.TicketStatusBacklog,
				RelativePath: "CHR-1-preview-command.md",
			},
		},
	})
	model.selected = 6
	model.selectedPane = 1

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	if cmd == nil {
		t.Fatal("cmd = nil, want external editor command")
	}

	if executor.path != ".chr/default/tickets/CHR-1-preview-command.md" {
		t.Fatalf("source path = %q", executor.path)
	}
}

func TestModelPageDownScrollsTicketDetailWithoutFocus(t *testing.T) {
	model := NewModel(Options{
		Tickets: []core.IndexedTicket{
			{
				Key:          "CHR-1",
				Title:        "Long detail",
				Status:       core.TicketStatusBacklog,
				RelativePath: "CHR-1-long-detail.md",
			},
		},
	})
	model.selected = 6
	model.height = 20

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyPgDown})

	scrolled, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if scrolled.detailScroll != 15 {
		t.Fatalf("detailScroll = %d, want 15", scrolled.detailScroll)
	}
}

func TestModelPageDownChangesVisibleTicketDetail(t *testing.T) {
	model := NewModel(Options{
		Tickets: []core.IndexedTicket{
			{
				Key:          "CHR-1",
				Title:        "Long detail",
				Status:       core.TicketStatusBacklog,
				RelativePath: "CHR-1-long-detail.md",
			},
		},
	})
	model.selected = 6
	model.width = 100
	model.height = 18
	model.tickets[0].BodyLines = []string{
		"Line 1",
		"Line 2",
		"Line 3",
		"Line 4",
		"Line 5",
		"Line 6",
		"Line 7",
		"Line 8",
		"Line 9",
		"Line 10",
		"Line 11",
		"Line 12",
		"Line 13",
		"Line 14",
		"Line 15",
	}

	before := model.View()
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyPgDown})

	scrolled, ok := updated.(Model)
	if !ok {
		t.Fatalf("updated model = %T, want Model", updated)
	}

	after := scrolled.View()

	if before == after {
		t.Fatal("View() did not change after PgDown")
	}

	if strings.Contains(after, "Line 1 ") {
		t.Fatalf("View() still contains first body line after PgDown:\n%s", after)
	}
}

func TestModelScrollsFocusedTicketDetail(t *testing.T) {
	model := NewModel(Options{
		Tickets: []core.IndexedTicket{
			{
				Key:          "CHR-1",
				Title:        "Long detail",
				Status:       core.TicketStatusBacklog,
				RelativePath: "CHR-1-long-detail.md",
			},
		},
	})
	model.selected = 6
	model.selectedPane = 1
	model.tickets[0].BodyLines = []string{
		"Line 1",
		"Line 2",
		"Line 3",
		"Line 4",
		"Line 5",
		"Line 6",
		"Line 7",
		"Line 8",
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})

	scrolled, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if scrolled.detailScroll != 1 {
		t.Fatalf("detailScroll = %d, want 1", scrolled.detailScroll)
	}
}

func TestModelReportsExternalEditorFailureContext(t *testing.T) {
	model := NewModel(Options{})

	updated, _ := model.Update(dumpEditorFinishedMsg{
		err:        errors.New("exit status 135"),
		path:       ".chr/default/dumps/project-dump.md",
		command:    "nvim project-dump.md",
		workingDir: "/workspace/.chr/default/dumps",
	})

	failed, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	for _, want := range []string{
		"Edit failed:",
		"command=nvim project-dump.md",
		"dir=/workspace/.chr/default/dumps",
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
			Message: "Edited dump .chr/default/dumps/second.md.",
			Dumps: []DumpView{
				{Title: "First", RelativePath: ".chr/default/dumps/first.md"},
				{Title: "Second", RelativePath: ".chr/default/dumps/second.md"},
			},
			SelectedPath: ".chr/default/dumps/second.md",
		},
	}
	model := NewModel(Options{
		DumpExecutor: executor,
		Dumps: []DumpView{
			{Title: "Second", RelativePath: ".chr/default/dumps/second.md"},
		},
	})

	updated, _ := model.Update(dumpEditorFinishedMsg{path: ".chr/default/dumps/second.md"})

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

	if dump.RelativePath != ".chr/default/dumps/second.md" {
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
			Message: "Created ticket CHR-1 at .chr/default/tickets/CHR-1-tui-created-ticket.md",
			Tickets: []core.IndexedTicket{
				{Key: "CHR-1", Title: "TUI created ticket", Status: "open"},
			},
			Documents: []DocumentView{
				{Index: 0, Title: "TUI Plan"},
			},
			AuthoringSessions: []core.AuthoringSession{
				{
					ID:        "author-20260522-100000",
					ProjectID: "project-1",
					Status:    core.AuthoringSessionStatusDraft,
					Mode:      core.AuthoringSessionModeTUI,
					Target:    core.AuthoringTargetSpec,
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
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "First", Status: "open"},
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

	if planned.ticketAction.Request.Key != "CHR-1" {
		t.Fatalf("ticket action key = %q, want CHR-1", planned.ticketAction.Request.Key)
	}

	if planned.ticketAction.Request.Status != "ready" {
		t.Fatalf("ticket action status = %q, want ready", planned.ticketAction.Request.Status)
	}

	view := planned.View()
	for _, want := range []string{
		"Confirm Ticket Action",
		"Move ticket status",
		"Ticket: CHR-1",
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
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "First", Status: "open"},
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
			Message: "Moved ticket CHR-1 to ready.",
			Tickets: []core.IndexedTicket{
				{Key: "CHR-1", Title: "First", Status: "ready"},
			},
		},
	}
	model := NewModel(Options{
		TicketExecutor: executor,
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "First", Status: "open"},
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

	if executor.request.Key != "CHR-1" || executor.request.Status != "ready" {
		t.Fatalf("executor request = %#v, want CHR-1 ready", executor.request)
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
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "First", Status: "open"},
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
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "First", Status: "ready"},
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

	if planned.workAction.Request.TicketKey != "CHR-1" {
		t.Fatalf("work action ticket key = %q, want CHR-1", planned.workAction.Request.TicketKey)
	}

	view := planned.View()
	for _, want := range []string{
		"Confirm Work Action",
		"Start selected ticket",
		"Ticket: CHR-1",
		"Enter: confirm",
		"Esc: cancel",
		"enter confirms work action",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelRendersDashAgentWorkPanel(t *testing.T) {
	model := NewModel(Options{
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "Ready frontend", Status: "ready_for_dev", Priority: "high"},
			{Key: "CHR-2", Title: "Backlog backend", Status: "backlog", Priority: "normal"},
		},
		Observability: ObservabilityView{
			LatestExecution: ExecutionView{ID: "exec-1", TicketKey: "CHR-1", Status: "succeeded"},
			LatestEvent:     AgentEventView{Phase: "merge_finished", Message: "CHR-1 merge finished: merged"},
			LatestLoop: LoopRunView{
				ID:             "loop-1",
				Status:         "succeeded",
				EligibleCount:  1,
				ProcessedCount: 1,
			},
		},
	})
	model.selected = 7

	view := model.View()
	for _, want := range []string{
		"Agent Work",
		"Queue",
		"CHR-1",
		"Ready frontend",
		"Current",
		"Loop: loop-1",
		"Phase: merge_finished",
		"Last event: CHR-1 merge finished: merged",
		"Execution: exec-1",
		"Recent",
		"a Run all ready tickets",
		"Run all ready tickets",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}

	if strings.Contains(view, "Backlog backend") {
		t.Fatalf("View() includes backlog ticket in ready queue:\n%s", view)
	}
}

func TestModelRendersDashWorkReadinessGuidance(t *testing.T) {
	model := NewModel(Options{
		WorkReadiness: WorkReadinessView{
			Status:            "blocked",
			CurrentBranch:     "feature/current",
			IntegrationBranch: "dev",
			Diagnostics: []WorkReadinessDiagnosticView{
				{
					Code:            "wrong_branch",
					Detail:          "expected dev but found feature/current",
					RepairCode:      "switch_branch",
					RepairAvailable: true,
					RepairSafe:      true,
				},
			},
		},
	})
	model.selected = 7

	view := model.View()
	for _, want := range []string{
		"Readiness",
		"Status: blocked",
		"Current branch: feature/current",
		"Expected branch: dev",
		"wrong_branch: expected dev but found feature/current",
		"Repair: x apply safe repair",
		"x Apply safe readiness repair",
		"x repairs readiness",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelDashRunAllKeyPlansWorkAction(t *testing.T) {
	model := NewModel(Options{})
	model.selected = 7

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})

	planned, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if !planned.workAction.Pending {
		t.Fatal("work action is not pending")
	}

	if planned.workAction.Request.Action != WorkActionRunAllReady {
		t.Fatalf("work action = %q, want %q", planned.workAction.Request.Action, WorkActionRunAllReady)
	}

	view := planned.View()
	for _, want := range []string{
		"Confirm Work Action",
		"Run all ready tickets",
		"Status filter: ready_for_dev",
		"Priority filter: all",
		"Order: oldest first",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelDashRepairReadinessKeyPlansWorkAction(t *testing.T) {
	model := NewModel(Options{
		WorkReadiness: WorkReadinessView{
			Status: "blocked",
			Diagnostics: []WorkReadinessDiagnosticView{
				{
					Code:            "runtime_artifacts_visible",
					RepairCode:      "ignore_runtime_artifacts",
					RepairAvailable: true,
					RepairSafe:      true,
				},
			},
		},
	})
	model.selected = 7

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	planned := requireUpdatedModel(t, updated)

	if !planned.workAction.Pending {
		t.Fatal("work action is not pending")
	}

	if planned.workAction.Request.Action != WorkActionRepairReadiness {
		t.Fatalf("work action = %q, want %q", planned.workAction.Request.Action, WorkActionRepairReadiness)
	}

	view := planned.View()
	for _, want := range []string{
		"Confirm Work Action",
		"Apply safe readiness repair",
		"Does not start agent work automatically.",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelDashCyclesAgentWorkFilters(t *testing.T) {
	model := NewModel(Options{
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "Ready normal", Status: "ready_for_dev", Priority: "normal", Severity: "normal"},
			{Key: "CHR-2", Title: "Ready critical", Status: "ready_for_dev", Priority: "critical", Severity: "critical"},
		},
	})
	model.selected = 7

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})

	filtered, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	view := filtered.View()
	for _, want := range []string{
		"Priority: critical",
		"CHR-2",
		"Ready critical",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}

	if strings.Contains(view, "Ready normal") {
		t.Fatalf("View() includes filtered ticket:\n%s", view)
	}
}

func TestModelDashRunAllUsesAgentWorkFilters(t *testing.T) {
	model := NewModel(Options{})
	model.selected = 7

	priorityChanged, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	orderChanged, _ := priorityChanged.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	planned, _ := orderChanged.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})

	modelWithAction, ok := planned.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", planned)
	}

	request := modelWithAction.workAction.Request
	if request.PriorityFilter != "critical" {
		t.Fatalf("PriorityFilter = %q, want critical", request.PriorityFilter)
	}

	if request.Order != core.TicketLoopOrderPriority {
		t.Fatalf("Order = %q, want priority", request.Order)
	}

	view := modelWithAction.View()
	for _, want := range []string{
		"Priority filter: critical",
		"Order: priority",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelRefreshesDashAgentWorkOnTick(t *testing.T) {
	executor := &fakeAgentWorkRefreshExecutor{
		result: AgentWorkRefreshResult{
			Tickets: []core.IndexedTicket{
				{Key: "CHR-2", Title: "Fresh ready", Status: "ready_for_dev"},
			},
			Observability: ObservabilityView{
				LatestExecution: ExecutionView{ID: "exec-2", TicketKey: "CHR-2", Status: "running"},
				LatestLoop:      LoopRunView{ID: "loop-2", Status: "running", EligibleCount: 1},
			},
			Readiness: WorkReadinessView{Status: "ready", CurrentBranch: "dev"},
		},
	}
	model := NewModel(Options{
		AgentWorkExecutor: executor,
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "Old ready", Status: "ready_for_dev"},
		},
	})
	model.selected = 7

	updated, cmd := model.Update(agentWorkRefreshTickMsg{})
	refreshing := requireUpdatedModel(t, updated)

	if cmd == nil {
		t.Fatal("refresh command is nil")
	}

	if !executor.called {
		cmd()
	}

	updated, _ = refreshing.Update(agentWorkRefreshFinishedMsg{result: executor.result})
	refreshed := requireUpdatedModel(t, updated)

	if len(refreshed.tickets) != 1 || refreshed.tickets[0].Key != "CHR-2" {
		t.Fatalf("tickets = %#v, want refreshed CHR-2", refreshed.tickets)
	}

	if refreshed.observability.LatestExecution.ID != "exec-2" {
		t.Fatalf("LatestExecution.ID = %q, want exec-2", refreshed.observability.LatestExecution.ID)
	}

	if refreshed.workReadiness.Status != "ready" {
		t.Fatalf("workReadiness.Status = %q, want ready", refreshed.workReadiness.Status)
	}
}

func TestModelRefreshesDashAgentWorkWhileWorkBusy(t *testing.T) {
	executor := &fakeAgentWorkRefreshExecutor{
		result: AgentWorkRefreshResult{
			Observability: ObservabilityView{
				RecentEvents: []AgentEventView{
					{Time: "12:00:00", TicketKey: "CHR-2", Message: "worker started"},
				},
			},
		},
	}
	model := NewModel(Options{AgentWorkExecutor: executor})
	model.selected = 7
	model.workBusy = true
	model.statusMessage = "Running all ready tickets..."

	updated, cmd := model.Update(agentWorkRefreshTickMsg{})
	refreshing := requireUpdatedModel(t, updated)

	if cmd == nil {
		t.Fatal("refresh command is nil")
	}

	if !refreshing.agentWorkRefreshInFlight {
		t.Fatal("agentWorkRefreshInFlight = false, want refresh in flight")
	}

	updated, _ = refreshing.Update(agentWorkRefreshFinishedMsg{result: executor.result})
	refreshed := requireUpdatedModel(t, updated)

	if len(refreshed.observability.RecentEvents) != 1 {
		t.Fatalf("RecentEvents length = %d, want 1", len(refreshed.observability.RecentEvents))
	}

	if refreshed.observability.RecentEvents[0].Message != "worker started" {
		t.Fatalf("RecentEvents[0].Message = %q, want worker started", refreshed.observability.RecentEvents[0].Message)
	}

	if !refreshed.workBusy {
		t.Fatal("workBusy = false, want foreground action retained during refresh")
	}
}

func TestModelSkipsDashAgentWorkRefreshWhenOneIsInFlight(t *testing.T) {
	executor := &fakeAgentWorkRefreshExecutor{}
	model := NewModel(Options{AgentWorkExecutor: executor})
	model.selected = 7

	updated, cmd := model.Update(agentWorkRefreshTickMsg{})
	refreshing := requireUpdatedModel(t, updated)

	if cmd == nil {
		t.Fatal("refresh command is nil")
	}

	updated, _ = refreshing.Update(agentWorkRefreshTickMsg{})
	stillRefreshing := requireUpdatedModel(t, updated)

	if !stillRefreshing.agentWorkRefreshInFlight {
		t.Fatal("agentWorkRefreshInFlight is false, want in-flight refresh retained")
	}

	if executor.called {
		t.Fatal("RefreshAgentWork called before refresh command runs")
	}
}

func TestModelDoesNotRefreshAgentWorkOutsideDash(t *testing.T) {
	executor := &fakeAgentWorkRefreshExecutor{}
	model := NewModel(Options{AgentWorkExecutor: executor})

	_, cmd := model.Update(agentWorkRefreshTickMsg{})
	if cmd == nil {
		t.Fatal("tick command is nil")
	}

	if executor.called {
		t.Fatal("RefreshAgentWork called outside Dash")
	}
}

func TestModelRendersDashRunningFailedAndBlockedStates(t *testing.T) {
	for _, testCase := range []struct {
		name          string
		observability ObservabilityView
		want          []string
	}{
		{
			name: "running",
			observability: ObservabilityView{
				LatestExecution: ExecutionView{ID: "exec-running", TicketKey: "CHR-1", Status: "running"},
				LatestLoop:      LoopRunView{ID: "loop-running", Status: "running", EligibleCount: 2, ProcessedCount: 1},
			},
			want: []string{"Status: running", "Execution status: running", "Loop loop-running  running"},
		},
		{
			name: "failed",
			observability: ObservabilityView{
				LatestExecution: ExecutionView{ID: "exec-failed", TicketKey: "CHR-2", Status: "failed"},
				LatestLoop:      LoopRunView{ID: "loop-failed", Status: "failed", EligibleCount: 2, FailedCount: 1},
			},
			want: []string{"Status: failed", "Execution status: failed", "Loop loop-failed  failed"},
		},
		{
			name: "blocked",
			observability: ObservabilityView{
				LatestExecution: ExecutionView{ID: "exec-blocked", TicketKey: "CHR-3", Status: "succeeded"},
				LatestLoop:      LoopRunView{ID: "loop-blocked", Status: "blocked", EligibleCount: 2, BlockedCount: 1},
			},
			want: []string{"Status: blocked", "Execution status: succeeded", "Loop loop-blocked  blocked"},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			model := NewModel(Options{Observability: testCase.observability})
			model.selected = 7

			view := model.View()
			for _, want := range testCase.want {
				if !strings.Contains(view, want) {
					t.Fatalf("View() missing %q:\n%s", want, view)
				}
			}
		})
	}
}

func TestModelCancelsWorkAction(t *testing.T) {
	model := NewModel(Options{
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "First", Status: "ready"},
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

func TestModelStartsWorkActionOnConfirm(t *testing.T) {
	executor := &fakeWorkActionExecutor{
		result: testSuccessfulWorkActionResult(),
	}
	model := NewModel(Options{
		WorkExecutor: executor,
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "First", Status: "ready"},
		},
	})

	ticket, ok := model.selectedTicketView()
	if !ok {
		t.Fatal("selectedTicketView returned no ticket")
	}

	model.workAction = NewWorkActionState(WorkActionStartTicket, ticket)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	busy := requireUpdatedModel(t, updated)

	if cmd == nil {
		t.Fatal("work action command is nil")
	}

	if !busy.workBusy {
		t.Fatal("workBusy = false, want true")
	}

	if busy.statusMessage != "Starting agent work for CHR-1..." {
		t.Fatalf("statusMessage = %q, want starting work status", busy.statusMessage)
	}

	action := busy.renderWorkAction()
	if !strings.Contains(action, "Work Action Running") || !strings.Contains(action, "Navigation is locked until this action finishes.") {
		t.Fatalf("renderWorkAction() = %q, want running overlay", action)
	}

	if !strings.Contains(action, "No agent events recorded yet.") {
		t.Fatalf("renderWorkAction() = %q, want empty event stream state", action)
	}
}

func TestModelBlocksWorkActionWhenReadinessBlocked(t *testing.T) {
	executor := &fakeWorkActionExecutor{}
	model := NewModel(Options{
		WorkExecutor:  executor,
		WorkReadiness: testBlockedWorkReadinessView(),
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "First", Status: "ready_for_dev"},
		},
	})
	model.selected = 7

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	planned := requireUpdatedModel(t, updated)

	updated, cmd := planned.Update(tea.KeyMsg{Type: tea.KeyEnter})
	blocked := requireUpdatedModel(t, updated)

	if cmd != nil {
		t.Fatal("blocked work action returned command")
	}

	if blocked.workBusy {
		t.Fatal("workBusy = true, want blocked before execution")
	}

	if executor.called {
		t.Fatal("executor was called for blocked readiness")
	}

	if blocked.statusMessage != "Work readiness blocked. Press x to apply safe repair." {
		t.Fatalf("statusMessage = %q, want readiness blocker", blocked.statusMessage)
	}
}

func TestModelRunsRepairReadinessAction(t *testing.T) {
	executor := &fakeWorkActionExecutor{
		result: WorkActionResult{
			Message: "Safe readiness repair applied. Retry the work action.",
			Readiness: WorkReadinessView{
				Status: "ready",
				RepairsApplied: []WorkReadinessAppliedRepairView{
					{Code: "ignore_runtime_artifacts", Detail: "excluded .chr runtime artifacts"},
				},
			},
		},
	}
	model := NewModel(Options{
		WorkExecutor:  executor,
		WorkReadiness: testBlockedWorkReadinessView(),
	})
	model.selected = 7

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	planned := requireUpdatedModel(t, updated)

	updated, cmd := planned.Update(tea.KeyMsg{Type: tea.KeyEnter})
	busy := requireUpdatedModel(t, updated)

	if cmd == nil {
		t.Fatal("repair readiness command is nil")
	}

	updated, _ = busy.Update(cmd())
	repaired := requireUpdatedModel(t, updated)

	if executor.request.Action != WorkActionRepairReadiness {
		t.Fatalf("executor action = %q, want %q", executor.request.Action, WorkActionRepairReadiness)
	}

	if repaired.workReadiness.Status != "ready" {
		t.Fatalf("workReadiness.Status = %q, want ready", repaired.workReadiness.Status)
	}

	if repaired.statusMessage != executor.result.Message {
		t.Fatalf("statusMessage = %q, want repair message", repaired.statusMessage)
	}
}

func TestModelRendersWorkActionEventStream(t *testing.T) {
	model := NewModel(Options{})
	model.workBusy = true
	model.statusMessage = "Running all ready tickets..."
	model.observability = ObservabilityView{
		RecentEvents: []AgentEventView{
			{Time: "11:31:00", Level: "info", Message: "loop started"},
			{
				Time:            "11:31:02",
				Level:           "info",
				TicketKey:       "CHR-2",
				ExecutionID:     "exec-2",
				Message:         "worker started",
				TicketPath:      ".chr/default/tickets/CHR-2.md",
				RuntimeLog:      "ops/default/run/00002-ticket-run/evidence/runtime.log",
				ValidationLog:   "ops/default/run/00002-ticket-run/evidence/validation.log",
				ExecutionReport: "ops/default/run/00002-ticket-run/evidence/execution-report.md",
			},
			{Time: "11:31:03", Level: "error", TicketKey: "CHR-2", Message: "validation failed"},
		},
	}

	action := model.renderWorkAction()
	for _, want := range []string{
		"Work Action Running",
		"Running all ready tickets...",
		"Navigation is locked until this action finishes.",
		"Events",
		"11:31:00 loop started",
		"11:31:02 CHR-2 worker started",
		"11:31:03 CHR-2 error: validation failed",
		"Ticket: .chr/default/tickets/CHR-2.md",
		"Execution report: ops/default/run/00002-ticket-run/evidence/execution-report.md",
		"Runtime log: ops/default/run/00002-ticket-run/evidence/runtime.log",
		"Validation log: ops/default/run/00002-ticket-run/evidence/validation.log",
		"Lookup: chr work logs exec-2",
	} {
		if !strings.Contains(action, want) {
			t.Fatalf("renderWorkAction() missing %q:\n%s", want, action)
		}
	}
}

func TestModelBlocksFooterMouseWhileWorkBusy(t *testing.T) {
	model := NewModel(Options{})
	model.width = 100
	model.height = 30
	model.workBusy = true

	updated, cmd := model.Update(tea.MouseMsg{
		X:      95,
		Y:      29,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	blocked := requireUpdatedModel(t, updated)

	if cmd != nil {
		t.Fatal("mouse update returned command while work action is running")
	}

	if blocked.statusMessage != "Work action already running." {
		t.Fatalf("statusMessage = %q, want work busy status", blocked.statusMessage)
	}
}

func TestModelCompletesWorkAction(t *testing.T) {
	executor := &fakeWorkActionExecutor{
		result: testSuccessfulWorkActionResult(),
	}
	model := NewModel(Options{
		WorkExecutor: executor,
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "First", Status: "ready"},
		},
	})

	ticket, ok := model.selectedTicketView()
	if !ok {
		t.Fatal("selectedTicketView returned no ticket")
	}

	model.workAction = NewWorkActionState(WorkActionStartTicket, ticket)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	busy := requireUpdatedModel(t, updated)
	updated, _ = busy.Update(cmd())
	confirmed := requireUpdatedModel(t, updated)

	if executor.request.Action != WorkActionStartTicket || executor.request.TicketKey != "CHR-1" {
		t.Fatalf("executor request = %#v, want start CHR-1", executor.request)
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
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "First", Status: "ready"},
		},
	})

	ticket, ok := model.selectedTicketView()
	if !ok {
		t.Fatal("selectedTicketView returned no ticket")
	}

	model.workAction = NewWorkActionState(WorkActionStartTicket, ticket)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	busy, ok := updated.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	updated, _ = busy.Update(cmd())
	failed := requireUpdatedModel(t, updated)

	if failed.workAction.Pending {
		t.Fatal("work action is pending after failure")
	}

	if failed.statusMessage != "Work action failed: runtime validation failed" {
		t.Fatalf("statusMessage = %q, want work action failure", failed.statusMessage)
	}
}

func TestModelRendersTicketBoardAndDetail(t *testing.T) {
	projectDir := t.TempDir()
	model := NewModel(Options{
		Tickets: ticketBoardTestTickets(t, projectDir),
	})
	model.selected = 6

	view := model.View()

	for _, want := range []string{
		"Board",
		"Backlog (1/6)",
		"CHR-1",
		"Fix publishing",
		"Selected Ticket",
		"Synopsis: Fix the publishing workflow.",
		"Source: plan_tracking",
		"Source Ref: .chr/default/plans/blog-plan.md",
		"Slice: Slice 1",
		"Checklist: 1 done / 2 total",
		"Comments: 1",
		"Path: .chr/default/tickets/open.md",
		"Body:",
		"Investigate publishing failure.",
		"- [x] Done",
		"- [ ] Pending",
		"# Comments",
		"2026-05-22T10:00:00Z chr",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelRendersTicketEvidenceDetail(t *testing.T) {
	projectDir := t.TempDir()
	ticketPath := writeTicketViewFile(t, projectDir, "evidence.md", testTicketEvidenceDetailBody())
	model := NewModel(Options{
		Tickets: []core.IndexedTicket{{
			Path:         ticketPath,
			RelativePath: "evidence.md",
			Key:          "CHR-1",
			Title:        "Add CLI behavior",
			Status:       core.TicketStatusBacklog,
			Type:         "feature",
		}},
	})
	model.selected = 6

	view := model.View()

	for _, want := range []string{
		"Evidence: complete",
		"Evidence strategy: complete (done)",
		"Gherkin: required (done)",
		"Tests: mixed (done)",
		"Validation: passed (done)",
		"Evidence summary: complete (done)",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelRendersFullTicketBodyWithoutPreviewSections(t *testing.T) {
	projectDir := t.TempDir()
	ticketPath := writeTicketViewFile(
		t,
		projectDir,
		"long.md",
		"# Description\n\nFirst paragraph.\n\n    indented code\n\n- [ ] Task\n\n# Comments\n\n## 2026-05-22T10:00:00Z chr\n\nComment body.\n",
	)
	model := NewModel(Options{
		Tickets: []core.IndexedTicket{
			{
				Path:         ticketPath,
				RelativePath: "long.md",
				Key:          "CHR-1",
				Title:        "Long body",
				Status:       core.TicketStatusBacklog,
			},
		},
	})
	model.selected = 6

	view := model.View()

	for _, want := range []string{
		"Body:",
		"# Description",
		"First paragraph.",
		"indented code",
		"- [ ] Task",
		"# Comments",
		"Comment body.",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}

	for _, unwanted := range []string{"Checklist Items:", "Recent Comments:"} {
		if strings.Contains(view, unwanted) {
			t.Fatalf("View() contains preview section %q:\n%s", unwanted, view)
		}
	}
}

func ticketBoardTestTickets(t *testing.T, projectDir string) []core.IndexedTicket {
	t.Helper()

	firstPath := writeTicketViewFile(
		t,
		projectDir,
		"open.md",
		"Investigate publishing failure.\n\n- [x] Done\n- [ ] Pending\n\n# Comments\n\n## 2026-05-22T10:00:00Z chr\n",
	)
	secondPath := writeTicketViewFile(t, projectDir, "ready.md", "- [ ] Ready\n")

	return []core.IndexedTicket{
		{
			Path:         firstPath,
			RelativePath: "open.md",
			Key:          "CHR-1",
			Title:        "Fix publishing",
			Synopsis:     "Fix the publishing workflow.",
			Status:       "open",
			Priority:     "high",
			Severity:     "critical",
			Type:         "bug",
			UseCase:      "cli-blog",
			ProjectArea:  "indexing",
			SourceKind:   "plan_tracking",
			SourceRef:    ".chr/default/plans/blog-plan.md",
			Slice:        "Slice 1",
		},
		{
			Path:         secondPath,
			RelativePath: "ready.md",
			Key:          "CHR-2",
			Title:        "Add completion",
			Status:       "ready",
			Priority:     "normal",
			Severity:     "normal",
			Type:         "feature",
			UseCase:      "cli-blog",
			ProjectArea:  "tasks",
		},
	}
}

func testTicketEvidenceDetailBody() string {
	return `# Description

<!-- chr:evidence-strategy:v1 -->
- [x] Evidence strategy: complete; change_kind=cli_behavior; risk=medium
- [x] Gherkin: required; path=.chr/default/acceptance/cli.feature
- [x] Tests: mixed; unit and CLI coverage
- [x] Validation: passed; make verify
- [x] Evidence summary: complete; ready for review
<!-- /chr:evidence-strategy -->
`
}

func TestModelSelectsTickets(t *testing.T) {
	model := NewModel(Options{
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "First", Status: "open"},
			{Key: "CHR-2", Title: "Second", Status: "open"},
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

	if ticket.Key != "CHR-2" {
		t.Fatalf("selected ticket = %q, want CHR-2", ticket.Key)
	}
}

func TestModelNavigatesBoardColumns(t *testing.T) {
	model := NewModel(Options{
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "First", Status: core.TicketStatusBacklog},
			{Key: "CHR-2", Title: "Second", Status: core.TicketStatusReadyForDev},
		},
	})
	model.selected = 6

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRight})
	navigated, ok := updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	if navigated.activeBoardColumn() != core.TicketStatusReadyForDev {
		t.Fatalf("board column = %q, want ready_for_dev", navigated.activeBoardColumn())
	}

	view := navigated.View()
	if !strings.Contains(view, "Ready For Dev (2/6)") {
		t.Fatalf("View() does not mark active Board column:\n%s", view)
	}

	ticket, ok := navigated.selectedTicketView()
	if !ok {
		t.Fatal("selectedTicketView returned no ticket")
	}

	if ticket.Key != "CHR-2" {
		t.Fatalf("selected ticket = %q, want CHR-2", ticket.Key)
	}
}

func TestModelRendersActiveBoardColumnInCompactMasterPane(t *testing.T) {
	model := NewModel(Options{
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "First", Status: core.TicketStatusBacklog},
			{Key: "CHR-2", Title: "Second", Status: core.TicketStatusReadyForDev},
		},
	})
	model.selected = 6

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 24})
	resized := requireUpdatedModel(t, updated)

	updated, _ = resized.Update(tea.KeyMsg{Type: tea.KeyRight})
	navigated := requireUpdatedModel(t, updated)
	view := navigated.View()

	for _, want := range []string{
		"RFD (2/6)",
		"> CHR-2",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelRendersCompactTicketRowsWithoutRawStatusOrWrapping(t *testing.T) {
	model := NewModel(Options{
		Tickets: []core.IndexedTicket{
			{
				Key:      "CHR-1",
				Title:    "Legacy ready ticket",
				Status:   core.TicketStatusReadyLegacy,
				Priority: "high",
			},
			{
				Key:      "CHR-15",
				Title:    "Implement filesystem-backed Post store and slug normalization",
				Status:   core.TicketStatusReadyForDev,
				Priority: "normal",
			},
		},
	})
	model.selected = 6

	lines := model.renderTicketFocusList(48, 8)
	rendered := strings.Join(lines, "\n")

	for _, want := range []string{
		"    Key     Sts   Pri   Title",
		"RFD",
		"RDY",
		"Norm",
		"Implement filesystem-ba",
		"…",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("compact ticket list missing %q:\n%s", want, rendered)
		}
	}

	for _, unwanted := range []string{
		"ready_for_dev",
		"Ready",
		"slug normalization",
	} {
		if strings.Contains(rendered, unwanted) {
			t.Fatalf("compact ticket list contains %q:\n%s", unwanted, rendered)
		}
	}

	for index, line := range lines {
		if lipgloss.Width(line) > 48 {
			t.Fatalf("line %d width = %d, want <= 48: %q", index, lipgloss.Width(line), line)
		}
	}
}

func TestModelTogglesBoardAttentionView(t *testing.T) {
	model := NewModel(Options{
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "First", Status: core.TicketStatusBacklog},
			{Key: "CHR-2", Title: "Blocked", Status: core.TicketStatusBlockedLegacy},
		},
	})
	model.selected = 6

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a"), Alt: true})
	attention := requireUpdatedModel(t, updated)

	if !attention.boardAttentionActive() {
		t.Fatal("board attention view is not active")
	}

	view := attention.View()
	for _, want := range []string{
		"Attention (1)",
		"! CHR-2",
		"Attention:",
		"status is blocked",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}

	updated, _ = attention.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a"), Alt: true})
	normal := requireUpdatedModel(t, updated)

	if normal.boardAttentionActive() {
		t.Fatal("board attention view is still active")
	}

	if normal.activeBoardColumn() != core.TicketStatusBacklog {
		t.Fatalf("board column = %q, want backlog", normal.activeBoardColumn())
	}
}

func TestModelRendersMaximizedBoardColumns(t *testing.T) {
	model := NewModel(Options{
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "Backlog ticket", Status: core.TicketStatusBacklog},
			{Key: "CHR-2", Title: "Ready ticket", Status: core.TicketStatusReadyForDev},
			{Key: "CHR-3", Title: "Progress ticket", Status: core.TicketStatusInProgress},
			{Key: "CHR-4", Title: "Review ticket", Status: core.TicketStatusReadyForReview},
			{Key: "CHR-5", Title: "Merge ticket", Status: core.TicketStatusReadyForMerge},
			{Key: "CHR-6", Title: "Done ticket", Status: core.TicketStatusDone},
		},
	})
	model.selected = 6

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 24})
	resized := requireUpdatedModel(t, updated)

	updated, _ = resized.Update(tea.KeyMsg{Type: tea.KeyRight})
	readyColumn := requireUpdatedModel(t, updated)

	updated, _ = readyColumn.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m"), Alt: true})
	maximized := requireUpdatedModel(t, updated)
	view := maximized.View()

	for _, want := range []string{
		"Backlog (1)",
		"Ready For Dev (1)",
		"In Progress (1)",
		"Ready For Review (1)",
		"Ready For Merge (1)",
		"Done (1)",
		"│> CHR-2  RFD",
		"├",
		"│Ready ticket",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}

	updated, _ = maximized.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m"), Alt: true})
	restored := requireUpdatedModel(t, updated)

	if restored.selectedTicket != 1 {
		t.Fatalf("selectedTicket = %d, want 1", restored.selectedTicket)
	}

	if restored.activeBoardColumn() != core.TicketStatusReadyForDev {
		t.Fatalf("activeBoardColumn = %q, want ready_for_dev", restored.activeBoardColumn())
	}
}

func TestModelRendersMaximizedBoardColumnAroundSelectedTicket(t *testing.T) {
	model := NewModel(Options{
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "First ticket", Status: core.TicketStatusReadyForDev},
			{Key: "CHR-2", Title: "Second ticket", Status: core.TicketStatusReadyForDev},
			{Key: "CHR-3", Title: "Third ticket with enough words to wrap across several card body lines", Status: core.TicketStatusReadyForDev},
		},
	})
	model.selectedTicket = 2

	lines := model.renderMaximizedBoardColumnLines(model.tickets, 30, 7)
	rendered := strings.Join(lines, "\n")

	for _, want := range []string{
		"│> CHR-3  RFD",
		"│Third ticket with enough",
		"│words to wrap across several",
		"│card body lines",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("maximized card missing %q:\n%s", want, rendered)
		}
	}

	if strings.Contains(rendered, "CHR-1") {
		t.Fatalf("maximized column did not scroll to selected ticket:\n%s", rendered)
	}
}

func TestModelRendersMaximizedBoardAttentionSeparately(t *testing.T) {
	model := NewModel(Options{
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "Normal ticket", Status: core.TicketStatusBacklog},
			{Key: "CHR-2", Title: "Blocked ticket", Status: core.TicketStatusBlockedLegacy},
		},
	})
	model.selected = 6

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 20})
	resized := requireUpdatedModel(t, updated)

	updated, _ = resized.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a"), Alt: true})
	attention := requireUpdatedModel(t, updated)

	updated, _ = attention.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m"), Alt: true})
	maximized := requireUpdatedModel(t, updated)
	view := maximized.View()

	for _, want := range []string{
		"Attention",
		"CHR-2",
		"Blocked ticket",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}

	if strings.Contains(view, "Backlog (") {
		t.Fatalf("maximized attention view rendered normal Board columns:\n%s", view)
	}
}

func TestTicketFilterMatchesMetadata(t *testing.T) {
	model := NewModel(Options{
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "Fix publishing", Status: "open"},
			{
				Key:        "CHR-2",
				Title:      "Add completion",
				Synopsis:   "Finish generated tickets from a plan.",
				Status:     "ready",
				SourceKind: "plan_tracking",
			},
		},
	})
	model.ticketFilter = "generated tickets"

	tickets := model.filteredTickets()
	if len(tickets) != 1 {
		t.Fatalf("len(filteredTickets) = %d, want 1", len(tickets))
	}

	if tickets[0].Key != "CHR-2" {
		t.Fatalf("filtered ticket = %q, want CHR-2", tickets[0].Key)
	}
}

func TestModelRendersDocumentsAndDetail(t *testing.T) {
	model := NewModel(Options{
		Specs: []SpecView{
			{
				Index:        0,
				UseCaseTitle: "Blog Publishing",
				Status:       "accepted",
				Title:        "TUI Navigation",
				RelativePath: ".chr/default/specs/tui-navigation.md",
				Content:      "# TUI Navigation\n\nStatus: accepted\n\nNavigation details.",
			},
		},
	})
	model.selected = 4

	view := model.View()

	for _, want := range []string{
		"Specs",
		"accepted",
		"Blog Publishing",
		"TUI Navigation",
		".chr/default/specs/tui-navigation.md",
		"Navigation details.",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestLoadDocumentViewsDiscoversProjectMarkdown(t *testing.T) {
	projectDir := t.TempDir()
	writeTicketViewFile(t, filepath.Join(projectDir, "ops", "default", "dump", "00000-context"), "artifact.md", "# Project Context Dump\n\nStatus: raw\n\nNotes.")
	writeTicketViewFile(t, filepath.Join(projectDir, ".prj", "specs", "specs"), "feature.md", "# Feature Spec\n\nStatus: accepted\n\nDetails.")
	writeTicketViewFile(t, filepath.Join(projectDir, ".prj", "trackers"), "feature-tracker.md", "# Feature Tracker\n\nStatus: planned\n")
	writeTicketViewFile(t, filepath.Join(projectDir, ".chr", "default", "plans", "archive"), "blog-plan-20260525.md", "# Archived Blog Plan\n\nStatus: planned\n")
	writeTicketViewFile(t, filepath.Join(projectDir, ".chr", "default", "specs", "archive"), "blog-spec-20260525.md", "# Archived Blog Spec\n\nStatus: draft\n")
	writeTicketViewFile(t, filepath.Join(projectDir, ".prj", "drafts"), "legacy-draft.md", "# Legacy Draft\n\nStatus: draft\n")
	writeTicketViewFile(t, filepath.Join(projectDir, ".chr", "default", "drafts"), "runtime-draft.md", "# Runtime Draft\n\nStatus: draft\n")

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

		if strings.Contains(document.RelativePath, "/archive/") {
			t.Fatalf("archived generated document loaded: %#v", document)
		}
	}
}

func TestModelRendersAuthoringSessionsAndDetail(t *testing.T) {
	updated := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	model := NewModel(Options{
		AuthoringSessions: []core.AuthoringSession{
			{
				ID:        "auth-1",
				ProjectID: "project-1",
				Status:    core.AuthoringSessionStatusActive,
				Mode:      core.AuthoringSessionModeAssisted,
				Target:    core.AuthoringTargetSpec,
				Topic:     "Billing sync",
				UpdatedAt: updated,
				Inputs: []core.AuthoringArtifactRef{
					{Path: ".prj/dumps/billing.md"},
				},
				Outputs: []core.AuthoringArtifactRef{
					{Path: ".chr/default/specs/auth-1-billing-sync.md"},
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
		"Transcript: .chr/authoring/sessions/auth-1/transcript.md",
		"Inputs: .prj/dumps/billing.md",
		"Outputs: .chr/default/specs/auth-1-billing-sync.md",
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
				TicketKey:      "CHR-1",
				Status:         "failed",
				UpdatedAt:      "2026-05-22T10:02:00Z",
				FailureReason:  "validation failed",
				RunDir:         "ops/default/run/00001-ticket-run",
				ResultPath:     "ops/default/run/00001-ticket-run/evidence/result.json",
				LogPath:        "ops/default/run/00001-ticket-run/evidence/runtime.log",
				ValidationPath: "ops/default/run/00001-ticket-run/evidence/validation.log",
			},
		},
	})
	model.views = []View{ViewAgent}
	model.selected = 0

	view := model.View()

	for _, want := range []string{
		"Agent",
		"Daemon: running",
		"PID: 1234",
		"Latest Loop",
		"ID: loop-1",
		"Counts: 3 eligible / 1 processed / 1 blocked / 0 failed",
		"Latest Execution",
		"Ticket: CHR-1",
		"Validation: ops/default/run/00001-ticket-run/evidence/validation.log",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}
}

func TestModelRendersRuntimeObservability(t *testing.T) {
	model := NewModel(Options{
		Observability: ObservabilityView{
			RuntimeImage:             "chr-agent-runtime:local",
			RuntimeDockerfile:        ".chr/default/runtime/agent/Dockerfile",
			RuntimeArtifactDir:       ".chr/default/runtime/runs",
			RuntimeValidationCommand: "make verify",
			RuntimeChecks: []RuntimeCheckView{
				{Name: "workspace", Status: "ok", Path: ".chr", Message: "ok"},
				{Name: "dockerfile", Status: "fail", Path: ".chr/default/runtime/agent/Dockerfile", Message: "missing"},
			},
		},
	})
	model.views = []View{ViewRuntime}
	model.selected = 0

	view := model.View()

	for _, want := range []string{
		"Runtime",
		"Image: chr-agent-runtime:local",
		"Dockerfile: .chr/default/runtime/agent/Dockerfile",
		"Artifact Dir: .chr/default/runtime/runs",
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
		core.NewStoppedDaemonState(newer),
		[]core.ExecutionRecord{
			{ID: "exec-old", Status: core.ExecutionStatusSucceeded, UpdatedAt: older},
			{ID: "exec-new", Status: core.ExecutionStatusFailed, UpdatedAt: newer},
		},
		[]core.LoopRunRecord{
			{ID: "loop-old", Status: core.LoopRunStatusSucceeded, UpdatedAt: older},
			{ID: "loop-new", Status: core.LoopRunStatusFailed, UpdatedAt: newer},
		},
		testObservabilityRuntimeResult(projectDir),
		testObservabilityAgentEvents(projectDir, newer),
	)

	if view.LatestExecution.ID != "exec-new" {
		t.Fatalf("LatestExecution.ID = %q, want exec-new", view.LatestExecution.ID)
	}

	if view.LatestLoop.ID != "loop-new" {
		t.Fatalf("LatestLoop.ID = %q, want loop-new", view.LatestLoop.ID)
	}

	if view.RuntimeDockerfile != ".chr/default/runtime/agent/Dockerfile" {
		t.Fatalf("RuntimeDockerfile = %q, want relative runtime path", view.RuntimeDockerfile)
	}

	if len(view.RecentEvents) != 1 {
		t.Fatalf("RecentEvents length = %d, want 1", len(view.RecentEvents))
	}

	if view.RecentEvents[0].Message != "CHR-2 worker started" {
		t.Fatalf("RecentEvents[0].Message = %q, want CHR-2 worker started", view.RecentEvents[0].Message)
	}

	if view.RecentEvents[0].RuntimeLog != "ops/default/run/00002-ticket-run/evidence/runtime.log" {
		t.Fatalf("RecentEvents[0].RuntimeLog = %q, want relative runtime log path", view.RecentEvents[0].RuntimeLog)
	}

	if view.LatestEvent.Phase != string(core.AgentEventPhaseWorkerStarted) {
		t.Fatalf("LatestEvent.Phase = %q, want worker_started", view.LatestEvent.Phase)
	}
}

func testObservabilityRuntimeResult(projectDir string) core.RuntimeValidationResult {
	return core.RuntimeValidationResult{
		Plan: core.RuntimePlan{
			Image:             "image",
			DockerfilePath:    filepath.Join(projectDir, ".chr", "default", "runtime", "agent", "Dockerfile"),
			ArtifactDir:       filepath.Join(projectDir, ".chr", "default", "runtime", "runs"),
			ValidationCommand: "make verify",
		},
		Checks: []core.RuntimeValidationCheck{
			{Name: "dockerfile", Path: filepath.Join(projectDir, ".chr", "default", "runtime", "agent", "Dockerfile"), OK: true, Message: "ok"},
		},
	}
}

func testObservabilityAgentEvents(projectDir string, timestamp time.Time) []core.AgentEvent {
	return []core.AgentEvent{
		{
			SchemaVersion: core.AgentEventSchemaVersion,
			Timestamp:     timestamp,
			LoopID:        "loop-new",
			ExecutionID:   "exec-new",
			TicketKey:     "CHR-2",
			Phase:         core.AgentEventPhaseWorkerStarted,
			Level:         core.AgentEventLevelInfo,
			Message:       "CHR-2 worker started",
			Refs: core.AgentEventRefs{
				"runtime_log_path": filepath.Join(projectDir, "ops", "default", "run", "00002-ticket-run", "evidence", "runtime.log"),
			},
		},
	}
}

func TestModelSelectsDocuments(t *testing.T) {
	model := NewModel(Options{
		Specs: []SpecView{
			{Index: 0, UseCaseTitle: "First", Title: "First Spec"},
			{Index: 1, UseCaseTitle: "Second", Title: "Second Spec"},
		},
	})
	model.selected = 4

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	navigated, ok := updated.(Model)

	if !ok {
		t.Fatalf("Update returned %T, want Model", updated)
	}

	document, ok := navigated.selectedSpecView()
	if !ok {
		t.Fatal("selectedSpecView returned no spec")
	}

	if document.UseCaseTitle != "Second" {
		t.Fatalf("selected use case = %q, want Second", document.UseCaseTitle)
	}
}

func TestModelRendersScopedDumps(t *testing.T) {
	updated := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	model := NewModel(Options{
		Dumps: []DumpView{
			{
				Title:        "Project Context",
				UseCaseTitle: "Project-wide",
				Status:       "raw",
				Scope:        DumpScope{},
				RelativePath: ".chr/default/dumps/project-context.md",
				SourcePath:   ".chr/default/dumps/project-context-raw.md",
				UpdatedAt:    updated,
				Preview:      "Project-wide notes.",
			},
			{
				Title:        "Blog Context",
				UseCaseTitle: "Blog Publishing",
				Status:       "curated",
				Scope:        DumpScope{CaseSlug: "blog"},
				RelativePath: ".chr/default/dumps/blog-context.md",
				UpdatedAt:    updated,
				Preview:      "Use-case notes.",
			},
		},
	})
	model.selected = 2

	view := model.View()
	for _, want := range []string{
		"Dumps",
		"Project-wide",
		"Blog Publishing",
		"Raw: .chr/default/dumps/project-context-raw.md",
		"Project-wide notes.",
		"enter edit  alt+e external  c create  r refresh",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("View() missing %q:\n%s", want, view)
		}
	}

	for _, unwanted := range []string{
		"Use Case     Status      Title",
		"Selected Dump",
		"Path: .chr/default/dumps/project-context.md",
		"Updated:",
		"Status:",
	} {
		if strings.Contains(view, unwanted) {
			t.Fatalf("View() contains unwanted %q:\n%s", unwanted, view)
		}
	}
}

func TestModelDumpListUsesUseCaseTitle(t *testing.T) {
	model := NewModel(Options{
		Dumps: []DumpView{
			{
				UseCaseTitle: "Blog Publishing",
				Title:        "Blog Context",
				Scope:        DumpScope{CaseSlug: "blog"},
				Preview:      "# Blog Context\n\nNotes.",
			},
		},
	})
	model.selected = 2

	view := model.View()
	if !strings.Contains(view, "> Blog Publishing") {
		t.Fatalf("View() missing use case title in dump list:\n%s", view)
	}

	if !strings.Contains(view, "Blog Context") {
		t.Fatalf("View() missing dump title in detail:\n%s", view)
	}
}

func TestModelRendersDumpContentWithMarkdownLineBreaks(t *testing.T) {
	model := NewModel(Options{
		Dumps: []DumpView{
			{
				Title:        "Project Context",
				RelativePath: ".chr/default/dumps/project-context.md",
				Preview:      "---\ntitle: \"Project Context\"\n---\n\n# Project Context\n\nNotes.",
			},
		},
	})
	model.selected = 2

	view := model.View()

	if !strings.Contains(view, "---\n") {
		t.Fatalf("View() does not preserve frontmatter line break:\n%s", view)
	}

	if !strings.Contains(view, "\n# Project Context\n") {
		t.Fatalf("View() does not preserve heading line break:\n%s", view)
	}
}

func TestModelSelectsDumpsIndependentlyFromDocuments(t *testing.T) {
	model := NewModel(Options{
		Dumps: []DumpView{
			{Title: "Project Context", RelativePath: ".chr/default/dumps/project-context.md"},
			{Title: "Blog Context", Scope: DumpScope{CaseSlug: "blog"}, RelativePath: ".chr/default/dumps/blog-context.md"},
		},
		Documents: []DocumentView{
			{Index: 0, Family: "spec", Title: "Spec", RelativePath: ".chr/default/specs/spec.md"},
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
				RawPath:        ".chr/default/quizzes/project-quiz-raw.md",
				CuratedPath:    ".chr/default/quizzes/project-quiz.md",
				RuntimePath:    ".chr/default/tmp/quiz-runtime/project.json",
				QuestionBudget: 50,
			},
			{
				Title:                 "Blog Publishing",
				Status:                "active",
				Mode:                  "fast",
				Scope:                 QuizScope{CaseSlug: "blog"},
				RawPath:               ".chr/default/quizzes/blog-quiz-raw.md",
				CuratedPath:           ".chr/default/quizzes/blog-quiz.md",
				RuntimePath:           ".chr/default/tmp/quiz-runtime/blog.json",
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
		"What should the workflow preserve?",
		"enter answers",
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
				RawPath:               ".chr/default/quizzes/blog-quiz-raw.md",
				CuratedPath:           ".chr/default/quizzes/blog-quiz.md",
				RuntimePath:           ".chr/default/tmp/quiz-runtime/blog.json",
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
			{Index: 0, Family: "spec", Title: "Spec", RelativePath: ".chr/default/specs/spec.md"},
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

	if confirmed.statusMessage != "Answering quiz." {
		t.Fatalf("statusMessage = %q, want answer status", confirmed.statusMessage)
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

	assertQuizAnswerHelperVisible(t, confirmed)
}

func TestModelGeneratesSpecFromSpecsKey(t *testing.T) {
	executor := &fakeSpecActionExecutor{
		result: SpecActionResult{
			Message: "Spec generated.",
			Documents: []DocumentView{
				{Family: "spec", Status: "draft", Title: "Generated Spec", RelativePath: ".chr/default/specs/drafts/blog-quiz-spec.md"},
			},
		},
	}
	model := NewModel(Options{
		SpecExecutor: executor,
		Specs: []SpecView{
			{Title: "Project", Scope: QuizScope{}, Status: "done"},
			{Title: "Blog", Scope: QuizScope{CaseSlug: "blog"}, Status: "done"},
		},
	})
	model.selected = 4
	model.selectedSpec = 1

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	busy := requireUpdatedModel(t, updated)

	if busy.specAction.Pending {
		t.Fatal("spec action is pending, want direct execution")
	}

	if !busy.specBusy {
		t.Fatal("spec action is not marked busy")
	}

	if busy.statusMessage != "Generating spec..." {
		t.Fatalf("statusMessage = %q, want Generating spec...", busy.statusMessage)
	}

	updated, _ = busy.Update(cmd())
	confirmed := requireUpdatedModel(t, updated)

	if executor.request.Action != SpecActionGenerate {
		t.Fatalf("executor action = %q, want %q", executor.request.Action, SpecActionGenerate)
	}

	if executor.request.Scope.CaseSlug != "blog" {
		t.Fatalf("scope = %q, want blog", executor.request.Scope.CaseSlug)
	}

	if confirmed.statusMessage != "Spec generated." {
		t.Fatalf("statusMessage = %q, want Spec generated.", confirmed.statusMessage)
	}

	if len(confirmed.documents) != 1 || confirmed.documents[0].Title != "Generated Spec" {
		t.Fatalf("documents = %#v, want generated spec", confirmed.documents)
	}
}

func TestModelGeneratesAcceptanceFeatureFromSpecsKey(t *testing.T) {
	executor := &fakeSpecActionExecutor{
		result: SpecActionResult{
			Message:                 "Acceptance feature needs more information.",
			AcceptanceFeatureStatus: "needs_more_info",
			AcceptanceFeatureRound:  ".chr/default/acceptance/blog-feature-round.md",
			AcceptanceFeatureReason: "Missing observable failure path.",
		},
	}
	model := NewModel(Options{
		SpecExecutor: executor,
		Specs: []SpecView{
			{Title: "Blog", Scope: QuizScope{CaseSlug: "blog"}, Status: "draft"},
		},
	})
	model.selected = 4

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	busy := requireUpdatedModel(t, updated)

	if !busy.specBusy {
		t.Fatal("spec action is not marked busy")
	}

	if busy.statusMessage != "Generating acceptance feature..." {
		t.Fatalf("statusMessage = %q, want acceptance feature busy message", busy.statusMessage)
	}

	updated, _ = busy.Update(cmd())
	confirmed := requireUpdatedModel(t, updated)

	if executor.request.Action != SpecActionGenerateFeature {
		t.Fatalf("executor action = %q, want %q", executor.request.Action, SpecActionGenerateFeature)
	}

	for _, want := range []string{
		"Acceptance feature needs more information.",
		"status: needs_more_info",
		"feature round: .chr/default/acceptance/blog-feature-round.md",
		"reason: Missing observable failure path.",
	} {
		if !strings.Contains(confirmed.statusMessage, want) {
			t.Fatalf("statusMessage missing %q: %q", want, confirmed.statusMessage)
		}
	}
}

func TestModelReportsSpecGenerationFailureShowsLastSavedSpec(t *testing.T) {
	executor := &fakeSpecActionExecutor{err: errors.New(`invoke OpenAI quiz model: Post "https://api.openai.com/v1/responses": context deadline exceeded`)}
	model := NewModel(Options{
		SpecExecutor: executor,
		Specs: []SpecView{
			{Title: "Blog Spec", Scope: QuizScope{CaseSlug: "blog"}, Status: "draft", Content: "# Blog Spec"},
		},
	})
	model.selected = 4

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	busy := requireUpdatedModel(t, updated)

	updated, _ = busy.Update(cmd())
	confirmed := requireUpdatedModel(t, updated)

	if confirmed.statusMessage != "Spec failed; showing last saved spec: provider timeout" {
		t.Fatalf("statusMessage = %q, want last saved spec failure", confirmed.statusMessage)
	}
}

func TestModelReportsAcceptanceFeatureGenerationFailureWithFeatureLabel(t *testing.T) {
	executor := &fakeSpecActionExecutor{err: errors.New("invalid JSON schema")}
	model := NewModel(Options{
		SpecExecutor: executor,
		Specs: []SpecView{
			{Title: "Blog Spec", Scope: QuizScope{CaseSlug: "blog"}, Status: "draft", Content: "# Blog Spec"},
		},
	})
	model.selected = 4

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	busy := requireUpdatedModel(t, updated)

	updated, _ = busy.Update(cmd())
	confirmed := requireUpdatedModel(t, updated)

	if confirmed.statusMessage != "Acceptance feature failed; showing last saved acceptance feature: invalid JSON schema" {
		t.Fatalf("statusMessage = %q, want acceptance feature failure", confirmed.statusMessage)
	}
}

//nolint:cyclop,funlen // End-to-end TUI action test validates several state updates together.
func TestModelGeneratesPlanFromPlansKey(t *testing.T) {
	executor := &fakePlanActionExecutor{
		result: PlanActionResult{
			Message: "Plan generated.",
			Documents: []DocumentView{
				{Family: "plan", Status: "planned", Title: "Generated Plan", RelativePath: ".chr/default/plans/blog-plan.md"},
			},
			Plans: []PlanView{
				{UseCaseTitle: "Project", Title: "Project", Scope: QuizScope{}, Status: "missing"},
				{UseCaseTitle: "Blog", Title: "Generated Plan", Scope: QuizScope{CaseSlug: "blog"}, Status: "planned"},
			},
			Scope: QuizScope{CaseSlug: "blog"},
		},
	}
	model := NewModel(Options{
		PlanExecutor: executor,
		Plans: []PlanView{
			{Title: "Project", Scope: QuizScope{}, Status: "missing"},
			{Title: "Blog", Scope: QuizScope{CaseSlug: "blog"}, Status: "missing"},
		},
	})
	model.selected = 5
	model.selectedPlan = 1

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	busy := requireUpdatedModel(t, updated)

	if busy.planAction.Pending {
		t.Fatal("plan action is pending, want direct execution")
	}

	if !busy.planBusy {
		t.Fatal("plan action is not marked busy")
	}

	if busy.statusMessage != "Generating plan..." {
		t.Fatalf("statusMessage = %q, want Generating plan...", busy.statusMessage)
	}

	if busy.selectedPane != 1 {
		t.Fatalf("selectedPane = %d, want 1", busy.selectedPane)
	}

	if busy.maximizedPane != 1 {
		t.Fatalf("maximizedPane = %d, want 1", busy.maximizedPane)
	}

	updated, _ = busy.Update(cmd())
	confirmed := requireUpdatedModel(t, updated)

	if executor.request.Action != PlanActionGenerate {
		t.Fatalf("executor action = %q, want %q", executor.request.Action, PlanActionGenerate)
	}

	if executor.request.Scope.CaseSlug != "blog" {
		t.Fatalf("scope = %q, want blog", executor.request.Scope.CaseSlug)
	}

	if confirmed.statusMessage != "Plan generated." {
		t.Fatalf("statusMessage = %q, want Plan generated.", confirmed.statusMessage)
	}

	if len(confirmed.documents) != 1 || confirmed.documents[0].Title != "Generated Plan" {
		t.Fatalf("documents = %#v, want generated plan", confirmed.documents)
	}

	plan, ok := confirmed.selectedPlanView()
	if !ok || plan.Title != "Generated Plan" {
		t.Fatalf("selected plan = %#v, want generated plan", plan)
	}
}

func TestModelReportsPlanGenerationFailureShowsLastSavedPlan(t *testing.T) {
	executor := &fakePlanActionExecutor{err: errors.New(`invoke OpenAI quiz model: Post "https://api.openai.com/v1/responses": context deadline exceeded`)}
	model := NewModel(Options{
		PlanExecutor: executor,
		Plans: []PlanView{
			{
				Title:        "Blog Plan",
				Scope:        QuizScope{CaseSlug: "blog"},
				Status:       "planned",
				RelativePath: ".chr/default/plans/blog-plan.md",
				Content:      "# Blog Plan",
			},
		},
	})
	model.selected = 5

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	pending := requireUpdatedModel(t, updated)

	updated, cmd := pending.Update(tea.KeyMsg{Type: tea.KeyEnter})
	busy := requireUpdatedModel(t, updated)

	updated, _ = busy.Update(cmd())
	confirmed := requireUpdatedModel(t, updated)

	if confirmed.statusMessage != "Plan failed; showing last saved plan: provider timeout" {
		t.Fatalf("statusMessage = %q, want last saved plan failure", confirmed.statusMessage)
	}
}

func TestModelRefusesGeneratePlanWhenPlanExists(t *testing.T) {
	executor := &fakePlanActionExecutor{}
	model := NewModel(Options{
		PlanExecutor: executor,
		Plans: []PlanView{
			{
				Title:        "Generated Plan",
				Scope:        QuizScope{CaseSlug: "blog"},
				Status:       "planned",
				RelativePath: ".chr/default/plans/blog-plan.md",
				Content:      "# Generated Plan\n\nExisting.",
			},
		},
	})
	model.selected = 5

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	confirmed := requireUpdatedModel(t, updated)

	if cmd != nil {
		t.Fatal("cmd is non-nil, want no generation command")
	}

	if confirmed.planBusy {
		t.Fatal("planBusy = true, want false")
	}

	if executor.request.Action != "" {
		t.Fatalf("executor action = %q, want no action", executor.request.Action)
	}

	if confirmed.statusMessage != "Plan already exists. Use regenerate plan to replace it." {
		t.Fatalf("statusMessage = %q, want existing plan refusal", confirmed.statusMessage)
	}
}

//nolint:cyclop,funlen,gocognit // Confirmation flow test keeps pending, busy, and complete states together.
func TestModelRegeneratesPlanWithConfirmation(t *testing.T) {
	executor := &fakePlanActionExecutor{
		result: PlanActionResult{
			Message: "Plan regenerated.",
			Plans: []PlanView{
				{
					Title:        "Regenerated Plan",
					Scope:        QuizScope{CaseSlug: "blog"},
					Status:       "planned",
					RelativePath: ".chr/default/plans/blog-plan.md",
				},
			},
			Scope: QuizScope{CaseSlug: "blog"},
		},
	}
	model := NewModel(Options{
		PlanExecutor: executor,
		Plans: []PlanView{
			{
				Title:        "Generated Plan",
				Scope:        QuizScope{CaseSlug: "blog"},
				Status:       "planned",
				RelativePath: ".chr/default/plans/blog-plan.md",
				Content:      "# Generated Plan\n\nExisting.",
			},
		},
	})
	model.selected = 5

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	pending := requireUpdatedModel(t, updated)

	if cmd != nil {
		t.Fatal("cmd is non-nil before confirmation")
	}

	if !pending.planAction.Pending {
		t.Fatal("plan action is not pending")
	}

	if pending.planAction.Request.Action != PlanActionRegenerate {
		t.Fatalf("pending action = %q, want %q", pending.planAction.Request.Action, PlanActionRegenerate)
	}

	if pending.selectedPane != 1 {
		t.Fatalf("selectedPane = %d, want 1", pending.selectedPane)
	}

	if pending.maximizedPane != 1 {
		t.Fatalf("maximizedPane = %d, want 1", pending.maximizedPane)
	}

	if pending.statusMessage != "Confirm plan action with Enter, or cancel with Esc." {
		t.Fatalf("statusMessage = %q, want confirmation guidance", pending.statusMessage)
	}

	if action := pending.renderPlanAction(); !strings.Contains(action, "Enter: confirm") || !strings.Contains(action, "Esc: cancel") {
		t.Fatalf("renderPlanAction() = %q, want confirmation keys", action)
	}

	updated, cmd = pending.Update(tea.KeyMsg{Type: tea.KeyEnter})
	busy := requireUpdatedModel(t, updated)

	if !busy.planBusy {
		t.Fatal("planBusy = false, want true")
	}

	if busy.statusMessage != "Regenerating plan..." {
		t.Fatalf("statusMessage = %q, want Regenerating plan...", busy.statusMessage)
	}

	action := busy.renderPlanAction()

	hasRunningOverlay := strings.Contains(action, "Plan Action Running") &&
		strings.Contains(action, "Regenerating plan...") &&
		strings.Contains(action, "Please wait")
	if !hasRunningOverlay {
		t.Fatalf("renderPlanAction() = %q, want running overlay", action)
	}

	if busy.maximizedPane != 1 {
		t.Fatalf("maximizedPane = %d, want 1 after confirmation", busy.maximizedPane)
	}

	updated, _ = busy.Update(cmd())
	confirmed := requireUpdatedModel(t, updated)

	if executor.request.Action != PlanActionRegenerate {
		t.Fatalf("executor action = %q, want %q", executor.request.Action, PlanActionRegenerate)
	}

	if confirmed.statusMessage != "Plan regenerated." {
		t.Fatalf("statusMessage = %q, want Plan regenerated.", confirmed.statusMessage)
	}

	if confirmed.maximizedPane != 1 {
		t.Fatalf("maximizedPane = %d, want 1 after regeneration", confirmed.maximizedPane)
	}
}

func TestModelShowsPlanTrackingOnlyWhenPlanExists(t *testing.T) {
	model := NewModel(Options{
		Plans: []PlanView{
			{
				Title:  "Missing Plan",
				Scope:  QuizScope{CaseSlug: "blog"},
				Status: "missing",
			},
		},
	})
	model.selected = 5

	if strings.Contains(model.planActionHint(), "alt+t tracking") {
		t.Fatalf("planActionHint() = %q, want no tracking hint for missing plan", model.planActionHint())
	}

	model.plans = normalizePlanViews([]PlanView{
		{
			Title:        "Generated Plan",
			Scope:        QuizScope{CaseSlug: "blog"},
			Status:       "planned",
			RelativePath: ".chr/default/plans/blog-plan.md",
			Content:      "# Generated Plan\n",
		},
	})

	if !strings.Contains(model.planActionHint(), "alt+t tracking") {
		t.Fatalf("planActionHint() = %q, want tracking hint for existing plan", model.planActionHint())
	}
}

func TestModelPlansPlanTrackingWithAltT(t *testing.T) {
	executor := &fakePlanActionExecutor{}
	model := NewModel(Options{
		PlanExecutor: executor,
		Plans: []PlanView{
			{
				Title:        "Generated Plan",
				Scope:        QuizScope{CaseSlug: "blog"},
				Status:       "planned",
				RelativePath: ".chr/default/plans/blog-plan.md",
				Content:      "# Generated Plan\n",
			},
		},
	})
	model.selected = 5

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t"), Alt: true})
	pending := requireUpdatedModel(t, updated)

	if cmd != nil {
		t.Fatal("cmd is non-nil before confirmation")
	}

	if !pending.planAction.Pending {
		t.Fatal("plan action is not pending")
	}

	if pending.planAction.Request.Action != PlanActionMaterializeTracking {
		t.Fatalf("pending action = %q, want %q", pending.planAction.Request.Action, PlanActionMaterializeTracking)
	}

	updated, cmd = pending.Update(tea.KeyMsg{Type: tea.KeyEnter})
	busy := requireUpdatedModel(t, updated)

	if cmd == nil {
		t.Fatal("cmd is nil, want tracking command")
	}

	if !busy.planBusy {
		t.Fatal("planBusy = false, want true")
	}

	if busy.statusMessage != "Creating plan tracking..." {
		t.Fatalf("statusMessage = %q, want Creating plan tracking...", busy.statusMessage)
	}

	action := busy.renderPlanAction()

	hasTrackingOverlay := strings.Contains(action, "Plan Action Running") &&
		strings.Contains(action, "Creating plan tracking...")
	if !hasTrackingOverlay {
		t.Fatalf("renderPlanAction() = %q, want tracking running overlay", action)
	}
}

func TestModelBlocksPlanTrackingWhileBusy(t *testing.T) {
	model := NewModel(Options{
		Plans: []PlanView{
			{
				Title:        "Generated Plan",
				Scope:        QuizScope{CaseSlug: "blog"},
				Status:       "planned",
				RelativePath: ".chr/default/plans/blog-plan.md",
				Content:      "# Generated Plan\n",
			},
		},
	})
	model.selected = 5
	model.planBusy = true

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t"), Alt: true})
	blocked := requireUpdatedModel(t, updated)

	if cmd != nil {
		t.Fatal("cmd is non-nil, want busy guard")
	}

	if blocked.planAction.Pending {
		t.Fatal("plan action is pending")
	}

	if blocked.statusMessage != "Plan action already running." {
		t.Fatalf("statusMessage = %q, want busy guard", blocked.statusMessage)
	}
}

func TestModelBlocksRegenerateWhileBusy(t *testing.T) {
	model := NewModel(Options{
		Plans: []PlanView{
			{
				Title:        "Generated Plan",
				Scope:        QuizScope{CaseSlug: "blog"},
				Status:       "planned",
				RelativePath: ".chr/default/plans/blog-plan.md",
				Content:      "# Generated Plan\n",
			},
		},
	})
	model.selected = 5
	model.planBusy = true

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	blocked := requireUpdatedModel(t, updated)

	if cmd != nil {
		t.Fatal("cmd is non-nil, want busy guard")
	}

	if blocked.planAction.Pending {
		t.Fatal("plan action is pending")
	}

	if blocked.statusMessage != "Plan action already running." {
		t.Fatalf("statusMessage = %q, want busy guard", blocked.statusMessage)
	}
}

func TestModelBlocksPlanTrackingWhenPlanIsMissing(t *testing.T) {
	model := NewModel(Options{
		Plans: []PlanView{
			{
				Title:  "Missing Plan",
				Scope:  QuizScope{CaseSlug: "blog"},
				Status: "missing",
			},
		},
	})
	model.selected = 5

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t"), Alt: true})
	blocked := requireUpdatedModel(t, updated)

	if cmd != nil {
		t.Fatal("cmd is non-nil, want no command")
	}

	if blocked.planAction.Pending {
		t.Fatal("plan action is pending")
	}

	if blocked.statusMessage != "Plan tracking requires an existing plan." {
		t.Fatalf("statusMessage = %q, want missing plan guidance", blocked.statusMessage)
	}
}

func TestPlanTrackingCommandOnlyAppearsWhenPlanExists(t *testing.T) {
	model := NewModel(Options{
		Plans: []PlanView{{Title: "Missing Plan", Status: "missing"}},
	})
	model.selected = 5
	model.commandPalette.Filter = "materialize-plan-tracking"

	if commands := model.filteredCommands(); len(commands) != 0 {
		t.Fatalf("commands = %#v, want no tracking command for missing plan", commands)
	}

	model.plans = normalizePlanViews([]PlanView{
		{
			Title:        "Generated Plan",
			Status:       "planned",
			RelativePath: ".chr/default/plans/project-plan.md",
			Content:      "# Generated Plan\n",
		},
	})

	commands := model.filteredCommands()
	if len(commands) != 1 || commands[0].Plan != PlanActionMaterializeTracking {
		t.Fatalf("commands = %#v, want tracking command", commands)
	}
}

func TestModelBlocksDuplicatePlanActionWhileBusy(t *testing.T) {
	executor := &fakePlanActionExecutor{}
	model := NewModel(Options{
		PlanExecutor: executor,
		Plans: []PlanView{
			{Title: "Blog", Scope: QuizScope{CaseSlug: "blog"}, Status: "missing"},
		},
	})
	model.selected = 5
	model.planBusy = true

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	confirmed := requireUpdatedModel(t, updated)

	if cmd != nil {
		t.Fatal("cmd is non-nil, want duplicate action blocked")
	}

	if confirmed.statusMessage != "Plan action already running." {
		t.Fatalf("statusMessage = %q, want busy guard", confirmed.statusMessage)
	}
}

func TestModelSpecListUsesUseCaseTitleWhenSpecExists(t *testing.T) {
	model := NewModel(Options{
		Specs: []SpecView{
			{
				UseCaseTitle: "Blog Publishing",
				Title:        "Generated Blog Workflow Spec",
				Status:       "draft",
				Scope:        QuizScope{CaseSlug: "blog"},
				Content:      "# Generated Blog Workflow Spec\n\nDetails.",
			},
		},
	})
	model.selected = 4

	view := model.View()
	if !strings.Contains(view, "> Blog Publishing") {
		t.Fatalf("View() missing use case title in list:\n%s", view)
	}

	if !strings.Contains(view, "Generated Blog Workflow Spec") {
		t.Fatalf("View() missing spec title in detail:\n%s", view)
	}
}

func assertQuizAnswerHelperVisible(t *testing.T, model Model) {
	t.Helper()

	view := model.View()
	if !strings.Contains(view, "Answer") {
		t.Fatalf("View() missing answer panel title:\n%s", view)
	}

	if !strings.Contains(view, "enter submits  shift+enter inserts line  alt+s skip  alt+d done  esc cancel") {
		t.Fatalf("View() missing quiz edit helper:\n%s", view)
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
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "First", Status: "open"},
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

type fakeSpecActionExecutor struct {
	request SpecActionRequest
	result  SpecActionResult
	err     error
}

func (e *fakeSpecActionExecutor) ExecuteSpecAction(request SpecActionRequest) (SpecActionResult, error) {
	e.request = request

	return e.result, e.err
}

type fakePlanActionExecutor struct {
	request PlanActionRequest
	result  PlanActionResult
	err     error
}

func (e *fakePlanActionExecutor) ExecutePlanAction(request PlanActionRequest) (PlanActionResult, error) {
	e.request = request

	return e.result, e.err
}

type fakeProjectMetadataExecutor struct {
	saved core.ProjectMetadata
}

func (e *fakeProjectMetadataExecutor) SaveProjectMetadata(
	metadata core.ProjectMetadata,
) (ProjectMetadataResult, error) {
	e.saved = metadata

	return ProjectMetadataResult{
		Message:  "Project metadata saved.",
		Metadata: metadata,
	}, nil
}

type fakeCaseMetadataExecutor struct {
	saved core.CaseMetadata
}

func (e *fakeCaseMetadataExecutor) SaveCaseMetadata(
	metadata core.CaseMetadata,
) (CaseMetadataResult, error) {
	e.saved = metadata

	e.saved.SchemaVersion = core.CaseMetadataSchemaVersion
	if e.saved.Slug == "" {
		e.saved.Slug = "blog-publishing"
	}

	e.saved.Exists = true

	return CaseMetadataResult{
		Message: "Case metadata saved.",
		Case:    e.saved,
		Cases:   []core.CaseMetadata{e.saved},
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
	called  bool
	request WorkActionRequest
	result  WorkActionResult
	err     error
}

type fakeAgentWorkRefreshExecutor struct {
	called bool
	result AgentWorkRefreshResult
	err    error
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

type fakeSourceOpenExecutor struct {
	path   string
	result SourceOpenResult
	err    error
}

func (e *fakeSourceOpenExecutor) OpenSource(path string) (SourceOpenResult, error) {
	e.path = path

	return e.result, e.err
}

func (e *fakeWorkActionExecutor) ExecuteWorkAction(request WorkActionRequest) (WorkActionResult, error) {
	e.called = true
	e.request = request

	return e.result, e.err
}

func testSuccessfulWorkActionResult() WorkActionResult {
	return WorkActionResult{
		Message: "Execution exec-1 for ticket CHR-1 finished with status succeeded.",
		Tickets: []core.IndexedTicket{
			{Key: "CHR-1", Title: "First", Status: "reviewing"},
		},
		Observability: ObservabilityView{
			LatestExecution: ExecutionView{ID: "exec-1", Status: "succeeded"},
		},
	}
}

func testBlockedWorkReadinessView() WorkReadinessView {
	return WorkReadinessView{
		Status: "blocked",
		Diagnostics: []WorkReadinessDiagnosticView{
			{
				Code:            "runtime_artifacts_visible",
				RepairCode:      "ignore_runtime_artifacts",
				RepairAvailable: true,
				RepairSafe:      true,
			},
		},
	}
}

func (e *fakeAgentWorkRefreshExecutor) RefreshAgentWork(AgentWorkRefreshRequest) (AgentWorkRefreshResult, error) {
	e.called = true

	return e.result, e.err
}

func assertConfirmedArtifactState(t *testing.T, model Model) {
	t.Helper()

	if len(model.tickets) != 1 || model.tickets[0].Key != "CHR-1" {
		t.Fatalf("tickets = %#v, want reloaded CHR-1 ticket", model.tickets)
	}

	if len(model.documents) != 1 || model.documents[0].Title != "TUI Plan" {
		t.Fatalf("documents = %#v, want reloaded TUI Plan document", model.documents)
	}

	if len(model.authoringSessions) != 1 || model.authoringSessions[0].ID != "author-20260522-100000" {
		t.Fatalf("authoringSessions = %#v, want reloaded authoring session", model.authoringSessions)
	}
}
