#!/usr/bin/env sh
set -eu

case_name="${1:-}"
ref_id="${2:-001}"

if [ -z "$case_name" ]; then
	echo "usage: scripts/qa/reset.sh <case> [ref]" >&2
	exit 2
fi

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
ref_dir="$repo_root/qa/ref/$ref_id/$case_name"
project_root="$repo_root/qa/project"
tmp_root="$repo_root/qa/tmp"
active_dir="$project_root/$case_name"
tmp_dir="$tmp_root/$case_name"
origin_dir="$tmp_root/$case_name-origin.git"

if [ ! -d "$ref_dir" ]; then
	echo "qa reference not found: qa/ref/$ref_id/$case_name" >&2
	exit 1
fi

rm -rf "$tmp_dir" "$origin_dir"
mkdir -p "$project_root" "$tmp_root"

if [ -e "$active_dir" ]; then
	mv "$active_dir" "$tmp_dir"
fi

mkdir -p "$active_dir"
cp -a "$ref_dir/." "$active_dir/"

git -C "$active_dir" init -q
git -C "$active_dir" checkout -B main >/dev/null 2>&1
git -C "$active_dir" config user.name "Chryso QA"
git -C "$active_dir" config user.email "qa@example.invalid"
git -C "$active_dir" add .
git -C "$active_dir" commit -q -m "chore: seed qa scenario"
git init --bare -q "$origin_dir"
git -C "$active_dir" remote add origin "$origin_dir"
git -C "$active_dir" checkout -B dev >/dev/null 2>&1
git -C "$active_dir" push -q -u origin dev

echo "qa project reset: qa/project/$case_name"
