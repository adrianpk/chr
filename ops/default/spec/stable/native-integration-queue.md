# Native Integration Queue and Merger

Status: accepted

Approved: 2026-10-11; maintainer approved the specification and authorized Looper execution.

Date: 2026-10-11

Product: chr

Model: [Native Integration Queue Model](native-integration-queue-model.md)

Related contract: [Ticket-Native Work Execution](ticket-native-workflow.md)
and [Ticket-Native Work Model](ticket-native-workflow-model.md).

Existing delivery: [Plan](../../plan/ticket-native-workflow.md) and
[Tracker](../../tracker/ticket-native-workflow.md).

This document specifies one cohesive corrective slice set. A slice set contains
multiple ordered slices; each slice remains one PR and each task remains one
ticket. The delivery boundaries are mapped to the approved
[corrective plan](../../plan/native-integration-queue.md) and
[tracker](../../tracker/native-integration-queue.md). Execution uses their
committed slice identities and task-ticket bindings.

## 1. Required Outcome

chr must finish approved ticket work through a real, guarded publication path.
The user can run one agent at a time or enable parallel independent work without
changing the correctness contract. chr owns integration scheduling, preparation,
conflict resolution, validation, publication, and recovery. Its core does not
depend on a forge's native merge queue.

A bounded merger role resolves understandable code and semantic conflicts in an
isolated checkout. Trusted host code decides whether its result is valid and
authorized. The role cannot approve itself, push remote refs, invoke a remote
merge, bypass protections, or mark tickets delivered.

One corrective slice set must deliver the missing production path, the merger
role, operating surfaces, recovery, and certification. Returning an unsupported
capability for every real provider is safe behavior but cannot satisfy delivery.

### 1.1 Scope

- Native queue behavior for standalone tickets and slice integration units.
- Serial default operation and bounded parallel independent implementation.
- Per-target publication exclusion, including writers outside chr.
- Exact prepared-result publication for `merge`, `rebase`, and `squash`.
- An explicit, certified Git publication profile for existing forge adapters.
- Intelligent conflict repair with finite attempts, time, and optional tokens.
- Durable uncertainty handling, receipts, ticket fanout, and closure barriers.
- Shared CLI, TUI, Web, daemon, and manual-harness service behavior.
- Reuse and completion of the existing ticket-native workflow delivery.

### 1.2 Boundaries

This work does not introduce a distributed consensus system, a new mandatory
service, speculative execution of later ordered slices, arbitrary concurrent
edits inside one slice, a new forge adapter, release publication, or automatic
promotion to `main`. Separate chr clones or hosts must use one registered
coordinator; unsupported coordinator connectivity blocks mutation.

External Git writers need not register with chr. Their updates are handled by
remote ref guards and renewed preparation. A Git server without the required
transaction or an incompatible branch policy remains explicitly unsupported
until an equally strong adapter mechanism is certified.

## 2. Existing Work and the Actual Gap

### 2.1 Audited baseline

The integrated baseline is `67c6edecf7b7b0f32474f04e6e34a7d24f530201` on
canonical `dev`, inspected on 2026-10-11. Eleven slices of the existing delivery
have integrated. The final assurance slice remains incomplete. Two assurance
increments exist on its retained branch, but are not integrated into this
baseline: `4f10f1c60de1f164f727d37b4df5c02343ab6e11` and
`44e8145aa618edfc51d5efb08330c641b02f7e58`.

| Area | Existing evidence | Treatment in this specification |
| --- | --- | --- |
| Ticket hierarchy, ordering, and frozen selection | Accepted ticket-native specification and workflow records. | Reuse; do not rematerialize delivered tickets. |
| Filesystem authority and revisions | [WorkflowStore](../../../../internal/core/workflow_store.go). | Extend the existing owner and journals; SQLite stays derived. |
| Shared capacity and role waits | [Capacity slots](../../../../internal/core/workflow_capacity_slot.go) and [role tests](../../../../internal/core/workflow_capacity_role_wait_test.go). | Reuse for executor, preparation, merger, reviewer, and closing roles. |
| Isolated integration preparation | [Git preparation](../../../../internal/core/workflow_preparation_git.go). | Extend exact-result preparation and commit provenance. |
| Candidate, validation, and review binding | [Preparation](../../../../internal/core/workflow_preparation.go). | Preserve immutable binding; add merger provenance and publication profile. |
| Integration orchestration | [Publication](../../../../internal/core/workflow_publication.go). | Extend the current service path, not a second orchestration engine. |
| Submission intent and bounded retries | [Submission preparation](../../../../internal/core/workflow_submission_prepare.go). | Reuse; distinguish Git application from forge observation. |
| Production publication adapter | [Provider publication](../../../../internal/core/workflow_publication_provider.go). | Missing: current capability has no atomic source/base or exact-result guarantee; submission refuses publication. |
| Merger role | Existing review repair and preparation hooks. | Missing as a complete integration-conflict contract with bounded semantic repair. |
| Composed assurance | Two retained assurance increments listed above. | Reconcile and reuse after approved delivery mapping; neither is current integrated certification. |

The publication blocker affects a single runner as well as parallel runners.
Reducing worker count does not implement a guarded publisher. Existing claims,
locks, journals, and tests also do not certify a real remote publication path.

### 2.2 Relationship to the accepted contract

The accepted contract already requires atomic source/base guards, exact method
proof, renewed validation after base movement, a single target publisher,
bounded parallel workers, and observation before delivery. This document makes
those requirements executable and adds the merger protocol and certification.

This accepted contract adds two explicit refinements. The Git transaction may rewrite the
CHR-owned canonical PR source ref to the exact prepared result. Original source
identity remains immutable evidence; published PR source identity is recorded
separately. Existing records that assume those identities are always equal must
be extended before this mechanism can be enabled. Rebase proof may additionally
contain an explicitly attributed integration-fix commit after the preserved
source sequence; such a commit must be visible in validation and review.

Approval covers those refinements, the supported branch-policy profile,
merger limits, and the model migration together. This specification refines the
earlier accepted contract for the corrective delivery; its historical evidence
and receipt schemas remain intact.

## 3. Decisions and Vocabulary

