package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
	workspaceconfig "forge.adrianpk.com/chr/internal/core/workspace/config"
)

type workflowPRCommandFixture struct {
	*fakeRemoteReviewRunner
	response []byte
	calls    []string
}

func (r *workflowPRCommandFixture) Output(ctx context.Context, command string, args ...string) ([]byte, error) {
	r.calls = append(r.calls, remoteReviewCommandKey(command, args...))

	return r.fakeRemoteReviewRunner.Output(ctx, command, args...)
}

func (r *workflowPRCommandFixture) Input(_ context.Context, command string, data []byte, args ...string) ([]byte, error) {
	key := remoteReviewCommandKey(command, args...)
	r.calls = append(r.calls, key)
	r.inputs[key] = data

	return r.response, nil
}

func workflowRemotePRBytes(t *testing.T, repository, marker string) []byte {
	t.Helper()

	pr := workflowPRWireFixture{URL: repository + marker + "7", State: "open"}
	pr.Head.Ref, pr.Head.SHA, pr.Head.Repo.URL = "chr/workflow/branch", strings.Repeat("a", 40), repository
	pr.Base.Ref, pr.Base.SHA, pr.Base.Repo.URL = "dev", strings.Repeat("b", 40), repository

	data, err := json.Marshal(pr)
	if err != nil {
		t.Fatal(err)
	}

	return data
}

// TestWorkflowPRAdapter preserves provider facts, binds API host and repository,
// submits structured content once and scans all states through the empty page.
func TestWorkflowPRAdapter(t *testing.T) {
	cases := []struct{ name, repository, marker, cli, prefix, input string }{
		{"forgejo", "https://forge.example:8443/test/repo", "/pulls/", "tea", "tea api -X GET https://forge.example:8443/api/v1/repos/test/repo", "tea api -X POST https://forge.example:8443/api/v1/repos/test/repo/pulls -d @-"},
		{"github", "https://github.com/test/repo", "/pull/", "gh", "gh api --hostname github.com -X GET repos/test/repo", "gh api --hostname github.com -X POST repos/test/repo/pulls --input -"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner := &workflowPRCommandFixture{fakeRemoteReviewRunner: newFakeRemoteReviewRunner(), response: workflowRemotePRBytes(t, tc.repository, tc.marker)}
			runner.outputs[tc.prefix+"/pulls?state=all&limit=50&per_page=50&page=1"] = []byte("[" + string(runner.response) + "]")
			runner.outputs[tc.prefix+"/pulls?state=all&limit=50&per_page=50&page=2"] = []byte("[]")
			runner.outputs[tc.prefix+"/pulls/7"] = runner.response

			commitField := "sha"
			if tc.name == "forgejo" {
				commitField = "id"
			}

			runner.outputs[tc.prefix+"/branches/chr%2Fworkflow%2Fbranch"] = []byte(`{"commit":{"` + commitField + `":"` + strings.Repeat("a", 40) + `"}}`)
			checkWorkflowPRAdapter(t, NewRemoteReviewAdapter(workspaceconfig.DefaultReviewProviderConfig(), runner), runner, tc.repository, tc.input)
		})
	}
}

