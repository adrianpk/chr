# Slice 12 workflow authority map

Source revision: `d0d65c114cf8ddffcd0492f7fc0b746a832adc0d`.

## Complete method map

The [method map](code-quality-remediation-slice-12-methods.csv) freezes all 908
current `WorkflowStore` methods across 234 files. Each row records its exact
source identity, signature, declaration hash, responsibility, direct calls and
reachable named lock/mutation/recovery entries. The target receiver owner is always
`internal/core/workflow`; the source revision still defines the type and its
methods together in `internal/core`. This is a relocation map, not evidence that
the family has already moved. A future stage policy owner identifies the consumer of
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
39 direct root couplings before family relocation after three shared ticket
text dependencies moved to their existing metadata owner. A temporary compile-time bridge, if necessary, must be recorded with
its explicit approved removal owner by Slice 15.

The current compiler graph also includes private struct-field uses and all
methods of every relocated type. Many remaining direct dependencies are fringe
workflow declarations that move in the same mechanical unit; they do not all
require another leaf package. Constructors and consumers of private fields move
with the owning type. Legacy read/replace consumers require real guarded file
operation contracts at the existing authority, preserving adopted-ticket refusal,
source equality, shared repository/non-Git locks and complete replacement. No
public recovery hook or mutable-store escape is implemented as a prerequisite.

The [signature entry map](code-quality-remediation-slice-12-signatures.csv)
reconciles all 71 assigned baseline rows to 68 current declarations. At entry,
48 rows require signature changes, 22 are legal five-parameter cohesion watches,
and splitWorkflowMarkdown already has its direct shared RequireMarkdown owner
from Slice 9. The original baseline identities remain frozen; final dispositions
will record actual changes, retained contracts and implementation evidence.

## Receipt duplication D08

The original D08 spans are `readWorkflowControlReceipt` and
`readMutationReceipt`. They are readers, despite the source review's proposed
"receipt writer" label. They share contained file read, absence and JSON decode
steps. Their receipt schemas, request/result fields, checksums and admission/
revision/operation guards differ. Any shared decode/read operation must retain
those checks at their current operation owner and preserve absence versus error;
a generic receipt writer cannot be inferred from this duplicate.

## Validation

The compiler-resolved source graph and exact CSV coverage/hash checks passed.
The frozen entry map is separate from runtime acceptance. The move ledger
records shared review values and ticket projections with exact declaration
preimages and current resolutions. Family relocation, stage ports, assigned
signature remediation and final candidate validation remain active Slice 12 work.