| Decision | Contract |
| --- | --- |
| Serial operation | `work.max_active_workers: 1` remains the default. No idle merger agent is launched. |
| Parallel operation | Independent eligible lanes may execute concurrently within the shared configured pool. |
| Ordered work | Tasks inside a slice and slices inside a set preserve their existing order and receipt barriers. |
| Queue ownership | chr owns queue state and scheduling; provider queues are not prerequisites. |
| Publication ownership | One active publication per canonical repository and full target ref. Different targets may proceed independently. |
| Role execution | The merger is a role of a configured harness, not necessarily a separate permanent agent or process. |
| Default method | Preserve existing `rebase` default and explicitly selected alternatives. No automatic method substitution. |
| Forge coverage | Certify the existing Forgejo and GitHub adapters for declared profiles; core rules remain provider neutral. |
| Remote changes | Human and other-machine updates are normal concurrent events, including when chr is configured serially. |
| Coordination | Reuse the registered repository coordinator; do not create independent per-checkout queues. |
| Delivery accounting | Preserve task tickets and one PR per slice. Granularity follows real deliverable tasks, without a fixed ticket-count target. |

Terms used below:

- **Integration unit:** one slice or one independent ticket.
- **Lane:** the isolated work identity already claimed by the invocation.
- **S:** admitted original source commit of the canonical PR.
- **B:** observed target commit against which preparation is performed.
- **P:** exact prepared commit that passes the combined gate and review.
- **Candidate revision:** immutable binding of S, B, P, method, contract, policy,
  preparation provenance, capability, and required evidence.
- **Target key:** normalized canonical repository identity plus full target ref.
- **Queue sequence:** monotonically assigned durable admission order, not task order.
- **Publication barrier:** exclusion retained while publication is active or uncertain.
- **Merger attempt:** one bounded harness execution against one immutable conflict input.
- **Verified receipt:** immutable evidence of actual Git delivery and canonical PR merge.

## 4. Invariants

| ID | Invariant |
| --- | --- |
| INV-01 | Tickets and their owning workflow records remain execution authority. |
| INV-02 | A slice has one canonical PR and one verified delivery receipt shared by its tasks. |
| INV-03 | At most one chr owner may mutate an integration unit at a time. |
| INV-04 | All agent roles consume the same repository worker-capacity pool. |
| INV-05 | No role shares another role's mutable checkout. |
| INV-06 | One target key has at most one active or uncertain publication. |
| INV-07 | Local exclusion never substitutes for remote source/base guards. |
| INV-08 | Every publication atomically guards S and B and publishes exactly P. |
| INV-09 | Every target update is independently proven fast-forward from B. |
| INV-10 | A no-op source ref update is never counted as a source guard. |
| INV-11 | Changed source, base, method, contract, or relevant policy invalidates the candidate's publication authorization. |
| INV-12 | Required validation runs on the exact prepared checkout and records P's tree and history proof. |
| INV-13 | The merger's confidence or narrative cannot replace validation or review. |
| INV-14 | No worker or merger owns publication credentials or receipt authority. |
| INV-15 | Branch protection, required signatures, and current publication policy are never bypassed. |
| INV-16 | A lost response does not justify another publication attempt. |
| INV-17 | Unknown publication retains the target barrier until reconciliation. |
| INV-18 | Lease expiry cannot reclaim a capacity slot or target barrier while the previous process or remote outcome is unresolved. |
| INV-19 | Restart, rebase, candidate replacement, and queue re-entry never refill consumed budgets. |
| INV-20 | The default serial mode launches at most one agent role at a time. |
| INV-21 | Later ordered work cannot become eligible before the preceding receipt barrier. |
| INV-22 | Read-only status never resumes, repairs, closes, or retries work. |
| INV-23 | A successful command, closed PR, or passing test alone never marks delivery. |
| INV-24 | Existing integration history and unrelated user work remain intact. |

## 5. Responsibilities and Trust Boundary

| Actor | Permitted responsibility | Prohibited responsibility |
| --- | --- | --- |
| Shared application service | Admission, policy, scheduling, checkpoints, user commands. | Surface-specific alternative state machines. |
| WorkflowStore and core domain | Revisions, claims, queues, journals, budgets, receipts, fanout. | Trusting a cache or harness output as authoritative state. |
| Executor | Implement the selected ticket in its owned lane. | Publishing or closing delivery records. |
| Merger | Repair an integration candidate in its isolated checkout and explain decisions. | Changing scope, target, method, budget, remote refs, or approval. |
| Reviewer | Evaluate exact candidate and evidence under configured review policy. | Reusing approval for a changed candidate. |
| Git preparer/publisher | Deterministic object preparation, exact transaction, ref observation. | Choosing semantic resolutions or changing merge method. |
| Forge adapter | Resolve identity, report policy capability, observe canonical PR facts. | Inventing merged metadata or treating closed as merged. |
| Human operator | Approve, pause, stop, amend scope, grant more budget, resolve ambiguity. | Implicit authorization through silence or status requests. |

Use existing `internal/core` and `internal/app` owners. Add small ports for new
capabilities rather than a provider dependency in queue or merger domain code.
Agent tools receive only local working permissions required by their role.
The trusted publication boundary retains credentials; their values never enter
prompts, command records, user events, or evidence artifacts.

## 6. Queue and Scheduling Contract

### 6.1 Admission and identity

1. Resolve canonical repository, target ref, owning coordinator, unit, and frozen invocation.
2. Validate prerequisites, method, source ownership, bounds, and policy profile.
3. Reuse an existing nonterminal entry for the same unit and invocation.
4. Allocate one queue sequence using the existing guarded workflow transaction.
5. Persist the entry and its owning invocation reference atomically.
6. Reject overflow before accepting work; never evict an accepted entry.

The queue extends the existing admitted frontier. It is not a second selection
system. Admission of an entire set does not enqueue its later slices as eligible
units. The next slice becomes eligible only after its predecessor is verified.

### 6.2 States

