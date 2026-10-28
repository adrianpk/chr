# Slice 12 workflow authority map

Source revision: `d0d65c114cf8ddffcd0492f7fc0b746a832adc0d`.

## Complete method map

The [method map](code-quality-remediation-slice-12-methods.csv) freezes all 908
current `WorkflowStore` methods across 234 files. Each row records its exact
source identity, signature, declaration hash, responsibility, direct calls and
reachable named lock/mutation/recovery entries. The target receiver owner is always
`internal/core/workflow`. T12.2 now relocates the actual type as `workflow.Store`
with all 908 methods. Current identities, signatures and declaration hashes sit
beside the frozen source fields. A future stage policy owner identifies the consumer of
store operations; it does not permit methods of a nonlocal type in another Go
package. Stage policy must become functions/methods of its own types over
consumer-owned operations when its approved extraction task runs.

Responsibilities: 337 coordinator, 136 closeout, 28 transaction, 56 merge,
15 durable-read, 158 integration, 12 durable-mutation and 166 forge/review methods.
These describe operation responsibility. Stage methods still perform durable
reads/writes through the current store. Static reachability records possible
calls, including callbacks in function literals; it does not prove branch order
or that a lock is held at an arbitrary instruction.

The scan uses compiler-resolved declarations and method selections, including
store methods in configuration, guided admission/control, plan resume and ticket
board/document files. Restricting a move to workflow-prefixed filenames would
leave methods behind and violate Go type ownership.

## Persistence and recovery authority

The [boundary map](code-quality-remediation-slice-12-boundaries.csv) records 100
concrete direct IO, locking, mutation and recovery entries in the workflow family
and its transitive root dependencies. It includes native key/operation locks,
configuration reads/upgrade, SQL index transactions, process supervision and
derived tracker projection in addition to the filesystem journal.

- `withRepositoryLock` resolves the Git common directory and locks the original repository-wide workflow lock. Preparation runs before the action; cleanup/unlock errors join the result.
- `withLock` guards coordinator mutation and replays a pending transition while exclusion is held, before the action.
- `commitFiles`/`commitAllocatedFiles` retain one bounded journal of ordered file images with the allocation guard and original hash. Replay validates before/after images, skips already-applied replacements, confirms all images, releases allocation, removes the journal and syncs the lock directory.
- Record creation/write, complete transitions, record/control receipts, claims, budgets and native publication images retain their existing combined commit sets. Stage consumers cannot obtain a second mutable store or commit half an operation independently.
- Submission and upgrade journals keep their original source/identity guards, reservation ownership, acknowledged/unknown outcomes and crash hooks.
- Tracker registration/output remains derived state with its existing guarded replacement, source hashes, stale-result semantics and repair protocol. SQLite workflow/integration indexes remain rebuildable projections; they do not replace authoritative files.

## Relocation prerequisites

The [dependency map](code-quality-remediation-slice-12-dependencies.csv) records
79 direct root declaration dependencies and their actual consumers. Configuration
observation must preserve defaults, explicit presence and raw revision/hash;
review/harness interfaces must preserve retained identity and authority; guided
proofs and ticket/board projections must keep their guards and actual operation
owner. Shared immutable values belong in an existing cohesive value owner.
Configuration/transport behavior must cross a real consumer-owned operation
boundary. The current resolution fields record ten review/cycle value dependencies
owned directly by workflow/state, eleven ticket projection dependencies owned by
core/ticket and one integer comparator supplied by the standard library. The shared parsed configuration and its admission/recovery operations now have
the workspace/config owner, resolving another fifteen dependencies. There are
39 remaining dependencies move with the workflow family, resolving all 79 entries.
The original source identities and hashes remain frozen; current resolutions
identify the actual capability owner. No compile-time bridge is introduced.

The current compiler graph also includes private struct-field uses and all
methods of every relocated type. Many remaining direct dependencies are fringe
workflow declarations that move in the same mechanical unit; they do not all
require another leaf package. Constructors and consumers of private fields move
with the owning type. Legacy read/replace consumers require real guarded file
operation contracts at the existing authority. `ReadLegacyTicketFile` and
`ReplaceLegacyTicketFile` now serve actual legacy consumers and preserve
adopted-ticket refusal, source equality, shared repository/non-Git locks and
complete replacement. No public recovery hook or mutable-store escape is added.

The move covers 2,710 production declarations across 353 files and 1,666 test
declarations across 276 files. Constant declaration groups stay intact. A
compiler-resolved rename removes package stutter from 236 public types/functions,
including implicit embedded-field references; field/tag order and wire values
stay unchanged. Exact token comparisons cover all 4,342 moved declaration groups.
Thirty-three fixture declarations without private owner references remain in
both test packages; tests that
inspect private workflow state move beside its owner. The package import guard
allows its 19 actual neutral dependencies and rejects root core and adapters.
The more-specific state guard continues to reject the coordinator and configuration.
All 100 frozen authority/recovery entries resolve to actual current declarations.

