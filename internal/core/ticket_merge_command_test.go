package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
)

func TestRunTicketMergeCommandDryRunDoesNotMutateTicket(t *testing.T) {
	ticket, path := writeMergeCommandTicket(t)
	before := readMergeCommandTicket(t, path)

	result, err := RunTicketMergeCommand(t.Context(), TicketMergeCommandOptions{
		Ticket:        ticket,
		Tickets:       []ticketcore.Indexed{ticket},
		Remote:        fakeTicketMergeCommandRemote{snapshot: mergeCommandSnapshot()},
		MergeStrategy: TicketMergeStrategyRebase,
		DryRun:        true,
		MergedAt:      testMergeCommandTime(),
	})
	if err != nil {
		t.Fatalf("RunTicketMergeCommand returned error: %v", err)
	}

	if !result.Preflight.Allowed || result.Status != ticketcore.StatusReadyForMerge {
		t.Fatalf("result = %#v, want allowed dry-run ready_for_merge", result)
	}

	if after := readMergeCommandTicket(t, path); after != before {
		t.Fatalf("ticket mutated on dry-run:\n%s", after)
	}
}

func TestRunTicketMergeCommandAppliesSuccessfulMerge(t *testing.T) {
	ticket, path := writeMergeCommandTicket(t)
	remote := fakeTicketMergeCommandRemote{
		snapshot: mergeCommandSnapshot(),
		merge: MergeResult{
			Reference: ReviewProviderReference{Provider: ReviewProviderForgejo},
			Strategy:  TicketMergeStrategyRebase,
			SHA:       "abc123",
		},
	}

	result, err := RunTicketMergeCommand(t.Context(), TicketMergeCommandOptions{
		Ticket:        ticket,
		Tickets:       []ticketcore.Indexed{ticket},
		Remote:        remote,
		MergeStrategy: TicketMergeStrategyRebase,
		MergedAt:      testMergeCommandTime(),
	})
	if err != nil {
		t.Fatalf("RunTicketMergeCommand returned error: %v", err)
	}

	if result.Status != ticketcore.StatusDone || result.MergeState != TicketMergeStateMerged {
		t.Fatalf("result = %#v, want done merged", result)
	}

	content := readMergeCommandTicket(t, path)
	for _, want := range []string{
		`status: "done"`,
		`merge_state: "merged"`,
		`merge_result: "merged"`,
		`merge_actor: "merger"`,
		`merge_commit: "abc123"`,
		`merged_at: "2026-06-02T12:00:00Z"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("ticket content does not contain %q:\n%s", want, content)
		}
	}
}

func TestRunTicketMergeCommandCleansWorktreeAfterSuccessfulMerge(t *testing.T) {
	ticket, _ := writeMergeCommandTicket(t)
	projectDir := t.TempDir()
	plan := testMergeCommandWorktreePlan(t, projectDir, ticket)
	git := newFakeWorktreeGit()
	git.entries = []WorktreeGitEntry{{Path: plan.WorktreePath, Branch: plan.Branch}}
	writeWorktreeTestMetadata(t, plan)

	remote := fakeTicketMergeCommandRemote{
		snapshot: mergeCommandSnapshot(),
		merge: MergeResult{
			Reference: ReviewProviderReference{Provider: ReviewProviderForgejo},
			Strategy:  TicketMergeStrategyRebase,
			SHA:       "abc123",
		},
	}

	result, err := RunTicketMergeCommand(t.Context(), TicketMergeCommandOptions{
		Ticket:        ticket,
		Tickets:       []ticketcore.Indexed{ticket},
		Remote:        remote,
		MergeStrategy: TicketMergeStrategyRebase,
		ProjectDir:    projectDir,
		BranchPolicy:  DefaultWorkBranchPolicy(),
		Workspace:     testCleanupWorktreeWorkspace(),
		WorktreeGit:   git,
		MergedAt:      testMergeCommandTime(),
	})
	if err != nil {
		t.Fatalf("RunTicketMergeCommand returned error: %v", err)
	}

	if result.WorktreeCleanup.Status != WorktreeCleanupStatusRemoved {
		t.Fatalf("WorktreeCleanup = %#v, want removed", result.WorktreeCleanup)
	}

	if len(git.removedWorktrees) != 1 || git.removedWorktrees[0] != plan.WorktreePath {
		t.Fatalf("removedWorktrees = %#v, want %q", git.removedWorktrees, plan.WorktreePath)
	}

	assertWorktreeMetadataMissing(t, plan)
}

func TestRunTicketMergeCommandRecordsProviderFailure(t *testing.T) {
	ticket, path := writeMergeCommandTicket(t)

	result, err := RunTicketMergeCommand(t.Context(), TicketMergeCommandOptions{
		Ticket:        ticket,
		Tickets:       []ticketcore.Indexed{ticket},
		Remote:        fakeTicketMergeCommandRemote{snapshot: mergeCommandSnapshot(), mergeErr: errors.New("provider failed")},
		MergeStrategy: TicketMergeStrategyRebase,
		MergedAt:      testMergeCommandTime(),
	})
	if err != nil {
		t.Fatalf("RunTicketMergeCommand returned error: %v", err)
	}

	if result.Preflight.Allowed || result.MergeState != TicketMergeStateBlocked {
		t.Fatalf("result = %#v, want blocked provider failure", result)
	}

	content := readMergeCommandTicket(t, path)
	for _, want := range []string{
		`status: "ready_for_merge"`,
		`merge_state: "blocked"`,
		`merge_result: "blocked"`,
		`merge_blocker: "provider failed"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("ticket content does not contain %q:\n%s", want, content)
		}
	}
}