| State | Meaning | Legal next states |
| --- | --- | --- |
| `queued` | Admitted and eligible; waiting for scheduling. | `preparing`, `paused`, `cancelled`, `blocked` |
| `preparing` | Snapshot and deterministic preparation in progress. | `merging`, `validating`, `queued`, `blocked`, `paused`, `cancelled` |
| `merging` | Bounded merger role owns local conflict repair. | `validating`, `queued`, `blocked`, `paused`, `cancelled` |
| `validating` | Host runs the combined gate on P. | `reviewing`, `merging`, `queued`, `blocked`, `paused`, `cancelled` |
| `reviewing` | Exact-candidate review in progress. | `waiting_review`, `ready_to_publish`, `merging`, `queued`, `blocked`, `paused`, `cancelled` |
| `waiting_review` | Human decision is required; no worker slot retained. | `ready_to_publish`, `merging`, `queued`, `blocked`, `paused`, `cancelled` |
| `ready_to_publish` | Current candidate is authorized; waiting for target exclusion. | `publishing`, `queued`, `blocked`, `paused`, `cancelled` |
| `publishing` | Durable intent exists; external operation may be in flight. | `observing`, `queued`, `blocked` |
| `observing` | Reconcile Git and forge outcome; target barrier retained. | `finalizing`, `blocked` |
| `finalizing` | Verified receipt exists; replay ticket fanout. | `integrated`, `blocked` |
| `blocked` | Typed reason requires reconciliation or explicit operator action. | Resume state recorded with the reason; never an unconditional retry |
| `paused` | Explicit pause before irreversible publication. | `queued`, `cancelled` |
| `cancelled` | No publication remains possible or unresolved. | Terminal |
| `integrated` | Receipt verified and fanout complete. | Terminal |

`publishing` may return to `queued` only after authoritative evidence proves the
transaction was not applied, the input changed, and no operation remains in
flight. A stop during publication records intent to stop after reconciliation;
it cannot label potentially delivered work cancelled.

### 6.3 Ordering and capacity

- Within a target queue, use admission sequence and stable unit ID as tie-breaker.
- Preserve existing task/set ordering and prerequisite exclusions before queue ranking.
- Skip paused, blocked, or human-review entries without holding publication exclusion.
- Preparation may overlap independent implementation within the shared worker limit.
- Deterministic Git IO is not an extra agent slot; configured agent validation,
  repair, review, and closing roles do consume a slot.
- Limit speculative prepared-but-unpublished candidates per target to two by default.
- Before launching another candidate, prefer an eligible head candidate that needs
  its last validation or review role. Prevent implementation from starving integration.
- Unknown publication blocks publication to that target, not unrelated targets
  or read-only operations. Other work may proceed within capacity and queue bounds.
- A repeatedly drifting unit reaches its drift limit and becomes visibly blocked;
  it cannot occupy the queue indefinitely through invisible retries.

### 6.4 Cross-checkout and cross-host behavior

Canonical remote aliases, worktrees, namespaces, and URL spellings must resolve
to the same target key and common capacity owner. Different credentials for the
same repository do not create distinct queues. Separate clones or hosts must
register with one coordinator or receive an explicit coordination error before
launch or publication. This specification does not equate remote CAS with a
distributed queue or lease service.

## 7. Candidate Preparation and Method Proof

### 7.1 Snapshot

Observe the canonical PR's repository, source repository, full source ref, S,
target ref, B, open/merged state, and policy revision. Fetch objects by verified
identity. Preserve the original source manifest, required ticket contracts,
existing task evidence, and current target tree. Reject ambiguous refs, missing
objects, source ownership changes, fork boundaries, and mismatched PR targets.

### 7.2 Deterministic preparation

Prepare in an owned isolated worktree. Do not mutate the worker lane, user's
checkout, original source evidence, or remote refs during preparation.

| Method | Required prepared history | Required evidence |
| --- | --- | --- |
| `merge` | P has parents B and S, in that order, and the resolved combined tree. | Parent identities, result tree, original source manifest, conflict resolutions. |
| `rebase` | Replayed linear source commits descend from B in original order; P is the last replayed commit. | Original-to-prepared commit map, ordered trees, preserved authors/messages, explicit empty-commit decisions. |
| `squash` | P has one parent B and the complete resolved change tree. | Original source manifest, combined diff, author attribution policy, result tree. |

Preserve author identity, author date, and original message for rebased commits;
record the integrator as committer. Rewriting invalidates old commit signatures.
If policy requires signatures on prepared commits, sign them through the trusted
host before validation; absence of the approved signing path blocks capability.
Do not claim that copied signature text validates a rewritten commit.

Reject a merge-containing source history for `rebase` unless a separately
implemented and certified preservation policy exists. Do not silently squash it.
An empty replayed commit is retained as an empty replay with its original message
and mapping; a truly empty integration unit requires explicit operator resolution.
Detect prior duplicate/cherry-picked effects without dropping work silently.

### 7.3 Exact-result sealing

The prepared result must satisfy all of the following before review:

1. P exists, its complete object graph is available, and B is its ancestor.
2. P differs from S; both publication ref updates will be non-no-op.
3. Method-specific history and original-to-prepared mapping are verified.
4. The working tree is clean after host-owned commit preparation.
5. All required validation evidence names P, its tree, and the combined contract.
6. Merger changes, including edits outside conflict markers, are fully recorded.
7. The candidate hash binds every relevant input and output.

A clean rebase that would reuse S still creates replayed commits with recorded
integration committer metadata, preserving authors and messages. If the result
collides with S, advance the recorded committer timestamp and reprepare before
validation, at most three times. Do not add an unrelated empty commit or modify
the source content solely to manufacture a guard.

## 8. Merger Role

### 8.1 Invocation conditions

Launch the role only when deterministic preparation reports conflicts, a combined
gate demonstrates an integration regression requiring repair, or the reviewer
returns a bounded actionable repair request. A clean, passing candidate needs no
merger-agent call. Target movement alone first triggers deterministic rebase.

### 8.2 Input contract

Provide a versioned, size-bounded input containing unit identity, candidate
revision, S, B, selected method, owned worktree, original commit manifest, task
intent and acceptance criteria, conflict paths, relevant target changes,
previous findings, permitted validation commands, and remaining attempt/time/
token budget. Include file/diff references and hashes rather than unlimited
repository dumps. Reject stale inputs before acquiring an agent slot.

### 8.3 Resolution rules

- Read both source intent and intervening target changes before choosing a result.
- Preserve independently required behavior from both sides where compatible.
- Inspect semantic interactions even after textual markers disappear: API
  contracts, call sites, configuration defaults, migrations, validation, and tests.
- Regenerate generated files and lockfiles using repository-owned tools when
  available; preserve their source inputs and record the command.