The [signature entry map](code-quality-remediation-slice-12-signatures.csv)
reconciles all 71 assigned baseline rows to 68 current declarations. At entry,
48 rows require signature changes, 22 are legal five-parameter cohesion watches,
and splitWorkflowMarkdown already has its direct shared RequireMarkdown owner
from Slice 9. The original baseline identities remain frozen; final dispositions
will record actual changes, retained contracts and implementation evidence.

## Consumer-owned stage operations

The [port map](code-quality-remediation-slice-12-ports.csv) records five contracts
and their actual production bindings. integrate.Validation validates and accepts
one claimed candidate; merge.Evidence verifies retained input; review.Decisions
accepts one bound review role; forge.Responses retains the untrusted response for
one journaled submission; closeout.Gate executes one bound closing gate. Private
workflow implementations retain claim/prepared/provider or closing/journal
identities. The store implements merger evidence checking directly.

The operation calls retain their original arguments, guards, critical sections,
cleanup and error joins. No store object, opaque handle or raw journal commit
crosses a stage contract. Each stage imports only workflow/state; 20 positive and
negative AST cases enforce that boundary. Policy extraction remains with the
approved T13-T15 tasks, while all 908 Store methods remain at the mutation owner.

Sixty shared protocol types and both evidence-enumeration methods move to state
with exact tokens after qualification and the EvidenceRefs name. Their existing
public type/embedded-field names, JSON order/tags, enum strings and nil/empty
representations remain intact. Six cross-package NativeUnsupported literals now
name the original Condition/Decision fields in the same order with unchanged
expressions. Native policy, authority validation, IO and recovery stay outside
this data owner. Both actual record/role wire fixtures and guarded stage workflows
pass at the new boundary.

## Receipt duplication D08

The original D08 spans are `readWorkflowControlReceipt` and
`readMutationReceipt`. They are readers, despite the source review's proposed
"receipt writer" label. They share contained file read, absence and JSON decode
steps. Their receipt schemas, request/result fields, checksums and admission/
revision/operation guards differ. Any shared decode/read operation must retain
those checks at their current operation owner and preserve absence versus error;
a generic receipt writer cannot be inferred from this duplicate.

T12.4 now shares only the contained read, missing-file distinction and typed JSON
decode in `readWorkflowOperationReceipt`. Both operation readers still validate
their original schema/checksum and request/operator/revision guards before
reporting a confirmed replay. Separate writers and their exact field/tag order
remain at the original transaction owner. Receipt and reservation states use a
named string type with the unchanged `reserved` and `confirmed` encodings.
The control readers' assigned result-shape work remains in Slice 15; no control
public API or admission/recovery contract changes here.

## Validation

The compiler-resolved source graph and exact CSV coverage/hash checks passed.
The frozen entry map is separate from runtime acceptance. The move ledger
records prerequisites and the compiling family relocation with exact declaration
preimages and current resolutions. Focused wire/storage, native Git/filesystem,
guarded legacy operations, application and dependency checks pass. Assigned signature remediation and final candidate validation remain active
Slice 12 work.

## Completed operation contracts

The signature ledger resolves all 71 frozen rows to 47 implemented declarations,
20 individually retained cohesive five-parameter declarations and the prior
shared RequireMarkdown fix. All current declarations satisfy the final parameter
and result shape; no new debt is added. Named observations retain absence/error
values, pagination, byte limits, owner identity and nil/empty behavior.

Coordinator progress/resume and supervised-process state types retain existing
JSON spellings and validation order; 337 literal references are replaced through
compiler-resolved types. Stop child contexts remain explicit at their callers,
with bound claim/cancellation input; heartbeat completion retains cancel-and-wait.
The real consumer reconciliation removes the app control-guard alias. No temporary
compile-time bridge remains. One Store still owns all 908 methods and all existing
locks, journal/recovery entries and atomic acceptance.

## Slice 15 final method inspection

The 908-row method map separates ownership disposition, body-specific basis and
inspected revision from original responsibility/reachability. It records 514
decisions: 428 guarded-persistence methods, 23 closing policies, one contained
evidence adapter, five coordinator functions, one private reader, 21 lifecycle
owners and 35 coordinator mappings. Lifecycle control now has an actual
coordinator and a cohesive Store acceptance operation using the original frozen
request/digest. The minute-bounded lock and complete action/receipt transaction
remain with Store; cancellation precedes continuation and receipt replay starts
no steps. Additional source bases cover recovery, admission, frontier, native
dispatch/delivery evidence and tracker order reads. The remaining 394 Store rows
stay explicitly pending. Reachable locking, IO or a lower method count alone is
not an ownership decision.
