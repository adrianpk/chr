# Codex Worker Context

## Ticket

- execution_id: exec-c4b72e7ae4e31f4d
- ticket_key: CHR-3
- ticket_path: /workspace/.chr/user/tickets/cli-gameplay/CHR-3-CLI-board-rendering.md

## Artifacts

- result_path: /workspace/.chr/user/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/result.json
- log_path: /workspace/.chr/user/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/runtime.log
- validation_path: /workspace/.chr/user/runtime/runs/CHR-3/exec-c4b72e7ae4e31f4d/validation.log

## Result Contract

Write a JSON object to `result_path` with these fields:

- `execution_id`: must match the execution ID above.
- `ticket_key`: must match the ticket key above.
- `status`: `succeeded` or `failed`.
- `summary`: concise human-readable result summary.
- `changed_files`: optional list of workspace-relative changed file summaries.
