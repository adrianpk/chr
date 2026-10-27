package core

import (
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativeMergerInput(t *testing.T) (*WorkflowStore, WorkflowMergerInput, time.Time) {
	t.Helper()

	store := workflowTestStore(t)
	now := time.Now().UTC()
	input := WorkflowMergerInput{
		WorkflowMergerIdentity: WorkflowMergerIdentity{
			ProtocolVersion: 1, ID: "merger-attempt-1", InvocationID: "native-run-1",
			EntryID: contenthash.Sum([]byte("entry")), Unit: WorkflowRef{Kind: "slice", ID: "slice-1"},
			CandidateRevision: 2, CandidateHash: contenthash.Sum([]byte("candidate")), ClaimGeneration: 3,
			ScopeHash: contenthash.Sum([]byte("scope")), Reason: "preparation_conflict",
		}, WorkflowMergerGit: WorkflowMergerGit{
			Repository: "https://forge.example.test/testing/native", SourceRef: "refs/heads/chr/work/source", TargetRef: "refs/heads/dev",
			Source: strings.Repeat("1", 40), Base: strings.Repeat("2", 40), Method: "merge",
			OwnedWorkspaceID: contenthash.Sum([]byte("owned-checkout")),
		}, WorkflowMergerScope: WorkflowMergerScope{AllowedPaths: []string{"src", "docs"}, ProtectedPaths: []string{"src/validation.go", "Makefile"}}, WorkflowMergerValidation: WorkflowMergerValidation{ValidationCommands: [][]string{{"go", "test", "./..."}}}, WorkflowMergerLimits: WorkflowMergerLimits{
			CapacityReservationID: "worker-slot-1", BudgetReservationID: "merger-call-1",
			RemainingAttempts: 3, WorkerTimeoutSeconds: 120, AdmittedAt: now, Deadline: now.Add(time.Minute),
		},
	}
	manifest := WorkflowNativeSourceManifest{Schema: "chr-native-source/v1", Repository: input.Repository, SourceRef: input.SourceRef, OriginalSource: input.Source, Commits: []string{input.Source}}
	input.SourceManifestRef = retainNativeRecordEvidence(t, store, input.InvocationID, "merger-source", manifest)
	input.ValidationContractRef = retainNativeRecordEvidence(t, store, input.InvocationID, "merger-validation", input.ValidationCommands)
	source := retainNativeRecordEvidence(t, store, input.InvocationID, "merger-intent", "Preserve source API behavior")
	target := retainNativeRecordEvidence(t, store, input.InvocationID, "merger-target", "Preserve target validation")
	input.IntentRefs, input.TargetChangeRefs = []WorkflowEvidenceRef{source}, []WorkflowEvidenceRef{target}
	input.Requirements = []WorkflowMergerRequirement{{ID: "source-api", Origin: "source", Evidence: source}, {ID: "target-validation", Origin: "target", Evidence: target}}
	input.Conflicts = []WorkflowMergerConflict{{ID: "conflict-1", Path: "src/service.go", Kind: "text", Evidence: source, RequirementIDs: []string{"source-api", "target-validation"}}}

	sealed, err := SealWorkflowMergerInput(input, now)
	if err != nil {
		t.Fatal(err)
	}

	return store, sealed, now
}

func nativeMergerResolved(input WorkflowMergerInput) WorkflowMergerResult {
	return WorkflowMergerResult{
		ProtocolVersion: 1, InputHash: input.InputHash, Outcome: "resolved", ChangedPaths: []string{"src/service.go"},
		ResolutionNotes:       []WorkflowMergerResolution{{ConflictID: "conflict-1", Path: "src/service.go", Rationale: "Retain the source API and target validation at the shared call site.", RequirementIDs: []string{"source-api", "target-validation"}}},
		PreservedRequirements: []string{"source-api", "target-validation"},
	}
}

func nativeMergerJSON(t *testing.T, value any) []byte {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}

	return data
}

// TestNativeMergerProtocol retains immutable input and rejects implicit role
// authority. Resolved, decision and failed proposals remain distinct outcomes.
func TestNativeMergerProtocol(t *testing.T) {
	_, input, now := nativeMergerInput(t)
	before := nativeMergerJSON(t, input)

	for _, outcome := range []string{"resolved", "needs_decision", "failed"} {
		t.Run(outcome, func(t *testing.T) {
			result := nativeMergerResolved(input)

			result.Outcome = outcome
			if outcome == "needs_decision" {
				result.DecisionRequests = []WorkflowMergerDecision{{ID: "decision-1", Issue: "Which API constraint controls this incompatible behavior?", RequirementIDs: []string{"source-api", "target-validation"}, Options: []string{"Keep the source API requirement", "Keep the target API requirement"}, Authority: "workflow_operator"}}
			}

			if outcome == "failed" {
				result.Failure = "No compatible repair was found within the bounded call."
			}

			parsed, err := ParseWorkflowMergerResult(nativeMergerJSON(t, result), input, workflowMergerFence(input), now)
			if err != nil || !workflowNativeJSONEqual(parsed, result) {
				t.Fatal("valid proposal lost its immutable binding", err)
			}
		})
	}

	if string(before) != string(nativeMergerJSON(t, input)) {
		t.Fatal("proposal parsing mutated the host input")
	}
}

