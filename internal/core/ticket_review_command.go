package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	opsmetadata "forge.adrianpk.com/chr/internal/core/ops/metadata"
	ticketcore "forge.adrianpk.com/chr/internal/core/ticket"
	ticketmetadata "forge.adrianpk.com/chr/internal/core/ticket/metadata"
)

const (
	// TicketReviewCommandResultPass defines a package boundary constant.
	TicketReviewCommandResultPass = "pass"
	// TicketReviewCommandResultNeedsChanges defines a package boundary constant.
	TicketReviewCommandResultNeedsChanges = "needs_changes"
	// TicketReviewCommandResultInsufficientEvidence defines a package boundary constant.
	TicketReviewCommandResultInsufficientEvidence = "insufficient_evidence"
	// TicketReviewCommandResultNeedsHumanDecision defines a package boundary constant.
	TicketReviewCommandResultNeedsHumanDecision = "needs_human_decision"
	// TicketReviewCommandActorReviewer defines a package boundary constant.
	TicketReviewCommandActorReviewer = "reviewer"
	// TicketReviewCommandCleanPassRemoteComment defines a package boundary constant.
	TicketReviewCommandCleanPassRemoteComment = false
	// TicketReviewCommandInspectionOnlyPolicySummary defines a package boundary constant.
	TicketReviewCommandInspectionOnlyPolicySummary = "inspect ticket, execution, validation, pull request, diff, comments, and checks without rerunning implementation"
)

// TicketReviewCommandOptions describes a package boundary value.
type TicketReviewCommandOptions struct {
	Ticket     ticketcore.Indexed
	Config     WorkspaceConfig
	Result     string
	Findings   []string
	Adapter    RemoteReviewAdapter
	DryRun     bool
	ReviewedAt time.Time
}

// TicketReviewCommandResult describes a package boundary value.
type TicketReviewCommandResult struct {
	TicketKey        string
	ReviewRef        ticketmetadata.ReviewRef
	RemoteSnapshot   RemoteReviewSnapshot
	ReviewState      string
	ReviewResult     string
	ReviewRound      int
	Status           string
	BlockingFindings []string
	RemoteComments   []RemoteReviewWrittenComment
	DryRun           bool
}

// RunTicketReviewCommand executes the corresponding package operation.
func RunTicketReviewCommand(ctx context.Context, opts TicketReviewCommandOptions) (TicketReviewCommandResult, error) {
	if err := ctx.Err(); err != nil {
		return TicketReviewCommandResult{}, err
	}

	reviewResult := strings.TrimSpace(opts.Result)
	if reviewResult == "" {
		reviewResult = TicketReviewCommandResultPass
	}

	if err := validateTicketReviewCommandInput(opts.Ticket, reviewResult, opts.Findings); err != nil {
		return TicketReviewCommandResult{}, err
	}

	ref, err := ticketReviewCommandPullRequestRef(opts.Ticket)
	if err != nil {
		return TicketReviewCommandResult{}, err
	}

	snapshot, err := opts.Adapter.Read(ctx, ref.URL)
	if err != nil {
		return TicketReviewCommandResult{}, fmt.Errorf("read remote review evidence: %w", err)
	}

	plan := ticketReviewCommandPlan(opts, ref, snapshot, reviewResult)
	if opts.DryRun {
		return plan, nil
	}

	comments, err := writeTicketReviewCommandFindings(ctx, opts.Adapter, ref.URL, opts.Findings)
	if err != nil {
		return TicketReviewCommandResult{}, err
	}

	plan.RemoteComments = comments

	err = applyTicketReviewCommandResult(ctx, opts.Ticket.Path, plan, opts.ReviewedAt)
	if err != nil {
		return TicketReviewCommandResult{}, err
	}

	return plan, nil
}

func validateTicketReviewCommandInput(ticket ticketcore.Indexed, result string, findings []string) error {
	if strings.TrimSpace(ticket.Key) == "" {
		return fmt.Errorf("ticket key is required")
	}

	if strings.TrimSpace(ticket.Path) == "" {
		return fmt.Errorf("ticket path is required")
	}

	if ticket.Status != ticketcore.StatusReadyForReview {
		return fmt.Errorf("ticket review requires status %s: %s", ticketcore.StatusReadyForReview, ticket.Status)
	}

	if !isValidTicketReviewCommandResult(result) {
		return fmt.Errorf("invalid ticket review result: %s", result)
	}

	if result == TicketReviewCommandResultNeedsChanges && len(trimmedTicketReviewFindings(findings)) == 0 {
		return fmt.Errorf("ticket review needs_changes requires at least one finding")
	}

	return nil
}

func isValidTicketReviewCommandResult(result string) bool {
	switch result {
	case TicketReviewCommandResultPass,
		TicketReviewCommandResultNeedsChanges,
		TicketReviewCommandResultInsufficientEvidence,
		TicketReviewCommandResultNeedsHumanDecision:
		return true
	default:
		return false
	}
}

