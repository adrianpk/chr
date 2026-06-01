# Dev To Main Gate

The dev-to-main gate collects promotion evidence before aligning `dev` into
`main`. It is designed for generic Go projects managed by ChrysoCoord.

## Command

```sh
chc gate dev-main --base <ref> --head <ref>
```

Required flags:

- `--base`: base Git ref for the promotion range.
- `--head`: head Git ref for the promotion range.

Optional flags:

- `--dir`: project directory override. Defaults to the current directory.

The command exits non-zero when validation fails or when configured
traceability policy requires agent evidence that is missing.

The command requires a managed project workspace. Create one with `chc init`
before running the gate in a project for the first time.

## Evidence

The gate records:

- base ref, head ref, resolved SHAs, and merge base;
- commits in the selected range;
- changed files and Git statuses;
- validation command, status, exit code, duration, stdout, and stderr;
- ChrysoCoord-managed work detected from execution reports;
- manual or external changes when listing is enabled;
- missing agent evidence when ChrysoCoord-owned work lacks result paths;
- high-risk manual changes when configured.

## Reports

Reports are written under:

```text
.chc/user/reports/gates/
```

Each run writes:

- `dev-main-<timestamp>.json`;
- `dev-main-<timestamp>.md`.

The JSON report is the machine-readable contract. The Markdown report is the
human-readable review artifact.

Top-level JSON fields:

- `version`;
- `created_at`;
- `status`;
- `base_ref`;
- `head_ref`;
- `base_sha`;
- `head_sha`;
- `merge_base_sha`;
- `commits`;
- `changed_files`;
- `validation`;
- `agent_managed_work`;
- `manual_or_external_changes`;
- `missing_agent_evidence`;
- `high_risk_manual_changes`;
- `failures`;
- `warnings`;
- `info`.

## Config

Project-local gate config lives at:

```text
.chc/user/config/gates.yml
```

Supported shape:

```yml
gates:
  dev_main:
    validation_command: make verify
    traceability:
      mode: agent_only
      fail_on_missing_agent_evidence: true
      warn_on_manual_changes: false
    manual_changes:
      list: true
      warn_on_high_risk_paths: false
    high_risk_paths:
      - internal/runtime/**
      - .woodpecker.yml
```

Defaults:

- `validation_command`: `go test ./...`
- `traceability.mode`: `agent_only`
- `traceability.fail_on_missing_agent_evidence`: `true`
- `traceability.warn_on_manual_changes`: `false`
- `manual_changes.list`: `true`
- `manual_changes.warn_on_high_risk_paths`: `false`
- `high_risk_paths`: empty

Manual or external changes are informational by default. They become warnings
only when policy enables warning behavior.

## Adoption Policy

For existing projects, start with:

```yml
gates:
  dev_main:
    validation_command: make verify
    traceability:
      mode: agent_only
      fail_on_missing_agent_evidence: false
      warn_on_manual_changes: true
    manual_changes:
      list: true
```

This records evidence and warns about manual or external changes without
blocking adoption because old work has no ChrysoCoord execution history.

After new work is consistently executed through ChrysoCoord tickets, projects
can tighten policy:

```yml
gates:
  dev_main:
    validation_command: make verify
    traceability:
      mode: agent_only
      fail_on_missing_agent_evidence: true
      warn_on_manual_changes: true
    manual_changes:
      list: true
      warn_on_high_risk_paths: true
```

## Woodpecker Setup

Generate a Woodpecker workflow:

```sh
chc gate woodpecker init
```

Dry-run output:

```sh
chc gate woodpecker init --dry-run
```

The generated workflow runs:

```sh
chc gate dev-main --dir . --base ${CHC_GATE_BASE:-origin/main} --head ${CHC_GATE_HEAD:-HEAD}
```

Woodpecker status is the first supported Forgejo feedback path. The pipeline
fails when the gate exits non-zero. Detailed evidence remains in
`.chc/user/reports/gates/`.
