package core

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowGitPublicationHook(t *testing.T, fixture workflowGitFixture, mode string) string {
	t.Helper()

	log := filepath.Join(filepath.Dir(fixture.remote), "commands")
	script := "#!/bin/sh\nset -eu\ncat > '" + strings.ReplaceAll(log, "'", "'\\''") + "'\n"

	switch mode {
	case "source":
		script += "unset GIT_QUARANTINE_PATH\ngit update-ref " + fixture.request.SourceRef + " " + fixture.request.Git.BaseCommit + "\n"
	case "target":
		script += "unset GIT_QUARANTINE_PATH\ngit update-ref " + fixture.request.TargetRef + " " + fixture.request.Git.SourceCommit + "\n"
	case "reject":
		script += "exit 1\n"
	}

	if err := os.WriteFile(filepath.Join(fixture.remote, "hooks/pre-receive"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}

	return log
}

// TestWorkflowGitContention runs real receive-pack hooks after advertisement
// across local, SSH and HTTPS transports. Both old/new commands participate;
// policy rejection and external source/target writes reject the entire update.
func TestWorkflowGitContention(t *testing.T) {
	endpoint, config := workflowSSHFixture(t)

	for _, transport := range []string{"local", "ssh", "https"} {
		for _, mode := range []string{"positive", "source", "target", "reject"} {
			t.Run(transport+"/"+mode, func(t *testing.T) {
				checkWorkflowGitContention(t, transport, mode, endpoint, config)
			})
		}
	}
}

func checkWorkflowGitContention(t *testing.T, transport, mode, endpoint, config string) {
	t.Helper()

	fixture := newWorkflowGitFixture(t, "merge")
	log := workflowGitPublicationHook(t, fixture, mode)
	publisher := NewWorkflowGitPublisher(fixture.request.Repository, fixture.remote)

	switch transport {
	case "ssh":
		publisher = NewWorkflowGitPublisher(fixture.request.Repository, endpoint+fixture.remote).WithSSHConfig(config)
	case "https":
		server, _ := workflowHTTPServer(t, fixture, "positive")
		publisher = NewWorkflowGitPublisher(fixture.request.Repository, server.URL+"/project.git").WithHTTPClient(server.Client())
	}

	result, err := publisher.PublishWorkflowGit(t.Context(), fixture.request)
	if mode == "positive" {
		checkWorkflowGitApplied(t, fixture, result, err)
	} else {
		checkWorkflowGitRejected(t, fixture, mode, result, err)
	}

	commands, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}

	expected := fixture.request.Git.SourceCommit + " " + fixture.request.Git.ResultCommit + " " + fixture.request.SourceRef + "\n" + fixture.request.Git.BaseCommit + " " + fixture.request.Git.ResultCommit + " " + fixture.request.TargetRef + "\n"
	if string(commands) != expected {
		t.Fatal("server did not receive both exact guarded update commands")
	}
}

func checkWorkflowGitApplied(t *testing.T, fixture workflowGitFixture, result WorkflowGitPublicationResult, err error) {
	t.Helper()

	if err != nil || result.State != "applied" || !result.CommandsSent || result.PublishedSourceCommit != fixture.request.Git.ResultCommit || result.OriginalSourceCommit != fixture.request.Git.SourceCommit {
		t.Fatalf("exact publication failed: %+v, %v", result, err)
	}

	checkWorkflowGitRefs(t, fixture, fixture.request.Git.ResultCommit, fixture.request.Git.ResultCommit)
}

func checkWorkflowGitRejected(t *testing.T, fixture workflowGitFixture, mode string, result WorkflowGitPublicationResult, err error) {
	t.Helper()

	if !errors.Is(err, workflowstate.ErrConflict) || result.State != "rejected" || !result.CommandsSent || result.PublishedSourceCommit != "" {
		t.Fatalf("atomic rejection lost: %+v, %v", result, err)
	}

	source, target := fixture.request.Git.SourceCommit, fixture.request.Git.BaseCommit

	switch mode {
	case "source":
		source = target
	case "target":
		target = source
	}

	checkWorkflowGitRefs(t, fixture, source, target)
}

// TestWorkflowGitNoopLease demonstrates why a CLI lease on an up-to-date source
// is insufficient: Git omits that command, so a concurrent source write survives
// while the supposedly guarded target update succeeds.
func TestWorkflowGitNoopLease(t *testing.T) {
	fixture := newWorkflowGitFixture(t, "merge")
	request := fixture.request
	workflowIntegrationCommand(t, request.Directory, "push", fixture.remote, request.Git.ResultCommit+":"+request.SourceRef)

	log := workflowGitPublicationHook(t, fixture, "source")
	workflowIntegrationCommand(t, request.Directory, "push", "--atomic", "--force-with-lease="+request.SourceRef+":"+request.Git.ResultCommit, "--force-with-lease="+request.TargetRef+":"+request.Git.BaseCommit, fixture.remote, request.Git.ResultCommit+":"+request.SourceRef, request.Git.ResultCommit+":"+request.TargetRef)

	commands, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(commands), request.SourceRef) || !strings.Contains(string(commands), request.TargetRef) {
		t.Fatal("fixture did not expose CLI no-op filtering")
	}

	checkWorkflowGitRefs(t, fixture, request.Git.BaseCommit, request.Git.ResultCommit)
}
