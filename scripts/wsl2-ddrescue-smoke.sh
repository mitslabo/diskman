#!/usr/bin/env bash
set -euo pipefail

WORK_DIR="${1:-$HOME/ddr-test}"
SIZE="${2:-1G}"
MAP_NAME="${3:-run1.map}"

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing command: $1" >&2
    exit 1
  fi
}

need_cmd ddrescue
need_cmd truncate
need_cmd mkdir
need_cmd rm
need_cmd cat

mkdir -p "$WORK_DIR"
cd "$WORK_DIR"

SRC="src.img"
DST="dst.img"

truncate -s "$SIZE" "$SRC"
truncate -s "$SIZE" "$DST"
rm -f "$MAP_NAME"

echo "[1/3] run ddrescue"
ddrescue -f "$SRC" "$DST" "$MAP_NAME"

echo "[2/3] mapfile"
cat "$MAP_NAME"

echo "[3/3] done"
echo "workdir: $WORK_DIR"
echo "source : $WORK_DIR/$SRC"
echo "dest   : $WORK_DIR/$DST"
echo "map    : $WORK_DIR/$MAP_NAME"
