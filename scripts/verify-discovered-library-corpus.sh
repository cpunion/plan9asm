#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/.." && pwd)
progress=0
if [[ "${1:-}" == --progress ]]; then
  progress=1
  shift
fi
reports_dir=${1:-"$repo_root/_out/discovered-library-corpus"}
shard_count=${2:-64}
if (( $# > 2 )) || ! [[ "$shard_count" =~ ^[1-9][0-9]*$ ]]; then
  echo "usage: $0 [--progress] [reports-directory] [shard-count]" >&2
  exit 2
fi

if (( progress )); then
  mkdir -p "$reports_dir"
  reports_dir=$(cd "$reports_dir" && pwd)
fi
corpus_args=(
  -manifest "$repo_root/testdata/corpus/reported-libraries.json"
  -repo-root "$repo_root"
  -discovery-ledger "$repo_root/testdata/discovery/ledger"
  -discovery-shard-count "$shard_count"
  -verify-discovery-reports "$reports_dir"
  -compare-assembly-ledger "$repo_root/testdata/discovery/assembly-ledger"
)
if (( progress )); then
  if [[ -e "$reports_dir/scan-status.json" || -L "$reports_dir/scan-status.json" ||
        -e "$reports_dir/assembly-progress.json" || -L "$reports_dir/assembly-progress.json" ]]; then
    echo "progress output already exists; preserve it and use a fresh reports directory" >&2
    exit 1
  fi
  pending_dir=$(mktemp -d "$reports_dir/.status-XXXXXX")
  trap 'rm -rf "$pending_dir"' EXIT
  go run -C "$repo_root" ./cmd/plan9asmdiscover \
    -status -out-dir "$repo_root/testdata/discovery/ledger" \
    > "$pending_dir/scan-status.json"
  mv "$pending_dir/scan-status.json" "$reports_dir/scan-status.json"

  # The CLI atomically publishes only fully audited progress. It still exits
  # nonzero for incomplete/failed reports or a committed-ledger mismatch.
  corpus_args+=(-discovery-progress-output "$pending_dir/assembly-progress.json")
  status=0
  go run -C "$repo_root" ./cmd/plan9asmcorpus "${corpus_args[@]}" || status=$?
  if [[ -f "$pending_dir/assembly-progress.json" ]]; then
    mv "$pending_dir/assembly-progress.json" "$reports_dir/assembly-progress.json"
  elif (( status == 0 )); then
    echo "strict reader returned success without its requested progress output" >&2
    exit 1
  fi
  exit "$status"
fi

go run -C "$repo_root" ./cmd/plan9asmdiscover \
  -status \
  -out-dir "$repo_root/testdata/discovery/ledger"
go run -C "$repo_root" ./cmd/plan9asmcorpus "${corpus_args[@]}"
