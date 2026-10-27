package app

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	executioncore "forge.adrianpk.com/chr/internal/core/execution"
	notifycore "forge.adrianpk.com/chr/internal/core/notify"
	telegramcore "forge.adrianpk.com/chr/internal/core/telegram"
)

type telegramUnavailableRead struct{ telegramReadService }

func (telegramUnavailableRead) InspectReceiver(int64) (telegramcore.ReceiverState, error) {
	return telegramcore.ReceiverState{}, errors.New("private unavailable-service body")
}

// An actual failed read produces one bounded recovery message without leaking
// backend text or inventing an execution result; retained native state is intact.
func TestTelegramPresentationFailureRead(t *testing.T) {
	payloads := make(chan telegramHelpHTTPBody, 2)
	server := telegramHelpHTTPServer(t, payloads, 0, "")
	receiver, _ := telegramAppTestReceiver(t, telegramcore.ReceiverAPI{Client: server.Client(), BaseURL: server.URL})
	seedTelegramSavedHandoff(t, receiver.env)

	receiver.operations = telegramUnavailableRead{receiver.operations}

	update := retainTelegramHelpInput(t, receiver, 1, "/status")
	if err := receiver.respond(t.Context(), update); err != nil {
		t.Fatal(err)
	}

	if err := receiver.drainDeliveries(t.Context()); err != nil {
		t.Fatal(err)
	}

	select {
	case reply := <-payloads:
		for _, value := range []string{"requested read is unavailable", "Saved state is unchanged", "chr telegram doctor"} {
			if !strings.Contains(reply.Text, value) {
				t.Fatal("missing recovery context", reply)
			}
		}

		if strings.Contains(reply.Text, "private") || len(reply.Text) > 1800 {
			t.Fatal("unbounded or unsafe service failure", reply)
		}
	default:
		t.Fatal("failed read produced no recovery message")
	}

	checkTelegramNoWork(t, receiver.env)
}

// Notification failure cannot change a successfully saved native loop result or
// its ticket execution. No transport receipt is substituted for native success.
func TestTelegramPresentationFailureNativeResult(t *testing.T) {
	projectDir, devRoot := newTestWorkspace(t), t.TempDir()
	t.Setenv("CHR_TELEGRAM_BOT_TOKEN", "token-1")
	installFakeTicketStartHarnessExecutor(t, nil)

	var calls atomic.Int32

	installTelegramNotificationSender(t, func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusForbidden)

		if _, err := w.Write([]byte(`{"ok":false,"error_code":403}`)); err != nil {
			t.Error(err)
		}
	})
	addProject(t, devRoot, "example", projectDir)
	writeRuntimeTemplate(t, projectDir)
	writeTelegramNotifyConfig(t, projectDir)
	writeTicketMarkdownOnly(t, projectDir, "CHR-1-Ready.md", testLoopTicket("CHR-1", "Ready frontend", "ready_for_dev", "high", "critical", "feature", "onboarding", "frontend", []string{"uxui"}))

	if err := run(t.Context(), []string{"ticket", "loop", "--dev-root", devRoot, "--dir", projectDir}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal("delivery failure changed native result", err)
	}

	loops := listProjectLoopRuns(t, devRoot, projectID(projectDir))
	if len(loops) != 1 || loops[0].Status != executioncore.LoopRunStatusSucceeded || loops[0].ProcessedCount != 1 {
		t.Fatal("notification changed completed loop", loops)
	}

	executions := listTicketExecutions(t, devRoot, projectID(projectDir), "CHR-1")
	if len(executions) != 1 {
		t.Fatal("notification repeated native execution", len(executions))
	}

	records := listNotificationEventsForTest(t, devRoot, projectID(projectDir))
	assertNotificationEvent(t, records, 0, notifycore.EventLoopFinished, notifycore.StatusFailed)

	if calls.Load() != 1 || records[0].Event.DeliveredAt != nil {
		t.Fatal("failed send claimed delivery or repeated dispatch", calls.Load(), records)
	}
}
