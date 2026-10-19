---
id: TKT-20261017203525
title: "Support Pi alongside Codex for execution and generation"
status: ready
ready_at: 2026-10-19T10:38:33Z
kind: task
severity: unclassified
priority: unclassified
scope: domain
tags: agents, pi, models
source: chat
reported_at: 2026-10-17T20:35:25Z
commits: []
---

# Support Pi alongside Codex for execution and generation

## Need

Allow users to choose Pi as an alternative to Codex in CHR. Keep Codex
supported without forcing existing users to replace their harness or migrate
their setup.

The motivation is model and provider choice: users should be able to use
models supported by Pi, including open-weight models and compatible local or
self-hosted endpoints.

## Desired Outcomes

1. Users can select Codex or Pi for headless ticket execution in Docker.
2. Users can select Codex or Pi for AI conversation and content processing,
   including Dumps, Quiz questions, Specs, Plans and related generation.
3. Users can choose the same harness for both roles or select them
   independently, with the desired model and provider supported by that harness.
4. Harness selection preserves CHR's ticket lifecycle, execution order,
   validation requirements, result semantics and workflow ownership.
5. Existing Codex setups remain usable without requiring Pi.

## Original Capture Boundary

The initial capture recorded the product need without prescribing protocols,
adapters, configuration schemas or container packaging. The subsequently
approved specification below now owns those implementation contracts.

## Approved Design and Delivery

The maintainer approved the specification and requested slicing on 2026-10-19.
The [approved specification](../../spec/stable/codex-pi-harnesses.md) and its
[model companion](../../spec/stable/codex-pi-harnesses-model.md) define the implementation.

This original feature ticket is assigned to Slice 8 for final integrated
acceptance and guidance. Slices 1-7 have their own focused delivery tickets.
This preserves one ticket per slice/PR without creating another parent progress
ledger. Ready status does not claim implementation or authorize live model calls.

## Delivery Assignment

Delivery set: [codex-pi-harnesses](../../plan/codex-pi-harnesses.md).

Slice: 8 of 8 — Cross-interface acceptance and operator guidance.

Order: [canonical tracker](../../tracker/codex-pi-harnesses.md).

Specification: [Selectable Codex and Pi Harnesses](../../spec/stable/codex-pi-harnesses.md).

Model: [Harness selection and retained bindings](../../spec/stable/codex-pi-harnesses-model.md).

Expected branch: `test/slice-8-harness-acceptance`.

Expected PR title: `test(slice-8): verify selectable harnesses and live configuration end to end`.

Expected report: `ops/default/report/slices/codex-pi-harnesses/slice-8-harness-acceptance.md`.

These are planned identities; no implementation branch, PR or report exists
merely because this ticket is ready.

Prerequisite: Slice 7 integrated into dev; implementation authorization for this set.

## Required Outcome

The complete approved contract has exact-candidate evidence, usable setup/recovery guidance and explicitly bounded live Pi proof.

Owners: Shared integration fixtures, User Guide/how-to/reference documentation and final slice report; original feature ticket.

Acceptance: HP-01, HP-02, HP-03, HP-04, HP-05, HP-06, HP-07, HP-08, HP-09, HP-10, HP-11, HP-12, HP-13, HP-14, HP-15, HP-16, HP-17, HP-18, HP-19, HP-20, HP-21.

## Ordered Tasks

| Task | Required work |
| --- | --- |
| T8.1 | Add one bounded cross-interface journey that switches defaults while prior worker/generation work remains active, exercises all four role combinations, resumes retained work and verifies a Codex-only setup and direct-API legacy selection. |
| T8.2 | Complete profile/runtime/credential setup and troubleshooting guidance, operator-visible switching semantics, model limitations and observed-versus-certified usage documentation; verify each HP criterion against its owning slice evidence and close any in-scope gap. |
| T8.3 | After explicit operator authorization, run one real Pi structured generation and one disposable Docker worker task with a fixed profile/model and bounded existing budgets; record exact evidence, run the final development gate on integrated dev and close the original feature ticket only after its PR is merged. |

## Validation

Use deterministic fixtures for routine regressions. The live fixture must be private, disposable and explicitly authorized, with no unrelated remote write or live bot message. Missing credentials/access or an unsupported capability is a real blocker, not a simulated success. Full provider/model, full Go/race and architecture matrices remain nightly.

Focused regression (planned): `go test ./internal/app -run '^TestHarnessJourney$' -count=1 -timeout=3m`.

Focused concurrency regression (planned): `go test -race ./internal/app -run '^TestHarnessJourney$' -count=1 -timeout=3m`.

Before slice review/integration, run `make verify` and `git diff --check`.
These commands are planned checks, not recorded successful executions.
Retain commits, PR, validation and blocker evidence here and in the slice report;
the tracker remains an order map.