- Treat binary files, submodule updates, and LFS conflicts as unsupported unless
  the repository declares a deterministic resolution and validation procedure.
- Never apply blanket `ours` or `theirs`, delete a failing requirement, weaken a
  gate, edit ticket scope, or select another merge method to obtain a green result.
- Stop with a precise decision request when preserving both intents requires a
  product choice, destructive migration, unknown ownership, or new scope.
- A role may propose tests needed to demonstrate its repair, but cannot replace
  the approved gate with a test that merely asserts its own chosen output.

### 8.4 Output and host verification

Return a bounded structured result: `resolved`, `needs_decision`, or `failed`;
input hash; changed paths; per-conflict rationale; preserved requirements;
local validation observations; unresolved questions; and reported usage.
Treat repository content and returned prose as untrusted input.

The host verifies ownership, allowed paths, input freshness, diff scope, and
method provenance. It commits the prepared result through its trusted Git port,
runs the combined gate itself, and obtains candidate-bound review. Agent claims
that tests passed are supplementary evidence only.

For rebase repair, retain an isolated checkpoint for the specific replayed
commit and its predecessors. Apply each resolution at the correct replay step,
then continue replay; do not collapse the sequence into an undocumented squash.
Review-driven semantic repair is mapped to the owning original commit when
unambiguous; otherwise it requires a declared attributed integration-fix commit
or an operator decision. Such a fix commit is part of the reviewed history proof.

### 8.5 Attempts, time, and tokens

At most three automatic merger launches are allowed per integration unit by
default. Deterministic preparation after target movement has a separate default
limit of five drift revisions. The existing default submission limit remains
three. These counters survive replacement candidates and restarts.

Use the existing worker timeout as the maximum merger runtime. Persist attempt
reservation before launch. Pause/stop uses the existing owned-process control
contract; timeout does not release a slot while the child remains alive.

An optional positive token budget is invocation-wide and includes executor,
merger, reviewer, and closing calls. Reserve a bounded per-call allowance before
launch, enforce it through the harness, and settle actual usage afterward. A
finite token budget requires a harness with enforceable limits and usage
reporting. Without token metering, finite attempts and time remain mandatory;
do not present that mode as enforcing a token ceiling.

Unknown usage retains the reservation. Restart never refunds it. Only an
explicit bounded budget grant can extend exhausted work. Reusing the same agent
implementation in another role does not bypass accounting.

## 9. Publication Mechanism

### 9.1 Required baseline: atomic Git updates of both refs

The required initial implementation is an opt-in `git_atomic_dual_ref` profile
for a CHR-owned source branch in the same canonical repository as the target.
The prepared result is published through one server transaction:

```text
source ref: S -> P
target ref: B -> P
```

Both refs must participate in the transaction. The source rewrite is an explicit
property of the selected profile, not a hidden side effect. The worker lane and
immutable original source manifest retain S. The canonical PR remains the same PR.

The trusted publisher supplies explicit expected old object IDs for each ref:

```text
git push --atomic \
  --force-with-lease=refs/heads/chr/work/example:<S> \
  --force-with-lease=refs/heads/dev:<B> \
  <canonical-remote> \
  <P>:refs/heads/chr/work/example \
  <P>:refs/heads/dev
```

This illustrates the intended transaction and is not a sufficient publisher
implementation by itself. The publication port must own the receive-pack
advertisement and command exchange, using a narrow tested transport component.
It checks that advertised refs still equal S and B, then explicitly sends both
expected-old/new ref commands with the negotiated atomic capability. It aborts
before sending update commands if either advertised value differs. Changes after
advertisement are rejected by the server's expected-old checks.

The port may reuse an existing transport library or a tightly scoped helper;
it must not rely solely on standard CLI no-op filtering or inspect a trace only
after publication. Use host Git for object generation and proof where practical.
Support the SSH and HTTPS transports used by the existing adapters through
host-owned authentication. Bound advertisement, packet, object-transfer, and
report-status processing by invocation deadlines and size limits. Persist any
uncertainty after commands might have been sent. Reject unnegotiated atomic
support, malformed/partial reports, and redirected canonical repository identity.
Transport implementation must have an explicit protocol/security review.

Git subprocesses receive argument vectors and validated full refs. Credentials
are redacted. No shell or implicit remote-tracking lease supplies the guard.

Explicit leases alone can authorize non-fast-forward changes. Therefore the host
must independently verify `B` is an ancestor of `P`; only the owned source ref
may be rewritten. No blanket `--force`, ref deletion, target recreation, wildcard
refspec, or fallback non-atomic push is permitted.

### 9.2 No-op guard trap

A standard Git client can omit a ref whose desired value equals the advertised
old value. Specifying a lease for that ref does not prove it participated in the
remote transaction. Consequently P must differ from S before validation and the
publisher must verify that both intended update commands are sent.

If either advertised ref already equals P at submission time, do not report a new guarded
publication. Enter observation/reconciliation. Prove an earlier application or
record a conflicting external update. Never skip one update and send the other.

### 9.3 Capability and policy admission

Enable the profile only after verifying:

1. Canonical source and target are refs in one repository.
2. The source ref is explicitly CHR-owned and rewrite-authorized for this unit.
3. Receive-pack advertises atomic ref transactions.
4. Current credentials may update both refs under the declared branch policy.
5. The target update is fast-forward and all required server protections remain enforced.
6. Any required signed-commit path is available for P and rewritten history.
7. The adapter can recognize this publication as the same canonical merged PR.
8. Current review mode permits host-bound approval of P under this profile.
9. The exact method has a certified result-proof and observation implementation.

The baseline policy profile is **CHR-managed candidate review with explicitly
authorized fast-forward target pushes**. It is not compatible by assumption with
a rule requiring every target update to use a forge merge API, or requiring
forge-native approval of a source SHA that the atomic transaction rewrites.
Reject such combinations unless server-enforced semantics and adapter
certification prove equivalence. Never temporarily disable protections, use
administrator bypass, or silently reinterpret required forge-native approval.

A capability report binds repository, PR, full refs, ownership revision, method,
server capability, policy fingerprint, observation contract, credential identity
fingerprint, and adapter version. Fingerprints reveal no credential secrets.
Recheck immediately before intent/submission; any relevant change invalidates
authorization. Actual receive-pack still enforces current remote permissions.

### 9.4 Same-PR observation and receipts

