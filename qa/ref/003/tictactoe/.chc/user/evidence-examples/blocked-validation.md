# Evidence Example: Blocked Validation

This example represents a ticket that cannot move to review because validation
did not pass.

<!-- chc:evidence-strategy:v1 -->
- [x] Evidence strategy: complete; change_kind=cli_behavior; risk=medium; test_timing=before_or_during
- [x] Gherkin: required; path=.chc/user/acceptance/cli-gameplay.feature
- [x] Tests: mixed; reason=table-driven game tests plus CLI command validation
- [ ] Validation: blocked; command=go test ./...; reason=runtime validation failed
- [ ] Evidence summary: pending; validation blocker must be resolved first
<!-- /chc:evidence-strategy -->

