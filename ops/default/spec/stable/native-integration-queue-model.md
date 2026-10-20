# Native Integration Queue Model

Status: accepted

Approved: 2026-10-11; accepted together with the behavioral specification.

Date: 2026-10-11

Product: chr

Specification: [Native Integration Queue and Merger](native-integration-queue.md)

Baseline: [Ticket-Native Work Model](ticket-native-workflow-model.md).

Every record and field below is an accepted additive contract for implementation. Names describe the
required domain model; they do not claim the current executable already stores
these fields. Preserve existing envelope identity, revision guards, storage
ownership, and immutable evidence. Do not introduce a second execution database.

## 1. Authority and Record Placement

| Information | Authoritative owner | Derived views |
| --- | --- | --- |
| Task status, membership, contracts | Existing ticket and set records. | Boards, tracker, SQLite indexes. |
| Admission and current integration phase | Owning invocation's integration extension. | Target queue order and surface status. |
| Canonical repository coordinator | Existing coordinator registration and guarded ownership records. | Checkout/namespace aliases. |
| Candidate and validation | Immutable evidence under the owning invocation. | Current candidate pointer. |
| Merger attempt and budget reservation | Owning invocation extension journals. | Remaining-budget counters and activity events. |
| Submission intent and observations | Existing submission journal, extended with the publication profile. | Queue publication barrier. |
| Verified delivery | Immutable integration receipt. | Ticket delivery fields and slice completion. |
| Closure | Existing owning-set closure receipt. | Set/tracker completion. |

Reuse `WorkflowStore`, `OpsLayout`, guarded file replacement, and current
invocation `extension/chr/` and `evidence/` directories. Suggested new records
live beneath those owned directories as `integration/queue.json`,
`integration/budget.json`, and immutable attempt/candidate evidence. Extend
the existing submission journal and receipt schema rather than creating
alternative journals with equivalent authority.

A target queue is a query over authoritative invocation entries, not a second
editable list. Repository coordinator allocation owns sequence assignment and
cross-namespace discovery. Any lookup index is rebuildable. A multi-file
transition uses the existing durable workflow transaction and replay protocol.

Runtime handles, sockets, raw credentials, private lease tokens, and live
worktree paths are host-owned. Versioned project records retain opaque
identities and evidence references, not machine-specific secret-bearing values.

## 2. Shared Types and Encoding

| Type | Contract |
| --- | --- |
| `RecordID` | Opaque stable ID, unique in the existing project/coordinator domain. |
| `Revision` | Positive monotonically increasing unsigned integer; compare-and-write required. |
| `ObjectID` | Full Git object ID with repository object-format tag; validate negotiated format. |
| `ContentHash` | Existing workflow content hash over versioned canonical encoding. |
| `RefName` | Validated full branch ref; no wildcard, deletion syntax, relative ref, or shell syntax. |
| `RepositoryID` | Normalized forge repository identity bound to canonical Git transport identity. |
| `EvidenceRef` | Relative, owned, bounded reference with content hash; no directory escape or symlink traversal. |
| `Timestamp` | UTC timestamp; deadlines persist through restart. |
| `TokenCount` | Nonnegative integer with checked arithmetic; maximum signed 64-bit value. |

JSON encodings are versioned. Hashes exclude display text and derived fields
unless those fields affect authorization. Unknown required schema versions
block mutation. Unknown optional extension fields survive round trips.
Missing nullable fields are not equivalent to invented zero-value proof.

## 3. Queue Entry

One entry belongs to one integration unit within one invocation. Standalone
tickets and slices use the same integration entry model and publication rules.

