package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

func TestARM64RawPoolSVETypedEffects(t *testing.T) {
	type effect struct {
		line string
		want bool
	}
	var cases []effect
	for _, width := range []string{"b", "h", "s", "d"} {
		gp := "w"
		if width == "d" {
			gp = "x"
		}
		cases = append(cases,
			effect{fmt.Sprintf("dup z31.%s, %s30", width, gp), true},
			effect{fmt.Sprintf("dup z31.%s, %s9", width, gp), false},
			effect{fmt.Sprintf("compact z31.%s, p7, z30.%s", width, width), true},
			effect{fmt.Sprintf("cnt z31.%s, p7/m, z30.%s", width, width), true},
			effect{fmt.Sprintf("cnt z31.%s, p7/z, z30.%s", width, width), true},
		)
		for _, inputs := range []struct {
			first, second string
			want          bool
		}{
			{"xzr", "x30", true}, {"x9", "x30", false},
			{"x30", "x9", false}, {"x9", "x9", false},
		} {
			cases = append(cases, effect{
				fmt.Sprintf("whilelo p15.%s, %s, %s", width, inputs.first, inputs.second), inputs.want,
			})
		}
	}
	for _, memory := range []struct {
		suffix, width string
		shift         int
	}{
		{"b", "b", 0}, {"b", "h", 0}, {"b", "s", 0}, {"b", "d", 0},
		{"h", "h", 1}, {"h", "s", 1}, {"h", "d", 1},
		{"w", "s", 2}, {"d", "d", 3},
	} {
		for _, load := range []bool{false, true} {
			op, mode := "st1", ""
			if load {
				op, mode = "ld1", "/z"
			}
			for _, address := range []struct {
				text string
				want bool
			}{
				{"[x30]", true}, {"[sp, #-8, mul vl]", true},
				{"[x9]", false}, {"[x9, #7, mul vl]", false},
				{fmt.Sprintf("[x30, x29, lsl #%d]", memory.shift), true},
				{fmt.Sprintf("[x30, x9, lsl #%d]", memory.shift), false},
				{fmt.Sprintf("[x9, x30, lsl #%d]", memory.shift), false},
			} {
				cases = append(cases, effect{
					fmt.Sprintf("%s%s {z31.%s}, p7%s, %s", op, memory.suffix, memory.width, mode, address.text), address.want,
				})
			}
		}
	}
	// A GP source cannot be mistaken for a vector register sharing its number.
	cases = append(cases, effect{"dup z31.d, x9", false})
	var lines []string
	for _, test := range cases {
		lines = append(lines, "adr x9, #64", "ldr w1, [x9]", test.line, "mov x9, xzr", "ret")
	}
	words := assembleARM64LLVMWords(t, lines, "+sve2p2")
	for i, test := range cases {
		t.Run(test.line, func(t *testing.T) {
			var instructions []Instr
			for _, word := range words[i*5 : (i+1)*5] {
				instructions = append(instructions, Instr{Op: OpWORD, Args: []Operand{{Kind: OpImm, Imm: int64(word)}}})
			}
			if got := arm64RawAddressOnlyLoaded(instructions, 0, len(instructions)); got != test.want {
				t.Fatalf("load-only proof=%v, want %v for %s", got, test.want, strings.Join(lines[i*5:(i+1)*5], "; "))
			}
		})
	}
	for _, word := range []uint32{0, 0x25ae1ff3, 0x6ea0f16c, 0x6ebee0e4} {
		if arm64RawPoolSVEIgnoresAddress(word, 9) {
			t.Fatalf("unknown or unmodeled encoding %#08x acquired a safe effect", word)
		}
	}
}
