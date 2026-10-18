package plan9asm

import "testing"

func TestX86EnqueueConstructedAddressGrammarBeforeIR(t *testing.T) {
	for _, arch := range []string{"386", "amd64"} {
		for _, op := range []Op{"ENQCMD", "ENQCMDS"} {
			for _, mutation := range []string{"offset-overflow", "offset-underflow", "orphan-scale", "invalid-scale",
				"stack-index", "foreign-extension", "constructed-memory", "invalid-segment", "invalid-16-bit-base"} {
				t.Run(arch+"/"+string(op)+"/"+mutation, func(t *testing.T) {
					ins := x86ScalarLaneInstruction(t, string(op)+" (BX),DI")
					switch mutation {
					case "offset-overflow":
						ins.Args[0].Mem.Off = 1 << 32
					case "offset-underflow":
						ins.Args[0].Mem.Off = -1<<31 - 1
					case "orphan-scale":
						ins.Args[0].Mem.Scale = 4
					case "invalid-scale":
						ins.Args[0].Mem.Index, ins.Args[0].Mem.Scale = CX, 3
					case "stack-index":
						ins.Args[0].Mem.Index, ins.Args[0].Mem.Scale = SP, 1
					case "foreign-extension":
						ins.Args[0].Mem.IndexExt = ExtendSXTW
					case "constructed-memory":
						ins.Args[0].ImmRaw = "$unknown"
					case "invalid-segment":
						ins.Args[0].Mem.Segment = AX
					case "invalid-16-bit-base":
						ins.x86AddressBits, ins.Args[0].Mem.Base = 16, AX
					}
					ctx, output := newAMD64CtxWithFuncForTest(t, Func{Instrs: []Instr{ins}}, FuncSig{Ret: Void}, nil)
					ctx.goarch = arch
					before, temporary := output.String(), ctx.tmp
					known, terminated, err := ctx.lowerEnqueue(op, ins)
					if !known || terminated || err == nil || output.String() != before || ctx.tmp != temporary {
						t.Fatal("unencodable enqueue address emitted IR", known, terminated, err)
					}
				})
			}
		}
	}
}
