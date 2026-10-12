package core

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func nativeMergerInputGuards() map[string]func(*WorkflowMergerInput) {
	return map[string]func(*WorkflowMergerInput){
		"version":               func(i *WorkflowMergerInput) { i.ProtocolVersion = 2 },
		"identity":              func(i *WorkflowMergerInput) { i.ClaimGeneration = 0 },
		"escaped attempt":       func(i *WorkflowMergerInput) { i.ID = "../attempt" },
		"opaque workspace":      func(i *WorkflowMergerInput) { i.OwnedWorkspaceID = "/user/checkout" },
		"relative evidence":     func(i *WorkflowMergerInput) { i.IntentRefs[0].Path = "run/native-run-1/evidence/../../escape.json" },
		"foreign evidence":      func(i *WorkflowMergerInput) { i.IntentRefs[0].Path = "run/other-run/evidence/intent.json" },
		"absolute evidence":     func(i *WorkflowMergerInput) { i.IntentRefs[0].Path = "/tmp/intent.json" },
		"unbounded attempts":    func(i *WorkflowMergerInput) { i.RemainingAttempts = 11 },
		"exhausted":             func(i *WorkflowMergerInput) { i.RemainingAttempts = 0 },
		"unbounded runtime":     func(i *WorkflowMergerInput) { i.WorkerTimeoutSeconds = 0 },
		"excessive deadline":    func(i *WorkflowMergerInput) { i.Deadline = i.AdmittedAt.Add(121 * time.Second) },
		"expired":               func(i *WorkflowMergerInput) { i.Deadline = i.AdmittedAt },
		"future admission":      func(i *WorkflowMergerInput) { i.AdmittedAt = i.Deadline },
		"clean candidate":       func(i *WorkflowMergerInput) { i.Conflicts = nil },
		"drift only":            func(i *WorkflowMergerInput) { i.Reason = "target_drift" },
		"unsupported method":    func(i *WorkflowMergerInput) { i.Method = "other" },
		"wrong format":          func(i *WorkflowMergerInput) { i.Base = strings.Repeat("2", 64) },
		"target alias":          func(i *WorkflowMergerInput) { i.TargetRef = "dev" },
		"same ref":              func(i *WorkflowMergerInput) { i.TargetRef = i.SourceRef },
		"missing target intent": func(i *WorkflowMergerInput) { i.TargetChangeRefs = nil },
		"missing source intent": func(i *WorkflowMergerInput) { i.IntentRefs = nil },
		"weakened requirements": func(i *WorkflowMergerInput) { i.Requirements[1].Origin = "source" },
		"foreign requirement":   func(i *WorkflowMergerInput) { i.Conflicts[0].RequirementIDs[0] = "undeclared" },
		"unknown conflict":      func(i *WorkflowMergerInput) { i.Conflicts[0].Kind = "automatic" },
		"empty gate":            func(i *WorkflowMergerInput) { i.ValidationCommands = nil },
		"command bound":         func(i *WorkflowMergerInput) { i.ValidationCommands[0][0] = strings.Repeat("x", 4097) },
		"invalid command":       func(i *WorkflowMergerInput) { i.ValidationCommands[0][0] = "go\x00" },
	}
}

// TestNativeMergerInput rejects malformed scope, unsupported triggers and
// missing bounded evidence before any role can acquire shared capacity.
func TestNativeMergerInput(t *testing.T) {
	for name, mutate := range nativeMergerInputGuards() {
		t.Run(name, func(t *testing.T) {
			_, input, now := nativeMergerInput(t)
			mutate(&input)

			if _, err := SealWorkflowMergerInput(input, now); !errors.Is(err, ErrWorkflowConflict) {
				t.Fatal("invalid role input was admitted", err)
			}
		})
	}

	_, input, now := nativeMergerInput(t)

	input.ValidationCommands[0][0] = "false"
	if err := validateWorkflowMergerInput(input, now); !errors.Is(err, ErrWorkflowConflict) {
		t.Fatal("mutated validation contract retained its old input hash", err)
	}
}

