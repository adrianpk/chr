package forge

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"strings"
)

// ReviewClient binds review/comment and merge protocol to configured host command IO.
// URL admission and permissions remain with the host adapter and its callers.
type ReviewClient struct {
	Reference Reference
	Command   string
	Commands  Commands
}

const (
	// CommentPlacementInline defines a package boundary constant.
	CommentPlacementInline = "inline"
	// CommentPlacementGeneral defines a package boundary constant.
	CommentPlacementGeneral = "general"
	// CommentPlacementSkipped defines a package boundary constant.
	CommentPlacementSkipped = "skipped"

	// CommentKindBlocking defines a package boundary constant.
	CommentKindBlocking = "blocking"
	// CommentKindNonBlocking defines a package boundary constant.
	CommentKindNonBlocking = "non_blocking"
)

// ReviewSnapshot describes a package boundary value.
type ReviewSnapshot struct {
	Reference          Reference
	PullRequestJSON    string
	Diff               string
	IssueCommentsJSON  string
	ReviewCommentsJSON string
	ChecksJSON         string
	Capabilities       ReviewCapabilities
}

// ReviewCapabilities describes a package boundary value.
type ReviewCapabilities struct {
	CanReadChecks       bool
	CanWriteInline      bool
	CanWriteGeneral     bool
	InlineFallbackNotes []string
}

// ReviewComment describes a package boundary value.
type ReviewComment struct {
	Body     string
	Path     string
	Line     int
	Side     string
	Blocking bool
}

// ReviewWriteResult describes a package boundary value.
type ReviewWriteResult struct {
	Reference Reference
	Comments  []WrittenComment
}

// MergeRequest describes a package boundary value.
type MergeRequest struct {
	URL      string
	Strategy string
}

// MergeResult describes a package boundary value.
type MergeResult struct {
	Reference Reference
	Strategy  string
	SHA       string
}

// WrittenComment describes a package boundary value.
type WrittenComment struct {
	Kind      string
	Placement string
	Path      string
	Line      int
	Reason    string
}

type remoteReviewPRMetadata struct {
	Head struct {
		SHA string `json:"sha"`
	} `json:"head"`
}

// Read executes the corresponding package operation.
func (adapter ReviewClient) Read(ctx context.Context) (ReviewSnapshot, error) {
	reference := adapter.Reference

	switch reference.Provider {
	case Forgejo:
		return adapter.readForgejo(ctx)
	case GitHub:
		return adapter.readGitHub(ctx)
	default:
		return ReviewSnapshot{}, fmt.Errorf("unsupported review provider: %s", reference.Provider)
	}
}

// WriteComments writes the requested output.
func (adapter ReviewClient) WriteComments(ctx context.Context, comments []ReviewComment) (ReviewWriteResult, error) {
	reference := adapter.Reference

	result := ReviewWriteResult{Reference: reference}

	for _, comment := range comments {
		written, err := adapter.writeComment(ctx, comment)
		if err != nil {
			return ReviewWriteResult{}, err
		}

		result.Comments = append(result.Comments, written)
	}

	return result, nil
}

// Merge executes the corresponding package operation.
func (a ReviewClient) Merge(ctx context.Context, strategy string) (MergeResult, error) {
	reference, command := a.Reference, a.Command

	payload, err := mergePayload(reference.Provider, strategy)
	if err != nil {
		return MergeResult{}, err
	}

	output, err := a.Commands.Input(ctx, command, payload, mergeArgs(reference, strategy)...)
	if err != nil {
		return MergeResult{}, err
	}

	return MergeResult{
		Reference: reference,
		Strategy:  strategy,
		SHA:       mergeResultSHA(output),
	}, nil
}

func (adapter ReviewClient) readForgejo(ctx context.Context) (ReviewSnapshot, error) {
	reference, command := adapter.Reference, adapter.Command

	prJSON, err := adapter.Commands.Output(ctx, command, "api", forgejoPullRequestPath(reference))
	if err != nil {
		return ReviewSnapshot{}, err
	}

	diff, err := adapter.Commands.Output(ctx, command, "api", forgejoPullRequestDiffPath(reference))
	if err != nil {
		return ReviewSnapshot{}, err
	}

	comments, err := adapter.Commands.Output(ctx, command, "api", forgejoIssueCommentsPath(reference))
	if err != nil {
		return ReviewSnapshot{}, err
	}

	checks, canReadChecks := adapter.readForgejoChecks(ctx, prJSON)

	return ReviewSnapshot{
		Reference:         reference,
		PullRequestJSON:   string(prJSON),
		Diff:              string(diff),
		IssueCommentsJSON: string(comments),
		ChecksJSON:        string(checks),
		Capabilities: ReviewCapabilities{
			CanReadChecks:   canReadChecks,
			CanWriteInline:  true,
			CanWriteGeneral: true,
		},
	}, nil
}

