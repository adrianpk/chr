package core

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	opsfile "forge.adrianpk.com/chr/internal/core/ops/file"
)

func retainNativeRecordEvidence(t *testing.T, store *WorkflowStore, id, name string, value any) WorkflowEvidenceRef {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}

	ref := WorkflowEvidenceRef{Path: "run/" + id + "/evidence/native-" + name + ".json", Hash: contenthash.Sum(data)}

	path, err := store.safePath(ref.Path)
	if err != nil {
		t.Fatal(err)
	}

	if err := opsfile.WriteAtomic(path, data); err != nil {
		t.Fatal(err)
	}

	return ref
}

func nativeRecordCandidate(t *testing.T, method string) (*WorkflowStore, WorkflowIntegrationCandidate) {
	t.Helper()

	f := newWorkflowPreparationFixture(t, method, workflowPreparationGate("test -f result.txt && if test -f target.txt; then printf combined; else printf source; fi"))

	candidate, err := f.store.PrepareWorkflowSliceIntegration(t.Context(), f.run.ID, f.provider, f.git)
	if err != nil {
		t.Fatal(err)
	}

	if candidate.Validation.Validation[0].Stdout != "combined" {
		t.Fatal("native fixture did not validate combined target content")
	}

	proof, pr := candidate.Input.Git, candidate.Input.Source.PullRequest
	cap := WorkflowNativeCapability{WorkflowPublicationCapability: WorkflowPublicationCapability{Repository: proof.Repository, PullRequestURL: pr.URL, Method: method, Mechanism: WorkflowPublicationGitAtomicDualRef, AtomicSourceBase: true, ExactResult: true}, SourceRef: "refs/heads/" + pr.HeadBranch, TargetRef: "refs/heads/" + pr.BaseBranch, OwnershipRevision: "native-fixture/v1", PolicyFingerprint: contenthash.Sum([]byte("policy")), ServerCapabilityFingerprint: contenthash.Sum([]byte("server")), CredentialIdentityFingerprint: contenthash.Sum([]byte("publisher")), ObservationContract: "forgejo_same_pr_git_auto_merge/v1", AdapterVersion: workflowNativeAdapterVersion, OriginalSourceCommit: proof.SourceCommit, ExpectedBaseCommit: proof.BaseCommit, PreparedResultCommit: proof.ResultCommit}

	cap, err = sealWorkflowNativeCapability(cap)
	if err != nil {
		t.Fatal(err)
	}

	candidate.Input.Capability = cap.WorkflowPublicationCapability
	candidate.Schema = workflowNativeCandidateSchema

	inputData, err := f.store.readWorkflowReviewImage(candidate.Validation.InputRef, candidate.Validation.InputHash)
	if err != nil {
		t.Fatal(err)
	}

	var input WorkflowTaskInput
	if err := json.Unmarshal(inputData, &input); err != nil {
		t.Fatal(err)
	}

	input.Preparation = &candidate.Input
	inputRef := retainNativeRecordEvidence(t, f.store, candidate.InvocationID, "input", input)
	candidate.Validation.InputRef, candidate.Validation.InputHash = inputRef.Path, inputRef.Hash
	b := WorkflowNativeRecordBinding{SchemaVersion: 1, EntryID: "entry-1", InvocationID: candidate.InvocationID, Unit: candidate.Input.Source.Unit, CanonicalPR: pr.URL, SourceRepository: pr.Repository, SourceRef: cap.SourceRef, TargetRef: cap.TargetRef, ObjectFormat: "sha1", OriginalSource: proof.SourceCommit, PublishedSource: proof.ResultCommit, ValidatedBase: proof.BaseCommit, PreparedResult: proof.ResultCommit, PreparedTree: proof.PreparedTree, Method: method, ContractHash: candidate.Input.Source.Slice.ContractHash, PolicyHash: cap.PolicyFingerprint, CapabilityHash: cap.Revision, Extensions: map[string]json.RawMessage{"future_optional": json.RawMessage(`{"retained":true}`)}}
	commits := strings.Fields(workflowIntegrationCommand(t, candidate.Lane.Worktree, "rev-list", "--reverse", "--first-parent", proof.OriginCommit+".."+proof.SourceCommit))
	b.SourceManifestRef = retainNativeRecordEvidence(t, f.store, b.InvocationID, "source", WorkflowNativeSourceManifest{Schema: "chr-native-source/v1", Repository: b.SourceRepository, SourceRef: b.SourceRef, OriginalSource: b.OriginalSource, Commits: commits})
	b.HistoryManifestRef = retainNativeRecordEvidence(t, f.store, b.InvocationID, "history", WorkflowNativeHistoryManifest{Schema: "chr-native-history/v1", Git: proof})
	b.PreparationRef = retainNativeRecordEvidence(t, f.store, b.InvocationID, "preparation", candidate.Input)
	b.CapabilityRef = retainNativeRecordEvidence(t, f.store, b.InvocationID, "capability", cap)
	b.ValidationRefs = []WorkflowEvidenceRef{retainNativeRecordEvidence(t, f.store, b.InvocationID, "validation", candidate.Validation)}
	candidate.Native = &WorkflowNativeCandidate{WorkflowNativeRecordBinding: b}

	candidate.Native.CandidateHash, err = WorkflowNativeCandidateHash(candidate)
	if err != nil {
		t.Fatal(err)
	}

	return f.store, candidate
}

