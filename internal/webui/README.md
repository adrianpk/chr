# internal/webui

Purpose: server-rendered local Web UI presentation boundary.

This package owns embedded templates, static assets, and HTTP handlers for the
local Web UI. It receives already-resolved project/config view data
from `internal/app`.

Boundary:

This package also owns the optional single-user HTTP login, bounded in-memory
sessions and private account-verifier encoding. It reads no project artifacts
for authentication. The application adapter resolves the private account file
and owns interactive setup and server startup.

- may render HTML pages and htmx fragments;
- may serve embedded static assets;
- may depend on presentation helper packages under `internal/webui`;
- must not parse project artifacts directly;
- must not mutate project state directly;
- must not be imported by `internal/core`, persistence, or TUI packages.


Handler assembly and routing stay in webui.go. Page, artifact and write handlers
have separate files; page_data.go owns read-only page composition, while access
control and request logging have their own owners. Guided leaf templates receive
narrow form, action, status, recovery, progress and result sections from the
existing guided page projection. No projection invokes an application operation.
