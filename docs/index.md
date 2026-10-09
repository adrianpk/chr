# Documentation

Chryso documentation follows Diataxis.

## Tutorials

Learning-oriented walkthroughs:

- [Getting started](tutorials/getting-started.md)

## How-to

Task-oriented guides:

- [User guide](how-to/user-guide/index.md)
- [Run the local Web UI](how-to/run-local-web-ui.md)
- [Expose the local Web UI deliberately](how-to/expose-local-web-ui.md)
- [Full-system lab runbook](how-to/full-system-lab-runbook.md)
- [Observe agent work](how-to/observe-agent-work.md)
- [Find execution and loop evidence](how-to/find-execution-and-loop-evidence.md)
- [Inspect and recover local state](how-to/inspect-and-recover-local-state.md)
- [Use Forgejo ticket review](how-to/use-forgejo-ticket-review.md)
- [Adopt the dev to main gate in an existing project](how-to/adopt-dev-main-gate-in-existing-project.md)
- [Use the dev to main gate with Woodpecker](how-to/use-dev-main-gate-with-woodpecker.md)
- [QA agent work scenario](how-to/qa-agent-work-scenario.md)
- [QA tictactoe agent work scenario](how-to/qa-tictactoe-agent-work-scenario.md)

## Reference

Exact contracts and command behavior:

### Surfaces

- [Commands](reference/commands.md)
- [TUI](reference/tui.md)
- [CLI TUI action matrix](reference/cli-tui-action-matrix.md)
- [Web UI action matrix](reference/web-ui-action-matrix.md)

### Project And Runtime

- [Project workspace](reference/project-workspace.md)
- [Work readiness](reference/work-readiness.md)
- [Agent work observability](reference/agent-work-observability.md)
- [Runtime configuration](reference/runtime-config.md)
- [Evidence strategy configuration](reference/evidence-strategy-config.md)
- [User environment](reference/user-environment.md)
- [Artifact generation backends](reference/artifact-generation-backends.md)

### Generated Artifacts

- [Acceptance feature generation](reference/acceptance-feature-generation.md)
- [Authoring sessions](reference/authoring-sessions.md)

### Integrations

- [Notifications](reference/notifications.md)
- [Notification configuration](reference/notification-config.md)
- [Remote commands](reference/remote-commands.md)

### Projects

- [Project CLI](reference/project-cli.md)
- [Project registry](reference/project-registry.md)
- [Dev to main gate](reference/dev-main-gate.md)

### Tickets

- [Ticket parser](reference/ticket-parser.md)
- [Ticket validation](reference/ticket-validation.md)
- [Ticket execution](reference/ticket-execution.md)
- [Workflow execution](reference/workflow-execution.md)
- [Workflow controls](reference/workflow-controls.md)
- [Ticket loop](reference/ticket-loop.md)
- [Ticket result application](reference/ticket-result-application.md)
- [Ticket status](reference/ticket-status.md)
- [Ticket comments](reference/ticket-comments.md)
- [Ticket file discovery](reference/ticket-file-discovery.md)
- [Ticket reconciliation](reference/ticket-reconciliation.md)

### Review And Merge

- [Review workflow configuration](reference/review-workflow-config.md)
- [Ticket review workflow](reference/ticket-review-workflow.md)
- [Ticket merge workflow](reference/ticket-merge-workflow.md)
- [Ticket review metadata](reference/ticket-review-metadata.md)
- [Review handoff summaries](reference/review-handoff-summaries.md)

### Reports

- [Reports](reference/reports.md)
- [Execution reports](reference/execution-reports.md)
- [Loop reports](reference/loop-reports.md)

### Maintenance And Indexes

- [Maintenance](reference/maintenance.md)
- [Index rebuild CLI](reference/index-rebuild-cli.md)
- [Index status CLI](reference/index-status-cli.md)
- [Ticket index](reference/ticket-index.md)
- [Ticket index rebuild](reference/ticket-index-rebuild.md)
- [Ticket index refresh](reference/ticket-index-refresh.md)
- [Ticket index status](reference/ticket-index-status.md)
- [SQLite migrations](reference/sqlite-migrations.md)

### Coordination And Proving

- [Daemon coordination](reference/daemon-coordination.md)
- [MVP agent proving ground](reference/mvp-agent-proving-ground.md)

## Explanation

Rationale and system shape:

- [Architecture](explanation/architecture.md)
- [Agent sandbox and runtime](explanation/agent-sandbox-and-runtime.md)
- [Document generation model](explanation/document-generation-model.md)
- [Agent review model](explanation/agent-review-model.md)
- [Testing and engineering doctrine](explanation/testing-and-engineering-doctrine.md)
- [Dev to main gate model](explanation/dev-main-gate-model.md)
- [Dev to main gate direction](explanation/dev-main-gate-direction.md)
- [Dev to main tooling gate spec draft](explanation/dev-main-tooling-gate-spec-draft.md)
- [Slicing workflow](explanation/slicing-workflow.md)
- [Ticket resolution lifecycle](explanation/ticket-resolution-lifecycle.md)
- [TUI interactive design notes](explanation/tui-interactive-design-notes.md)

## Gallery

Visual examples:

- [Gallery](explanation/gallery.md)
- [Old gallery](explanation/old-gallery.md)
