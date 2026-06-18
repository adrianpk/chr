package core

import (
	"context"
	"strings"
	"testing"
)

func TestAgentOperationLockBlocksConcurrentOperation(t *testing.T) {
	projectDir := t.TempDir()

	lock, err := AcquireAgentOperationLock(context.Background(), AgentOperationLockOptions{
		ProjectDir: projectDir,
		Operation:  AgentOperationDevelopment,
		Owner:      "test",
	})
	if err != nil {
		t.Fatalf("AcquireAgentOperationLock returned error: %v", err)
	}

	defer func() {
		if err := lock.Release(); err != nil {
			t.Fatalf("Release returned error: %v", err)
		}
	}()

	_, err = AcquireAgentOperationLock(context.Background(), AgentOperationLockOptions{
		ProjectDir: projectDir,
		Operation:  AgentOperationReview,
		Owner:      "test",
	})
	if err == nil || !strings.Contains(err.Error(), "agent operation lock is already held") {
		t.Fatalf("second AcquireAgentOperationLock error = %v, want held lock", err)
	}
}

func TestWorkSelectionPreflightPrioritizesReviewRepair(t *testing.T) {
	tickets := []IndexedTicket{
		{Key: "CHC-2", Status: TicketStatusReadyForDev, ReviewState: "", ReviewResult: ""},
		{Key: "CHC-1", Status: TicketStatusReadyForDev, ReviewState: "changes_requested", ReviewResult: "needs_changes"},
	}

	result, err := RunWorkSelectionPreflight(context.Background(), WorkSelectionPreflightOptions{
		ProjectDir: t.TempDir(),
		Tickets:    tickets,
	})
	if err != nil {
		t.Fatalf("RunWorkSelectionPreflight returned error: %v", err)
	}

	if !result.Allowed {
		t.Fatalf("Allowed = false, want true")
	}

	if len(result.SelectedTickets) != 1 || result.SelectedTickets[0].Key != "CHC-1" {
		t.Fatalf("SelectedTickets = %#v, want CHC-1 repair", result.SelectedTickets)
	}
}

func TestWorkSelectionPreflightBlocksLaterTicketWhenRepairPending(t *testing.T) {
	tickets := []IndexedTicket{
		{Key: "CHC-1", Status: TicketStatusReadyForDev, ReviewState: "changes_requested"},
		{Key: "CHC-2", Status: TicketStatusReadyForDev},
	}

	result, err := RunWorkSelectionPreflight(context.Background(), WorkSelectionPreflightOptions{
		ProjectDir:        t.TempDir(),
		Tickets:           tickets,
		SelectedTicketKey: "CHC-2",
	})
	if err != nil {
		t.Fatalf("RunWorkSelectionPreflight returned error: %v", err)
	}

	if result.Allowed {
		t.Fatalf("Allowed = true, want false")
	}

	if len(result.BlockReasons) != 1 || result.BlockReasons[0].Code != WorkSelectionBlockReviewRepairPending {
		t.Fatalf("BlockReasons = %#v, want review repair pending", result.BlockReasons)
	}
}

func TestWorkSelectionPreflightBlocksStaleDev(t *testing.T) {
	result, err := RunWorkSelectionPreflight(context.Background(), WorkSelectionPreflightOptions{
		ProjectDir:        t.TempDir(),
		Tickets:           []IndexedTicket{{Key: "CHC-1", Status: TicketStatusReadyForDev}},
		RequireUpdatedDev: true,
		Git: fakeWorkSelectionGit{
			branch:      "feat/slice-7-serialized-work-preflight",
			clean:       true,
			containsRef: true,
		},
	})
	if err != nil {
		t.Fatalf("RunWorkSelectionPreflight returned error: %v", err)
	}

	if result.Allowed {
		t.Fatalf("Allowed = true, want false")
	}

	if len(result.BlockReasons) != 1 || result.BlockReasons[0].Code != WorkSelectionBlockStaleDev {
		t.Fatalf("BlockReasons = %#v, want stale dev", result.BlockReasons)
	}
}

func TestWorkSelectionPreflightBlocksDevBehindOriginDev(t *testing.T) {
	result, err := RunWorkSelectionPreflight(context.Background(), WorkSelectionPreflightOptions{
		ProjectDir:        t.TempDir(),
		Tickets:           []IndexedTicket{{Key: "CHC-1", Status: TicketStatusReadyForDev}},
		RequireUpdatedDev: true,
		Git: fakeWorkSelectionGit{
			branch:      "dev",
			clean:       true,
			containsRef: false,
		},
	})
	if err != nil {
		t.Fatalf("RunWorkSelectionPreflight returned error: %v", err)
	}

	if result.Allowed {
		t.Fatalf("Allowed = true, want false")
	}

	if len(result.BlockReasons) != 1 || result.BlockReasons[0].Code != WorkSelectionBlockStaleDev {
		t.Fatalf("BlockReasons = %#v, want stale dev", result.BlockReasons)
	}
}

type fakeWorkSelectionGit struct {
	branch      string
	clean       bool
	containsRef bool
}

func (git fakeWorkSelectionGit) CurrentBranch(context.Context, string) (string, error) {
	return git.branch, nil
}

func (git fakeWorkSelectionGit) WorktreeClean(context.Context, string) (bool, error) {
	return git.clean, nil
}

func (git fakeWorkSelectionGit) BranchContainsRef(context.Context, string, string, string) (bool, error) {
	return git.containsRef, nil
}
