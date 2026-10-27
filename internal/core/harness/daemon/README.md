# Daemon lifecycle

This package owns daemon state, start guards, persistent heartbeat/stop requests
and enabled-project classification. Application startup, dispatch and shutdown
remain assembled by the application.

The SQLite constructor consumes the existing `database.DataDirectory` capability;
it does not load workspace configuration. Project loading reads the existing
project registry and workspace directory marker. No runner or workflow
coordinator is imported.

State fields, SQL statements, stale thresholds and failure messages retain their
existing representation and behavior. Tests use the production migration assets
against owned in-memory databases; the real user-environment path integration
remains beside that environment.
