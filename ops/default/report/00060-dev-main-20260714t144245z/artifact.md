# Dev to Main Gate Report

## Summary

- status: failed
- created_at: 2026-07-14T14:42:45Z

## Git Evidence

- base: origin/main b5e6ae7a678ecfb0c58784bec519b7ad999b8af2
- head: origin/dev 2ab51a301378212849918457a44d22f976021a37
- merge_base: b5e6ae7a678ecfb0c58784bec519b7ad999b8af2
- diff_stat: 416 files changed, 1871 insertions(+), 1869 deletions(-)
- commits: 2
- changed_files: 416

### Commits

- `2ab51a301378212849918457a44d22f976021a37` refactor(paths): retire legacy chr user workspace
- `e1abf82be376d0496c06a342004bc2319f34af0e` refactor(paths): retire legacy chr user workspace

### Changed Files

- M .agents/ARCHITECTURE.md
- R100 .chr/user/agents/ARCHITECTURE.md -> .chr/default/agents/ARCHITECTURE.md
- R100 .chr/user/agents/CONVENTIONS.md -> .chr/default/agents/CONVENTIONS.md
- R100 .chr/user/agents/DOCUMENTING.md -> .chr/default/agents/DOCUMENTING.md
- R100 .chr/user/agents/INVARIANTS.md -> .chr/default/agents/INVARIANTS.md
- R100 .chr/user/agents/LIFECYCLE.md -> .chr/default/agents/LIFECYCLE.md
- R100 .chr/user/agents/LINTING.md -> .chr/default/agents/LINTING.md
- R083 .chr/user/agents/PLAYBOOKS.md -> .chr/default/agents/PLAYBOOKS.md
- R100 .chr/user/agents/ROUTING.md -> .chr/default/agents/ROUTING.md
- R100 .chr/user/agents/SLICES.md -> .chr/default/agents/SLICES.md
- R100 .chr/user/agents/VERSIONING.md -> .chr/default/agents/VERSIONING.md
- R100 .chr/user/agents/playbooks/dump.md -> .chr/default/agents/playbooks/dump.md
- R100 .chr/user/agents/playbooks/planned-delivery-workflow.md -> .chr/default/agents/playbooks/planned-delivery-workflow.md
- R100 .chr/user/agents/playbooks/quiz.md -> .chr/default/agents/playbooks/quiz.md
- R100 .chr/user/agents/playbooks/release-alignment.md -> .chr/default/agents/playbooks/release-alignment.md
- R100 .chr/user/agents/playbooks/slice-delivery.md -> .chr/default/agents/playbooks/slice-delivery.md
- R100 .chr/user/agents/playbooks/specification-workflow.md -> .chr/default/agents/playbooks/specification-workflow.md
- R100 .chr/user/agents/playbooks/ticket-workflow.md -> .chr/default/agents/playbooks/ticket-workflow.md
- R100 .chr/user/agents/playbooks/version-control-workflow.md -> .chr/default/agents/playbooks/version-control-workflow.md
- R100 .chr/user/cases/context-capture.md -> .chr/default/cases/context-capture.md
- R095 .chr/user/cases/dev-main-gate.md -> .chr/default/cases/dev-main-gate.md
- R100 .chr/user/cases/qa-scenarios.md -> .chr/default/cases/qa-scenarios.md
- R100 .chr/user/cases/spec-to-tickets.md -> .chr/default/cases/spec-to-tickets.md
- R100 .chr/user/cases/ticket-execution-loop.md -> .chr/default/cases/ticket-execution-loop.md
- R100 .chr/user/cases/tui-operations.md -> .chr/default/cases/tui-operations.md
- R100 .chr/user/config.yml -> .chr/default/config.yml
- R100 .chr/user/config/gates.yml -> .chr/default/config/gates.yml
- R085 .chr/user/docs/adoption-baseline.md -> .chr/default/docs/adoption-baseline.md
- R069 .chr/user/docs/legacy-material-map.md -> .chr/default/docs/legacy-material-map.md
- R098 .chr/user/dumps/curated/2026-06-02-merger-agent-role.md -> .chr/default/dumps/curated/2026-06-02-merger-agent-role.md
- R097 .chr/user/dumps/curated/dev-main-gate-context.md -> .chr/default/dumps/curated/dev-main-gate-context.md
- R100 .chr/user/dumps/curated/document-generation-context.md -> .chr/default/dumps/curated/document-generation-context.md
- R087 .chr/user/dumps/curated/project-context.md -> .chr/default/dumps/curated/project-context.md
- R100 .chr/user/dumps/curated/runtime-execution-context.md -> .chr/default/dumps/curated/runtime-execution-context.md
- R100 .chr/user/dumps/raw/2026-06-02-merger-agent-role-raw.md -> .chr/default/dumps/raw/2026-06-02-merger-agent-role-raw.md
- R100 .chr/user/manifest.md -> .chr/default/manifest.md
- R100 .chr/user/plans/dev-main-gate-plan.md -> .chr/default/plans/dev-main-gate-plan.md
- R090 .chr/user/plans/merger-agent-workflow-plan.md -> .chr/default/plans/merger-agent-workflow-plan.md
- R092 .chr/user/plans/ticket-review-integration-plan.md -> .chr/default/plans/ticket-review-integration-plan.md
- R085 .chr/user/project.md -> .chr/default/project.md
- R095 .chr/user/quizzes/curated/ticket-review-integration-quiz.md -> .chr/default/quizzes/curated/ticket-review-integration-quiz.md
- R098 .chr/user/quizzes/raw/ticket-review-integration-quiz-raw.md -> .chr/default/quizzes/raw/ticket-review-integration-quiz-raw.md
- R100 .chr/user/runtime/agent/Dockerfile -> .chr/default/runtime/agent/Dockerfile
- R092 .chr/user/specs/dev-main-gate-spec.md -> .chr/default/specs/dev-main-gate-spec.md
- R100 .chr/user/specs/merger-agent-workflow.md -> .chr/default/specs/merger-agent-workflow.md
- R100 .chr/user/specs/ticket-review-integration-draft.md -> .chr/default/specs/ticket-review-integration-draft.md
- R100 .chr/user/trackers/merger-agent-workflow-tracker.md -> .chr/default/trackers/merger-agent-workflow-tracker.md
- R100 .chr/user/trackers/ticket-review-integration-tracker.md -> .chr/default/trackers/ticket-review-integration-tracker.md
- M .gitignore
- M .prj/fixtures/mvp-agent-proving-ground/todolist/.chr/config.yml
- R100 .prj/fixtures/mvp-agent-proving-ground/todolist/.chr/runtime/agent/Dockerfile -> .prj/fixtures/mvp-agent-proving-ground/todolist/.chr/default/runtime/agent/Dockerfile
- M .prj/fixtures/mvp-agent-proving-ground/todolist/README.md
- M .prj/plans/slicing-board-ticket-operations-plan.md
- M .prj/plans/slicing-codex-artifact-generation-plan.md
- M .prj/plans/slicing-llm-file-tracing-plan.md
- M .prj/plans/slicing-merger-agent-workflow-plan.md
- M .prj/plans/slicing-qa-agent-work-scenario-plan.md
- M .prj/plans/slicing-ticket-review-integration-plan.md
- M .prj/quizzes/curated/dashboard-board-ticket-operations-quiz.md
- M .prj/quizzes/curated/ticket-review-integration-quiz.md
- M .prj/quizzes/raw/dashboard-board-ticket-operations-quiz-raw.md
- M .prj/quizzes/raw/ticket-review-integration-quiz-raw.md
- M .prj/reports/llm-trace-payloads-2026-05-27.md
- M .prj/reports/slices/agent-test-evidence-strategy-pass/slice-1-evidence-strategy-config.md
- M .prj/reports/slices/agent-test-evidence-strategy-pass/slice-5-gherkin-test-policy.md
- M .prj/reports/slices/agent-test-evidence-strategy-pass/slice-6-evidence-visibility-docs.md
- M .prj/reports/slices/agent-work-observability/slice-2-work-lifecycle-events.md
- M .prj/reports/slices/board-ticket-operations/slice-12-board-docs-validation.md
- M .prj/reports/slices/board-ticket-operations/slice-2-userspace-resolver.md
- M .prj/reports/slices/codex-artifact-generation/slice-1-plan-tracking-codex-backend.md
- M .prj/reports/slices/codex-artifact-generation/slice-4-artifact-run-bundle-audit.md
- M .prj/reports/slices/dev-main-agent-evidence-gate/slice-1-gate-command-git-evidence.md
- M .prj/reports/slices/dev-main-agent-evidence-gate/slice-2-validation-report-output.md
- M .prj/reports/slices/dev-main-agent-evidence-gate/slice-4-gate-policy-configuration.md
- M .prj/reports/slices/dev-main-agent-evidence-gate/slice-5-forgejo-woodpecker-integration.md
- M .prj/reports/slices/dev-main-agent-evidence-gate/slice-5-forgejo-woodpecker-investigation.md
- M .prj/reports/slices/gherkin-acceptance-generation/slice-3-feature-artifact-checks.md
- M .prj/reports/slices/gherkin-acceptance-generation/slice-4-spec-feature-generation.md
- M .prj/reports/slices/gherkin-acceptance-generation/slice-5-acceptance-workflow-docs.md
- M .prj/reports/slices/isolated-ticket-worktrees/slice-1-workspace-execution-config.md
- M .prj/reports/slices/llm-file-tracing/slice-1-llm-trace-sink.md
- M .prj/reports/slices/llm-file-tracing/slice-2-quiz-model-trace-integration.md
- M .prj/reports/slices/mvp-agent-proving-ground/slice-1-proving-ground-fixture.md
- M .prj/reports/slices/qa-agent-work-scenario/slice-3-seeded-chr-context-adr-docs.md
- M .prj/reports/slices/separate-acceptance-feature-workflow/slice-1-spec-feature-decoupling.md
- M .prj/reports/slices/separate-acceptance-feature-workflow/slice-2-acceptance-feature-service.md
- M .prj/reports/slices/separate-acceptance-feature-workflow/slice-4-feature-round-defaults.md
- M .prj/reports/slices/separate-acceptance-feature-workflow/slice-6-acceptance-workflow-docs-and-fixture.md
- M .prj/reports/slices/serialized-integration-loop/slice-3-high-level-agentic-review.md
- M .prj/reports/slices/serialized-integration-loop/slice-6-observability-docs-and-qa.md
- M .prj/reports/slices/ticket-review-integration/slice-3-review-mode-configuration.md
- M .prj/reports/slices/ticket-review-integration/slice-4-forge-provider-diagnostics.md
- M .prj/reports/slices/ticket-review-integration/slice-7-serialized-work-preflight.md
- M .prj/reports/slices/tictactoe-qa-reference/slice-1-tictactoe-qa-reference.md
- M .prj/scripts/probe-codex-artifact-generation.sh
- M .prj/specs/drafts/ticket-review-integration-draft.md
- M .prj/specs/specs/board-ticket-operations-model.md
- M .prj/specs/specs/board-ticket-operations.md
- M .prj/specs/specs/merger-agent-workflow.md
- M .prj/specs/specs/qa-agent-work-scenario.md
- M .prj/trackers/slicing-board-ticket-operations-tracker.md
- M .prj/trackers/slicing-merger-agent-workflow-tracker.md
- M .prj/trackers/slicing-qa-agent-work-scenario-tracker.md
- M .prj/trackers/slicing-ticket-review-integration-tracker.md
- M CHANGELOG.md
- M cmd/chr/assets/workspace/profiles/standard/agents/PLAYBOOKS.md
- M cmd/chr/assets/workspace/profiles/standard/root/AGENTS.md
- M docs/explanation/architecture.md
- M docs/explanation/dev-main-gate-direction.md
- M docs/explanation/dev-main-gate-model.md
- M docs/explanation/dev-main-tooling-gate-spec-draft.md
- M docs/explanation/old-gallery.md
- M docs/explanation/tui-interactive-design-notes.md
- M docs/how-to/adopt-dev-main-gate-in-existing-project.md
- M docs/how-to/full-system-lab-runbook.md
- M docs/how-to/qa-tictactoe-agent-work-scenario.md
- M docs/how-to/use-dev-main-gate-with-woodpecker.md
- M docs/how-to/use-forgejo-ticket-review.md
- M docs/how-to/user-guide/index.md
- M docs/reference/acceptance-feature-generation.md
- M docs/reference/authoring-sessions.md
- M docs/reference/commands.md
- M docs/reference/dev-main-gate.md
- M docs/reference/evidence-strategy-config.md
- M docs/reference/index-rebuild-cli.md
- M docs/reference/index-status-cli.md
- M docs/reference/notification-config.md
- M docs/reference/project-workspace.md
- M docs/reference/review-workflow-config.md
- M docs/reference/runtime-config.md
- M docs/reference/ticket-execution.md
- M docs/reference/ticket-file-discovery.md
- M docs/reference/ticket-index-rebuild.md
- M docs/reference/ticket-index-refresh.md
- M docs/reference/ticket-index-status.md
- M docs/reference/ticket-reconciliation.md
- M docs/reference/ticket-review-workflow.md
- M docs/reference/tui.md
- M docs/reference/user-environment.md
- M docs/reference/work-readiness.md
- M docs/tutorials/getting-started.md
- M internal/app/cli_test.go
- M internal/app/planning_baseline_test.go
- M internal/app/service_acceptance_feature_test.go
- M internal/app/service_dump_curation_test.go
- M internal/app/service_plan_test.go
- M internal/app/service_quiz_test.go
- M internal/app/service_spec_test.go
- M internal/app/service_test.go
- M internal/app/service_work_test.go
- M internal/app/tui_adapter_parity_test.go
- M internal/core/acceptance_feature_test.go
- M internal/core/agent_event_readback_test.go
- M internal/core/agent_event_test.go
- M internal/core/agent_run_test.go
- M internal/core/agentic_review_test.go
- M internal/core/assisted_artifact_synthesis_test.go
- M internal/core/assisted_context_test.go
- M internal/core/authoring_artifact_layout_test.go
- M internal/core/authoring_contextual_quiz_test.go
- M internal/core/authoring_session_test.go
- M internal/core/changed_file_discovery_test.go
- M internal/core/codex_worker_context_test.go
- M internal/core/codex_worker_test.go
- M internal/core/config.go
- M internal/core/config_test.go
- M internal/core/context_agent_protocols_test.go
- M internal/core/context_composer_test.go
- M internal/core/context_materializer_test.go
- M internal/core/context_previous_attempt_test.go
- M internal/core/context_project_documents_test.go
- M internal/core/context_readiness_handoff_test.go
- M internal/core/context_tree_search_test.go
- M internal/core/diagnostic_bundle_test.go
- M internal/core/evidence_policy.go
- M internal/core/evidence_policy_test.go
- M internal/core/evidence_readiness_test.go
- M internal/core/evidence_strategy_checklist_test.go
- M internal/core/evidence_strategy_summary_test.go
- M internal/core/execution_artifact_drift_test.go
- M internal/core/execution_report_test.go
- M internal/core/execution_repository_test.go
- M internal/core/execution_runner_test.go
- M internal/core/execution_test.go
- M internal/core/gate_report_test.go
- M internal/core/harness_invocation_test.go
- M internal/core/llm_trace_test.go
- M internal/core/plan_generation_test.go
- M internal/core/project_userspace_test.go
- M internal/core/proving_test.go
- M internal/core/quiz_artifact_repository_test.go
- M internal/core/quiz_model_test.go
- M internal/core/quiz_runtime_test.go
- M internal/core/result_application_test.go
- M internal/core/runtime_init_test.go
- M internal/core/runtime_plan.go
- M internal/core/runtime_plan_test.go
- M internal/core/runtime_validate_test.go
- M internal/core/spec_generation.go
- M internal/core/spec_generation_test.go
- M internal/core/ticket_discovery_test.go
- M internal/core/ticket_index_rebuild_test.go
- M internal/core/ticket_index_status_test.go
- M internal/core/ticket_index_test.go
- M internal/core/ticket_merge_preflight_test.go
- M internal/core/ticket_parser_test.go
- M internal/core/ticket_reconciliation_test.go
- M internal/core/ticket_repository_test.go
- M internal/core/ticket_test.go
- M internal/core/ticket_validation_test.go
- M internal/core/work_readiness_service.go
- M internal/core/work_readiness_service_test.go
- M internal/core/workspace_context_test.go
- M internal/core/workspace_graph_test.go
- M internal/core/workspace_index_cache_test.go
- M internal/core/workspace_index_test.go
- M internal/core/workspace_init.go
- M internal/core/workspace_init_test.go
- M internal/core/worktree_service_test.go
- M internal/tui/documents.go
- M internal/tui/model_test.go
- M internal/tui/source_actions.go
- M internal/tui/work_readiness_test.go
- M internal/webui/webui_test.go
- M ops/default/dump/00001-curated/artifact.md
- M ops/default/dump/00001-curated/files/legacy/2026-06-02-merger-agent-role.md
- M ops/default/dump/00001-curated/files/legacy/dev-main-gate-context.md
- M ops/default/dump/00001-curated/files/legacy/project-context.md
- M ops/default/dump/00001-curated/metadata.json
- M ops/default/dump/00002-raw/artifact.md
- M ops/default/dump/00002-raw/metadata.json
- M ops/default/plan/00008-dev-main-gate-plan-md/artifact.md
- M ops/default/plan/00008-dev-main-gate-plan-md/metadata.json
- M ops/default/plan/00009-merger-agent-workflow-plan-md/artifact.md
- M ops/default/plan/00009-merger-agent-workflow-plan-md/files/legacy/merger-agent-workflow-plan.md
- M ops/default/plan/00009-merger-agent-workflow-plan-md/metadata.json
- M ops/default/plan/00010-ticket-review-integration-plan-md/artifact.md
- M ops/default/plan/00010-ticket-review-integration-plan-md/files/legacy/ticket-review-integration-plan.md
- M ops/default/plan/00010-ticket-review-integration-plan-md/metadata.json
- M ops/default/plan/00026-2026-06-01-dev-main-agent-evidence-gate-plan-md/files/legacy/2026-06-01-dev-main-agent-evidence-gate-plan.md
- M ops/default/plan/00028-2026-06-06-separate-acceptance-feature-workflow-plan-md/files/legacy/2026-06-06-separate-acceptance-feature-workflow-plan.md
- M ops/default/plan/00031-2026-06-18-tictactoe-qa-reference-plan-md/files/legacy/2026-06-18-tictactoe-qa-reference-plan.md
- M ops/default/plan/00032-2026-06-19-agent-test-evidence-strategy-pass-plan-md/files/legacy/2026-06-19-agent-test-evidence-strategy-pass-plan.md
- M ops/default/quiz/00003-curated/artifact.md
- M ops/default/quiz/00003-curated/files/legacy/ticket-review-integration-quiz.md
- M ops/default/quiz/00003-curated/metadata.json
- M ops/default/quiz/00004-raw/artifact.md
- M ops/default/quiz/00004-raw/files/legacy/ticket-review-integration-quiz-raw.md
- M ops/default/quiz/00004-raw/metadata.json
- M ops/default/report/00036-ops-layout-migration/files/manifest.json
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/cases/cli-gameplay.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/cases/cli-gameplay.md
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/config.yml -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/config.yml
- R095 qa/ref/003/tictactoe/.chr/user/dumps/cli-gameplay-context-raw.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/dumps/cli-gameplay-context-raw.md
- R095 qa/ref/003/tictactoe/.chr/user/dumps/cli-gameplay-context.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/dumps/cli-gameplay-context.md
- R094 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/dumps/project-context-raw.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/dumps/project-context-raw.md
- R094 qa/ref/003/tictactoe/.chr/user/dumps/project-context.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/dumps/project-context.md
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/manifest.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/manifest.md
- R098 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/plans/archive/cli-gameplay-plan-20260618-084729438415104.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/plans/archive/cli-gameplay-plan-20260618-084729438415104.md
- R098 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/plans/cli-gameplay-plan.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/plans/cli-gameplay-plan.md
- R082 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/plans/cli-gameplay-tracking.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/plans/cli-gameplay-tracking.md
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/project.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/project.md
- R096 qa/ref/003/tictactoe/.chr/user/quizzes/cli-gameplay-quiz-raw.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/quizzes/cli-gameplay-quiz-raw.md
- R091 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/quizzes/cli-gameplay-quiz.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/quizzes/cli-gameplay-quiz.md
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/agent/Dockerfile -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/agent/Dockerfile
- R074 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/plan-generation.log -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/plan-generation.log
- R055 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-1/exec-648ea4e465b090f7/context.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-1/exec-648ea4e465b090f7/context.md
- R069 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-1/exec-648ea4e465b090f7/execution-report.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-1/exec-648ea4e465b090f7/execution-report.md
- A ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-1/exec-648ea4e465b090f7/result-application.md
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-1/exec-648ea4e465b090f7/result.json -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-1/exec-648ea4e465b090f7/result.json
- R098 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-1/exec-648ea4e465b090f7/runtime.log -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-1/exec-648ea4e465b090f7/runtime.log
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-1/exec-648ea4e465b090f7/validation.log -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-1/exec-648ea4e465b090f7/validation.log
- R055 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-2/exec-409755cd38adfe67/context.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-2/exec-409755cd38adfe67/context.md
- R070 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-2/exec-409755cd38adfe67/execution-report.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-2/exec-409755cd38adfe67/execution-report.md
- A ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-2/exec-409755cd38adfe67/result-application.md
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-2/exec-409755cd38adfe67/result.json -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-2/exec-409755cd38adfe67/result.json
- R087 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-2/exec-409755cd38adfe67/runtime.log -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-2/exec-409755cd38adfe67/runtime.log
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-2/exec-409755cd38adfe67/validation.log -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-2/exec-409755cd38adfe67/validation.log
- R055 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/context.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/context.md
- R070 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/execution-report.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/execution-report.md
- A ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/result-application.md
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/result.json -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/result.json
- R083 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/runtime.log -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/runtime.log
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/validation.log -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/validation.log
- R054 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-4/exec-6326a1544f559788/context.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-4/exec-6326a1544f559788/context.md
- R071 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-4/exec-6326a1544f559788/execution-report.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-4/exec-6326a1544f559788/execution-report.md
- A ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-4/exec-6326a1544f559788/result-application.md
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-4/exec-6326a1544f559788/result.json -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-4/exec-6326a1544f559788/result.json
- R096 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-4/exec-6326a1544f559788/runtime.log -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-4/exec-6326a1544f559788/runtime.log
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-4/exec-6326a1544f559788/validation.log -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-4/exec-6326a1544f559788/validation.log
- R055 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-5/exec-7fe25b30b4955e1e/context.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-5/exec-7fe25b30b4955e1e/context.md
- R070 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-5/exec-7fe25b30b4955e1e/execution-report.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-5/exec-7fe25b30b4955e1e/execution-report.md
- A ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-5/exec-7fe25b30b4955e1e/result-application.md
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-5/exec-7fe25b30b4955e1e/result.json -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-5/exec-7fe25b30b4955e1e/result.json
- R095 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-5/exec-7fe25b30b4955e1e/runtime.log -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-5/exec-7fe25b30b4955e1e/runtime.log
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-5/exec-7fe25b30b4955e1e/validation.log -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-5/exec-7fe25b30b4955e1e/validation.log
- R098 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/specs/project-spec.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/specs/project-spec.md
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/tickets/.gitkeep -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/tickets/.gitkeep
- R085 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/tickets/cli-gameplay/CHR-1-Game-outcome-rules.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/tickets/cli-gameplay/CHR-1-Game-outcome-rules.md
- R087 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/tickets/cli-gameplay/CHR-2-Move-state-transitions.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/tickets/cli-gameplay/CHR-2-Move-state-transitions.md
- R087 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/tickets/cli-gameplay/CHR-3-CLI-board-rendering.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/tickets/cli-gameplay/CHR-3-CLI-board-rendering.md
- R088 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/tickets/cli-gameplay/CHR-4-Interactive-gameplay-loop.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/tickets/cli-gameplay/CHR-4-Interactive-gameplay-loop.md
- R088 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/tickets/cli-gameplay/CHR-5-Test-coverage-hardening.md -> ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/tickets/cli-gameplay/CHR-5-Test-coverage-hardening.md
- D ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-1/exec-648ea4e465b090f7/result-application.md
- D ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-2/exec-409755cd38adfe67/result-application.md
- D ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/result-application.md
- D ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-4/exec-6326a1544f559788/result-application.md
- D ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-5/exec-7fe25b30b4955e1e/result-application.md
- M ops/default/report/00037-tictactoe-agent-work-success/metadata.json
- M ops/default/spec/00005-dev-main-gate-spec-md/artifact.md
- M ops/default/spec/00005-dev-main-gate-spec-md/files/legacy/dev-main-gate-spec.md
- M ops/default/spec/00005-dev-main-gate-spec-md/metadata.json
- M ops/default/spec/00006-merger-agent-workflow-md/artifact.md
- M ops/default/spec/00006-merger-agent-workflow-md/metadata.json
- M ops/default/spec/00007-ticket-review-integration-draft-md/artifact.md
- M ops/default/spec/00007-ticket-review-integration-draft-md/metadata.json
- M ops/default/spec/00011-2026-06-01-dev-main-agent-evidence-gate-md/files/legacy/2026-06-01-dev-main-agent-evidence-gate.md
- M ops/default/spec/00012-2026-06-06-gherkin-acceptance-generation-md/files/legacy/2026-06-06-gherkin-acceptance-generation.md
- M ops/default/spec/00013-2026-06-06-separate-acceptance-feature-workflow-md/files/legacy/2026-06-06-separate-acceptance-feature-workflow.md
- M ops/default/spec/00014-2026-06-18-agent-work-observability-model-md/files/legacy/2026-06-18-agent-work-observability-model.md
- M ops/default/spec/00018-2026-06-18-configurable-work-readiness-md/files/legacy/2026-06-18-configurable-work-readiness.md
- M ops/default/spec/00019-2026-06-19-agent-test-evidence-strategy-pass-md/files/legacy/2026-06-19-agent-test-evidence-strategy-pass.md
- M ops/default/spec/00020-2026-06-19-isolated-ticket-worktrees-md/files/legacy/2026-06-19-isolated-ticket-worktrees.md
- M ops/default/spec/00021-2026-06-20-serialized-integration-loop-model-md/files/legacy/2026-06-20-serialized-integration-loop-model.md
- M ops/default/spec/00022-2026-06-20-serialized-integration-loop-md/files/legacy/2026-06-20-serialized-integration-loop.md
- M ops/default/spec/00023-2026-06-29-execution-unit-runtime-refinements-md/files/legacy/2026-06-29-execution-unit-runtime-refinements.md
- M ops/default/spec/00024-2026-06-29-optional-local-web-ui-md/files/legacy/2026-06-29-optional-local-web-ui.md
- M ops/default/spec/00025-drafts/files/legacy/2026-06-01-dev-main-agent-evidence-gate.md
- M ops/default/spec/00025-drafts/files/legacy/2026-06-06-gherkin-acceptance-generation.md
- M ops/plans/2026-06-01-dev-main-agent-evidence-gate-plan.md
- M ops/plans/2026-06-06-separate-acceptance-feature-workflow-plan.md
- M ops/plans/2026-06-18-tictactoe-qa-reference-plan.md
- M ops/plans/2026-06-19-agent-test-evidence-strategy-pass-plan.md
- M ops/specs/2026-06-01-dev-main-agent-evidence-gate.md
- M ops/specs/2026-06-06-gherkin-acceptance-generation.md
- M ops/specs/2026-06-06-separate-acceptance-feature-workflow.md
- M ops/specs/2026-06-18-agent-work-observability-model.md
- M ops/specs/2026-06-18-configurable-work-readiness.md
- M ops/specs/2026-06-19-agent-test-evidence-strategy-pass.md
- M ops/specs/2026-06-19-isolated-ticket-worktrees.md
- M ops/specs/2026-06-20-serialized-integration-loop-model.md
- M ops/specs/2026-06-20-serialized-integration-loop.md
- M ops/specs/2026-06-29-execution-unit-runtime-refinements.md
- M ops/specs/2026-06-29-optional-local-web-ui.md
- M ops/trackers/2026-06-01-dev-main-agent-evidence-gate-tracker.md
- M ops/trackers/2026-06-06-gherkin-acceptance-generation-tracker.md
- M ops/trackers/2026-06-06-separate-acceptance-feature-workflow-tracker.md
- M ops/trackers/2026-06-19-agent-test-evidence-strategy-pass-tracker.md
- M ops/trackers/2026-06-20-serialized-integration-loop-tracker.md
- R100 qa/ref/001/blog/.chr/user/cases/comments.md -> qa/ref/001/blog/.chr/default/cases/comments.md
- R100 qa/ref/001/blog/.chr/user/config.yml -> qa/ref/001/blog/.chr/default/config.yml
- R100 qa/ref/001/blog/.chr/user/dumps/comments-context.md -> qa/ref/001/blog/.chr/default/dumps/comments-context.md
- R100 qa/ref/001/blog/.chr/user/manifest.md -> qa/ref/001/blog/.chr/default/manifest.md
- R100 qa/ref/001/blog/.chr/user/project.md -> qa/ref/001/blog/.chr/default/project.md
- R094 qa/ref/001/blog/.chr/user/quizzes/comments-quiz-raw.md -> qa/ref/001/blog/.chr/default/quizzes/comments-quiz-raw.md
- R091 qa/ref/001/blog/.chr/user/quizzes/comments-quiz.md -> qa/ref/001/blog/.chr/default/quizzes/comments-quiz.md
- R100 qa/ref/001/blog/.chr/user/tickets/CHR-1-fix-unknown-post-status.md -> qa/ref/001/blog/.chr/default/tickets/CHR-1-fix-unknown-post-status.md
- R100 qa/ref/002/dialer/.chr/user/cases/call-outcome-capture.md -> qa/ref/002/dialer/.chr/default/cases/call-outcome-capture.md
- R100 qa/ref/002/dialer/.chr/user/cases/conversion.md -> qa/ref/002/dialer/.chr/default/cases/conversion.md
- R100 qa/ref/002/dialer/.chr/user/cases/followups.md -> qa/ref/002/dialer/.chr/default/cases/followups.md
- R100 qa/ref/002/dialer/.chr/user/cases/lead-queue.md -> qa/ref/002/dialer/.chr/default/cases/lead-queue.md
- R100 qa/ref/002/dialer/.chr/user/config.yml -> qa/ref/002/dialer/.chr/default/config.yml
- R100 qa/ref/002/dialer/.chr/user/dumps/call-outcome-capture-context.md -> qa/ref/002/dialer/.chr/default/dumps/call-outcome-capture-context.md
- R100 qa/ref/002/dialer/.chr/user/dumps/conversion-context.md -> qa/ref/002/dialer/.chr/default/dumps/conversion-context.md
- R100 qa/ref/002/dialer/.chr/user/dumps/followups-context.md -> qa/ref/002/dialer/.chr/default/dumps/followups-context.md
- R100 qa/ref/002/dialer/.chr/user/dumps/lead-queue-context.md -> qa/ref/002/dialer/.chr/default/dumps/lead-queue-context.md
- R100 qa/ref/002/dialer/.chr/user/dumps/project-context.md -> qa/ref/002/dialer/.chr/default/dumps/project-context.md
- R100 qa/ref/002/dialer/.chr/user/manifest.md -> qa/ref/002/dialer/.chr/default/manifest.md
- R100 qa/ref/002/dialer/.chr/user/plans/conversion-plan.md -> qa/ref/002/dialer/.chr/default/plans/conversion-plan.md
- R090 qa/ref/002/dialer/.chr/user/plans/conversion-tracking.md -> qa/ref/002/dialer/.chr/default/plans/conversion-tracking.md
- R100 qa/ref/002/dialer/.chr/user/plans/followups-plan.md -> qa/ref/002/dialer/.chr/default/plans/followups-plan.md
- R091 qa/ref/002/dialer/.chr/user/plans/followups-tracking.md -> qa/ref/002/dialer/.chr/default/plans/followups-tracking.md
- R100 qa/ref/002/dialer/.chr/user/plans/lead-queue-plan.md -> qa/ref/002/dialer/.chr/default/plans/lead-queue-plan.md
- R090 qa/ref/002/dialer/.chr/user/plans/lead-queue-tracking.md -> qa/ref/002/dialer/.chr/default/plans/lead-queue-tracking.md
- R100 qa/ref/002/dialer/.chr/user/project.md -> qa/ref/002/dialer/.chr/default/project.md
- R084 qa/ref/002/dialer/.chr/user/quizzes/conversion-quiz-raw.md -> qa/ref/002/dialer/.chr/default/quizzes/conversion-quiz-raw.md
- R090 qa/ref/002/dialer/.chr/user/quizzes/conversion-quiz.md -> qa/ref/002/dialer/.chr/default/quizzes/conversion-quiz.md
- R082 qa/ref/002/dialer/.chr/user/quizzes/followups-quiz-raw.md -> qa/ref/002/dialer/.chr/default/quizzes/followups-quiz-raw.md
- R088 qa/ref/002/dialer/.chr/user/quizzes/followups-quiz.md -> qa/ref/002/dialer/.chr/default/quizzes/followups-quiz.md
- R085 qa/ref/002/dialer/.chr/user/quizzes/lead-queue-quiz-raw.md -> qa/ref/002/dialer/.chr/default/quizzes/lead-queue-quiz-raw.md
- R090 qa/ref/002/dialer/.chr/user/quizzes/lead-queue-quiz.md -> qa/ref/002/dialer/.chr/default/quizzes/lead-queue-quiz.md
- R100 qa/ref/002/dialer/.chr/user/specs/conversion-spec.md -> qa/ref/002/dialer/.chr/default/specs/conversion-spec.md
- R100 qa/ref/002/dialer/.chr/user/specs/followups-spec.md -> qa/ref/002/dialer/.chr/default/specs/followups-spec.md
- R100 qa/ref/002/dialer/.chr/user/specs/lead-queue-spec.md -> qa/ref/002/dialer/.chr/default/specs/lead-queue-spec.md
- R094 qa/ref/002/dialer/.chr/user/tickets/conversion/CHR-104-add-prospect-domain.md -> qa/ref/002/dialer/.chr/default/tickets/conversion/CHR-104-add-prospect-domain.md
- R093 qa/ref/002/dialer/.chr/user/tickets/conversion/CHR-105-conversion-endpoint.md -> qa/ref/002/dialer/.chr/default/tickets/conversion/CHR-105-conversion-endpoint.md
- R092 qa/ref/002/dialer/.chr/user/tickets/conversion/CHR-106-conversion-funnel-metrics.md -> qa/ref/002/dialer/.chr/default/tickets/conversion/CHR-106-conversion-funnel-metrics.md
- R093 qa/ref/002/dialer/.chr/user/tickets/conversion/CHR-107-reject-invalid-conversions.md -> qa/ref/002/dialer/.chr/default/tickets/conversion/CHR-107-reject-invalid-conversions.md
- R094 qa/ref/002/dialer/.chr/user/tickets/followups/CHR-108-follow-up-domain.md -> qa/ref/002/dialer/.chr/default/tickets/followups/CHR-108-follow-up-domain.md
- R095 qa/ref/002/dialer/.chr/user/tickets/followups/CHR-109-queue-time-source-decision.md -> qa/ref/002/dialer/.chr/default/tickets/followups/CHR-109-queue-time-source-decision.md
- R093 qa/ref/002/dialer/.chr/user/tickets/followups/CHR-110-due-followups-return-to-queue.md -> qa/ref/002/dialer/.chr/default/tickets/followups/CHR-110-due-followups-return-to-queue.md
- R094 qa/ref/002/dialer/.chr/user/tickets/followups/CHR-111-gallery-sample-data.md -> qa/ref/002/dialer/.chr/default/tickets/followups/CHR-111-gallery-sample-data.md
- R093 qa/ref/002/dialer/.chr/user/tickets/followups/CHR-112-calendar-export-spike.md -> qa/ref/002/dialer/.chr/default/tickets/followups/CHR-112-calendar-export-spike.md
- R095 qa/ref/002/dialer/.chr/user/tickets/queue/CHR-101-filter-do-not-call-leads.md -> qa/ref/002/dialer/.chr/default/tickets/queue/CHR-101-filter-do-not-call-leads.md
- R094 qa/ref/002/dialer/.chr/user/tickets/queue/CHR-102-add-callback-visibility-tests.md -> qa/ref/002/dialer/.chr/default/tickets/queue/CHR-102-add-callback-visibility-tests.md
- R093 qa/ref/002/dialer/.chr/user/tickets/queue/CHR-103-queue-summary-metadata.md -> qa/ref/002/dialer/.chr/default/tickets/queue/CHR-103-queue-summary-metadata.md
- R100 qa/ref/003/tictactoe/.chr/user/acceptance/cli-gameplay.feature -> qa/ref/003/tictactoe/.chr/default/acceptance/cli-gameplay.feature
- R100 qa/ref/003/tictactoe/.chr/user/cases/cli-gameplay.md -> qa/ref/003/tictactoe/.chr/default/cases/cli-gameplay.md
- R100 qa/ref/003/tictactoe/.chr/user/config.yml -> qa/ref/003/tictactoe/.chr/default/config.yml
- R095 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/dumps/cli-gameplay-context-raw.md -> qa/ref/003/tictactoe/.chr/default/dumps/cli-gameplay-context-raw.md
- R095 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/dumps/cli-gameplay-context.md -> qa/ref/003/tictactoe/.chr/default/dumps/cli-gameplay-context.md
- R094 qa/ref/003/tictactoe/.chr/user/dumps/project-context-raw.md -> qa/ref/003/tictactoe/.chr/default/dumps/project-context-raw.md
- R094 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/dumps/project-context.md -> qa/ref/003/tictactoe/.chr/default/dumps/project-context.md
- R087 qa/ref/003/tictactoe/.chr/user/evidence-examples/behavior-gherkin.md -> qa/ref/003/tictactoe/.chr/default/evidence-examples/behavior-gherkin.md
- R087 qa/ref/003/tictactoe/.chr/user/evidence-examples/blocked-validation.md -> qa/ref/003/tictactoe/.chr/default/evidence-examples/blocked-validation.md
- R100 qa/ref/003/tictactoe/.chr/user/evidence-examples/docs-no-tests.md -> qa/ref/003/tictactoe/.chr/default/evidence-examples/docs-no-tests.md
- R095 qa/ref/003/tictactoe/.chr/user/fixtures/serialized-loop-tickets/CHR-1-board-model.md -> qa/ref/003/tictactoe/.chr/default/fixtures/serialized-loop-tickets/CHR-1-board-model.md
- R095 qa/ref/003/tictactoe/.chr/user/fixtures/serialized-loop-tickets/CHR-2-cli-game-loop.md -> qa/ref/003/tictactoe/.chr/default/fixtures/serialized-loop-tickets/CHR-2-cli-game-loop.md
- R075 qa/ref/003/tictactoe/.chr/user/fixtures/serialized-loop-tickets/README.md -> qa/ref/003/tictactoe/.chr/default/fixtures/serialized-loop-tickets/README.md
- R100 qa/ref/003/tictactoe/.chr/user/manifest.md -> qa/ref/003/tictactoe/.chr/default/manifest.md
- R100 qa/ref/003/tictactoe/.chr/user/project.md -> qa/ref/003/tictactoe/.chr/default/project.md
- R096 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/quizzes/cli-gameplay-quiz-raw.md -> qa/ref/003/tictactoe/.chr/default/quizzes/cli-gameplay-quiz-raw.md
- R091 qa/ref/003/tictactoe/.chr/user/quizzes/cli-gameplay-quiz.md -> qa/ref/003/tictactoe/.chr/default/quizzes/cli-gameplay-quiz.md
- R100 qa/ref/003/tictactoe/.chr/user/tickets/.gitkeep -> qa/ref/003/tictactoe/.chr/default/tickets/.gitkeep
- M qa/ref/003/tictactoe/README.md
- M scripts/qa/gate.sh
- M scripts/qa/smoke.sh

