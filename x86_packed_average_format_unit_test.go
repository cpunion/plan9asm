package plan9asm

import (
	"fmt"
	"strings"
	"testing"
)

// Object-only instruction units retain every original format row, including
// MMX and unbound native operands. These are not public translations or runs.
// Actual typed entry arguments still describe logical FP byte layout only.
func x86AverageFormatObjectUnit(t *testing.T, file *File, signatures map[string]FuncSig, arch, triple string) string {
	t.Helper()
	return x86FixedBinaryFormatObjectUnit(t, file, signatures, arch, triple, 0, func(c *amd64Ctx, ins Instr) (bool, bool, error) {
		return c.lowerPackedUnsignedAverage(ins.Op, ins)
	})
}

func x86WordMultiplyFormatObjectUnit(t *testing.T, file *File, signatures map[string]FuncSig, arch, triple string, frame int64) string {
	t.Helper()
	return x86FixedBinaryFormatObjectUnit(t, file, signatures, arch, triple, frame, func(c *amd64Ctx, ins Instr) (bool, bool, error) {
		return c.lowerPackedWordMultiply(ins.Op, ins)
	})
}

func x86FixedBinaryFormatObjectUnit(t *testing.T, file *File, signatures map[string]FuncSig, arch, triple string,
	frame int64, lower func(*amd64Ctx, Instr) (bool, bool, error)) string {
	t.Helper()
	resolve := func(name string) string { return strings.TrimPrefix(name, "·") }
	var ir strings.Builder
	fmt.Fprintf(&ir, "target triple = %q\n", triple)
	emitAMD64Prelude(&ir, arch, file)
	emitExternSBGlobals(&ir, file, resolve, signatures)
	attributes := newFeatureAttrRegistry()
	emitted := 0
	for index, function := range file.Funcs {
		sig, found := signatures[resolve(function.Sym)]
		if !found || sig.Ret != Void || function.FrameSize != frame || len(function.Instrs) < 2 {
			t.Fatal("fixed binary format unit lost its complete original Void/frame role")
		}
		var arguments []string
		for ordinal, typ := range sig.Args {
			arguments = append(arguments, fmt.Sprintf("%s %%arg%d", typ, ordinal))
		}
		features := attributes.ref(inferFuncTargetFeaturesForGOARCH(file.Arch, arch, function))
		fmt.Fprintf(&ir, "define void @FixedBinaryFormat%d(%s) %s {\n", index, strings.Join(arguments, ", "), features)
		consumer := newX86Ctx(&ir, function, sig, resolve, signatures, arch, triple, false)
		if err := consumer.emitEntryAllocas(); err != nil {
			t.Fatal(err)
		}
		for ordinal, instruction := range function.Instrs {
			switch instruction.Op {
			case OpTEXT:
				if ordinal != 0 {
					t.Fatal("format unit has an additional original TEXT")
				}
			case OpRET:
				if ordinal != len(function.Instrs)-1 || len(instruction.Args) != 0 {
					t.Fatal("format unit lost its original final return")
				}
				ir.WriteString("  ret void\n")
			default:
				known, terminated, err := lower(consumer, instruction)
				if !known || terminated || err != nil {
					t.Fatal("complete original fixed binary format row not emitted", instruction.Raw, err)
				}
				emitted++
			}
		}
		ir.WriteString("}\n")
	}
	attributes.emit(&ir)
	t.Logf("all %d original fixed binary format rows in emitter-only objects; no public or runtime success claim", emitted)
	return ir.String()
}
