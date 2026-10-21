# Native Integration

Native integration retains one invocation-owned entry per integration unit and
uses the existing shared workflow service. CLI, TUI, Web, daemon and manual
service callers observe the same retained facts. For setup, see
[Configure native integration](../how-to/configure-native-integration.md).

## Executable configuration

The current project configuration accepts these integration-related settings:

| Key | Default | Accepted values and meaning |
| --- | --- | --- |
| `work.integration.publication_profile` | `disabled` | `disabled` or explicit `git_atomic_dual_ref`. Independent of version and review mode. |
| `work.integration.max_prepared_per_target` | `2` | 1–8, further limited by ready queue capacity. |
| `work.max_active_workers` | `1` | 1–8 shared agent-role slots. |
| `work.ready_queue_capacity` | `64` | 1–1024 unfinished lanes. |
| `work.max_integration_attempts` | `3` | 1–10 retained submission attempts per unit. |
| `work.merge_strategy` | `rebase` | `rebase`, `merge` or `squash`; no method fallback. |

`integration` is the one supported nested mapping inside `work`: its scalar
keys use four spaces. Other workflow, work and review scalar keys use two
spaces. Unsupported keys, duplicate sections or fields, tabs and invalid bounds
reject execution. See [Workflow configuration](workflow-configuration.md).

The service's `WorkflowExecutionPolicy.MaxMergerAttempts` accepts 1–10 and
defaults to 3 when omitted. `MaxDriftRevisions` currently accepts 1–10 and
defaults to 5 when omitted. These are service admission fields, not project YAML
keys or CLI flags. Submission and observation IO use three-minute deadlines;
there is no configurable observation-timeout key.

`WorkflowExecutionPolicy.AgentTokenBudget` is an optional positive `int64`
invocation ceiling. Zero omits the ceiling; negative values reject admission.
Current CLI admission and project YAML do not expose this service field.
Finite-token operation requires enforceable, reported whole-call accounting:
input includes cached input, and output includes reasoning. Calls reserve their
maximum before dispatch. Missing, interrupted or incompatible usage keeps that
reservation charged; restart cannot refund it. A token grant extends only an
already finite ceiling.

Admission freezes policy and consumption. Status distinguishes admission limits,
current limits, consumed and remaining attempts, settled and reserved tokens, and
the available finite balance. Candidate replacement, pause/resume and repeated
starts preserve accounting. Operator grants are positive, attributable, bounded
extensions, not resets or publication approval.

## Source rewrite and policy support

The `git_atomic_dual_ref` profile binds original source S, validated target B and
prepared result P. The publisher requires expected old values S and B and moves
both the owned source and the canonical target to P in one atomic transaction.
B must be an ancestor of P; P equal to S requires fresh preparation or an explicit
blocker. Source ownership covers one exact full branch ref in the canonical
repository. Forks and human-owned sources are unsupported.

The baseline requires all of the following:

- A current host ownership grant authorizes the source rewrite and target
  fast-forward, with a known authorized publisher identity.
- Repository policy permits the configured method and host-bound review of P.
- The source and target are unprotected and no enforced branch-rule responses
  are present. Protected branches and nonempty native rules are currently blocked.
- The receive-pack server supports atomic dual-ref updates and advertises the
  bound S and B.
- The existing forge adapter's certified observation contract recognizes the
  original canonical PR as merged by the Git transaction.

The same original PR is observed through the existing Forgejo or GitHub adapter.
Neither source-only legacy merge endpoints nor an adapter-specific retry loop
provide this guarantee. Rebase, merge and squash retain their exact history,
combined-tree validation and candidate review before publication. A capability
change invalidates prior authorization.

Support is evaluated for the actual repository, refs, method and current policy.
An enabled profile is not evidence of live certification. The current executable
blocks protected branches and nonempty enforced-rule responses. Supporting those
policies requires equivalent certified enforcement; existing protections remain
in place while publication is blocked.

The default application service connects the remote adapter to the invocation's
retained source ownership and the host's private `native-publishers.json` file.
Each host profile binds one canonical repository and full target ref, Git
transport, credential reference, API identity fingerprint, observation contract
and certified methods with an acceptance-evidence reference. The file is bounded
to 64 KiB and 64 unique repository/target pairs. Unknown fields, duplicate
profiles, unsupported methods, missing certification or nonprivate files reject
the publisher. See [Configure native integration](../how-to/configure-native-integration.md)
for its format and location.

| Condition | Required action |
| --- | --- |
| `profile_disabled` | Explicitly select the authorized profile. |
| `fork_source`, `source_rewrite_unauthorized`, `ownership_binding` | Resolve ownership of the exact same-repository source ref. |
| `target_push_unauthorized`, `publisher_access_denied` | Resolve authorized publication access to a writable target. |
| `method_disabled`, `method_uncertified` | Permit and certify the configured method. |
| `native_approval_required`, `review_policy_unknown` | Resolve host-bound review under current policy. |
| `native_protections_required`, `native_rules_required` | Certify equivalent enforcement while retaining protections. |
| `observation_uncertified` | Certify same-original-PR observation for this repository. |
| `atomic_transport_unavailable` | Use a server supporting the required atomic transaction. |

## Retained observations and actions

Status shows the integration unit, canonical repository/target and queue position;
original source, validated base and prepared result; actual occupied role slots;
budget accounting; publication state; and the precise blocker or question.
Review waiting or approval can coexist with a retained `reviewing` queue phase.
Approval does not infer publication readiness or delivery.

Private claim credentials, raw lease/process fields, publisher fingerprints and
harness inputs are excluded from the integration projection. Read-only inspection
does not fetch, retry, repair, reserve a role or replay a pending transition.
Malformed or pending state requires its explicit reconciliation owner.

A guarded operator action uses the displayed invocation revision, integration
entry identity/revision, budget revision and candidate hash. A decision answer
also binds the displayed question revision/hash, choice and authority. CLI and Web native
approval require the displayed integration guards and exact review binding.
Manual service approval always requires the displayed invocation revision and
review binding; it can additionally check the supplied integration guard. Stale input rejects
without adopting new authorization from a refreshed status view. See
[Workflow controls](workflow-controls.md#native-budget-grants-and-decisions).

A possibly dispatched submission retains target exclusion until canonical
confirmation or an exact stable-ID unpublished seal. A timeout, cancellation,
open PR or accepted provider response is not a delivery receipt. Verified Git and
canonical PR observations precede receipt finalization. For method proofs,
recovery and receipt contracts, see [Workflow execution](workflow-execution.md).