// TestNativeMergerScope checks directory boundaries and denies workflow,
// ticket-scope and gate edits even when a result narrates successful repair.
func TestNativeMergerScope(t *testing.T) {
	paths := []string{"../escape", "/tmp/result", "src/../escape", "src/validation.go", "src/validation.go/check", "src-other/service.go", "src\\service.go", "src/.git/config", "src/.chr/run.json", "src/AGENTS.md", "src/.agents/LINTING.md"}
	_, input, now := nativeMergerInput(t)

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			result := nativeMergerResolved(input)

			result.ChangedPaths = []string{path}
			if _, err := ParseWorkflowMergerResult(nativeMergerJSON(t, result), input, workflowMergerFence(input), now); !errors.Is(err, ErrWorkflowConflict) {
				t.Fatal("out-of-scope role change was accepted", err)
			}
		})
	}

	for _, path := range []string{"ops/default/ticket/ready", "ops/default/spec/stable", "ops/default/plan", "ops/default/tracker", ".git", ".chr", ".agents"} {
		t.Run(path, func(t *testing.T) {
			input.AllowedPaths = []string{path}
			if _, err := SealWorkflowMergerInput(input, now); !errors.Is(err, ErrWorkflowConflict) {
				t.Fatal("workflow or scope mutation was granted to the role", err)
			}
		})
	}
}

func nativeMergerResultGuards() map[string]func(*WorkflowMergerResult) {
	return map[string]func(*WorkflowMergerResult){
		"unsupported version":   func(r *WorkflowMergerResult) { r.ProtocolVersion = 2 },
		"old input":             func(r *WorkflowMergerResult) { r.InputHash = workflowHash([]byte("old")) },
		"approval outcome":      func(r *WorkflowMergerResult) { r.Outcome = "approved" },
		"no changes":            func(r *WorkflowMergerResult) { r.ChangedPaths = nil },
		"duplicate paths":       func(r *WorkflowMergerResult) { r.ChangedPaths = []string{"src/service.go", "src/service.go"} },
		"lost requirement":      func(r *WorkflowMergerResult) { r.PreservedRequirements = []string{"source-api"} },
		"duplicate requirement": func(r *WorkflowMergerResult) { r.PreservedRequirements = []string{"source-api", "source-api"} },
		"foreign requirement":   func(r *WorkflowMergerResult) { r.PreservedRequirements[0] = "other" },
		"missing rationale":     func(r *WorkflowMergerResult) { r.ResolutionNotes = nil },
		"empty rationale":       func(r *WorkflowMergerResult) { r.ResolutionNotes[0].Rationale = " " },
		"oversized rationale":   func(r *WorkflowMergerResult) { r.ResolutionNotes[0].Rationale = strings.Repeat("x", 4097) },
		"wrong conflict":        func(r *WorkflowMergerResult) { r.ResolutionNotes[0].ConflictID = "other" },
		"wrong path":            func(r *WorkflowMergerResult) { r.ResolutionNotes[0].Path = "src/other.go" },
		"lost conflict intent":  func(r *WorkflowMergerResult) { r.ResolutionNotes[0].RequirementIDs = []string{"source-api"} },
		"invented regeneration": func(r *WorkflowMergerResult) {
			r.ResolutionNotes[0].RegenerationCommand = []string{"git", "checkout", "--ours", "."}
		},
		"escaped check": func(r *WorkflowMergerResult) {
			r.LocalCheckRefs = []WorkflowEvidenceRef{{Path: "run/other/evidence/check.json", Hash: workflowHash(nil)}}
		},
		"empty question":        func(r *WorkflowMergerResult) { r.Outcome = "needs_decision" },
		"empty failure":         func(r *WorkflowMergerResult) { r.Outcome = "failed" },
		"contradictory success": func(r *WorkflowMergerResult) { r.Failure = "Still unresolved" },
	}
}

// TestNativeMergerResults rejects weakened intent, invented commands and
// internally inconsistent outcomes without interpreting the returned prose.
func TestNativeMergerResults(t *testing.T) {
	_, input, now := nativeMergerInput(t)
	for name, mutate := range nativeMergerResultGuards() {
		t.Run(name, func(t *testing.T) {
			result := nativeMergerResolved(input)
			mutate(&result)

			if _, err := ParseWorkflowMergerResult(nativeMergerJSON(t, result), input, workflowMergerFence(input), now); !errors.Is(err, ErrWorkflowConflict) {
				t.Fatal("invalid repair proposal was accepted", err)
			}
		})
	}
}