Git refs and forge metadata are not one atomic database. A successful push first
proves the Git transaction, then requires canonical PR observation. Do not call
an unguarded merge API afterward, manufacture a merge flag, or close the PR and
interpret that as merged.

The adapter must tolerate its certified forge's metadata lag and eventual source
head update from S to P. It records both original source S and published source P.
The method is established by actual prepared history, not an ambiguous forge
display label for a manually detected merge.

The Git transaction guards refs, not simultaneous edits to forge metadata. If
another actor closes or retargets the PR after preflight, Git application may
already be irreversible. Observe that condition as an inconsistent publication,
retain the barrier, and require reconciliation; never claim full atomicity of
PR state, rollback the target automatically, or report a clean failed attempt.

Issue a receipt only after proving:

- Actual target contains P, or a later descendant still contains its verified proof.
- Method-specific parents, ordered history, and result tree match the candidate.
- The canonical repository and original PR identity match the retained binding.
- The forge reports that PR merged and supplies a compatible merge result.
- Published source movement to P is attributed to the retained transaction or
  equivalent authoritative evidence; a later source update is recorded separately.

Compatibility is pinned to the admitted, versioned observation contract. The
existing v1 contracts require the forge's reported merge result to equal P.
The explicit GitHub `github_same_pr_git_source_attribution/v2` contract also
permits the original PR to attribute a rebase to S, provided the complete verified
replay and any separately attributed integration repairs bind original S to exact
P, the Git transaction and target proof establish P, and canonical source
metadata agrees with the independently observed source
rewrite. A later source head requires retained transaction attribution and is
recorded separately. The original PR must still be reported merged. Retain the
forge's raw result S separately from delivered P; never substitute P into forge
metadata or infer a merge from reachable commits alone. Other results, methods,
providers, or missing proof remain incompatible. Certification changes invalidate
prior admission, and historical v1 evidence keeps its original interpretation.

If Git application is proven but forge merge metadata remains missing, retain
`observing` and the target barrier. After the observation deadline, expose a
typed reconciliation blocker. Do not roll back delivered target history, retry
the push, or mark tasks done. Certification must demonstrate that the supported
profile converges for each claimed method.

### 9.5 Other providers and mechanisms

A provider-native API can be another mechanism only if it atomically guards S
and B, delivers exact P or an equivalently proven immutable result, preserves the
same PR, and implements observation. A source-only `sha` check is insufficient.
No implementation may switch methods or substitute a native merge queue to
claim compliance.

Unsupported profiles return structured reasons before task execution when
detectable at admission, or before publication when policy has changed. Include
the unsupported condition and the required operator decision. Do not recommend
disabling safeguards as an automatic recovery action.

## 10. Revalidation and Concurrency

| Change | Required response |
| --- | --- |
| Target B changes before preparation | Fetch current target and prepare a new revision. |
| Target changes after validation or approval | Invalidate combined evidence and approval; reprepare, validate, and review. |
| Source S changes before publication | Freeze publication, verify source ownership and new work scope, then admit a new candidate revision or request a decision. |
| Source changes during atomic publication | Entire transaction fails; target remains unchanged by this attempt. |
| Target changes during atomic publication | Entire transaction fails; source remains unchanged by this attempt. |
| Method, required command, contract, or review policy changes | New candidate binding and required review; no reuse of old authorization. |
| Only an unrelated queue projection changes | Preserve candidate evidence; refresh projection without redoing domain work. |
| Another chr unit integrates first | Rebase on the confirmed new target; keep valid task-level evidence but rerun the combined gate. |
| Another target has uncertain publication | This target may proceed if coordinator capacity and ownership allow. |
| External writer integrates the same work | Verify exact equivalence and canonical PR facts before adopting a receipt; otherwise request reconciliation. |

Required combined validation cannot be replaced by a conflict-free rebase check.
Use fetch/object caches and unchanged task evidence to avoid repeating unrelated
implementation. Do not optimize away a required gate on the new combined tree.

The target publication lock covers final observation, authorization checks,
intent, submission, and uncertainty reconciliation. It is not held across normal
implementation, merger-agent deliberation, or human review. If the final check
finds a moved base, release the lock before bounded re-preparation.

Claim generations fence stale local callbacks before any intent is created.
Once IO may have started, ownership cannot transfer into another publication
until the prior outcome is reconciled. A new owner must not reuse identical
source/base leases while an old publisher might still execute.

## 11. Failure and Recovery

| Failure boundary | Durable facts | Recovery |
| --- | --- | --- |
| Before queue admission commit | No accepted entry. | Retry same idempotency key. |
| After admission, before preparation | Entry and frozen invocation. | Resume scheduling within retained budgets. |
| Merger launch or timeout | Reserved attempt, process identity, usage allowance. | Reap/observe owned process; settle or retain usage; no hidden replacement. |
| Validation failure | Exact candidate and failing command evidence. | Bounded actionable repair or typed block. |
| Approval becomes stale | Old approval retained as history. | Prepare/review current candidate; no automatic approval transfer. |
| Intent persisted, IO definitely not started | Submission ID and guard. | Same-attempt dispatch only after proving no surviving dispatcher. |
| Process dies around dispatch | Intent; IO may have started. | Observe; do not infer not-applied from missing response. |
| Atomic push definitively rejected | Rejection and independently observed refs. | Reprepare after proved guard conflict, or block permissions/capability. |
| Transport disconnect or timeout | Unknown application. | Retain target barrier and observe with the same submission ID. |
| Git applied, forge metadata pending | P and transaction proof; no receipt yet. | Poll within deadline, then explicit reconciliation. |
| Receipt persisted, fanout interrupted | Immutable receipt and finalization journal. | Replay missing member updates exactly once. |
| Target subsequently advances | Receipt proof remains reachable. | Preserve delivery; no false invalidation from a later descendant. |
| Target rewritten to remove P | Previously valid receipt, conflicting current observation. | Expose history divergence; never silently erase or recreate delivery. |
| Stop requested after possible application | Stop intent and submission journal. | Reconcile delivery before terminal stop state. |

Observation does not consume another submission attempt. A remote rejection is
not enough to infer all refs stayed unchanged if transport output is incomplete;
observe actual refs and server transaction evidence. Exponential polling remains
bounded by a persisted deadline. Read-only status reports uncertainty without
performing repair.

