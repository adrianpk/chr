#!/usr/bin/env sh
set -eu

case_name="${1:-}"

if [ -z "$case_name" ]; then
	echo "usage: scripts/qa/test.sh <case>" >&2
	exit 2
fi

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
active_dir="$repo_root/qa/project/$case_name"

if [ ! -d "$active_dir" ]; then
	echo "qa project not found: qa/project/$case_name" >&2
	echo "run: make qa-reset QA_CASE=$case_name" >&2
	exit 1
fi

if [ ! -f "$active_dir/go.mod" ]; then
	echo "qa project has no Go module yet: qa/project/$case_name"
	exit 0
fi

go -C "$active_dir" test ./...
