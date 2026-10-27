# Host runtime planning and launch

This package owns runtime Config, Plan and Invocation values, artifact layout,
container path mapping, launch command assembly and filesystem validation.
ResolvePlan preserves the existing project-directory checks; NewInvocation
validates its explicit execution identity and assembles commands without starting
a process. Validate reports the same required directories and regular files.

Workspace owns ExecutionConfig, execution modes, base/cleanup policies, defaults
and validation. Native configuration parsing, ticket execution, workflow capacity
admission and workspace asset initialization remain with their actual owners.
Consumers import the runtime and workspace values directly.

Canonical ops paths, retained private migration paths, container argv and
mounts, default image/workdir and validation command retain their existing
values. Planning does not launch containers or discover provider resources.
The package cannot import the root configuration or workflow coordinator.

WorkerMounts grants only the explicit checkout and its Git common directory.
WorkerInstructions reads the explicit project contract within the original
regular-file, single-link, UTF-8 and size bounds. Reservation and HoldReservation
hold shared references to existing admitted slot/guard files, release on early
errors and observe caller cancellation. Native workflow admission supplies those
paths; this package cannot create a second capacity pool or choose a slot.
