# Harness host policy

This package owns host executable resolution, bounded file fingerprints and
retained command admission. It uses binding values from the parent harness
package and preserves workflow conflict identities through workflow/state.

Executable resolution never launches a program. Fingerprints stream at most
512 MiB from regular executable files; symlink resolution, no-follow opens,
nonblocking file admission and exact argv JSON hashes retain their original
semantics. Callers use this owner directly.

The parent harness package contains pure identities, selections and capabilities
and cannot import this host owner. Adapter-neutral runner packages remain
independent of project policy and workflow state.

The host also owns retained generation operations and private artifact IO, model
request dispatch, usage/validation metadata, and the concrete authoring/quiz host
adapters. Original JSON fields, checksum encoding, record bounds, private file
modes and retained lookup order remain intact.

Native configuration admission remains with the configuration loader. It resolves
the actual locked image and supplies a mandatory admission callback; operation
values are validated before that callback, executable checks and artifact writes.
The preparation value contains no context. Callers derive generation contexts and
own cancellation, deadlines and cleanup. Structured model requests retain their
native projection into the host output validator.
