# Workspace identity and configuration files

DirectoryName retains the .chr directory spelling. UserDirectoryName,
UserRelativePath and UserPath retain the existing default pseudo-user spelling
and lexical filepath composition, including variadic elements. These values
select no authority or layout and perform no IO.

ConfigPath owns the existing contained .chr/default configuration path guard.
It preserves absolute path resolution, lexical bounds, ancestor symlink rejection,
regular-file checks and original errors. WithConfigLock owns the existing
config.lock file, cancellation-aware acquisition, private creation, callback,
unlock and joined cleanup errors. Configuration interpretation, authorization,
revision checks, upgrades and retained journals stay with their operation owners.

This package uses the existing operational file owner for advisory locking and
has no dependency on the configuration or workflow coordinator.

Userspace resolution validates the active operator and resolves the project
through the existing project directory owner. It retains the default userspace
path and original environment/error contract. EnsureUserspace honors the
caller context before creating that directory. Host trace consumers use this
owner directly without importing the root configuration package.

ExecutionConfig contains the shared execution mode, worktree base and cleanup
policy. DefaultExecutionConfig, FillExecutionDefaults and ValidateExecutionConfig
retain the original zero value, field defaults, fixed worktree root and error
contracts. Runtime planning and native worktree execution consume these values;
this owner does not choose tickets or acquire workflow capacity.

The [configuration child](config/README.md) owns parsed project settings, defaults,
raw-image revisions, explicit-key metadata and recoverable harness-selection
changes. It consumes these path and locking operations; the parent has no import
of the child or the workflow coordinator.
