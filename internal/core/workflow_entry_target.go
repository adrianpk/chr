package core

import (
	"context"
	"fmt"
	"strings"

	projectcore "forge.adrianpk.com/chr/internal/core/project"
)

// WorkflowEntryRepository resolves only an explicit public canonical identity.
// SSH aliases, credentials and local transports cannot become durable URLs.
func WorkflowEntryRepository(ctx context.Context, projectDir, remote string) (string, error) {
	if strings.TrimSpace(remote) == "" || strings.HasPrefix(remote, "-") || strings.ContainsAny(remote, "\x00\r\n") {
		return "", fmt.Errorf("work canonical_remote is required for ticket authority")
	}

	reader := workflowIntegrationReader{GitReader: projectcore.GitReader{Directory: projectDir}}

	raw, err := reader.Output(ctx, "remote", "get-url", "--", remote)
	if err != nil {
		return "", fmt.Errorf("read canonical remote identity: %w", err)
	}

	return workflowCanonicalRepository(strings.TrimSpace(raw))
}
