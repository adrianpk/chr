# internal/webui/ui

Purpose: small server-rendered UI primitives for Web UI templates.

This package owns reusable HTML primitives such as buttons, chips, and
navigation. Primitives must escape user-facing text and stay presentation-only.

Boundary:

- may render trusted component HTML from escaped fields;
- may compose htmx attributes;
- may expose template functions;
- must not include project workflow, persistence, or command execution logic.