## 12. Configuration and Operating Surfaces

Existing defaults and explicitly configured policies are preserved. Proposed
new keys are requirements for implementation, not current executable options.

| Setting | Default | Bounds and behavior |
| --- | --- | --- |
| `work.max_active_workers` | `1` | Existing `1..8`, shared by every agent role. |
| `work.ready_queue_capacity` | `64` | Existing `1..1024`, counts unfinished admitted lanes once. |
| `work.max_integration_attempts` | `3` | Existing `1..10`, retained across candidate revisions. |
| `work.integration.publication_profile` | `disabled` | Explicit opt-in `git_atomic_dual_ref`; no unsafe legacy fallback. |
| `work.integration.max_prepared_per_target` | `2` | `1..8`, further bounded by admitted capacity. |
| `work.integration.max_drift_revisions` | `5` | `1..20`, retained per unit/invocation. |
| `work.integration.max_merger_attempts` | `3` | `1..10`, retained per unit/invocation. |
| `work.integration.observation_timeout_seconds` | `180` | `10..3600`; deadline expiry blocks, never retries publication. |
| `work.agent_token_budget` | Unset | Optional positive invocation-wide ceiling; requires enforced per-call grants and usage reporting. |
| Merger runtime | Existing worker timeout | Snapshot at admission; cannot exceed the invocation's remaining time. |

Source ownership is a validated workflow binding, not a configurable wildcard
permission. Setup explains source rewrite behavior and tests the profile before
work is admitted. No hidden enablement during upgrade. Policy edits affect new
admission and current publication authorization as specified; they never refill
existing counters.

All execution surfaces invoke the same service commands and expose the same
observed facts:

- Integration unit, target, canonical PR, phase, queue position, and wait reason.
- Current source/base/result identities when useful for diagnosis.
- Active role, last meaningful progress, attempts remaining, and budget status.
- Whether publication is not started, in flight, unknown, Git-applied, or verified.
- A precise human decision, review request, exhausted-budget grant, or unsupported-profile blocker.
- Actions authorized by state: approve exact candidate, pause, stop, bounded
  budget grant, resolve decision, or explicit reconciliation.

Do not expose raw leases, private tokens, credential fingerprints, or internal
paths as normal product decisions. A serial user sees one active agent and a
clear integration phase. A parallel user can inspect each lane and target queue.
No surface may implement its own retry loop or optimistic delivery status.

## 13. Migration and Existing Delivery

1. Add versioned extension records and strict readers without rewriting old receipts.
2. Rebuild queue indexes from authoritative records; do not infer execution from trackers.
3. Preserve current claims, unknown submissions, ownership, and exhausted budgets.
4. Treat old candidates lacking dual-ref proof as ineligible for this mechanism
   until explicitly re-prepared and re-reviewed.
5. Keep original and published source identities distinct in new receipts;
   historical receipts retain their earlier schema and verification rules.
6. Preserve the eleven integrated slices and their task receipts.
7. During approved planning, map the two retained assurance increments to the
   corrective work once; rebase/adapt their checks rather than duplicating tickets.
8. Leave the old assurance invocation blocked until an approved transition names
   its continuation or replacement and preserves retained evidence.
9. Complete the outstanding original live acceptance and final closure gates
   after the corrective set delivers the real publisher.

Specification approval does not certify the retained assurance branch or
close its tickets. Execution starts through the approved corrective plan and
Looper configuration. Final integration and ticket closure remain distinct
verified state transitions.

## 14. Acceptance Contract

Acceptance IDs describe observable obligations. They do not require one test or
one ticket each. Reuse meaningful existing checks and composed scenarios. Record
the command, exact revision, fixture/profile, result, and evidence reference for
every claimed case. A mock-only result cannot certify remote Git or forge behavior.

### 14.1 Queue, ordering, and capacity

| ID | Scenario and required result |
| --- | --- |
| AC-001 | With default configuration and queued work, at most one executor/merger/reviewer agent process is active at a time. |
| AC-002 | With capacity two, independent eligible lanes execute concurrently in isolated checkouts. |
| AC-003 | Two task tickets in one slice preserve task order even when capacity exceeds one. |
| AC-004 | A later slice does not start before the previous slice's receipt barrier. |
| AC-005 | Repeated admission with the same key returns the retained entry without duplicating queue capacity. |
| AC-006 | Queue overflow rejects new admission while preserving every previously accepted entry. |
| AC-007 | Worktree, namespace, and remote aliases share the canonical capacity and target key. |
| AC-008 | Two chr publishers for one target cannot both reach external dispatch concurrently. |
| AC-009 | Two independent targets can publish concurrently when capacity and policy allow. |
| AC-010 | Human-review waiting releases the worker slot and does not hold the publication lock. |
| AC-011 | An unknown publication blocks its target while unrelated targets remain usable. |
| AC-012 | Separate clones or hosts without registered coordination fail before launching mutable work. |
| AC-013 | Eligible integration receives capacity under sustained new implementation admission. |
| AC-014 | A blocked or paused unit does not prevent the next eligible independent unit from progressing. |
| AC-015 | Reducing capacity waits for live roles without killing or overcounting them. |

### 14.2 Preparation, methods, and merger

| ID | Scenario and required result |
| --- | --- |
| AC-016 | A clean merge produces parents B/S and the exact validated result tree. |
| AC-017 | A clean rebase preserves ordered commit intent, authors, and messages and descends from B. |
| AC-018 | Squash produces one combined commit on B with traceable source attribution. |
| AC-019 | A clean candidate passes without launching a merger agent. |
| AC-020 | A textual conflict invokes the bounded merger with source intent and target changes. |
| AC-021 | Textually clean but semantically incompatible code is rejected by the combined gate/review and repaired or blocked. |
| AC-022 | Resolution preserves required behavior from both branches and records per-conflict rationale. |
| AC-023 | An ambiguous product decision stops with a precise question and no remote mutation. |
| AC-024 | Blanket conflict-side selection, weakened gates, or ticket-scope edits are rejected. |
| AC-025 | Merger output with stale input hash or paths outside permitted scope is rejected. |
| AC-026 | Agent-reported green checks do not bypass host execution of the required gate. |
| AC-027 | Rebase conflict repair preserves the replay sequence and original-to-prepared mapping. |
| AC-028 | Integration-fix commits are explicit in the manifest and candidate review. |
| AC-029 | Generated files and lockfiles use declared regeneration; unsupported binary/submodule/LFS repair blocks. |
| AC-030 | A merge-containing source cannot silently become a squash under rebase policy. |
| AC-031 | Empty replay commits retain attribution and mapping; duplicate effects do not silently discard intent. |
| AC-032 | Required signature policy blocks unsigned rewritten commits and accepts a certified signed path. |
| AC-033 | Candidate preparation leaves user checkouts and the worker lane unchanged. |
| AC-034 | Merger timeout retains process ownership until the child is confirmed stopped. |
| AC-035 | Merger/review repair budgets remain consumed across target movement and restart. |

