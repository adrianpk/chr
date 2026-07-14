package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestNewWorktreePlanUsesRemoteBaseAndDeterministicPath(t *testing.T) {
	dir := t.TempDir()

	plan, err := NewWorktreePlan(WorktreePlanOptions{
		ProjectDir: dir,
		TicketKey:  "CHR-1 Add CLI",
		Branch:     "feat/chr-1-add-cli",
	})
	if err != nil {
		t.Fatalf("NewWorktreePlan returned error: %v", err)
	}

	wantRoot := filepath.Join(dir, "ops", "default", "worktree")

	if plan.BaseRef != "origin/dev" {
		t.Fatalf("BaseRef = %q, want origin/dev", plan.BaseRef)
	}

	if plan.WorktreeRoot != wantRoot {
		t.Fatalf("WorktreeRoot = %q, want %q", plan.WorktreeRoot, wantRoot)
	}

	if plan.WorktreePath != filepath.Join(wantRoot, "chr-1-add-cli") {
		t.Fatalf("WorktreePath = %q, want deterministic ticket path", plan.WorktreePath)
	}
}

func TestNewWorktreePlanUsesConfiguredOpsNamespace(t *testing.T) {
	t.Setenv(OpsUserEnv, "adrian")
	dir := t.TempDir()

	plan, err := NewWorktreePlan(WorktreePlanOptions{
		ProjectDir: dir,
		TicketKey:  "CHR-1",
		Branch:     "feat/chr-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := plan.WorktreeRoot, filepath.Join(dir, "ops", "adrian", "worktree"); got != want {
		t.Fatalf("WorktreeRoot = %q, want %q", got, want)
	}
}

func TestWorktreePathComponentSanitizesTicketKeys(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "simple", value: "CHR-1", want: "chr-1"},
		{name: "spaces", value: "CHR-1 Add CLI", want: "chr-1-add-cli"},
		{name: "slashes", value: "../CHR/1", want: "chr-1"},
		{name: "symbols", value: "!!!CHR@@@1!!!", want: "chr-1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := WorktreePathComponent(test.value); got != test.want {
				t.Fatalf("WorktreePathComponent(%q) = %q, want %q", test.value, got, test.want)
			}
		})
	}
}

func TestParseWorktreeGitListReadsPorcelainEntries(t *testing.T) {
	entries := parseWorktreeGitList(`worktree /repo
HEAD abc123
branch refs/heads/dev

worktree /repo/.chr/default/runtime/worktrees/chr-1
HEAD def456
branch refs/heads/feat/chr-1

worktree /tmp/orphan
HEAD 000000
detached
prunable gitdir file points to non-existent location
`)

	if len(entries) != 3 {
		t.Fatalf("len(entries) = %d, want 3", len(entries))
	}

	if entries[1].Path != "/repo/.chr/default/runtime/worktrees/chr-1" || entries[1].Branch != "feat/chr-1" {
		t.Fatalf("entry[1] = %#v, want worktree branch entry", entries[1])
	}

	if !entries[2].Detached || !entries[2].Prunable {
		t.Fatalf("entry[2] = %#v, want detached prunable entry", entries[2])
	}
}

func TestWorktreeServicePrepareCreatesBranchAndAddsMissingWorktree(t *testing.T) {
	dir := t.TempDir()
	git := newFakeWorktreeGit()
	service := WorktreeService{Git: git}

	result, err := service.Prepare(context.Background(), WorktreePlanOptions{
		ProjectDir: dir,
		TicketKey:  "CHR-1",
		Branch:     "feat/chr-1",
	})
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}

	if !slices.Equal(git.fetched, []string{"origin"}) {
		t.Fatalf("fetched = %#v, want origin", git.fetched)
	}

	if !result.CreatedBranch || !result.AddedWorktree {
		t.Fatalf("result = %#v, want created branch and added worktree", result)
	}

	if got := git.createdBranches["feat/chr-1"]; got != "origin/dev" {
		t.Fatalf("created branch base = %q, want origin/dev", got)
	}

	if len(git.addedWorktrees) != 1 || git.addedWorktrees[0].branch != "feat/chr-1" {
		t.Fatalf("added worktrees = %#v, want feat/chr-1 add", git.addedWorktrees)
	}
}

