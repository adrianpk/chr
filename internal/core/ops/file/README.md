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

ProviderOutput and CommandOutput retain the original file-bounded provider bytes,
16 KiB command prefix, overflow/truncation flags and full-input write counts.
CommandOutput preserves optional per-write file writes and sync error joining.
Process execution and native workflow lifecycle policy stay with callers.
