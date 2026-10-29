package tui

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// CommandAction identifies the kind of command palette operation to execute.
type CommandAction string

// Command actions define command palette routing targets.
const (
	CommandActionNavigate           CommandAction = "navigate"
	CommandActionRefresh            CommandAction = "refresh"
	CommandActionHelp               CommandAction = "help"
	CommandActionOpenSelectedSource CommandAction = "open_selected_source"
	CommandActionPlanArtifact       CommandAction = "plan_artifact"
	CommandActionPlanDump           CommandAction = "plan_dump"
	CommandActionPlanQuiz           CommandAction = "plan_quiz"
	CommandActionPlanSpec           CommandAction = "plan_spec"
	CommandActionPlanPlan           CommandAction = "plan_plan"
	CommandActionPlanTicket         CommandAction = "plan_ticket"
	CommandActionPlanWork           CommandAction = "plan_work"
	CommandActionGuided             CommandAction = "guided"
	CommandActionClearCase          CommandAction = "clear_case"
)

// Command describes one command palette entry.
type Command struct {
	ID          string
	Title       string
	Description string
	Action      CommandAction
	View        View
	Artifact    ArtifactActionID
	Dump        DumpActionID
	Quiz        QuizActionID
	Spec        SpecActionID
	Plan        PlanActionID
	Ticket      TicketActionID
	Status      string
	Work        WorkActionID
}

// CommandPaletteState stores command palette filtering and selection state.
type CommandPaletteState struct {
	Open     bool
	Filter   string
	Selected int
	Commands []Command
}

// NewCommandPaletteState initializes command palette state with a copy of the
// available commands.
func NewCommandPaletteState(commands []Command) CommandPaletteState {
	return CommandPaletteState{
		Commands: append([]Command(nil), commands...),
	}
}