## Agent-Managed Work

- none

## Manual or External Changes

- M .agents/ARCHITECTURE.md
- R100 .chr/default/agents/ARCHITECTURE.md
- R100 .chr/default/agents/CONVENTIONS.md
- R100 .chr/default/agents/DOCUMENTING.md
- R100 .chr/default/agents/INVARIANTS.md
- R100 .chr/default/agents/LIFECYCLE.md
- R100 .chr/default/agents/LINTING.md
- R083 .chr/default/agents/PLAYBOOKS.md
- R100 .chr/default/agents/ROUTING.md
- R100 .chr/default/agents/SLICES.md
- R100 .chr/default/agents/VERSIONING.md
- R100 .chr/default/agents/playbooks/dump.md
- R100 .chr/default/agents/playbooks/planned-delivery-workflow.md
- R100 .chr/default/agents/playbooks/quiz.md
- R100 .chr/default/agents/playbooks/release-alignment.md
- R100 .chr/default/agents/playbooks/slice-delivery.md
- R100 .chr/default/agents/playbooks/specification-workflow.md
- R100 .chr/default/agents/playbooks/ticket-workflow.md
- R100 .chr/default/agents/playbooks/version-control-workflow.md
- R100 .chr/default/cases/context-capture.md
- R095 .chr/default/cases/dev-main-gate.md
- R100 .chr/default/cases/qa-scenarios.md
- R100 .chr/default/cases/spec-to-tickets.md
- R100 .chr/default/cases/ticket-execution-loop.md
- R100 .chr/default/cases/tui-operations.md
- R100 .chr/default/config.yml
- R100 .chr/default/config/gates.yml
- R085 .chr/default/docs/adoption-baseline.md
- R069 .chr/default/docs/legacy-material-map.md
- R098 .chr/default/dumps/curated/2026-06-02-merger-agent-role.md
- R097 .chr/default/dumps/curated/dev-main-gate-context.md
- R100 .chr/default/dumps/curated/document-generation-context.md
- R087 .chr/default/dumps/curated/project-context.md
- R100 .chr/default/dumps/curated/runtime-execution-context.md
- R100 .chr/default/dumps/raw/2026-06-02-merger-agent-role-raw.md
- R100 .chr/default/manifest.md
- R100 .chr/default/plans/dev-main-gate-plan.md
- R090 .chr/default/plans/merger-agent-workflow-plan.md
- R092 .chr/default/plans/ticket-review-integration-plan.md
- R085 .chr/default/project.md
- R095 .chr/default/quizzes/curated/ticket-review-integration-quiz.md
- R098 .chr/default/quizzes/raw/ticket-review-integration-quiz-raw.md
- R100 .chr/default/runtime/agent/Dockerfile
- R092 .chr/default/specs/dev-main-gate-spec.md
- R100 .chr/default/specs/merger-agent-workflow.md
- R100 .chr/default/specs/ticket-review-integration-draft.md
- R100 .chr/default/trackers/merger-agent-workflow-tracker.md
- R100 .chr/default/trackers/ticket-review-integration-tracker.md
- M .gitignore
- M .prj/fixtures/mvp-agent-proving-ground/todolist/.chr/config.yml
- R100 .prj/fixtures/mvp-agent-proving-ground/todolist/.chr/default/runtime/agent/Dockerfile
- M .prj/fixtures/mvp-agent-proving-ground/todolist/README.md
- M .prj/plans/slicing-board-ticket-operations-plan.md
- M .prj/plans/slicing-codex-artifact-generation-plan.md
- M .prj/plans/slicing-llm-file-tracing-plan.md
- M .prj/plans/slicing-merger-agent-workflow-plan.md
- M .prj/plans/slicing-qa-agent-work-scenario-plan.md
- M .prj/plans/slicing-ticket-review-integration-plan.md
- M .prj/quizzes/curated/dashboard-board-ticket-operations-quiz.md
- M .prj/quizzes/curated/ticket-review-integration-quiz.md
- M .prj/quizzes/raw/dashboard-board-ticket-operations-quiz-raw.md
- M .prj/quizzes/raw/ticket-review-integration-quiz-raw.md
- M .prj/reports/llm-trace-payloads-2026-05-27.md
- M .prj/reports/slices/agent-test-evidence-strategy-pass/slice-1-evidence-strategy-config.md
- M .prj/reports/slices/agent-test-evidence-strategy-pass/slice-5-gherkin-test-policy.md
- M .prj/reports/slices/agent-test-evidence-strategy-pass/slice-6-evidence-visibility-docs.md
- M .prj/reports/slices/agent-work-observability/slice-2-work-lifecycle-events.md
- M .prj/reports/slices/board-ticket-operations/slice-12-board-docs-validation.md
- M .prj/reports/slices/board-ticket-operations/slice-2-userspace-resolver.md
- M .prj/reports/slices/codex-artifact-generation/slice-1-plan-tracking-codex-backend.md
- M .prj/reports/slices/codex-artifact-generation/slice-4-artifact-run-bundle-audit.md
- M .prj/reports/slices/dev-main-agent-evidence-gate/slice-1-gate-command-git-evidence.md
- M .prj/reports/slices/dev-main-agent-evidence-gate/slice-2-validation-report-output.md
- M .prj/reports/slices/dev-main-agent-evidence-gate/slice-4-gate-policy-configuration.md
- M .prj/reports/slices/dev-main-agent-evidence-gate/slice-5-forgejo-woodpecker-integration.md
- M .prj/reports/slices/dev-main-agent-evidence-gate/slice-5-forgejo-woodpecker-investigation.md
- M .prj/reports/slices/gherkin-acceptance-generation/slice-3-feature-artifact-checks.md
- M .prj/reports/slices/gherkin-acceptance-generation/slice-4-spec-feature-generation.md
- M .prj/reports/slices/gherkin-acceptance-generation/slice-5-acceptance-workflow-docs.md
- M .prj/reports/slices/isolated-ticket-worktrees/slice-1-workspace-execution-config.md
- M .prj/reports/slices/llm-file-tracing/slice-1-llm-trace-sink.md
- M .prj/reports/slices/llm-file-tracing/slice-2-quiz-model-trace-integration.md
- M .prj/reports/slices/mvp-agent-proving-ground/slice-1-proving-ground-fixture.md
- M .prj/reports/slices/qa-agent-work-scenario/slice-3-seeded-chr-context-adr-docs.md
- M .prj/reports/slices/separate-acceptance-feature-workflow/slice-1-spec-feature-decoupling.md
- M .prj/reports/slices/separate-acceptance-feature-workflow/slice-2-acceptance-feature-service.md
- M .prj/reports/slices/separate-acceptance-feature-workflow/slice-4-feature-round-defaults.md
- M .prj/reports/slices/separate-acceptance-feature-workflow/slice-6-acceptance-workflow-docs-and-fixture.md
- M .prj/reports/slices/serialized-integration-loop/slice-3-high-level-agentic-review.md
- M .prj/reports/slices/serialized-integration-loop/slice-6-observability-docs-and-qa.md
- M .prj/reports/slices/ticket-review-integration/slice-3-review-mode-configuration.md
- M .prj/reports/slices/ticket-review-integration/slice-4-forge-provider-diagnostics.md
- M .prj/reports/slices/ticket-review-integration/slice-7-serialized-work-preflight.md
- M .prj/reports/slices/tictactoe-qa-reference/slice-1-tictactoe-qa-reference.md
- M .prj/scripts/probe-codex-artifact-generation.sh
- M .prj/specs/drafts/ticket-review-integration-draft.md
- M .prj/specs/specs/board-ticket-operations-model.md
- M .prj/specs/specs/board-ticket-operations.md
- M .prj/specs/specs/merger-agent-workflow.md
- M .prj/specs/specs/qa-agent-work-scenario.md
- M .prj/trackers/slicing-board-ticket-operations-tracker.md
- M .prj/trackers/slicing-merger-agent-workflow-tracker.md
- M .prj/trackers/slicing-qa-agent-work-scenario-tracker.md
- M .prj/trackers/slicing-ticket-review-integration-tracker.md
- M CHANGELOG.md
- M cmd/chr/assets/workspace/profiles/standard/agents/PLAYBOOKS.md
- M cmd/chr/assets/workspace/profiles/standard/root/AGENTS.md
- M docs/explanation/architecture.md
- M docs/explanation/dev-main-gate-direction.md
- M docs/explanation/dev-main-gate-model.md
- M docs/explanation/dev-main-tooling-gate-spec-draft.md
- M docs/explanation/old-gallery.md
- M docs/explanation/tui-interactive-design-notes.md
- M docs/how-to/adopt-dev-main-gate-in-existing-project.md
- M docs/how-to/full-system-lab-runbook.md
- M docs/how-to/qa-tictactoe-agent-work-scenario.md
- M docs/how-to/use-dev-main-gate-with-woodpecker.md
- M docs/how-to/use-forgejo-ticket-review.md
- M docs/how-to/user-guide/index.md
- M docs/reference/acceptance-feature-generation.md
- M docs/reference/authoring-sessions.md
- M docs/reference/commands.md
- M docs/reference/dev-main-gate.md
- M docs/reference/evidence-strategy-config.md
- M docs/reference/index-rebuild-cli.md
- M docs/reference/index-status-cli.md
- M docs/reference/notification-config.md
- M docs/reference/project-workspace.md
- M docs/reference/review-workflow-config.md
- M docs/reference/runtime-config.md
- M docs/reference/ticket-execution.md
- M docs/reference/ticket-file-discovery.md
- M docs/reference/ticket-index-rebuild.md
- M docs/reference/ticket-index-refresh.md
- M docs/reference/ticket-index-status.md
- M docs/reference/ticket-reconciliation.md
- M docs/reference/ticket-review-workflow.md
- M docs/reference/tui.md
- M docs/reference/user-environment.md
- M docs/reference/work-readiness.md
- M docs/tutorials/getting-started.md
- M internal/app/cli_test.go
- M internal/app/planning_baseline_test.go
- M internal/app/service_acceptance_feature_test.go
- M internal/app/service_dump_curation_test.go
- M internal/app/service_plan_test.go
- M internal/app/service_quiz_test.go
- M internal/app/service_spec_test.go
- M internal/app/service_test.go
- M internal/app/service_work_test.go
- M internal/app/tui_adapter_parity_test.go
- M internal/core/acceptance_feature_test.go
- M internal/core/agent_event_readback_test.go
- M internal/core/agent_event_test.go
- M internal/core/agent_run_test.go
- M internal/core/agentic_review_test.go
- M internal/core/assisted_artifact_synthesis_test.go
- M internal/core/assisted_context_test.go
- M internal/core/authoring_artifact_layout_test.go
- M internal/core/authoring_contextual_quiz_test.go
- M internal/core/authoring_session_test.go
- M internal/core/changed_file_discovery_test.go
- M internal/core/codex_worker_context_test.go
- M internal/core/codex_worker_test.go
- M internal/core/config.go
- M internal/core/config_test.go
- M internal/core/context_agent_protocols_test.go
- M internal/core/context_composer_test.go
- M internal/core/context_materializer_test.go
- M internal/core/context_previous_attempt_test.go
- M internal/core/context_project_documents_test.go
- M internal/core/context_readiness_handoff_test.go
- M internal/core/context_tree_search_test.go
- M internal/core/diagnostic_bundle_test.go
- M internal/core/evidence_policy.go
- M internal/core/evidence_policy_test.go
- M internal/core/evidence_readiness_test.go
- M internal/core/evidence_strategy_checklist_test.go
- M internal/core/evidence_strategy_summary_test.go
- M internal/core/execution_artifact_drift_test.go
- M internal/core/execution_report_test.go
- M internal/core/execution_repository_test.go
- M internal/core/execution_runner_test.go
- M internal/core/execution_test.go
- M internal/core/gate_report_test.go
- M internal/core/harness_invocation_test.go
- M internal/core/llm_trace_test.go
- M internal/core/plan_generation_test.go
- M internal/core/project_userspace_test.go
- M internal/core/proving_test.go
- M internal/core/quiz_artifact_repository_test.go
- M internal/core/quiz_model_test.go
- M internal/core/quiz_runtime_test.go
- M internal/core/result_application_test.go
- M internal/core/runtime_init_test.go
- M internal/core/runtime_plan.go
- M internal/core/runtime_plan_test.go
- M internal/core/runtime_validate_test.go
- M internal/core/spec_generation.go
- M internal/core/spec_generation_test.go
- M internal/core/ticket_discovery_test.go
- M internal/core/ticket_index_rebuild_test.go
- M internal/core/ticket_index_status_test.go
- M internal/core/ticket_index_test.go
- M internal/core/ticket_merge_preflight_test.go
- M internal/core/ticket_parser_test.go
- M internal/core/ticket_reconciliation_test.go
- M internal/core/ticket_repository_test.go
- M internal/core/ticket_test.go
- M internal/core/ticket_validation_test.go
- M internal/core/work_readiness_service.go
- M internal/core/work_readiness_service_test.go
- M internal/core/workspace_context_test.go
- M internal/core/workspace_graph_test.go
- M internal/core/workspace_index_cache_test.go
- M internal/core/workspace_index_test.go
- M internal/core/workspace_init.go
- M internal/core/workspace_init_test.go
- M internal/core/worktree_service_test.go
- M internal/tui/documents.go
- M internal/tui/model_test.go
- M internal/tui/source_actions.go
- M internal/tui/work_readiness_test.go
- M internal/webui/webui_test.go
- M ops/default/dump/00001-curated/artifact.md
- M ops/default/dump/00001-curated/files/legacy/2026-06-02-merger-agent-role.md
- M ops/default/dump/00001-curated/files/legacy/dev-main-gate-context.md
- M ops/default/dump/00001-curated/files/legacy/project-context.md
- M ops/default/dump/00001-curated/metadata.json
- M ops/default/dump/00002-raw/artifact.md
- M ops/default/dump/00002-raw/metadata.json
- M ops/default/plan/00008-dev-main-gate-plan-md/artifact.md
- M ops/default/plan/00008-dev-main-gate-plan-md/metadata.json
- M ops/default/plan/00009-merger-agent-workflow-plan-md/artifact.md
- M ops/default/plan/00009-merger-agent-workflow-plan-md/files/legacy/merger-agent-workflow-plan.md
- M ops/default/plan/00009-merger-agent-workflow-plan-md/metadata.json
- M ops/default/plan/00010-ticket-review-integration-plan-md/artifact.md
- M ops/default/plan/00010-ticket-review-integration-plan-md/files/legacy/ticket-review-integration-plan.md
- M ops/default/plan/00010-ticket-review-integration-plan-md/metadata.json
- M ops/default/plan/00026-2026-06-01-dev-main-agent-evidence-gate-plan-md/files/legacy/2026-06-01-dev-main-agent-evidence-gate-plan.md
- M ops/default/plan/00028-2026-06-06-separate-acceptance-feature-workflow-plan-md/files/legacy/2026-06-06-separate-acceptance-feature-workflow-plan.md
- M ops/default/plan/00031-2026-06-18-tictactoe-qa-reference-plan-md/files/legacy/2026-06-18-tictactoe-qa-reference-plan.md
- M ops/default/plan/00032-2026-06-19-agent-test-evidence-strategy-pass-plan-md/files/legacy/2026-06-19-agent-test-evidence-strategy-pass-plan.md
- M ops/default/quiz/00003-curated/artifact.md
- M ops/default/quiz/00003-curated/files/legacy/ticket-review-integration-quiz.md
- M ops/default/quiz/00003-curated/metadata.json
- M ops/default/quiz/00004-raw/artifact.md
- M ops/default/quiz/00004-raw/files/legacy/ticket-review-integration-quiz-raw.md
- M ops/default/quiz/00004-raw/metadata.json
- M ops/default/report/00036-ops-layout-migration/files/manifest.json
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/cases/cli-gameplay.md
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/config.yml
- R095 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/dumps/cli-gameplay-context-raw.md
- R095 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/dumps/cli-gameplay-context.md
- R094 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/dumps/project-context-raw.md
- R094 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/dumps/project-context.md
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/manifest.md
- R098 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/plans/archive/cli-gameplay-plan-20260618-084729438415104.md
- R098 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/plans/cli-gameplay-plan.md
- R082 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/plans/cli-gameplay-tracking.md
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/project.md
- R096 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/quizzes/cli-gameplay-quiz-raw.md
- R091 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/quizzes/cli-gameplay-quiz.md
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/agent/Dockerfile
- R074 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/plan-generation.log
- R055 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-1/exec-648ea4e465b090f7/context.md
- R069 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-1/exec-648ea4e465b090f7/execution-report.md
- A ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-1/exec-648ea4e465b090f7/result-application.md
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-1/exec-648ea4e465b090f7/result.json
- R098 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-1/exec-648ea4e465b090f7/runtime.log
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-1/exec-648ea4e465b090f7/validation.log
- R055 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-2/exec-409755cd38adfe67/context.md
- R070 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-2/exec-409755cd38adfe67/execution-report.md
- A ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-2/exec-409755cd38adfe67/result-application.md
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-2/exec-409755cd38adfe67/result.json
- R087 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-2/exec-409755cd38adfe67/runtime.log
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-2/exec-409755cd38adfe67/validation.log
- R055 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/context.md
- R070 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/execution-report.md
- A ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/result-application.md
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/result.json
- R083 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/runtime.log
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/validation.log
- R054 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-4/exec-6326a1544f559788/context.md
- R071 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-4/exec-6326a1544f559788/execution-report.md
- A ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-4/exec-6326a1544f559788/result-application.md
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-4/exec-6326a1544f559788/result.json
- R096 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-4/exec-6326a1544f559788/runtime.log
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-4/exec-6326a1544f559788/validation.log
- R055 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-5/exec-7fe25b30b4955e1e/context.md
- R070 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-5/exec-7fe25b30b4955e1e/execution-report.md
- A ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-5/exec-7fe25b30b4955e1e/result-application.md
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-5/exec-7fe25b30b4955e1e/result.json
- R095 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-5/exec-7fe25b30b4955e1e/runtime.log
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/runtime/runs/CHR-5/exec-7fe25b30b4955e1e/validation.log
- R098 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/specs/project-spec.md
- R100 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/tickets/.gitkeep
- R085 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/tickets/cli-gameplay/CHR-1-Game-outcome-rules.md
- R087 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/tickets/cli-gameplay/CHR-2-Move-state-transitions.md
- R087 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/tickets/cli-gameplay/CHR-3-CLI-board-rendering.md
- R088 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/tickets/cli-gameplay/CHR-4-Interactive-gameplay-loop.md
- R088 ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/default/tickets/cli-gameplay/CHR-5-Test-coverage-hardening.md
- D ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-1/exec-648ea4e465b090f7/result-application.md
- D ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-2/exec-409755cd38adfe67/result-application.md
- D ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/result-application.md
- D ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-4/exec-6326a1544f559788/result-application.md
- D ops/default/report/00037-tictactoe-agent-work-success/files/legacy/tictactoe-agent-work-success/project/.chr/user/runtime/runs/CHR-5/exec-7fe25b30b4955e1e/result-application.md
- M ops/default/report/00037-tictactoe-agent-work-success/metadata.json
- M ops/default/spec/00005-dev-main-gate-spec-md/artifact.md
- M ops/default/spec/00005-dev-main-gate-spec-md/files/legacy/dev-main-gate-spec.md
- M ops/default/spec/00005-dev-main-gate-spec-md/metadata.json
- M ops/default/spec/00006-merger-agent-workflow-md/artifact.md
- M ops/default/spec/00006-merger-agent-workflow-md/metadata.json
- M ops/default/spec/00007-ticket-review-integration-draft-md/artifact.md
- M ops/default/spec/00007-ticket-review-integration-draft-md/metadata.json
- M ops/default/spec/00011-2026-06-01-dev-main-agent-evidence-gate-md/files/legacy/2026-06-01-dev-main-agent-evidence-gate.md
- M ops/default/spec/00012-2026-06-06-gherkin-acceptance-generation-md/files/legacy/2026-06-06-gherkin-acceptance-generation.md
- M ops/default/spec/00013-2026-06-06-separate-acceptance-feature-workflow-md/files/legacy/2026-06-06-separate-acceptance-feature-workflow.md
- M ops/default/spec/00014-2026-06-18-agent-work-observability-model-md/files/legacy/2026-06-18-agent-work-observability-model.md
- M ops/default/spec/00018-2026-06-18-configurable-work-readiness-md/files/legacy/2026-06-18-configurable-work-readiness.md
- M ops/default/spec/00019-2026-06-19-agent-test-evidence-strategy-pass-md/files/legacy/2026-06-19-agent-test-evidence-strategy-pass.md
- M ops/default/spec/00020-2026-06-19-isolated-ticket-worktrees-md/files/legacy/2026-06-19-isolated-ticket-worktrees.md
- M ops/default/spec/00021-2026-06-20-serialized-integration-loop-model-md/files/legacy/2026-06-20-serialized-integration-loop-model.md
- M ops/default/spec/00022-2026-06-20-serialized-integration-loop-md/files/legacy/2026-06-20-serialized-integration-loop.md
- M ops/default/spec/00023-2026-06-29-execution-unit-runtime-refinements-md/files/legacy/2026-06-29-execution-unit-runtime-refinements.md
- M ops/default/spec/00024-2026-06-29-optional-local-web-ui-md/files/legacy/2026-06-29-optional-local-web-ui.md
- M ops/default/spec/00025-drafts/files/legacy/2026-06-01-dev-main-agent-evidence-gate.md
- M ops/default/spec/00025-drafts/files/legacy/2026-06-06-gherkin-acceptance-generation.md
- M ops/plans/2026-06-01-dev-main-agent-evidence-gate-plan.md
- M ops/plans/2026-06-06-separate-acceptance-feature-workflow-plan.md
- M ops/plans/2026-06-18-tictactoe-qa-reference-plan.md
- M ops/plans/2026-06-19-agent-test-evidence-strategy-pass-plan.md
- M ops/specs/2026-06-01-dev-main-agent-evidence-gate.md
- M ops/specs/2026-06-06-gherkin-acceptance-generation.md
- M ops/specs/2026-06-06-separate-acceptance-feature-workflow.md
- M ops/specs/2026-06-18-agent-work-observability-model.md
- M ops/specs/2026-06-18-configurable-work-readiness.md
- M ops/specs/2026-06-19-agent-test-evidence-strategy-pass.md
- M ops/specs/2026-06-19-isolated-ticket-worktrees.md
- M ops/specs/2026-06-20-serialized-integration-loop-model.md
- M ops/specs/2026-06-20-serialized-integration-loop.md
- M ops/specs/2026-06-29-execution-unit-runtime-refinements.md
- M ops/specs/2026-06-29-optional-local-web-ui.md
- M ops/trackers/2026-06-01-dev-main-agent-evidence-gate-tracker.md
- M ops/trackers/2026-06-06-gherkin-acceptance-generation-tracker.md
- M ops/trackers/2026-06-06-separate-acceptance-feature-workflow-tracker.md
- M ops/trackers/2026-06-19-agent-test-evidence-strategy-pass-tracker.md
- M ops/trackers/2026-06-20-serialized-integration-loop-tracker.md
- R100 qa/ref/001/blog/.chr/default/cases/comments.md
- R100 qa/ref/001/blog/.chr/default/config.yml
- R100 qa/ref/001/blog/.chr/default/dumps/comments-context.md
- R100 qa/ref/001/blog/.chr/default/manifest.md
- R100 qa/ref/001/blog/.chr/default/project.md
- R094 qa/ref/001/blog/.chr/default/quizzes/comments-quiz-raw.md
- R091 qa/ref/001/blog/.chr/default/quizzes/comments-quiz.md
- R100 qa/ref/001/blog/.chr/default/tickets/CHR-1-fix-unknown-post-status.md
- R100 qa/ref/002/dialer/.chr/default/cases/call-outcome-capture.md
- R100 qa/ref/002/dialer/.chr/default/cases/conversion.md
- R100 qa/ref/002/dialer/.chr/default/cases/followups.md
- R100 qa/ref/002/dialer/.chr/default/cases/lead-queue.md
- R100 qa/ref/002/dialer/.chr/default/config.yml
- R100 qa/ref/002/dialer/.chr/default/dumps/call-outcome-capture-context.md
- R100 qa/ref/002/dialer/.chr/default/dumps/conversion-context.md
- R100 qa/ref/002/dialer/.chr/default/dumps/followups-context.md
- R100 qa/ref/002/dialer/.chr/default/dumps/lead-queue-context.md
- R100 qa/ref/002/dialer/.chr/default/dumps/project-context.md
- R100 qa/ref/002/dialer/.chr/default/manifest.md
- R100 qa/ref/002/dialer/.chr/default/plans/conversion-plan.md
- R090 qa/ref/002/dialer/.chr/default/plans/conversion-tracking.md
- R100 qa/ref/002/dialer/.chr/default/plans/followups-plan.md
- R091 qa/ref/002/dialer/.chr/default/plans/followups-tracking.md
- R100 qa/ref/002/dialer/.chr/default/plans/lead-queue-plan.md
- R090 qa/ref/002/dialer/.chr/default/plans/lead-queue-tracking.md
- R100 qa/ref/002/dialer/.chr/default/project.md
- R084 qa/ref/002/dialer/.chr/default/quizzes/conversion-quiz-raw.md
- R090 qa/ref/002/dialer/.chr/default/quizzes/conversion-quiz.md
- R082 qa/ref/002/dialer/.chr/default/quizzes/followups-quiz-raw.md
- R088 qa/ref/002/dialer/.chr/default/quizzes/followups-quiz.md
- R085 qa/ref/002/dialer/.chr/default/quizzes/lead-queue-quiz-raw.md
- R090 qa/ref/002/dialer/.chr/default/quizzes/lead-queue-quiz.md
- R100 qa/ref/002/dialer/.chr/default/specs/conversion-spec.md
- R100 qa/ref/002/dialer/.chr/default/specs/followups-spec.md
- R100 qa/ref/002/dialer/.chr/default/specs/lead-queue-spec.md
- R094 qa/ref/002/dialer/.chr/default/tickets/conversion/CHR-104-add-prospect-domain.md
- R093 qa/ref/002/dialer/.chr/default/tickets/conversion/CHR-105-conversion-endpoint.md
- R092 qa/ref/002/dialer/.chr/default/tickets/conversion/CHR-106-conversion-funnel-metrics.md
- R093 qa/ref/002/dialer/.chr/default/tickets/conversion/CHR-107-reject-invalid-conversions.md
- R094 qa/ref/002/dialer/.chr/default/tickets/followups/CHR-108-follow-up-domain.md
- R095 qa/ref/002/dialer/.chr/default/tickets/followups/CHR-109-queue-time-source-decision.md
- R093 qa/ref/002/dialer/.chr/default/tickets/followups/CHR-110-due-followups-return-to-queue.md
- R094 qa/ref/002/dialer/.chr/default/tickets/followups/CHR-111-gallery-sample-data.md
- R093 qa/ref/002/dialer/.chr/default/tickets/followups/CHR-112-calendar-export-spike.md
- R095 qa/ref/002/dialer/.chr/default/tickets/queue/CHR-101-filter-do-not-call-leads.md
- R094 qa/ref/002/dialer/.chr/default/tickets/queue/CHR-102-add-callback-visibility-tests.md
- R093 qa/ref/002/dialer/.chr/default/tickets/queue/CHR-103-queue-summary-metadata.md
- R100 qa/ref/003/tictactoe/.chr/default/acceptance/cli-gameplay.feature
- R100 qa/ref/003/tictactoe/.chr/default/cases/cli-gameplay.md
- R100 qa/ref/003/tictactoe/.chr/default/config.yml
- R095 qa/ref/003/tictactoe/.chr/default/dumps/cli-gameplay-context-raw.md
- R095 qa/ref/003/tictactoe/.chr/default/dumps/cli-gameplay-context.md
- R094 qa/ref/003/tictactoe/.chr/default/dumps/project-context-raw.md
- R094 qa/ref/003/tictactoe/.chr/default/dumps/project-context.md
- R087 qa/ref/003/tictactoe/.chr/default/evidence-examples/behavior-gherkin.md
- R087 qa/ref/003/tictactoe/.chr/default/evidence-examples/blocked-validation.md
- R100 qa/ref/003/tictactoe/.chr/default/evidence-examples/docs-no-tests.md
- R095 qa/ref/003/tictactoe/.chr/default/fixtures/serialized-loop-tickets/CHR-1-board-model.md
- R095 qa/ref/003/tictactoe/.chr/default/fixtures/serialized-loop-tickets/CHR-2-cli-game-loop.md
- R075 qa/ref/003/tictactoe/.chr/default/fixtures/serialized-loop-tickets/README.md
- R100 qa/ref/003/tictactoe/.chr/default/manifest.md
- R100 qa/ref/003/tictactoe/.chr/default/project.md
- R096 qa/ref/003/tictactoe/.chr/default/quizzes/cli-gameplay-quiz-raw.md
- R091 qa/ref/003/tictactoe/.chr/default/quizzes/cli-gameplay-quiz.md
- R100 qa/ref/003/tictactoe/.chr/default/tickets/.gitkeep
- M qa/ref/003/tictactoe/README.md
- M scripts/qa/gate.sh
- M scripts/qa/smoke.sh

