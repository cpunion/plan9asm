package plan9asm

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/arch/x86/x86asm"
)

func x86PackedReductionOriginalRawRows(t *testing.T, rows []x86PackedReductionOracleRow) []string {
	t.Helper()
	var assembly strings.Builder
	assembly.WriteString("TEXT raw(SB),4,$0-0\n")
	for _, row := range rows {
		assembly.WriteString(row.row + "\n")
	}
	code := x86PackedShiftActualGoBytes(t, "amd64", assembly.String())
	var raw []string
	offset := 0
	for _, row := range rows {
		_, length, known, err := decodedX86VEXHorizontalFloatInstruction(code[offset:], 64)
		if !known {
			instruction, decodeErr := x86asm.Decode(code[offset:], 64)
			length, err = instruction.Len, decodeErr
		}
		if err != nil || length < 1 || length > len(code)-offset {
			t.Fatal("original reduction whole-instruction boundary", row.row, offset, err)
		}
		raw = append(raw, strings.TrimSuffix(x86RawDisplacementDirectives(code[offset:offset+length]), "\n"))
		offset += length
	}
	if offset != len(code) {
		// Keep the failed source's complete actual byte stream. A finite
		// independent listing adds only RET and exposes Go's own PCs without
		// trusting our instruction lengths or discarding unexplained bytes.
		body := strings.TrimPrefix(assembly.String(), "TEXT raw(SB),4,$0-0\n")
		listing, err := x86ExtensionGoListing(t, "amd64", body)
		t.Fatalf("original reduction bytes not completely consumed: decoded=%d actual=%d suffix=%x complete=%x\nindependent finite Go PC listing (extra terminal RET), error=%v:\n%s", offset, len(code), code[offset:], code, err, listing)
	}
	return raw
}

func x86PackedReductionOracleState(width int) x86ResultState {
	state := x86ResultState{flags: x86ByteFlagsAll, direction: x86ResultDirectionBackward}
	for index := 0; index < 3; index++ {
		state.writeVector(index, 64, uint64(0xa5a55a5affff0000)|x86CallerVectorByteMask(int64(width)), true)
	}
	state.gp[3], state.pointerGP[3] = 0xff, true
	return state
}

func TestX86PackedReductionEverySelectorMeaningfulByteAndUpperAlias(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		for _, row := range x86PackedReductionOracleRows() {
			if strings.Contains(row.row, "(BX)") || arch == "386" && row.dot && !row.legacy {
				continue
			}
			t.Run(arch+"/"+row.row, func(t *testing.T) {
				selectors := 1
				if row.dot {
					selectors = 256
				}
				for selector := 0; selector < selectors; selector++ {
					spelling := strings.Replace(row.row, "$255", fmt.Sprintf("$%d", selector), 1)
					ins := x86ScalarLaneInstruction(t, spelling)
					initial := x86PackedReductionOracleState(row.width)
					destination, _ := amd64VectorRegisterIndex(ins.Args[row.destination].Reg, row.width)
					want := initial.clone()
					want.writeVector(destination, row.width, x86CallerVectorByteMask(int64(row.width)), row.legacy)
					state := initial.clone()
					if !x86ResultTransfer(ins, &state, nil, 0, arch) || !reflect.DeepEqual(state, want) {
						t.Fatal("complete initialized reduction changed unrelated input/FLAGS/DF/upper bytes", spelling)
					}
					for _, input := range []int{row.first, row.second} {
						source, _ := amd64VectorRegisterIndex(ins.Args[input].Reg, row.width)
						for missing := 0; missing < row.width; missing++ {
							required := true
							if row.dot {
								laneBytes := 4
								if strings.HasSuffix(row.mnemonic, "pd") {
									laneBytes = 8
								}
								required = selector&(1<<uint(4+(missing%16)/laneBytes)) != 0
							}
							state := initial.clone()
							state.writeVector(source, 64, state.vectorBytes(source)&^(uint64(1)<<uint(missing)), true)
							before := state.clone()
							accepted := x86ResultTransfer(ins, &state, nil, 0, arch)
							if accepted == required || !accepted && !reflect.DeepEqual(state, before) {
								t.Fatal("selector borrowed a missing meaningful byte or rejected a suppressed lane", spelling, input, missing, required)
							}
						}
					}
				}
			})
		}
	}
}

func TestX86PackedReductionRawIdentityAndStandaloneBoundaries(t *testing.T) {
	rows := x86PackedReductionOracleRows()
	raw := x86PackedReductionOriginalRawRows(t, rows)
	for _, arch := range []string{"386", "amd64"} {
		var source strings.Builder
		for index := range rows {
			fmt.Fprintf(&source, "TEXT F%d(SB),4,$0-0\n%s\nRET\n", index, raw[index])
		}
		original := mustParseX86RawReturn(t, source.String())
		file, err := normalizeX86RawFile(original, arch)
		if err != nil || len(file.Funcs) != 34 {
			t.Fatal("complete original reduction decoder failed", arch, err)
		}
		for index, row := range rows {
			t.Run(arch+"/"+row.row, func(t *testing.T) {
				function := file.Funcs[index]
				ins := function.Instrs[1]
				if !x86RawPackedReductionSource(ins, arch) {
					t.Fatal("missing complete physical source identity", row.row)
				}
				state := x86PackedReductionOracleState(row.width)
				before := state.clone()
				if x86ResultTransfer(ins, &state, nil, 0, arch) || !reflect.DeepEqual(state, before) {
					t.Fatal("standalone query acquired raw source/native memory authority")
				}
				if accepted, known := x86LoweredRawPackedReductionStep(ins, &state, nil, arch); !accepted || !known {
					t.Fatal("exact complete raw inputs did not compose with paired native opcode")
				}
				for _, mutate := range []func(*Instr){
					func(ins *Instr) { ins.Op = "UNKNOWN" },
					func(ins *Instr) {
						ins.Args = append([]Operand(nil), ins.Args...)
						ins.Args[0] = Operand{Kind: OpReg, Reg: "X7"}
					},
					func(ins *Instr) { ins.x86Encoded = !ins.x86Encoded },
					func(ins *Instr) { ins.x86AddressBits = 32 },
					func(ins *Instr) { ins.x86VectorBytes++ },
					func(ins *Instr) { ins.x86SegmentPrefix = 0x64 },
					func(ins *Instr) { ins.x86UnboundAddressOverride = true },
					func(ins *Instr) { ins.x86SourceRaw = nil },
					func(ins *Instr) {
						origin := *ins.x86SourceRaw
						origin.packedReduction = x86RawPackedReductionIdentity{}
						ins.x86SourceRaw = &origin
					},
				} {
					changed := ins
					mutate(&changed)
					if x86RawPackedReductionSource(changed, arch) {
						t.Fatal("edited selector/physical/source identity borrowed a raw continuation")
					}
					state := before.clone()
					if accepted, _ := x86LoweredRawPackedReductionStep(changed, &state, nil, arch); accepted || !reflect.DeepEqual(state, before) {
						t.Fatal("edited raw reduction changed a rejected state")
					}
				}
				other := "386"
				if arch == "386" {
					other = "amd64"
				}
				if x86RawPackedReductionSource(ins, other) {
					t.Fatal("raw physical source borrowed another creation mode")
				}
				cloned := cloneX86ReplayInstrs(function.Instrs)
				if cloned[1].x86SourceRaw == ins.x86SourceRaw || !x86RawPackedReductionSource(cloned[1], arch) {
					t.Fatal("immutable raw replay clone lost its physical identity snapshot")
				}
			})
		}
	}
}
