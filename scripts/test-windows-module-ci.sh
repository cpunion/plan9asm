#!/usr/bin/env bash
set -euo pipefail

module=${1:?root, plan9asm or plan9asmll required}
phase=${2:?execute or complete required}
case "$module" in
  root) directory=.; parts=16; shards=32 ;;
  plan9asm|plan9asmll) directory="cmd/$module"; parts=4; shards=8 ;;
  *) echo 'unknown Windows module' >&2; exit 1 ;;
esac
case "$phase" in
  execute)
    if [[ $# != 4 || ! $3 =~ ^[012]$ ]]; then
      echo 'execute requires one owner and its root part (or -1)' >&2
      exit 1
    fi
    if [[ $3 == 0 ]]; then
      if ! [[ $4 =~ ^([0-9]|[1-9][0-9])$ ]] || (( $4 >= parts )); then
        echo 'invalid Windows root partition' >&2
        exit 1
      fi
    elif [[ $4 != -1 ]]; then
      echo 'non-root owners require part -1' >&2
      exit 1
    fi
    args=("-module-owner=$3" "-module-root-partition=$4"
      "-out=_out/windows-module-$module")
    ;;
  complete)
    if [[ $# != 2 ]]; then
      echo 'complete does not execute an owner' >&2
      exit 1
    fi
    args=("-verify-module=_out/windows-module-$module-owners"
      "-out=_out/windows-module-$module-verified")
    ;;
  *) echo 'unknown Windows module phase' >&2; exit 1 ;;
esac

# The LLVM setup action selects byollvm. Both the inherited and explicit Go
# build context must match; never clear it and select different helper sources.
if [[ ${GOFLAGS:-} != -tags=byollvm ]]; then
  echo 'Windows LLVM module suite requires the declared byollvm context' >&2
  exit 1
fi
export GOENV=off GOWORK=off GOTOOLCHAIN=local GOMAXPROCS=2
if [[ $(go env GOHOSTOS) != windows ]]; then
  echo 'Windows module driver requires an actual Windows Go toolchain' >&2
  exit 1
fi
native_root=$(cygpath -m "$PWD")
mkdir -p _out
go run ./cmd/plan9asmtest -module "-dir=$directory" "${args[@]}" \
  "-shards=$shards" -workers=2 -timeout=30m "-module-root-partitions=$parts" -- \
  -tags=byollvm "-coverprofile=$native_root/coverage-windows-$module.out"
if [[ $phase == complete ]]; then
  test -s "coverage-windows-$module.out"
fi
