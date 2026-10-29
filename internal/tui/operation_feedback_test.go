package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"forge.adrianpk.com/chr/internal/app/operation"
)

// A submitted answer remains visible until acknowledgement so a failed or late
// response cannot destroy the exact draft that the user may need to recover.
func TestOperationDraft(t *testing.T) {
	model := NewModel(Options{Quizzes: []QuizView{{ActiveQuestionSummary: "Keep which input?"}}})
	model.selectView(ViewQuizzes)
	model.startQuizAnswer()
	model.quizAnswer.Input.SetValue("Unsent sentinel answer")

	model.quizExecutor = feedbackQuizExecutor{}

	command := model.saveQuizAnswer()
	if command == nil || !model.quizAnswer.Active || model.quizAnswer.Input.Value() != "Unsent sentinel answer" {
		t.Fatal("submission discarded the answer before acknowledgement")
	}
}

type feedbackQuizExecutor struct{}

func (feedbackQuizExecutor) ExecuteQuizAction(QuizActionRequest) (QuizActionResult, error) {
	return QuizActionResult{}, nil
}

// Two queued replies must remain attached to their captured interaction; an
// older completion cannot replace a newer result or its selected document.
func TestOperationLate(t *testing.T) {
	model := NewModel(Options{})
	model.selectView(ViewSpecs)

	model.specExecutor = feedbackSpecExecutor("Older reply")
	older := model.executeSpecActionCommand(SpecActionRequest{Action: SpecActionGenerate})
	model.specExecutor = feedbackSpecExecutor("Newer reply")
	newer := model.executeSpecActionCommand(SpecActionRequest{Action: SpecActionGenerate})
	model.Update(newer())
	model.Update(older())

	if model.statusMessage != "Newer reply" {
		t.Fatal("late completion replaced the current interaction", model.statusMessage)
	}
}

type feedbackSpecExecutor string

func (reply feedbackSpecExecutor) ExecuteSpecAction(SpecActionRequest) (SpecActionResult, error) {
	return SpecActionResult{Message: string(reply)}, nil
}

// A rejected save leaves the complete textarea and only confirmed reload can
// replace it with the newly observed body and revision from the same source.
func TestOperationReload(t *testing.T) {
	executor := &feedbackDumpExecutor{}
	model := NewModel(Options{DumpExecutor: executor})
	model.selectView(ViewDumps)
	model.startDumpEdit("bound.md", "Original", "old")
	model.dumpEdit.Input.SetValue("Unsent sentinel")
	model.saveDumpEdit()

	if model.dumpEdit.Input.Value() != "Unsent sentinel" || model.feedback.Result.Status != operation.Stale {
		t.Fatal("stale save lost its draft")
	}

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	model.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if executor.reads != 0 || model.dumpEdit.Input.Value() != "Unsent sentinel" {
		t.Fatal("unconfirmed reload changed input")
	}

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if executor.reads != 1 || model.dumpEdit.Input.Value() != "Canonical newer body" || model.dumpEdit.Revision != "new" {
		t.Fatal("confirmed reload did not replace the captured source")
	}
}

type feedbackDumpExecutor struct {
	reads    int
	requests []DumpActionRequest
	pending  bool
}

func (executor *feedbackDumpExecutor) ExecuteDumpAction(request DumpActionRequest) (DumpActionResult, error) {
	executor.requests = append(executor.requests, request)
	if request.Action == DumpActionEdit {
		executor.reads++

		return DumpActionResult{SelectedPath: request.Path, Content: "Canonical newer body", Revision: "new", Outcome: operation.Result{Status: operation.OK}}, nil
	}

	if executor.pending && len(executor.requests) > 1 {
		return DumpActionResult{Outcome: operation.Result{Status: operation.OK}}, nil
	}

	outcome := operation.Result{Status: operation.Stale, Reload: true, Summary: "Reload the changed dump."}
	if executor.pending {
		outcome = operation.Result{Status: operation.Blocked, Retry: true, Summary: "Retained operation pending."}
	}

	return DumpActionResult{Outcome: outcome}, errors.New("rejected original write")
}

