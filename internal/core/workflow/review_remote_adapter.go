package workflow

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

// OSReviewRemoteCommandRunner describes a package boundary value.
type OSReviewRemoteCommandRunner struct{}

// RemoteReviewAdapter describes a package boundary value.
type RemoteReviewAdapter struct {
	config workspaceconfig.ReviewProviderConfig
	runner workflowforge.Commands
}

// Output executes the corresponding package operation.
func (OSReviewRemoteCommandRunner) Output(ctx context.Context, command string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("run %s: %w", command, err)
	}

	return output, nil
}

// Input executes the corresponding package operation.
func (OSReviewRemoteCommandRunner) Input(ctx context.Context, command string, input []byte, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdin = bytes.NewReader(input)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("run %s: %w", command, err)
	}

	return output, nil
}

// NewRemoteReviewAdapter constructs the corresponding package service or value.
func NewRemoteReviewAdapter(config workspaceconfig.ReviewProviderConfig, runner workflowforge.Commands) RemoteReviewAdapter {
	if runner == nil {
		runner = OSReviewRemoteCommandRunner{}
	}

	return RemoteReviewAdapter{
		config: config,
		runner: runner,
	}
}

type reviewProviderEndpoint struct {
	reference workflowforge.Reference
	config    workspaceconfig.ReviewProviderTargetConfig
}

func (adapter RemoteReviewAdapter) resolve(rawURL string) (reviewProviderEndpoint, error) {
	reference, err := workflowforge.ParseReference(rawURL)
	if err != nil {
		return reviewProviderEndpoint{}, err
	}

	config, ok := reviewProviderTargetConfig(adapter.config, reference.Provider)
	if !ok || !config.Enabled {
		return reviewProviderEndpoint{}, fmt.Errorf("review provider is not enabled: %s", reference.Provider)
	}

	if strings.TrimSpace(config.CLI) == "" {
		return reviewProviderEndpoint{}, fmt.Errorf("review provider CLI is not configured: %s", reference.Provider)
	}

	return reviewProviderEndpoint{reference: reference, config: config}, nil
}

// Read resolves enabled provider configuration before review snapshot IO.
func (adapter RemoteReviewAdapter) Read(ctx context.Context, rawURL string) (workflowforge.ReviewSnapshot, error) {
	resolved, err := adapter.resolve(rawURL)
	if err != nil {
		return workflowforge.ReviewSnapshot{}, err
	}

	client := adapter.reviewClient(resolved)

	return client.Read(ctx)
}

// WriteComments retains the configured endpoint and comment fallback protocol.
func (adapter RemoteReviewAdapter) WriteComments(ctx context.Context, rawURL string, comments []workflowforge.ReviewComment) (workflowforge.ReviewWriteResult, error) {
	resolved, err := adapter.resolve(rawURL)
	if err != nil {
		return workflowforge.ReviewWriteResult{}, err
	}

	client := adapter.reviewClient(resolved)

	return client.WriteComments(ctx, comments)
}

// Merge binds an admitted source-review request to configured provider IO.
func (adapter RemoteReviewAdapter) Merge(ctx context.Context, request workflowforge.MergeRequest) (workflowforge.MergeResult, error) {
	resolved, err := adapter.resolve(request.URL)
	if err != nil {
		return workflowforge.MergeResult{}, err
	}

	client := adapter.reviewClient(resolved)

	return client.Merge(ctx, request.Strategy)
}

func (adapter RemoteReviewAdapter) reviewClient(resolved reviewProviderEndpoint) workflowforge.ReviewClient {
	return workflowforge.ReviewClient{Reference: resolved.reference, Command: resolved.config.CLI, Commands: adapter.runner}
}