| Field | Type / nullability | Rule |
| --- | --- | --- |
| `schema_version` | Positive integer | Extension schema version. |
| `id`, `revision` | RecordID, Revision | Stable entry identity and guarded revision. |
| `invocation_id` | RecordID | Existing owning invocation. |
| `unit_kind` | `slice` or `ticket` | Frozen at admission. |
| `unit_id` | RecordID | Owning slice or independent ticket. |
| `project_id`, `namespace` | Existing identities | Never inferred from checkout basename. |
| `coordinator_id` | RecordID | Registered common owner. |
| `repository_id`, `target_ref` | RepositoryID, RefName | Canonical target key components. |
| `queue_sequence` | Positive integer | Coordinator-allocated, monotonically increasing. |
| `admission_key` | Bounded string | Existing idempotent request binding. |
| `frozen_scope_hash` | ContentHash | Ticket membership, ordering, contract, and invocation scope. |
| `state` | Enum from specification section 6.2 | Current integration phase. |
| `state_revision` | Revision | Changes with every guarded phase transition. |
| `candidate_ref` | EvidenceRef, nullable | Current immutable candidate. |
| `claim_generation` | Existing generation | Fences stale local callbacks. |
| `budget_ref` | Owned journal reference | Persisted invocation/unit consumption. |
| `submission_ref` | Owned journal reference, nullable | Current retained submission, including unknown outcome. |
| `receipt_ref` | EvidenceRef, nullable | Verified immutable delivery receipt only. |
| `blocker` | Structured blocker, nullable | Typed reason, required action, and allowed resume state. |
| `stop_intent` | Existing stop intent, nullable | Preserved until publication reconciliation. |
| `created_at`, `updated_at` | Timestamp | Server-owned timestamps. |

`state` refines the existing invocation phase; it is not an independent authority
over task status. Existing coarse invocation statuses remain readable. The shared
service derives their current projection from the guarded entry and journals.

Examples of projection: queued/capacity waits use existing pending or
waiting-capacity status; human review uses waiting-review; publishing/observing
use waiting-integration with a precise subphase. Integrated unit state advances
the existing invocation frontier and need not terminate an entire set invocation.

### 3.1 Uniqueness and transitions

- Unique nonterminal `(invocation_id, unit_kind, unit_id)` entry.
- Unique `(coordinator_id, queue_sequence)` allocation.
- Current invocation reference must name that exact entry and revision.
- Transition expectations include prior entry revision, claim generation,
  candidate hash where relevant, and required budget reservation.
- Admission and current-entry pointers commit together through the workflow journal.
- A target barrier is derived from retained publishing/observing journals even
  if a stale queue index says the entry is otherwise eligible.

## 4. Canonical Target and Source Ownership

| Record | Required fields | Constraint |
| --- | --- | --- |
| Target binding | `repository_id`, `target_ref`, `coordinator_id`, identity revision | Same remote through different aliases resolves to one target. |
| Source ownership | `unit_id`, canonical PR identity, source repository/ref, workflow owner, rewrite permission, ownership revision | Rewrite permission is explicit, exact, and limited to this unit. |
| Target exclusion | Existing lock/reservation identity, owner generation, active submission ID, uncertainty flag | No concurrent publication or premature lease reclamation. |

Source ownership is not inferred from a branch name prefix. A prefix is useful
for allocation but cannot prove an existing human branch is safe to rewrite.
Ownership transfer requires an explicit guarded action at a quiescent checkpoint.
Cross-repository source/target pairs cannot claim the dual-ref capability.

Do not add a remote generation ref as an implicit distributed lock. The supported
coordinator retains exclusion until previous remote IO is settled. A future
distributed scheme requires its own approved contract and certification.

## 5. Immutable Candidate

Extend the existing integration candidate with the following explicit bindings.
All relevant fields participate in `candidate_hash`.

| Field | Type | Rule |
| --- | --- | --- |
| `schema_version`, `id`, `revision` | Version and identities | Candidate revisions append; never overwrite reviewed evidence. |
| `entry_id`, `invocation_id`, `unit_id` | RecordID | Exact owner linkage. |
| `canonical_pr` | RepositoryID and stable PR identity/URL | One canonical PR for the unit. |
| `source_repository`, `source_ref` | RepositoryID, RefName | Frozen source binding. |
| `original_source` | ObjectID | S, before publication rewrite. |
| `target_ref`, `validated_base` | RefName, ObjectID | Target and B. |
| `prepared_result`, `prepared_tree` | ObjectID | P and exact validated result tree. |
| `method` | `merge`, `rebase`, `squash` | Explicit effective policy; no fallback. |
| `source_manifest_ref` | EvidenceRef | Original ordered commits and intent. |
| `history_manifest_ref` | EvidenceRef | Prepared method proof and original/prepared mapping. |
| `preparation_ref` | EvidenceRef | Owned preparation identity, clean-tree proof, and provenance. |
| `merger_attempt_refs` | List of EvidenceRef | Only attempts contributing to this result. |
| `contract_hash` | ContentHash | Required ticket/set behavior and validation contract. |
| `validation_refs` | List of EvidenceRef | Host-run required combined checks on P. |
| `capability_hash` | ContentHash | Exact current publication profile. |
| `policy_hash` | ContentHash | Review mode, method, required signatures and protection profile. |
| `candidate_hash` | ContentHash | Canonical immutable authorization binding. |