func TestWorktreeServicePrepareReusesExistingCleanWorktree(t *testing.T) {
	dir := t.TempDir()
	plan := testWorktreePlan(t, dir)
	git := newFakeWorktreeGit()
	git.branches[plan.Branch] = true
	git.entries = []WorktreeGitEntry{{Path: plan.WorktreePath, Branch: plan.Branch, Head: "base-sha"}}
	service := WorktreeService{Git: git}

	result, err := service.Prepare(context.Background(), WorktreePlanOptions{
		ProjectDir: dir,
		TicketKey:  plan.TicketKey,
		Branch:     plan.Branch,
	})
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}

	if result.Inspection.State != WorktreeStateClean {
		t.Fatalf("Inspection.State = %q, want clean", result.Inspection.State)
	}

	if result.AddedWorktree {
		t.Fatal("AddedWorktree = true, want existing worktree reuse")
	}
}

func TestWorktreeServicePrepareBlocksExistingBranchWithoutBase(t *testing.T) {
	dir := t.TempDir()
	plan := testWorktreePlan(t, dir)
	git := newFakeWorktreeGit()
	git.branches[plan.Branch] = true
	git.branchContainsRef[branchContainsRefKey(plan.Branch, plan.BaseRef)] = false
	service := WorktreeService{Git: git}

	_, err := service.Prepare(context.Background(), WorktreePlanOptions{
		ProjectDir: dir,
		TicketKey:  plan.TicketKey,
		Branch:     plan.Branch,
	})
	if err == nil {
		t.Fatal("Prepare returned nil error")
	}

	if got := err.Error(); got != "worktree branch feat/chr-1 does not contain base origin/dev" {
		t.Fatalf("error = %q, want branch base blocker", got)
	}
}

func TestWorktreeServiceInspectClassifiesTargetStates(t *testing.T) {
	dir := t.TempDir()
	plan := testWorktreePlan(t, dir)

	tests := []struct {
		name     string
		entries  []WorktreeGitEntry
		statuses map[string][]string
		want     WorktreeState
	}{
		{name: "missing", want: WorktreeStateMissing},
		{
			name:    "clean",
			entries: []WorktreeGitEntry{{Path: plan.WorktreePath, Branch: plan.Branch}},
			want:    WorktreeStateClean,
		},
		{
			name:     "dirty",
			entries:  []WorktreeGitEntry{{Path: plan.WorktreePath, Branch: plan.Branch}},
			statuses: map[string][]string{plan.WorktreePath: {" M main.go"}},
			want:     WorktreeStateDirty,
		},
		{
			name:    "mismatched",
			entries: []WorktreeGitEntry{{Path: plan.WorktreePath, Branch: "feat/other"}},
			want:    WorktreeStateMismatched,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertWorktreeInspectionState(t, dir, worktreeInspectionStateAssertion{
				entries:  test.entries,
				statuses: test.statuses,
				want:     test.want,
			})
		})
	}
}

func TestWorktreeServiceInspectClassifiesBranchInUseAndOrphaned(t *testing.T) {
	dir := t.TempDir()
	plan := testWorktreePlan(t, dir)

	branchInUse := []WorktreeGitEntry{{Path: filepath.Join(dir, "other-worktree"), Branch: plan.Branch}}
	assertWorktreeInspectionState(t, dir, worktreeInspectionStateAssertion{
		entries: branchInUse,
		want:    WorktreeStateBranchInUse,
	})

	if err := writeWorktreeTestDir(plan.WorktreePath); err != nil {
		t.Fatalf("create orphaned worktree path: %v", err)
	}

	assertWorktreeInspectionState(t, dir, worktreeInspectionStateAssertion{want: WorktreeStateOrphaned})
}

func TestWorktreeMetadataRoundTrips(t *testing.T) {
	dir := t.TempDir()
	plan := testWorktreePlan(t, dir)
	updatedAt := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	metadata := NewWorktreeMetadata(plan, updatedAt)
	metadata.LastExecutionID = "exec-1"
	metadata.LastExecutionStatus = "ready_for_review"

	path, err := WorktreeMetadataPath(plan)
	if err != nil {
		t.Fatalf("WorktreeMetadataPath returned error: %v", err)
	}

	if err := WriteWorktreeMetadata(path, metadata); err != nil {
		t.Fatalf("WriteWorktreeMetadata returned error: %v", err)
	}

	got, err := ReadWorktreeMetadata(path)
	if err != nil {
		t.Fatalf("ReadWorktreeMetadata returned error: %v", err)
	}

	if got != metadata {
		t.Fatalf("metadata = %#v, want %#v", got, metadata)
	}
}