// TestNativeMergerMalformed rejects ambiguous JSON and invented approval,
// publication or receipt fields, including fields nested inside role prose.
func TestNativeMergerMalformed(t *testing.T) {
	_, input, now := nativeMergerInput(t)
	valid := string(nativeMergerJSON(t, nativeMergerResolved(input)))
	cases := map[string]string{
		"empty": "", "null": "null", "array": "[]", "trailing": valid + "{}",
		"oversized":         strings.Repeat(" ", WorkflowFileLimit+1),
		"duplicate":         strings.Replace(valid, `"outcome":"resolved"`, `"outcome":"failed","outcome":"resolved"`, 1),
		"nested duplicate":  strings.Replace(valid, `"conflict_id":"conflict-1"`, `"conflict_id":"other","conflict_id":"conflict-1"`, 1),
		"approval":          strings.Replace(valid, `"outcome":"resolved"`, `"approval":true,"outcome":"resolved"`, 1),
		"receipt":           strings.Replace(valid, `"outcome":"resolved"`, `"receipt":{},"outcome":"resolved"`, 1),
		"publication":       strings.Replace(valid, `"outcome":"resolved"`, `"remote_mutation_grant":true,"outcome":"resolved"`, 1),
		"gate claim":        strings.Replace(valid, `"outcome":"resolved"`, `"host_gate_passed":true,"outcome":"resolved"`, 1),
		"nested permission": strings.Replace(valid, `"conflict_id":"conflict-1"`, `"method":"squash","conflict_id":"conflict-1"`, 1),
		"case duplicate":    strings.Replace(valid, `"outcome":"resolved"`, `"Outcome":"failed","outcome":"resolved"`, 1),
		"case alias":        strings.Replace(valid, `"outcome":"resolved"`, `"Outcome":"resolved"`, 1),
		"invalid utf8":      strings.Replace(valid, "Retain the source", string([]byte{0xff}), 1),
		"deep nesting":      strings.Repeat("[", 34) + "0" + strings.Repeat("]", 34),
	}

	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseWorkflowMergerResult([]byte(data), input, workflowMergerFence(input), now); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("malformed or authority-bearing proposal was accepted", err)
			}
		})
	}
}

// TestNativeMergerFence rejects results after each host acceptance identity
// changes, without trusting a role's unchanged input hash or success narrative.
func TestNativeMergerFence(t *testing.T) {
	_, input, now := nativeMergerInput(t)
	cases := map[string]func(*WorkflowMergerFence){
		"entry":      func(f *WorkflowMergerFence) { f.EntryID = contenthash.Sum([]byte("other")) },
		"candidate":  func(f *WorkflowMergerFence) { f.CandidateHash = contenthash.Sum([]byte("new candidate")) },
		"revision":   func(f *WorkflowMergerFence) { f.CandidateRevision++ },
		"generation": func(f *WorkflowMergerFence) { f.ClaimGeneration++ },
		"scope":      func(f *WorkflowMergerFence) { f.ScopeHash = contenthash.Sum([]byte("expanded scope")) },
		"workspace":  func(f *WorkflowMergerFence) { f.OwnedWorkspaceID = contenthash.Sum([]byte("user checkout")) },
		"source":     func(f *WorkflowMergerFence) { f.Source = strings.Repeat("3", 40) },
		"base":       func(f *WorkflowMergerFence) { f.Base = strings.Repeat("4", 40) },
		"method":     func(f *WorkflowMergerFence) { f.Method = "squash" },
		"target":     func(f *WorkflowMergerFence) { f.TargetRef = "refs/heads/main" },
		"repository": func(f *WorkflowMergerFence) { f.Repository = "https://forge.example.test/testing/other" },
		"unit":       func(f *WorkflowMergerFence) { f.Unit.ID = "slice-2" },
	}
	data := nativeMergerJSON(t, nativeMergerResolved(input))

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			current := workflowMergerFence(input)
			mutate(&current)

			if _, err := ParseWorkflowMergerResult(data, input, current, now); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("stale proposal passed the host fence", err)
			}
		})
	}

	if _, err := ParseWorkflowMergerResult(data, input, workflowMergerFence(input), input.Deadline); !errors.Is(err, workflowstate.ErrConflict) {
		t.Fatal("elapsed deadline was treated as continuing authority", err)
	}
}

// TestNativeMergerUsage preserves unknown accounting and rejects overflow,
// unsupported output-only metering or a report larger than the granted call.
func TestNativeMergerUsage(t *testing.T) {
	_, input, now := nativeMergerInput(t)
	input.TokenAllowance = &WorkflowTokenAllowance{Maximum: 100, Input: 60, Output: 40, WholeCall: true, CapabilityRevision: contenthash.Sum([]byte("harness")), Deadline: input.Deadline}

	input, err := SealWorkflowMergerInput(input, now)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name  string
		usage *WorkflowMergerUsage
		valid bool
	}{
		{"unknown", nil, true},
		{"whole call", &WorkflowMergerUsage{Input: 60, Output: 40, WholeCall: true, CapabilityRevision: input.TokenAllowance.CapabilityRevision}, true},
		{"overspend", &WorkflowMergerUsage{Input: 60, Output: 41, WholeCall: true, CapabilityRevision: input.TokenAllowance.CapabilityRevision}, false},
		{"negative", &WorkflowMergerUsage{Input: -1, WholeCall: true, CapabilityRevision: input.TokenAllowance.CapabilityRevision}, false},
		{"overflow", &WorkflowMergerUsage{Input: math.MaxInt64, Output: 1, WholeCall: true, CapabilityRevision: input.TokenAllowance.CapabilityRevision}, false},
		{"output only", &WorkflowMergerUsage{Output: 20, CapabilityRevision: input.TokenAllowance.CapabilityRevision}, false},
		{"other harness", &WorkflowMergerUsage{Input: 20, WholeCall: true, CapabilityRevision: contenthash.Sum([]byte("other"))}, false},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			result := nativeMergerResolved(input)
			result.Usage = test.usage

			_, err := ParseWorkflowMergerResult(nativeMergerJSON(t, result), input, workflowMergerFence(input), now)
			if (err == nil) != test.valid {
				t.Fatal("usage escaped the immutable call allowance", err)
			}
		})
	}
}
