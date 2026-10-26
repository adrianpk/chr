package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

// AuthoringSessionView is the terminal projection of an authoring session.
type AuthoringSessionView struct {
	Index          int
	ID             string
	Status         string
	Mode           string
	Target         string
	Topic          string
	UpdatedAt      time.Time
	SessionPath    string
	TranscriptPath string
	ArtifactsPath  string
	Inputs         []string
	Outputs        []string
}

// NewAuthoringSessionViews converts core authoring sessions into TUI rows.
func NewAuthoringSessionViews(sessions []core.AuthoringSession) []AuthoringSessionView {
	views := make([]AuthoringSessionView, 0, len(sessions))

	for index, session := range sessions {
		paths := core.AuthoringSessionLayout("", session.ID)
		views = append(views, AuthoringSessionView{
			Index:          index,
			ID:             session.ID,
			Status:         string(session.Status),
			Mode:           string(session.Mode),
			Target:         string(session.Target),
			Topic:          session.Topic,
			UpdatedAt:      session.UpdatedAt,
			SessionPath:    filepath.ToSlash(paths.SessionPath),
			TranscriptPath: filepath.ToSlash(paths.TranscriptPath),
			ArtifactsPath:  filepath.ToSlash(paths.ArtifactsPath),
			Inputs:         authoringArtifactPaths(session.Inputs),
			Outputs:        authoringArtifactPaths(session.Outputs),
		})
	}

	return views
}

func (m *Model) renderAuthoring() string {
	if len(m.authoringSessions) == 0 {
		return "Authoring\n  No authoring sessions were found."
	}

	var output strings.Builder

	output.WriteString("Authoring\n")
	output.WriteString("  Status      Target     Topic\n")

	for _, session := range m.authoringSessions {
		prefix := "  "
		if session.Index == m.selectedAuthoringSession {
			prefix = "> "
		}

		output.WriteString(prefix)
		output.WriteString(renderAuthoringSessionRow(session))
		output.WriteString("\n")
	}

	output.WriteString("\n")
	output.WriteString(m.renderSelectedAuthoringSessionDetail())

	return strings.TrimRight(output.String(), "\n")
}

func renderAuthoringSessionRow(session AuthoringSessionView) string {
	return fmt.Sprintf(
		"%-11s %-10s %s",
		session.Status,
		session.Target,
		session.Topic,
	)
}

func (m *Model) renderSelectedAuthoringSessionDetail() string {
	session, ok := m.selectedAuthoringSessionView()
	if !ok {
		return "Selected Authoring Session\n  No session selected."
	}

	var output strings.Builder

	output.WriteString("Selected Authoring Session\n")
	fmt.Fprintf(&output, "  Topic: %s\n", session.Topic)
	fmt.Fprintf(&output, "  ID: %s\n", session.ID)
	fmt.Fprintf(&output, "  Status: %s\n", session.Status)
	fmt.Fprintf(&output, "  Mode: %s\n", session.Mode)
	fmt.Fprintf(&output, "  Target: %s\n", session.Target)
	fmt.Fprintf(&output, "  Updated: %s\n", renderTime(session.UpdatedAt))
	fmt.Fprintf(&output, "  Session: %s\n", session.SessionPath)
	fmt.Fprintf(&output, "  Transcript: %s\n", session.TranscriptPath)
	fmt.Fprintf(&output, "  Artifacts: %s\n", session.ArtifactsPath)
	fmt.Fprintf(&output, "  Inputs: %s\n", renderStringList(session.Inputs))
	fmt.Fprintf(&output, "  Outputs: %s", renderStringList(session.Outputs))

	return output.String()
}

func (m *Model) selectedAuthoringSessionView() (AuthoringSessionView, bool) {
	for _, session := range m.authoringSessions {
		if session.Index == m.selectedAuthoringSession {
			return session, true
		}
	}

	return AuthoringSessionView{}, false
}

func (m *Model) selectPreviousAuthoringSession() {
	if len(m.authoringSessions) == 0 {
		return
	}

	if m.selectedAuthoringSession == 0 {
		m.selectedAuthoringSession = m.authoringSessions[len(m.authoringSessions)-1].Index

		return
	}

	m.selectedAuthoringSession--
}

func (m *Model) selectNextAuthoringSession() {
	if len(m.authoringSessions) == 0 {
		return
	}

	m.selectedAuthoringSession++
	if m.selectedAuthoringSession >= len(m.authoringSessions) {
		m.selectedAuthoringSession = 0
	}
}

func authoringArtifactPaths(refs []core.AuthoringArtifactRef) []string {
	paths := make([]string, 0, len(refs))
	for _, ref := range refs {
		if ref.Path != "" {
			paths = append(paths, ref.Path)
		}
	}

	return paths
}

func renderStringList(values []string) string {
	if len(values) == 0 {
		return "none"
	}

	return strings.Join(values, ", ")
}
