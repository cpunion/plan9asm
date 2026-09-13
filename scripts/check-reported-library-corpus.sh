#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)
suite=${1:-all}
if [[ $# -gt 1 ]]; then
  echo "usage: $0 [suite-id|all]" >&2
  exit 2
fi

if [[ -n "${LLVM_CONFIG:-}" ]]; then
  llvm_bin_dir=$("$LLVM_CONFIG" --bindir)
  llc_cmd="$llvm_bin_dir/llc"
else
  llc_cmd=$(command -v llc || command -v llc-23 || command -v llc-22 || command -v llc-21 || command -v llc-20 || command -v llc-19 || true)
fi
if [[ ! -x "$llc_cmd" ]]; then
  echo "llc not found through LLVM_CONFIG or PATH" >&2
  exit 1
fi

tmp_root=$(mktemp -d)
trap 'rm -rf "$tmp_root"' EXIT
translator="$tmp_root/plan9asmll"
runner="$tmp_root/plan9asmcorpus"
go build -C "$repo_root/cmd/plan9asmll" -o "$translator" .
go build -C "$repo_root" -o "$runner" ./cmd/plan9asmcorpus

check_latest=${PLAN9ASM_CORPUS_CHECK_LATEST:-true}
"$runner" \
  -manifest="$repo_root/testdata/corpus/reported-libraries.json" \
  -corpus-dir="$repo_root/testdata/corpus" \
  -repo-root="$repo_root" \
  -translator="$translator" \
  -llc="$llc_cmd" \
  -suite="$suite" \
  -check-latest="$check_latest"
