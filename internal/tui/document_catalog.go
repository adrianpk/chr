package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type documentCatalogItem struct {
	Index                                             int
	UseCaseTitle, Title, Status, ScopeStatus, Content string
}

type documentCatalogView struct {
	kind, title, hint, message string
	items                      []documentCatalogItem
	selected                   int
	busy                       bool
}

func (m *Model) documentCatalog(kind View) documentCatalogView {
	if kind == ViewPlans {
		items := make([]documentCatalogItem, len(m.plans))
		for i, plan := range m.plans {
			items[i] = documentCatalogItem{Index: plan.Index, UseCaseTitle: planUseCaseTitle(plan), Title: plan.Title, Status: plan.Status, ScopeStatus: planScopeStatus(plan), Content: plan.Content}
		}

		return documentCatalogView{kind: "plan", title: "Plans", hint: m.planActionHint(), message: m.statusMessage, items: items, selected: m.selectedPlan, busy: m.planBusy}
	}

	items := make([]documentCatalogItem, len(m.specs))
	for i, spec := range m.specs {
		items[i] = documentCatalogItem{Index: spec.Index, UseCaseTitle: specUseCaseTitle(spec), Title: spec.Title, Status: spec.Status, ScopeStatus: specScopeStatus(spec), Content: spec.Content}
	}

	return documentCatalogView{kind: "spec", title: "Specs", hint: m.specActionHint(), message: m.statusMessage, items: items, selected: m.selectedSpec, busy: m.specBusy}
}

func (catalog documentCatalogView) selectedItem() (documentCatalogItem, bool) {
	for _, item := range catalog.items {
		if item.Index == catalog.selected {
			return item, true
		}
	}

	if len(catalog.items) > 0 {
		return catalog.items[0], true
	}

	return documentCatalogItem{}, false
}

func (catalog documentCatalogView) focusList(limit int) []string {
	if len(catalog.items) == 0 {
		return []string{"No use cases found."}
	}

	lines := []string{shellMutedStyle.Render("Use Case")}

	for _, spec := range catalog.items {
		prefix := " "
		if spec.Index == catalog.selected {
			prefix = ">"
		}

		lines = append(lines, prefix+" "+spec.UseCaseTitle)
		if len(lines) >= limit {
			break
		}
	}

	return lines
}

func (catalog documentCatalogView) detailLines(width int) []string {
	spec, ok := catalog.selectedItem()
	if !ok {
		return []string{"No use case selected."}
	}

	lines := []string{
		shellMutedStyle.Render(catalog.hint),
		"",
		spec.Title,
		shellMutedStyle.Render(spec.ScopeStatus),
	}

	if strings.TrimSpace(spec.Content) == "" {
		lines = append(lines, "", "No generated "+catalog.kind+" found for this use case.")
		if catalog.busy {
			lines = append(lines, "", "Working...")
			lines = append(lines, "  "+catalog.message)
		}

		return lines
	}

	if catalog.busy {
		lines = append(lines, "", "Working...")
		lines = append(lines, "  "+catalog.message)
	}

	lines = append(lines, "")
	lines = append(lines, renderContentReadLines(spec.Content, maxInt(width, 1))...)

	return lines
}

func (catalog documentCatalogView) render() string {
	if len(catalog.items) == 0 {
		return catalog.title + "\n  No use cases found.\n\n  " + catalog.hint
	}

	var output strings.Builder

	output.WriteString(catalog.title + "\n")

	for _, spec := range catalog.items {
		prefix := "  "
		if spec.Index == catalog.selected {
			prefix = "> "
		}

		output.WriteString(prefix)
		output.WriteString(spec.UseCaseTitle)
		output.WriteString("  ")
		output.WriteString(shellMutedStyle.Render(spec.Status))
		output.WriteString("\n")
	}

	output.WriteString("\n")
	output.WriteString(strings.Join(catalog.detailLines(72), "\n"))

	return strings.TrimRight(output.String(), "\n")
}

func (m *Model) renderDocumentCatalogWorkspace(width, height int, catalog documentCatalogView) string {
	list := catalog.focusList(height - 2)
	detail := catalog.detailLines(maxInt(width-4, 1))

	switch m.maximizedPane {
	case 0:
		return fitBlock(renderPanel(catalog.title, list, width, height, true), width, height)
	case 1:
		return fitBlock(m.renderDetailPanel(strings.TrimSuffix(catalog.title, "s"), detail, width, height, true), width, height)
	}

	horizontal := m.splitWidths(width)
	leftWidth, rightWidth := horizontal.Left, horizontal.Right
	detail = catalog.detailLines(maxInt(rightWidth-4, 1))

	return fitBlock(lipgloss.JoinHorizontal(
		lipgloss.Top,
		renderPanel(catalog.title, list, leftWidth, height, m.selectedPane == 0),
		panelVerticalGap(height),
		m.renderDetailPanel(strings.TrimSuffix(catalog.title, "s"), detail, rightWidth, height, m.selectedPane == 1),
	), width, height)
}