### 14.3 Atomic publication and external writers

| ID | Scenario and required result |
| --- | --- |
| AC-036 | A real receive-pack transaction updates owned source S->P and target B->P together. |
| AC-037 | Target movement before push rejects the complete transaction and preserves the source. |
| AC-038 | Source movement before push rejects the complete transaction and preserves the target. |
| AC-039 | A hook moves source after advertisement; the non-no-op dual-ref transaction rejects and target remains B. |
| AC-040 | A fixture proves an up-to-date source ref can be omitted; product code refuses to treat its lease as a guard. |
| AC-041 | P equal to S is re-prepared before validation or explicitly blocked; never partially published. |
| AC-042 | A server without atomic support fails capability with no fallback push. |
| AC-043 | A target that is not an ancestor of P is rejected even when explicit leases would permit force. |
| AC-044 | A fork PR across repositories is rejected by the dual-ref mechanism before publication. |
| AC-045 | A human-owned source ref cannot be rewritten through this profile. |
| AC-046 | Permission or branch-protection denial leaves both refs unchanged by the attempt. |
| AC-047 | A policy requiring incompatible native merge/review semantics is reported unsupported without bypass. |
| AC-048 | A policy/profile change after review invalidates authorization before dispatch. |
| AC-049 | A serial chr run handles an external target writer by re-preparing and revalidating. |
| AC-050 | Two parallel independent lanes integrate in turn with renewed base validation. |
| AC-051 | No merge-method substitution or legacy unguarded API call occurs on failure. |
| AC-052 | The prepared P and actual published tree/history match for every selected method. |
| AC-053 | A source or target already at P in receive-pack advertisement triggers observation before sending commands, not a partial new push. |
| AC-054 | Explicit leases use full refs and object IDs, independent of stale tracking refs or background fetches. |
| AC-055 | Ref names, remote URLs, and process arguments cannot inject shell or wildcard operations. |

### 14.4 Recovery, budgets, and delivery

| ID | Scenario and required result |
| --- | --- |
| AC-056 | Crash after intent but around dispatch resumes observation before any resubmission. |
| AC-057 | Lost successful push response is recovered from Git and forge facts without a second publication. |
| AC-058 | Unknown outcome retains the target barrier across restart and lease expiry. |
| AC-059 | A stale owner cannot publish after replacement; replacement waits while earlier IO may still occur. |
| AC-060 | Stop during possible application reconciles delivery before recording terminal stop. |
| AC-061 | Git-applied but metadata-pending publication remains visible and does not mark tickets done. |
| AC-062 | A closed-unmerged PR never creates a delivery receipt. |
| AC-063 | Receipt verification distinguishes original S from published source P. |
| AC-064 | A later descendant target preserves valid receipt reachability. |
| AC-065 | A target rewrite removing P produces explicit history-divergence evidence. |
| AC-066 | Interrupted receipt fanout replays exactly once to every task in the slice. |
| AC-067 | Exhausted submission, drift, or merger allowance blocks with its persisted reason. |
| AC-068 | Explicit bounded budget extension resumes the existing unit without erasing consumption. |
| AC-069 | Enforced token grants account for executor, merger, reviewer, and closing calls together. |
| AC-070 | Missing token-limit/usage capability rejects finite-token mode before an agent launch. |
| AC-071 | Unknown usage keeps its reservation; restart cannot refund tokens. |
| AC-072 | Read-only status leaves files, claims, budgets, tickets, and remote state unchanged. |
| AC-073 | Malformed or partially migrated records block safely and preserve the original bytes. |
| AC-074 | SQLite/index rebuild reproduces queue views without changing authoritative execution state. |
| AC-075 | No credential value, private owner token, or secret-bearing remote URL enters public evidence. |

### 14.5 Surfaces and real certification

| ID | Scenario and required result |
| --- | --- |
| AC-076 | CLI, TUI, Web, daemon, and manual service callers report the same retained phase and blocker. |
| AC-077 | Approval from every applicable surface binds the exact candidate; stale approval fails consistently. |
| AC-078 | A serial user sees one active agent and a clear integration/decision/completion status. |
| AC-079 | Parallel status identifies each lane and target queue without conflating their progress. |
| AC-080 | Setup explains and verifies the source-rewrite profile before work admission. |
| AC-081 | Upgrade preserves explicit old configuration and leaves the new publication profile disabled. |
| AC-082 | Forgejo real repository: merge publishes exact P and the same canonical PR is observed merged. |
| AC-083 | Forgejo real repository: rebase publishes exact P and the same canonical PR is observed merged. |
| AC-084 | Forgejo real repository: squash publishes exact P and the same canonical PR is observed merged. |
| AC-085 | GitHub real repository: merge publishes exact P and the same canonical PR is observed merged. |
| AC-086 | GitHub real repository: rebase publishes exact P and the same canonical PR is observed merged. |
| AC-087 | GitHub real repository: squash publishes exact P and the same canonical PR is observed merged. |
| AC-088 | Real certified profiles demonstrate external-source and external-target contention, not only clean merges. |
| AC-089 | Real policy-denial profiles demonstrate unsupported behavior without automatic protection changes. |
| AC-090 | All certification entries bind server/version or API profile, credentials role, branch policy, method, PR, and exact commit evidence. |
| AC-091 | Existing ticket-native composed assurance continues to pass after adaptation to the real publisher. |
| AC-092 | Final gate on the integrated development revision includes required strict lint, tests, focused race/concurrency checks, and live acceptance evidence. |
| AC-093 | Every completed slice has a verified canonical PR merge and correct task-ticket finalization. |
| AC-094 | The original delivery closes only after its remaining assurance and closure obligations are verified. |
| AC-095 | The corrective set closes only with reconciled queues, no unknown submissions, and a verified set closure receipt. |

