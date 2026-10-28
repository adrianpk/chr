package closeout

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"

	workflowstate "forge.adrianpk.com/chr/internal/core/workflow/state"
)

type gateHost struct {
	prepareError error
	commandError error
	checkError   error
	exitCode     int
	calls        []string
	afterCheck   func()
}

func (host *gateHost) Prepare(context.Context) (string, error) {
	host.calls = append(host.calls, "prepare")

	return "checkout", host.prepareError
}

func (host *gateHost) Run(_ context.Context, directory string, command []string) (workflowstate.CommandResult, error) {
	host.calls = append(host.calls, command[0])

	if directory != "checkout" {
		return workflowstate.CommandResult{}, workflowstate.ErrConflict
	}

	return workflowstate.CommandResult{Command: slices.Clone(command), ExitCode: host.exitCode, Stdout: "retained output", Stderr: "retained diagnostic", Truncated: true}, host.commandError
}

func (host *gateHost) Check(_ context.Context, directory string) error {
	host.calls = append(host.calls, "check")

	if directory != "checkout" {
		return workflowstate.ErrConflict
	}

	if host.afterCheck != nil {
		host.afterCheck()
	}

	return host.checkError
}

type gateCase struct {
	name          string
	host          gateHost
	waived        bool
	cancelAtCheck bool
	calls         []string
	results       int
	completed     bool
	failure       error
}

// TestExecute proves the stage/host boundary preserves partial results, stops at
// the first failure, checks even waived gates and never confirms a receipt.
func TestExecute(t *testing.T) {
	failure := errors.New("host failure")

	cases := []gateCase{
		{name: "completed", calls: []string{"prepare", "first", "second", "check"}, results: 2, completed: true},
		{name: "checkout failure", host: gateHost{prepareError: failure}, calls: []string{"prepare"}, failure: failure},
		{name: "partial command failure", host: gateHost{commandError: failure}, calls: []string{"prepare", "first"}, results: 1, failure: failure},
		{name: "command rejected", host: gateHost{exitCode: 7}, calls: []string{"prepare", "first"}, results: 1, failure: workflowstate.ErrConflict},
		{name: "checkout changed", host: gateHost{checkError: failure}, calls: []string{"prepare", "first", "second", "check"}, results: 2, failure: failure},
		{name: "explicit waiver", waived: true, calls: []string{"prepare", "check"}, completed: true},
		{name: "canceled after checks", cancelAtCheck: true, calls: []string{"prepare", "first", "second", "check"}, results: 2, completed: true, failure: context.Canceled},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) { runGateCase(t, test) })
	}
}

func runGateCase(t *testing.T, test gateCase) {
	t.Helper()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	host := test.host
	if test.cancelAtCheck {
		host.afterCheck = cancel
	}

	execution := Execution{Input: workflowstate.ClosingInput{Schema: "chr-set-closing-input/v1", SetID: "set", Validation: workflowstate.Validation{Required: true, Commands: [][]string{{"first"}, {"second"}}}}, Evidence: workflowstate.ClosingEvidence{Ref: "input.json", Hash: "retained-hash"}}
	if test.waived {
		execution.Input.Validation = workflowstate.Validation{WaiverReason: "Approved explicit waiver"}
	}

	receipt, err := Execute(ctx, execution, &host, &host)
	if !errors.Is(err, test.failure) || !slices.Equal(host.calls, test.calls) || len(receipt.Validation) != test.results {
		t.Fatalf("execution boundary: calls=%v results=%+v err=%v", host.calls, receipt.Validation, err)
	}

	checkGateReceipt(t, execution, receipt, test)
}

func checkGateReceipt(t *testing.T, execution Execution, receipt workflowstate.ClosingReceipt, test gateCase) {
	t.Helper()

	wantInput := execution.Input
	wantInput.Schema = "chr-set-closure/v1"

	if !reflect.DeepEqual(receipt.ClosingInput, wantInput) || receipt.Input != execution.Evidence || receipt.Waiver != execution.Input.Validation.WaiverReason || !receipt.ConfirmedAt.IsZero() {
		t.Fatal("execution changed frozen input/evidence or confirmed the receipt")
	}

	wantOutcome := "closing_validation_failed"
	if test.completed {
		wantOutcome = "completed"
	}

	if receipt.Outcome != wantOutcome {
		t.Fatalf("outcome=%q", receipt.Outcome)
	}

	checkGateOutput(t, receipt.Validation, test.host.exitCode)

	if test.results == 0 && receipt.Validation != nil {
		t.Fatal("empty validation changed from nil")
	}
}

func checkGateOutput(t *testing.T, results []workflowstate.CommandResult, exitCode int) {
	t.Helper()

	for _, result := range results {
		if result.Stdout != "retained output" || result.Stderr != "retained diagnostic" || !result.Truncated || result.ExitCode != exitCode {
			t.Fatalf("partial command evidence lost: %+v", result)
		}
	}
}
