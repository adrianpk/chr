# Model trace policy

This package owns bounded host model trace files, environment controls, payload
retention, redaction and rotation. It resolves the shared workspace userspace and
imports no root configuration or workflow coordinator.

Recording retains its existing best-effort contract. Explicit writes return
filesystem failures, honor caller cancellation and preserve byte limits and
rotation counts. Payloads remain omitted unless enabled. Secret-like fields,
bearer tokens and known environment secrets use the original redaction rules.
