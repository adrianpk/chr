# Code Quality Remediation

Status: accepted
Date: 2026-10-26
Delivery set: code-quality-remediation
Planning authorization: maintainer request, 2026-10-26.
Implementation approval: maintainer, 2026-10-26; complete 16-slice proposal, local AST/lint work and one Looper execution approved.
Format: OpenSpec requirements and scenarios, under the shared Specification Contract. No OpenSpec executable is required.

Review: [Quality review](../report/code-quality-review-2026-10-26/report.md)
Detections: [Complete baseline lists](../report/code-quality-review-2026-10-26/detections.md)
Model: [Compatibility model](code-quality-remediation-model.md)
Plan: [Single-cycle plan](../plan/code-quality-remediation.md)
Tracker: [Execution tracker](../tracker/code-quality-remediation.md)

## Purpose

Resolve the accepted production-code findings from the 2026-10-26 review in one delivery set of at most 16 slices. Reduce structural complexity without changing the authoring, ticket, execution, approval, publication or receipt contracts that users rely on.

The review is evidence and a record of maintainer decisions, not an instruction to apply every suggested implementation mechanically. This specification records the remediations and the necessary corrections to those suggestions.

## Baseline and authority

The source baseline is `28c3db337cc0b08d249c65c8f2fd68efd5d0966b`. It matches local dev at planning time. The source report and detections remain unchanged. Runtime behavior is still governed by the existing specifications in their original formats, especially [ticket-native work](stable/ticket-native-workflow.md), [native integration](stable/native-integration-queue.md), [guided use cases](stable/guided-use-case-workflow.md), [Telegram presentation](stable/telegram-message-presentation.md) and [harness contracts](stable/codex-pi-harnesses.md).

## Requirements

### Requirement: Complete single-cycle accountability

The delivery set SHALL account for findings R1 through R10, every row of baseline inventories A, B and C, the 14 large files, six wide records/views, 11 duplicate pairs, seven dead symbols and the minor findings. All actionable accepted work SHALL finish within this plan and its 16 slices. An exclusion SHALL identify the exact contract, false-positive evidence or measured reason; moving code or deferring it to another cycle is not a resolution.

#### Scenario: Baseline detection is moved

- **GIVEN** an unresolved baseline signature violation
- **WHEN** its function moves or is renamed
- **THEN** its inventory identity remains linked to its new owner and owning slice
- **AND** closure requires a fixed signature or a proven external-interface exception, not disappearance at its original path.

#### Scenario: A compatibility path must remain

- **GIVEN** a legacy path still serves a documented supported consumer
- **WHEN** the owning slice reviews it
- **THEN** it records the consumer, governing contract, validation and concrete removal condition
- **AND** it completes every safe simplification in this cycle without deleting that supported behavior.

### Requirement: Cohesive input and result contracts

Production Go functions SHALL end with at most five parameters, counting context and excluding the receiver. Repeated values SHALL be grouped only when they express one actual operation. Functions SHALL return no result, one result, (T, error), or (T, bool). Externally imposed signatures SHALL have a narrow, evidence-backed exception. Project-owned interfaces and helper methods SHALL obey the policy. A five-parameter function is permitted; it is reviewed for cohesion rather than reduced mechanically.

#### Scenario: Lookup has expected absence

- **GIVEN** a lookup currently returns a value, a found flag and an error
- **WHEN** it is converted to a named result
- **THEN** absent, present and failed outcomes remain distinguishable
- **AND** existing normal-absence behavior does not become a fatal operation failure.

#### Scenario: Bubble Tea owns the method signature

- **GIVEN** Model.Update implements the installed tea.Model interface
- **WHEN** the production signature check runs
- **THEN** only that proved method may retain the imposed result shape
- **AND** internal update helpers return commands or named outcomes under the normal policy.

### Requirement: Permanent enforcement without hidden lint debt

The implementation SHALL preserve all existing lint protections. During migration it MAY use a finite, shrinking declaration-specific debt manifest. It SHALL reject new/worsened debt and stale allowances on both feature branches and dev. Final closure SHALL remove temporary debt and enforce five parameters, the result policy, approved semantic duplicate rules and exact external-interface exceptions across the production tree. Test signatures SHALL not be swept into this production-only policy by accident.

#### Scenario: Candidate is already on dev

- **GIVEN** the integrated candidate is the current dev commit
- **WHEN** validation runs
- **THEN** it still examines the complete production tree
- **AND** a self-comparison against dev cannot produce a false green result.

### Requirement: Explicit acyclic capability ownership

Core behavior SHALL remain under internal/core with cohesive capability packages. The central application operation service SHALL remain in internal/app. CLI, TUI, Web and Telegram SHALL map interaction and transport to shared operations. Workflow storage SHALL own durable mutation, transactions, locks and recovery; stage packages SHALL own policy over narrow consumer-owned ports and SHALL not import their coordinator. Root core SHALL not remain an oversized coordinator behind renamed prefixes.

#### Scenario: A package move exposes hidden coupling

- **GIVEN** a move would create a cycle or require exposing many unrelated private helpers
- **WHEN** its task establishes the new boundary
- **THEN** it resolves ownership through a genuine shared value owner or a narrow operation port before moving callers
- **AND** it does not add an untyped store escape hatch, generic helper package or permanent pass-through facade.

