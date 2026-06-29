# internal/webui/htmx

Purpose: local htmx helper primitives for Web UI templates and handlers.

This package owns small htmx attribute builders and request/response header
helpers. It is copied/adapted into Chryso so the Web UI can use htmx
without depending on an external helper module.

Boundary:

- may build safe `html/template` attributes;
- may inspect htmx request headers;
- may set htmx response headers;
- must not know Chryso project state or business operations.
