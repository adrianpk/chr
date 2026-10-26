package core

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRunTicketMergePreflightAllowsReadyTicket(t *testing.T) {
	ticket := testMergePreflightTicket()
	remote := fakeTicketMergePreflightRemote{
		snapshot: RemoteReviewSnapshot{
			Reference:       ReviewProviderReference{Provider: ReviewProviderForgejo},
			PullRequestJSON: `{"state":"open","base":{"ref":"develop"},"head":{"ref":"feat/slice-2-merger-preflight"},"mergeable":true}`,
		},
	}

	result, err := RunTicketMergePreflight(t.Context(), TicketMergePreflightOptions{
		Ticket:        ticket,
		Tickets:       []IndexedTicket{ticket},
		Remote:        remote,
		ExpectedBase:  "develop",
		MergeStrategy: TicketMergeStrategyRebase,
	})
	if err != nil {
		t.Fatalf("RunTicketMergePreflight returned error: %v", err)
	}

	if !result.Allowed {
		t.Fatalf("Allowed = false, blockers = %#v", result.Blockers)
	}

	if result.ReviewRef.URL != "https://forge.example.com/owner/repo/pulls/284" {
		t.Fatalf("ReviewRef = %#v, want PR ref", result.ReviewRef)
	}
}

func TestRunTicketMergePreflightBlocksMissingHandoff(t *testing.T) {
	ticket := testMergePreflightTicket()
	ticket.ReviewResult = TicketReviewCommandResultNeedsChanges

	result, err := RunTicketMergePreflight(t.Context(), TicketMergePreflightOptions{
		Ticket: ticket,
		Remote: fakeTicketMergePreflightRemote{},
	})
	if err != nil {
		t.Fatalf("RunTicketMergePreflight returned error: %v", err)
	}

	assertMergePreflightBlocker(t, result, TicketMergePreflightBlockMissingHandoff, TicketMergePreflightBlockKindOperational)
}

func TestRunTicketMergePreflightBlocksProviderUnreadable(t *testing.T) {
	result, err := RunTicketMergePreflight(t.Context(), TicketMergePreflightOptions{
		Ticket: testMergePreflightTicket(),
		Remote: fakeTicketMergePreflightRemote{err: errors.New("missing token")},
	})
	if err != nil {
		t.Fatalf("RunTicketMergePreflight returned error: %v", err)
	}

	assertMergePreflightBlocker(t, result, TicketMergePreflightBlockProviderUnreadable, TicketMergePreflightBlockKindOperational)
}

func TestRunTicketMergePreflightBlocksWrongBase(t *testing.T) {
	result, err := RunTicketMergePreflight(t.Context(), TicketMergePreflightOptions{
		Ticket: testMergePreflightTicket(),
		Remote: fakeTicketMergePreflightRemote{
			snapshot: RemoteReviewSnapshot{
				Reference:       ReviewProviderReference{Provider: ReviewProviderForgejo},
				PullRequestJSON: `{"state":"open","base":{"ref":"main"},"head":{"ref":"feat/slice-2-merger-preflight"},"mergeable":true}`,
			},
		},
	})
	if err != nil {
		t.Fatalf("RunTicketMergePreflight returned error: %v", err)
	}

	assertMergePreflightBlocker(t, result, TicketMergePreflightBlockWrongBase, TicketMergePreflightBlockKindOperational)
}

func TestRunTicketMergePreflightBlocksHeadMismatch(t *testing.T) {
	result, err := RunTicketMergePreflight(t.Context(), TicketMergePreflightOptions{
		Ticket: testMergePreflightTicket(),
		Remote: fakeTicketMergePreflightRemote{
			snapshot: RemoteReviewSnapshot{
				Reference:       ReviewProviderReference{Provider: ReviewProviderGitHub},
				PullRequestJSON: `{"state":"OPEN","baseRefName":"dev","headRefName":"feat/other","mergeable":"MERGEABLE"}`,
			},
		},
	})
	if err != nil {
		t.Fatalf("RunTicketMergePreflight returned error: %v", err)
	}

	assertMergePreflightBlocker(t, result, TicketMergePreflightBlockHeadMismatch, TicketMergePreflightBlockKindRepair)
}

