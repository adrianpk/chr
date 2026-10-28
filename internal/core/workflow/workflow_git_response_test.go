package workflow

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowGitReceiveProxy is the owned real-process fault fixture invoked
// by sshd. It forwards Git's advertisement and commands, then truncates the
// report only after the real receive-pack has successfully updated both refs.
func TestWorkflowGitReceiveProxy(t *testing.T) {
	remote := os.Getenv("CHR_NATIVE_LOST_REPORT")
	if remote == "" {
		return
	}

	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	command := workflowPublicationCommand(ctx, "git", "receive-pack", "--", remote)

	session, err := startWorkflowReceivePack(command, cancel)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := session.Close(); err != nil {
			t.Error(err)
		}
	}()

	workflowForwardGitAdvertisement(t, session.Advertisement())

	commands, err := io.ReadAll(io.LimitReader(os.Stdin, workflowPublicationPackLimit+workflowPublicationTextLimit))
	if err != nil {
		t.Fatal(err)
	}

	report, err := session.Send(commands)
	if err != nil {
		t.Fatal(err)
	}

	data, err := io.ReadAll(io.LimitReader(report, workflowPublicationTextLimit))
	if err != nil {
		t.Fatal(err)
	}

	if err := session.Finish(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stdout.Write(data[:len(data)/2]); err != nil {
		t.Fatal(err)
	}
}

func workflowForwardGitAdvertisement(t *testing.T, reader io.Reader) {
	t.Helper()

	for {
		packet, err := readWorkflowPublicationPacket(reader)
		if err != nil {
			t.Fatal(err)
		}

		if packet == nil {
			if _, err := os.Stdout.WriteString("0000"); err != nil {
				t.Fatal(err)
			}

			return
		}

		if _, err := os.Stdout.Write(workflowPublicationPacket(string(packet))); err != nil {
			t.Fatal(err)
		}
	}
}

// TestWorkflowGitSSHUnknown proves that loss of a real SSH publication report
// preserves unknown outcome even when both remote refs have already become P.
func TestWorkflowGitSSHUnknown(t *testing.T) {
	fixture := newWorkflowGitFixture(t, "merge")

	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	quote := func(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'" }
	force := "env CHR_NATIVE_LOST_REPORT=" + quote(fixture.remote) + " " + quote(executable) + " -test.run=^TestWorkflowGitReceiveProxy$"
	endpoint, config := workflowSSHFixtureCommand(t, force)
	publisher := NewWorkflowGitPublisher(fixture.request.Repository, endpoint+fixture.remote).WithSSHConfig(config)

	result, err := publisher.PublishWorkflowGit(t.Context(), fixture.request)
	if !errors.Is(err, ErrWorkflowObservationRequired) || result.State != workflowstate.GitPublicationUnknown || !result.CommandsSent || result.PublishedSourceCommit != "" {
		t.Fatalf("lost SSH response was not retained: %+v, %v", result, err)
	}

	checkWorkflowGitRefs(t, fixture, fixture.request.Git.ResultCommit, fixture.request.Git.ResultCommit)
}
