package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderOverlayWorkspace(width, height int, overlay string) string {
	content := []string{
		shellTitleStyle.Render("Active interaction"),
		"",
	}
	content = append(content, strings.Split(overlay, "\n")...)

	return renderPanel("Action", content, width, height, true)
}

func (m *Model) renderShellOverlay() string {
	var overlays []string

	for _, overlay := range []string{
		m.renderArtifactAction(),
		m.renderDumpAction(),
		m.renderTicketAction(),
		m.renderWorkAction(),
		m.renderQuizAction(),
		m.renderSpecAction(),
		m.renderPlanAction(),
		m.renderCommandPalette(),
	} {
		if strings.TrimSpace(overlay) != "" {
			overlays = append(overlays, strings.TrimSpace(overlay))
		}
	}

	return strings.Join(overlays, "\n\n")
}

func (m *Model) renderShellStatus(width int) string {
	if m.statusMessage == "" {
		return ""
	}

	return shellStatusStyle.Render(fitPlainLine(" "+m.statusMessage, width))
}

func (m *Model) renderShellFooter(width int) string {
	if m.guided.Active {
		return strings.Join(wrapPlainText(m.guidedHint(), width), "\n")
	}

	return m.renderShellButtons(shellFooterButtons(), width)
}

func shellFooterButtons() []shellButton {
	return []shellButton{
		{Key: "F1", Label: "Project", View: ViewProject},
		{Key: "F2", Label: "Cases", View: ViewCases},
		{Key: "F3", Label: "Dumps", View: ViewDumps},
		{Key: "F4", Label: "Quizzes", View: ViewQuizzes},
		{Key: "F5", Label: "Specs", View: ViewSpecs},
		{Key: "F6", Label: "Plans", View: ViewPlans},
		{Key: "F7", Label: "Board", View: ViewBoard},
		{Key: "F8", Label: "Dash", View: ViewDash},
		{Key: "F9", Label: "Help", View: ViewHelp},
		{Key: "F10", Label: "Exit"},
	}
}

func (m *Model) renderShellButtons(buttons []shellButton, width int) string {
	widths := distributedWidths(width, len(buttons))
	cells := make([]string, 0, len(buttons))

	for index, button := range buttons {
		style := shellButtonInactiveStyle
		if button.View != "" && button.View == m.SelectedView() {
			style = shellButtonStyle
		}

		cellWidth := widths[index]
		style = style.Width(cellWidth)
		cells = append(cells, style.Render(centerShellButtonLines(button.Key, button.Label, cellWidth)))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
}

func shellFooterButtonAt(x, y, width, height int) (shellButton, bool) {
	if y < height-2 || y >= height || x < 0 || x >= width {
		return shellButton{}, false
	}

	buttons := shellFooterButtons()
	offset := 0

	for index, cellWidth := range distributedWidths(width, len(buttons)) {
		if x >= offset && x < offset+cellWidth {
			return buttons[index], true
		}

		offset += cellWidth
	}

	return shellButton{}, false
}

func centerShellButtonLines(key, label string, width int) string {
	if width <= 0 {
		return ""
	}

	keyRunes := []rune(key)
	labelRunes := []rune(label)
	keyWidth := minInt(len(keyRunes), width)
	labelWidth := minInt(len(labelRunes), width)

	anchorWidth := maxInt(keyWidth, labelWidth)
	anchorLeft := maxInt((width-anchorWidth)/2, 0)
	keyLeft := anchorLeft + maxInt((anchorWidth-keyWidth)/2, 0)
	labelLeft := anchorLeft + maxInt((anchorWidth-labelWidth)/2, 0)

	return placePlainLine(key, width, keyLeft) + "\n" +
		placePlainLine(label, width, labelLeft)
}