func TestRunTicketMergeCommandRecordsBlockedPreflight(t *testing.T) {
	ticket, path := writeMergeCommandTicket(t)

	result, err := RunTicketMergeCommand(t.Context(), TicketMergeCommandOptions{
		Ticket:        ticket,
		Tickets:       []ticketcore.Indexed{ticket},
		Remote:        fakeTicketMergeCommandRemote{snapshot: blockedMergeCommandSnapshot()},
		MergeStrategy: TicketMergeStrategyRebase,
		MergedAt:      testMergeCommandTime(),
	})
	if err != nil {
		t.Fatalf("RunTicketMergeCommand returned error: %v", err)
	}

	if result.Preflight.Allowed || result.MergeState != TicketMergeStateBlocked {
		t.Fatalf("result = %#v, want blocked preflight", result)
	}

	content := readMergeCommandTicket(t, path)
	for _, want := range []string{
		`status: "ready_for_merge"`,
		`merge_state: "blocked"`,
		`merge_result: "blocked"`,
		`merge_blocker: "not_mergeable"`,
		`merged_at: ""`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("ticket content does not contain %q:\n%s", want, content)
		}
	}
}

type fakeTicketMergeCommandRemote struct {
	snapshot RemoteReviewSnapshot
	merge    MergeResult
	mergeErr error
}

func (remote fakeTicketMergeCommandRemote) Read(context.Context, string) (RemoteReviewSnapshot, error) {
	return remote.snapshot, nil
}

func (remote fakeTicketMergeCommandRemote) Merge(context.Context, MergeRequest) (MergeResult, error) {
	if remote.mergeErr != nil {
		return MergeResult{}, remote.mergeErr
	}

	return remote.merge, nil
}

func writeMergeCommandTicket(t *testing.T) (ticketcore.Indexed, string) {
	t.Helper()

	now := testMergeCommandTime()
	ticket := ticketcore.Ticket{
		Key:          "CHR-3",
		Title:        "Merge reviewed ticket",
		Status:       ticketcore.StatusReadyForMerge,
		Branch:       "feat/slice-3-merge-execution-command",
		ReviewState:  TicketReviewCommandResultPass,
		ReviewResult: TicketReviewCommandResultPass,
		ReviewRefs: []ticketmetadata.ReviewRef{{
			Type:     "pull_request",
			Provider: "forgejo",
			URL:      "https://forge.example.com/owner/repo/pulls/285",
			ID:       "285",
			Status:   "open",
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}

	content, err := ticketcore.NewRenderer().Render(ticket)
	if err != nil {
		t.Fatalf("RenderTicket returned error: %v", err)
	}

	path := filepath.Join(t.TempDir(), "ticket.md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write ticket: %v", err)
	}

	indexed := ticketcore.Indexed{
		IndexedIdentity: ticketcore.IndexedIdentity{
			Key:  ticket.Key,
			Path: path,
		}, IndexedClassification: ticketcore.IndexedClassification{Status: ticket.Status}, IndexedSource: ticketcore.IndexedSource{Branch: ticket.Branch}, IndexedReview: ticketcore.IndexedReview{
			ReviewState:  ticket.ReviewState,
			ReviewResult: ticket.ReviewResult,
			ReviewRefs:   ticket.ReviewRefs,
		},
	}

	return indexed, path
}

func readMergeCommandTicket(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ticket: %v", err)
	}

	return string(content)
}

func mergeCommandSnapshot() RemoteReviewSnapshot {
	return RemoteReviewSnapshot{
		Reference:       ReviewProviderReference{Provider: ReviewProviderForgejo},
		PullRequestJSON: `{"state":"open","base":{"ref":"dev"},"head":{"ref":"feat/slice-3-merge-execution-command"},"mergeable":true}`,
	}
}

func blockedMergeCommandSnapshot() RemoteReviewSnapshot {
	return RemoteReviewSnapshot{
		Reference:       ReviewProviderReference{Provider: ReviewProviderForgejo},
		PullRequestJSON: `{"state":"open","base":{"ref":"dev"},"head":{"ref":"feat/slice-3-merge-execution-command"},"mergeable":false}`,
	}
}

func testMergeCommandTime() time.Time {
	return time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)
}

func testMergeCommandWorktreePlan(t *testing.T, projectDir string, ticket ticketcore.Indexed) WorktreePlan {
	t.Helper()

	plan, err := NewWorktreePlan(WorktreePlanOptions{
		ProjectDir:   projectDir,
		TicketKey:    ticket.Key,
		Branch:       ticket.Branch,
		BranchPolicy: DefaultWorkBranchPolicy(),
		Workspace:    testCleanupWorktreeWorkspace(),
	})
	if err != nil {
		t.Fatalf("NewWorktreePlan returned error: %v", err)
	}

	return plan
}
