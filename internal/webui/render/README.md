# internal/webui/render

Purpose: template rendering helpers for the local Web UI.

This package owns template parsing, page rendering, fragment rendering, and
function-map composition for server-rendered HTML.

Boundary:

- may parse embedded template files;
- may write full-page or fragment HTML responses;
- may compose presentation helper function maps;
- must not resolve projects or execute Chryso operations.
