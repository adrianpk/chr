# Execution records, readiness and workspaces

This package owns execution/loop records, runner and artifact capture, event
streams and bounded read-only SQLite observation, branch policy, readiness
observation and explicit safe repair,
and ticket worktree planning, inspection, preparation, metadata and cleanup.
D03 normalization is shared before concern splitting. Check operations retain
read interfaces; repair alone receives its mutation interface and permission gates.

User checkout state, ticket worktree state, preview and authoritative workflow
admission remain distinct. Native workflow orchestration stays with its actual
owner and imports these operations directly. This owner cannot import the root
coordinator or ticket domain rules.

Existing diagnostic/metadata values, nullable/empty results, path conventions,
Git commands, filesystem permissions, error meanings and cleanup guards are
preserved. Named operation/repair/status/configuration results are in-memory
contracts; they do not change persisted fields or wire spellings.

Record/result and event wire fields, SQL column/bind order, nullable timestamps,
read limits, artifact paths and resource cleanup retain their existing contracts.
Ticket-file eligibility and native workflow coordination stay with their own owner.
