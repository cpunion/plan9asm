package plan9asm

import (
	"fmt"
	"regexp"
)

// Keep this reader independent of the LLVM-backed fixture so its text boundary
// and strict mutation controls can also run as an explicit Go file pair.
func x86StringMemoryIRAccessError(body string, width, reads, writes int) error {
	typ := fmt.Sprintf("i%d", 8*width)
	read := regexp.MustCompile(`(?m)^  %[^\n]+ = load volatile ` + typ + `, ptr [^\n]+, align 1(?:,|$)`)
	write := regexp.MustCompile(`(?m)^  store volatile ` + typ + ` [^\n]+, ptr [^\n]+, align 1(?:,|$)`)
	if gotReads, gotWrites := len(read.FindAllString(body, -1)), len(write.FindAllString(body, -1)); gotReads != reads || gotWrites != writes {
		return fmt.Errorf("lost or invented observable native memory: got read/write %d/%d, want %d/%d\n%s",
			gotReads, gotWrites, reads, writes, body)
	}
	if regexp.MustCompile(`(?m)^  (?:%[^ ]+ = load|store) i\d+ [^\n]+, ptr [^\n]+, align 1(?:,|$)`).MatchString(body) {
		return fmt.Errorf("string access became nonvolatile\n%s", body)
	}
	return nil
}
