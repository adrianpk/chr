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
