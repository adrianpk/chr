package app

import (
	"slices"
	"strings"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

func ticketEditorGroups(view ticketEditor) []ticketEditorGroup {
	groups := []ticketEditorGroup{
		{Title: "Workflow", Fields: []ticketEditorField{{Name: "status", Label: "Status"}}},
		{Title: "Classification", Fields: []ticketEditorField{
			{Name: "use_case", Label: "Use Case"},
			{Name: "priority", Label: "Priority"},
			{Name: "type", Label: "Kind / type"},
			{Name: "severity", Label: "Severity"},
			{Name: "project_area", Label: "Scope / project area"},
			{Name: "source_kind", Label: "Source"},
			{Name: "resolution", Label: "Resolution"},
		}},
		{Title: "Ownership", Fields: []ticketEditorField{{Name: "assignee", Label: "Assignee"}, {Name: "owner", Label: "Owner"}}},
		{Title: "Delivery", Fields: []ticketEditorField{{Name: "branch", Label: "Branch"}, {Name: "pr", Label: "Pull request"}, {Name: "commits", Label: "Commit references", Multiline: true}}},
	}
	for i := range groups {
		for j := range groups[i].Fields {
			groups[i].Fields[j] = ticketEditorFieldValue(view, groups[i].Fields[j])
		}
	}

	return groups
}

func ticketEditorFieldValue(view ticketEditor, field ticketEditorField) ticketEditorField {
	field.Value, field.ReadOnly = view.Fields[field.Name], view.ReadOnly[field.Name]

	field.Options = ticketEditorFieldOptions(view, field.Name)
	if field.Name == "resolution" && !slices.Contains([]string{"solved", "closed", string(ticketcore.StatusDone)}, view.Status) {
		field.ReadOnly = "Available when the ticket is resolved."
	}

	if field.Name == "use_case" {
		field.Options = []ticketEditorOption{{Value: "", Label: "Project-wide"}}
		for _, item := range view.Cases {
			field.Options = append(field.Options, ticketEditorOption{Value: item.Slug, Label: item.Name})
		}
	}

	if len(field.Options) > 0 && !slices.ContainsFunc(field.Options, func(option ticketEditorOption) bool { return option.Value == field.Value }) {
		field.Options = append(field.Options, ticketEditorOption{Value: field.Value, Label: field.Value})
	}

	return field
}

func ticketEditorFieldOptions(view ticketEditor, name string) []ticketEditorOption {
	if name == "status" {
		return ticketEditorStatusOptions(view)
	}

	values := map[string][]string{
		"priority":   {"", "unclassified", "low", "normal", "high", "critical"},
		"severity":   {"", "unclassified", "low", "normal", "high", "critical"},
		"type":       {"", "task", "bug", "feature", "cleanup", "research", "docs"},
		"resolution": {"", "fixed", "wont_do", "duplicate", "obsolete", "moved", "answered"},
	}
	if strings.HasPrefix(view.Key, "TKT-") {
		repositoryTicketFieldOptions(values)
	}

	options := make([]ticketEditorOption, 0, len(values[name]))
	for _, value := range values[name] {
		label := strings.ReplaceAll(value, "_", " ")
		if label == "" {
			label = "Unspecified"
		}

		options = append(options, ticketEditorOption{Value: value, Label: label})
	}

	return options
}

func repositoryTicketFieldOptions(values map[string][]string) {
	values["priority"] = []string{"unclassified", "low", "normal", "high", "urgent"}
	values["severity"] = []string{"unclassified", "low", "medium", "high", "critical"}
	values["type"] = []string{"unclassified", "bug", "task", "cleanup", "follow_up", "research", "decision"}
	values["project_area"] = []string{"unclassified", "domain", "api", "persistence", "ui", "ops", "docs", "infra"}
	values["source_kind"] = []string{"chat", "manual_test", "review", "trace", "user_report", "implementation"}
}

func ticketEditorStatusOptions(view ticketEditor) []ticketEditorOption {
	options := []ticketEditorOption{
		{Value: string(ticketcore.StatusBacklog), Label: "Backlog"}, {Value: string(ticketcore.StatusReadyForDev), Label: "Ready For Dev"},
	}
	if view.New || view.NativeWorkflow && view.Phase == "pending" {
		return options
	}

	return append(options,
		ticketEditorOption{Value: string(ticketcore.StatusInProgress), Label: "In Progress"},
		ticketEditorOption{Value: string(ticketcore.StatusReadyForReview), Label: "Ready For Review"},
		ticketEditorOption{Value: string(ticketcore.StatusReadyForMerge), Label: "Ready For Merge"},
		ticketEditorOption{Value: string(ticketcore.StatusDone), Label: "Done"},
	)
}