Changing S, B, P, method, contract, preparation, or relevant policy creates a new
candidate revision. Earlier task-level evidence may remain reusable; earlier
combined validation and review do not authorize the new candidate.

### 5.1 History manifest

| Method | Required manifest fields |
| --- | --- |
| Merge | `base`, `original_source`, ordered parent IDs `[B,S]`, prepared tree. |
| Rebase | Ordered source commit IDs; each original/prepared mapping with author, author date, message hash, original/prepared parent, tree, empty-replay decision; declared integration-fix commits. |
| Squash | Original source commit IDs, combined diff hash, attribution policy, parent B, prepared tree and commit. |

Common fields include object format, integrator committer metadata, preparation
nonce/identity, no-op collision count, signature requirements/results, and final
P. A nonce can be retained as provenance without modifying user content or
inserting an unrelated commit. Signing secrets never enter the manifest.

An integration-fix commit is separately attributed to an actionable finding and
reviewed as part of the prepared sequence. It does not erase the ordered source
mapping. Acceptance of this additive proof requires the refinement noted in the
behavioral specification; old receipt verification keeps its original rules.

## 6. Merger Attempt and Decision

### 6.1 Input

| Field | Rule |
| --- | --- |
| `id`, `entry_id`, `candidate_revision`, `claim_generation` | Stable attempt and exact stale-result fence. |
| `input_hash`, `protocol_version` | Canonical immutable input and supported output protocol. |
| `reason` | Preparation conflict, combined-gate regression, or actionable review finding. |
| `source`, `base`, `method`, `history_checkpoint` | Exact S/B and replay position. |
| `intent_refs`, `conflict_refs`, `target_change_refs`, `finding_refs` | Bounded evidence references with hashes. |
| `allowed_paths`, `validation_contract_ref` | Declared scope and repository-owned commands. |
| `capacity_reservation_id`, `budget_reservation_id` | Persisted before launch. |
| `deadline`, `token_allowance` | Finite runtime; optional enforced token grant. |
| `owned_workspace_id` | Opaque host-owned worktree binding, not a user-selected path. |

### 6.2 Result

Result fields are `input_hash`, `outcome`, `changed_paths`, `resolution_notes`,
`preserved_requirements`, `local_check_refs`, `decision_requests`, and `usage`.
Outcome is `resolved`, `needs_decision`, or `failed`. Host acceptance records
result hash, allowed-path verification, method-proof verification, resulting
candidate reference, and rejection reason if applicable.

Record actual changes independently of the returned list. Validate structured
output, bounded size, references, and current generation before accepting it.
The result never contains a forged receipt, approval, or remote mutation grant.

### 6.3 Decision request

Store request ID, owning entry/candidate hash, issue, conflicting requirements,
bounded options, required authority, and resolution status. An answer binds to
the request's current revision. A later source/base change can stale the answer.
No-answer remains waiting; elapsed time is not approval.

## 7. Budget Ledger

Use one invocation-wide ledger with per-unit counters and per-call reservations.
All agent roles share its optional token ceiling. Existing worker, repair,
submission, and closure counters are referenced or migrated, never reset.

| Field | Type and semantics |
| --- | --- |
| `invocation_id`, `revision` | Existing owner and guarded ledger revision. |
| `token_limit` | Nullable positive TokenCount; null means no claimed token ceiling. |
| `tokens_settled`, `tokens_reserved` | Checked totals; their sum cannot exceed a finite limit. |
| `unit_counters` | Per unit: merger launches, drift revisions, submissions; preserve existing worker/review counts. |
| `limits_snapshot` | Admission-time limits and current explicitly granted extensions. |
| `grant_history` | Immutable bounded operator grants with actor, reason, prior revision and timestamp. |
| `reservation_refs` | Agent-call and submission reservations with idempotent IDs. |

