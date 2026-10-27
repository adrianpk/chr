# Operational artifact layout and envelopes

This package owns namespace layout, namespace-wide numeric artifact allocation,
canonical and named artifact paths, envelope directories and shared metadata.
Authoring and workflow consumers call it directly. It does not import their
coordinators or own session receipts, approvals or multi-file journals.

Allocation uses the existing exclusive id.lock and preserves cleanup errors.
Metadata retains ops-layout/v1, original field names, timestamps, file indexing
and private-extension exclusion. Known envelopes are preserved when present.
Numeric path-prefix recognition intentionally retains its existing prefix rule;
it is not a stronger complete artifact identity validator.

Public callers use Layout, Artifact, ArtifactOptions and ResolveLayout, plus
cohesive artifact operations. No root core aliases or forwarding methods remain.
