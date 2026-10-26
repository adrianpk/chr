---
id: TKT-20261026084824
title: "Quality slice 4: Typed commands and central application operations"
status: reviewing
kind: cleanup
severity: unclassified
priority: normal
scope: domain
tags: code-quality, remediation, slice-4
source: review
reported_at: 2026-10-26T08:48:24Z
commits: [c6da9e52f81ab6a2260dfbaa3e374c6dce3e6bfa, 398dbf67a8aa92ac0bd45a9c6aac313fec6b3d88, 1f0c7c7c2f8237bef5653c975623a55d4b77b36a, f40d62d4775628abf58d527d93e11e8fae557a7e, 42db97f8af41b95f61120415afdc16931891b47c]
started_at: 2026-10-26T10:38:54Z
branch: refactor/slice-4-quality-commands
reviewing_at: 2026-10-26T11:35:18Z
ready_at: 2026-10-26T08:57:09Z
---

# Quality Slice 4: Typed commands and central application operations

Delivery set: code-quality-remediation
Plan: [Single-cycle plan](../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../tracker/code-quality-remediation.md)
Spec: [Requirements](../../spec/code-quality-remediation.md)
Model: [Compatibility](../../spec/code-quality-remediation-model.md)
Source: [Review findings](../../report/code-quality-review-2026-10-26/report.md)

## Problem and evidence

This review unit owns R1, R2, R4, R7, R8, R10 as scoped in the plan, including every assigned declaration in [the baseline inventory](../../plan/code-quality-remediation-inventory.csv). Baseline: `28c3db337cc0b08d249c65c8f2fd68efd5d0966b`. It is part of one 16-slice delivery set, not a separate runtime work invocation.

## Work

1. **T4.1:** Replace root Metadata dependency lookup with one typed command constructor value and method actions. Cover every command constructor, not just the reported 15; remove ignored dependency and output parameters and the service locator only after all production callers move.
2. **T4.2:** Mechanically split ticket.go and author.go by actual subcommand responsibility. Split maintenance.go and service_work.go where their independent operations justify it. Keep moves, signature changes and semantic extractions in distinct task commits.
3. **T4.3:** Move reusable ticket loop, review, merge and start-preflight orchestration behind the existing central service in service_ticket files. Use one cohesive ticket-loop value while passing context explicitly. Route existing CLI, TUI, Web and Telegram callers through the same operation where semantics match.
4. **T4.4:** Remove service.dumpDirectory and the other app-owned test-only symbols after checking production call paths; redirect retained tests to live operations. Resolve newTelegramServeCommand, serviceFromCommand and owned error-name/unparam findings without deleting meaningful regression coverage.
5. **T4.5:** Finish every signature row assigned to this slice, including command entrypoints, gate/report adapters and ticket-loop results. Preserve CLI flags, output, cancellation, finite attempts and host-owned publication authority.
6. **T4.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

## Acceptance criteria

1. Command wiring is compile-time typed; adapters contain parsing and presentation; ticket orchestration has one central operation owner; no metadata service locator remains.
2. Every assigned baseline entry has a precise fixed or retained-contract disposition, linked to task commit and actual validation.
3. Existing authoring/execution/approval/recovery and durable/wire/hash contracts are preserved as required by the specification and model.
4. The slice compiles and passes its focused selection, make verify and git diff --check; its report and tracker identify the actual evidence boundary.

## Dependencies and delivery

Predecessor: verified integration of Slice 3 into dev.
Planned branch: `refactor/slice-4-quality-commands`.
Expected PR title: `refactor(slice-4): inject command dependencies and share ticket operations`.
Expected report: `ops/default/report/slices/code-quality-remediation/slice-4-commands.md`.

The tracker owns task execution; this ticket mirrors verified slice state. The maintainer approved this delivery set on 2026-10-26; this ticket is ready subject to its recorded predecessor. Use the normal in_progress/reviewing/solved transitions and record real commits and PR only when they exist. No independent ticket branch or extra PR is added beside the slice PR.

## Validation selection

CLI command-construction and ticket loop/start/review/merge regressions with real local Git where affected; make verify.

All commands above are planned. None ran as implementation validation during this planning task.

Implementation: all five implementation tasks are committed; [Slice 4 report](../../report/slices/code-quality-remediation/slice-4-commands.md) and tracker hold contract and check evidence. Canonical PR reference and final exact-HEAD controller validation follow before handoff.
