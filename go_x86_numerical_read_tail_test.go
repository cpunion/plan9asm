package plan9asm

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestGoX86NumericalDirectTailStaticListingOriginalGo(t *testing.T) {
	// The owned assembler already supplies -p oracle. Historical Go listing
	// formatting leaves a leading-middle-dot symbol's package blank despite
	// that flag; explicit source qualification retains one exact identity on
	// both tools without guessing or rewriting their native listing names.
	for _, arch := range []string{"386", "amd64"} {
		t.Run(arch, func(t *testing.T) {
			listing := x86InsertExtractGoListing(t, arch, "TEXT oracle·entry(SB),4,$0-0\nJMP worker<>(SB)\nTEXT worker<>(SB),4,$0\nCMPB feature+7(SB),$1\nRET\n")
			rows, err := numericalListingRows(listing)
			if err != nil || len(rows) != 2 || len(rows["worker"]) != 2 || len(rows["oracle.entry"]) != 1 {
				for _, line := range bytes.Split(listing, []byte("\n")) {
					if bytes.Contains(line, []byte("STEXT")) {
						t.Log(string(line))
					}
				}
				t.Fatal("static source body was assigned to another TEXT listing", err)
			}
		})
	}
}

func TestGoX86NumericalDirectTailListingUnknownHeadersFailClosed(t *testing.T) {
	const header = "example.entry STEXT nosplit asm size=1 args=0x0 locals=0x0 funcid=0x0 align=0x0\n"
	const row = "\t0x0000 00000 (source.s:4)\tRET\n"
	for _, bad := range []string{
		"worker STEXT mystery nosplit size=1 args=0x0 locals=0x0",
		"worker STEXT static asm size=1 args=0x0 locals=0x0",
		"worker STEXT nosplit size=x args=0x0 locals=0x0",
		"worker STEXT nosplit size=1 args=0x0",
		"worker STEXT STEXT nosplit size=1 args=0x0 locals=0x0",
		"STEXT static nosplit size=1 args=0x0 locals=0x0",
	} {
		t.Run(strings.Fields(bad)[0]+"/"+bad, func(t *testing.T) {
			rows, err := numericalListingRows([]byte(header + row + bad + "\n" + row))
			if rows != nil || !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatal("unknown header attributed instructions to the preceding TEXT", rows, err)
			}
		})
	}
	for _, raw := range []string{row, header + row + header + row} {
		if rows, err := numericalListingRows([]byte(raw)); rows != nil || !errors.Is(err, ErrProbeNeedsContext) {
			t.Fatal("unbound or ambiguous native instruction ownership was accepted", rows, err)
		}
	}
}
