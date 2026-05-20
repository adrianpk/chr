#!/usr/bin/env sh
set -eu

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/../../.." && pwd)"
source_dir="$repo_root/.prj/fixtures/mvp-agent-proving-ground/todolist"
target_dir="$repo_root/_workspace/testing/projects/todolist"

rm -rf "$target_dir"
mkdir -p "$(dirname "$target_dir")"
cp -R "$source_dir" "$target_dir"

printf 'materialized: %s\n' "$target_dir"
