package plan9asm

// Linux ignores RLIMIT_CORE when core_pattern names a pipe collector. Every
// standalone C oracle opts out before main so its fault children inherit the
// setting, without changing the Go test process or the machine-wide policy.
// Keep each child's checked zero rlimit as well for the other Unix platforms.
// See https://man7.org/linux/man-pages/man5/core.5.html.
const runtimeOracleProcessSetupC = `
#if defined(__linux__)
#include <sys/prctl.h>
#include <unistd.h>
__attribute__((constructor)) static void plan9asm_oracle_process_setup(void) {
    if (prctl(PR_SET_DUMPABLE, 0L, 0L, 0L, 0L) != 0) _exit(125);
}
#endif
`
