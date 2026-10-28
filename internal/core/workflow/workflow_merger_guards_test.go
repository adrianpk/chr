package workflow

import (
	"errors"
	"strings"
	"testing"
	"time"

	"forge.adrianpk.com/chr/internal/core/contenthash"
	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

func nativeMergerInputGuards() map[string]func(*MergerInput) {
	return map[string]func(*MergerInput){
		"version":               func(i *MergerInput) { i.ProtocolVersion = 2 },
		"identity":              func(i *MergerInput) { i.ClaimGeneration = 0 },
		"escaped attempt":       func(i *MergerInput) { i.ID = "../attempt" },
		"opaque workspace":      func(i *MergerInput) { i.OwnedWorkspaceID = "/user/checkout" },
		"relative evidence":     func(i *MergerInput) { i.IntentRefs[0].Path = "run/native-run-1/evidence/../../escape.json" },
		"foreign evidence":      func(i *MergerInput) { i.IntentRefs[0].Path = "run/other-run/evidence/intent.json" },
		"absolute evidence":     func(i *MergerInput) { i.IntentRefs[0].Path = "/tmp/intent.json" },
		"unbounded attempts":    func(i *MergerInput) { i.RemainingAttempts = 11 },
		"exhausted":             func(i *MergerInput) { i.RemainingAttempts = 0 },
		"unbounded runtime":     func(i *MergerInput) { i.WorkerTimeoutSeconds = 0 },
		"excessive deadline":    func(i *MergerInput) { i.Deadline = i.AdmittedAt.Add(121 * time.Second) },
		"expired":               func(i *MergerInput) { i.Deadline = i.AdmittedAt },
		"future admission":      func(i *MergerInput) { i.AdmittedAt = i.Deadline },
		"clean candidate":       func(i *MergerInput) { i.Conflicts = nil },
		"drift only":            func(i *MergerInput) { i.Reason = "target_drift" },
		"unsupported method":    func(i *MergerInput) { i.Method = "other" },
		"wrong format":          func(i *MergerInput) { i.Base = strings.Repeat("2", 64) },
		"target alias":          func(i *MergerInput) { i.TargetRef = "dev" },
		"same ref":              func(i *MergerInput) { i.TargetRef = i.SourceRef },
		"missing target intent": func(i *MergerInput) { i.TargetChangeRefs = nil },
		"missing source intent": func(i *MergerInput) { i.IntentRefs = nil },
		"weakened requirements": func(i *MergerInput) { i.Requirements[1].Origin = "source" },
		"foreign requirement":   func(i *MergerInput) { i.Conflicts[0].RequirementIDs[0] = "undeclared" },
		"unknown conflict":      func(i *MergerInput) { i.Conflicts[0].Kind = "automatic" },
		"empty gate":            func(i *MergerInput) { i.ValidationCommands = nil },
		"command bound":         func(i *MergerInput) { i.ValidationCommands[0][0] = strings.Repeat("x", 4097) },
		"invalid command":       func(i *MergerInput) { i.ValidationCommands[0][0] = "go\x00" },
	}
}

// TestNativeMergerInput rejects malformed scope, unsupported triggers and
// missing bounded evidence before any role can acquire shared capacity.
func TestNativeMergerInput(t *testing.T) {
	for name, mutate := range nativeMergerInputGuards() {
		t.Run(name, func(t *testing.T) {
			_, input, now := nativeMergerInput(t)
			mutate(&input)

			if _, err := SealWorkflowMergerInput(input, now); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("invalid role input was admitted", err)
			}
		})
	}

	_, input, now := nativeMergerInput(t)

	input.ValidationCommands[0][0] = "false"
	if err := validateWorkflowMergerInput(input, now); !errors.Is(err, workflowstate.ErrConflict) {
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
			if _, err := ParseWorkflowMergerResult(nativeMergerJSON(t, result), input, workflowMergerFence(input), now); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("out-of-scope role change was accepted", err)
			}
		})
	}

	for _, path := range []string{"ops/default/ticket/ready", "ops/default/spec/stable", "ops/default/plan", "ops/default/tracker", ".git", ".chr", ".agents"} {
		t.Run(path, func(t *testing.T) {
			input.AllowedPaths = []string{path}
			if _, err := SealWorkflowMergerInput(input, now); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("workflow or scope mutation was granted to the role", err)
			}
		})
	}
}

func nativeMergerResultGuards() map[string]func(*MergerResult) {
	return map[string]func(*MergerResult){
		"unsupported version":   func(r *MergerResult) { r.ProtocolVersion = 2 },
		"old input":             func(r *MergerResult) { r.InputHash = contenthash.Sum([]byte("old")) },
		"approval outcome":      func(r *MergerResult) { r.Outcome = "approved" },
		"no changes":            func(r *MergerResult) { r.ChangedPaths = nil },
		"duplicate paths":       func(r *MergerResult) { r.ChangedPaths = []string{"src/service.go", "src/service.go"} },
		"lost requirement":      func(r *MergerResult) { r.PreservedRequirements = []string{"source-api"} },
		"duplicate requirement": func(r *MergerResult) { r.PreservedRequirements = []string{"source-api", "source-api"} },
		"foreign requirement":   func(r *MergerResult) { r.PreservedRequirements[0] = "other" },
		"missing rationale":     func(r *MergerResult) { r.ResolutionNotes = nil },
		"empty rationale":       func(r *MergerResult) { r.ResolutionNotes[0].Rationale = " " },
		"oversized rationale":   func(r *MergerResult) { r.ResolutionNotes[0].Rationale = strings.Repeat("x", 4097) },
		"wrong conflict":        func(r *MergerResult) { r.ResolutionNotes[0].ConflictID = "other" },
		"wrong path":            func(r *MergerResult) { r.ResolutionNotes[0].Path = "src/other.go" },
		"lost conflict intent":  func(r *MergerResult) { r.ResolutionNotes[0].RequirementIDs = []string{"source-api"} },
		"invented regeneration": func(r *MergerResult) {
			r.ResolutionNotes[0].RegenerationCommand = []string{"git", "checkout", "--ours", "."}
		},
		"escaped check": func(r *MergerResult) {
			r.LocalCheckRefs = []workflowstate.EvidenceRef{{Path: "run/other/evidence/check.json", Hash: contenthash.Sum(nil)}}
		},
		"empty question":        func(r *MergerResult) { r.Outcome = "needs_decision" },
		"empty failure":         func(r *MergerResult) { r.Outcome = "failed" },
		"contradictory success": func(r *MergerResult) { r.Failure = "Still unresolved" },
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

			if _, err := ParseWorkflowMergerResult(nativeMergerJSON(t, result), input, workflowMergerFence(input), now); !errors.Is(err, workflowstate.ErrConflict) {
				t.Fatal("invalid repair proposal was accepted", err)
			}
		})
	}
}
