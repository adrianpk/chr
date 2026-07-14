# Codex Worker Context

## Ticket

- execution_id: exec-409755cd38adfe67
- ticket_key: CHR-2
- ticket_path: /workspace/.chr/user/tickets/cli-gameplay/CHR-2-Move-state-transitions.md

## Artifacts

- result_path: /workspace/.chr/user/runtime/runs/CHR-2/exec-409755cd38adfe67/result.json
- log_path: /workspace/.chr/user/runtime/runs/CHR-2/exec-409755cd38adfe67/runtime.log
- validation_path: /workspace/.chr/user/runtime/runs/CHR-2/exec-409755cd38adfe67/validation.log

## Result Contract

Write a JSON object to `result_path` with these fields:

- `execution_id`: must match the execution ID above.
- `ticket_key`: must match the ticket key above.
- `status`: `succeeded` or `failed`.
- `summary`: concise human-readable result summary.
- `changed_files`: optional list of workspace-relative changed file summaries.
