package core

import (
	"strings"
	"testing"
	"time"
)

func TestParseRemoteCommandRequestParsesSupportedCommands(t *testing.T) {
	for _, test := range supportedRemoteCommandTests() {
		t.Run(test.name, func(t *testing.T) {
			request, err := ParseRemoteCommandRequest(remoteParseOptions(test.text))
			if err != nil {
				t.Fatalf("ParseRemoteCommandRequest returned error: %v", err)
			}

			if request.Intent != test.intent {
				t.Fatalf("intent = %q, want %q", request.Intent, test.intent)
			}

			if request.ProjectSelector != test.projectSelector {
				t.Fatalf("project selector = %q, want %q", request.ProjectSelector, test.projectSelector)
			}

			test.assert(t, request)
		})
	}
}

func TestParseRemoteCommandRequestRejectsUnsupportedCommands(t *testing.T) {
	for _, test := range unsupportedRemoteCommandTests() {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseRemoteCommandRequest(remoteParseOptions(test.text))
			if err == nil {
				t.Fatal("ParseRemoteCommandRequest returned nil error")
			}

			if !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("error = %q, want %q", err.Error(), test.wantErr)
			}
		})
	}
}

func TestValidateRemoteCommandRequestRejectsInvalidRequests(t *testing.T) {
	for _, test := range invalidRemoteCommandRequestTests() {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateRemoteCommandRequest(test.request)
			if err == nil {
				t.Fatal("ValidateRemoteCommandRequest returned nil error")
			}

			if !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("error = %q, want %q", err.Error(), test.wantErr)
			}
		})
	}
}

func TestRemoteCommandIntentValidatesKnownValues(t *testing.T) {
	for _, intent := range []RemoteCommandIntent{
		RemoteCommandIntentStatus,
		RemoteCommandIntentLoopStart,
		RemoteCommandIntentTicketStart,
		RemoteCommandIntentTicketComment,
	} {
		if !intent.Valid() {
			t.Fatalf("RemoteCommandIntent.Valid() = false for %s", intent)
		}
	}

	if RemoteCommandIntent("delete_everything").Valid() {
		t.Fatal("unknown remote command intent is valid")
	}
}

func supportedRemoteCommandTests() []struct {
	name            string
	text            string
	intent          RemoteCommandIntent
	projectSelector string
	assert          func(t *testing.T, request RemoteCommandRequest)
} {
	return []struct {
		name            string
		text            string
		intent          RemoteCommandIntent
		projectSelector string
		assert          func(t *testing.T, request RemoteCommandRequest)
	}{
		{name: "status", text: "@example status", intent: RemoteCommandIntentStatus, projectSelector: "example", assert: assertRemoteStatus},
		{name: "loop", text: "@example loop start --status ready --priority high --max-iterations 2", intent: RemoteCommandIntentLoopStart, projectSelector: "example", assert: assertRemoteLoop},
		{name: "ticket start", text: "ticket start CHR-12", intent: RemoteCommandIntentTicketStart, assert: assertRemoteTicketStart},
		{name: "ticket comment", text: "ticket comment CHR-12 blocked on CI logs", intent: RemoteCommandIntentTicketComment, assert: assertRemoteTicketComment},
	}
}

func unsupportedRemoteCommandTests() []struct {
	name    string
	text    string
	wantErr string
} {
	return []struct {
		name    string
		text    string
		wantErr string
	}{
		{name: "empty", text: " ", wantErr: "remote command text is required"},
		{name: "unknown root", text: "merge now", wantErr: "unsupported remote command"},
		{name: "status args", text: "status now", wantErr: "status remote command does not accept arguments"},
		{name: "loop unsafe arg", text: "loop start now", wantErr: "argument must be a flag"},
		{name: "loop unknown flag", text: "loop start --delete true", wantErr: "unsupported loop remote command flag"},
		{name: "ticket unknown action", text: "ticket delete CHR-1", wantErr: "unsupported ticket remote command"},
	}
}

func invalidRemoteCommandRequestTests() []struct {
	name    string
	request RemoteCommandRequest
	wantErr string
} {
	return []struct {
		name    string
		request RemoteCommandRequest
		wantErr string
	}{
		{name: "sender", request: remoteRequestWith(func(request *RemoteCommandRequest) { request.Sender.ExternalID = "" }), wantErr: "sender external id is required"},
		{name: "intent", request: remoteRequestWith(func(request *RemoteCommandRequest) { request.Intent = "merge" }), wantErr: "intent is invalid"},
		{name: "ticket key", request: remoteRequestWith(func(request *RemoteCommandRequest) { request.TicketKey = "" }), wantErr: "ticket key is required"},
		{name: "comment", request: remoteRequestWith(func(request *RemoteCommandRequest) {
			request.Intent = RemoteCommandIntentTicketComment
			request.Comment = ""
		}), wantErr: "ticket comment is required"},
		{name: "time", request: remoteRequestWith(func(request *RemoteCommandRequest) { request.ReceivedAt = time.Time{} }), wantErr: "received time is required"},
	}
}

func remoteParseOptions(text string) RemoteCommandParseOptions {
	return RemoteCommandParseOptions{
		Sender: RemoteCommandSender{
			Channel:    "telegram",
			ExternalID: "sender-1",
		},
		Text:       text,
		ReceivedAt: remoteCommandTestTime(),
	}
}

func remoteRequestWith(edit func(*RemoteCommandRequest)) RemoteCommandRequest {
	request, err := ParseRemoteCommandRequest(remoteParseOptions("ticket start CHR-1"))
	if err != nil {
		panic(err)
	}

	edit(&request)

	return request
}

func assertRemoteStatus(t *testing.T, request RemoteCommandRequest) {
	t.Helper()

	if request.TicketKey != "" || request.Comment != "" || !request.Loop.empty() {
		t.Fatalf("status request has arguments: %#v", request)
	}
}

func assertRemoteLoop(t *testing.T, request RemoteCommandRequest) {
	t.Helper()

	if request.Loop.MaxIterations != 2 {
		t.Fatalf("max iterations = %d, want 2", request.Loop.MaxIterations)
	}

	if strings.Join(request.Loop.Statuses, ",") != "ready" {
		t.Fatalf("statuses = %#v, want ready", request.Loop.Statuses)
	}

	if strings.Join(request.Loop.Priorities, ",") != "high" {
		t.Fatalf("priorities = %#v, want high", request.Loop.Priorities)
	}
}

func assertRemoteTicketStart(t *testing.T, request RemoteCommandRequest) {
	t.Helper()

	if request.TicketKey != "CHR-12" {
		t.Fatalf("ticket key = %q, want CHR-12", request.TicketKey)
	}
}

func assertRemoteTicketComment(t *testing.T, request RemoteCommandRequest) {
	t.Helper()

	if request.TicketKey != "CHR-12" {
		t.Fatalf("ticket key = %q, want CHR-12", request.TicketKey)
	}

	if request.Comment != "blocked on CI logs" {
		t.Fatalf("comment = %q, want expected text", request.Comment)
	}
}

func remoteCommandTestTime() time.Time {
	return time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
}
