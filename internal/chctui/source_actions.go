package chctui

import (
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type SourceOpenExecutor interface {
	OpenSource(path string) (SourceOpenResult, error)
}

type SourceOpenResult struct {
	EditorCommand            *exec.Cmd
	EditorCommandDescription string
}

type sourceEditorFinishedMsg struct {
	err        error
	path       string
	command    string
	workingDir string
}

func (m *Model) openSelectedSource() tea.Cmd {
	path := strings.TrimSpace(m.selectedSourcePath())
	if path == "" {
		m.statusMessage = "No selected source path is available."

		return nil
	}

	if m.sourceExecutor == nil {
		m.statusMessage = "Source editor executor is unavailable."

		return nil
	}

	result, err := m.sourceExecutor.OpenSource(path)
	if err != nil {
		m.statusMessage = "Open source failed: " + err.Error()

		return nil
	}

	if result.EditorCommand == nil {
		m.statusMessage = "Open source failed: editor command is unavailable."

		return nil
	}

	m.statusMessage = "External editor opened. chc will resume when the editor process exits."

	return tea.ExecProcess(result.EditorCommand, func(err error) tea.Msg {
		return sourceEditorFinishedMsg{
			err:        err,
			path:       path,
			command:    result.EditorCommandDescription,
			workingDir: result.EditorCommand.Dir,
		}
	})
}

func (m *Model) completeSourceEdit(msg sourceEditorFinishedMsg) {
	if msg.err != nil {
		m.statusMessage = "Edit failed: " + sourceEditorErrorMessage(msg)

		return
	}

	if m.workspaceExecutor == nil {
		m.statusMessage = "Edit completed."

		return
	}

	m.refreshWorkspace()
}

func sourceEditorErrorMessage(msg sourceEditorFinishedMsg) string {
	parts := []string{}
	if strings.TrimSpace(msg.command) != "" {
		parts = append(parts, "command="+msg.command)
	}

	if strings.TrimSpace(msg.workingDir) != "" {
		parts = append(parts, "dir="+msg.workingDir)
	}

	parts = append(parts, msg.err.Error())

	return strings.Join(parts, " ")
}

func (m *Model) resetDetailScrollIfSourceChanged(previousPath string) {
	if previousPath != m.selectedSourcePath() {
		m.detailScroll = 0
	}
}

func (m *Model) scrollDetail(delta int) {
	if !m.detailPaneCanScroll() {
		return
	}

	m.detailScroll += delta
	if m.detailScroll < 0 {
		m.detailScroll = 0
	}
}

func (m *Model) scrollDetailPage(direction int) {
	if !m.detailPaneCanScroll() {
		return
	}

	pageSize := maxInt(m.height-5, 5)
	if direction < 0 {
		m.scrollDetail(-pageSize)

		return
	}

	m.scrollDetail(pageSize)
}

func (m Model) detailPaneCanScroll() bool {
	switch m.SelectedView() {
	case ViewBoard, ViewTickets, ViewDocuments, ViewDumps, ViewQuizzes, ViewSpecs, ViewPlans, ViewAuthoring, ViewAgent, ViewRuntime:
		return true
	default:
		return false
	}
}

func scrollLines(lines []string, offset, visible int) []string {
	if visible <= 0 || len(lines) <= visible {
		return lines
	}

	if offset < 0 {
		offset = 0
	}

	maxOffset := len(lines) - visible
	if offset > maxOffset {
		offset = maxOffset
	}

	return lines[offset : offset+visible]
}

func ticketDisplayPath(relativePath string) string {
	relativePath = strings.TrimSpace(relativePath)
	if relativePath == "" {
		return ""
	}

	if strings.HasPrefix(relativePath, ".chc/") || strings.HasPrefix(relativePath, ".prj/") {
		return relativePath
	}

	return fmt.Sprintf(".chc/tickets/%s", relativePath)
}
