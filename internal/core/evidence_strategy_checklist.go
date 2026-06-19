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

	items, err := parseChecklistItems(block)
	if err != nil {
		return EvidenceStrategyChecklist{}, err
	}

	return EvidenceStrategyChecklist{
		Found: true,
		Items: items,
	}, nil
}

// RenderEvidenceStrategyChecklist renders a deterministic v1 strategy block.
func RenderEvidenceStrategyChecklist(checklist EvidenceStrategyChecklist) (string, error) {
	items, err := requireChecklistItems(checklist.Items)
	if err != nil {
		return "", err
	}

	var output strings.Builder
	output.WriteString(EvidenceStrategyChecklistStartAnchor)
	output.WriteString("\n")

	for _, label := range EvidenceStrategyChecklistLabels() {
		output.WriteString(renderEvidenceStrategyChecklistItem(items[label]))
		output.WriteString("\n")
	}

	output.WriteString(EvidenceStrategyChecklistEndAnchor)
	output.WriteString("\n")

	return output.String(), nil
}

// UpsertEvidenceStrategyChecklist inserts or updates the v1 strategy block.
func UpsertEvidenceStrategyChecklist(content string, updates EvidenceStrategyChecklist) (string, error) {
	existing, err := ParseEvidenceStrategyChecklist(content)
	if err != nil {
		return "", err
	}

	checklist, err := checklistForUpsert(existing, updates)
	if err != nil {
		return "", err
	}

	rendered, err := RenderEvidenceStrategyChecklist(checklist)
	if err != nil {
		return "", err
	}

	if existing.Found {
		return replaceEvidenceStrategyChecklistBlock(content, rendered)
	}

	return insertEvidenceStrategyChecklistBlock(content, rendered), nil
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

func parseChecklistItems(block string) (map[string]EvidenceStrategyChecklistItem, error) {
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

	return requireChecklistItems(items)
}

func parseEvidenceStrategyChecklistLine(line string) (EvidenceStrategyChecklistItem, bool, error) {
	if !strings.HasPrefix(line, "- [") {
		return EvidenceStrategyChecklistItem{}, false, nil
	}

	checked, err := parseEvidenceStrategyChecklistChecked(line)
	if err != nil {
		return EvidenceStrategyChecklistItem{}, true, err
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
		Checked: checked,
		Status:  status,
		Detail:  detail,
	}, true, nil
}

func parseEvidenceStrategyChecklistChecked(line string) (bool, error) {
	if len(line) < len("- [ ] ") || line[4] != ']' {
		return false, fmt.Errorf("invalid checklist checked state")
	}

	switch line[3] {
	case ' ':
		return false, nil
	case 'x', 'X':
		return true, nil
	default:
		return false, fmt.Errorf("invalid checklist checked state")
	}
}

func parseEvidenceStrategyChecklistValue(value string) (string, string) {
	value = strings.TrimSpace(value)

	status, detail, found := strings.Cut(value, ";")
	if !found {
		return strings.TrimSpace(value), ""
	}

	return strings.TrimSpace(status), strings.TrimSpace(detail)
}

func checklistForUpsert(existing, updates EvidenceStrategyChecklist) (EvidenceStrategyChecklist, error) {
	if existing.Found {
		return mergeChecklists(existing, updates)
	}

	for label, item := range updates.Items {
		if err := validateEvidenceStrategyChecklistItem(label, item); err != nil {
			return EvidenceStrategyChecklist{}, err
		}
	}

	return updates, nil
}

func mergeChecklists(existing, updates EvidenceStrategyChecklist) (EvidenceStrategyChecklist, error) {
	merged := EvidenceStrategyChecklist{
		Found: true,
		Items: map[string]EvidenceStrategyChecklistItem{},
	}

	for _, label := range EvidenceStrategyChecklistLabels() {
		merged.Items[label] = existing.Items[label]
	}

	for label, item := range updates.Items {
		if err := validateEvidenceStrategyChecklistItem(label, item); err != nil {
			return EvidenceStrategyChecklist{}, err
		}

		merged.Items[label] = item
	}

	return merged, nil
}

func validateEvidenceStrategyChecklistItem(label string, item EvidenceStrategyChecklistItem) error {
	if !isEvidenceStrategyChecklistLabel(label) {
		return fmt.Errorf("unknown evidence strategy checklist label: %s", label)
	}

	if item.Label != label {
		return fmt.Errorf("evidence strategy checklist item label mismatch: %s", label)
	}

	if !isValidEvidenceStrategyChecklistStatus(label, item.Status) {
		return fmt.Errorf("invalid %s status: %s", label, item.Status)
	}

	return nil
}

func renderEvidenceStrategyChecklistItem(item EvidenceStrategyChecklistItem) string {
	checked := " "
	if item.Checked {
		checked = "x"
	}

	value := strings.TrimSpace(item.Status)

	detail := strings.TrimSpace(item.Detail)
	if detail != "" {
		value += "; " + detail
	}

	return fmt.Sprintf("- [%s] %s: %s", checked, item.Label, value)
}

func replaceEvidenceStrategyChecklistBlock(content, rendered string) (string, error) {
	start := strings.Index(content, EvidenceStrategyChecklistStartAnchor)

	end := strings.Index(content, EvidenceStrategyChecklistEndAnchor)
	if start < 0 || end < 0 || end < start {
		return "", fmt.Errorf("evidence strategy checklist anchors are required")
	}

	end += len(EvidenceStrategyChecklistEndAnchor)

	return content[:start] + strings.TrimRight(rendered, "\n") + content[end:], nil
}

func insertEvidenceStrategyChecklistBlock(content, rendered string) string {
	insertAt := evidenceStrategyChecklistInsertOffset(content)
	prefix := strings.TrimRight(content[:insertAt], "\n")
	suffix := strings.TrimLeft(content[insertAt:], "\n")

	if suffix == "" {
		return prefix + "\n\n" + rendered
	}

	return prefix + "\n\n" + rendered + "\n" + suffix
}

func evidenceStrategyChecklistInsertOffset(content string) int {
	if !strings.HasPrefix(content, "---\n") {
		return 0
	}

	remaining := content[len("---\n"):]

	end := strings.Index(remaining, "\n---")
	if end < 0 {
		return 0
	}

	return len("---\n") + end + len("\n---")
}

func requireChecklistItems(items map[string]EvidenceStrategyChecklistItem) (map[string]EvidenceStrategyChecklistItem, error) {
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