func (adapter ReviewClient) readGitHub(ctx context.Context) (ReviewSnapshot, error) {
	reference, command := adapter.Reference, adapter.Command

	prJSON, err := adapter.Commands.Output(ctx, command, "pr", "view", reference.URL, "--json", "number,title,state,baseRefName,headRefName,headRefOid,mergeable,isDraft,reviewDecision")
	if err != nil {
		return ReviewSnapshot{}, err
	}

	diff, err := adapter.Commands.Output(ctx, command, "pr", "diff", reference.URL)
	if err != nil {
		return ReviewSnapshot{}, err
	}

	issueComments, err := adapter.Commands.Output(ctx, command, "api", githubIssueCommentsPath(reference))
	if err != nil {
		return ReviewSnapshot{}, err
	}

	reviewComments, err := adapter.Commands.Output(ctx, command, "api", githubReviewCommentsPath(reference))
	if err != nil {
		return ReviewSnapshot{}, err
	}

	checks, err := adapter.Commands.Output(ctx, command, "pr", "checks", reference.URL, "--json", "name,state,link,bucket")
	if err != nil {
		return ReviewSnapshot{}, err
	}

	return ReviewSnapshot{
		Reference:          reference,
		PullRequestJSON:    string(prJSON),
		Diff:               string(diff),
		IssueCommentsJSON:  string(issueComments),
		ReviewCommentsJSON: string(reviewComments),
		ChecksJSON:         string(checks),
		Capabilities: ReviewCapabilities{
			CanReadChecks:   true,
			CanWriteInline:  true,
			CanWriteGeneral: true,
		},
	}, nil
}

func (adapter ReviewClient) readForgejoChecks(ctx context.Context, prJSON []byte) ([]byte, bool) {
	reference, command := adapter.Reference, adapter.Command

	metadata := remoteReviewPRMetadata{}
	if err := json.Unmarshal(prJSON, &metadata); err != nil || strings.TrimSpace(metadata.Head.SHA) == "" {
		return nil, false
	}

	checks, err := adapter.Commands.Output(ctx, command, "api", forgejoCommitStatusesPath(reference, metadata.Head.SHA))
	if err != nil {
		return nil, false
	}

	return checks, true
}

func (adapter ReviewClient) writeComment(ctx context.Context, comment ReviewComment) (WrittenComment, error) {
	if isInlineComment(comment) {
		written, err := adapter.writeInlineComment(ctx, comment)
		if err == nil {
			return written, nil
		}
	}

	return adapter.writeGeneralComment(ctx, comment)
}

func (adapter ReviewClient) writeInlineComment(ctx context.Context, comment ReviewComment) (WrittenComment, error) {
	reference, command := adapter.Reference, adapter.Command

	payload, err := inlineCommentPayload(reference.Provider, comment)
	if err != nil {
		return WrittenComment{}, err
	}

	_, err = adapter.Commands.Input(ctx, command, payload, inlineCommentArgs(reference)...)
	if err != nil {
		return WrittenComment{}, err
	}

	return WrittenComment{
		Kind:      remoteReviewCommentKind(comment),
		Placement: CommentPlacementInline,
		Path:      comment.Path,
		Line:      comment.Line,
	}, nil
}

func (adapter ReviewClient) writeGeneralComment(ctx context.Context, comment ReviewComment) (WrittenComment, error) {
	reference, command := adapter.Reference, adapter.Command

	payload, err := json.Marshal(map[string]string{"body": generalCommentBody(comment)})
	if err != nil {
		return WrittenComment{}, fmt.Errorf("encode general comment payload: %w", err)
	}

	_, err = adapter.Commands.Input(ctx, command, payload, generalCommentArgs(reference)...)
	if err != nil {
		return WrittenComment{}, err
	}

	return WrittenComment{
		Kind:      remoteReviewCommentKind(comment),
		Placement: CommentPlacementGeneral,
		Path:      comment.Path,
		Line:      comment.Line,
		Reason:    "inline_unavailable",
	}, nil
}

func isInlineComment(comment ReviewComment) bool {
	return strings.TrimSpace(comment.Path) != "" && comment.Line > 0
}

func remoteReviewCommentKind(comment ReviewComment) string {
	if comment.Blocking {
		return CommentKindBlocking
	}

	return CommentKindNonBlocking
}

