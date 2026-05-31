#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  scripts/assets/normalize-gallery-screenshots.sh IMAGE [...]

Normalizes ChrysoCoord TUI gallery screenshots in place.

Supported source geometries:
  1920x1080 -> 1850x888, crop 1850x888+34+142
  1920x1062 -> 1850x888, crop 1850x888+34+142
  1884x1005 -> 1852x887, crop 1852x887+15+86

Already-normalized geometries are skipped:
  1850x888
  1852x887
USAGE
}

if [[ $# -eq 0 ]]; then
  usage
  exit 2
fi

if ! command -v magick >/dev/null 2>&1; then
  echo "error: ImageMagick 'magick' command is required" >&2
  exit 1
fi

for image in "$@"; do
  if [[ ! -f "$image" ]]; then
    echo "error: file not found: $image" >&2
    exit 1
  fi

  geometry=$(magick identify -format '%wx%h' "$image")

  case "$geometry" in
    1920x1080|1920x1062)
      crop='1850x888+34+142'
      ;;
    1884x1005)
      crop='1852x887+15+86'
      ;;
    1850x888|1852x887)
      echo "skip $image ($geometry already normalized)"
      continue
      ;;
    *)
      echo "error: unsupported geometry for $image: $geometry" >&2
      exit 1
      ;;
  esac

  magick "$image" -crop "$crop" +repage "$image"
  normalized=$(magick identify -format '%wx%h' "$image")
  echo "normalized $image ($geometry -> $normalized)"
done