### 14.6 Transport and boundary races

| ID | Scenario and required result |
| --- | --- |
| AC-096 | Source becomes P between application preflight and receive-pack advertisement; the publisher aborts before commands rather than relying on CLI no-op filtering. |
| AC-097 | The guarded exchange, contention checks, and unknown-outcome recovery pass over both supported SSH and HTTPS transports. |
| AC-098 | Malformed capability packets, partial report-status, missing atomic negotiation, and repository-changing redirects cannot yield success or an automatic retry. |
| AC-099 | An external close or retarget of the PR around Git application produces explicit reconciliation without false receipt or target rollback. |
| AC-100 | Concurrent agent-role admissions cannot reserve more than the finite invocation token ceiling. |

## 15. One Corrective Slice Set: Delivery Boundaries

These ordered work packages map to the committed delivery plan and tracker.
Each executable slice must remain reviewable and integrate through the
repository's normal development process. Work packages do not prescribe a
fixed ticket count.

| Package | Deliverable boundary | Required evidence before dependent work |
| --- | --- | --- |
| W01 | Positive vertical publication path: advertisement-aware atomic dual-ref transport, exact result, capability and observer ports, no-op guard defense. | AC-036..043, 051..055 with actual Git processes and protocol review; no all-unsupported implementation. |
| W02 | Forge adapter profile and original/published source refinement. | Positive real canonical-PR observation for default rebase on each existing forge before freezing the mechanism. |
| W03 | Versioned records, migration readers, queue entry identity, journals and recovery ownership. | Revision/idempotency/migration cases and no duplicate authority. |
| W04 | Canonical target scheduling, fairness, shared capacity and coordinator routing. | Serial/parallel and cross-worktree process evidence. |
| W05 | Complete method preparation, provenance, empty-commit and signing policy. | Exact merge/rebase/squash history and no-op collision cases. |
| W06 | Merger input/output protocol and trusted host execution. | Conflicting and ambiguous-intent fixtures; role cannot publish. |
| W07 | Semantic repair, combined validation, candidate review and stale-evidence invalidation. | Textually clean regression and target-drift composed cases. |
| W08 | Persistent merger/drift/time/token limits and operator decisions. | Exhaustion, restart, unknown usage, explicit extension. |
| W09 | Publication exclusion, dispatch fencing, unknown outcomes and observation. | Crash/transport/lease/stop fault matrix with actual processes. |
| W10 | Canonical receipts, fanout, descendant proof, divergence and closure. | Idempotent finalization and no premature ticket completion. |
| W11 | Shared service surface projections/actions and profile setup. | Consistent CLI/TUI/Web/daemon/manual behavior. |
| W12 | Provider/method certification and permission-policy matrix. | All six positive real-provider cases plus contention and denial profiles. |
| W13 | Adapt retained assurance, document support boundaries, final integrated gate and old/new closure. | AC-091..095 with exact integrated commit and verified receipts. |

W01/W02 are a feasibility gate, not permission to weaken the contract later.
If either existing forge cannot report the same PR merged for the required
profile, stop dependent implementation and revise this specification with maintainer approval. Do not continue
building a queue around an unpublishable mechanism or close the set by marking
every production profile unsupported.

The approved plan must map each invariant and acceptance ID to a package and
evidence owner. It must also identify how unmerged assurance work is reused and
how the original delivery is resumed or superseded. One cohesive set may be
longer than the earlier delivery; correctness is not traded for a slice limit.

## 16. Approved Decisions and External Prerequisites

The maintainer approved the following design decisions on 2026-10-11.
Approval authorizes implementation, not a claim of delivered behavior:

1. Enable an opt-in same-repository owned-source rewrite profile with the dual-ref transaction.
2. Keep one registered coordinator as the supported cross-host chr topology.
3. Use three merger attempts, five drift revisions, and existing submission/worker bounds by default.
4. Add optional enforceable invocation token budgeting without requiring a new agent service.
5. Deliver one corrective set and finish the original assurance/closure through the approved mapped transition.
6. Permit explicitly attributed integration-fix commits in reviewed rebase proof.

Before W02 and W12, designate disposable repositories for both existing forges,
approved credential roles, target branches, source-ownership rules, and branch
policies. Test access and evidence retention must be concrete. Missing live
profiles are a certification prerequisite, not a reason to claim mocks prove
remote compatibility. Credentials remain host-owned and outside versioned docs.

No claim is made here that current hosted policy profiles have been certified.
The architectural mechanism has a local design probe; the product implementation
and real forge acceptance remain work for the approved corrective set.

## 17. Completion Evidence

Completion requires the positive real publisher, bounded merger, concurrency and
recovery contract, shared surfaces, six provider/method certification cases,
supported/unsupported policy documentation, exact integrated-revision quality
gate, and verified ticket/set closure. Passing isolated unit tests, a mergeable
PR, all-false capability reports, or a running controller is not completion.

The implementation report must distinguish code integration, local fault tests,
live provider certification, remaining unsupported profiles, original-delivery
closure, and corrective-set closure. Preserve the evidence of rejected attempts
and unresolved outcomes until their retention policy permits removal.

## 18. Design Evidence and Primary References

On 2026-10-11, an ephemeral local Git 2.55.0 probe verified five cases with real
receive-pack processes: successful dual-ref application, target-lease rejection,
source-lease rejection, the omitted no-op-source guard, and rejection when a
pre-receive hook moved a participating source ref after advertisement. Temporary
fixtures were removed after execution. This is design evidence, not production
code validation or live forge certification; the approved work must retain
repeatable acceptance fixtures in the repository.

- [Git push](https://git-scm.com/docs/git-push): atomic updates and explicit expected-value leases.
- [Git protocol capabilities](https://git-scm.com/docs/gitprotocol-capabilities): receive-pack atomic capability negotiation.
- [Git receive-pack](https://git-scm.com/docs/git-receive-pack): ref update processing and hooks.
- [Git rebase](https://git-scm.com/docs/git-rebase): replay and rewritten commit behavior.
- [GitHub merge API](https://docs.github.com/en/rest/pulls/pulls?apiVersion=2022-11-28#merge-a-pull-request): source-SHA guard does not supply the required target-base guard.
