package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opscore "forge.adrianpk.com/chr/internal/core/ops"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

// TestWorkflowUpgradeRecovery checks each durable policy boundary. Ordinary
// reads fail closed until the same explicit request finishes the pending work.
func TestWorkflowUpgradeRecovery(t *testing.T) {
	for _, stage := range []int{0, 1, 2} {
		t.Run(fmt.Sprintf("boundary %d", stage), func(t *testing.T) {
			store, opts := workflowUpgradeFixture(t, "")

			store.afterReplace = func(index int) error {
				if index == stage {
					return fmt.Errorf("owned write fault")
				}

				return nil
			}
			if _, err := store.UpgradeWorkflowConfiguration(t.Context(), opts); err == nil {
				t.Fatal("owned fault not reached")
			}

			checkWorkflowUpgradePending(t, store, opts)

			store.afterReplace = nil

			result, err := store.UpgradeWorkflowConfiguration(t.Context(), opts)
			if err != nil {
				t.Fatal(err)
			}

			checkWorkflowUpgradeSnapshot(t, store, opts, result)
		})
	}
}

func checkWorkflowUpgradePending(t *testing.T, store *WorkflowStore, opts WorkflowUpgradeOptions) {
	t.Helper()

	if _, err := LoadWorkspaceConfig(store.layout.ProjectDir()); err == nil {
		t.Fatal("partial upgrade exposed policy")
	}

	if _, err := store.workflowReviewPolicy(workflowInvocationFile{}); err == nil {
		t.Fatal("partial upgrade exposed publication policy")
	}

	checkWorkflowUpgradePendingDomain(t, store)

	changed := opts

	changed.ReviewMode = ReviewWorkflowModeTrustedAutoMerge
	if _, err := store.UpgradeWorkflowConfiguration(t.Context(), changed); err == nil {
		t.Fatal("foreign request resumed partial upgrade")
	}
}

func checkWorkflowUpgradePendingDomain(t *testing.T, store *WorkflowStore) {
	t.Helper()

	if _, err := store.AdoptPlanWorkflow(t.Context(), WorkflowPlanAdoptionOptions{}); err == nil || !strings.Contains(err.Error(), "upgrade pending") {
		t.Fatal("ordinary adoption ignored pending upgrade", err)
	}

	if _, err := store.MaterializePlan(t.Context(), WorkflowMaterializationOptions{}); err == nil || !strings.Contains(err.Error(), "upgrade pending") {
		t.Fatal("ordinary materialization ignored pending upgrade", err)
	}
}

// TestWorkflowUpgradeAdoptionRecovery resumes a torn selected-record transition
// before enabling policy, preserving identities and excluding a foreign retry.
func TestWorkflowUpgradeAdoptionRecovery(t *testing.T) {
	store, selection := workflowPlanAdoptionFixture(t)
	opts := WorkflowUpgradeOptions{ExpectedRevision: contenthash.Sum(nil), StateAuthority: WorkStateAuthorityTickets, Adoption: &selection, CheckQuiescent: func(context.Context) error { return nil }}
	writes := 0

	store.afterReplace = func(int) error {
		writes++
		if writes == 2 {
			return fmt.Errorf("owned adoption write fault")
		}

		return nil
	}
	if _, err := store.UpgradeWorkflowConfiguration(t.Context(), opts); err == nil {
		t.Fatal("owned adoption fault not reached")
	}

	checkWorkflowUpgradePending(t, store, opts)

	store.afterReplace = nil

	result, err := store.UpgradeWorkflowConfiguration(t.Context(), opts)
	if err != nil || result.Adopted == nil {
		t.Fatal("selected upgrade did not recover", err)
	}

	checkAdoptedWorkflowTasks(t, store, *result.Adopted, selection)
	checkWorkflowUpgradeSnapshot(t, store, opts, result)
}

