# Evidence Example: Behavior Work With Gherkin

This example represents user-visible CLI behavior while acceptance feature
artifacts are active in the project.

<!-- chc:evidence-strategy:v1 -->
- [x] Evidence strategy: complete; change_kind=cli_behavior; risk=medium; test_timing=before_or_during
- [x] Gherkin: required; path=.chc/user/acceptance/cli-gameplay.feature
- [x] Tests: mixed; reason=table-driven game tests plus CLI command validation
- [x] Validation: passed; command=go test ./...
- [x] Evidence summary: complete; behavior evidence and Gherkin were updated
<!-- /chc:evidence-strategy -->

