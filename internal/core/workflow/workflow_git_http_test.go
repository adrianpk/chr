package workflow

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

type workflowHTTPFixture struct {
	remote string
	result string
	mode   string
	posts  atomic.Int64
}

func (fixture *workflowHTTPFixture) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	username, password, ok := request.BasicAuth()
	if !ok || username != "fixture" || password != "synthetic-token" {
		writer.Header().Set("WWW-Authenticate", "Basic realm=fixture")
		writer.WriteHeader(http.StatusUnauthorized)

		return
	}

	if request.URL.Path == "/project.git/info/refs" {
		fixture.advertise(writer, request)

		return
	}

	if request.URL.Path == "/project.git/git-receive-pack" && request.Method == http.MethodPost {
		fixture.receive(writer, request)

		return
	}

	writer.WriteHeader(http.StatusNotFound)
}

func (fixture *workflowHTTPFixture) advertise(writer http.ResponseWriter, request *http.Request) {
	if fixture.mode == "redirect" {
		http.Redirect(writer, request, "/other.git/info/refs", http.StatusTemporaryRedirect)

		return
	}

	if strings.HasPrefix(fixture.mode, "already_") {
		ref := "refs/heads/dev"
		if fixture.mode == "already_source" {
			ref = "refs/heads/chr/workflow/fixture"
		}

		command := workflowPublicationCommand(request.Context(), "git", "-C", fixture.remote, "update-ref", ref, fixture.result)
		if err := command.Run(); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)

			return
		}
	}

	command := workflowPublicationCommand(request.Context(), "git", "receive-pack", "--stateless-rpc", "--advertise-refs", fixture.remote)
	output := &workflowPublicationBuffer{limit: workflowPublicationTextLimit}

	command.Stdout = output
	if err := command.Run(); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)

		return
	}

	writer.Header().Set("Content-Type", "application/x-git-receive-pack-advertisement")

	data := append(workflowPublicationPacket("# service=git-receive-pack\n"), []byte("0000")...)

	data = append(data, output.Bytes()...)
	if _, err := writer.Write(data); err != nil {
		return
	}
}

func (fixture *workflowHTTPFixture) receive(writer http.ResponseWriter, request *http.Request) {
	fixture.posts.Add(1)

	command := workflowPublicationCommand(request.Context(), "git", "receive-pack", "--stateless-rpc", fixture.remote)
	output := &workflowPublicationBuffer{limit: workflowPublicationTextLimit}
	command.Stdout = output

	command.Stdin = io.LimitReader(request.Body, workflowPublicationPackLimit+workflowPublicationTextLimit)
	if err := command.Run(); err != nil {
		writer.WriteHeader(http.StatusInternalServerError)

		return
	}

	writer.Header().Set("Content-Type", "application/x-git-receive-pack-result")

	data := output.Bytes()
	if fixture.mode == "partial" {
		data = data[:len(data)/2]
	}

	if _, err := writer.Write(data); err != nil {
		return
	}
}

func workflowHTTPServer(t *testing.T, fixture workflowGitFixture, mode string) (*httptest.Server, *workflowHTTPFixture) {
	t.Helper()

	handler := &workflowHTTPFixture{remote: fixture.remote, result: fixture.request.Git.ResultCommit, mode: mode}
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)

	credentials := filepath.Join(t.TempDir(), "credentials")

	endpoint, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	endpoint.User = url.UserPassword("fixture", "synthetic-token")
	if err := os.WriteFile(credentials, []byte(endpoint.String()+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	workflowIntegrationCommand(t, fixture.request.Directory, "config", "credential.helper", "")
	workflowIntegrationCommand(t, fixture.request.Directory, "config", "--add", "credential.helper", "store --file="+credentials)

	return server, handler
}

// TestWorkflowGitHTTPS proves smart TLS publication through real receive-pack
// and Git credential lookup; redirects and truncated reports cannot yield success.
func TestWorkflowGitHTTPS(t *testing.T) {
	for _, mode := range []string{"positive", "source", "target", "redirect", "partial", "already_source", "already_target"} {
		t.Run(mode, func(t *testing.T) { checkWorkflowGitHTTPS(t, mode) })
	}
}

func checkWorkflowGitHTTPS(t *testing.T, mode string) {
	t.Helper()

	fixture := newWorkflowGitFixture(t, "merge")

	source, target := fixture.request.Git.SourceCommit, fixture.request.Git.BaseCommit
	if mode == "source" {
		source = target
		workflowIntegrationCommand(t, fixture.remote, "update-ref", fixture.request.SourceRef, source)
	}

	if mode == "target" {
		target = source
		workflowIntegrationCommand(t, fixture.remote, "update-ref", fixture.request.TargetRef, target)
	}

	if strings.HasPrefix(mode, "already_") {
		workflowIntegrationCommand(t, fixture.request.Directory, "push", fixture.remote, fixture.request.Git.ResultCommit+":refs/heads/objects-only")
	}

	server, handler := workflowHTTPServer(t, fixture, mode)
	publisher := NewWorkflowGitPublisher(fixture.request.Repository, server.URL+"/project.git").WithHTTPClient(server.Client())
	result, err := publisher.PublishWorkflowGit(t.Context(), fixture.request)
	checkWorkflowHTTPOutcome(t, mode, result, err, handler.posts.Load())

	if mode == "positive" || mode == "partial" {
		source, target = fixture.request.Git.ResultCommit, fixture.request.Git.ResultCommit
	}

	if mode == "already_source" {
		source = fixture.request.Git.ResultCommit
	}

	if mode == "already_target" {
		target = fixture.request.Git.ResultCommit
	}

	checkWorkflowGitRefs(t, fixture, source, target)
}

func checkWorkflowHTTPOutcome(t *testing.T, mode string, result GitPublicationResult, err error, posts int64) {
	t.Helper()

	expectedState, sent := "not_sent", false

	switch mode {
	case "positive":
		expectedState, sent = "applied", true

		if err != nil {
			t.Fatal(err)
		}
	case "partial":
		expectedState, sent = "unknown", true

		if !errors.Is(err, ErrWorkflowObservationRequired) {
			t.Fatalf("partial response lost uncertainty: %v", err)
		}
	default:
		if err == nil {
			t.Fatal("invalid or stale HTTP exchange succeeded")
		}
	}

	checkWorkflowHTTPStatus(t, result, err, posts, expectedState, sent)
}

func checkWorkflowHTTPStatus(t *testing.T, result GitPublicationResult, err error, posts int64, expectedState string, sent bool) {
	t.Helper()

	if result.State != expectedState || result.CommandsSent != sent || (posts == 1) != sent || posts > 1 {
		t.Fatalf("unexpected HTTP result: %+v, posts=%d, error=%v", result, posts, err)
	}

	if err != nil && strings.Contains(err.Error(), "synthetic-token") {
		t.Fatal("publication error exposed authorization")
	}
}

// TestWorkflowHTTPIdentity rejects credential-bearing and repository-changing
// URLs before contacting a server, keeping canonical identity explicit.
func TestWorkflowHTTPIdentity(t *testing.T) {
	for _, endpoint := range []string{"http://forge.example.invalid/project.git", "https://user:secret@forge.example.invalid/project.git", "https://forge.example.invalid/../project.git", "https://forge.example.invalid/project.git?other=repo", "https://forge.example.invalid/project%2F.git"} {
		remote, err := url.Parse(endpoint)
		if err == nil && validWorkflowHTTPRemote(remote) {
			t.Fatalf("unsafe HTTPS transport accepted: %s", endpoint)
		}
	}
}
