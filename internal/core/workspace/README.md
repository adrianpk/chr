# Workspace identity

DirectoryName retains the existing .chr directory spelling shared by initialization, retained references and consumers. This value owns no layout selection, configuration, IO or capability policy.

UserDirectoryName, UserRelativePath and UserPath own the existing default
pseudo-user spelling and lexical filepath composition used by authoring,
initialization and native consumers. They retain the original append, Join,
FromSlash and ToSlash sequence, including variadic elements. They validate no
authority and perform no IO; operation owners retain layout selection, path
permission, locking, transactions and recovery. This package remains stdlib-only.
