# Codex Worker Context

## Ticket

- execution_id: exec-7fe25b30b4955e1e
- ticket_key: CHC-5
- ticket_path: /workspace/.chc/user/tickets/cli-gameplay/CHC-5-Test-coverage-hardening.md

## Artifacts

- result_path: /workspace/.chc/user/runtime/runs/CHC-5/exec-7fe25b30b4955e1e/result.json
- log_path: /workspace/.chc/user/runtime/runs/CHC-5/exec-7fe25b30b4955e1e/runtime.log
- validation_path: /workspace/.chc/user/runtime/runs/CHC-5/exec-7fe25b30b4955e1e/validation.log

## Result Contract

Write a JSON object to `result_path` with these fields:

- `execution_id`: must match the execution ID above.
- `ticket_key`: must match the ticket key above.
- `status`: `succeeded` or `failed`.
- `summary`: concise human-readable result summary.
- `changed_files`: optional list of workspace-relative changed file summaries.