// TestNativeRecordMethods reads exact S/B/P evidence for all existing Git methods
// and preserves optional extensions in the immutable authorization hash.
func TestNativeRecordMethods(t *testing.T) {
	for _, method := range []string{"merge", "rebase", "squash"} {
		t.Run(method, func(t *testing.T) {
			store, candidate := nativeRecordCandidate(t, method)
			if err := store.checkWorkflowNativeCandidate(candidate); err != nil {
				t.Fatal(err)
			}

			data, err := json.Marshal(candidate)
			if err != nil {
				t.Fatal(err)
			}

			var again WorkflowIntegrationCandidate
			if err := json.Unmarshal(data, &again); err != nil {
				t.Fatal(err)
			}

			if err := store.checkWorkflowNativeCandidate(again); err != nil || again.Native.OriginalSource == again.Native.PublishedSource || string(again.Native.Extensions["future_optional"]) != `{"retained":true}` {
				t.Fatalf("lost native binding: %v", err)
			}
		})
	}
}

// TestNativeRecordGuards rejects unsupported/incomplete and stale candidate
// authorization before any new-schema record can be treated as publishable.
func TestNativeRecordGuards(t *testing.T) {
	store, original := nativeRecordCandidate(t, "rebase")

	cases := []struct {
		name   string
		change func(*WorkflowIntegrationCandidate)
	}{
		{"unsupported", func(c *WorkflowIntegrationCandidate) { c.Schema = "chr-integration-candidate/v3" }},
		{"downgrade", func(c *WorkflowIntegrationCandidate) { c.Schema = "chr-integration-candidate/v1" }},
		{"missing", func(c *WorkflowIntegrationCandidate) { c.Native = nil }},
		{"version", func(c *WorkflowIntegrationCandidate) { c.Native.SchemaVersion = 2 }},
		{"source", func(c *WorkflowIntegrationCandidate) { c.Native.OriginalSource = c.Native.PublishedSource }},
		{"owner", func(c *WorkflowIntegrationCandidate) { c.Native.InvocationID = "other" }},
		{"unit", func(c *WorkflowIntegrationCandidate) { c.Native.Unit.ID += "other" }},
		{"method", func(c *WorkflowIntegrationCandidate) { c.Native.Method = "squash" }},
		{"format", func(c *WorkflowIntegrationCandidate) { c.Native.ObjectFormat = "sha256" }},
		{"path", func(c *WorkflowIntegrationCandidate) {
			c.Native.HistoryManifestRef.Path = "run/other/evidence/proof.json"
		}},
		{"escape", func(c *WorkflowIntegrationCandidate) { c.Native.HistoryManifestRef.Path += "/../proof.json" }},
		{"hash", func(c *WorkflowIntegrationCandidate) { c.Native.CapabilityRef.Hash = strings.Repeat("a", 64) }},
		{"extension", func(c *WorkflowIntegrationCandidate) {
			c.Native.Extensions = map[string]json.RawMessage{"changed": json.RawMessage(`true`)}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			candidate := original
			native := *original.Native
			candidate.Native = &native
			tc.change(&candidate)

			if err := store.checkWorkflowNativeCandidate(candidate); err == nil {
				t.Fatal("accepted incomplete or altered native proof")
			}
		})
	}

	testNativeRecordEvidence(t, store, original)
}

