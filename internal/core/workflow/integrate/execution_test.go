package integrate

import (
	"context"
	"errors"
	"testing"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type validationPort struct {
	completion Completion
	acceptErr  error
	cleanupErr error
}

func (port *validationPort) Validate(ctx context.Context, _ *workflowstate.TaskResult) error {
	return ctx.Err()
}

func (port *validationPort) Accept(ctx context.Context, completion Completion) error {
	port.completion, port.cleanupErr = completion, ctx.Err()
	completion.Candidate.Validation.WaiverReason = "retained by guarded acceptance"

	return port.acceptErr
}

// Validation deadlines and operator stops must remain distinct at acceptance.
// A detached, bounded cleanup context must still retain the acceptance result
// and both errors when validation or the guarded store transition fails.
func TestValidationCompletion(t *testing.T) {
	acceptErr := errors.New("acceptance rejected stale evidence")
	for _, test := range []struct {
		name        string
		expired     bool
		stopped     bool
		acceptErr   error
		validateErr error
	}{
		{name: "accepted"},
		{name: "acceptance rejected", acceptErr: acceptErr},
		{name: "deadline", expired: true, validateErr: context.DeadlineExceeded},
		{name: "stop and rejection", stopped: true, acceptErr: acceptErr, validateErr: context.Canceled},
	} {
		t.Run(test.name, func(t *testing.T) {
			watched, stop := context.WithCancel(t.Context())
			defer stop()

			until := time.Now().Add(time.Hour)
			if test.expired {
				until = time.Now().Add(-time.Second)
			}

			ctx, cancel := context.WithDeadline(watched, until)
			defer cancel()

			if test.stopped {
				stop()
			}

			port := &validationPort{acceptErr: test.acceptErr}

			candidate, err := ExecuteWorkflowValidation(ctx, watched, ValidationInput{InvocationID: "invocation", Round: 7}, port)
			checkValidationErrors(t, err, test.validateErr, test.acceptErr)

			if port.cleanupErr != nil || port.completion.Interrupted != test.stopped || !errors.Is(port.completion.Failure, test.validateErr) {
				t.Fatalf("acceptance received wrong cancellation or failure: %+v, cleanup: %v", port.completion, port.cleanupErr)
			}

			if candidate.Validation.Successful != (test.validateErr == nil) || candidate.Validation.FinishedAt.IsZero() || candidate.Validation.WaiverReason != "retained by guarded acceptance" {
				t.Fatalf("candidate lost its validation or accepted result: %+v", candidate)
			}
		})
	}
}

func checkValidationErrors(t *testing.T, actual, validation, acceptance error) {
	t.Helper()

	if validation != nil && !errors.Is(actual, validation) || acceptance != nil && !errors.Is(actual, acceptance) || validation == nil && acceptance == nil && actual != nil {
		t.Fatalf("result lost validation or acceptance error: %v", actual)
	}
}