func ticketReviewCommandPullRequestRef(ticket ticketcore.Indexed) (ticketmetadata.ReviewRef, error) {
	for _, ref := range ticket.ReviewRefs {
		if ref.Type == "pull_request" && strings.TrimSpace(ref.URL) != "" {
			return ref, nil
		}
	}

	return ticketmetadata.ReviewRef{}, fmt.Errorf("ticket review requires a pull_request review ref")
}

func ticketReviewCommandPlan(opts TicketReviewCommandOptions, ref ticketmetadata.ReviewRef, snapshot RemoteReviewSnapshot, result string) TicketReviewCommandResult {
	transition := ticketReviewCommandStateAndStatus(result)

	return TicketReviewCommandResult{
		TicketKey:        opts.Ticket.Key,
		ReviewRef:        ref,
		RemoteSnapshot:   snapshot,
		ReviewState:      transition.state,
		ReviewResult:     result,
		Status:           transition.status,
		BlockingFindings: trimmedTicketReviewFindings(opts.Findings),
		ReviewRound:      opts.Ticket.ReviewRound + 1,
		DryRun:           opts.DryRun,
	}
}

type ticketReviewTransition struct {
	state, status string
}

func ticketReviewCommandStateAndStatus(result string) ticketReviewTransition {
	switch result {
	case TicketReviewCommandResultPass:
		return ticketReviewTransition{state: "pass", status: ticketcore.StatusReadyForMerge}
	case TicketReviewCommandResultNeedsChanges:
		return ticketReviewTransition{state: "changes_requested", status: ticketcore.StatusReadyForDev}
	case TicketReviewCommandResultInsufficientEvidence:
		return ticketReviewTransition{state: "insufficient_evidence", status: ticketcore.StatusReadyForDev}
	case TicketReviewCommandResultNeedsHumanDecision:
		return ticketReviewTransition{state: "needs_human_decision", status: ticketcore.StatusReadyForReview}
	default:
		return ticketReviewTransition{status: ticketcore.StatusReadyForReview}
	}
}

func writeTicketReviewCommandFindings(ctx context.Context, adapter RemoteReviewAdapter, url string, findings []string) ([]RemoteReviewWrittenComment, error) {
	trimmed := trimmedTicketReviewFindings(findings)
	if len(trimmed) == 0 {
		return nil, nil
	}

	comments := make([]RemoteReviewComment, 0, len(trimmed))
	for _, finding := range trimmed {
		comments = append(comments, RemoteReviewComment{Body: finding, Blocking: true})
	}

	result, err := adapter.WriteComments(ctx, url, comments)
	if err != nil {
		return nil, fmt.Errorf("write remote review findings: %w", err)
	}

	return result.Comments, nil
}

func applyTicketReviewCommandResult(ctx context.Context, path string, result TicketReviewCommandResult, reviewedAt time.Time) error {
	content, err := readLegacyTicketFile(path)
	if err != nil {
		return fmt.Errorf("read ticket file: %w", err)
	}

	updatedAt := reviewedAt.UTC()
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	updated := string(content)
	for _, field := range []struct {
		key        string
		value      string
		beforeKeys []string
	}{
		{key: "status", value: opsmetadata.Quote(result.Status), beforeKeys: []string{"priority", "severity", "type"}},
		{key: "review_state", value: opsmetadata.Quote(result.ReviewState), beforeKeys: []string{"review_result", "review_comments_url", "review_round", "blocking_findings", "created_at"}},
		{key: "review_result", value: opsmetadata.Quote(result.ReviewResult), beforeKeys: []string{"review_comments_url", "review_round", "blocking_findings", "created_at"}},
		{key: "review_comments_url", value: opsmetadata.Quote(result.ReviewRef.URL), beforeKeys: []string{"review_round", "blocking_findings", "created_at"}},
		{key: "review_round", value: strconvTicketReviewRound(result), beforeKeys: []string{"blocking_findings", "created_at"}},
		{key: "blocking_findings", value: opsmetadata.QuoteList(result.BlockingFindings), beforeKeys: []string{"created_at"}},
		{key: "updated_at", value: opsmetadata.Timestamp(updatedAt), beforeKeys: nil},
	} {
		updated, err = setTicketFrontmatterScalarField(updated, field.key, field.value, field.beforeKeys)
		if err != nil {
			return err
		}
	}

	err = writeLegacyTicketFile(ctx, path, content, []byte(updated))
	if err != nil {
		return fmt.Errorf("write ticket file: %w", err)
	}

	return nil
}

func strconvTicketReviewRound(result TicketReviewCommandResult) string {
	round := result.ReviewRound
	if round <= 0 {
		round = 1
	}

	return fmt.Sprintf("%d", round)
}

func trimmedTicketReviewFindings(findings []string) []string {
	trimmed := make([]string, 0, len(findings))
	for _, finding := range findings {
		finding = strings.TrimSpace(finding)
		if finding != "" {
			trimmed = append(trimmed, finding)
		}
	}

	return trimmed
}