### 7.1 Token reservation

A reservation records ID, role, entry/attempt ID, approved maximum tokens,
harness capability revision, input allowance, output allowance, deadline,
state, reported usage, and settlement proof. States are `reserved`, `launched`,
`settled`, `unknown`, or `released_before_launch`.

Token units count reported input plus output once; cached input remains input
and reasoning already included in output is not counted twice. Adapters must
declare compatible accounting and enforce a whole-call allowance, including
any internal multi-turn tool loop. Output-only limits cannot certify a total
token ceiling. Reject unsupported metering before launch.

Reserve before launching under the same coordinator mutation exclusion as other
budget reservations. Simultaneous role admissions cannot oversubscribe the
remaining ceiling. Settle once by reservation ID. Unknown usage keeps the full
reservation charged until reliable reconciliation; a restart cannot release it.
No new pricing, currency budget, or billing service is introduced by this model.

### 7.2 Counter transitions

- Merger launch increments once at durable dispatch reservation.
- Drift increments once per accepted replacement base revision.
- Submission increments only on a new external mutation attempt, not observation.
- A proven pre-launch cancellation may release its reservation; possible launch
  cannot be refunded solely because no response was recorded.
- Existing per-invocation limits also constrain per-unit actions; neither is bypassed.
- Grants extend ceilings by explicit positive amounts; consumed history remains immutable.

## 8. Capability Certificate and Review Grant

| Capability field | Rule |
| --- | --- |
| `mechanism` | `git_atomic_dual_ref` for the required baseline. |
| `repository_id`, `canonical_pr`, `source_ref`, `target_ref`, `method` | Exact scope; never a provider-wide boolean. |
| `source_ownership_revision` | Exact rewrite grant and owning unit. |
| `atomic_source_base`, `exact_result` | True only for implemented, verified semantics. |
| `transport_profile`, `object_format`, `atomic_advertised` | Negotiated transport contract and server capability. |
| `protection_profile_hash`, `review_profile_hash`, `signing_profile_hash` | Current permission and policy bindings. |
| `credential_role_fingerprint` | Nonsecret credential identity/profile binding; never credential bytes. |
| `adapter_revision`, `observation_contract_revision` | Versioned certification identity. |
| `certification_ref`, `certificate_hash` | Profile/method acceptance evidence and current binding hash. |
| `unsupported_reason` | Structured reason when any required property is unavailable. |

Capability is re-observed before submission. A certificate is not permission to
ignore current server rules. Store relevant profile fingerprints and verified
capability, not an unverifiable claim that all forges share the same semantics.

Review grants retain existing actor, mode, candidate hash, policy hash, decision,
finding references and timestamp. Add the capability hash and published-source
rewrite property to the binding. Required native forge review incompatible with
that rewrite is an unsupported profile unless certified server enforcement exists.

## 9. Submission Journal and Receive-Pack Exchange

Extend the existing submission journal with these facts:

| Field | Rule |
| --- | --- |
| `submission_id`, `attempt_number` | Existing idempotent identity and persisted attempt count. |
| `entry_id`, `candidate_hash`, `claim_generation` | Exact dispatch owner and immutable result. |
| `mechanism`, `capability_hash`, `review_grant_ref` | Authorized publication profile and evidence. |
| `source_ref`, `expected_source`, `new_source` | Exact source command S -> P. |
| `target_ref`, `expected_target`, `new_target` | Exact target command B -> P. |
| `fast_forward_proof_ref` | Host proof that B is an ancestor of P. |
| `intent_time`, `dispatch_state`, `dispatch_identity` | Durable IO boundary and surviving dispatcher identity. |
| `advertisement_observation_ref` | Actual advertised refs/capabilities; expected-value comparison before sending commands. |
| `command_manifest_ref` | Both explicit old/new update commands and negotiated atomic flag. |
| `transport_response_ref` | Bounded sanitized report-status or unknown-outcome evidence. |
| `application_state` | `not_started`, `unknown`, `not_applied`, or `git_applied`. |
| `observation_deadline`, `observation_refs` | Persisted reconciliation window and append-only observations. |
| `receipt_ref` | Null until verified Git plus canonical PR delivery. |

