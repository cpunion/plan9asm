package plan9asm

import (
	"errors"
	"strings"
	"testing"
)

func TestX86LockScalarRMWTypedRolesAgreeWithoutMemoryPermission(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, row := range x86LockScalarOracleForms(arch) {
			t.Run(arch+"/"+row.text, func(t *testing.T) {
				file, err := Parse(ArchAMD64, "TEXT ·f(SB),4,$0-0\n"+row.text+"\nRET\n")
				if err != nil {
					t.Fatal(err)
				}
				ins := file.Funcs[0].Instrs[0]
				form, known, err := x86LockScalarRMWProperties(ins.Op, ins, arch)
				if err != nil || !known {
					t.Fatal("complete accepted grammar row lost its typed form", err)
				}
				effect, known, err := x86TypedInstructionByteEffect(ins.Op, ins, arch)
				if err != nil || !known || len(effect.operands) != len(ins.Args) {
					t.Fatal("accepted lowering row lacks its complete ordered byte effect", err)
				}
				for index, operand := range effect.operands {
					if operand.index != index || operand.width != int64(form.spec.bits/8) || !operand.read ||
						operand.write != (index == len(ins.Args)-1) {
						t.Fatal("read/write/native width role changed", effect)
					}
				}
				stem := row.op[:len(row.op)-1]
				defined, killed, read := x86ByteFlagsAll, uint8(0), uint8(0)
				switch stem {
				case "ADC", "SBB":
					read = x86ByteFlagCF
				case "AND", "OR", "XOR":
					defined &^= x86ByteFlagAF
					killed = x86ByteFlagAF
				case "INC", "DEC":
					defined &^= x86ByteFlagCF
				case "NOT":
					defined = 0
				case "BTC", "BTR", "BTS":
					defined = x86ByteFlagCF
					killed = x86ByteFlagOF | x86ByteFlagSF | x86ByteFlagAF | x86ByteFlagPF
				}
				bitRegister := strings.HasPrefix(stem, "BT") && ins.Args[0].Kind == OpReg
				if effect.flagsRead != read || effect.flagsDefined != defined || effect.flagsKilled != killed ||
					effect.dynamicMemory != bitRegister {
					t.Fatal("defined/preserved/killed flags or full bit-string memory role changed", effect)
				}
				for _, address := range []string{"p+0(FP)", "local-16(SP)"} {
					body := strings.Replace(row.text, "1(SI)", address, 1)
					source := strings.Replace(x86NamedLockOracleSource(arch, "CLC\nLOCK\n"+body), "$0-", "$32-", 1)
					x86LockScalarOriginalInstruction(t, arch, source)
					for _, api := range []string{"text", "module", "Go"} {
						ir, err := x86OrderingHintTranslateAPI(t, api, source,
							"package ordering\nfunc locked(p *[16]byte)\n", arch, "")
						if !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
							t.Fatal("pure effect borrowed virtual FP/AUTO native backing", api, address, err)
						}
					}
				}
			})
		}
	}
}

func TestX86LockScalarRMWMalformedSourceAndDispatchRejectBeforeIR(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, body := range []string{
			"ADDL CX", "ADDL (SI),(DI)", "ADDL X0,(SI)", "ADDL CX,$1",
			"ADDL $1.0,(SI)", "ADDL.BAD CX,(SI)", "BTCL $128,(SI)",
			"BTCL $-129,(SI)", "INCL $1", "NOTL (SI),(DI)",
			"ADDL CX,(SI)(SP*1)", "ADDL CX,(SI)(DI*3)",
		} {
			t.Run(arch+"/"+body, func(t *testing.T) {
				source := x86NamedLockOracleSource(arch, "CLC\nLOCK\n"+body)
				goAssemblerOracleOutput(t, arch, source, "", false)
				for _, api := range []string{"text", "module", "Go"} {
					ir, err := x86OrderingHintTranslateAPI(t, api, source,
						"package ordering\nfunc locked(p *[16]byte)\n", arch, "")
					if err == nil || ir != "" {
						t.Fatal("malformed complete source emitted IR", api, err)
					}
				}
			})
		}
	}
	ctx, output := newAMD64CtxWithFuncForTest(t, Func{}, FuncSig{Ret: Void}, nil)
	ctx.namedAtomicLock, ctx.lockScalarNS = true, "__bound_"
	before, temporary := output.String(), ctx.tmp
	ins := Instr{Op: "SUBL", Args: []Operand{{Kind: OpReg, Reg: CX}, {Kind: OpMem, Mem: MemRef{Base: SI}}}}
	known, terminated, err := ctx.lowerLockScalarRMW("ADDL", ins)
	if !known || terminated || err == nil || before != output.String() || temporary != ctx.tmp {
		t.Fatal("contradictory dispatch mutated IR or temporary state", known, terminated, err)
	}
}

func TestX86LockScalarRMWCarryDoesNotAuthorizeFrameReads(t *testing.T) {
	for _, arch := range []string{"amd64", "386"} {
		for _, address := range []string{"0(SP)", "0(BP)", "missing+32(FP)"} {
			t.Run(arch+"/"+address, func(t *testing.T) {
				source := x86NamedLockOracleSource(arch, "CMPL "+address+",AX\nLOCK\nADCL CX,(SI)")
				x86LockScalarOriginalInstruction(t, arch, source)
				for _, api := range []string{"text", "module", "Go"} {
					ir, err := x86OrderingHintTranslateAPI(t, api, source,
						"package ordering\nfunc locked(p *[16]byte)\n", arch, "")
					if !errors.Is(err, ErrProbeNeedsContext) || ir != "" {
						t.Fatal("CF observation borrowed an unproved stack/frame read", api, err)
					}
				}
			})
		}
	}
}
