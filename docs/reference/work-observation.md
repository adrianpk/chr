# Work Observation

The central operation service's `ObserveWork` reads one project and active
namespace under its configured work authority. It returns queue readiness,
separate current contexts, retained results and recent activity. These are
observations of existing records, not new execution or scheduling authority.

## Sources and Identity

In ticket authority mode, `InspectWorkflowWork` reads native tasks, sets and
invocations under the existing repository inspection lock. It uses the shared
selection, prerequisite and slice/task ordering rules without scanning an
unrelated legacy Board. Set groups have deterministic creation/identity order;
loose work remains a separate shared selection. The display does not impose a
new cross-set scheduling policy.

Each current context retains its project, invocation, attempt and ticket
binding. An event from another invocation cannot supply its phase. Native
activity derives from retained invocation revisions and attempt checkpoints;
no tool activity is invented when the source has none. A required pending
transition is unavailable until its owning workflow resolves it. Inspection
does not recover journals or rewrite maps.

In tracker compatibility mode, the shared legacy file reader applies ready
status, declared ordering, adoption exclusion and selection preflight. Existing
execution and loop records are opened through SQLite read-only mode, without
schema initialization. Structured events must match the project and original
execution or loop. Lifecycle records remain authoritative when no phase event
exists, including after completion or interruption.

## Bounds

- Native records use the existing 4096-record and per-file bounds and a
  ten-second inspection deadline.
- Legacy observation retains up to 128 executions and 128 loops, prioritizing
  active records and then recent results.
- At most 32 event streams are read, with at most the final 256 KiB per stream.
  Each retains at most 128 valid entries and bounded malformed-line diagnostics.
- The shared result exposes a 128-item queue page with total and continuation,
  up to 128 current contexts, 16 recent results and 128 deduplicated events.
- The service operation has a fifteen-second deadline. Source failures return
  errors rather than an empty queue or invented lifecycle transition.

## State and Evidence

Native validated-task counts and confirmed-ticket delivery counts remain
separate. A completed host invocation can retain validated work without a
confirmed integration receipt. Legacy loop counters keep their existing
eligible, processed, blocked and failed definitions; execution status does not
create unsupported counters.

Observation time is separate from persisted state and activity times. Quiet
work remains active according to its retained lifecycle state. A read cannot
claim, admit, start, stop, migrate, maintain tracker processing state or alter
execution policy. Existing control operations retain their own authorization
and guards.
