package core

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
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
	Ticket     IndexedTicket
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
	ReviewRef        TicketReviewRef
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

	plan := ticketReviewCommandPlan(opts.Ticket, ref, snapshot, reviewResult, opts.Findings, opts.DryRun)
	if opts.DryRun {
		return plan, nil
	}

	comments, err := writeTicketReviewCommandFindings(ctx, opts.Adapter, ref.URL, opts.Findings)
	if err != nil {
		return TicketReviewCommandResult{}, err
	}

	plan.RemoteComments = comments

	err = applyTicketReviewCommandResult(opts.Ticket.Path, plan, opts.ReviewedAt)
	if err != nil {
		return TicketReviewCommandResult{}, err
	}

	return plan, nil
}

func validateTicketReviewCommandInput(ticket IndexedTicket, result string, findings []string) error {
	if strings.TrimSpace(ticket.Key) == "" {
		return fmt.Errorf("ticket key is required")
	}

	if strings.TrimSpace(ticket.Path) == "" {
		return fmt.Errorf("ticket path is required")
	}

	if ticket.Status != TicketStatusReadyForReview {
		return fmt.Errorf("ticket review requires status %s: %s", TicketStatusReadyForReview, ticket.Status)
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

func ticketReviewCommandPullRequestRef(ticket IndexedTicket) (TicketReviewRef, error) {
	for _, ref := range ticket.ReviewRefs {
		if ref.Type == "pull_request" && strings.TrimSpace(ref.URL) != "" {
			return ref, nil
		}
	}

	return TicketReviewRef{}, fmt.Errorf("ticket review requires a pull_request review ref")
}

func ticketReviewCommandPlan(
	ticket IndexedTicket,
	ref TicketReviewRef,
	snapshot RemoteReviewSnapshot,
	result string,
	findings []string,
	dryRun bool,
) TicketReviewCommandResult {
	state, status := ticketReviewCommandStateAndStatus(result)

	return TicketReviewCommandResult{
		TicketKey:        ticket.Key,
		ReviewRef:        ref,
		RemoteSnapshot:   snapshot,
		ReviewState:      state,
		ReviewResult:     result,
		Status:           status,
		BlockingFindings: trimmedTicketReviewFindings(findings),
		ReviewRound:      ticket.ReviewRound + 1,
		DryRun:           dryRun,
	}
}

func ticketReviewCommandStateAndStatus(result string) (string, string) {
	switch result {
	case TicketReviewCommandResultPass:
		return "pass", TicketStatusReadyForMerge
	case TicketReviewCommandResultNeedsChanges:
		return "changes_requested", TicketStatusReadyForDev
	case TicketReviewCommandResultInsufficientEvidence:
		return "insufficient_evidence", TicketStatusReadyForDev
	case TicketReviewCommandResultNeedsHumanDecision:
		return "needs_human_decision", TicketStatusReadyForReview
	default:
		return "", TicketStatusReadyForReview
	}
}

func writeTicketReviewCommandFindings(
	ctx context.Context,
	adapter RemoteReviewAdapter,
	url string,
	findings []string,
) ([]RemoteReviewWrittenComment, error) {
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

func applyTicketReviewCommandResult(path string, result TicketReviewCommandResult, reviewedAt time.Time) error {
	content, err := os.ReadFile(path)
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
		{key: "status", value: yamlString(result.Status), beforeKeys: []string{"priority", "severity", "type"}},
		{key: "review_state", value: yamlString(result.ReviewState), beforeKeys: []string{"review_result", "review_comments_url", "review_round", "blocking_findings", "created_at"}},
		{key: "review_result", value: yamlString(result.ReviewResult), beforeKeys: []string{"review_comments_url", "review_round", "blocking_findings", "created_at"}},
		{key: "review_comments_url", value: yamlString(result.ReviewRef.URL), beforeKeys: []string{"review_round", "blocking_findings", "created_at"}},
		{key: "review_round", value: strconvTicketReviewRound(result), beforeKeys: []string{"blocking_findings", "created_at"}},
		{key: "blocking_findings", value: yamlStringList(result.BlockingFindings), beforeKeys: []string{"created_at"}},
		{key: "updated_at", value: yamlTimestamp(updatedAt), beforeKeys: nil},
	} {
		updated, err = setTicketFrontmatterScalarField(updated, field.key, field.value, field.beforeKeys)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile(path, []byte(updated), 0o644)
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
