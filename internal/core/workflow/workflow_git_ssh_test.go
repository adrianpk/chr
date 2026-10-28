package workflow

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func workflowSSHCommand(t *testing.T, name string, args ...string) string {
	t.Helper()

	output, err := exec.CommandContext(t.Context(), name, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("fixture command %s failed: %v", name, err)
	}

	return strings.TrimSpace(string(output))
}

func workflowSSHFile(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}

func workflowSSHFixture(t *testing.T) (string, string) {
	t.Helper()

	return workflowSSHFixtureCommand(t, "")
}

func workflowSSHFixtureCommand(t *testing.T, force string) (string, string) {
	t.Helper()

	root := t.TempDir()
	hostKey, clientKey := filepath.Join(root, "host"), filepath.Join(root, "client")
	workflowSSHCommand(t, "ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", hostKey)
	workflowSSHCommand(t, "ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", clientKey)

	clientPublic := workflowSSHCommand(t, "ssh-keygen", "-y", "-f", clientKey)
	workflowSSHFile(t, filepath.Join(root, "authorized_keys"), clientPublic+"\n")

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}

	_, port, err := net.SplitHostPort(address)
	if err != nil {
		t.Fatal(err)
	}

	account, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}

	workflowSSHServer(t, root, hostKey, port, account.Username, force)

	hostPublic := workflowSSHCommand(t, "ssh-keygen", "-y", "-f", hostKey)
	workflowSSHFile(t, filepath.Join(root, "known_hosts"), "[127.0.0.1]:"+port+" "+hostPublic+"\n")

	config := filepath.Join(root, "ssh_config")
	workflowSSHFile(t, config, fmt.Sprintf("Host 127.0.0.1\n  IdentityFile %s\n  IdentitiesOnly yes\n  StrictHostKeyChecking yes\n  UserKnownHostsFile %s\n", clientKey, filepath.Join(root, "known_hosts")))

	return "ssh://" + account.Username + "@" + address, config
}

func workflowSSHServer(t *testing.T, root, hostKey, port, account, force string) {
	t.Helper()

	config := filepath.Join(root, "sshd_config")

	contents := fmt.Sprintf("Port %s\nListenAddress 127.0.0.1\nHostKey %s\nPidFile %s\nAuthorizedKeysFile %s\nStrictModes no\nUsePAM no\nPasswordAuthentication no\nKbdInteractiveAuthentication no\nPubkeyAuthentication yes\nAllowUsers %s\nLogLevel ERROR\n", port, hostKey, filepath.Join(root, "pid"), filepath.Join(root, "authorized_keys"), account)
	if force != "" {
		contents += "ForceCommand " + force + "\n"
	}

	workflowSSHFile(t, config, contents)

	binary, err := exec.LookPath("sshd")
	if err != nil {
		t.Fatal("real SSH fixture requires sshd")
	}

	ctx, cancel := context.WithCancel(t.Context())

	command := workflowPublicationCommand(ctx, binary, "-D", "-e", "-f", config)
	if err := command.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}

	t.Cleanup(func() {
		cancel()

		if err := command.Wait(); err == nil {
			t.Error("fixture SSH server exited without owned cancellation")
		}
	})

	workflowSSHReady(t, "127.0.0.1:"+port)
}

func workflowSSHReady(t *testing.T, address string) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		connection, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
		if err == nil {
			if err := connection.Close(); err != nil {
				t.Fatal(err)
			}

			return
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("owned SSH fixture did not become ready")
}

// TestWorkflowGitSSH proves real SSH authentication and receive-pack publication,
// including a changed advertisement that must preserve both remote refs.
func TestWorkflowGitSSH(t *testing.T) {
	endpoint, config := workflowSSHFixture(t)
	for _, stale := range []bool{false, true} {
		t.Run(fmt.Sprintf("stale=%t", stale), func(t *testing.T) {
			fixture := newWorkflowGitFixture(t, "merge")
			transport := endpoint + fixture.remote
			publisher := NewWorkflowGitPublisher(fixture.request.Repository, transport).WithSSHConfig(config)

			source, target := fixture.request.Git.SourceCommit, fixture.request.Git.BaseCommit
			if stale {
				source = target
				workflowIntegrationCommand(t, fixture.remote, "update-ref", fixture.request.SourceRef, source)
			}

			result, err := publisher.PublishWorkflowGit(t.Context(), fixture.request)
			if stale {
				checkWorkflowGitStale(t, result, err)
				checkWorkflowGitRefs(t, fixture, source, target)

				return
			}

			if err != nil || result.State != "applied" {
				t.Fatalf("publication: %+v, %v", result, err)
			}

			checkWorkflowGitRefs(t, fixture, fixture.request.Git.ResultCommit, fixture.request.Git.ResultCommit)
		})
	}
}

func checkWorkflowGitStale(t *testing.T, result workflowstate.GitPublicationResult, err error) {
	t.Helper()

	if !errors.Is(err, workflowstate.ErrConflict) || result.CommandsSent {
		t.Fatalf("stale publication: %+v, %v", result, err)
	}
}

// TestWorkflowSSHArguments refuses option and credential injection, and keeps
// shell metacharacters in a repository path inside one quoted remote argument.
func TestWorkflowSSHArguments(t *testing.T) {
	for _, transport := range []string{"ssh://-option/project.git", "ssh://user:secret@forge.example.invalid/project.git", "ssh://forge.example.invalid:0/project.git", "ssh://forge.example.invalid/project.git?change=repo", "ssh://forge.example.invalid/project%0A.git"} {
		publisher := NewWorkflowGitPublisher("", transport)
		if _, err := publisher.publicationSSHArgs(); err == nil {
			t.Fatalf("unsafe SSH transport accepted: %s", transport)
		}
	}

	remote := &url.URL{Scheme: "ssh", Host: "forge.example.invalid", Path: "/repo'$(touch injected).git"}
	publisher := NewWorkflowGitPublisher("", remote.String())

	args, err := publisher.publicationSSHArgs()
	if err != nil || args[len(args)-1] != "git-receive-pack '/repo'\\''$(touch injected).git'" {
		t.Fatalf("SSH path was not quoted: %v, %v", args, err)
	}
}