// TestWorkflowUpgradeEdited rejects manual config and journal changes while
// preserving both images for reconciliation rather than overwriting them.
func TestWorkflowUpgradeEdited(t *testing.T) {
	for _, target := range []string{"config.yml", workflowUpgradePending} {
		t.Run(target, func(t *testing.T) {
			store, opts := workflowUpgradeFixture(t, "")

			store.afterReplace = func(int) error { return fmt.Errorf("owned write fault") }
			if _, err := store.UpgradeWorkflowConfiguration(t.Context(), opts); err == nil {
				t.Fatal("owned fault not reached")
			}

			path := UserPath(store.layout.ProjectDir(), target)

			image := []byte("# Manual policy edit\n")
			if err := os.WriteFile(path, image, 0o600); err != nil {
				t.Fatal(err)
			}

			store.afterReplace = nil
			if _, err := store.UpgradeWorkflowConfiguration(t.Context(), opts); err == nil {
				t.Fatal("manual edit overwritten")
			}

			observed, err := os.ReadFile(path)
			if err != nil || string(observed) != string(image) {
				t.Fatal("manual reconciliation evidence lost", err)
			}
		})
	}
}

// TestWorkflowUpgradeOwned rejects actual native capacity and legacy namespace
// locks, then permits the upgrade only after their owners release them.
func TestWorkflowUpgradeOwned(t *testing.T) {
	store, opts := workflowUpgradeFixture(t, "")

	common, err := workflowRepositoryDirectory(t.Context(), store.layout.ProjectDir())
	if err != nil {
		t.Fatal(err)
	}

	guard, err := workflowCapacityFile(common, "chr-workflow-worker.lock", syscall.LOCK_SH)
	if err != nil {
		t.Fatal(err)
	}

	_, blocked := store.UpgradeWorkflowConfiguration(t.Context(), opts)
	if err := guard.Close(); err != nil {
		t.Fatal(err)
	}

	if !errors.Is(blocked, ErrWorkflowCapacity) {
		t.Fatal("native capacity did not block upgrade", blocked)
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: store.layout.ProjectDir(), Namespace: "other"})
	if err != nil {
		t.Fatal(err)
	}

	lock, err := acquireWorkflowAdoptionLock(layout)
	if err != nil {
		t.Fatal(err)
	}

	_, blocked = store.UpgradeWorkflowConfiguration(t.Context(), opts)
	if err := lock.Release(); err != nil {
		t.Fatal(err)
	}

	if blocked == nil {
		t.Fatal("legacy namespace owner did not block upgrade")
	}

	if _, err := store.UpgradeWorkflowConfiguration(t.Context(), opts); err != nil {
		t.Fatal("released owners still blocked upgrade", err)
	}
}

// TestWorkflowUpgradeNative blocks retained unfinished invocation ownership,
// even when no worker process currently holds the repository capacity guard.
func TestWorkflowUpgradeNative(t *testing.T) {
	store, _, request := workflowExecutionFixture(t, []string{"true"}, 30)
	workflowTestStart(t, store, request)

	prior, err := ResolveWorkflowConfiguration(store.layout.ProjectDir(), "")
	if err != nil {
		t.Fatal(err)
	}

	opts := WorkflowUpgradeOptions{ExpectedRevision: prior.Revision, StateAuthority: WorkStateAuthorityTickets, CheckQuiescent: func(context.Context) error { return nil }}
	if _, err := store.UpgradeWorkflowConfiguration(t.Context(), opts); err == nil || !strings.Contains(err.Error(), "unfinished workflow") {
		t.Fatal("retained native ownership ignored", err)
	}
}

