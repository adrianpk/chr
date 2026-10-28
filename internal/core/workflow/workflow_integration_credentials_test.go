package workflow

import (
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// Isolated object stores must retain the source's credential-helper context.
// Both preparation and later verification fetch the actual authenticated target.
func TestWorkflowIntegrationCredentials(t *testing.T) {
	f := newWorkflowGitFixture(t, "rebase")
	backend := &cgi.Handler{Path: filepath.Join(workflowIntegrationCommand(t, f.request.Directory, "--exec-path"), "git-http-backend"), Env: []string{"GIT_PROJECT_ROOT=" + filepath.Dir(f.remote), "GIT_HTTP_EXPORT_ALL=1"}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "fixture" || password != "synthetic-token" {
			w.Header().Set("WWW-Authenticate", "Basic realm=fixture")
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		backend.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)
	workflowPreparationCredential(t, f, server.URL)

	query := PreparationQuery{IntegrationQuery: f.request.Git.IntegrationQuery, BaseCommit: f.request.Git.BaseCommit, Lane: workflowstate.Lane{Unit: workflowstate.Ref{Kind: "ticket", ID: "TST-1"}, Branch: "refs/heads/chr/workflow/credential-preparation", Worktree: filepath.Join(t.TempDir(), "prepared"), BaseCommit: f.request.Git.BaseCommit, HeadCommit: f.request.Git.BaseCommit}}
	query.SourceDirectory, query.ResultCommit = f.request.Directory, ""
	transport := server.URL + "/" + filepath.Base(f.remote)

	proof, err := NewWorkflowPreparationGit(f.request.Repository, transport).PrepareWorkflowIntegration(t.Context(), query)
	if err != nil {
		t.Fatal("authenticated preparation lost the source credential context", err)
	}

	if proof.PreparedTree != f.request.Git.PreparedTree {
		t.Fatal("authenticated preparation changed the combined content")
	}

	workflowIntegrationCommand(t, f.request.Directory, "push", f.remote, proof.ResultCommit+":refs/heads/dev")

	observation := proof.IntegrationQuery
	observation.SourceDirectory = f.request.Directory

	verified, err := NewWorkflowIntegrationGit(f.request.Repository, transport).VerifyPreparedWorkflowIntegration(t.Context(), observation, proof)
	if err != nil || verified.VerifiedTree != proof.PreparedTree {
		t.Fatal("authenticated verification lost the source credential context", err)
	}
}

func workflowPreparationCredential(t *testing.T, f workflowGitFixture, endpoint string) {
	t.Helper()

	location, err := url.Parse(endpoint)
	if err != nil {
		t.Fatal(err)
	}

	location.User = url.UserPassword("fixture", "synthetic-token")

	path := filepath.Join(t.TempDir(), "credentials")
	if err := os.WriteFile(path, []byte(location.String()+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	workflowIntegrationCommand(t, f.request.Directory, "config", "credential.helper", "")
	workflowIntegrationCommand(t, f.request.Directory, "config", "--add", "credential.helper", "store --file="+path)
}
