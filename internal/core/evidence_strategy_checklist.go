package core

import (
	"fmt"
	"strings"
)

const (
	// EvidenceStrategyChecklistStartAnchor defines the v1 machine start anchor.
	EvidenceStrategyChecklistStartAnchor = "<!-- chc:evidence-strategy:v1 -->"
	// EvidenceStrategyChecklistEndAnchor defines the v1 machine end anchor.
	EvidenceStrategyChecklistEndAnchor = "<!-- /chc:evidence-strategy -->"
)

const (
	// EvidenceStrategyChecklistLabelStrategy defines the strategy decision item.
	EvidenceStrategyChecklistLabelStrategy = "Evidence strategy"
	// EvidenceStrategyChecklistLabelGherkin defines the Gherkin evidence item.
	EvidenceStrategyChecklistLabelGherkin = "Gherkin"
	// EvidenceStrategyChecklistLabelTests defines the tests evidence item.
	EvidenceStrategyChecklistLabelTests = "Tests"
	// EvidenceStrategyChecklistLabelValidation defines the validation evidence item.
	EvidenceStrategyChecklistLabelValidation = "Validation"
	// EvidenceStrategyChecklistLabelSummary defines the final evidence summary item.
	EvidenceStrategyChecklistLabelSummary = "Evidence summary"
)

// EvidenceStrategyChecklistItem describes one anchored strategy checklist item.
type EvidenceStrategyChecklistItem struct {
	Label   string
	Checked bool
	Status  string
	Detail  string
}

// EvidenceStrategyChecklist describes the anchored v1 strategy checklist block.
type EvidenceStrategyChecklist struct {
	Found bool
	Items map[string]EvidenceStrategyChecklistItem
}

// ParseEvidenceStrategyChecklist parses a ticket evidence strategy checklist.
func ParseEvidenceStrategyChecklist(content string) (EvidenceStrategyChecklist, error) {
	block, found, err := evidenceStrategyChecklistBlock(content)
	if err != nil || !found {
		return EvidenceStrategyChecklist{Found: found}, err
	}

	items, err := parseEvidenceStrategyChecklistItems(block)
	if err != nil {
		return EvidenceStrategyChecklist{}, err
	}

	return EvidenceStrategyChecklist{
		Found: true,
		Items: items,
	}, nil
}

// EvidenceStrategyChecklistLabels returns labels in canonical render order.
func EvidenceStrategyChecklistLabels() []string {
	return []string{
		EvidenceStrategyChecklistLabelStrategy,
		EvidenceStrategyChecklistLabelGherkin,
		EvidenceStrategyChecklistLabelTests,
		EvidenceStrategyChecklistLabelValidation,
		EvidenceStrategyChecklistLabelSummary,
	}
}

// Item returns one checklist item by canonical label.
func (c EvidenceStrategyChecklist) Item(label string) (EvidenceStrategyChecklistItem, bool) {
	item, ok := c.Items[label]

	return item, ok
}

func evidenceStrategyChecklistBlock(content string) (string, bool, error) {
	start := strings.Index(content, EvidenceStrategyChecklistStartAnchor)
	end := strings.Index(content, EvidenceStrategyChecklistEndAnchor)

	switch {
	case start < 0 && end < 0:
		return "", false, nil
	case start < 0:
		return "", false, fmt.Errorf("evidence strategy checklist end anchor without start anchor")
	case end < 0:
		return "", false, fmt.Errorf("evidence strategy checklist start anchor without end anchor")
	case end < start:
		return "", false, fmt.Errorf("evidence strategy checklist end anchor appears before start anchor")
	}

	remaining := content[start+len(EvidenceStrategyChecklistStartAnchor):]
	if strings.Contains(remaining[:end-start-len(EvidenceStrategyChecklistStartAnchor)], EvidenceStrategyChecklistStartAnchor) {
		return "", false, fmt.Errorf("evidence strategy checklist contains duplicate start anchor")
	}

	return content[start+len(EvidenceStrategyChecklistStartAnchor) : end], true, nil
}

