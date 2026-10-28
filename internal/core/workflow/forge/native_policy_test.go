package forge

import (
	"encoding/json"
	"testing"
)

// TestNativeMergePolicy uses each provider's actual repository field and keeps
// explicit denial or missing policy authoritative over the other API's field.
func TestNativeMergePolicy(t *testing.T) {
	cases := []struct {
		name     string
		contract string
		policy   string
		allowed  bool
	}{
		{"forgejo_enabled", "forgejo_same_pr_git_auto_merge/v1", `"allow_merge_commits":true`, true},
		{"forgejo_disabled", "forgejo_same_pr_git_auto_merge/v1", `"allow_merge_commits":false`, false},
		{"forgejo_wrong_field", "forgejo_same_pr_git_auto_merge/v1", `"allow_merge_commit":true`, false},
		{"forgejo_conflict", "forgejo_same_pr_git_auto_merge/v1", `"allow_merge_commits":false,"allow_merge_commit":true`, false},
		{"forgejo_missing", "forgejo_same_pr_git_auto_merge/v1", `"archived":false`, false},
		{"github_enabled", "github_same_pr_git_merge/v1", `"allow_merge_commit":true`, true},
		{"github_disabled", "github_same_pr_git_merge/v1", `"allow_merge_commit":false`, false},
		{"github_wrong_field", "github_same_pr_git_merge/v1", `"allow_merge_commits":true`, false},
		{"github_conflict", "github_same_pr_git_merge/v1", `"allow_merge_commit":false,"allow_merge_commits":true`, false},
		{"github_missing", "github_same_pr_git_merge/v1", `"archived":false`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			responses := []json.RawMessage{
				json.RawMessage(`{"permissions":{"push":true},` + tc.policy + `}`),
				json.RawMessage(`{"protected":false}`), json.RawMessage(`{"protected":false}`),
			}

			observation, err := nativeRemotePolicy(responses, "merge", tc.contract)
			if err != nil || (len(observation.Unsupported) == 0) != tc.allowed {
				t.Fatalf("merge policy allowed=%t, want %t; reasons=%+v, error=%v", len(observation.Unsupported) == 0, tc.allowed, observation.Unsupported, err)
			}
		})
	}
}