// TestWorkflowPRBranch accepts the distinct Forgejo/GitHub branch commit fields
// and rejects absent, short or contradictory canonical source observations.
func TestWorkflowPRBranch(t *testing.T) {
	commit := strings.Repeat("a", 40)

	cases := []struct {
		name, data string
		valid      bool
	}{
		{"forgejo ID", `{"commit":{"id":"` + commit + `"}}`, true},
		{"github SHA", `{"commit":{"sha":"` + commit + `"}}`, true},
		{"missing commit", `{}`, false},
		{"short commit", `{"commit":{"id":"abc"}}`, false},
		{"conflicting fields", `{"commit":{"id":"` + commit + `","sha":"` + strings.Repeat("b", 40) + `"}}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runner := newFakeRemoteReviewRunner()
			runner.outputs["tea api -X GET https://forge.example/api/v1/repos/test/repo/branches/dev"] = []byte(tc.data)

			adapter := NewRemoteReviewAdapter(workspaceconfig.DefaultReviewProviderConfig(), runner)

			observed, err := adapter.ReadWorkflowBranch(t.Context(), "https://forge.example/test/repo", "dev")
			if (err == nil) != tc.valid {
				t.Fatalf("canonical branch observation: %s, %v", observed, err)
			}

			if tc.valid && observed != commit {
				t.Fatal("branch parser changed the full commit")
			}
		})
	}
}

func checkWorkflowPRAdapter(t *testing.T, adapter RemoteReviewAdapter, runner *workflowPRCommandFixture, repository, inputKey string) {
	t.Helper()

	request := workflowstate.PullRequestRequest{PullRequestQuery: workflowstate.PullRequestQuery{Repository: repository, HeadBranch: "chr/workflow/branch", BaseBranch: "dev"}, Title: "Deliver slice", Body: "A literal body\nwith `code` and $(text)."}

	items, err := adapter.FindWorkflowPullRequests(t.Context(), request.PullRequestQuery)
	if err != nil || len(items) != 1 {
		t.Fatalf("provider discovery: %+v, %v", items, err)
	}

	pr, err := adapter.ReadWorkflowPullRequest(t.Context(), items[0].URL)
	if err != nil || pr != items[0] {
		t.Fatal("provider read changed canonical facts")
	}

	commit, err := adapter.ReadWorkflowBranch(t.Context(), repository, request.HeadBranch)
	if err != nil || commit != pr.HeadCommit {
		t.Fatal("source branch was not observed")
	}

	created, err := adapter.CreateWorkflowPullRequest(t.Context(), request)
	if err != nil || created != pr {
		t.Fatal("structured PR creation did not retain provider result")
	}

	checkWorkflowPRPayload(t, runner, inputKey, request)
}

func checkWorkflowPRPayload(t *testing.T, runner *workflowPRCommandFixture, inputKey string, request workflowstate.PullRequestRequest) {
	t.Helper()

	var payload map[string]string
	if err := json.Unmarshal(runner.inputs[inputKey], &payload); err != nil || payload["body"] != request.Body || payload["head"] != request.HeadBranch || len(runner.calls) != 5 {
		t.Fatal("PR payload, method or bounded paging changed")
	}
}

// TestWorkflowPRAdapterBounds rejects null/malformed/oversized pages and retains
// a nonempty short page as incomplete until later pages have been observed.
func TestWorkflowPRAdapterBounds(t *testing.T) {
	for _, tc := range []struct {
		name string
		data []byte
	}{{"null", []byte("null")}, {"malformed", []byte("{")}, {"oversized", []byte(strings.Repeat("x", workflowstate.FileLimit+1))}} {
		t.Run(tc.name, func(t *testing.T) {
			runner := newFakeRemoteReviewRunner()
			runner.outputs["tea api -X GET https://forge.example/api/v1/repos/test/repo/pulls?state=all&limit=50&per_page=50&page=1"] = tc.data

			adapter := NewRemoteReviewAdapter(workspaceconfig.DefaultReviewProviderConfig(), runner)
			if _, err := adapter.FindWorkflowPullRequests(t.Context(), workflowstate.PullRequestQuery{Repository: "https://forge.example/test/repo", HeadBranch: "chr/workflow/branch", BaseBranch: "dev"}); err == nil {
				t.Fatal("invalid provider page accepted")
			}
		})
	}

	runner := &workflowPRCommandFixture{fakeRemoteReviewRunner: newFakeRemoteReviewRunner()}
	for page := 1; page <= workflowstate.PageLimit; page++ {
		runner.outputs[fmt.Sprintf("tea api -X GET https://forge.example/api/v1/repos/test/repo/pulls?state=all&limit=50&per_page=50&page=%d", page)] = []byte(`[{"head":{"ref":"unrelated"}}]`)
	}

	adapter := NewRemoteReviewAdapter(workspaceconfig.DefaultReviewProviderConfig(), runner)
	if _, err := adapter.FindWorkflowPullRequests(t.Context(), workflowstate.PullRequestQuery{Repository: "https://forge.example/test/repo", HeadBranch: "chr/workflow/branch", BaseBranch: "dev"}); err == nil {
		t.Fatal("incomplete short-page pagination licensed replacement PR creation")
	}
}

type workflowPRWireFixture struct {
	URL          string                      `json:"html_url"`
	State        string                      `json:"state"`
	Merged       bool                        `json:"merged"`
	ResultCommit string                      `json:"merge_commit_sha"`
	Head         workflowPRWireBranchFixture `json:"head"`
	Base         workflowPRWireBranchFixture `json:"base"`
}

type workflowPRWireBranchFixture struct {
	Ref  string `json:"ref"`
	SHA  string `json:"sha"`
	Repo struct {
		URL string `json:"html_url"`
	} `json:"repo"`
}
