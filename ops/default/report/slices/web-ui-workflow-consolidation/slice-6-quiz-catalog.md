# Slice 6: Scope-driven Quiz Catalog

Status: drafting
Delivery set: web-ui-workflow-consolidation
Plan: `ops/default/plan/web-ui-workflow-consolidation.md`
Tracker: `ops/default/tracker/web-ui-workflow-consolidation.md`
Branch: `feat/slice-6-quiz-catalog`
PR: pending

## Purpose

Deliver the retained-answer catalog and named pending-scope entry in the approved
[plan](../../../plan/web-ui-workflow-consolidation.md) and
[tracker](../../../tracker/web-ui-workflow-consolidation.md). Slice 5 is integrated
through PR #402 at `303db92f10651f5eb36d730e25a742ff78d06cd6`.

## Delivered Behavior

The shared catalog always includes Project Quiz, followed by existing named cases with confirmed answered entries. Only scopes without answers appear as pending. Partial, cancelled and completed work uses retained entries; runtime counters, curated files and skipped entries do not imply an answer.

## Implementation Notes

T6.1 implements a bounded read-only projection through the existing quiz repository. Each item retains its scope, reference, observed session and revision. The focused view re-observes that same identity; catalog reads do not start sessions or invoke generation. T6.2 is pending.

## Contracts Added or Changed

`ReadQuizCatalog` separates answered cards from eligible unanswered named scopes without changing the existing CLI/TUI inventory.

## Files of Interest

`internal/app/service_quiz_catalog.go` and its real repository regressions.

## Validation

`go test ./internal/app -run '^TestQuizCatalog' -count=1` passed, including absent, partial, skipped, cancelled, completed, counter-only and curated-only states, retained answers without runtime, all-answered and new-case membership.

`git diff --check` passed. Final clean-HEAD controller validation follows T6.2 and PR-reference commits.

## Risks and Follow-ups

None recorded.