func TestRunTicketMergePreflightBlocksNotMergeable(t *testing.T) {
	result, err := RunTicketMergePreflight(t.Context(), TicketMergePreflightOptions{
		Ticket: testMergePreflightTicket(),
		Remote: fakeTicketMergePreflightRemote{
			snapshot: RemoteReviewSnapshot{
				Reference:       ReviewProviderReference{Provider: ReviewProviderGitHub},
				PullRequestJSON: `{"state":"OPEN","baseRefName":"dev","headRefName":"feat/slice-2-merger-preflight","mergeable":"CONFLICTING"}`,
			},
		},
	})
	if err != nil {
		t.Fatalf("RunTicketMergePreflight returned error: %v", err)
	}

	assertMergePreflightBlocker(t, result, TicketMergePreflightBlockNotMergeable, TicketMergePreflightBlockKindRepair)
}

func TestRunTicketMergePreflightBlocksSerialRepair(t *testing.T) {
	ticket := testMergePreflightTicket()

	result, err := RunTicketMergePreflight(t.Context(), TicketMergePreflightOptions{
		Ticket: ticket,
		Tickets: []IndexedTicket{
			ticket,
			{IndexedTicketIdentity: IndexedTicketIdentity{Key: "CHR-1"}, IndexedTicketClassification: IndexedTicketClassification{Status: TicketStatusReadyForDev}, IndexedTicketReview: IndexedTicketReview{ReviewState: "changes_requested"}},
		},
		Remote: fakeTicketMergePreflightRemote{},
	})
	if err != nil {
		t.Fatalf("RunTicketMergePreflight returned error: %v", err)
	}

	assertMergePreflightBlocker(t, result, TicketMergePreflightBlockRepairPending, TicketMergePreflightBlockKindRepair)
}

func TestRunTicketMergePreflightBlocksInvalidMergeStrategy(t *testing.T) {
	result, err := RunTicketMergePreflight(t.Context(), TicketMergePreflightOptions{
		Ticket:        testMergePreflightTicket(),
		Remote:        fakeTicketMergePreflightRemote{},
		MergeStrategy: "octopus",
	})
	if err != nil {
		t.Fatalf("RunTicketMergePreflight returned error: %v", err)
	}

	assertMergePreflightBlocker(t, result, TicketMergePreflightBlockInvalidMergeStrategy, TicketMergePreflightBlockKindOperational)
}

func testMergePreflightTicket() IndexedTicket {
	now := time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)

	return IndexedTicket{
		IndexedTicketIdentity: IndexedTicketIdentity{
			ProjectID:    "project-1",
			Path:         "/tmp/project/.chr/default/tickets/CHR-2-Test.md",
			RelativePath: "CHR-2-Test.md",
			Key:          "CHR-2",
		}, IndexedTicketSummary: IndexedTicketSummary{Title: "Test ticket"}, IndexedTicketClassification: IndexedTicketClassification{Status: TicketStatusReadyForMerge}, IndexedTicketSource: IndexedTicketSource{Branch: "feat/slice-2-merger-preflight"}, IndexedTicketReview: IndexedTicketReview{
			ReviewState:  TicketReviewCommandResultPass,
			ReviewResult: TicketReviewCommandResultPass,
			ReviewRefs: []TicketReviewRef{{
				Type:     "pull_request",
				Provider: "forgejo",
				URL:      "https://forge.example.com/owner/repo/pulls/284",
				ID:       "284",
				Status:   "open",
			}},
		}, IndexedTicketTimes: IndexedTicketTimes{
			CreatedAt: &now,
			UpdatedAt: &now,
		},
	}
}

func assertMergePreflightBlocker(t *testing.T, result TicketMergePreflightResult, code, kind string) {
	t.Helper()

	if result.Allowed {
		t.Fatalf("Allowed = true, want false")
	}

	for _, blocker := range result.Blockers {
		if blocker.Code == code && blocker.Kind == kind {
			return
		}
	}

	t.Fatalf("Blockers = %#v, want %s/%s", result.Blockers, code, kind)
}

type fakeTicketMergePreflightRemote struct {
	snapshot RemoteReviewSnapshot
	err      error
}

func (remote fakeTicketMergePreflightRemote) Read(context.Context, string) (RemoteReviewSnapshot, error) {
	if remote.err != nil {
		return RemoteReviewSnapshot{}, remote.err
	}

	if strings.TrimSpace(remote.snapshot.PullRequestJSON) == "" {
		return RemoteReviewSnapshot{}, errors.New("unexpected remote read")
	}

	return remote.snapshot, nil
}