func TestCleanupMergedTicketWorktreeRemovesCleanWorktreeAndMetadata(t *testing.T) {
	dir := t.TempDir()
	plan := testWorktreePlan(t, dir)
	git := newFakeWorktreeGit()
	git.entries = []WorktreeGitEntry{{Path: plan.WorktreePath, Branch: plan.Branch}}
	writeWorktreeTestMetadata(t, plan)

	result, err := CleanupMergedTicketWorktree(context.Background(), WorktreeCleanupOptions{
		ProjectDir: dir,
		TicketKey:  plan.TicketKey,
		Branch:     plan.Branch,
		Workspace:  testCleanupWorktreeWorkspace(),
		Git:        git,
	})
	if err != nil {
		t.Fatalf("CleanupMergedTicketWorktree returned error: %v", err)
	}

	if result.Status != WorktreeCleanupStatusRemoved || !result.RemovedWorktree || !result.RemovedMetadata {
		t.Fatalf("result = %#v, want removed worktree and metadata", result)
	}

	if len(git.removedWorktrees) != 1 || git.removedWorktrees[0] != plan.WorktreePath {
		t.Fatalf("removedWorktrees = %#v, want %q", git.removedWorktrees, plan.WorktreePath)
	}

	assertWorktreeMetadataMissing(t, plan)
}

func TestCleanupMergedTicketWorktreeBlocksDirtyWorktree(t *testing.T) {
	dir := t.TempDir()
	plan := testWorktreePlan(t, dir)
	git := newFakeWorktreeGit()
	git.entries = []WorktreeGitEntry{{Path: plan.WorktreePath, Branch: plan.Branch}}
	git.statuses[plan.WorktreePath] = []string{" M main.go"}

	result, err := CleanupMergedTicketWorktree(context.Background(), WorktreeCleanupOptions{
		ProjectDir: dir,
		TicketKey:  plan.TicketKey,
		Branch:     plan.Branch,
		Workspace:  testCleanupWorktreeWorkspace(),
		Git:        git,
	})
	if err != nil {
		t.Fatalf("CleanupMergedTicketWorktree returned error: %v", err)
	}

	if result.Status != WorktreeCleanupStatusBlocked || result.RemovedWorktree {
		t.Fatalf("result = %#v, want blocked without removal", result)
	}

	if len(git.removedWorktrees) != 0 {
		t.Fatalf("removedWorktrees = %#v, want none", git.removedWorktrees)
	}
}

func TestCleanupMergedTicketWorktreePrunesMissingMetadata(t *testing.T) {
	dir := t.TempDir()
	plan := testWorktreePlan(t, dir)
	git := newFakeWorktreeGit()

	writeWorktreeTestMetadata(t, plan)

	result, err := CleanupMergedTicketWorktree(context.Background(), WorktreeCleanupOptions{
		ProjectDir: dir,
		TicketKey:  plan.TicketKey,
		Branch:     plan.Branch,
		Workspace:  testCleanupWorktreeWorkspace(),
		Git:        git,
	})
	if err != nil {
		t.Fatalf("CleanupMergedTicketWorktree returned error: %v", err)
	}

	if result.Status != WorktreeCleanupStatusSkipped || !result.RemovedMetadata {
		t.Fatalf("result = %#v, want skipped with metadata prune", result)
	}

	assertWorktreeMetadataMissing(t, plan)
}

func TestCleanupMergedTicketWorktreeBlocksBranchInUse(t *testing.T) {
	dir := t.TempDir()
	plan := testWorktreePlan(t, dir)
	git := newFakeWorktreeGit()
	git.entries = []WorktreeGitEntry{{Path: filepath.Join(dir, "other-worktree"), Branch: plan.Branch}}

	result, err := CleanupMergedTicketWorktree(context.Background(), WorktreeCleanupOptions{
		ProjectDir: dir,
		TicketKey:  plan.TicketKey,
		Branch:     plan.Branch,
		Workspace:  testCleanupWorktreeWorkspace(),
		Git:        git,
	})
	if err != nil {
		t.Fatalf("CleanupMergedTicketWorktree returned error: %v", err)
	}

	if result.Status != WorktreeCleanupStatusBlocked || result.Detail == "" {
		t.Fatalf("result = %#v, want blocked branch-in-use detail", result)
	}
}

func testCleanupWorktreeWorkspace() WorkWorkspaceConfig {
	workspace := DefaultWorkWorkspaceConfig()
	workspace.Mode = WorkWorkspaceModeWorktree

	return workspace
}

func writeWorktreeTestMetadata(t *testing.T, plan WorktreePlan) {
	t.Helper()

	path, err := WorktreeMetadataPath(plan)
	if err != nil {
		t.Fatalf("WorktreeMetadataPath returned error: %v", err)
	}

	if err := WriteWorktreeMetadata(path, NewWorktreeMetadata(plan, time.Now())); err != nil {
		t.Fatalf("WriteWorktreeMetadata returned error: %v", err)
	}
}

