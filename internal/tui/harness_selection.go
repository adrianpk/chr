package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"forge.adrianpk.com/chr/internal/core"
)

// HarnessSelectionExecutor observes shared project defaults and retained work.
type HarnessSelectionExecutor interface {
	InspectHarnessSelections() (HarnessSelectionView, error)
	SaveHarnessSelections(HarnessSelectionRequest) (HarnessSelectionView, error)
}

// HarnessSelectionView contains public selection identities, never private paths.
type HarnessSelectionView struct {
	RequestID          string
	Snapshot           core.HarnessSelectionSnapshot
	Worker, Generation core.HarnessSelection
	Profiles           []HarnessProfileView
	Retained           []HarnessRetainedView
	More               bool
	RetainedProblem    string
}

// HarnessProfileView offers named choices with role-specific local readiness.
type HarnessProfileView struct {
	Name, Harness, Provider, Model string
	Worker, Generation             core.HarnessReadiness
}

// HarnessRetainedView separates a retained worker binding from current defaults.
type HarnessRetainedView struct {
	ID, Status string
	Binding    *core.HarnessBinding
}

const harnessKeepProfile = ":keep"

type harnessSelectionState struct {
	Active, Busy       bool
	Saving             bool
	Session            uint64
	View               HarnessSelectionView
	Worker, Generation string
	Field, Scroll      int
	Message            string
}

type harnessSelectionLoadedMsg struct {
	Session uint64
	View    HarnessSelectionView
	Err     error
}

func (m *Model) openHarnessPicker() tea.Cmd {
	if m.harnessExecutor == nil {
		m.statusMessage = "Harness selection is unavailable."

		return nil
	}

	m.harnessSelection = harnessSelectionState{Active: true}

	return m.reloadHarnessPicker()
}

func (m *Model) reloadHarnessPicker() tea.Cmd {
	m.harnessSequence++
	m.harnessSelection.Session = m.harnessSequence
	m.harnessSelection.Busy = true
	m.harnessSelection.Message = "Loading current defaults…"
	session, executor := m.harnessSequence, m.harnessExecutor

	return func() tea.Msg {
		view, err := executor.InspectHarnessSelections()

		return harnessSelectionLoadedMsg{Session: session, View: view, Err: err}
	}
}

func (m *Model) completeHarnessRead(msg harnessSelectionLoadedMsg) {
	state := &m.harnessSelection
	if !state.Active || !state.Busy || state.Saving || state.Session != msg.Session {
		return
	}

	state.Busy = false
	if msg.Err != nil {
		state.Message = "Cannot load choices. Check private profile setup and reload."

		return
	}

	state.View = msg.View
	state.Worker, state.Generation = harnessKeepProfile, harnessKeepProfile
	state.Scroll, state.Message = 0, ""
}

func (m *Model) updateHarnessKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.harnessSelection.Saving {
		return m, nil
	}

	if msg.Type == tea.KeyEsc {
		m.harnessSelection = harnessSelectionState{}

		return m, nil
	}

	if m.harnessSelection.Busy {
		return m, nil
	}

	switch msg.String() {
	case "r":
		return m, m.reloadHarnessPicker()
	case "enter", "ctrl+s":
		return m, m.saveHarnessPicker()
	default:
		m.updateHarnessChoiceKey(msg.String())
	}

	return m, nil
}

func (m *Model) updateHarnessChoiceKey(key string) {
	switch key {
	case "tab", "shift+tab":
		m.harnessSelection.Scroll = 0
		m.harnessSelection.Field = 1 - m.harnessSelection.Field
	case "left", "up", "k":
		m.cycleHarnessProfile(-1)
	case "right", "down", "j":
		m.cycleHarnessProfile(1)
	case "c":
		m.harnessSelection.Scroll = 0
		*m.harnessChoice() = ""
	case "pgdown":
		maximum := maxInt(len(m.harnessSelectionLines(maxInt(m.width-4, 56)))-m.harnessVisibleRows(), 0)
		m.harnessSelection.Scroll = minInt(m.harnessSelection.Scroll+1, maximum)
	case "pgup":
		m.harnessSelection.Scroll = maxInt(m.harnessSelection.Scroll-1, 0)
	}
}

func (m *Model) harnessChoice() *string {
	if m.harnessSelection.Field == 0 {
		return &m.harnessSelection.Worker
	}

	return &m.harnessSelection.Generation
}

func (m *Model) cycleHarnessProfile(direction int) {
	m.harnessSelection.Scroll = 0

	choices := []string{harnessKeepProfile, ""}
	for _, profile := range m.harnessSelection.View.Profiles {
		choices = append(choices, profile.Name)
	}

	selected := m.harnessChoice()
	for index, choice := range choices {
		if choice == *selected {
			*selected = choices[(index+direction+len(choices))%len(choices)]

			return
		}
	}

	*selected = harnessKeepProfile
}