// TestWorkflowUpgradePath rejects symlinks, oversize policy and already-versioned
// projects without changing their existing policy or external target bytes.
func TestWorkflowUpgradePath(t *testing.T) {
	for _, kind := range []string{"symlink", "oversize", "versioned", "invalid"} {
		t.Run(kind, func(t *testing.T) {
			store, opts := workflowUpgradeFixture(t, "")

			path := UserPath(store.layout.ProjectDir(), "config.yml")
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}

			image := workflowUpgradeInvalidImage(kind)
			writeWorkflowUpgradeInvalid(t, kind, path, image)

			opts.ExpectedRevision = contenthash.Sum(image)
			if _, err := store.UpgradeWorkflowConfiguration(t.Context(), opts); err == nil {
				t.Fatal("unsafe policy path accepted")
			}

			observed, err := os.ReadFile(path)
			if err != nil || string(observed) != string(image) {
				t.Fatal("invalid policy or symlink target changed", err)
			}
		})
	}
}

func workflowUpgradeInvalidImage(kind string) []byte {
	switch kind {
	case "oversize":
		return []byte(strings.Repeat("#", workflowstate.FileLimit+1))
	case "versioned":
		return []byte("workflow:\n  version: 1\n")
	case "invalid":
		return []byte("review:\n  mode: unknown\n")
	default:
		return []byte("# External file\n")
	}
}

func writeWorkflowUpgradeInvalid(t *testing.T, kind, path string, image []byte) {
	t.Helper()

	if kind != "symlink" {
		if err := os.WriteFile(path, image, 0o600); err != nil {
			t.Fatal(err)
		}

		return
	}

	external := filepath.Join(t.TempDir(), "external.yml")
	if err := os.WriteFile(external, image, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(external, path); err != nil {
		t.Fatal(err)
	}
}

// TestWorkflowUpgradeProcess crashes only its owned fixture after publishing
// the marker, leaving the durable prior-policy snapshot and exclusion locks.
func TestWorkflowUpgradeProcess(t *testing.T) {
	dir := os.Getenv("CHR_UPGRADE_DIR")
	if dir == "" {
		return
	}

	layout, err := opscore.ResolveLayout(opscore.LayoutOptions{ProjectDir: dir, Namespace: "default"})
	if err != nil {
		t.Fatal(err)
	}

	store := NewWorkflowStore(layout, "project-1")
	opts := WorkflowUpgradeOptions{ExpectedRevision: contenthash.Sum(nil), StateAuthority: WorkStateAuthorityTickets, CheckQuiescent: func(context.Context) error { return nil }}

	store.afterReplace = func(stage int) error {
		if stage == 2 {
			os.Exit(17)
		}

		return nil
	}
	if _, err := store.UpgradeWorkflowConfiguration(t.Context(), opts); err != nil {
		t.Fatal(err)
	}

	t.Fatal("owned crash was not reached")
}

// TestWorkflowUpgradeRestart proves a fresh host resumes a lost response with
// the original prior-policy snapshot and reclaims only matching dead locks.
func TestWorkflowUpgradeRestart(t *testing.T) {
	store, opts := workflowUpgradeFixture(t, "")
	command := exec.Command(os.Args[0], "-test.run=^TestWorkflowUpgradeProcess$")

	command.Env = append(os.Environ(), "CHR_UPGRADE_DIR="+store.layout.ProjectDir())
	output, err := command.CombinedOutput()

	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 17 {
		t.Fatalf("owned upgrade crash = %v: %s", err, output)
	}

	data, err := store.readWorkflowUpgradeFile(workflowUpgradePending)
	if err != nil {
		t.Fatal(err)
	}

	var journal workflowUpgradeJournal
	if err := json.Unmarshal(data, &journal); err != nil {
		t.Fatal(err)
	}

	if _, err := store.readWorkflowUpgradeFile("workflow-upgrades/" + journal.RequestHash + ".json"); err != nil {
		t.Fatal("marker published without prior-policy snapshot", err)
	}

	result, err := store.UpgradeWorkflowConfiguration(t.Context(), opts)
	if err != nil {
		t.Fatal(err)
	}

	checkWorkflowUpgradeSnapshot(t, store, opts, result)
}
