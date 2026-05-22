package chctui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"forge.adrianpk.com/chc/internal/chc"
)

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

func NewAuthoringSessionViews(sessions []chc.AuthoringSession) []AuthoringSessionView {
	views := make([]AuthoringSessionView, 0, len(sessions))

	for index, session := range sessions {
		paths := chc.AuthoringSessionLayout("", session.ID)
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

func (m Model) renderAuthoring() string {
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

func (m Model) renderSelectedAuthoringSessionDetail() string {
	session, ok := m.selectedAuthoringSessionView()
	if !ok {
		return "Selected Authoring Session\n  No session selected."
	}

	var output strings.Builder

	output.WriteString("Selected Authoring Session\n")
	output.WriteString(fmt.Sprintf("  Topic: %s\n", session.Topic))
	output.WriteString(fmt.Sprintf("  ID: %s\n", session.ID))
	output.WriteString(fmt.Sprintf("  Status: %s\n", session.Status))
	output.WriteString(fmt.Sprintf("  Mode: %s\n", session.Mode))
	output.WriteString(fmt.Sprintf("  Target: %s\n", session.Target))
	output.WriteString(fmt.Sprintf("  Updated: %s\n", renderTime(session.UpdatedAt)))
	output.WriteString(fmt.Sprintf("  Session: %s\n", session.SessionPath))
	output.WriteString(fmt.Sprintf("  Transcript: %s\n", session.TranscriptPath))
	output.WriteString(fmt.Sprintf("  Artifacts: %s\n", session.ArtifactsPath))
	output.WriteString(fmt.Sprintf("  Inputs: %s\n", renderStringList(session.Inputs)))
	output.WriteString(fmt.Sprintf("  Outputs: %s", renderStringList(session.Outputs)))

	return output.String()
}

func (m Model) selectedAuthoringSessionView() (AuthoringSessionView, bool) {
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

func authoringArtifactPaths(refs []chc.AuthoringArtifactRef) []string {
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
