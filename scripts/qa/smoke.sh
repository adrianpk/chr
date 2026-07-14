#!/usr/bin/env sh
set -eu

case_name="${1:-}"
ref_id="${2:-001}"

if [ -z "$case_name" ]; then
	echo "usage: scripts/qa/smoke.sh <case> [ref]" >&2
	exit 2
fi

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
active_dir="$repo_root/qa/project/$case_name"

"$repo_root/scripts/qa/reset.sh" "$case_name" "$ref_id" >/dev/null

if [ ! -d "$active_dir/.git" ]; then
	echo "qa smoke failed: active project has no independent git repository" >&2
	exit 1
fi

active_top="$(git -C "$active_dir" rev-parse --show-toplevel)"
expected_top="$(CDPATH= cd -- "$active_dir" && pwd)"
if [ "$active_top" != "$expected_top" ]; then
	echo "qa smoke failed: active project git root is $active_top" >&2
	exit 1
fi

if ! git -C "$repo_root" check-ignore -q "qa/project/$case_name"; then
	echo "qa smoke failed: qa/project/$case_name is not ignored by the parent repository" >&2
	exit 1
fi

output_file="$(mktemp)"
trap 'rm -f "$output_file"' EXIT HUP INT TERM

if ! (cd "$repo_root" && go run ./cmd/chr work doctor --dir "$active_dir" >"$output_file" 2>&1); then
	echo "qa smoke failed: work readiness doctor failed" >&2
	cat "$output_file" >&2
	exit 1
fi

if ! grep -q '^status: ready$' "$output_file"; then
	echo "qa smoke failed: work readiness is not ready after reset" >&2
	cat "$output_file" >&2
	exit 1
fi

if [ "$case_name" = "blog" ]; then
	if "$repo_root/scripts/qa/test.sh" "$case_name" >"$output_file" 2>&1; then
		echo "qa smoke failed: baseline test unexpectedly passed" >&2
		exit 1
	fi

	if ! grep -q 'status = 500, want 404' "$output_file"; then
		echo "qa smoke failed: baseline failure did not match the expected unknown-post bug" >&2
		cat "$output_file" >&2
		exit 1
	fi

	echo "qa smoke ok: qa/project/$case_name reset, ignored, git-initialized, and failing as expected"

	exit 0
fi

if ! "$repo_root/scripts/qa/test.sh" "$case_name" >"$output_file" 2>&1; then
	echo "qa smoke failed: baseline tests failed" >&2
	cat "$output_file" >&2
	exit 1
fi

if [ "$case_name" = "tictactoe" ]; then
	for fixture in \
		"ops/default/spec/00000-cli-gameplay-acceptance-feature/artifact.md" \
		"ops/default/report/00000-docs-no-tests-evidence-example/artifact.md" \
		"ops/default/report/00000-behavior-gherkin-evidence-example/artifact.md" \
		"ops/default/report/00000-blocked-validation-evidence-example/artifact.md"
	do
		if [ ! -f "$active_dir/$fixture" ]; then
			echo "qa smoke failed: missing tictactoe evidence fixture $fixture" >&2
			exit 1
		fi
	done

	if ! grep -q 'Validation: blocked' "$active_dir/ops/default/report/00000-blocked-validation-evidence-example/artifact.md"; then
		echo "qa smoke failed: blocked validation fixture is missing blocked evidence" >&2
		exit 1
	fi
fi

echo "qa smoke ok: qa/project/$case_name reset, ignored, git-initialized, and tests passed"