func (m *Model) renderHarnessWorkspace(width, height int) string {
	keys := []string{"Tab: role · arrows: profile · c: clear", "Enter: save · r: reload/discard · Esc: cancel · PgUp/Dn"}
	content := m.harnessSelectionLines(maxInt(width-4, 1))
	lines := append(keys, scrollLines(content, m.harnessSelection.Scroll, maxInt(height-3-len(keys), 1))...)

	return renderPanel("Harness defaults", lines, width, height, true)
}

func (m *Model) harnessSelectionLines(width int) []string {
	state := m.harnessSelection

	lines := harnessWrap("New operations only. Running, paused and resumed work keeps its binding.", width)
	if state.Message != "" {
		lines = append(lines, harnessWrap(state.Message, width)...)
	}

	if state.View.Snapshot.Revision == "" {
		return lines
	}

	lines = append(lines, harnessWrap(m.harnessChoiceLine("Worker", state.Worker, state.Field == 0), width)...)
	lines = append(lines, harnessWrap(m.harnessChoiceLine("Generation", state.Generation, state.Field == 1), width)...)
	lines = append(lines, harnessRoleLines("Worker", state.View.Worker, false, width)...)
	lines = append(lines, harnessRoleLines("Generation", state.View.Generation, true, width)...)
	lines = append(lines, "Retained worker bindings:")

	for _, retained := range state.View.Retained {
		label := "Legacy worker (original settings)"
		if b := retained.Binding; b != nil {
			label = harnessIdentity(b.Profile, b.Harness, b.ModelProvider, b.Model)
		}

		lines = append(lines, harnessWrap(retained.ID+" · "+retained.Status+" · "+label, width)...)
	}

	if state.View.RetainedProblem != "" {
		lines = append(lines, harnessWrap(state.View.RetainedProblem, width)...)
	}

	if state.View.More {
		lines = append(lines, "First four retained invocations; use workflow status.")
	}

	return lines
}

func harnessRoleLines(name string, selection core.HarnessSelection, generation bool, width int) []string {
	label := harnessIdentity(selection.Profile, selection.Harness, selection.ModelProvider, selection.Model)
	if selection.Profile == "" {
		label = "Legacy dispatch · Codex CLI default"
		if generation {
			label = "Legacy dispatch · varies by operation"
		}
	}

	r := selection.Readiness
	setup := fmt.Sprintf("Local %s; credentials %s; runtime %s; remote auth %s.", harnessReadyLabel(r), r.Credentials, r.Runtime, r.RemoteAuthentication)
	lines := harnessWrap(name+" default: "+label, width)

	lines = append(lines, harnessWrap(strings.ReplaceAll(setup, "_", " "), width)...)
	if r.Problem != "" {
		lines = append(lines, harnessWrap(r.Problem, width)...)
	}

	return lines
}

func (m *Model) harnessChoiceLine(role, choice string, focused bool) string {
	label := "Unavailable: " + choice
	if choice == harnessKeepProfile {
		label = "Keep current default"
	}

	if choice == "" {
		label = "Restore legacy dispatch"
	}

	for _, profile := range m.harnessSelection.View.Profiles {
		if profile.Name != choice {
			continue
		}

		ready := profile.Worker
		if role == "Generation" {
			ready = profile.Generation
		}

		label = harnessIdentity(profile.Name, profile.Harness, profile.Provider, profile.Model) + " · " + harnessReadyLabel(ready)
	}

	prefix := "  "
	if focused {
		prefix = "> "
	}

	return prefix + role + " choice: " + label
}

func harnessReadyLabel(readiness core.HarnessReadiness) string {
	if readiness.Ready() {
		return "ready"
	}

	return "not ready"
}

func harnessIdentity(profile, identity, provider, model string) string {
	identity = map[string]string{"codex": "Codex", "pi": "Pi"}[identity]

	if model == "cli_default" {
		model = "CLI default"
	}

	if provider != "" && provider != "cli_default" {
		model = provider + "/" + model
	}

	return strings.Trim(strings.Join([]string{profile, identity, model}, " · "), " ·")
}

func harnessWrap(text string, width int) []string {
	return strings.Split(lipgloss.NewStyle().Width(maxInt(width, 1)).Render(text), "\n")
}

func (m *Model) harnessVisibleRows() int {
	width := maxInt(m.width, 60)
	height := maxInt(m.height, minShellHeight)
	height -= lineCount(m.renderShellTopbar(width)) + lineCount(m.renderShellFooter(width)) + lineCount(m.renderShellStatus(width))

	return maxInt(height-5, 1)
}
