# Slice 4: Validation Runner

Status: delivered
Slicing: MVP Agent Proving Ground
Plan: `.prj/plans/slicing-mvp-agent-proving-ground-plan.md`
Tracker: `.prj/trackers/slicing-mvp-agent-proving-ground-tracker.md`
Branch: `feat/slice-4-proving-validation-runner`
PR: #75

## Purpose

Capture compile, lint, test, and verify evidence for one proving objective.

## Delivered Behavior

`chc` now exposes:

```sh
chc proving validate --dir <project> --objective <objective>
```

The command runs the fixture validation commands and writes a Markdown
validation artifact to the objective's declared validation path.

The validation runner records:

- command string;
- exit code;
- stdout;
- stderr;
- aggregate pass/fail result.

## Implementation Notes

Validation runs outside the sandbox for this slice. The goal is evidence
capture that the later Ralph Loop can consume.

The command returns a non-zero error when any validation command fails. This is
expected for the current fixture because the proving objectives are still
unfixed.

## Contracts Added or Changed

New temporary CLI command:

- `chc proving validate`
- `--dir`
- `--objective`

New core behavior:

- `RunProvingValidation`
- `ProvingValidationCommands`
- `RenderProvingValidationResult`

The default proving validation command list is:

- `make compile`
- `make lint`
- `make test`
- `make verify`

## Files of Interest

- `cmd/chc/main.go`
- `internal/chc/proving.go`
- `internal/chc/proving_test.go`
- `docs/reference/mvp-agent-proving-ground.md`

## Validation

- `go test ./internal/chc -run 'TestRunProvingValidation|TestParseProving|TestPlanProving|TestRunProving|TestRenderProving'`: pass
- `go run ./cmd/chc proving validate --dir _workspace/testing/projects/todolist --objective .chc/proving/objectives/TODO-1-fix-task-id-reuse.md`: expected failure with validation artifact
- `go run ./cmd/chc proving validate --dir _workspace/testing/projects/todolist --objective .chc/proving/objectives/TODO-2-complete-task.md`: expected failure with validation artifact
- `make verify`: pass

## Risks and Follow-ups

The validation runner currently uses a fixed command list from the proving
contract. Slice 5 should feed this validation artifact back into the next Codex
attempt.
