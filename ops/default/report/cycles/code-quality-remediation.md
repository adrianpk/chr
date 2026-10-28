# Code quality remediation

Status: prepared; all 16 slices integrated, controller gate pending
Delivery set: code-quality-remediation
Plan: [Approved delivery](../../plan/code-quality-remediation.md)
Tracker: [Task and integration evidence](../../tracker/code-quality-remediation.md)
Baseline: `28c3db337cc0b08d249c65c8f2fd68efd5d0966b`
Final integrated candidate: pending
Final gate evidence: pending

## Delivered source changes

All 16 recorded slices are integrated into dev through canonical PRs #438
through #453. The final implementation commit is
`dbaa4bce2c86daaf33f8b41b2e025be44535a5db`; each PR's exact branch/title, dev base,
merge SHA and ancestry were verified against the controller checkpoint. Slice 16
completed permanent enforcement, residual error naming/cleanup explanations,
source reconciliation and durable architecture references. Its implementation
entries are closed. This cycle remains in progress until the controller's
unchanged full gate passes on the candidate it freezes after metadata preparation.

| Approved finding | Final source disposition |
| --- | --- |
| 121 over-five, 230 five-input and 420 result-policy occurrences | All 771 identities reconciled: 570 conforming named contracts, 200 cohesive five-input cases and one imposed external interface. Permanent whole-source maximum five; no temporary allowance. |
| Root package and Store ownership | Real capability packages, five workflow stages and shared state/configuration owners. All 908 source Store methods assessed individually; 777 retain actual guarded persistence. No root forwarding facade or workflow type alias; 22 finite consumer ports. |
| 14 large files | Operation/ownership splits with original/current declarations and per-capability regressions. Cohesive schema literals retain exact justification. |
| Six wide values | Cohesive record/projection sections preserve flattened field/wire/SQL contracts. TUI uses one pointer model with 17 substates and captured asynchronous inputs. |
| Eleven duplicate pairs | Shared identical rules, with separately recorded schema, approval, source, mutation and presentation differences. Permanent production dupl threshold 150. |
| Seven dead production symbols | Removed or confined to test helpers; live request/renderer/explicit-mode paths remain. |
| Command wiring and SQL | 134 constructor mappings use typed factory injection; fourteen SQL binding/scan boundaries across eight tables retain explicit columns, binds and scans. |
| Minor candidates | All 454 source-frozen secondary rows assessed, including 130 boolean and 291 compatibility references. Eight error names corrected. All 23 baseline bare cleanup directives have resource-specific reasons, including twelve test-fixture sites. |

The [complete final reconciliation](../../plan/code-quality-remediation-slice-16-findings.md)
links every category to its original identities, implementation commits, actual
owners and merged task checks. The source review and detections remain preserved
for controller-owned final closure. Their original counts are not silently
rewritten to the broader final signature-check scope.

## Compatibility and measured exclusions

Record and wire fixtures retain field names/order, optional/NULL/time/bool values,
schema versions, error identity, content hashes, budgets, request/receipt identity
and atomic source/recovery guards. The final capability map and glossary preserve
the distinction between ticket-native authority and forge publication.

Supported retained protocols and commands have individual current consumers and
explicit retirement conditions. Distinct receipt writers retain separate schemas,
identity validation and guarded transactions. The exact Bubble Tea Update method
is the single external result exception. Existing grammar/dispatch/schema
complexity exceptions retain their specific source rationale. None is a temporary
migration bridge or an unknown category labeled as excluded.

The final benchmark on `50d562611bfb5d8e725ff62da389fd9c2e372818` retains the same
49,832-byte model. Idle and resize updates allocate zero bytes; focus uses 1,096
bytes/nine allocations, compared with baseline 229,376/172,032/345,162 bytes for
idle/resize/focus. Plain/shell View allocation counts remain 703/1,404. Every
sample and the latency comparison are in the linked reconciliation. These
measurements do not justify blanket pointer changes for 4,374 heuristic warnings.

## Validation and closure

Owning slice reports retain exact task and PR-head checks, failures/corrections,
controller receipts and verified merge references. Slice 16 controller receipt
`52XOJZYFDDGRBWZ2VM6OGGREK4` passed ordered make verify, 34 boundary roots, four race roots and diff check on exact merged
`dbaa4bce2c86daaf33f8b41b2e025be44535a5db` in 61.665 seconds, with no failures,
skips or owned fixture leaks. Its benchmark retains its separate measured
revision. Controller-owned integrated validation remains pending.

No main promotion, tag, release, mirror publication, successor plan or additional
slice is part of this report. Final closure must record the controller's exact
integrated commit, ordered gate commands, evidence identity and outcomes before
closing the delivery-set status and final acceptance ledger.

## Prepared gate prerequisites

The frozen configuration retains exactly this full gate:

```sh
make verify && go test ./internal/app -run '^TestGuidedUseCaseParityNative$' -count=1 -timeout=3m && CHR_WEB_BROWSER=/usr/bin/chromium go test ./internal/app -run '^Test(GuidedUseCaseWebBrowser|WebQuizBrowser)$' -count=1 -timeout=3m
```

The configured Chromium executable exists and reports version 151.0.7922.173.
Go 1.27.1, make, Python 3 and Git are present. All three exact Go test entrypoints
exist. The selected native/browser tests create their own temporary project,
Git/SQLite, local HTTP and isolated browser-profile fixtures, then own cleanup.
No persistent fixture needs to remain running for the controller. This
preparation does not run any selected test or change the frozen gate.