The journal states augment existing submission handling. The server cannot be
assumed to store a caller idempotency token. Repeating a Git transaction is not
the idempotency strategy; retained dispatch identity, expected refs, exclusion,
observation, and immutable result proof provide safe recovery.

The transport profile identifies SSH or HTTPS and its certified authentication
integration. Record object format, advertised capabilities, explicit command
manifest, and bounded status parsing results. Do not store authentication
headers, URL passwords, or an unrestricted packet trace containing secrets.

The transport must read the advertisement and compare advertised source/target
to S/B before emitting update commands. Standard CLI success or a post hoc trace
cannot replace that check. Both expected-old/new commands are explicitly part of
one atomic receive-pack exchange. A source already advertised at P cannot be
silently omitted and leave only a target update.

`not_applied` requires definitive transaction rejection plus reconciled refs and
no surviving dispatcher. Missing acknowledgment is `unknown`. Every possible
remote application retains the target barrier until authoritative reconciliation.

## 10. Observation and Receipt

Each observation is immutable evidence containing submission ID, timestamp,
transport/canonical repository identity, actual refs, object/history proof,
canonical PR state/head/merge result, adapter revision, and interpretation.
Retain the admitted `observation_contract` with new observations and receipts;
it must agree with the immutable capability evidence. An absent field preserves
historical exact-P interpretation and cannot authorize source-result attribution.
Interpretation is `pending`, `not_applied`, `applied_unconfirmed_pr`,
`verified`, `conflicting`, or `unobservable`.

### 10.1 Extended receipt

| Field | Rule |
| --- | --- |
| Existing receipt identity and owner | Preserve immutable integration unit and invocation linkage. |
| `schema_version`, `mechanism` | Distinguish new dual-ref receipts from historical schemas. |
| `candidate_hash`, `submission_id`, `capability_hash` | Exact authorized attempt. |
| `canonical_repository`, `canonical_pr` | Same original PR and repository. |
| `original_source` | S; never overwritten with the new PR head. |
| `published_source` | P for the dual-ref mechanism. |
| `validated_base`, `prepared_result` | B and P. |
| `method`, `history_manifest_ref`, `prepared_tree` | Actual method proof and exact resulting tree. |
| `transaction_proof_ref`, `observation_refs` | Git application and canonical PR merged evidence. |
| `actual_merge_result`, `observed_target` | Raw forge merge attribution and independently observed current target. |
| `observation_contract` | Admitted versioned interpretation; never inferred or upgraded on read. |
| `verification_time` | Time the complete receipt was verified. |

A later target descendant is valid if proof of P remains reachable. A later
source update does not rewrite the receipt; retain separate observation history.
An incompatible canonical merge result blocks rather than substituting a
different method. Closing an unmerged PR creates no receipt.

Under explicit GitHub `github_same_pr_git_source_attribution/v2` certification,
`actual_merge_result` may equal original S for rebase while `prepared_result`,
`published_source`, and the receipt's delivered Git result remain exact P. The
ordered replay and any separately attributed integration repairs must bind S to
P, actual target proof must retain P, and source rewrite attribution and canonical
merged PR identity must be independently
verified. Raw S is preserved rather than normalized to P. The receipt,
observation, and retained capability must agree on this contract; legacy v1
evidence still requires a reported merge result of P.

### 10.2 Fanout and closure

Existing finalization journal keys remain `(receipt_id, member_key)`. All slice
tasks receive the same verified receipt/PR with their own membership identity.
Independent tickets receive their own unit receipt. Queue `integrated` follows
completed fanout; a receipt with incomplete fanout stays `finalizing`.

Set closure requires all intended unit receipts, final required gate evidence,
no live owner requiring work, no pending finalization, and no uncertain
publication. Original-delivery and corrective-set closure records are distinct;
neither is inferred from the other's integrated code.

## 11. Queries and Rebuildable Indexes

Required queries include:

