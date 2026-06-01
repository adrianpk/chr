package chc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

const (
	RemoteReviewCommentPlacementInline  = "inline"
	RemoteReviewCommentPlacementGeneral = "general"
	RemoteReviewCommentPlacementSkipped = "skipped"

	RemoteReviewCommentKindBlocking    = "blocking"
	RemoteReviewCommentKindNonBlocking = "non_blocking"
)

type ReviewRemoteCommandRunner interface {
	Output(ctx context.Context, command string, args ...string) ([]byte, error)
	Input(ctx context.Context, command string, input []byte, args ...string) ([]byte, error)
}

type OSReviewRemoteCommandRunner struct{}

type RemoteReviewAdapter struct {
	config ReviewProviderConfig
	runner ReviewRemoteCommandRunner
}

type RemoteReviewSnapshot struct {
	Reference          ReviewProviderReference
	PullRequestJSON    string
	Diff               string
	IssueCommentsJSON  string
	ReviewCommentsJSON string
	ChecksJSON         string
	Capabilities       RemoteReviewCapabilities
}

type RemoteReviewCapabilities struct {
	CanReadChecks       bool
	CanWriteInline      bool
	CanWriteGeneral     bool
	InlineFallbackNotes []string
}

type RemoteReviewComment struct {
	Body     string
	Path     string
	Line     int
	Side     string
	Blocking bool
}

type RemoteReviewWriteResult struct {
	Reference ReviewProviderReference
	Comments  []RemoteReviewWrittenComment
}

type RemoteReviewWrittenComment struct {
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

func (OSReviewRemoteCommandRunner) Output(ctx context.Context, command string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("run %s: %w", command, err)
	}

	return output, nil
}

func (OSReviewRemoteCommandRunner) Input(
	ctx context.Context,
	command string,
	input []byte,
	args ...string,
) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdin = bytes.NewReader(input)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("run %s: %w", command, err)
	}

	return output, nil
}

func NewRemoteReviewAdapter(config ReviewProviderConfig, runner ReviewRemoteCommandRunner) RemoteReviewAdapter {
	if runner == nil {
		runner = OSReviewRemoteCommandRunner{}
	}

	return RemoteReviewAdapter{
		config: config,
		runner: runner,
	}
}

func (adapter RemoteReviewAdapter) Read(ctx context.Context, rawURL string) (RemoteReviewSnapshot, error) {
	reference, config, err := adapter.resolve(rawURL)
	if err != nil {
		return RemoteReviewSnapshot{}, err
	}

	switch reference.Provider {
	case ReviewProviderForgejo:
		return adapter.readForgejo(ctx, reference, config)
	case ReviewProviderGitHub:
		return adapter.readGitHub(ctx, reference, config)
	default:
		return RemoteReviewSnapshot{}, fmt.Errorf("unsupported review provider: %s", reference.Provider)
	}
}

func (adapter RemoteReviewAdapter) WriteComments(
	ctx context.Context,
	rawURL string,
	comments []RemoteReviewComment,
) (RemoteReviewWriteResult, error) {
	reference, config, err := adapter.resolve(rawURL)
	if err != nil {
		return RemoteReviewWriteResult{}, err
	}

	result := RemoteReviewWriteResult{Reference: reference}
	for _, comment := range comments {
		written, err := adapter.writeComment(ctx, reference, config, comment)
		if err != nil {
			return RemoteReviewWriteResult{}, err
		}

		result.Comments = append(result.Comments, written)
	}

	return result, nil
}

func (adapter RemoteReviewAdapter) resolve(
	rawURL string,
) (ReviewProviderReference, ReviewProviderTargetConfig, error) {
	reference, err := DetectReviewProviderReference(rawURL)
	if err != nil {
		return ReviewProviderReference{}, ReviewProviderTargetConfig{}, err
	}

	config, ok := reviewProviderTargetConfig(adapter.config, reference.Provider)
	if !ok || !config.Enabled {
		return ReviewProviderReference{}, ReviewProviderTargetConfig{}, fmt.Errorf("review provider is not enabled: %s", reference.Provider)
	}

	if strings.TrimSpace(config.CLI) == "" {
		return ReviewProviderReference{}, ReviewProviderTargetConfig{}, fmt.Errorf("review provider CLI is not configured: %s", reference.Provider)
	}

	return reference, config, nil
}

