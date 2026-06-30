# internal/webui

Purpose: server-rendered local Web UI presentation boundary.

This package owns embedded templates, static assets, and HTTP handlers for the
local Web UI. It receives already-resolved project/config view data
from `internal/app`.

Boundary:

- may render HTML pages and htmx fragments;
- may serve embedded static assets;
- may depend on presentation helper packages under `internal/webui`;
- must not parse project artifacts directly;
- must not mutate project state directly;
- must not be imported by `internal/core`, persistence, or TUI packages.
