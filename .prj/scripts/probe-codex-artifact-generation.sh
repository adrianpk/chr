#!/usr/bin/env bash
set -euo pipefail

repo_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
source_workspace="${1:-/home/adrian/Projects/labs/chc/projects/chc-blog-work}"
workspace_root="${CHC_PROBE_WORKSPACE_ROOT:-$(mktemp -d -t chc-codex-artifacts-XXXXXX)}"
workspace="${workspace_root}/chc-blog-work"
case_slug="artifact-backend-audit"

if [[ ! -d "${source_workspace}/.chc/user" ]]; then
  echo "source workspace does not look like a ChrysoCoord workspace: ${source_workspace}" >&2
  exit 1
fi

mkdir -p "${workspace_root}"
cp -a "${source_workspace}" "${workspace}"

rm -rf \
  "${workspace}/.chc/user/log/llm" \
  "${workspace}/.chc/user/tmp/agent-runs" \
  "${workspace}/.chc/user/specs/${case_slug}-spec.md" \
  "${workspace}/.chc/user/plans/${case_slug}-plan.md" \
  "${workspace}/.chc/user/plans/${case_slug}-tracking.md" \
  "${workspace}/.chc/user/tickets/${case_slug}"

mkdir -p \
  "${workspace}/.chc/user/cases" \
  "${workspace}/.chc/user/quizzes" \
  "${workspace}/.chc/user/dumps"

cat >"${workspace}/.chc/user/cases/${case_slug}.md" <<'CASE'
---
schema_version: 1
name: "Artifact Backend Audit"
description: "Verify Codex-backed dump, spec, plan, and tracking generation on a copied reference workspace."
---

Exercise the artifact generation pipeline without modifying the source
workspace. The generated artifacts should prove that Codex owns synthesis while
Go owns file writes, ticket keys, ticket paths, and traceable materialization.
CASE

cat >"${workspace}/.chc/user/quizzes/${case_slug}-quiz.md" <<'QUIZ'
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
- Generated outputs must remain under `.chc/user`.
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

export CHC_LLM_TRACE=1
export CHC_LLM_TRACE_PAYLOADS=1
export CHC_LLM_TRACE_KEEP_FILES=1
export CHC_LLM_TRACE_MAX_BYTES="${CHC_LLM_TRACE_MAX_BYTES:-2000000}"
export CHC_OPERATOR_ID="${CHC_OPERATOR_ID:-codex-artifact-probe}"
export CHC_CODEX_TIMEOUT_SECONDS="${CHC_CODEX_TIMEOUT_SECONDS:-1200}"

run_chc() {
  (cd "${repo_dir}" && go run ./cmd/chc "$@")
}

run_chc dump new \
  --dir "${workspace}" \
  --use-case "${case_slug}" \
  --title "Artifact backend audit notes" \
  --content "Audit the default Codex artifact generation path. Preserve trace evidence, generated files, and ticket materialization boundaries."

run_chc spec generate --dir "${workspace}" --use-case "${case_slug}"
run_chc plan generate --dir "${workspace}" --use-case "${case_slug}"
run_chc plan tracking create --dir "${workspace}" --use-case "${case_slug}"

trace_path="${workspace}/.chc/user/log/llm/llm.ndjson"
agent_runs_dir="${workspace}/.chc/user/tmp/agent-runs"

echo "workspace=${workspace}"
echo "trace=${trace_path}"
echo "agent_runs=${agent_runs_dir}"
echo "trace_events=$(wc -l <"${trace_path}")"
echo "agent_run_dirs=$(find "${agent_runs_dir}" -mindepth 1 -maxdepth 1 -type d | wc -l)"
echo "generated_files:"
find "${workspace}/.chc/user" \
  \( -path "*/${case_slug}*" -o -path "*/tickets/${case_slug}/*" \) \
  -type f \
  | sort
