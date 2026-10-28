package core

import (
	"context"
	"errors"
	"strings"
	"testing"

	workflowcore "forge.adrianpk.com/chr/internal/core/workflow"
	workflowforge "forge.adrianpk.com/chr/internal/core/workflow/forge"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

func TestRemoteReviewAdapterReadsForgejoSnapshot(t *testing.T) {
	runner := newFakeRemoteReviewRunner()
	runner.outputs["tea api /repos/example/repo/pulls/12"] = []byte(`{"head":{"sha":"abc123"}}`)
	runner.outputs["tea api /repos/example/repo/pulls/12.diff"] = []byte("diff --git a/file.go b/file.go")
	runner.outputs["tea api /repos/example/repo/issues/12/comments"] = []byte(`[{"body":"comment"}]`)
	runner.outputs["tea api /repos/example/repo/commits/abc123/statuses"] = []byte(`[{"state":"success"}]`)

	adapter := workflowcore.NewRemoteReviewAdapter(workspaceconfig.DefaultReviewProviderConfig(), runner)

	snapshot, err := adapter.Read(t.Context(), "https://codeberg.org/example/repo/pulls/12")
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}

	if snapshot.Reference.Provider != workflowforge.Forgejo {
		t.Fatalf("Provider = %q, want %q", snapshot.Reference.Provider, workflowforge.Forgejo)
	}

	if !strings.Contains(snapshot.Diff, "diff --git") {
		t.Fatalf("Diff = %q", snapshot.Diff)
	}

	if !snapshot.Capabilities.CanReadChecks {
		t.Fatal("CanReadChecks = false, want true")
	}

	if !snapshot.Capabilities.CanWriteInline || !snapshot.Capabilities.CanWriteGeneral {
		t.Fatalf("Capabilities = %#v, want inline and general writing", snapshot.Capabilities)
	}
}

func TestRemoteReviewAdapterReadsGitHubSnapshot(t *testing.T) {
	runner := newFakeRemoteReviewRunner()
	runner.outputs["gh pr view https://github.com/example/repo/pull/34 --json number,title,state,baseRefName,headRefName,headRefOid,mergeable,isDraft,reviewDecision"] = []byte(`{"number":34}`)
	runner.outputs["gh pr diff https://github.com/example/repo/pull/34"] = []byte("diff --git a/file.go b/file.go")
	runner.outputs["gh api repos/example/repo/issues/34/comments"] = []byte(`[{"body":"issue"}]`)
	runner.outputs["gh api repos/example/repo/pulls/34/comments"] = []byte(`[{"body":"review"}]`)
	runner.outputs["gh pr checks https://github.com/example/repo/pull/34 --json name,state,link,bucket"] = []byte(`[{"name":"test","state":"SUCCESS"}]`)

	adapter := workflowcore.NewRemoteReviewAdapter(workspaceconfig.DefaultReviewProviderConfig(), runner)

	snapshot, err := adapter.Read(t.Context(), "https://github.com/example/repo/pull/34")
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}

	if snapshot.Reference.Provider != workflowforge.GitHub {
		t.Fatalf("Provider = %q, want %q", snapshot.Reference.Provider, workflowforge.GitHub)
	}

	if !strings.Contains(snapshot.ReviewCommentsJSON, "review") {
		t.Fatalf("ReviewCommentsJSON = %q", snapshot.ReviewCommentsJSON)
	}

	if !snapshot.Capabilities.CanReadChecks {
		t.Fatal("CanReadChecks = false, want true")
	}
}

