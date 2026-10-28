package forge

import (
	"context"
	"errors"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Submitter dispatches an already journaled intent through the host's guarded provider.
// Acceptance is only a reply; canonical observation still establishes delivery.
type Submitter interface {
	SubmitWorkflowIntegration(context.Context, workflowstate.Submission) (workflowstate.SubmissionResponse, error)
}

// Submit dispatches once under the watched context, then retains the untrusted
// response with the original five-second detached cleanup bound. The host binds
// Responses to the exact invocation and submission journal identity.
func Submit(ctx, watched context.Context, submission workflowstate.Submission, provider Submitter, responses Responses) (workflowstate.Submission, error) {
	response, submitErr := provider.SubmitWorkflowIntegration(watched, submission)

	cleanup, finish := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer finish()

	submission, err := responses.Retain(cleanup, Response{Reply: response, Failure: submitErr})

	return submission, errors.Join(submitErr, err)
}