// DefaultCommandRegistry returns the built-in TUI command palette entries.
func DefaultCommandRegistry() []Command {
	commands := []Command{
		{ID: "refresh", Title: "Refresh", Description: "Reload workspace files into the runtime graph.", Action: CommandActionRefresh},
		{ID: "clear-case-context", Title: "Clear case context", Description: "Show all case scopes without changing documents.", Action: CommandActionClearCase},
		{ID: "help", Title: "Help", Description: "Navigate to the Help view.", Action: CommandActionHelp, View: ViewHelp},
		{ID: "open-source", Title: "Open selected source path", Description: "Show the selected source path for external opening.", Action: CommandActionOpenSelectedSource},
		{ID: "create-ticket", Title: "Create ticket", Description: "Create a default task ticket after confirmation.", Action: CommandActionPlanArtifact, Artifact: ArtifactActionCreateTicket},
		{ID: "append-dump", Title: "Append to selected dump", Description: "Enter text for the captured dump and revision.", Action: CommandActionPlanDump, Dump: DumpActionAppend},
		{ID: "create-dump", Title: "Create dump", Description: "Enter scope, title and body before one Create.", Action: CommandActionPlanDump, Dump: DumpActionCreate},
		{ID: "edit-dump", Title: "Edit dump", Description: "Open the selected dump Markdown file in the editor.", Action: CommandActionPlanDump, Dump: DumpActionEdit},
		{ID: "refresh-dumps", Title: "Refresh dumps", Description: "Reload scoped dumps.", Action: CommandActionPlanDump, Dump: DumpActionRefresh},
		{ID: "quiz-start", Title: "Start selected quiz", Description: "Start the selected scoped quiz.", Action: CommandActionPlanQuiz, Quiz: QuizActionStart},
		{ID: "quiz-resume", Title: "Resume selected quiz", Description: "Reload the selected scoped quiz runtime.", Action: CommandActionPlanQuiz, Quiz: QuizActionResume},
		{ID: "quiz-answer", Title: "Answer selected quiz", Description: "Enter an answer for the active quiz question.", Action: CommandActionPlanQuiz, Quiz: QuizActionAnswer},
		{ID: "quiz-skip", Title: "Skip quiz question", Description: "Skip the active selected quiz question.", Action: CommandActionPlanQuiz, Quiz: QuizActionSkip},
		{ID: "quiz-complete", Title: "Complete selected quiz", Description: "Mark the selected quiz as complete.", Action: CommandActionPlanQuiz, Quiz: QuizActionComplete},
		{ID: "quiz-cancel", Title: "Cancel selected quiz", Description: "Cancel the selected quiz loop.", Action: CommandActionPlanQuiz, Quiz: QuizActionCancel},
		{ID: "quiz-refresh", Title: "Refresh quizzes", Description: "Reload scoped quiz state.", Action: CommandActionPlanQuiz, Quiz: QuizActionRefresh},
		{ID: "quiz-mode-deep", Title: "Switch quiz to deep mode", Description: "Use deep generation for the selected quiz.", Action: CommandActionPlanQuiz, Quiz: QuizActionDeepMode},
		{ID: "quiz-mode-fast", Title: "Switch quiz to fast mode", Description: "Use fast candidate generation for the selected quiz.", Action: CommandActionPlanQuiz, Quiz: QuizActionFastMode},
		{ID: "generate-spec", Title: "Generate spec", Description: "Generate a draft spec from available project and use-case context.", Action: CommandActionPlanSpec, Spec: SpecActionGenerate},
		{ID: "generate-acceptance-feature", Title: "Generate acceptance feature", Description: "Generate a Gherkin acceptance feature from the selected spec context.", Action: CommandActionPlanSpec, Spec: SpecActionGenerateFeature},
		{ID: "refresh-specs", Title: "Refresh specs", Description: "Reload scoped specs.", Action: CommandActionPlanSpec, Spec: SpecActionRefresh},
		{ID: "generate-plan", Title: "Generate plan", Description: "Generate an implementation plan from available project and use-case context.", Action: CommandActionPlanPlan, Plan: PlanActionGenerate},
		{ID: "regenerate-plan", Title: "Regenerate plan", Description: "Explicitly replace the selected implementation plan.", Action: CommandActionPlanPlan, Plan: PlanActionRegenerate},
		{ID: "materialize-plan-tracking", Title: "Create plan tracking", Description: "Create tracking and tickets from the selected plan.", Action: CommandActionPlanPlan, Plan: PlanActionMaterializeTracking},
		{ID: "refresh-plans", Title: "Refresh plans", Description: "Reload scoped plans.", Action: CommandActionPlanPlan, Plan: PlanActionRefresh},
		{ID: "create-authoring-session", Title: "Create authoring session", Description: "Create a spec-oriented authoring session after confirmation.", Action: CommandActionPlanArtifact, Artifact: ArtifactActionCreateAuthoringSession},
		{ID: "create-context-quiz", Title: "Create context-aware quiz", Description: "Plan a context-aware quiz creation action.", Action: CommandActionPlanArtifact, Artifact: ArtifactActionCreateContextQuiz},
		{ID: "ticket-add-comment", Title: "Add ticket comment", Description: "Add a standard comment to the selected ticket after confirmation.", Action: CommandActionPlanTicket, Ticket: TicketActionAddComment},
		{ID: "ticket-status-backlog", Title: "Move ticket to backlog", Description: "Move the selected ticket to backlog through canonical guards.", Action: CommandActionPlanTicket, Ticket: TicketActionMoveStatus, Status: "backlog"},
		{ID: "ticket-status-ready-for-dev", Title: "Move ticket to ready_for_dev", Description: "Move the selected ticket to ready_for_dev through canonical guards.", Action: CommandActionPlanTicket, Ticket: TicketActionMoveStatus, Status: "ready_for_dev"},
		{ID: "ticket-status-in-progress", Title: "Move ticket to in_progress", Description: "Move the selected ticket to in_progress through canonical guards.", Action: CommandActionPlanTicket, Ticket: TicketActionMoveStatus, Status: "in_progress"},
		{ID: "ticket-status-ready-for-review", Title: "Move ticket to ready_for_review", Description: "Move the selected ticket to ready_for_review through canonical guards.", Action: CommandActionPlanTicket, Ticket: TicketActionMoveStatus, Status: "ready_for_review"},
		{ID: "ticket-status-ready-for-merge", Title: "Move ticket to ready_for_merge", Description: "Move the selected ticket to ready_for_merge through canonical guards.", Action: CommandActionPlanTicket, Ticket: TicketActionMoveStatus, Status: "ready_for_merge"},
		{ID: "ticket-status-done", Title: "Move ticket to done", Description: "Move the selected ticket to done through canonical guards.", Action: CommandActionPlanTicket, Ticket: TicketActionMoveStatus, Status: "done"},
		{ID: "ticket-checklist-toggle", Title: "Toggle ticket checklist item", Description: "Validate checklist toggle availability for the selected ticket.", Action: CommandActionPlanTicket, Ticket: TicketActionToggleChecklist},
		{ID: "ticket-checklist-add", Title: "Add ticket checklist item", Description: "Validate checklist creation availability for the selected ticket.", Action: CommandActionPlanTicket, Ticket: TicketActionAddChecklistItem},
		{ID: "work-start-ticket", Title: "Start selected ticket", Description: "Run the agent harness for the selected ticket after confirmation.", Action: CommandActionPlanWork, Work: WorkActionStartTicket},
		{ID: "work-start-ready-loop", Title: "Start ready ticket loop", Description: "Run one ready ticket through the agent loop after confirmation.", Action: CommandActionPlanWork, Work: WorkActionStartReadyLoop},
		{ID: "work-run-all-ready", Title: "Run all ready tickets", Description: "Run all ready_for_dev tickets through the agent loop after confirmation.", Action: CommandActionPlanWork, Work: WorkActionRunAllReady},
		{ID: "work-repair-readiness", Title: "Apply safe readiness repair", Description: "Run safe work readiness repair after confirmation.", Action: CommandActionPlanWork, Work: WorkActionRepairReadiness},
		{ID: "work-request-stop", Title: "Request agent stop", Description: "Request daemon stop after confirmation.", Action: CommandActionPlanWork, Work: WorkActionRequestStop},
		{ID: "guided-use-cases", Title: "Guided use cases", Description: "Develop a new use case or continue a named guided run.", Action: CommandActionGuided},
	}

	commands = append(commands, nativeWorkflowCommands()...)

	for _, view := range append(slices.Clone(defaultViews), auxiliaryViews...) {
		commands = append(commands, Command{
			ID:          "view:" + strings.ToLower(string(view)),
			Title:       "Go to " + string(view),
			Description: "Navigate to the " + string(view) + " view.",
			Action:      CommandActionNavigate,
			View:        view,
		})
	}

	return commands
}

