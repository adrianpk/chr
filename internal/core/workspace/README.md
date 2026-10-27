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
