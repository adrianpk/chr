package tui

import (
	"os"
	"strings"

	"forge.adrianpk.com/chr/internal/core"
)

type ticketMarkdownDetailView struct {
	ChecklistDone  int
	ChecklistTotal int
	CommentCount   int
	BodyLines      []string
	Evidence       core.EvidenceStrategySummary
}

type ticketMarkdownParser struct {
	detail            ticketMarkdownDetailView
	inFrontMatter     bool
	frontMatterClosed bool
	inComments        bool
}

func ticketMarkdownDetail(path string) ticketMarkdownDetailView {
	content, err := os.ReadFile(path)
	if err != nil {
		return ticketMarkdownDetailView{}
	}

	parser := ticketMarkdownParser{
		detail: ticketMarkdownDetailView{
			Evidence: core.SummarizeEvidenceStrategyChecklist(string(content)),
		},
	}

	for _, line := range strings.Split(string(content), "\n") {
		parser.readLine(line)
	}

	return parser.detail
}

func (p *ticketMarkdownParser) readLine(line string) {
	trimmed := strings.TrimSpace(line)
	if p.handleFrontMatter(trimmed) || p.inFrontMatter {
		return
	}

	p.detail.BodyLines = append(p.detail.BodyLines, line)

	if trimmed == "" {
		return
	}

	switch {
	case strings.HasPrefix(trimmed, "# "):
		p.readHeading(trimmed)
	case strings.HasPrefix(trimmed, "- ["):
		p.readChecklistItem(trimmed)
	case p.inComments && strings.HasPrefix(trimmed, "## "):
		p.detail.CommentCount++
	}
}

func (p *ticketMarkdownParser) handleFrontMatter(trimmed string) bool {
	if trimmed != "---" || p.frontMatterClosed {
		return false
	}

	p.inFrontMatter = !p.inFrontMatter
	if !p.inFrontMatter {
		p.frontMatterClosed = true
	}

	return true
}

func (p *ticketMarkdownParser) readHeading(trimmed string) {
	p.inComments = trimmed == "# Comments"
}

func (p *ticketMarkdownParser) readChecklistItem(trimmed string) {
	p.detail.ChecklistTotal++

	if strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]") {
		p.detail.ChecklistDone++
	}
}
