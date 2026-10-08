package plan9asm

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
	"golang.org/x/arch/x86/x86asm"
)

// This first instruction consumer domain is scalar integer MOV and CMP reads.
// Other families, pointer/address/write forms and SIMD transport do not acquire
// a read lease merely because they reference the same numerical variable.
func x86NumericalSBReadShape(ins Instr, arch string) (int, int64, bool) {
	if len(ins.Args) != 2 {
		return 0, 0, false
	}
	widths := map[Op]int64{
		"MOVB": 1, "MOVW": 2, "MOVL": 4, "MOVQ": 8,
		"MOVBWSX": 1, "MOVBWZX": 1, "MOVBLSX": 1, "MOVBLZX": 1,
		"MOVBQSX": 1, "MOVBQZX": 1, "MOVWLSX": 2, "MOVWLZX": 2,
		"MOVWQSX": 2, "MOVWQZX": 2, "MOVLQSX": 4, "MOVLQZX": 4,
		"CMPB": 1, "CMPW": 2, "CMPL": 4, "CMPQ": 8,
	}
	width, known := widths[ins.Op]
	if !known {
		return 0, 0, false
	}
	source := 0
	if strings.HasPrefix(string(ins.Op), "MOV") {
		_, gp := x86CallerGP(ins.Args[1].Reg)
		if ins.Args[1].Kind != OpReg || !gp {
			return 0, 0, false
		}
	} else {
		if ins.Args[1].Kind == OpSym && ins.Args[0].Kind == OpReg {
			source = 1
		} else if ins.Args[1].Kind != OpReg && ins.Args[1].Kind != OpImm {
			return 0, 0, false
		}
	}
	if ins.Args[source].Kind != OpSym || strings.HasPrefix(ins.Args[source].Sym, "$") {
		return 0, 0, false
	}
	return source, width, true
}

type numericalSBRead struct {
	function          string
	instruction       int
	operand           int
	original          Instr
	pc, offset, width int64
	code              []byte
	relocation        gostaticstorage.TextRelocation
	variable          gostaticstorage.Variable
	owner             *GoNumericalStorageOwner
	ownerOrdinal      int
}

func numericalNativeName(name, pkg string) string {
	name = strings.ReplaceAll(strings.ReplaceAll(name, "∕", "/"), "·", ".")
	if strings.HasPrefix(name, ".") {
		name = pkg + name
	}
	return name
}

func numericalSameInstruction(a, b Instr, pkg string) bool {
	if a.Op != b.Op || len(a.Args) != len(b.Args) {
		return false
	}
	for i, operand := range a.Args {
		other := b.Args[i]
		if operand.Kind == OpReg && other.Kind == OpReg {
			width := int64(4)
			name := string(a.Op)
			if strings.HasSuffix(name, "SX") || strings.HasSuffix(name, "ZX") {
				name = name[:len(name)-2]
			}
			if strings.HasSuffix(name, "B") {
				width = 1
			} else if strings.HasSuffix(name, "W") {
				width = 2
			} else if strings.HasSuffix(name, "Q") {
				width = 8
			}
			root, valid := amd64FullRegBase(operand.Reg)
			otherRoot, otherValid := amd64FullRegBase(other.Reg)
			if !valid || !otherValid || root != otherRoot || x86CallerByteMask(operand.Reg, width) != x86CallerByteMask(other.Reg, width) {
				return false
			}
			// Actual cmd/asm byte-register rewriting may print AL for source
			// AX. Compare the exact byte lane, never just the wider GP alias.
			operand.Reg, other.Reg = root, root
		}
		if operand.Kind == OpSym {
			base, offset, ok := parseSBRef(operand.Sym)
			otherBase, otherOffset, otherOK := parseSBRef(other.Sym)
			if !ok || !otherOK {
				return false
			}
			operand.Sym = fmt.Sprintf("%s%+d(SB)", numericalNativeName(base, pkg), offset)
			other.Sym = fmt.Sprintf("%s%+d(SB)", numericalNativeName(otherBase, pkg), otherOffset)
		}
		if !reflect.DeepEqual(operand, other) {
			return false
		}
	}
	return true
}