func (m *Model) renderCommandPalette() string {
	if !m.commandPalette.Open {
		return ""
	}

	var output strings.Builder

	output.WriteString("\n\nCommand Palette\n")
	fmt.Fprintf(&output, "  Filter: %s\n", m.commandPalette.Filter)

	commands := m.filteredCommands()
	if len(commands) == 0 {
		output.WriteString("  No commands match.")

		return output.String()
	}

	for index, command := range commands {
		prefix := "  "
		if index == m.commandPalette.Selected {
			prefix = "> "
		}

		output.WriteString(prefix)
		output.WriteString(command.Title)
		output.WriteString(" - ")
		output.WriteString(command.Description)
		output.WriteString("\n")
	}

	return strings.TrimRight(output.String(), "\n")
}

func (m *Model) updateCommandPaletteKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc:
		m.closeCommandPalette()
	case tea.KeyEnter:
		command := m.executeSelectedCommand()

		return command
	case tea.KeyBackspace:
		m.commandPalette.Filter = strings.TrimSuffix(m.commandPalette.Filter, lastRune(m.commandPalette.Filter))
		m.boundSelectedCommand()
	case tea.KeyUp:
		m.selectPreviousCommand()
	case tea.KeyDown:
		m.selectNextCommand()
	default:
		return m.updateCommandPaletteRunes(msg)
	}

	return nil
}

func (m *Model) updateCommandPaletteRunes(msg tea.KeyMsg) tea.Cmd {
	if msg.Type != tea.KeyRunes {
		return nil
	}

	m.commandPalette.Filter += msg.String()
	m.commandPalette.Selected = 0

	return nil
}

func (m *Model) openCommandPalette() {
	m.commandPalette.Open = true
	m.commandPalette.Filter = ""
	m.commandPalette.Selected = 0
	m.statusMessage = ""
}

func (m *Model) closeCommandPalette() {
	m.commandPalette.Open = false
	m.commandPalette.Filter = ""
	m.commandPalette.Selected = 0
}

func (m *Model) selectPreviousCommand() {
	commands := m.filteredCommands()
	if len(commands) == 0 {
		return
	}

	if m.commandPalette.Selected == 0 {
		m.commandPalette.Selected = len(commands) - 1

		return
	}

	m.commandPalette.Selected--
}

func (m *Model) selectNextCommand() {
	commands := m.filteredCommands()
	if len(commands) == 0 {
		return
	}

	m.commandPalette.Selected++
	if m.commandPalette.Selected >= len(commands) {
		m.commandPalette.Selected = 0
	}
}

func (m *Model) boundSelectedCommand() {
	commands := m.filteredCommands()
	if len(commands) == 0 || m.commandPalette.Selected >= len(commands) {
		m.commandPalette.Selected = 0
	}
}

func (m *Model) executeSelectedCommand() tea.Cmd {
	commands := m.filteredCommands()
	if len(commands) == 0 {
		m.statusMessage = "No command selected."
		m.closeCommandPalette()

		return nil
	}

	command := commands[m.commandPalette.Selected]
	m.closeCommandPalette()

	return m.executeCommand(command)
}

