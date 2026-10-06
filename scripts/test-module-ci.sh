#!/usr/bin/env bash
set -euo pipefail

# One actual main-module inventory, three exhaustive scheduling owners. Nested
# CLI modules keep their independent workflow steps; Windows does not use this.
flavor=${1:?test, race or coverage required}
phase=${2:?execute or complete required}
flags=()
case "$flavor" in
  test) timeout=45m ;;
  race) timeout=60m; flags=(-race) ;;
  coverage) timeout=45m; flags=(-coverprofile=coverage-root.out) ;;
  *) echo "unknown module suite role" >&2; exit 1 ;;
esac
if [[ -n "${GOFLAGS:-}" ]]; then
  echo "module suite requires explicit empty GOFLAGS" >&2
  exit 1
fi
export GOFLAGS= GOENV=off GOWORK=off GOTOOLCHAIN=local GOMAXPROCS=2
mkdir -p _out
args=(-module -out="_out/module-$flavor" -shards=32 -workers=2 "-timeout=$timeout")
case "$phase" in
  execute)
    if [[ $# != 3 || ! $3 =~ ^[012]$ ]]; then
      echo "execute requires exactly one owner 0, 1 or 2" >&2
      exit 1
    fi
    args+=("-module-owner=$3")
    ;;
  complete)
    if [[ $# != 2 ]]; then
      echo "complete does not execute an owner" >&2
      exit 1
    fi
    args=(-module -out="_out/module-$flavor-verified" -shards=32 -workers=2
      "-timeout=$timeout" "-verify-module=_out/module-$flavor-owners")
    ;;
  *) echo "unknown module suite phase" >&2; exit 1 ;;
esac
go run ./cmd/plan9asmtest "${args[@]}" -- "${flags[@]}"