func inlineCommentPayload(provider Provider, comment ReviewComment) ([]byte, error) {
	switch provider {
	case Forgejo:
		return json.Marshal(map[string]any{
			"event": "COMMENT",
			"comments": []map[string]any{
				{
					"path":         comment.Path,
					"body":         comment.Body,
					"new_position": comment.Line,
				},
			},
		})
	case GitHub:
		return json.Marshal(map[string]any{
			"body": comment.Body,
			"path": comment.Path,
			"line": comment.Line,
			"side": remoteReviewCommentSide(comment.Side),
		})
	default:
		return nil, fmt.Errorf("unsupported review provider: %s", provider)
	}
}

func generalCommentBody(comment ReviewComment) string {
	if !isInlineComment(comment) {
		return comment.Body
	}

	return fmt.Sprintf("%s\n\nLocation: `%s:%d`", comment.Body, comment.Path, comment.Line)
}

func remoteReviewCommentSide(side string) string {
	side = strings.ToUpper(strings.TrimSpace(side))
	if side == "LEFT" {
		return "LEFT"
	}

	return "RIGHT"
}

func inlineCommentArgs(reference Reference) []string {
	switch reference.Provider {
	case Forgejo:
		return []string{"api", "-X", "POST", forgejoPullRequestReviewsPath(reference), "-d", "@-"}
	case GitHub:
		return []string{"api", "-X", "POST", githubReviewCommentsPath(reference), "--input", "-"}
	default:
		return nil
	}
}

func generalCommentArgs(reference Reference) []string {
	switch reference.Provider {
	case Forgejo:
		return []string{"api", "-X", "POST", forgejoIssueCommentsPath(reference), "-d", "@-"}
	case GitHub:
		return []string{"api", "-X", "POST", githubIssueCommentsPath(reference), "--input", "-"}
	default:
		return nil
	}
}

func forgejoPullRequestPath(reference Reference) string {
	return path.Join("/repos", reference.Owner, reference.Repo, "pulls", strconv.Itoa(reference.Number))
}

func forgejoPullRequestDiffPath(reference Reference) string {
	return path.Join("/repos", reference.Owner, reference.Repo, "pulls", strconv.Itoa(reference.Number)+".diff")
}

func forgejoIssueCommentsPath(reference Reference) string {
	return path.Join("/repos", reference.Owner, reference.Repo, "issues", strconv.Itoa(reference.Number), "comments")
}

func forgejoPullRequestReviewsPath(reference Reference) string {
	return path.Join("/repos", reference.Owner, reference.Repo, "pulls", strconv.Itoa(reference.Number), "reviews")
}

func forgejoCommitStatusesPath(reference Reference, sha string) string {
	return path.Join("/repos", reference.Owner, reference.Repo, "commits", sha, "statuses")
}

func githubIssueCommentsPath(reference Reference) string {
	return path.Join("repos", reference.Owner, reference.Repo, "issues", strconv.Itoa(reference.Number), "comments")
}

func githubReviewCommentsPath(reference Reference) string {
	return path.Join("repos", reference.Owner, reference.Repo, "pulls", strconv.Itoa(reference.Number), "comments")
}

func mergePayload(provider Provider, strategy string) ([]byte, error) {
	switch provider {
	case Forgejo:
		return json.Marshal(map[string]string{"Do": forgejoMergeStrategy(strategy)})
	case GitHub:
		return []byte("{}"), nil
	default:
		return nil, fmt.Errorf("unsupported review provider: %s", provider)
	}
}

func mergeArgs(reference Reference, strategy string) []string {
	switch reference.Provider {
	case Forgejo:
		return []string{"api", "-X", "PUT", forgejoPullRequestMergePath(reference), "-d", "@-"}
	case GitHub:
		return []string{"pr", "merge", reference.URL, githubMergeStrategyFlag(strategy)}
	default:
		return nil
	}
}

func forgejoMergeStrategy(strategy string) string {
	switch strategy {
	case "merge":
		return "merge"
	case "squash":
		return "squash"
	default:
		return "rebase"
	}
}

func githubMergeStrategyFlag(strategy string) string {
	switch strategy {
	case "merge":
		return "--merge"
	case "squash":
		return "--squash"
	default:
		return "--rebase"
	}
}

func forgejoPullRequestMergePath(reference Reference) string {
	return path.Join("/repos", reference.Owner, reference.Repo, "pulls", strconv.Itoa(reference.Number), "merge")
}

func mergeResultSHA(output []byte) string {
	var parsed struct {
		SHA string `json:"sha"`
	}

	if err := json.Unmarshal(output, &parsed); err != nil {
		return ""
	}

	return strings.TrimSpace(parsed.SHA)
}
