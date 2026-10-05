# Exhaustive root-package test partitions

`plan9asmtest` runs one exact Go package in fresh, serial child processes. It
discovers the complete current `Test`, runnable `Example` and `Fuzz` inventory
with actual `go test -list=.`; no checked-in test list is used. Go's separately
listed benchmarks are retained in the report but not run, matching ordinary
`go test` without `-bench`. Fuzz targets run their normal seed corpus, not a
fuzzing campaign. Every emitted subtest and seed must also terminate.

From the repository root, select the required Go/LLVM tools described in
[validation](../../docs/development/validation.md), then use a new output path:

```sh
mkdir -p _out
GOFLAGS= go run ./cmd/plan9asmtest \
  -out _out/root-partitions-current -shards 32 -timeout 45m
```

This covers **only the root package**, not root `./...`, the nested CLI
modules, strict stdlib/object gates or cross-runtime jobs. Those still require
their existing complete commands. To run every other root-module package
without repeating the root:

```sh
go list ./... > _out/non-root-packages.txt || exit 1
suite_status=0
while IFS= read -r pkg; do
  if [ "$pkg" != github.com/xgo-dev/plan9asm ]; then
    go test "$pkg" -count=1 -timeout=45m || suite_status=1
  fi
done < _out/non-root-packages.txt
test "$suite_status" -eq 0
```

Use the same race/coverage/build flags for those other packages when required.
Ordinary mode's coverage profile is root-package
coverage only, not a replacement for the original module-wide profile.

Build/instrumentation flags follow `--`, for example:

```sh
GOFLAGS= go run ./cmd/plan9asmtest \
  -out _out/root-partitions-race-cover -shards 32 -timeout 60m -- \
  -race -covermode=atomic -coverprofile=coverage-root.out -p=1
```

The tool can be built/run from a separate tool checkout while targeting a
clean frozen source checkout through `-dir /path/to/frozen-source`. Neither
checkout is edited. An absolute `-out` separates tool evidence from source.

The actual list and every execution use identical build flags. Each nonempty
partition has an independently generated coverage profile. Only after all
children and JSON validations succeed are the profiles merged into a **new**
destination, checking identical block/statement inventories and modes. `set`
counters combine by OR; `count`/`atomic` combine by checked uint64 addition.
Failed runs retain their separate profiles, not a misleading merged result.

Supported flags are `-race`, `-cover`, `-covermode`, `-coverpkg`, `-tags`, `-p`,
`-gcflags`, `-asmflags`, `-ldflags`, `-mod`, `-modfile`, `-overlay`, `-trimpath`,
`-asan`, `-msan` and `-work`. Filters, `-count`, `-fuzz`, benchmarks, custom
executors and test arguments are rejected. `GOFLAGS` must be empty: pass its
supported build flags explicitly instead of allowing hidden test filters.
The runner forces `-count=1`; successful cached runs cannot replace execution.

## Proof and failure behavior

The sorted inventory hashes newline-separated names with SHA-256. The first
eight bytes of each name's SHA-256, interpreted big-endian modulo the shard
count, determine its partition. Sorting/sharding never deletes a name. Empty
partitions are explicitly recorded; an entirely empty inventory fails.
The exact escaped alternation selects every planned root in its partition.
A pattern over 16,000 bytes fails and asks for more shards; it is never cut to
fit command-line limits.

`report.json` retains the complete plan, commands, per-root/subtest/seed
outcomes, elapsed times, incomplete counts and errors. Raw actual inventory,
Go version, JSON and stderr logs remain alongside it. Each root must have one
actual run and one terminal pass/fail/skip; unknown, missing or repeated roots,
unterminated seeds, malformed JSON, failed package/child exits and timeouts
fail. Other partitions are still attempted after a failure. Existing Go skips
are counted separately from assertions passed; they do not prove execution of
the skipped semantics. Required cross/native counterpart jobs remain required.

The wall deadline also covers compilation; Go receives the same test timeout.
Combined stdout/stderr capture is bounded at 64 MiB per command. On Unix every
child owns a process group, terminated on cancellation **and ordinary exit**,
with a bounded output-pipe wait. Interrupt, TERM and HUP cancel owned groups.
This releases LLVM contexts between shards on Linux and Darwin.
The initial tool intentionally fails on non-Unix systems until equivalent
descendant-cleanup proof exists; it does not pretend Windows execution passed.

