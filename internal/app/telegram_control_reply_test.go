package app

import (
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramHoldControlAck struct {
	client  core.TelegramHTTPClient
	release <-chan struct{}
	held    atomic.Bool
}

func (c *telegramHoldControlAck) Do(request *http.Request) (*http.Response, error) {
	prompt, err := telegramControlPromptRequest(request)
	if err != nil {
		return nil, err
	}

	response, err := c.client.Do(request)
	if err != nil || !prompt || !c.held.CompareAndSwap(false, true) {
		return response, err
	}

	select {
	case <-c.release:
		return response, nil
	case <-request.Context().Done():
		return nil, errors.Join(request.Context().Err(), response.Body.Close())
	}
}

func telegramControlPromptRequest(request *http.Request) (bool, error) {
	if !strings.HasSuffix(request.URL.Path, "/sendMessage") || request.GetBody == nil {
		return false, nil
	}

	body, err := request.GetBody()
	if err != nil {
		return false, err
	}

	var data struct {
		Text string `json:"text"`
	}

	readErr := json.NewDecoder(body).Decode(&data)

	return strings.Contains(data.Text, "Native control preview:"), errors.Join(readErr, body.Close())
}

// TestTelegramControlReplyArrival holds an actual HTTP send acknowledgement while
// intake receives its reply. No model or effect runs until the binding commits.
func TestTelegramControlReplyArrival(t *testing.T) {
	f, operations, ref := telegramWorkAdmitted(t)

	before, err := operations.StopWorkflow(t.Context(), ref.Invocation.ID, "pause")
	if err != nil {
		t.Fatal(err)
	}

	release := make(chan struct{})

	var once sync.Once

	f.receiver.api.Client = &telegramHoldControlAck{client: f.receiver.api.Client, release: release}
	telegramControlRunReceiver(t, f.receiver, func() { once.Do(func() { close(release) }) })
	telegramWorkQueueMessage(t, f, "/cancel "+ref.Invocation.ID)

	prompt := telegramWaitNotice(t, f, "Native control preview:", 3*time.Second)
	queueTelegramControlReply(t, f, prompt.ID, "cancel")
	checkTelegramControlReplyWaiting(t, f)

	if current := checkTelegramWorkRunCount(t, operations, 1); current.Revision != before.Revision || current.StopIntent != "pause" {
		t.Fatal("unbound reply changed native stop intent", current)
	}

	once.Do(func() { close(release) })
	telegramWaitNotice(t, f, "Native control confirmed.", 3*time.Second)
	checkTelegramControlOutcome(t, "cancel", checkTelegramWorkRunCount(t, operations, 1))
}

func queueTelegramControlReply(t *testing.T, f *telegramConversationAppFixture, messageID int64, text string) {
	t.Helper()

	payload, err := json.Marshal([]any{map[string]any{"update_id": f.next, "message": map[string]any{"message_id": f.next, "from": map[string]any{"id": 100001}, "chat": map[string]any{"id": 100001, "type": "private"}, "text": text, "reply_to_message": map[string]any{"message_id": messageID}}}})
	if err != nil {
		t.Fatal(err)
	}

	f.next++
	f.updates <- payload
}

func checkTelegramControlReplyWaiting(t *testing.T, f *telegramConversationAppFixture) {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		pending, err := f.receiver.store.PendingApplicationInputs(t.Context(), f.receiver.botID, core.TelegramProcessingDecision, time.Now())
		if err != nil {
			t.Fatal(err)
		}

		if len(pending) == 1 {
			_, recognized, err := f.receiver.store.ResolveTelegramActionAnswer(t.Context(), pending[0])
			if !recognized || !errors.Is(err, core.ErrTelegramReplyPending) {
				t.Fatal("reply escaped pending definite acknowledgement", recognized, err)
			}

			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("early reply was acknowledged before its prompt binding")
}

// TestTelegramControlLaneFence keeps a continuation accepted by a late reply
// from executing in a reserved decision slot. The ordinary retry does one step.
func TestTelegramControlLaneFence(t *testing.T) {
	f, operations, ref := telegramWorkAdmitted(t)
	token := telegramControlOffer(t, f, "/continue "+ref.Invocation.ID)
	input := telegramOperationRetainInput(t, f, "/answer "+token+" continue")
	lane := core.TelegramProcessingDecision
	f.receiver.processingLane = &lane

	images := workflowCLIImage(t, operations.projectContext.ProjectDir)
	if err := f.receiver.respond(t.Context(), input.Update); !errors.Is(err, errTelegramLaneChanged) || !maps.Equal(images, workflowCLIImage(t, operations.projectContext.ProjectDir)) {
		t.Fatal("continuation executed in a reserved slot", err)
	}

	lane = core.TelegramProcessingOrdinary

	if err := f.receiver.respond(t.Context(), input.Update); err != nil {
		t.Fatal(err)
	}

	checkTelegramControlOutcome(t, "continue", checkTelegramWorkRunCount(t, operations, 1))
}

// TestTelegramControlAuthoringLane prevents a newly bound slash-command reply
// from reaching its real model provider through a reserved decision activation.
func TestTelegramControlAuthoringLane(t *testing.T) {
	f := newTelegramConversationAppFixture(t)
	calls := telegramAuthoringBridgeHTTP(t, f, func(input authoringModelInput) string {
		return telegramAuthoringBridgeGeneration(t, input)
	}, telegramAuthoringBridgeCapture)
	f.send(t, "/prepare --project example-project Reply capacity", 0)

	c := f.sessions(t)[0]
	prompt := f.send(t, "/reply "+c.ID+" Keep bounded retries.", 0)
	queueTelegramControlReply(t, f, prompt.ID, "/done "+c.ID)

	lane := core.TelegramProcessingDecision

	f.receiver.processingLane = &lane
	if _, err := f.receiver.step(t.Context()); !errors.Is(err, errTelegramLaneChanged) || calls.Load() != 0 {
		t.Fatal("authoring entered the provider from a reserved activation", calls.Load(), err)
	}

	lane = core.TelegramProcessingOrdinary

	if err := f.receiver.drain(t.Context()); err != nil {
		t.Fatal(err)
	}

	telegramWaitNotice(t, f, "saved as a dump", time.Second)

	if calls.Load() != 1 {
		t.Fatal("ordinary reply did not generate exactly once", calls.Load())
	}
}
