package core

import (
	"fmt"
	"strings"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
)

const (
	// DefaultTicketBranchPrefix defines the branch namespace for derived ticket branches.
	DefaultTicketBranchPrefix = "chr"
)

// TicketBranchSelection describes the branch selected for one ticket.
type TicketBranchSelection struct {
	TicketKey      string
	Branch         string
	ExistingBranch string
	Derived        bool
}

// SelectTicketBranch resolves the implementation branch for one ticket.
func SelectTicketBranch(ticket ticketcore.Indexed) (TicketBranchSelection, error) {
	key := strings.TrimSpace(ticket.Key)
	if key == "" {
		return TicketBranchSelection{}, fmt.Errorf("ticket key is required")
	}

	existingBranch := strings.TrimSpace(ticket.Branch)
	if existingBranch != "" {
		return TicketBranchSelection{
			TicketKey:      key,
			Branch:         existingBranch,
			ExistingBranch: existingBranch,
		}, nil
	}

	component := executioncore.WorktreePathComponent(key + " " + ticket.Title)
	if component == "" {
		return TicketBranchSelection{}, fmt.Errorf("ticket %s cannot derive branch name", key)
	}

	return TicketBranchSelection{
		TicketKey: key,
		Branch:    DefaultTicketBranchPrefix + "/" + component,
		Derived:   true,
	}, nil
}