func parseEvidenceStrategyChecklistItems(block string) (map[string]EvidenceStrategyChecklistItem, error) {
	items := map[string]EvidenceStrategyChecklistItem{}

	for lineNumber, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		item, ok, err := parseEvidenceStrategyChecklistLine(trimmed)
		if err != nil {
			return nil, fmt.Errorf("parse evidence strategy checklist line %d: %w", lineNumber+1, err)
		}

		if !ok {
			continue
		}

		if _, exists := items[item.Label]; exists {
			return nil, fmt.Errorf("duplicate evidence strategy checklist label: %s", item.Label)
		}

		items[item.Label] = item
	}

	return requireEvidenceStrategyChecklistItems(items)
}

func parseEvidenceStrategyChecklistLine(line string) (EvidenceStrategyChecklistItem, bool, error) {
	if !strings.HasPrefix(line, "- [") {
		return EvidenceStrategyChecklistItem{}, false, nil
	}

	if len(line) < len("- [ ] ") || line[3] != ' ' && line[3] != 'x' && line[3] != 'X' || line[4] != ']' {
		return EvidenceStrategyChecklistItem{}, true, fmt.Errorf("invalid checklist checked state")
	}

	remainder := strings.TrimSpace(line[len("- [ ] "):])
	label, value, found := strings.Cut(remainder, ":")
	if !found {
		return EvidenceStrategyChecklistItem{}, true, fmt.Errorf("checklist item requires label and value")
	}

	label = strings.TrimSpace(label)
	if !isEvidenceStrategyChecklistLabel(label) {
		return EvidenceStrategyChecklistItem{}, true, fmt.Errorf("unknown evidence strategy checklist label: %s", label)
	}

	status, detail := parseEvidenceStrategyChecklistValue(value)
	if !isValidEvidenceStrategyChecklistStatus(label, status) {
		return EvidenceStrategyChecklistItem{}, true, fmt.Errorf("invalid %s status: %s", label, status)
	}

	return EvidenceStrategyChecklistItem{
		Label:   label,
		Checked: line[3] == 'x' || line[3] == 'X',
		Status:  status,
		Detail:  detail,
	}, true, nil
}

func parseEvidenceStrategyChecklistValue(value string) (string, string) {
	value = strings.TrimSpace(value)
	status, detail, found := strings.Cut(value, ";")
	if !found {
		return strings.TrimSpace(value), ""
	}

	return strings.TrimSpace(status), strings.TrimSpace(detail)
}

func requireEvidenceStrategyChecklistItems(
	items map[string]EvidenceStrategyChecklistItem,
) (map[string]EvidenceStrategyChecklistItem, error) {
	for _, label := range EvidenceStrategyChecklistLabels() {
		if _, ok := items[label]; ok {
			continue
		}

		return nil, fmt.Errorf("missing evidence strategy checklist label: %s", label)
	}

	return items, nil
}

func isEvidenceStrategyChecklistLabel(label string) bool {
	for _, candidate := range EvidenceStrategyChecklistLabels() {
		if label == candidate {
			return true
		}
	}

	return false
}

func isValidEvidenceStrategyChecklistStatus(label, status string) bool {
	status = strings.TrimSpace(status)
	if status == "" {
		return false
	}

	allowed, ok := evidenceStrategyChecklistStatuses()[label]
	if !ok {
		return false
	}

	_, ok = allowed[status]

	return ok
}

func evidenceStrategyChecklistStatuses() map[string]map[string]struct{} {
	return map[string]map[string]struct{}{
		EvidenceStrategyChecklistLabelStrategy: {
			"planned":  {},
			"blocked":  {},
			"complete": {},
		},
		EvidenceStrategyChecklistLabelGherkin: {
			"required":       {},
			"not_applicable": {},
			"disabled":       {},
			"blocked":        {},
		},
		EvidenceStrategyChecklistLabelTests: {
			"none":        {},
			"existing":    {},
			"unit":        {},
			"table":       {},
			"integration": {},
			"fuzz":        {},
			"property":    {},
			"mutation":    {},
			"runtime":     {},
			"mixed":       {},
			"blocked":     {},
		},
		EvidenceStrategyChecklistLabelValidation: {
			"pending": {},
			"passed":  {},
			"failed":  {},
			"blocked": {},
		},
		EvidenceStrategyChecklistLabelSummary: {
			"pending":  {},
			"complete": {},
			"blocked":  {},
		},
	}
}
