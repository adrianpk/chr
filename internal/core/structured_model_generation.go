package core

import (
	"context"

	harnesshost "forge.adrianpk.com/chr/internal/core/harness/host"
	"forge.adrianpk.com/chr/internal/harness"
)

// ValidateStructuredModelOutput records the original caller's typed/schema
// validation, independently of JSON framing and process/network success.
func ValidateStructuredModelOutput(ctx context.Context, request StructuredModelRequest, validationErr error) error {
	return harnesshost.ValidateGenerationOutput(ctx, harness.AgentRunRequest{ID: request.ID, Purpose: request.Purpose, ProjectDir: request.ProjectDir}, validationErr)
}
