package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	authoringcore "forge.adrianpk.com/chr/internal/core/authoring"
	workspacecore "forge.adrianpk.com/chr/internal/core/workspace"
)

// TestGuidedUseCaseAdmissionPrerequisites blocks real setup failures before native admission or host dispatch.
func TestGuidedUseCaseAdmissionPrerequisites(t *testing.T) {
	t.Parallel()

	cases := []struct{ name, action string }{{"host", "Start a managed workflow host, then retry this retained operation."}, {"worker", "available worker harness"}, {"provider", "canonical review provider"}, {"publisher", "private certified publisher"}, {"contract", "native"}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			operations, prepared, host := guidedAdmissionFixture(t)
			breakGuidedPrerequisite(t, &operations, prepared, tc.name)

			blocked, err := operations.AdmitGuidedWork(t.Context(), guidedQuizCommand(prepared, "admit", "admit", ""))
			if err == nil || blocked.Guided.Work.Invocation != nil || len(host.ids) != 0 || !strings.Contains(blocked.Guided.Blocker, tc.action) {
				t.Fatal("prerequisite failure claimed execution or lost its corrective action", err, blocked.Guided.Blocker)
			}

			page, err := operations.WorkflowStatusPage(t.Context(), 0, 8)
			if err != nil || len(page.Items) != 0 {
				t.Fatal("blocked prerequisite allocated native work", err)
			}
		})
	}
}

func breakGuidedPrerequisite(t *testing.T, operations *service, prepared authoringcore.Session, name string) {
	t.Helper()

	config := "workflow:\n  version: 1\nreview:\n  mode: human_review\nwork:\n  state_authority: tickets\n"

	switch name {
	case "host":
		operations.guidedWorkHost = nil
	case "worker":
		operations.workflowWorker = nil
		config += "harnesses:\n  worker: unavailable\n"
	case "provider":
		operations.workflowPullRequests = nil
		config = "workflow:\n  version: 1\nwork:\n  state_authority: tickets\nreview:\n  mode: human_review\n  forgejo_enabled: false\n"
	case "publisher":
		config = "workflow:\n  version: 1\nreview:\n  mode: human_review\nwork:\n  state_authority: tickets\n  integration:\n    publication_profile: git_atomic_dual_ref\n"
	case "contract":
		member := prepared.Guided.Work.Materialization.Members[0]

		path := filepath.Join(operations.projectContext.ProjectDir, "ops", prepared.Namespace, filepath.FromSlash(member.Path))
		if err := os.WriteFile(path, []byte("Invalid replacement contract"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, config)
}

// TestGuidedUseCaseAdmissionPolicy preserves a displayed policy after configuration changes instead of silently refreshing authorization.
func TestGuidedUseCaseAdmissionPolicy(t *testing.T) {
	t.Parallel()

	operations, prepared, host := guidedAdmissionFixture(t)
	path := workspacecore.UserPath(operations.projectContext.ProjectDir, "config.yml")

	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	operations.afterGuidedAdmission = func(phase string) error {
		if phase == "preview" {
			writeCLIWorkspaceConfig(t, operations.projectContext.ProjectDir, "workflow:\n  version: 1\nreview:\n  mode: human_review\nwork:\n  state_authority: tickets\n  max_worker_attempts: 1\n")
		}

		return nil
	}
	command := guidedQuizCommand(prepared, "admit", "admit", "")

	blocked, err := operations.AdmitGuidedWork(t.Context(), command)
	if err == nil || blocked.Guided.Work.Admission == nil || blocked.Guided.Work.Invocation != nil || len(host.ids) != 0 {
		t.Fatal("changed policy was silently accepted", err)
	}

	operations.afterGuidedAdmission = nil
	if _, err := operations.AdmitGuidedWork(t.Context(), command); err == nil {
		t.Fatal("retry replaced the retained policy")
	}

	restoreGuidedAdmissionPolicy(t, operations, blocked, original)

	if len(host.ids) != 1 {
		t.Fatal("restored policy did not activate exactly once")
	}
}

func restoreGuidedAdmissionPolicy(t *testing.T, operations service, blocked authoringcore.Session, original []byte) {
	t.Helper()

	path := workspacecore.UserPath(operations.projectContext.ProjectDir, "config.yml")
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}

	command, ready := guidedProgressCommand(blocked)
	if !ready {
		t.Fatal("pending admission was not resumable")
	}

	recovered, err := operations.AdmitGuidedWork(t.Context(), command)
	if err != nil || recovered.Guided.Work.Admission.RequestKey != blocked.Guided.Work.Admission.RequestKey {
		t.Fatal("restored prerequisite did not reuse original admission", err)
	}
}