## High-Risk Manual Changes

- none

## Missing Agent Evidence

- none

## Validation

- command: make verify
- passed: false
- exit_code: 2

```text
make[1]: Entering directory '/home/adrian/Projects/chr-operational-layout'
internal/app/ticket.go:1327:1: calculated cyclomatic complexity for function applyTicketResult is 11, max is 10 (cyclop)
func applyTicketResult(ctx context.Context, env core.UserEnvironment, plan ticketApplyResultPlan, events ticketExecutionEvents, assetsFS fs.FS) error {
^
internal/core/ops_artifact.go:35:1: calculated cyclomatic complexity for function CreateOpsArtifact is 14, max is 10 (cyclop)
func CreateOpsArtifact(opts CreateOpsArtifactOptions) (OpsArtifact, error) {
^
internal/core/ops_artifact_test.go:11:1: calculated cyclomatic complexity for function TestCreateOpsArtifactAllocatesGlobalIDAndEnvelope is 11, max is 10 (cyclop)
func TestCreateOpsArtifactAllocatesGlobalIDAndEnvelope(t *testing.T) {
^
internal/core/ops_artifact_test.go:73:1: calculated cyclomatic complexity for function TestRefreshOpsArtifactMetadataIndexesSharedFiles is 11, max is 10 (cyclop)
func TestRefreshOpsArtifactMetadataIndexesSharedFiles(t *testing.T) {
^
internal/core/ops_artifact.go:55:17: Error return value of `os.Remove` is not checked (errcheck)
	defer func() { _ = os.Remove(lockPath) }()
	               ^
internal/core/ops_artifact.go:202:8: Error return value of `fmt.Sscanf` is not checked (errcheck)
				_, _ = fmt.Sscanf(match[1], "%d", &number)
				   ^
internal/core/proving.go:450:3: Error return value of `logFile.Close` is not checked (errcheck)
		_ = logFile.Close()
		^
internal/app/ticket.go:1896:85: non-wrapping format verb for fmt.Errorf. Use `%w` to format errors (errorlint)
				return executed, fmt.Errorf("%w; refresh shared artifact metadata: %v", runErr, err)
				                                                                                ^
internal/core/config_test.go:583:6: Function 'TestValidateWorkConfigRejectsInvalidWorkspaceConfig' is too long (71 > 60) (funlen)
func TestValidateWorkConfigRejectsInvalidWorkspaceConfig(t *testing.T) {
     ^
internal/core/ops_artifact.go:107:1: cognitive complexity 20 of func `EnsureOpsArtifactEnvelope` is high (> 15) (gocognit)
func EnsureOpsArtifactEnvelope(opts CreateOpsArtifactOptions, id string) (OpsArtifact, error) {
^
internal/core/report_listing.go:53:1: cognitive complexity 17 of func `(ReportLister).ListProjectReports` is high (> 15) (gocognit)
func (l ReportLister) ListProjectReports(
^
internal/core/runtime_plan.go:182:1: cognitive complexity 21 of func `runtimeArtifactPaths` is high (> 15) (gocognit)
func runtimeArtifactPaths(
^
internal/core/ops_artifact.go:241:1: File is not properly formatted (gofumpt)
func RefreshOpsArtifactMetadata(dir string, status string, now time.Time) error {
^
internal/core/runtime_init_test.go:32:1: File is not properly formatted (gofumpt)

^
internal/core/runtime_plan.go:243:1: File is not properly formatted (gofumpt)
func runtimeOpsLayout(projectDir string, configuredArtifactDir string, opts RuntimePlanOptions) (OpsLayout, string, bool, error) {
^
internal/app/ticket.go:1893:1: `if executed.Artifacts.ArtifactID != ""` has complex nested blocks (complexity: 3) (nestif)
	if executed.Artifacts.ArtifactID != "" {
^
internal/core/ops_artifact.go:146:1: `if errors.Is(err, os.ErrNotExist)` has complex nested blocks (complexity: 3) (nestif)
	if _, err := os.Stat(metadataPath); errors.Is(err, os.ErrNotExist) {
^
internal/core/runtime_plan.go:193:1: `if canonical` has complex nested blocks (complexity: 10) (nestif)
	if canonical {
^
internal/app/ticket.go:1116:2: return with no blank line before (nlreturn)
	return record, nil
	^
internal/app/ticket.go:1898:4: return with no blank line before (nlreturn)
			return executed, err
			^
internal/core/ops_artifact.go:189:2: return with no blank line before (nlreturn)
	return file.Close()
	^
internal/core/ops_artifact.go:207:2: return with no blank line before (nlreturn)
	return maximum + 1
	^
internal/core/ops_artifact.go:213:2: return with no blank line before (nlreturn)
	return strings.Trim(value, "-")
	^
internal/core/ops_artifact.go:231:3: return with no blank line before (nlreturn)
		return nil
		^
internal/core/ops_artifact.go:237:2: return with no blank line before (nlreturn)
	return files, nil
	^
internal/core/ops_artifact_test.go:123:2: return with no blank line before (nlreturn)
	return true
	^
internal/core/ops_layout_test.go:49:4: return with no blank line before (nlreturn)
			return ""
			^
internal/core/proving.go:451:3: return with no blank line before (nlreturn)
		return nil, err
		^
internal/core/proving.go:771:2: return with no blank line before (nlreturn)
	return EnsureOpsArtifactEnvelope(CreateOpsArtifactOptions{
	^
internal/core/worktree_plan.go:84:2: return with no blank line before (nlreturn)
	return layout.WorktreeRoot(), nil
	^
internal/core/runtime_init_test.go:33:1: unnecessary trailing newline (whitespace)
}
^
internal/core/runtime_init_test.go:56:1: unnecessary trailing newline (whitespace)
}
^
internal/app/agent_work_events.go:16:2: missing whitespace above this line (invalid statement above assign) (wsl_v5)
	config, err := core.LoadWorkspaceConfig(projectDir)
	^
internal/app/service_work.go:318:2: missing whitespace above this line (too many statements above if) (wsl_v5)
	record, err = allocateLoopRunArtifact(s.projectContext.ProjectDir, record)
	^
internal/app/service_work.go:322:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if err := repository.CreateLoopRun(s.ctx, record); err != nil {
	^
internal/app/service_work_test.go:77:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if result.Execution == nil {
	^
internal/app/service_work_test.go:80:2: missing whitespace above this line (invalid statement above expr) (wsl_v5)
	assertPathExists(t, filepath.Join(result.Execution.Artifacts.RunDir, "artifact.md"))
	^
internal/app/ticket.go:944:2: missing whitespace above this line (too many statements above if) (wsl_v5)
	record, err = allocateLoopRunArtifact(projectContext.ProjectDir, record)
	^
internal/app/ticket.go:1005:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if record.ArtifactID != "" {
	^
internal/app/ticket.go:1038:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if record.ArtifactID != "" {
	^
internal/app/ticket.go:1104:2: missing whitespace above this line (invalid statement above assign) (wsl_v5)
	artifact, err := core.CreateOpsArtifact(core.CreateOpsArtifactOptions{
	^
internal/app/ticket.go:1114:2: missing whitespace above this line (invalid statement above assign) (wsl_v5)
	record.ArtifactID = artifact.ID
	^
internal/app/ticket.go:1386:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if plan.Record.Artifacts.ArtifactID != "" {
	^
internal/app/ticket.go:1611:2: missing whitespace above this line (invalid statement above assign) (wsl_v5)
	layout, err := core.ResolveOpsLayout(core.ResolveOpsLayoutOptions{ProjectDir: projectContext.ProjectDir})
	^
internal/app/ticket.go:1615:2: missing whitespace above this line (invalid statement above assign) (wsl_v5)
	artifact, err := core.CreateOpsArtifact(core.CreateOpsArtifactOptions{
	^
internal/app/ticket.go:1625:2: missing whitespace above this line (invalid statement above assign) (wsl_v5)
	record.Artifacts.ArtifactID = artifact.ID
	^
internal/core/agent_event_stream.go:57:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if strings.TrimSpace(record.Artifacts.ArtifactID) != "" {
	^
internal/core/agent_event_stream.go:69:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if strings.TrimSpace(record.ArtifactID) != "" {
	^
internal/core/agentic_review_test.go:93:2: missing whitespace above this line (invalid statement above expr) (wsl_v5)
	assertAgenticReviewArtifactExists(t, artifact.ContextPath)
	^
internal/core/execution_report.go:353:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if strings.TrimSpace(record.Artifacts.ArtifactID) != "" {
	^
internal/core/execution_test.go:221:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if record.Artifacts.ArtifactID != plan.ArtifactID {
	^
internal/core/ops_artifact.go:39:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if strings.TrimSpace(opts.Title) == "" {
	^
internal/core/ops_artifact.go:42:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if strings.TrimSpace(opts.Producer) == "" {
	^
internal/core/ops_artifact.go:61:2: missing whitespace above this line (invalid statement above assign) (wsl_v5)
	id := fmt.Sprintf("%05d-%s", nextOpsArtifactNumber(opts.Layout.Root()), slug)
	^
internal/core/ops_artifact.go:62:2: missing whitespace above this line (too many statements above if) (wsl_v5)
	dir, err := opts.Layout.ArtifactDirectory(opts.Type, id)
	^
internal/core/ops_artifact.go:66:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if err := os.MkdirAll(dir, 0o755); err != nil {
	^
internal/core/ops_artifact.go:87:2: missing whitespace above this line (too many statements above if) (wsl_v5)
	encoded, err := json.MarshalIndent(metadata, "", "  ")
	^
internal/core/ops_artifact.go:91:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if err := os.WriteFile(filepath.Join(dir, "artifact.md"), []byte("# "+strings.TrimSpace(opts.Title)+"\n"), 0o644); err != nil {
	^
internal/core/ops_artifact.go:94:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), append(encoded, '\n'), 0o644); err != nil {
	^
internal/core/ops_artifact.go:97:2: missing whitespace above this line (invalid statement above range) (wsl_v5)
	for _, path := range []string{"files", "evidence", filepath.Join("extension", strings.TrimSpace(opts.Producer))} {
	^
internal/core/ops_artifact.go:111:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if strings.TrimSpace(id) == "" || filepath.Base(id) != id {
	^
internal/core/ops_artifact.go:114:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if strings.TrimSpace(opts.Title) == "" {
	^
internal/core/ops_artifact.go:117:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if strings.TrimSpace(opts.Producer) == "" {
	^
internal/core/ops_artifact.go:130:2: missing whitespace above this line (invalid statement above range) (wsl_v5)
	for _, path := range []string{"", "files", "evidence", filepath.Join("extension", strings.TrimSpace(opts.Producer))} {
	^
internal/core/ops_artifact.go:164:3: missing whitespace above this line (too many statements above if) (wsl_v5)
		encoded, err := json.MarshalIndent(metadata, "", "  ")
		^
internal/core/ops_artifact.go:168:3: missing whitespace above this line (invalid statement above if) (wsl_v5)
		if err := os.WriteFile(metadataPath, append(encoded, '\n'), 0o644); err != nil {
		^
internal/core/ops_artifact.go:182:2: missing whitespace above this line (invalid statement above assign) (wsl_v5)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	^
internal/core/ops_artifact.go:186:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if err != nil {
	^
internal/core/ops_artifact.go:194:2: missing whitespace above this line (no shared variables above range) (wsl_v5)
	for artifactType := range opsArtifactTypes {
	^
internal/core/ops_artifact.go:199:3: missing whitespace above this line (invalid statement above range) (wsl_v5)
		for _, entry := range entries {
		^
internal/core/ops_artifact.go:219:2: missing whitespace above this line (too many statements above if) (wsl_v5)
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, walkErr error) error {
	^
internal/core/ops_artifact.go:223:3: missing whitespace above this line (invalid statement above if) (wsl_v5)
		if entry.IsDir() || path == filepath.Join(dir, "metadata.json") || strings.HasPrefix(path, filepath.Join(dir, "extension")+string(filepath.Separator)) {
		^
internal/core/ops_artifact.go:226:3: missing whitespace above this line (invalid statement above assign) (wsl_v5)
		relative, err := filepath.Rel(dir, path)
		^
internal/core/ops_artifact.go:230:3: missing whitespace above this line (invalid statement above assign) (wsl_v5)
		files = append(files, filepath.ToSlash(relative))
		^
internal/core/ops_artifact.go:236:2: missing whitespace above this line (invalid statement above expr) (wsl_v5)
	sort.Strings(files)
	^
internal/core/ops_artifact.go:247:2: missing whitespace above this line (too many statements above if) (wsl_v5)
	content, err := os.ReadFile(path)
	^
internal/core/ops_artifact.go:261:2: missing whitespace above this line (invalid statement above assign) (wsl_v5)
	metadata["files"] = files
	^
internal/core/ops_artifact.go:269:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if err := os.WriteFile(path, append(encoded, '\n'), 0o644); err != nil {
	^
internal/core/ops_artifact_test.go:13:2: missing whitespace above this line (too many statements above if) (wsl_v5)
	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: projectDir})
	^
internal/core/ops_artifact_test.go:17:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if err := os.MkdirAll(filepath.Join(layout.Root(), "plan", "00001-existing"), 0o755); err != nil {
	^
internal/core/ops_artifact_test.go:32:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if artifact.ID != "00002-parser-validation" {
	^
internal/core/ops_artifact_test.go:35:2: missing whitespace above this line (invalid statement above range) (wsl_v5)
	for _, path := range []string{"artifact.md", "metadata.json", "files", "evidence", filepath.Join("extension", "chr")} {
	^
internal/core/ops_artifact_test.go:45:2: missing whitespace above this line (never cuddle decl) (wsl_v5)
	var metadata struct {
	^
internal/core/ops_artifact_test.go:51:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if len(metadata.ExternalIDs) != 1 || metadata.ExternalIDs[0] != "exec-1" {
	^
internal/core/ops_artifact_test.go:61:2: missing whitespace above this line (invalid statement above assign) (wsl_v5)
	lockPath := filepath.Join(layout.Root(), "lock", "id.lock")
	^
internal/core/ops_artifact_test.go:65:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if err := os.WriteFile(lockPath, []byte("busy\n"), 0o644); err != nil {
	^
internal/core/ops_artifact_test.go:68:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if _, err := CreateOpsArtifact(CreateOpsArtifactOptions{Layout: layout, Type: "run", Title: "Parser", Producer: "chr"}); err == nil {
	^
internal/core/ops_artifact_test.go:78:2: missing whitespace above this line (invalid statement above assign) (wsl_v5)
	artifact, err := CreateOpsArtifact(CreateOpsArtifactOptions{Layout: layout, Type: "run", Title: "Parser", Producer: "chr"})
	^
internal/core/ops_artifact_test.go:82:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if err := os.WriteFile(filepath.Join(artifact.Dir, "files", "context.md"), []byte("context\n"), 0o644); err != nil {
	^
internal/core/ops_artifact_test.go:85:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if err := os.WriteFile(filepath.Join(artifact.Dir, "evidence", "result.json"), []byte("{}\n"), 0o644); err != nil {
	^
internal/core/ops_artifact_test.go:88:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if err := os.WriteFile(filepath.Join(artifact.Dir, "extension", "chr", "state.json"), []byte("{}\n"), 0o644); err != nil {
	^
internal/core/ops_artifact_test.go:91:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if err := RefreshOpsArtifactMetadata(artifact.Dir, "succeeded", time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC)); err != nil {
	^
internal/core/ops_artifact_test.go:99:2: missing whitespace above this line (never cuddle decl) (wsl_v5)
	var metadata struct {
	^
internal/core/ops_artifact_test.go:106:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if metadata.Status != "succeeded" {
	^
internal/core/ops_artifact_test.go:109:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if got, want := metadata.Files, []string{"artifact.md", "evidence/result.json", "files/context.md"}; !equalStringSlices(got, want) {
	^
internal/core/ops_artifact_test.go:118:2: missing whitespace above this line (invalid statement above range) (wsl_v5)
	for index := range got {
	^
internal/core/ops_layout.go:57:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if namespace == "" {
	^
internal/core/ops_layout.go:60:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if !IsValidOpsNamespace(namespace) {
	^
internal/core/ops_layout.go:103:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if strings.TrimSpace(id) == "" || filepath.Base(id) != id {
	^
internal/core/ops_layout_test.go:22:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if layout.Root() != filepath.Join(projectDir, "ops", "default") {
	^
internal/core/ops_layout_test.go:73:2: missing whitespace above this line (too many statements above if) (wsl_v5)
	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: projectDir, Namespace: "adrian"})
	^
internal/core/ops_layout_test.go:82:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if want := filepath.Join(projectDir, "ops", "adrian", "run", "00001-parser-validation"); runDir != want {
	^
internal/core/ops_layout_test.go:90:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if want := filepath.Join(projectDir, "ops", "adrian", "worktree", "00001-parser-validation"); worktreeDir != want {
	^
internal/core/ops_layout_test.go:104:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if _, err := layout.ArtifactDirectory("run", "../outside"); err == nil {
	^
internal/core/proving.go:248:2: missing whitespace above this line (invalid statement above assign) (wsl_v5)
	artifact, err := ensureProvingOpsArtifact(projectDir, planned.Objective)
	^
internal/core/proving.go:278:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if err := RefreshOpsArtifactMetadata(artifact.Dir, provingValidationArtifactStatus(result), time.Now()); err != nil {
	^
internal/core/proving.go:449:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if err := RefreshOpsArtifactMetadata(artifact.Dir, "running", time.Now()); err != nil {
	^
internal/core/proving.go:835:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if err := RefreshOpsArtifactMetadata(artifact.Dir, result.Status, time.Now()); err != nil {
	^
internal/core/report_listing.go:72:2: missing whitespace above this line (invalid statement above assign) (wsl_v5)
	entries := []ReportListEntry{}
	^
internal/core/report_listing_test.go:142:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if reports[0].ExecutionID != "adrian-exec" {
	^
internal/core/runtime_plan.go:87:2: missing whitespace above this line (too many statements above if) (wsl_v5)
	artifactDir, runArtifactDir, workerContextPath, resultPath, logPath, validationPath, err := runtimeArtifactPaths(projectDir, executionWorkspaceDir, config.ArtifactDir, opts)
	^
internal/core/runtime_plan.go:227:2: missing whitespace above this line (too many statements above if) (wsl_v5)
	runArtifactDir := artifactDir
	^
internal/core/runtime_plan_test.go:149:2: missing whitespace above this line (too many statements above if) (wsl_v5)
	layout, err := ResolveOpsLayout(ResolveOpsLayoutOptions{ProjectDir: dir})
	^
internal/core/runtime_plan_test.go:169:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if plan.WorkerContextPath != filepath.Join(wantRunDir, "files", "context.md") {
	^
internal/core/runtime_plan_test.go:172:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if plan.ResultPath != filepath.Join(wantRunDir, "evidence", "result.json") {
	^
internal/core/work_config.go:66:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if filepath.ToSlash(filepath.Clean(config.WorktreeRoot)) != DefaultWorktreeRoot {
	^
internal/core/work_readiness_service.go:344:2: missing whitespace above this line (no shared variables above range) (wsl_v5)
	for _, pattern := range runtimeArtifactIgnorePatterns() {
	^
internal/core/work_readiness_service.go:349:3: missing whitespace above this line (invalid statement above assign) (wsl_v5)
		changed = changed || patternChanged
		^
internal/core/work_selection_preflight_test.go:22:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if want := filepath.Join(projectDir, OpsDirectoryName, DefaultOpsNamespace, "lock", "agent-operation.lock"); lock.path != want {
	^
internal/core/worktree_service_test.go:42:2: missing whitespace above this line (invalid statement above assign) (wsl_v5)
	dir := t.TempDir()
	^
internal/core/worktree_service_test.go:52:2: missing whitespace above this line (invalid statement above if) (wsl_v5)
	if got, want := plan.WorktreeRoot, filepath.Join(dir, "ops", "adrian", "worktree"); got != want {
	^
121 issues:
* cyclop: 4
* errcheck: 3
* errorlint: 1
* funlen: 1
* gocognit: 3
* gofumpt: 3
* nestif: 3
* nlreturn: 12
* whitespace: 2
* wsl_v5: 89
make[1]: *** [Makefile:36: lint-strict] Error 1
make[1]: Leaving directory '/home/adrian/Projects/chr-operational-layout'
```

## Failures

- validation failed

## Warnings

- manual or external changes present

## Info

- none

## Artifacts

- json: /home/adrian/Projects/chr-operational-layout/.chr/default/reports/gates/dev-main-20260714T144245Z.json
- markdown: /home/adrian/Projects/chr-operational-layout/.chr/default/reports/gates/dev-main-20260714T144245Z.md

