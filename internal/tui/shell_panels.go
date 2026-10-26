package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func wrapPlainText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	lines := []string{}
	line := ""

	for _, word := range words {
		if line == "" {
			line = word

			continue
		}

		next := line + " " + word
		if lipgloss.Width(next) <= width {
			line = next

			continue
		}

		lines = append(lines, line)
		line = word
	}

	if line != "" {
		lines = append(lines, line)
	}

	return lines
}

func renderPanel(title string, lines []string, width, height int, focused bool) string {
	style := shellPanelStyle
	if focused {
		style = shellFocusedPanelStyle
	}

	contentHeight := maxInt(height-2, 1)

	content := append([]string{shellTitleStyle.Render(title)}, lines...)

	return style.Width(maxInt(width-2, 1)).Height(contentHeight).Render(fitBlock(strings.Join(content, "\n"), maxInt(width-4, 1), contentHeight))
}

func (m *Model) renderDetailPanel(title string, lines []string, width, height int, focused bool) string {
	lines = scrollLines(lines, m.detailScroll, maxInt(height-3, 1))

	return renderPanel(title, lines, width, height, focused)
}

func panelVerticalGap(height int) string {
	lines := make([]string, maxInt(height, 1))
	for index := range lines {
		lines[index] = strings.Repeat(" ", panelVerticalGapSize)
	}

	return strings.Join(lines, "\n")
}

func normalizeBlock(content string, width, height int) string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	if len(lines) > height {
		lines = lines[:height]
	}

	for len(lines) < height {
		lines = append(lines, "")
	}

	for index, line := range lines {
		lines[index] = fitStyledLine(line, width)
	}

	return strings.Join(lines, "\n")
}

func fitBlock(content string, width, height int) string {
	return normalizeBlock(content, width, height)
}

func fitStyledLine(line string, width int) string {
	visible := lipgloss.Width(line)
	if visible >= width {
		return line
	}

	return line + strings.Repeat(" ", width-visible)
}

func fitPlainLine(line string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(line)
	if len(runes) > width {
		if width == 1 {
			return "…"
		}

		return string(runes[:width-1]) + "…"
	}

	return line + strings.Repeat(" ", width-len(runes))
}

func placePlainLine(line string, width, left int) string {
	if width <= 0 {
		return ""
	}

	fitted := truncatePlainLine(line, width)
	lineWidth := lipgloss.Width(fitted)

	if lineWidth >= width {
		return fitted
	}

	left = minInt(maxInt(left, 0), width-lineWidth)
	right := width - lineWidth - left

	return strings.Repeat(" ", left) + fitted + strings.Repeat(" ", right)
}

func truncatePlainLine(line string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(line)
	if len(runes) <= width {
		return line
	}

	if width == 1 {
		return "…"
	}

	return string(runes[:width-1]) + "…"
}

func (m *Model) splitWidths(width int) horizontalPanelSplit {
	available := maxInt(width-panelVerticalGapSize, 1)
	left := available * m.verticalSplitPercent / 100
	right := available - left

	return horizontalPanelSplit{Left: left, Right: right}
}

func (m *Model) splitHeights(height int) verticalPanelSplit {
	available := maxInt(height, 1)
	top := maxInt(available*m.horizontalSplitPercent/100, 1)
	bottom := maxInt(available-top, 1)

	return verticalPanelSplit{Top: top, Bottom: bottom}
}

func distributedWidths(width, count int) []int {
	if count <= 0 {
		return nil
	}

	base := width / count
	remainder := width % count
	widths := make([]int, count)

	for index := range widths {
		widths[index] = base
		if index < remainder {
			widths[index]++
		}
	}

	return widths
}

func lineCount(value string) int {
	if value == "" {
		return 0
	}

	return strings.Count(value, "\n") + 1
}

func emptyValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}

	return value
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}

	return right
}

func percentFromPosition(position, total int) int {
	if total <= 0 {
		return defaultPanelSplitPercent
	}

	percent := position * 100 / total

	return minInt(maxInt(percent, minPanelSplitPercent), maxPanelSplitPercent)
}

func nearInt(left, right, distance int) bool {
	delta := left - right
	if delta < 0 {
		delta = -delta
	}

	return delta <= distance
}

func minInt(left, right int) int {
	if left < right {
		return left
	}

	return right
}

type horizontalPanelSplit struct {
	Left  int
	Right int
}

type verticalPanelSplit struct {
	Top    int
	Bottom int
}
