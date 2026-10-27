# Acceptance features

This package owns canonical feature and feature-round note paths, validation,
static Gherkin checks, note answer retention, rendering and the existing writes.
The governing contract is docs/reference/acceptance-feature-generation.md.
The spec remains authoritative; features retain their supporting-evidence role.

Feature publication requires the existing static checks before any write. Round
notes preserve existing answers and focused prompts. Canonical paths, generated
metadata, timestamps, traceability, error text and file modes remain unchanged.
The application service owns context, provider selection and generation decisions.
The shared generation package owns model payload contracts. There is no root core,
workflow coordinator, provider IO or model client dependency in this owner.
