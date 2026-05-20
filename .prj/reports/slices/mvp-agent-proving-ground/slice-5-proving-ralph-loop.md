# Slice 5: Minimal Ralph Loop

Status: reviewing
Slicing: MVP Agent Proving Ground
Plan: `.prj/plans/slicing-mvp-agent-proving-ground-plan.md`
Tracker: `.prj/trackers/slicing-mvp-agent-proving-ground-tracker.md`
Branch: `feat/slice-5-proving-ralph-loop`
PR: #76

## Purpose

Add a fixture-scoped Ralph Loop that retries one proving objective after
validation failure.

## Delivered Behavior

`chc` now exposes:

```sh
chc proving ralph --dir <project> --objective <objective>
```

The loop runs a sandboxed Codex invocation, validates the project, and retries
with the previous validation failure embedded in the next worker context.

The loop stops when:

- validation passes;
- the same validation failure repeats;
- agent execution fails;
- validation execution fails;
- max attempts are reached.

## Implementation Notes

Each attempt receives its own objective identity and artifact directory under
the objective artifact root:

```text
<artifact_root>/attempt-N/
```

This preserves context, result, agent log, and validation artifacts per attempt.
Codex does not receive implicit conversation memory. The only cross-attempt
state is the validation failure rendered into the next `context.md`.

## Contracts Added or Changed

New temporary CLI command:

- `chc proving ralph`
- `--dir`
- `--objective`
- `--max-attempts`

New core behavior:

- `RunProvingRalphLoop`
- `ProvingRalphLoopOptions`
- `ProvingRalphLoopResult`
- `ProvingValidationFailureSignature`
- `RenderProvingValidationFeedback`
- `RenderProvingRalphLoopResult`

## Files of Interest

- `cmd/chc/main.go`
- `internal/chc/proving.go`
- `internal/chc/proving_test.go`
- `docs/reference/mvp-agent-proving-ground.md`

## Validation

- `go test ./internal/chc ./cmd/chc`: pass
- `make verify`: pass

## Risks and Follow-ups

Slice 5 proves loop mechanics with injected validation in unit tests. Slice 6
must run the real fixture end to end through Docker and Codex.