func TestRemoteReviewAdapterWritesInlineComment(t *testing.T) {
	runner := newFakeRemoteReviewRunner()
	adapter := workflowcore.NewRemoteReviewAdapter(workspaceconfig.DefaultReviewProviderConfig(), runner)

	result, err := adapter.WriteComments(t.Context(), "https://github.com/example/repo/pull/34", []workflowcore.RemoteReviewComment{
		{Body: "Handle nil input.", Path: "internal/example.go", Line: 42, Blocking: true},
	})
	if err != nil {
		t.Fatalf("WriteComments returned error: %v", err)
	}

	if len(result.Comments) != 1 {
		t.Fatalf("Comments = %d, want 1", len(result.Comments))
	}

	written := result.Comments[0]
	if written.Placement != workflowcore.RemoteReviewCommentPlacementInline || written.Kind != workflowcore.RemoteReviewCommentKindBlocking {
		t.Fatalf("written comment = %#v", written)
	}

	command := "gh api -X POST repos/example/repo/pulls/34/comments --input -"
	input := string(runner.inputs[command])

	for _, want := range []string{
		`"body":"Handle nil input."`,
		`"path":"internal/example.go"`,
		`"line":42`,
		`"side":"RIGHT"`,
	} {
		if !strings.Contains(input, want) {
			t.Fatalf("inline payload does not contain %q:\n%s", want, input)
		}
	}
}

func TestRemoteReviewAdapterFallsBackToGeneralComment(t *testing.T) {
	runner := newFakeRemoteReviewRunner()
	runner.inputErrors["tea api -X POST /repos/example/repo/pulls/12/reviews -d @-"] = errors.New("inline unsupported")

	adapter := workflowcore.NewRemoteReviewAdapter(workspaceconfig.DefaultReviewProviderConfig(), runner)

	result, err := adapter.WriteComments(t.Context(), "https://codeberg.org/example/repo/pulls/12", []workflowcore.RemoteReviewComment{
		{Body: "Handle nil input.", Path: "internal/example.go", Line: 42},
	})
	if err != nil {
		t.Fatalf("WriteComments returned error: %v", err)
	}

	written := result.Comments[0]
	if written.Placement != workflowcore.RemoteReviewCommentPlacementGeneral || written.Reason != "inline_unavailable" {
		t.Fatalf("written comment = %#v", written)
	}

	command := "tea api -X POST /repos/example/repo/issues/12/comments -d @-"
	input := string(runner.inputs[command])

	for _, want := range []string{
		"Handle nil input.",
		"Location: `internal/example.go:42`",
	} {
		if !strings.Contains(input, want) {
			t.Fatalf("general payload does not contain %q:\n%s", want, input)
		}
	}
}

func TestRemoteReviewAdapterRejectsDisabledProvider(t *testing.T) {
	config := workspaceconfig.DefaultReviewProviderConfig()
	config.GitHub.Enabled = false

	adapter := workflowcore.NewRemoteReviewAdapter(config, newFakeRemoteReviewRunner())

	_, err := adapter.Read(t.Context(), "https://github.com/example/repo/pull/34")
	if err == nil {
		t.Fatal("Read returned nil error")
	}

	if !strings.Contains(err.Error(), "review provider is not enabled: github") {
		t.Fatalf("error = %q", err.Error())
	}
}

type fakeRemoteReviewRunner struct {
	outputs     map[string][]byte
	outputErrs  map[string]error
	inputs      map[string][]byte
	inputErrors map[string]error
}

func newFakeRemoteReviewRunner() *fakeRemoteReviewRunner {
	return &fakeRemoteReviewRunner{
		outputs:     map[string][]byte{},
		outputErrs:  map[string]error{},
		inputs:      map[string][]byte{},
		inputErrors: map[string]error{},
	}
}

func (runner *fakeRemoteReviewRunner) Output(_ context.Context, command string, args ...string) ([]byte, error) {
	key := remoteReviewCommandKey(command, args...)
	if err := runner.outputErrs[key]; err != nil {
		return nil, err
	}

	output, ok := runner.outputs[key]
	if !ok {
		return nil, errors.New("missing output for " + key)
	}

	return output, nil
}

func (runner *fakeRemoteReviewRunner) Input(
	_ context.Context,
	command string,
	input []byte,
	args ...string,
) ([]byte, error) {
	key := remoteReviewCommandKey(command, args...)
	if err := runner.inputErrors[key]; err != nil {
		return nil, err
	}

	runner.inputs[key] = input

	return []byte(`{}`), nil
}

func remoteReviewCommandKey(command string, args ...string) string {
	return strings.Join(append([]string{command}, args...), " ")
}
