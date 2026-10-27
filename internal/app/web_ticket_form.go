package app

import (
	"slices"
	"strings"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	"forge.adrianpk.com/chr/internal/webui"
)

func (p webViewProvider) completeTicketEditor(view webui.TicketDetailView) (webui.TicketDetailView, error) {
	view.RequestID = webTicketMutationBinding("").RequestID
	view.CommentRequestID = webTicketMutationBinding("").RequestID

	cases, err := authoringcore.NewCaseMetadataService(p.projectContext.ProjectDir).LoadAll()
	if err != nil {
		return view, err
	}

	for _, item := range cases {
		view.Cases = append(view.Cases, webui.TicketCaseView{Slug: item.Slug, Name: item.Name})
	}

	view.Groups = webTicketFieldGroups(view)

	return view, nil
}

func webTicketFieldGroups(view webui.TicketDetailView) []webui.TicketFieldGroupView {
	groups := []webui.TicketFieldGroupView{
		{Title: "Workflow", Fields: []webui.TicketFieldView{{Name: "status", Label: "Status"}}},
		{Title: "Classification", Fields: []webui.TicketFieldView{
			{Name: "use_case", Label: "Use Case"},
			{Name: "priority", Label: "Priority"},
			{Name: "type", Label: "Kind / type"},
			{Name: "severity", Label: "Severity"},
			{Name: "project_area", Label: "Scope / project area"},
			{Name: "source_kind", Label: "Source"},
			{Name: "resolution", Label: "Resolution"},
		}},
		{Title: "Ownership", Fields: []webui.TicketFieldView{{Name: "assignee", Label: "Assignee"}, {Name: "owner", Label: "Owner"}}},
		{Title: "Delivery", Fields: []webui.TicketFieldView{{Name: "branch", Label: "Branch"}, {Name: "pr", Label: "Pull request"}, {Name: "commits", Label: "Commit references", Multiline: true}}},
	}
	for i := range groups {
		for j := range groups[i].Fields {
			groups[i].Fields[j] = webTicketField(view, groups[i].Fields[j])
		}
	}

	return groups
}

func webTicketField(view webui.TicketDetailView, field webui.TicketFieldView) webui.TicketFieldView {
	field.Value, field.ReadOnly = view.Fields[field.Name], view.ReadOnly[field.Name]

	field.Options = webTicketFieldOptions(view, field.Name)
	if field.Name == "resolution" && !slices.Contains([]string{"solved", "closed", ticketcore.TicketStatusDone}, view.Status) {
		field.ReadOnly = "Available when the ticket is resolved."
	}

	if field.Name == "use_case" {
		field.Options = []webui.TicketOptionView{{Value: "", Label: "Project-wide"}}
		for _, item := range view.Cases {
			field.Options = append(field.Options, webui.TicketOptionView{Value: item.Slug, Label: item.Name})
		}
	}

	if len(field.Options) > 0 && !slices.ContainsFunc(field.Options, func(option webui.TicketOptionView) bool { return option.Value == field.Value }) {
		field.Options = append(field.Options, webui.TicketOptionView{Value: field.Value, Label: field.Value})
	}

	return field
}

func webTicketFieldOptions(view webui.TicketDetailView, name string) []webui.TicketOptionView {
	if name == "status" {
		return webTicketStatusOptions(view)
	}

	values := map[string][]string{
		"priority":   {"", "unclassified", "low", "normal", "high", "critical"},
		"severity":   {"", "unclassified", "low", "normal", "high", "critical"},
		"type":       {"", "task", "bug", "feature", "cleanup", "research", "docs"},
		"resolution": {"", "fixed", "wont_do", "duplicate", "obsolete", "moved", "answered"},
	}
	if strings.HasPrefix(view.Key, "TKT-") {
		webRepositoryFieldOptions(values)
	}

	options := make([]webui.TicketOptionView, 0, len(values[name]))
	for _, value := range values[name] {
		label := strings.ReplaceAll(value, "_", " ")
		if label == "" {
			label = "Unspecified"
		}

		options = append(options, webui.TicketOptionView{Value: value, Label: label})
	}

	return options
}

func webRepositoryFieldOptions(values map[string][]string) {
	values["priority"] = []string{"unclassified", "low", "normal", "high", "urgent"}
	values["severity"] = []string{"unclassified", "low", "medium", "high", "critical"}
	values["type"] = []string{"unclassified", "bug", "task", "cleanup", "follow_up", "research", "decision"}
	values["project_area"] = []string{"unclassified", "domain", "api", "persistence", "ui", "ops", "docs", "infra"}
	values["source_kind"] = []string{"chat", "manual_test", "review", "trace", "user_report", "implementation"}
}

func webTicketStatusOptions(view webui.TicketDetailView) []webui.TicketOptionView {
	options := []webui.TicketOptionView{
		{Value: ticketcore.TicketStatusBacklog, Label: "Backlog"}, {Value: ticketcore.TicketStatusReadyForDev, Label: "Ready For Dev"},
	}
	if view.New || view.NativeWorkflow && view.Phase == "pending" {
		return options
	}

	return append(options,
		webui.TicketOptionView{Value: ticketcore.TicketStatusInProgress, Label: "In Progress"},
		webui.TicketOptionView{Value: ticketcore.TicketStatusReadyForReview, Label: "Ready For Review"},
		webui.TicketOptionView{Value: ticketcore.TicketStatusReadyForMerge, Label: "Ready For Merge"},
		webui.TicketOptionView{Value: ticketcore.TicketStatusDone, Label: "Done"},
	)
}