func assertWorktreeMetadataMissing(t *testing.T, plan WorktreePlan) {
	t.Helper()

	path, err := WorktreeMetadataPath(plan)
	if err != nil {
		t.Fatalf("WorktreeMetadataPath returned error: %v", err)
	}

	_, err = os.Stat(path)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("metadata path %s stat error = %v, want missing", path, err)
	}
}

type worktreeInspectionStateAssertion struct {
	entries  []WorktreeGitEntry
	statuses map[string][]string
	want     WorktreeState
}

func assertWorktreeInspectionState(t *testing.T, dir string, assertion worktreeInspectionStateAssertion) {
	t.Helper()

	plan := testWorktreePlan(t, dir)
	git := newFakeWorktreeGit()
	git.entries = assertion.entries
	git.statuses = assertion.statuses
	service := WorktreeService{Git: git}

	inspection, err := service.Inspect(context.Background(), WorktreePlanOptions{
		ProjectDir: dir,
		TicketKey:  plan.TicketKey,
		Branch:     plan.Branch,
	})
	if err != nil {
		t.Fatalf("Inspect returned error: %v", err)
	}

	if inspection.State != assertion.want {
		t.Fatalf("Inspection.State = %q, want %q; diagnostics=%#v", inspection.State, assertion.want, inspection.Diagnostics)
	}
}

func testWorktreePlan(t *testing.T, dir string) WorktreePlan {
	t.Helper()

	plan, err := NewWorktreePlan(WorktreePlanOptions{
		ProjectDir: dir,
		TicketKey:  "CHR-1",
		Branch:     "feat/chr-1",
	})
	if err != nil {
		t.Fatalf("NewWorktreePlan returned error: %v", err)
	}

	return plan
}

func writeWorktreeTestDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

type fakeWorktreeGit struct {
	fetched           []string
	refs              map[string]string
	branches          map[string]bool
	branchContainsRef map[string]bool
	createdBranches   map[string]string
	addedWorktrees    []fakeAddedWorktree
	removedWorktrees  []string
	entries           []WorktreeGitEntry
	statuses          map[string][]string
	fetchErr          error
}

type fakeAddedWorktree struct {
	path   string
	branch string
}

func newFakeWorktreeGit() *fakeWorktreeGit {
	return &fakeWorktreeGit{
		refs:              map[string]string{"origin/dev": "base-sha"},
		branches:          map[string]bool{},
		branchContainsRef: map[string]bool{},
		createdBranches:   map[string]string{},
		statuses:          map[string][]string{},
	}
}

func (git *fakeWorktreeGit) FetchRemote(_ context.Context, _, remote string) error {
	if git.fetchErr != nil {
		return git.fetchErr
	}

	git.fetched = append(git.fetched, remote)

	return nil
}

func (git *fakeWorktreeGit) ResolveRef(_ context.Context, _, ref string) (string, error) {
	sha, ok := git.refs[ref]
	if !ok {
		return "", errors.New("missing ref")
	}

	return sha, nil
}

func (git *fakeWorktreeGit) BranchExists(_ context.Context, _, branch string) (bool, error) {
	return git.branches[branch], nil
}

func (git *fakeWorktreeGit) BranchContainsRef(_ context.Context, _, branch, ref string) (bool, error) {
	contains, ok := git.branchContainsRef[branchContainsRefKey(branch, ref)]
	if !ok {
		return true, nil
	}

	return contains, nil
}

func (git *fakeWorktreeGit) CreateBranch(_ context.Context, _, branch, baseRef string) error {
	git.branches[branch] = true
	git.createdBranches[branch] = baseRef

	return nil
}

func (git *fakeWorktreeGit) AddWorktree(_ context.Context, _, path, branch string) error {
	git.addedWorktrees = append(git.addedWorktrees, fakeAddedWorktree{path: path, branch: branch})

	return nil
}

func (git *fakeWorktreeGit) RemoveWorktree(_ context.Context, _, path string) error {
	git.removedWorktrees = append(git.removedWorktrees, path)

	return nil
}

func (git *fakeWorktreeGit) WorktreeList(context.Context, string) ([]WorktreeGitEntry, error) {
	return append([]WorktreeGitEntry(nil), git.entries...), nil
}

func (git *fakeWorktreeGit) WorktreeStatus(_ context.Context, path string) ([]string, error) {
	return append([]string(nil), git.statuses[path]...), nil
}

func branchContainsRefKey(branch, ref string) string {
	return branch + "\x00" + ref
}