// Recovery retries freeze the original request, including its operation and
// revision; changing the visible editor cannot submit a different payload.
func TestOperationRetry(t *testing.T) {
	executor := &feedbackDumpExecutor{pending: true}
	model := NewModel(Options{DumpExecutor: executor})
	model.selectView(ViewDumps)
	model.startDumpEdit("bound.md", "Original", "old")
	model.dumpEdit.Input.SetValue("Retained sentinel")
	model.saveDumpEdit()
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})

	if len(executor.requests) != 2 || executor.requests[0] != executor.requests[1] {
		t.Fatal("retry replaced the original write binding")
	}
}

// Escape asks before losing entered text in every existing manual editor;
// cancelling that confirmation leaves its complete draft untouched.
func TestOperationDiscard(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(*Model)
		read  func(*Model) string
	}{
		{"dump", func(m *Model) {
			m.selectView(ViewDumps)
			m.startDumpEdit("bound.md", "", "rev")
			m.dumpEdit.Input.SetValue("sentinel")
		}, func(m *Model) string { return m.dumpEdit.Input.Value() }},
		{"quiz", func(m *Model) {
			m.selectView(ViewQuizzes)

			m.quizzes = normalizeQuizViews([]QuizView{{ActiveQuestionSummary: "Required question"}})
			m.startQuizAnswer()
			m.quizAnswer.Input.SetValue("sentinel")
		}, func(m *Model) string { return m.quizAnswer.Input.Value() }},
		{"project", func(m *Model) {
			m.selectView(ViewProject)

			m.projectEdit = newProjectEditState(0, "")
			m.projectEdit.Input.SetValue("sentinel")
		}, func(m *Model) string { return m.projectEdit.Input.Value() }},
		{"case", func(m *Model) { m.caseEdit = newProjectEditState(0, ""); m.caseEdit.Input.SetValue("sentinel") }, func(m *Model) string { return m.caseEdit.Input.Value() }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			model := NewModel(Options{})
			tc.setup(model)
			model.Update(tea.KeyMsg{Type: tea.KeyEsc})

			if !model.draftConfirmation.Active || tc.read(model) != "sentinel" {
				t.Fatal("Escape discarded unsent input")
			}

			model.Update(tea.KeyMsg{Type: tea.KeyEsc})

			if tc.read(model) != "sentinel" {
				t.Fatal("cancelled discard lost input")
			}

			model.Update(tea.KeyMsg{Type: tea.KeyEsc})
			model.Update(tea.KeyMsg{Type: tea.KeyEnter})

			if tc.read(model) != "" {
				t.Fatal("confirmed discard did not release the editor")
			}
		})
	}
}

// The real shell Action panel exposes distinct outcomes and complete findings;
// successful results stay available without stealing focus from the next input.
func TestOperationFeedback(t *testing.T) {
	for _, status := range []string{operation.OK, operation.ValidationFailed, operation.Stale, operation.Blocked} {
		t.Run(status, func(t *testing.T) {
			model := NewModel(Options{})
			model.Update(tea.WindowSizeMsg{Width: 110, Height: 30})
			model.recordOperation(operation.Result{Status: status, Summary: "Concrete result", ChangedFiles: []string{"ops/default/dump/example.md"}, Validation: []string{"content: keep entered text"}, BlockedReason: "Exact reason"})

			if status == operation.OK {
				model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o"), Alt: true})
			}

			for _, want := range []string{"Action", status, "Concrete result", "content: keep entered text", "Changed: ops/default/dump/example.md", "Exact reason"} {
				if !strings.Contains(model.View(), want) {
					t.Fatal("Action panel omitted outcome evidence", want, model.View())
				}
			}
		})
	}
}

// A delayed successful answer cannot close a newly opened editor even when its
// case and empty starting text match the submitted interaction.
func TestOperationNewDraft(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	model := NewModel(Options{Quizzes: []QuizView{{ActiveQuestionSummary: "Which input?"}}})
	model.selectView(ViewQuizzes)
	model.startQuizAnswer()
	model.quizAnswer.Input.SetValue("Original answer")

	model.quizExecutor = feedbackQuizAction(func(QuizActionRequest) (QuizActionResult, error) {
		close(started)
		<-release

		return QuizActionResult{Message: "Old answer accepted"}, nil
	})
	command := model.saveQuizAnswer()

	messages := make(chan tea.Msg, 1)
	go func() { messages <- command() }()

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		close(release)
		t.Fatal("owned executor never started")
	}

	model.startQuizAnswer()
	model.quizAnswer.Input.SetValue("New unsent sentinel")
	close(release)

	select {
	case message := <-messages:
		model.Update(message)
	case <-time.After(5 * time.Second):
		t.Fatal("owned executor did not return")
	}

	if !model.quizAnswer.Active || model.quizAnswer.Input.Value() != "New unsent sentinel" || model.statusMessage == "Old answer accepted" {
		t.Fatal("old acknowledgement replaced the new draft")
	}
}