// Both listings are actual assembler output. The expanded listing's physical
// line associates a unique original expanded instruction with a native PC;
// original listing and binary PC tables retain its original CPP source role.
// No line-only mapping or relocation name can independently create a read.
func numericalListingRows(raw []byte) (map[string]map[int64][]struct {
	position string
	ins      Instr
}, error) {
	rows := make(map[string]map[int64][]struct {
		position string
		ins      Instr
	})
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(nil, 4<<20)
	name := ""
	for scanner.Scan() {
		line := scanner.Text()
		if header := goAMD64FrameHeader.FindStringSubmatch(line); header != nil {
			name = header[1]
			if rows[name] != nil {
				return nil, numericalSBContext("duplicate listing TEXT")
			}
			rows[name] = make(map[int64][]struct {
				position string
				ins      Instr
			})
			continue
		}
		match := goAMD64FrameRow.FindStringSubmatch(line)
		if match == nil {
			continue // Complete binary records, not this display parser, validate objects.
		}
		pc, err := strconv.ParseInt(match[1], 16, 64)
		if err != nil || rows[name] == nil {
			return nil, numericalSBContext("unbound native listing PC")
		}
		parsed, err := Parse(ArchAMD64, "TEXT ·row(SB),4,$0\n"+match[3]+"\n")
		if err != nil || len(parsed.Funcs) != 1 || len(parsed.Funcs[0].Instrs) != 2 || parsed.Funcs[0].Instrs[1].Op == OpTEXT {
			continue // TEXT/FUNCDATA/PCDATA are not a machine read.
		}
		rows[name][pc] = append(rows[name][pc], struct {
			position string
			ins      Instr
		}{match[2], parsed.Funcs[0].Instrs[1]})
	}
	return rows, scanner.Err()
}

func numericalRelocationEqual(a, b []gostaticstorage.TextRelocation) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		x, y := a[i], b[i]
		if x.Offset != y.Offset || x.Size != y.Size || x.Type != y.Type || x.Add != y.Add ||
			x.Target.Name != y.Target.Name || x.Target.ABI != y.Target.ABI || x.Target.ABIKnown != y.Target.ABIKnown || x.Target.Package != y.Target.Package {
			return false
		}
	}
	return true
}

func numericalInstructionRelocation(text gostaticstorage.TextRecord, pc int64, native x86asm.Inst, arch string) (gostaticstorage.TextRelocation, int64, error) {
	var selected []gostaticstorage.TextRelocation
	for _, relocation := range text.Relocations {
		if relocation.Offset < pc+int64(native.Len) && relocation.Offset+relocation.Size > pc {
			selected = append(selected, relocation)
		}
	}
	if len(selected) != 1 || selected[0].Size != 4 || !selected[0].Target.ABIKnown || selected[0].Target.ABI != 0 {
		return gostaticstorage.TextRelocation{}, 0, numericalSBContext("read needs one exact numerical storage relocation/ABI")
	}
	r := selected[0]
	field := native.PCRelOff
	offset := r.Add
	if arch == "amd64" {
		if native.PCRel != 4 || r.Type != 14 || native.AddrSize != 64 {
			return r, 0, numericalSBContext("unknown AMD64 PC-relative read rule")
		}
		tail := int64(native.Len - native.PCRelOff - native.PCRel)
		if tail < 0 || offset > math.MaxInt64-tail {
			return r, 0, numericalSBContext("PC-relative read addend overflow")
		}
		offset += tail
	} else {
		if native.PCRel != 0 || r.Type != 1 || native.AddrSize != 32 {
			return r, 0, numericalSBContext("unknown 386 absolute read rule")
		}
		// These complete scalar forms use mod=00/rm=101's absolute disp32.
		// Locate that physical field from the opcode/ModRM, not from Add or
		// the listing. No SIB/index or address-size-prefix permission is given.
		code := text.Code[pc : pc+int64(native.Len)]
		prefix := 0
		for prefix < len(code) && code[prefix] == 0x66 {
			prefix++
		}
		if prefix >= len(code) {
			return r, 0, numericalSBContext("truncated absolute read opcode")
		}
		modrm := prefix + 1
		if code[prefix] == 0x0f {
			modrm++
		}
		if modrm >= len(code) || code[modrm]&0xc7 != 5 {
			return r, 0, numericalSBContext("unbound absolute scalar ModRM")
		}
		field = modrm + 1
	}
	if r.Offset != pc+int64(field) || field < 0 || field+4 > native.Len {
		return r, 0, numericalSBContext("relocation does not bind the decoded address field")
	}
	return r, offset, nil
}

func numericalOwnerContains(variable gostaticstorage.Variable, offset, width int64) bool {
	if offset < 0 || width <= 0 || offset > variable.Size || width > variable.Size-offset {
		return false
	}
	for byteOffset := offset; byteOffset < offset+width; byteOffset++ {
		found := false
		for _, span := range variable.Numerical {
			if span.Offset <= byteOffset && byteOffset-span.Offset < span.Length {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
