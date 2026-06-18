package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// TicketMergePreflightBlockKindOperational defines a package boundary constant.
	TicketMergePreflightBlockKindOperational = "operational"
	// TicketMergePreflightBlockKindRepair defines a package boundary constant.
	TicketMergePreflightBlockKindRepair = "repair"

	// TicketMergePreflightBlockMissingHandoff defines a package boundary constant.
	TicketMergePreflightBlockMissingHandoff = "missing_handoff"
	// TicketMergePreflightBlockProviderUnreadable defines a package boundary constant.
	TicketMergePreflightBlockProviderUnreadable = "provider_unreadable"
	// TicketMergePreflightBlockPullRequestClosed defines a package boundary constant.
	TicketMergePreflightBlockPullRequestClosed = "pull_request_closed"
	// TicketMergePreflightBlockWrongBase defines a package boundary constant.
	TicketMergePreflightBlockWrongBase = "wrong_base"
	// TicketMergePreflightBlockHeadMismatch defines a package boundary constant.
	TicketMergePreflightBlockHeadMismatch = "head_mismatch"
	// TicketMergePreflightBlockNotMergeable defines a package boundary constant.
	TicketMergePreflightBlockNotMergeable = "not_mergeable"
	// TicketMergePreflightBlockRepairPending defines a package boundary constant.
	TicketMergePreflightBlockRepairPending = "serial_repair_pending"
	// TicketMergePreflightBlockInvalidMergeStrategy defines a package boundary constant.
	TicketMergePreflightBlockInvalidMergeStrategy = "invalid_merge_strategy"

	// TicketMergeStrategyMerge defines a package boundary constant.
	TicketMergeStrategyMerge = "merge"
	// TicketMergeStrategyRebase defines a package boundary constant.
	TicketMergeStrategyRebase = "rebase"
	// TicketMergeStrategySquash defines a package boundary constant.
	TicketMergeStrategySquash = "squash"
)

// TicketMergePreflightRemoteReader defines a package boundary interface.
type TicketMergePreflightRemoteReader interface {
	Read(ctx context.Context, rawURL string) (RemoteReviewSnapshot, error)
}

// TicketMergePreflightOptions describes a package boundary value.
type TicketMergePreflightOptions struct {
	Ticket        IndexedTicket
	Tickets       []IndexedTicket
	Remote        TicketMergePreflightRemoteReader
	ExpectedBase  string
	MergeStrategy string
}

// TicketMergePreflightResult describes a package boundary value.
type TicketMergePreflightResult struct {
	Allowed       bool
	TicketKey     string
	ReviewRef     TicketReviewRef
	Remote        RemoteReviewSnapshot
	ExpectedBase  string
	MergeStrategy string
	Blockers      []TicketMergePreflightBlocker
}

// TicketMergePreflightBlocker describes a package boundary value.
type TicketMergePreflightBlocker struct {
	Code   string
	Kind   string
	Detail string
}

type ticketMergeRemoteMetadata struct {
	State     string
	Base      string
	Head      string
	Mergeable string
}

// RunTicketMergePreflight executes the corresponding package operation.
func RunTicketMergePreflight(ctx context.Context, opts TicketMergePreflightOptions) (TicketMergePreflightResult, error) {
	if err := ctx.Err(); err != nil {
		return TicketMergePreflightResult{}, err
	}

	result := newTicketMergePreflightResult(opts)

	handoff, err := ValidateTicketMergeHandoff(opts.Ticket)
	if err != nil {
		result.addBlocker(TicketMergePreflightBlockMissingHandoff, TicketMergePreflightBlockKindOperational, err.Error())

		return result, nil
	}

	result.ReviewRef = handoff.ReviewRef

	if blocker := ticketMergeStrategyBlocker(result.MergeStrategy); blocker != nil {
		result.Blockers = append(result.Blockers, *blocker)

		return result.finalize(), nil
	}

	if blocker := ticketMergeSerialRepairBlocker(opts.Ticket, opts.Tickets); blocker != nil {
		result.Blockers = append(result.Blockers, *blocker)

		return result.finalize(), nil
	}

	if opts.Remote == nil {
		result.addBlocker(TicketMergePreflightBlockProviderUnreadable, TicketMergePreflightBlockKindOperational, "merge preflight remote reader is required")

		return result.finalize(), nil
	}

	snapshot, err := opts.Remote.Read(ctx, handoff.ReviewRef.URL)
	if err != nil {
		result.addBlocker(TicketMergePreflightBlockProviderUnreadable, TicketMergePreflightBlockKindOperational, err.Error())

		return result.finalize(), nil
	}

	result.Remote = snapshot
	result.Blockers = append(result.Blockers, ticketMergeRemoteBlockers(handoff, result.ExpectedBase, snapshot)...)

	return result.finalize(), nil
}

func newTicketMergePreflightResult(opts TicketMergePreflightOptions) TicketMergePreflightResult {
	expectedBase := strings.TrimSpace(opts.ExpectedBase)
	if expectedBase == "" {
		expectedBase = DefaultWorkIntegrationBranch
	}

	mergeStrategy := strings.TrimSpace(opts.MergeStrategy)
	if mergeStrategy == "" {
		mergeStrategy = TicketMergeStrategyRebase
	}

	return TicketMergePreflightResult{
		Allowed:       true,
		TicketKey:     opts.Ticket.Key,
		ExpectedBase:  expectedBase,
		MergeStrategy: mergeStrategy,
	}
}

func (result *TicketMergePreflightResult) addBlocker(code, kind, detail string) {
	result.Blockers = append(result.Blockers, TicketMergePreflightBlocker{
		Code:   code,
		Kind:   kind,
		Detail: detail,
	})
	result.Allowed = false
}