- Resolve one integration unit's current entry and immutable evidence.
- List eligible entries for a canonical target by queue sequence and stable ID.
- Find publication barriers from authoritative journals, including unknown IO.
- Find available shared role capacity and unexpired live process reservations.
- Retrieve per-unit counters and invocation-wide reserved/settled token usage.
- Find a retained submission by idempotency key, candidate, or owning unit.
- Locate entries awaiting review, decision, observation, or explicit budget grant.
- Verify all task members share the owning slice's verified receipt.
- Determine whether a set can close without mutating state during status.

Suggested derived indexes use `(repository_id, target_ref, state, queue_sequence)`,
`(invocation_id, unit_id)`, `submission_id`, `candidate_hash`, and `receipt_id`.
Uniqueness is enforced by authoritative guarded records and coordinator
allocation; SQLite uniqueness alone cannot provide execution safety.

Index corruption or deletion must be recoverable from records and journals.
Do not replay mutations merely to answer a query. An interrupted transition is
reported as requiring reconciliation until an authorized mutation repairs it.

## 12. Migration and Validation

1. Add readers that distinguish historical and new schema versions.
2. Validate new fields and ownership before writing new-schema records.
3. Preserve old receipt bytes, consumed counters, unknown submissions, and claims.
4. Add missing queue entries only through an explicit guarded migration/admission,
   not by guessing from a tracker or ticket title.
5. Reprepare/review old candidates before enabling source rewrite publication.
6. Preserve new original/published source fields in every subsequent read/write,
   public projection, fanout, and receipt verifier.
7. Rebuild indexes after the authoritative transition; interrupted index work
   cannot invalidate an already committed domain record.
8. Retain malformed source bytes and expose a typed blocker; never normalize
   unsupported data into a seemingly runnable record.

Cross-record validation rejects mismatched project/namespace/unit, invalid
target identity, duplicate active unit entries, stale generations, out-of-bounds
counter arithmetic, invalid evidence paths, P equal to S, missing method proof,
missing review binding, unsupported profile, and receipts without canonical
merged PR proof.

An integration-fix manifest or source rewrite is not silently interpreted by
old code. A reader incapable of verifying the new schema blocks mutation and
reports the required compatible version.

## 13. Minimal Wire Examples

These fragments omit existing envelopes and required evidence fields. They
illustrate separate identities and budget ownership; they are not runnable
configuration or complete records. Symbolic object IDs require substitution
with validated full IDs in real records.

```json
{
  "mechanism": "git_atomic_dual_ref",
  "original_source": "<S>",
  "validated_base": "<B>",
  "prepared_result": "<P>",
  "published_source": "<P>",
  "application_state": "unknown",
  "receipt_ref": null
}
```

```json
{
  "role": "merger",
  "state": "unknown",
  "approved_max_tokens": 24000,
  "reported_usage": null,
  "reservation_released": false
}
```

## 14. Model-to-Behavior Traceability

| Model owner | Behavioral obligations | Main acceptance group |
| --- | --- | --- |
| Queue entry, target binding, capacity | Specification sections 4..6; INV-01..07, 18, 20..22. | AC-001..015, 058..060, 072, 074. |
| Candidate and history manifest | Sections 7 and 10; INV-08..13, 21. | AC-016..019, 027..033, 041, 049..054. |
| Merger attempt and decision | Section 8; INV-04..05, 13..14, 19..20. | AC-020..035, 067..071. |
| Budget ledger | Sections 8 and 12; INV-18..20. | AC-034..035, 067..071. |
| Capability and review grant | Section 9; INV-08..15. | AC-032, 042..048, 051..055, 077, 080..090. |
| Submission and observation | Sections 9..11; INV-06..10, 16..18, 23. | AC-036..065. |
| Receipt, fanout, and closure | Sections 9, 11, 13, 17; INV-01..02, 23..24. | AC-061..066, 091..095. |
| Migration and surface projections | Sections 12..13; INV-01, 22..24. | AC-072..081, 091..095. |
| Transport, metadata races, shared token reservation | Sections 8..11; INV-08..10, 15..19. | AC-096..100. |

The approved delivery plan must refine this map to evidence owners and concrete
commands. This model does not allocate task tickets, activate slices, or grant
publication access.
