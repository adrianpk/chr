package app

import (
	"context"
	"errors"
	"maps"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

// TestTelegramRecoveryStartup replays an interrupted authoring acknowledgement
// through receiver startup while preserving another conversation's paused work.
func TestTelegramRecoveryStartup(t *testing.T) {
	for _, state := range []string{"pause", "review"} {
		t.Run(state, func(t *testing.T) { checkTelegramRecoveryStartup(t, state) })
	}
}

func checkTelegramRecoveryStartup(t *testing.T, state string) {
	t.Helper()

	f, operations, before := telegramRecoveryRetainedWork(t, state)
	calls := telegramAuthoringBridgeHTTP(t, f, func(authoringModelInput) string {
		return telegramAuthoringServiceOutput(t, "Preserve the retained scope.", "")
	}, telegramAuthoringBridgeCapture)
	f.send(t, "/prepare --project "+operations.projectContext.ProjectID+" restart scope", 0)

	conversation := telegramRecoveryCapture(t, f)
	operations = telegramOperationConcreteService(t, f.receiver)
	input := telegramOperationRetainInput(t, f, "/done "+conversation)

	f.receiver.operations = telegramAuthoringLostAck{service: operations, lost: &atomic.Bool{}}
	if err := f.receiver.respond(t.Context(), input.Update); !errors.Is(err, errTelegramAuthoringLostAck) {
		t.Fatal("owner acknowledgement was not interrupted", err)
	}

	image := workflowCLIImage(t, operations.projectContext.ProjectDir)
	telegramAuthoringReopen(t, f)

	f.receiver.operations = operations

	cancel, finished := startTelegramRecoveryReceiver(t, f.receiver)
	defer cancel()

	waitTelegramRecoveryPrepared(t, f, conversation)
	cancel()
	waitTelegramRecoveryReceiver(t, finished)

	if calls.Load() != 1 || !maps.Equal(image, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("startup repeated the accepted authoring effect", calls.Load())
	}

	if after := checkTelegramWorkRunCount(t, operations, 1); !reflect.DeepEqual(before, after) {
		t.Fatal("startup changed paused invocation, attempts or budget", after)
	}
}

func telegramRecoveryRetainedWork(t *testing.T, state string) (*telegramConversationAppFixture, service, core.WorkflowInvocation) {
	t.Helper()

	var (
		f          *telegramConversationAppFixture
		operations service
		id         string
	)
	if state == "review" {
		f, operations, id = telegramControlDecisionFixture(t, "approve")
		telegramControlOffer(t, f, "/approve "+id)
	} else {
		var ref telegramcore.WorkReference

		f, operations, ref = telegramWorkAdmitted(t)

		id = ref.Invocation.ID
		if _, err := operations.StopWorkflow(t.Context(), id, "pause"); err != nil {
			t.Fatal(err)
		}
	}

	before, err := operations.ReadWorkflow(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	return f, operations, before
}

func telegramRecoveryCapture(t *testing.T, f *telegramConversationAppFixture) string {
	t.Helper()

	for _, c := range f.sessions(t) {
		if c.Phase == "collecting" {
			return c.ID
		}
	}

	t.Fatal("capture unavailable")

	return ""
}

func startTelegramRecoveryReceiver(t *testing.T, receiver telegramReceiver) (context.CancelFunc, <-chan error) {
	t.Helper()

	ctx, cancel := context.WithCancel(t.Context())

	finished := make(chan error, 1)
	go func() { finished <- receiver.run(ctx); close(finished) }()

	t.Cleanup(func() { cancel(); waitTelegramRecoveryReceiver(t, finished) })

	return cancel, finished
}

func waitTelegramRecoveryReceiver(t *testing.T, finished <-chan error) {
	t.Helper()

	select {
	case err := <-finished:
		if err != nil {
			t.Error("receiver shutdown failed", err)
		}
	case <-time.After(30 * time.Second):
		t.Error("receiver did not finish its owned activations")
	}
}

func waitTelegramRecoveryPrepared(t *testing.T, f *telegramConversationAppFixture, id string) {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		c := f.session(t, id)
		if c.Phase == "prepared" && c.AuthoringSessionRef != "" {
			return
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("startup did not recover retained authoring")
}
