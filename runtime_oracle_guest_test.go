package plan9asm

import "testing"

// Keep the complete execution prefix; the guest argument regression observes
// whether this historical single-program spelling preserves the real argv.
func runtimeOracleGuestArguments(t *testing.T, prefix []string, executable string, arguments ...string) []string {
	t.Helper()
	command := append([]string(nil), prefix...)
	command = append(command, executable)
	return append(command, arguments...)
}
