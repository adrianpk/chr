# Quality inventory reconciliation

Baseline: `3a35bdf3445120cad3c47874d59712ef47227531`; production matches the
review baseline `28c3db337cc0b08d249c65c8f2fd68efd5d0966b`.
Owner: T1.1 in [the tracker](../tracker/code-quality-remediation.md).

## Declaration and caller evidence

[Reconciled declarations](code-quality-remediation-reconciled.csv) accounts for
all 121 A, 230 B and 420 C rows. Every row matches its exact source path, line
and declaration name. Three interface-method labels have their explanatory
suffix removed when matching; their source identities remain unchanged.
Normalized signatures, parameter/result counts, syntactic call counts and up
to three candidate call sites are recorded per row. Selector calls are matched
by method name, not resolved receiver type. These are caller inspection entry
points, not whole-program reachability or proof of dead code. Owning slices
must inspect their concrete callers before changing the contracts.

The extended scan includes function types, callbacks and all interface methods.
It finds 122 signatures above five parameters, 232 at five and 423 outside the
result policy. [Additional detections](code-quality-remediation-additional.json)
records the six additional category entries with exact signatures and owners.
They stay in the approved capability slices. Category overlap is not summed as
unique functions. Exactly-five entries remain cohesion review candidates.

## Secondary findings

[Source evidence](code-quality-remediation-secondary.csv) enumerates the 14
files above 800 lines, all eight error implementations without an Error suffix,
11 bare errcheck directives, 130 boolean-signature review candidates and 291
legacy-reference lines. The source review's 23 bare-suppression and 88 control
parameter figures are not reproducible as counts of these source categories.
Do not treat the discrepancy as fixes: inspect each retained source location in
its owner. The boolean search is intentionally broader than behavior-selecting
flags; a bool result, domain datum or compatibility reference is not a defect.
The original review and detections remain byte-for-byte unchanged.

All 14 large files and D01-D11 match the exact pairs and owners in the plan's
secondary inventory and the original review. Their lines and compared blocks
remain present. No duplicate is accepted as semantically identical solely from
text matching. Each pair keeps its governing-contract comparison in the owning
task. None is resolved in this guardrail slice.

| Suspected dead symbol | Production declaration | Live inspection and owner |
| --- | --- | --- |
| service.dumpDirectory | app/service_dump.go:123 | Only service_test.go calls it; Slice 4 |
| service.newQuizModelRequest | app/service_quiz.go:174 | Four service_quiz_test.go calls; Slice 8 |
| telegramGuidedResultExport | app/telegram_guided_document.go:61 | Type alias used only by telegram_guided_document_test.go; Slice 7 |
| telegramHelpReply | app/telegram_help.go:17 | Two telegram_help_test.go calls; Slice 7 |
| timePtr | core/ticket_index.go:565 | Two ticket_index_test.go calls; Slice 11 |
| WorkflowStore.prepareWorkflowTask | core/workflow_execution_prepare.go:18 | Four calls in three workflow test files; delegates to the live mode-aware method; Slice 15 |
| withoutWorkflowPublicationProfile | core/workflow_upgrade_journal.go:279 | config_native_profile_test.go calls it; Slice 15 |

Paths in this table are relative to internal/. Direct source references were
checked across all production and test Go files. Deletion/movement and the
retained live-path regression evidence belong to the recorded later tasks.
The alias is not a function; the wrapper still exercises real production code.

## Evidence limits

This is reconciliation and ownership evidence, not remediation closure.
SQL mapping, durable bytes, semantic duplication, control flags and supported
legacy contracts require their already approved behavior checks. Baseline
counts do not replace the exact candidate's production gate.
