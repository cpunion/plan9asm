//go:build go1.27

package plan9asm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// Actual SDK factory products remain the owners. These foreign-source
// negatives reuse genuine captured certificates but never manufacture a
// valid one for a different instruction, package, source or native address.
func TestX86ScalarReciprocalExistingOwnersPrecedeNativeFamilies(t *testing.T) {
	source, err := numericalSDKMapsReadCapture(t, "linux/amd64", "memhash_amd64.s")
	if err != nil {
		t.Fatal("actual complete SDK owner/source precondition", err)
	}
	ctx := source.sdk.originalContext
	minimum, err := CaptureGoX86SDKLiteralAlignment(ctx, source.literals.owner)
	if err != nil {
		t.Fatal("actual independent SDK literal alignment precondition", err)
	}
	t.Cleanup(func() { minimum.Close() })
	pkg, original, options := numericalSDKLiteralPackage(t, source.literals.source, source.request.AssemblyFile)
	bound, err := source.BindNumericalReadsWithLiteralAlignment(ctx, pkg, original, options, minimum)
	if err != nil {
		t.Fatal("complete unchanged SDK numerical/literal binding", err)
	}
	frozen, _, err := freezeNumericalReadInputs(original, bound)
	if err != nil {
		t.Fatal("complete SDK source/declaration/certificate identity", err)
	}
	var literal *x86SDKLiteralCertificate
	var numerical *x86NumericalReadCertificate
	var literalOriginal, numericalOriginal Instr
	for _, function := range frozen.Funcs {
		for _, instruction := range function.Instrs {
			if instruction.x86SDKLiteral != nil && literal == nil {
				if _, valid := certifiedSDKLiteralAccess(instruction, "amd64"); !valid {
					t.Fatal("actual original SDK literal certificate was not valid before use")
				}
				literal, literalOriginal = instruction.x86SDKLiteral, instruction
			}
			if instruction.x86NumericalRead != nil && numerical == nil {
				if _, _, valid := numericalCertifiedRead(instruction, "amd64"); !valid {
					t.Fatal("actual original mutable SDK read certificate was not valid before use")
				}
				numerical, numericalOriginal = instruction.x86NumericalRead, instruction
			}
		}
	}
	if literal == nil || numerical == nil {
		t.Fatal("actual complete SDK source lost either independent owner class")
	}
	for _, api := range []string{"text", "module", "go-module"} {
		t.Run("actual-original-SDK/"+api, func(t *testing.T) {
			if ir := numericalFamilyTranslate(t, api, pkg, original, bound); ir == "" {
				t.Fatal("actual original whole-SDK valid certificate control emitted no module")
			}
		})
	}

	for _, op := range []string{"MOVSS", "MOVSD", "RCPSS", "RSQRTSS", "VRCPSS", "VRSQRTSS"} {
		for _, memory := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/memory=%t", op, memory), func(t *testing.T) {
				text, declaration, target := x86ScalarReciprocalOwnerProgram(op, memory)
				goAssemblerOracleOutput(t, "amd64", text, "", true)
				for _, api := range []string{"text", "module", "Go"} {
					t.Run("baseline/"+api, func(t *testing.T) {
						if _, err := x86ScalarFPBitViewAPI(t, text, declaration, "amd64", "x86_64-unknown-linux-gnu", api); err != nil {
							t.Fatal("complete legitimate source baseline was relabeled or rejected", err)
						}
					})
				}
				program, err := Parse(ArchAMD64, text)
				if err != nil {
					t.Fatal(err)
				}
				programPackage := mustGoPackage(t, "ownercontrol", "package ownercontrol\n"+declaration)
				resolve := func(name string) string { return strings.TrimPrefix(name, "·") }
				signatures, err := goSigsForAsmFile(programPackage, program, resolve, "amd64", nil, nil, nil)
				if err != nil {
					t.Fatal(err)
				}
				function := program.Funcs[0]
				ordinal := -1
				for index, instruction := range function.Instrs {
					if instruction.Raw == target {
						if ordinal != -1 {
							t.Fatal("ambiguous complete original scalar consumer")
						}
						ordinal = index
					}
				}
				if ordinal == -1 {
					t.Fatal("complete original scalar instruction was not retained", target)
				}
				for _, owner := range []string{"SDKLiteral", "NumericalRead"} {
					t.Run(owner+"/original-certificate-on-foreign-source", func(t *testing.T) {
						changed := function
						changed.Instrs = append([]Instr(nil), function.Instrs...)
						instruction := &changed.Instrs[ordinal]
						if owner == "SDKLiteral" {
							instruction.x86SDKLiteral = literal
							if _, valid := certifiedSDKLiteralAccess(*instruction, "amd64"); valid {
								t.Fatal("actual original SDK certificate was forged into a foreign source")
							}
						} else {
							instruction.x86NumericalRead = numerical
							if _, _, valid := numericalCertifiedRead(*instruction, "amd64"); valid {
								t.Fatal("actual original mutable-read certificate was forged into a foreign source")
							}
						}
						consumer := newX86Ctx(new(strings.Builder), changed, signatures["Value"], resolve, signatures,
							"amd64", "x86_64-unknown-linux-gnu", false)
						if err := consumer.checkLoweredScalarFPResultInitialization(); !errors.Is(err, ErrProbeNeedsContext) {
							t.Fatal("family native continuation bypassed the existing owner's invalid complete identity", owner, err)
						}
						if !reflect.DeepEqual(program.Funcs[0], function) {
							t.Fatal("negative changed the complete original source control")
						}
					})
				}
			})
		}
	}
	if _, valid := certifiedSDKLiteralAccess(literalOriginal, "amd64"); !valid {
		t.Fatal("foreign-source negative revoked or mutated the independent original literal owner")
	}
	if _, _, valid := numericalCertifiedRead(numericalOriginal, "amd64"); !valid {
		t.Fatal("foreign-source negative revoked or mutated the independent original numerical owner")
	}
}

func x86ScalarReciprocalOwnerProgram(op string, memory bool) (source, declaration, target string) {
	load, scalar := "MOVSS", "uint32"
	if op == "MOVSD" {
		load, scalar = "MOVSD", "uint64"
	}
	input, preload := "(AX)", ""
	if !memory {
		input, preload = "X0", load+" (AX),X0\n"
	}
	args, destination := input+",X1", "X1"
	if strings.HasPrefix(op, "V") {
		args, destination = input+",X1,X2", "X2"
	}
	target = op + " " + args
	source = "TEXT ·Value(SB),4,$0-24\nMOVQ out+0(FP),DI\nMOVQ input+8(FP),AX\nMOVQ upper+16(FP),BX\n" +
		"MOVUPS (BX),X1\nSTC\n" + preload + target + "\nMOVUPS " + destination + ",(DI)\nSETCS 16(DI)\nSTMXCSR 20(DI)\nRET\n"
	declaration = fmt.Sprintf("func Value(out *[24]byte,input *%s,upper *[16]byte)\n", scalar)
	return source, declaration, target
}