func (result TicketMergePreflightResult) finalize() TicketMergePreflightResult {
	result.Allowed = len(result.Blockers) == 0

	return result
}

func ticketMergeStrategyBlocker(strategy string) *TicketMergePreflightBlocker {
	switch strings.TrimSpace(strategy) {
	case TicketMergeStrategyMerge, TicketMergeStrategyRebase, TicketMergeStrategySquash:
		return nil
	default:
		return &TicketMergePreflightBlocker{
			Code:   TicketMergePreflightBlockInvalidMergeStrategy,
			Kind:   TicketMergePreflightBlockKindOperational,
			Detail: fmt.Sprintf("merge strategy is not allowed: %s", strategy),
		}
	}
}

func ticketMergeSerialRepairBlocker(ticket IndexedTicket, tickets []IndexedTicket) *TicketMergePreflightBlocker {
	for _, candidate := range tickets {
		if candidate.Key == ticket.Key {
			continue
		}

		if !isReviewRepairTicket(candidate) {
			continue
		}

		return &TicketMergePreflightBlocker{
			Code:   TicketMergePreflightBlockRepairPending,
			Kind:   TicketMergePreflightBlockKindRepair,
			Detail: fmt.Sprintf("review repair ticket must be resolved first: %s", candidate.Key),
		}
	}

	return nil
}

func ticketMergeRemoteBlockers(handoff TicketMergeHandoff, expectedBase string, snapshot RemoteReviewSnapshot) []TicketMergePreflightBlocker {
	metadata := parseTicketMergeRemoteMetadata(snapshot)
	blockers := []TicketMergePreflightBlocker{}

	if isTicketMergePullRequestClosed(metadata.State) {
		blockers = append(blockers, TicketMergePreflightBlocker{
			Code:   TicketMergePreflightBlockPullRequestClosed,
			Kind:   TicketMergePreflightBlockKindOperational,
			Detail: fmt.Sprintf("pull request is not open: %s", metadata.State),
		})
	}

	if metadata.Base != "" && metadata.Base != expectedBase {
		blockers = append(blockers, TicketMergePreflightBlocker{
			Code:   TicketMergePreflightBlockWrongBase,
			Kind:   TicketMergePreflightBlockKindOperational,
			Detail: fmt.Sprintf("pull request base is %q, expected %q", metadata.Base, expectedBase),
		})
	}

	if metadata.Head != "" && metadata.Head != handoff.Branch {
		blockers = append(blockers, TicketMergePreflightBlocker{
			Code:   TicketMergePreflightBlockHeadMismatch,
			Kind:   TicketMergePreflightBlockKindRepair,
			Detail: fmt.Sprintf("pull request head is %q, expected ticket branch %q", metadata.Head, handoff.Branch),
		})
	}

	if isTicketMergeExplicitlyNotMergeable(metadata.Mergeable) {
		blockers = append(blockers, TicketMergePreflightBlocker{
			Code:   TicketMergePreflightBlockNotMergeable,
			Kind:   TicketMergePreflightBlockKindRepair,
			Detail: fmt.Sprintf("pull request is not mergeable: %s", metadata.Mergeable),
		})
	}

	return blockers
}

func parseTicketMergeRemoteMetadata(snapshot RemoteReviewSnapshot) ticketMergeRemoteMetadata {
	switch snapshot.Reference.Provider {
	case ReviewProviderGitHub:
		return parseGitHubTicketMergeMetadata(snapshot.PullRequestJSON)
	default:
		return parseForgeTicketMergeMetadata(snapshot.PullRequestJSON)
	}
}

func parseForgeTicketMergeMetadata(rawJSON string) ticketMergeRemoteMetadata {
	type forgeRef struct {
		Ref string `json:"ref"`
	}

	type forgePullRequest struct {
		State     string   `json:"state"`
		Base      forgeRef `json:"base"`
		Head      forgeRef `json:"head"`
		Mergeable any      `json:"mergeable"`
	}

	var parsed forgePullRequest
	if err := json.Unmarshal([]byte(rawJSON), &parsed); err != nil {
		return ticketMergeRemoteMetadata{}
	}

	return ticketMergeRemoteMetadata{
		State:     strings.ToLower(parsed.State),
		Base:      parsed.Base.Ref,
		Head:      parsed.Head.Ref,
		Mergeable: normalizeTicketMergeableValue(parsed.Mergeable),
	}
}

func parseGitHubTicketMergeMetadata(rawJSON string) ticketMergeRemoteMetadata {
	type githubPullRequest struct {
		State       string `json:"state"`
		BaseRefName string `json:"baseRefName"`
		HeadRefName string `json:"headRefName"`
		Mergeable   any    `json:"mergeable"`
	}

	var parsed githubPullRequest
	if err := json.Unmarshal([]byte(rawJSON), &parsed); err != nil {
		return ticketMergeRemoteMetadata{}
	}

	return ticketMergeRemoteMetadata{
		State:     strings.ToLower(parsed.State),
		Base:      parsed.BaseRefName,
		Head:      parsed.HeadRefName,
		Mergeable: normalizeTicketMergeableValue(parsed.Mergeable),
	}
}

func normalizeTicketMergeableValue(value any) string {
	switch typed := value.(type) {
	case bool:
		if typed {
			return "true"
		}

		return "false"
	case string:
		return strings.ToLower(strings.TrimSpace(typed))
	default:
		return ""
	}
}

func isTicketMergePullRequestClosed(state string) bool {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "", "open":
		return false
	default:
		return true
	}
}

func isTicketMergeExplicitlyNotMergeable(mergeable string) bool {
	switch strings.ToLower(strings.TrimSpace(mergeable)) {
	case "false", "conflicting", "dirty", "not_mergeable":
		return true
	default:
		return false
	}
}
