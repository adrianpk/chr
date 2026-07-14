#!/usr/bin/env bash
set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
source_workspace="${1:-/home/adrian/Projects/labs/chr/projects/chr-blog-work}"
workspace_root="${CHR_PROBE_WORKSPACE_ROOT:-$(mktemp -d -t chr-codex-artifacts-XXXXXX)}"
workspace="${workspace_root}/chr-blog-work"
case_slug="artifact-backend-audit"

if [[ ! -d "${source_workspace}/.chr/default" ]]; then
  echo "source workspace does not look like a Chryso workspace: ${source_workspace}" >&2
  exit 1
fi

mkdir -p "${workspace_root}"
cp -a "${source_workspace}" "${workspace}"

rm -rf \
  "${workspace}/.chr/default/log/llm" \
  "${workspace}/.chr/default/tmp/agent-runs" \
  "${workspace}/.chr/default/specs/${case_slug}-spec.md" \
  "${workspace}/.chr/default/plans/${case_slug}-plan.md" \
  "${workspace}/.chr/default/plans/${case_slug}-tracking.md" \
  "${workspace}/.chr/default/tickets/${case_slug}"

mkdir -p \
  "${workspace}/.chr/default/cases" \
  "${workspace}/.chr/default/quizzes" \
  "${workspace}/.chr/default/dumps"

cat >"${workspace}/.chr/default/cases/${case_slug}.md" <<'CASE'
---
schema_version: 1
name: "Artifact Backend Audit"
description: "Verify Codex-backed dump, spec, plan, and tracking generation on a copied reference workspace."
---

Exercise the artifact generation pipeline without modifying the source
workspace. The generated artifacts should prove that Codex owns synthesis while
Go owns file writes, ticket keys, ticket paths, and traceable materialization.
CASE

cat >"${workspace}/.chr/default/quizzes/${case_slug}-quiz.md" <<'QUIZ'
---
project_id: "project-a2274745ed8d848a"
case_slug: "artifact-backend-audit"
scope: "use_case"
status: "curated"
updated_at: "2026-05-29T12:00:00Z"
---

# Quiz Notes

## Decisions

- The audit must run against a copied reference workspace.
- Codex should be the default backend for long artifact generation.
- Generated outputs must remain under `.chr/default`.
- Go must own artifact paths, ticket keys, and ticket file creation.
- The trace should show Codex request and response events.

## Constraints

- Do not modify the source reference workspace.
- Do not use direct OpenAI backend overrides.
- Keep the generated case intentionally small enough for repeatable testing.

## Evidence

- A curated dump is created for the audit case.
- A spec is generated for the audit case.
- A plan is generated for the audit case.
- Plan tracking and tickets are materialized for the audit case.
- The workspace trace log contains Codex events.
- The workspace agent run bundle directory contains prompts, schemas, logs, and
  outputs.
QUIZ

export CHR_LLM_TRACE=1
export CHR_LLM_TRACE_PAYLOADS=1
export CHR_LLM_TRACE_KEEP_FILES=1
export CHR_LLM_TRACE_MAX_BYTES="${CHR_LLM_TRACE_MAX_BYTES:-2000000}"
export CHR_OPERATOR_ID="${CHR_OPERATOR_ID:-codex-artifact-probe}"
export CHR_CODEX_TIMEOUT_SECONDS="${CHR_CODEX_TIMEOUT_SECONDS:-1200}"

run_chr() {
  (cd "${repo_dir}" && go run ./cmd/chr "$@")
}

run_chr dump new \
  --dir "${workspace}" \
  --use-case "${case_slug}" \
  --title "Artifact backend audit notes" \
  --content "Audit the default Codex artifact generation path. Preserve trace evidence, generated files, and ticket materialization boundaries."

run_chr spec generate --dir "${workspace}" --use-case "${case_slug}"
run_chr plan generate --dir "${workspace}" --use-case "${case_slug}"
run_chr plan tracking create --dir "${workspace}" --use-case "${case_slug}"

trace_path="${workspace}/.chr/default/log/llm/llm.ndjson"
agent_runs_dir="${workspace}/.chr/default/tmp/agent-runs"

echo "workspace=${workspace}"
echo "trace=${trace_path}"
echo "agent_runs=${agent_runs_dir}"
echo "trace_events=$(wc -l <"${trace_path}")"
echo "agent_run_dirs=$(find "${agent_runs_dir}" -mindepth 1 -maxdepth 1 -type d | wc -l)"
echo "generated_files:"
find "${workspace}/.chr/default" \
  \( -path "*/${case_slug}*" -o -path "*/tickets/${case_slug}/*" \) \
  -type f \
  | sort
