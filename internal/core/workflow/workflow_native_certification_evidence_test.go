package workflow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type nativeCertificationEvidence struct {
	Schema            string                             `json:"schema"`
	Method            string                             `json:"method"`
	Scenario          string                             `json:"scenario"`
	Transport         string                             `json:"transport,omitempty"`
	Policy            json.RawMessage                    `json:"remote_policy,omitempty"`
	HTTPStatus        int                                `json:"http_status,omitempty"`
	Posts             int                                `json:"publication_posts,omitempty"`
	Server            string                             `json:"server_api_profile"`
	Role              string                             `json:"credential_role"`
	StartedAt         time.Time                          `json:"started_at"`
	FinishedAt        time.Time                          `json:"finished_at"`
	Passed            bool                               `json:"passed"`
	ReceiptCompatible bool                               `json:"receipt_observation_compatible"`
	ObservedSource    string                             `json:"observed_source,omitempty"`
	ObservedTarget    string                             `json:"observed_target,omitempty"`
	ContendedPR       *workflowstate.PullRequestSnapshot `json:"contended_pull_request,omitempty"`
	Native            nativeLiveEvidence                 `json:"native"`
}

func readNativeCertificationRef(t *testing.T, profile nativeLiveProfile, dir, ref string) string {
	t.Helper()

	fields := strings.Fields(workflowIntegrationCommand(t, dir, "ls-remote", profile.GitRemote, ref))
	if len(fields) != 2 || fields[1] != ref || !workflowstate.ValidCommitID(fields[0]) {
		t.Fatal("canonical fixture ref is unavailable")
	}

	return fields[0]
}

func checkNativeReceiptCompatibility(t *testing.T, evidence *nativeCertificationEvidence) {
	t.Helper()

	binding := evidence.Native.Binding

	observed := evidence.Native.Observations[len(evidence.Native.Observations)-1]
	if observed.State != workflowstate.NativeMerged || observed.Git == nil || observed.Git.ObservedTarget != binding.Git.ResultCommit {
		t.Fatal("original canonical PR lacks exact merged Git proof")
	}

	git := *observed.Git
	b := workflowstate.NativeRecordBinding{
		SourceRepository: git.Repository, SourceRef: binding.Capability.SourceRef, TargetRef: binding.Capability.TargetRef,
		Method: git.Method, OriginalSource: git.SourceCommit, ValidatedBase: git.BaseCommit,
		PreparedResult: git.ResultCommit, PublishedSource: git.ResultCommit, PreparedTree: git.PreparedTree,
		CanonicalPR: binding.PullRequest.URL,
	}
	native := workflowstate.NativeReceipt{NativeRecordBinding: b, ActualMergeResult: observed.PullRequest.ResultCommit}
	native.ObservationContract = observed.ObservationContract

	receipt := Receipt{
		Native: &native, Repository: git.Repository, Method: git.Method, SourceCommit: git.SourceCommit,
		BaseCommit: git.BaseCommit, PreparedTree: git.PreparedTree, HistoryProof: git.HistoryHash,
		ResultCommit: git.ResultCommit, VerifiedTree: git.VerifiedTree, ObservedTarget: git.ObservedTarget,
	}
	if err := validateWorkflowNativeReceiptObservation(receipt, observed); err != nil || !validWorkflowNativeReceiptTransaction(native, binding.Publication) {
		t.Fatalf("real publication is incompatible with receipt proof: %v", err)
	}

	evidence.ReceiptCompatible = true
}

func retainNativeCertification(t *testing.T, state, key string, evidence *nativeCertificationEvidence) {
	t.Helper()

	evidence.FinishedAt, evidence.Passed = time.Now().UTC(), !t.Failed()
	if t.Failed() && evidence.Native.Failure == "" {
		evidence.Native.Failure = "live case failed; retained command output provides the failure boundary"
	}

	data, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	dir := filepath.Join(state, "chr", "qa", "native-integration-slice-12")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, evidence.Native.Profile+"-"+key+".json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
}
