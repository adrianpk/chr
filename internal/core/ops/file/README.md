# Operational file IO

Read and ReadBounded preserve the original default and explicit limits, overflow
checks and read-only close policy. WriteAtomic and WriteAtomicBounded preserve
adjacent temporary images, file sync, rename, directory sync and cleanup errors.
All existing bounds, file modes, error text and temporary names are retained.
They do not replace project.WriteFileAtomically, whose durability contract differs.

Lock retains cancellation-aware advisory exclusive locking and the original
retry interval. Callers still own descriptor lifetime, unlock, lock scope,
recovery, journals and atomic multi-file transitions. No context is stored.
The only internal dependency is workflow/state for the existing default bound.