### Requirement: Stable durable and hash-sensitive representations

Record grouping, typed states and SQL row mappings SHALL preserve existing durable and wire contracts, including names, optionality, ordering where hashed, null/empty distinctions, state spellings, identity, revisions and receipt meaning. A Go-only structural change SHALL not require a database migration or invalidate retained operations. Existing hashed records SHALL remain verifiable without rewriting their history.

#### Scenario: Embedded fields alter JSON order

- **GIVEN** a record's serialized bytes contribute to an input or evidence hash
- **WHEN** fields are grouped into embedded structs
- **THEN** a before/after fixture proves exact compatible serialization and verification
- **AND** unchanged JSON keys alone are insufficient acceptance evidence.

#### Scenario: Same-type SQL columns can be swapped

- **GIVEN** a repository binds many string or integer fields
- **WHEN** its row mapping is refactored
- **THEN** a real SQLite test writes distinguishable values to every column and reads them back after insert and update
- **AND** the existing transaction and rollback boundary remains authoritative.

### Requirement: Measured and safe presentation state

The TUI SHALL use explicit pointer-model ownership with cohesive sub-states where this removes measured copying. Background commands SHALL not gain unsynchronized access to mutable model state. Web projections SHALL follow existing page sections and preserve the current shared forms, operation semantics and visual language. These refactors SHALL not introduce new screens or redesign workflows.

#### Scenario: A command completes after navigation

- **GIVEN** a background operation was started from one view
- **WHEN** the user changes views before its result arrives
- **THEN** the message updates the correct retained state without racing with model mutation or corrupting the active interaction.

### Requirement: Semantic duplication and removal discipline

Each reported duplicate SHALL be compared with both governing contracts. The same rule SHALL have one owner; coincidentally equal rules MAY remain separate with precise evidence. Test-only production symbols SHALL be removed or moved to test helpers after checking their real callers. Cleanup errors that can alter publication, durability or process release SHALL remain observable.

#### Scenario: Two generators share mechanics but different approvals

- **GIVEN** spec and acceptance generation share context-loading mechanics
- **WHEN** common steps are extracted
- **THEN** each artifact retains its own schema, approval, publication and retry decisions
- **AND** a generic pipeline cannot silently merge those separate contracts.

### Requirement: Code-backed names and compatible vocabulary

Package names SHALL be plain nouns and identifiers SHALL use package qualification without repeated prefixes. Domain terms SHALL be defined from existing contracts before any rename. Ticket-native authority SHALL not be renamed to forge merely because a different native publication path uses a forge provider. Serialized values and existing specification identities SHALL remain stable.

#### Scenario: An ambiguous native identifier is reviewed

- **GIVEN** native denotes canonical ticket-owned execution in a governing spec
- **WHEN** a publication package is named forge
- **THEN** the ticket-native concept retains its distinct meaning
- **AND** a mass native-to-forge replacement is rejected.

### Requirement: Preserved recovery and human authority

Refactoring SHALL preserve finite attempts, deadlines, capacity ownership, explicit cancellation, revision/hash guards, exact approved scope, uncertain-outcome reconciliation and receipt-backed completion. A structural change SHALL not authorize an agent, approve a PR, start work, reset consumed budgets or convert validation into integration proof.

#### Scenario: A publication succeeds but its acknowledgement is lost

- **GIVEN** publication has applied under the existing guarded identity
- **WHEN** the reorganized host resumes
- **THEN** it reconciles the original operation and retained receipt
- **AND** it neither republishes blindly nor invents a completed result.

### Requirement: Proportional integrated completion evidence

Every slice SHALL compile and pass its recorded focused checks and the repository-required make verify. The complete set SHALL run its selected real-boundary gate on the exact integrated candidate. Final acceptance SHALL include before/after counts, no temporary bridges or debt, persistence/hash compatibility, representative adapter journeys, and verified ticket/tracker/report closure. Structural checks and Markdown scenarios SHALL not be represented as runtime acceptance results.

#### Scenario: Planning artifacts exist before implementation

- **GIVEN** this specification, plan, tracker and tickets have been prepared without maintainer approval
- **WHEN** their consistency checks pass
- **THEN** they remain proposed planning artifacts
- **AND** no Looper execution, merge delegation or tool approval is implied.

## Accepted decisions

1. One layered delivery set with the exact 16-slice map in the plan, including package decomposition and final enforcement.
2. One repository-owned Go AST checker, using the standard library and integrated with existing make checks; no new external modules, service or OpenSpec tooling.
3. A declaration-specific migration debt manifest instead of the report's global merge-base lint filter; final thresholds of five parameters and two results, plus the semantic second-result rule.
4. Preserve supported compatibility paths and precise existing domain terms when evidence shows they remain necessary. Record exact retained cases within this cycle, not a generic promise of later work.

On 2026-10-26 the maintainer approved the complete proposal and requested its conventional commit, push and startup through one Looper run. The plan and tracker record implementation authority and conditional Rebase + Fast-forward merge delegation for the 16 recorded slice PRs into dev. This approval includes the standard-library AST checker and proposed lint changes.