type feedbackQuizAction func(QuizActionRequest) (QuizActionResult, error)

func (execute feedbackQuizAction) ExecuteQuizAction(request QuizActionRequest) (QuizActionResult, error) {
	return execute(request)
}

// A result can be inspected after navigation, but recovery must not reuse it
// against a newer draft that the user has typed since the failure.
func TestOperationChangedRetry(t *testing.T) {
	executor := &feedbackDumpExecutor{pending: true}
	model := NewModel(Options{DumpExecutor: executor})
	model.selectView(ViewDumps)
	model.startDumpEdit("bound.md", "", "rev")
	model.dumpEdit.Input.SetValue("Original sentinel")
	model.saveDumpEdit()
	model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" newer text")})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o"), Alt: true})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})

	if len(executor.requests) != 1 || !strings.Contains(model.dumpEdit.Input.Value(), "newer text") {
		t.Fatal("old recovery submitted or discarded the newer draft")
	}
}

// Long finding lists use the same Action panel and remain readable by scrolling
// without submitting, clearing the result or changing its originating view.
func TestOperationScroll(t *testing.T) {
	model := NewModel(Options{})
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 18})

	findings := make([]string, 30)
	for i := range findings {
		findings[i] = "Field finding"
	}

	findings[len(findings)-1] = "LAST_FINDING_SENTINEL"
	model.recordOperation(operation.Result{Status: operation.ValidationFailed, Validation: findings})
	model.Update(tea.KeyMsg{Type: tea.KeyEnd})

	if !strings.Contains(model.View(), "LAST_FINDING_SENTINEL") || model.SelectedView() != ViewCases {
		t.Fatal("Action scroll hid the final finding or changed the view", model.View())
	}
}

// A confirmation belongs to its captured editor generation. A later editor
// cannot be discarded by accepting an older prompt, even with matching text.
func TestOperationNewConfirmation(t *testing.T) {
	model := NewModel(Options{})
	model.selectView(ViewDumps)
	model.startDumpEdit("first.md", "Original", "rev")
	model.dumpEdit.Input.SetValue("Sentinel")
	model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model.startDumpEdit("second.md", "Sentinel", "new")
	model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !model.dumpEdit.Active || model.dumpEdit.Path != "second.md" || model.dumpEdit.Input.Value() != "Sentinel" {
		t.Fatal("obsolete confirmation discarded another editor")
	}
}

// Native execution and status may overlap; an older observed revision cannot
// replace a newer workflow or install its obsolete confirmation controls.
func TestOperationLateRevision(t *testing.T) {
	model := NewModel(Options{})
	model.workflow = WorkflowView{ID: "run", Revision: 4}
	model.workExecutor = feedbackWorkAction(func(WorkActionRequest) (WorkActionResult, error) {
		return WorkActionResult{Workflow: &WorkflowView{ID: "run", Revision: 4}, Confirmation: &WorkActionState{Pending: true, Title: "Old confirmation"}}, nil
	})
	command := model.executeWorkActionCommand(WorkActionRequest{Action: WorkActionWorkflowStatus, Workflow: &WorkflowActionRequest{ID: "run", ExpectedRevision: 4}})
	model.workflow.Revision = 5
	model.workAction = WorkActionState{Pending: true, Title: "Newer confirmation"}
	model.Update(command())

	if model.workflow.Revision != 5 || model.workAction.Title != "Newer confirmation" {
		t.Fatal("older workflow observation replaced current authorization")
	}
}

type feedbackWorkAction func(WorkActionRequest) (WorkActionResult, error)

func (execute feedbackWorkAction) ExecuteWorkAction(request WorkActionRequest) (WorkActionResult, error) {
	return execute(request)
}