Freeze source, environment and tools throughout a real suite run. Reports
describe the invoked Go tests, not assembly-ledger or external-module coverage.
The runner never deletes caller caches or modifies baseline/scan inputs.

## Selected cross-runtime CI owners

The required cross-runtime job uses a separate
`actual_go_selected_ci_partitions_v1` protocol. Its original complete five-prefix
selector is applied to the complete actual `-list=.` inventory and checked
against a second actual Go `-list=<selector>` observation with identical build
flags. Nonselected roots remain in the full inventory; benchmarks remain listed,
not executed. Ordinary schema-1 reports and full-root behavior are unchanged.

Sixteen jobs own the selected roots exhaustively and disjointly. The exact
`TestCrossLinuxRuntimeMatrix` root alone belongs to owner zero; the first eight
SHA-256 bytes of every other selected name, big-endian modulo fifteen plus one,
assign owners one through fifteen. There is no concurrency cap or fail-fast
matrix cancellation. Each owned root runs in a fresh bounded child with the
existing **20-minute** deadline and `-count=1`. Failures stay sticky while later
roots are attempted. Empty owners are explicitly receipted, not credited as
executing semantics. This mechanism does not prove the giant root itself fits
its unchanged deadline.

For example, with the required SDK/LLVM/cross tools already installed:

```sh
GOFLAGS= GOENV=off GOWORK=off GOTOOLCHAIN=local PLAN9ASM_CROSS_EXEC=1 \
  go run ./cmd/plan9asmtest -out _out/cross-runtime-owner \
  -ci-selector '^(TestCrossLinuxRuntimeMatrix|TestARM64Conformance|TestARM64RawSQSHLURuntimeSemantics|TestARMIntegerMemoryRuntimeCross|TestARMRawSaturationRuntimeCross)' \
  -ci-isolate-root TestCrossLinuxRuntimeMatrix -ci-owner 0 -shards 16 -timeout 20m
```

All sixteen artifacts are required by `-verify-ci-partitions <directory>`, with
the identical selector/isolate/shard/timeout/build configuration. It independently
recaptures both actual inventories and source/SDK/tool/environment scope,
recomputes ownership, and reads every root's bounded raw Go JSON and actual exit.
Missing, duplicated, foreign, incomplete, failed or skipped roots/subtests/seeds
fail completion. Paired plan/receipt deletion cannot shrink the denominator.
It also checks raw inventory/environment/version/tool-origin observations and
before/after guards. The required native tool names and bytes are retained;
cache and workspace spelling are not cross-job source identity. Inherited
`GOFLAGS`, user `GOENV`, workspace switching and toolchain fallback are forbidden.

The environment scope retains **every actual `go env -json` field**, including
unknown future fields and the complete compiler flags. Only explicitly named
SDK/source/storage roots and the Go-generated prefix-map temporary nonce are
normalized. That nonce must be a direct `go-build<decimal>` child of the observed
temporary root mapped to exactly `/tmp/go-build`; all other flag bytes, ordering
and quote spelling stay exact. Raw physical observations remain in artifacts
and the aggregate reconstructs this comparison independently. Lazy SDK tools
are resolved with the selected driver's actual `go tool -n`, not guessed paths.

Process controls are a separate explicit boundary: all `PLAN9ASM_*` and
`QEMU_*` variables, `SELF_BL_SINK`, LLVM/tool search, Go runtime, compiler include,
loader and sanitizer controls are retained. This does not claim complete ambient
machine/ISA/sysroot/library identity or copy unrelated CI credentials. Required
original/native preflights and runtime comparisons still determine those
semantics; hosted-runner labels are not ISA certificates.

Worker artifacts retain physical paths, failures and logs; matching source and
SDK content hashes permit separate checkout locations, not different source or
compiler versions. Test completion is not an assembly-ledger, external-library
runtime, or unexecuted architecture claim. Coverage-profile merging remains an
ordinary full-root feature rather than a selected-CI waiver.
External compiler-input flags (`-overlay`, `-modfile`, `-gcflags`, `-asmflags`,
`-ldflags`) are not registered in this initial CI protocol: its source guard
cannot silently authenticate files supplied outside the checkout. Ordinary
mode keeps its existing support for those flags.
External cache/package drivers and compound compiler-wrapper commands likewise
remain unregistered. The native scope binds the actual `CC`, `CXX`, `AR` and
`PKG_CONFIG` executable bytes as well as the required LLVM/cross/QEMU tools.
