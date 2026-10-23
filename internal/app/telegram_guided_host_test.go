package app

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core"
)

type telegramGuidedWaitingModel struct {
	base    core.QuizModelAdapter
	entered chan struct{}
	release chan struct{}
}

func (model telegramGuidedWaitingModel) InvokeQuizModel(ctx context.Context, request core.QuizModelRequest) (core.QuizModelResponse, error) {
	select {
	case model.entered <- struct{}{}:
	default:
	}

	select {
	case <-model.release:
	case <-ctx.Done():
		return core.QuizModelResponse{}, ctx.Err()
	}

	return model.base.InvokeQuizModel(ctx, request)
}

// TestGuidedUseCaseTelegramManaged keeps receiver status responsive during
// generation, then pauses and resumes at the canonical authoring boundary.
func TestGuidedUseCaseTelegramManaged(t *testing.T) {
	f, operations := guidedTelegramFixture(t)
	start := f.send(t, "Develop a new use case", 0)
	name := f.send(t, "Managed clarification", start.ID)
	f.send(t, "Retain finite work limits.", name.ID)

	entered, release := make(chan struct{}, 1), make(chan struct{})
	question := "What should the workflow preserve?\n" + strings.Repeat("Retain complete context. ", 150)

	base, ok := operations.quizModel.(*fakeServiceQuizModel)
	if !ok {
		t.Fatal("missing owned quiz fixture")
	}

	base.responses[0].Question.Prompt = question
	operations.quizModel = telegramGuidedWaitingModel{base: operations.quizModel, entered: entered, release: release}
	f.receiver.operations = operations

	var once sync.Once

	telegramControlRunReceiver(t, f.receiver, func() { once.Do(func() { close(release) }) })
	telegramWorkQueueMessage(t, f, "Finish capture")

	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("managed clarification did not enter")
	}

	telegramWorkQueueMessage(t, f, "/status")
	telegramWaitNotice(t, f, "Receiver:", 3*time.Second)
	once.Do(func() { close(release) })
	telegramWaitNotice(t, f, "Clarification for Managed clarification", 8*time.Second)

	upload := readGuidedTelegramUpload(t, f)
	if string(upload.Payload) != question {
		t.Fatal("managed question attachment was shortened")
	}

	telegramWorkQueueMessage(t, f, "Pause use case")
	telegramWaitNotice(t, f, "paused", 5*time.Second)

	c := f.sessions(t)[0]

	paused := readGuidedTelegramSession(t, operations, c)
	if paused.Guided.Status != "paused" {
		t.Fatal("canonical pause was not retained")
	}

	telegramWorkQueueMessage(t, f, "Resume use case")
	telegramWaitNotice(t, f, "Clarification for Managed clarification", 8*time.Second)

	resumed := readGuidedTelegramSession(t, operations, f.sessions(t)[0])
	if resumed.ID != paused.ID || resumed.Guided.Quiz == nil || !strings.Contains(resumed.Guided.Capture, "finite work limits") {
		t.Fatal("resume replaced the original run or input")
	}
}
