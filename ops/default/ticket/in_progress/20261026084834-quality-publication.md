---
id: TKT-20261026084834
title: "Quality slice 14: Forge publication and review stages"
status: in_progress
kind: cleanup
severity: unclassified
priority: normal
scope: domain
tags: code-quality, remediation, slice-14
source: review
reported_at: 2026-10-26T08:48:34Z
commits: [42253fe479a1a47d9a465c7359bba1a0df3f3586, ee47cb41ee02ee2df60d475a29128783b0e9ef37, 09431ee3293fdd5a279e5d3b7170dbf7037070a6, ca61d8a6fadc0d1c98685dbed31d7239196a0b90, 09b9e4c262dee11a83b7a34d5c886c2ceadb5907, 3e053ee5515e9943ab6ddac1a8024f5fb3f73d13, 932090eb2531dba22e53b2ec57264c38c49d9429, 23453a58cbb8f780065ea29f780e4dd37b76d0fc, 187a9ec4c15fef69e137c158fe574163aaae5497]
started_at: 2026-10-28T04:43:13Z
branch: refactor/slice-14-quality-publication
ready_at: 2026-10-26T08:57:09Z
---

# Quality Slice 14: Forge publication and review stages

Delivery set: code-quality-remediation
Plan: [Single-cycle plan](../../plan/code-quality-remediation.md)
Tracker: [Task evidence](../../tracker/code-quality-remediation.md)
Spec: [Requirements](../../spec/code-quality-remediation.md)
Model: [Compatibility](../../spec/code-quality-remediation-model.md)
Source: [Review findings](../../report/code-quality-review-2026-10-26/report.md)

## Problem and evidence

This review unit owns R1, R2, R3, R10 as scoped in the plan, including every assigned declaration in [the baseline inventory](../../plan/code-quality-remediation-inventory.csv). Baseline: `28c3db337cc0b08d249c65c8f2fd68efd5d0966b`. It is part of one 16-slice delivery set, not a separate runtime work invocation.

## Work

1. **T14.1:** Extract provider publication, submission and pull-request interaction into internal/core/workflow/forge. Extract review decisions into internal/core/workflow/review. Keep the provider-independent workflow coordinator separate from these adapters.
2. **T14.2:** Distinguish ticket-native authority from forge-native publication: native is not a global synonym for forge. Rename only symbols whose actual meaning is proven; retain established domain vocabulary and persisted spellings where it remains precise.
3. **T14.3:** Preserve exact target/head binding, source identity, permission/protection checks, policy revision, explicit human approval and host-owned publication. An open PR or worker exit cannot become confirmed delivery.
4. **T14.4:** Convert assigned submission/review tuples and repeated parameters into precise stage values. Preserve idempotent dispatch, uncertain outcome reconciliation, duplicate acknowledgement and transactional receipt ownership.
5. **T14.5:** Update service/adapter callers and glossary, finish local state/flag/error/suppression work, and verify package dependencies and direct consumer interfaces without exposing store internals.
6. **T14.6:** Run the required focused checks and make verify; record results, update the ticket/tracker, produce this slice report, and verify its dev integration before marking it delivered.

## Acceptance criteria

1. Forge transport and review policy are separate from orchestration, without changing approval authority or delivery proof; no blanket native-to-forge rename occurs.
2. Every assigned baseline entry has a precise fixed or retained-contract disposition, linked to task commit and actual validation.
3. Existing authoring/execution/approval/recovery and durable/wire/hash contracts are preserved as required by the specification and model.
4. The slice compiles and passes its focused selection, make verify and git diff --check; its report and tracker identify the actual evidence boundary.

## Dependencies and delivery

Predecessor: verified integration of Slice 13 into dev.
Planned branch: `refactor/slice-14-quality-publication`.
Expected PR title: `refactor(slice-14): isolate forge publication and review decisions`.
Expected report: `ops/default/report/slices/code-quality-remediation/slice-14-publication.md`.

The tracker owns task execution; this ticket mirrors verified slice state. The maintainer approved this delivery set on 2026-10-26; this ticket is ready subject to its recorded predecessor. Use the normal in_progress/reviewing/solved transitions and record real commits and PR only when they exist. No independent ticket branch or extra PR is added beside the slice PR.

## Validation selection

Existing native provider/transport permission and review fixtures through local HTTP/Git; existing review/publication receipt tests; make verify.

Task checks for review and forge extraction pass: 42 real review/preparation/
application-control checks, 17 real submission/recovery checks and focused races,
with pinned lint and finite-debt validation. Exact commands are in the canonical
report. The first review completion unit's evidence remains recorded separately. The
publication value conversion adds 35 real native/legacy checks and six focused
race/wire/dependency checks.
`make verify` and final exact-head controller validation remain pending.

Slice 13 PR #450 is verified merged into dev at
`53b2adb23b9b1d20529d00bbe3b32b3be5990b4e`. Slice 14 is active. Fourteen pure
policy declarations have independent review/forge owners and matching mechanical
token evidence; completion and dispatch use consumer-owned ports with host
claim/process/write authority retained. Fifty assigned rows are implemented, including all category A/C rows;
11 category B rows remain pending. Exact debt is 24. Remaining PR/native observation, binding
policy and state/flag/error/glossary work must finish before a PR is ready.

The publication value conversion preserves dispatch permission, transport,
canonical observation, revision fences and receipt evidence as distinct outcomes.
Native-path/publication barriers retain partial/error behavior and cannot be read
as successful delivery. All direct consumers retain their current authority and
assertions; real IO/recovery and focused races are recorded in the report.