func testNativeRecordEvidence(t *testing.T, store *WorkflowStore, candidate WorkflowIntegrationCandidate) {
	t.Helper()

	ref := candidate.Native.HistoryManifestRef
	path := filepath.Join(store.layout.Root(), ref.Path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, append(data, ' '), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := store.checkWorkflowNativeCandidate(candidate); err == nil {
		t.Fatal("accepted changed evidence image")
	}

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	outside := filepath.Join(t.TempDir(), "history.json")
	if err := os.WriteFile(outside, data, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(outside, path); err != nil {
		t.Fatal(err)
	}

	if err := store.checkWorkflowNativeCandidate(candidate); err == nil {
		t.Fatal("accepted symlinked proof")
	}
}

// TestNativeRecordVersions rejects unsupported evidence versions even when their
// new file hashes and candidate binding are internally consistent.
func TestNativeRecordVersions(t *testing.T) {
	store, original := nativeRecordCandidate(t, "rebase")

	cases := []struct {
		name   string
		change func(*testing.T, *WorkflowIntegrationCandidate)
	}{
		{"source", func(t *testing.T, c *WorkflowIntegrationCandidate) {
			b := &c.Native.WorkflowNativeRecordBinding

			var source WorkflowNativeSourceManifest
			if err := store.readWorkflowNativeEvidence(b.InvocationID, b.SourceManifestRef, &source); err != nil {
				t.Fatal(err)
			}

			source.Schema = "chr-native-source/v2"
			b.SourceManifestRef = retainNativeRecordEvidence(t, store, b.InvocationID, "unsupported-source", source)
		}},
		{"history", func(t *testing.T, c *WorkflowIntegrationCandidate) {
			b := &c.Native.WorkflowNativeRecordBinding
			b.HistoryManifestRef = retainNativeRecordEvidence(t, store, b.InvocationID, "unsupported-history", WorkflowNativeHistoryManifest{Schema: "chr-native-history/v2", Git: c.Input.Git})
		}},
		{"capability", func(t *testing.T, c *WorkflowIntegrationCandidate) {
			b := &c.Native.WorkflowNativeRecordBinding

			var cap WorkflowNativeCapability
			if err := store.readWorkflowNativeEvidence(b.InvocationID, b.CapabilityRef, &cap); err != nil {
				t.Fatal(err)
			}

			cap.AdapterVersion = "chr-native-publication/v2"
			b.CapabilityRef = retainNativeRecordEvidence(t, store, b.InvocationID, "unsupported-capability", cap)
		}},
		{"validation", func(t *testing.T, c *WorkflowIntegrationCandidate) {
			b := &c.Native.WorkflowNativeRecordBinding
			result := c.Validation
			result.Schema = "chr-integration-validation/v2"
			b.ValidationRefs = []WorkflowEvidenceRef{retainNativeRecordEvidence(t, store, b.InvocationID, "unsupported-validation", result)}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			candidate := original
			native := *original.Native
			candidate.Native = &native
			tc.change(t, &candidate)

			var err error

			candidate.Native.CandidateHash, err = WorkflowNativeCandidateHash(candidate)
			if err != nil {
				t.Fatal(err)
			}

			if err := store.checkWorkflowNativeCandidate(candidate); err == nil {
				t.Fatal("accepted unsupported evidence version")
			}
		})
	}
}