func (adapter RemoteReviewAdapter) readForgejo(
	ctx context.Context,
	reference ReviewProviderReference,
	config ReviewProviderTargetConfig,
) (RemoteReviewSnapshot, error) {
	prJSON, err := adapter.runner.Output(ctx, config.CLI, "api", forgejoPullRequestPath(reference))
	if err != nil {
		return RemoteReviewSnapshot{}, err
	}

	diff, err := adapter.runner.Output(ctx, config.CLI, "api", forgejoPullRequestDiffPath(reference))
	if err != nil {
		return RemoteReviewSnapshot{}, err
	}

	comments, err := adapter.runner.Output(ctx, config.CLI, "api", forgejoIssueCommentsPath(reference))
	if err != nil {
		return RemoteReviewSnapshot{}, err
	}

	checks, canReadChecks := adapter.readForgejoChecks(ctx, reference, config, prJSON)

	return RemoteReviewSnapshot{
		Reference:         reference,
		PullRequestJSON:   string(prJSON),
		Diff:              string(diff),
		IssueCommentsJSON: string(comments),
		ChecksJSON:        string(checks),
		Capabilities: RemoteReviewCapabilities{
			CanReadChecks:   canReadChecks,
			CanWriteInline:  true,
			CanWriteGeneral: true,
		},
	}, nil
}

func (adapter RemoteReviewAdapter) readGitHub(
	ctx context.Context,
	reference ReviewProviderReference,
	config ReviewProviderTargetConfig,
) (RemoteReviewSnapshot, error) {
	prJSON, err := adapter.runner.Output(ctx, config.CLI, "pr", "view", reference.URL, "--json", "number,title,state,baseRefName,headRefName,headRefOid,mergeable,isDraft,reviewDecision")
	if err != nil {
		return RemoteReviewSnapshot{}, err
	}

	diff, err := adapter.runner.Output(ctx, config.CLI, "pr", "diff", reference.URL)
	if err != nil {
		return RemoteReviewSnapshot{}, err
	}

	issueComments, err := adapter.runner.Output(ctx, config.CLI, "api", githubIssueCommentsPath(reference))
	if err != nil {
		return RemoteReviewSnapshot{}, err
	}

	reviewComments, err := adapter.runner.Output(ctx, config.CLI, "api", githubReviewCommentsPath(reference))
	if err != nil {
		return RemoteReviewSnapshot{}, err
	}

	checks, err := adapter.runner.Output(ctx, config.CLI, "pr", "checks", reference.URL, "--json", "name,state,link,bucket")
	if err != nil {
		return RemoteReviewSnapshot{}, err
	}

	return RemoteReviewSnapshot{
		Reference:          reference,
		PullRequestJSON:    string(prJSON),
		Diff:               string(diff),
		IssueCommentsJSON:  string(issueComments),
		ReviewCommentsJSON: string(reviewComments),
		ChecksJSON:         string(checks),
		Capabilities: RemoteReviewCapabilities{
			CanReadChecks:   true,
			CanWriteInline:  true,
			CanWriteGeneral: true,
		},
	}, nil
}

func (adapter RemoteReviewAdapter) readForgejoChecks(
	ctx context.Context,
	reference ReviewProviderReference,
	config ReviewProviderTargetConfig,
	prJSON []byte,
) ([]byte, bool) {
	metadata := remoteReviewPRMetadata{}
	if err := json.Unmarshal(prJSON, &metadata); err != nil || strings.TrimSpace(metadata.Head.SHA) == "" {
		return nil, false
	}

	checks, err := adapter.runner.Output(ctx, config.CLI, "api", forgejoCommitStatusesPath(reference, metadata.Head.SHA))
	if err != nil {
		return nil, false
	}

	return checks, true
}

func (adapter RemoteReviewAdapter) writeComment(
	ctx context.Context,
	reference ReviewProviderReference,
	config ReviewProviderTargetConfig,
	comment RemoteReviewComment,
) (RemoteReviewWrittenComment, error) {
	if isInlineComment(comment) {
		written, err := adapter.writeInlineComment(ctx, reference, config, comment)
		if err == nil {
			return written, nil
		}
	}

	return adapter.writeGeneralComment(ctx, reference, config, comment)
}

