# Codex Worker Context

## Ticket

- execution_id: exec-648ea4e465b090f7
- ticket_key: CHR-1
- ticket_path: /workspace/.chr/default/tickets/cli-gameplay/CHR-1-Game-outcome-rules.md

## Artifacts

- result_path: /workspace/.chr/default/runtime/runs/CHR-1/exec-648ea4e465b090f7/result.json
- log_path: /workspace/.chr/default/runtime/runs/CHR-1/exec-648ea4e465b090f7/runtime.log
- validation_path: /workspace/.chr/default/runtime/runs/CHR-1/exec-648ea4e465b090f7/validation.log

## Result Contract

Write a JSON object to `result_path` with these fields:

- `execution_id`: must match the execution ID above.
- `ticket_key`: must match the ticket key above.
- `status`: `succeeded` or `failed`.
- `summary`: concise human-readable result summary.
- `changed_files`: optional list of workspace-relative changed file summaries.
