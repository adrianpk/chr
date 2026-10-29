package tui

import "fmt"

// GuidedResultView is verified native delivery and validation evidence.
type GuidedResultView struct {
	Complete     bool
	Summary      string
	ClosingCount int
	Deliveries   []GuidedDeliveryView
	Validation   []GuidedValidationView
}

type (
	// GuidedDeliveryView identifies a confirmed native delivery receipt.
	GuidedDeliveryView struct{ Repository, Branch, Commit, PullRequest, Method, Reference, Hash string }
	// GuidedValidationView retains task or closing validation and its evidence.
	GuidedValidationView struct {
		Label, Outcome, Reference, Waiver string
		Commands                          []GuidedValidationCommand
	}
)

// GuidedValidationCommand displays a host-observed command and exit code.
type GuidedValidationCommand struct {
	Command  string
	ExitCode int
}

func (m *Model) guidedResultLines() []string {
	result := m.guided.View.Result
	if m.SelectedView() != ViewBoard || result == nil {
		return nil
	}

	lines := []string{"Results", result.Summary}
	for _, delivery := range result.Deliveries {
		lines = append(lines, "Confirmed delivery", "Repository: "+delivery.Repository, "Branch: "+delivery.Branch, "Commit: "+delivery.Commit, "Method: "+delivery.Method, "PR: "+delivery.PullRequest, "Receipt: "+delivery.Reference, "Receipt hash: "+delivery.Hash)
	}

	for _, validation := range result.Validation {
		lines = append(lines, guidedValidationLines(validation)...)
	}

	return append(lines, fmt.Sprintf("Closing receipts: %d", result.ClosingCount), "v JSON export · r read current state")
}

func guidedValidationLines(validation GuidedValidationView) []string {
	lines := []string{validation.Label + ": " + validation.Outcome, "Evidence: " + validation.Reference}
	if validation.Waiver != "" {
		lines = append(lines, "Waiver: "+validation.Waiver)
	}

	for _, command := range validation.Commands {
		lines = append(lines, fmt.Sprintf("Validation: %s · exit %d", command.Command, command.ExitCode))
	}

	return lines
}