func (adapter RemoteReviewAdapter) writeInlineComment(
	ctx context.Context,
	reference ReviewProviderReference,
	config ReviewProviderTargetConfig,
	comment RemoteReviewComment,
) (RemoteReviewWrittenComment, error) {
	payload, err := inlineCommentPayload(reference.Provider, comment)
	if err != nil {
		return RemoteReviewWrittenComment{}, err
	}

	_, err = adapter.runner.Input(ctx, config.CLI, payload, inlineCommentArgs(reference)...)
	if err != nil {
		return RemoteReviewWrittenComment{}, err
	}

	return RemoteReviewWrittenComment{
		Kind:      remoteReviewCommentKind(comment),
		Placement: RemoteReviewCommentPlacementInline,
		Path:      comment.Path,
		Line:      comment.Line,
	}, nil
}

func (adapter RemoteReviewAdapter) writeGeneralComment(
	ctx context.Context,
	reference ReviewProviderReference,
	config ReviewProviderTargetConfig,
	comment RemoteReviewComment,
) (RemoteReviewWrittenComment, error) {
	payload, err := json.Marshal(map[string]string{"body": generalCommentBody(comment)})
	if err != nil {
		return RemoteReviewWrittenComment{}, fmt.Errorf("encode general comment payload: %w", err)
	}

	_, err = adapter.runner.Input(ctx, config.CLI, payload, generalCommentArgs(reference)...)
	if err != nil {
		return RemoteReviewWrittenComment{}, err
	}

	return RemoteReviewWrittenComment{
		Kind:      remoteReviewCommentKind(comment),
		Placement: RemoteReviewCommentPlacementGeneral,
		Path:      comment.Path,
		Line:      comment.Line,
		Reason:    "inline_unavailable",
	}, nil
}

func isInlineComment(comment RemoteReviewComment) bool {
	return strings.TrimSpace(comment.Path) != "" && comment.Line > 0
}

func remoteReviewCommentKind(comment RemoteReviewComment) string {
	if comment.Blocking {
		return RemoteReviewCommentKindBlocking
	}

	return RemoteReviewCommentKindNonBlocking
}

func inlineCommentPayload(provider ReviewProvider, comment RemoteReviewComment) ([]byte, error) {
	switch provider {
	case ReviewProviderForgejo:
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
	case ReviewProviderGitHub:
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

func generalCommentBody(comment RemoteReviewComment) string {
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

func inlineCommentArgs(reference ReviewProviderReference) []string {
	switch reference.Provider {
	case ReviewProviderForgejo:
		return []string{"api", "-X", "POST", forgejoPullRequestReviewsPath(reference), "-d", "@-"}
	case ReviewProviderGitHub:
		return []string{"api", "-X", "POST", githubReviewCommentsPath(reference), "--input", "-"}
	default:
		return nil
	}
}

func generalCommentArgs(reference ReviewProviderReference) []string {
	switch reference.Provider {
	case ReviewProviderForgejo:
		return []string{"api", "-X", "POST", forgejoIssueCommentsPath(reference), "-d", "@-"}
	case ReviewProviderGitHub:
		return []string{"api", "-X", "POST", githubIssueCommentsPath(reference), "--input", "-"}
	default:
		return nil
	}
}

func forgejoPullRequestPath(reference ReviewProviderReference) string {
	return path.Join("/repos", reference.Owner, reference.Repo, "pulls", strconv.Itoa(reference.Number))
}

func forgejoPullRequestDiffPath(reference ReviewProviderReference) string {
	return path.Join("/repos", reference.Owner, reference.Repo, "pulls", strconv.Itoa(reference.Number)+".diff")
}

func forgejoIssueCommentsPath(reference ReviewProviderReference) string {
	return path.Join("/repos", reference.Owner, reference.Repo, "issues", strconv.Itoa(reference.Number), "comments")
}

func forgejoPullRequestReviewsPath(reference ReviewProviderReference) string {
	return path.Join("/repos", reference.Owner, reference.Repo, "pulls", strconv.Itoa(reference.Number), "reviews")
}

func forgejoCommitStatusesPath(reference ReviewProviderReference, sha string) string {
	return path.Join("/repos", reference.Owner, reference.Repo, "commits", sha, "statuses")
}

func githubIssueCommentsPath(reference ReviewProviderReference) string {
	return path.Join("repos", reference.Owner, reference.Repo, "issues", strconv.Itoa(reference.Number), "comments")
}

func githubReviewCommentsPath(reference ReviewProviderReference) string {
	return path.Join("repos", reference.Owner, reference.Repo, "pulls", strconv.Itoa(reference.Number), "comments")
}