//nolint:cyclop // Central command dispatch keeps palette behavior easy to audit.
func (m *Model) executeCommand(command Command) tea.Cmd {
	switch command.Action {
	case CommandActionClearCase:
		m.bindCaseContext("", "")

		m.selectedCase = -1
		m.statusMessage = "Case context cleared. All scopes are visible."

		return nil
	case CommandActionGuided:
		return m.openGuided()
	case CommandActionNavigate, CommandActionHelp:
		m.selectView(command.View)

		m.statusMessage = "Navigated to " + string(command.View) + "."
	case CommandActionRefresh:
		m.refreshWorkspace()
	case CommandActionOpenSelectedSource:
		m.statusMessage = m.selectedSourcePathMessage()
	case CommandActionPlanArtifact:
		m.planArtifactAction(command.Artifact)
	case CommandActionPlanDump:
		m.planDumpAction(command.Dump)
	case CommandActionPlanQuiz:
		m.planQuizAction(command.Quiz)
	case CommandActionPlanSpec:
		m.planSpecAction(command.Spec)
	case CommandActionPlanPlan:
		m.planPlanAction(command.Plan)
	case CommandActionPlanTicket:
		m.planTicketAction(command.Ticket, command.Status)
	case CommandActionPlanWork:
		m.planWorkAction(command.Work)
	}

	return nil
}

func (m *Model) selectView(view View) {
	if m.SelectedView() == ViewCases && view != ViewCases {
		m.captureCaseContext()
	}

	if view == ViewHelp && m.SelectedView() != ViewHelp {
		m.helpReturnView, m.helpReturnPane = m.SelectedView(), m.selectedPane
	}

	index := slices.Index(m.views, view)
	if index < 0 && !slices.Contains(auxiliaryViews, view) {
		return
	}

	m.auxiliaryView = ""
	if index < 0 {
		m.auxiliaryView = view
	} else {
		m.selected = index
	}

	m.selectPrimaryPane()

	m.maximizedPane = -1
	m.reconcileCaseTicketSelection()
	m.explainCaseContext()
}

//nolint:wsl_v5 // Filtering keeps predicate and append adjacent.
func (m *Model) filteredCommands() []Command {
	filter := strings.ToLower(strings.TrimSpace(m.commandPalette.Filter))
	if filter == "" {
		commands := []Command{}
		for _, command := range m.commandPalette.Commands {
			if m.commandAvailableInCurrentView(command) {
				commands = append(commands, command)
			}
		}

		return commands
	}

	commands := []Command{}

	for _, command := range m.commandPalette.Commands {
		if !m.commandAvailableInCurrentView(command) {
			continue
		}

		if command.Matches(filter) {
			commands = append(commands, command)
		}
	}

	return commands
}

func (m *Model) commandAvailableInCurrentView(command Command) bool {
	if command.Action == CommandActionPlanSpec {
		return m.SelectedView() == ViewSpecs
	}

	if command.Action == CommandActionPlanPlan {
		if command.Plan == PlanActionMaterializeTracking {
			return m.SelectedView() == ViewPlans && m.selectedPlanCanMaterializeTracking()
		}

		return m.SelectedView() == ViewPlans
	}

	return true
}

// Matches reports whether a command should be shown for the current filter.
func (c Command) Matches(filter string) bool {
	return strings.Contains(strings.ToLower(c.ID), filter) ||
		strings.Contains(strings.ToLower(c.Title), filter) ||
		strings.Contains(strings.ToLower(c.Description), filter)
}

func (m *Model) selectedSourcePathMessage() string {
	path := m.selectedSourcePath()
	if path == "" {
		return "No selected source path is available."
	}

	return "Selected source path: " + path
}

//nolint:cyclop // Source resolution is intentionally centralized by view.
func (m *Model) selectedSourcePath() string {
	switch m.SelectedView() {
	case ViewBoard, ViewTickets:
		ticket, ok := m.selectedTicketView()
		if ok {
			return ticketDisplayPath(ticket.RelativePath)
		}
	case ViewDumps:
		dump, ok := m.selectedDumpView()
		if ok {
			return dump.RelativePath
		}
	case ViewQuizzes:
		return m.selectedQuizSourcePath()
	case ViewSpecs:
		spec, ok := m.selectedSpecView()
		if ok {
			return spec.RelativePath
		}
	case ViewPlans:
		plan, ok := m.selectedPlanView()
		if ok {
			return plan.RelativePath
		}
	case ViewDocuments:
		return m.selectedDocumentSourcePath()
	case ViewAuthoring:
		session, ok := m.selectedAuthoringSessionView()
		if ok {
			return session.SessionPath
		}
	}

	return ""
}

func (m *Model) selectedQuizSourcePath() string {
	quiz, ok := m.selectedQuizView()
	if !ok {
		return ""
	}

	if strings.TrimSpace(quiz.CuratedPath) != "" {
		return quiz.CuratedPath
	}

	return quiz.RawPath
}

func (m *Model) selectedDocumentSourcePath() string {
	document, ok := m.selectedDocumentViewFor(m.currentDocumentViews())
	if !ok {
		return ""
	}

	return document.RelativePath
}
