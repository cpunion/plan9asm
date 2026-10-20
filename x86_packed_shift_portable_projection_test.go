package plan9asm

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// This test-only projection removes exactly the original functions' checked
// ISA selection attributes. Their integer operations, guarded volatile accesses,
// aliases and control flow remain unchanged. It is not native x86 execution.
func x86PackedShiftPortableIR(source, ir string) (string, error) {
	file, err := Parse(ArchAMD64, source)
	if err != nil {
		return "", err
	}
	file, err = normalizeX86RawFile(file, "amd64")
	if err != nil {
		return "", err
	}
	profiles := make(map[string]string)
	for _, function := range file.Funcs {
		name := strings.TrimPrefix(function.Sym, "·")
		if _, duplicate := profiles[name]; duplicate || len(function.Instrs) == 0 ||
			function.Instrs[0].Op != OpTEXT || len(function.X86RawText) != 0 {
			return "", fmt.Errorf("portable packed-shift source lost its unique ordinary function: %s", name)
		}
		shifts := 0
		for ordinal, instruction := range function.Instrs {
			switch instruction.Op {
			case OpTEXT:
				// Parse binds this sole header to function.Sym. It is not an
				// executable body operation or a general directive exemption.
				if ordinal != 0 || len(instruction.Args) != 0 {
					return "", fmt.Errorf("portable packed-shift TEXT is not its associated first header: %s", name)
				}
				continue
			case "MOVQ", "KMOVQ", "VMOVUPS", "RET":
				continue
			}
			form, recognized, formErr := x86PackedShiftProperties(instruction.Op, instruction, "amd64")
			if !recognized || formErr != nil || form.legacy || x86PackedShiftHasNativeCount(form) {
				return "", fmt.Errorf("portable packed-shift source contains an unobserved operation: %s", instruction.Raw)
			}
			shifts++
		}
		if shifts != 1 {
			return "", fmt.Errorf("portable packed-shift function lost its single original operation: %s/%d", name, shifts)
		}
		profiles[name] = inferFuncTargetFeaturesForGOARCH(ArchAMD64, "amd64", function)
	}
	return x86PackedShiftPortableFeatureProjection(ir, profiles)
}

func x86PackedShiftPortableFeatureProjection(ir string, profiles map[string]string) (string, error) {
	if len(profiles) == 0 || regexp.MustCompile(`\basm\b|@llvm\.(?:x86|aarch64|arm|wasm|ppc|riscv)\.|"target-cpu"`).MatchString(ir) {
		return "", fmt.Errorf("portable packed-shift IR contains native behavior or no owned functions")
	}
	groupPattern := regexp.MustCompile(`(?m)^attributes (#[0-9]+) = \{ "target-features"="([^"]+)" \}\r?\n`)
	groups := groupPattern.FindAllStringSubmatch(ir, -1)
	byProfile, byGroup := make(map[string]string), make(map[string]string)
	for _, group := range groups {
		if byProfile[group[2]] != "" || byGroup[group[1]] != "" {
			return "", fmt.Errorf("duplicate portable packed-shift feature group")
		}
		byProfile[group[2]], byGroup[group[1]] = group[1], group[2]
	}
	if len(regexp.MustCompile(`(?m)^attributes `).FindAllString(ir, -1)) != len(groups) ||
		strings.Count(ir, "target-features") != len(groups) {
		return "", fmt.Errorf("unobserved portable packed-shift attribute group")
	}
	definitionPattern := regexp.MustCompile(`(?m)^define [^\r\n]*@([A-Za-z_][A-Za-z0-9_.$]*)\([^\r\n]*\)[^\r\n]*\{`)
	definitions := definitionPattern.FindAllStringSubmatch(ir, -1)
	if len(definitions) != len(profiles) {
		return "", fmt.Errorf("portable packed-shift definition inventory changed: %d/%d", len(definitions), len(profiles))
	}
	seen, uses := make(map[string]bool), make(map[string]int)
	referencePattern := regexp.MustCompile(`#[0-9]+\b`)
	for _, definition := range definitions {
		name := definition[1]
		profile, found := profiles[name]
		if !found || seen[name] {
			return "", fmt.Errorf("unowned or duplicate portable packed-shift definition: %s", name)
		}
		seen[name] = true
		references := referencePattern.FindAllString(definition[0], -1)
		if profile == "" {
			if len(references) != 0 {
				return "", fmt.Errorf("feature-free portable packed-shift function gained an attribute: %s", name)
			}
			continue
		}
		group := byProfile[profile]
		if group == "" || len(references) != 1 || references[0] != group {
			return "", fmt.Errorf("portable packed-shift function lost exact ISA ownership: %s", name)
		}
		uses[group]++
	}
	for _, group := range groups {
		// Count exact numeric tokens, not prefixes such as #20 and #200.
		references := regexp.MustCompile(regexp.QuoteMeta(group[1])+`\b`).FindAllString(ir, -1)
		if uses[group[1]] == 0 || len(references) != uses[group[1]]+1 {
			return "", fmt.Errorf("portable packed-shift feature group has missing or foreign references: %s", group[1])
		}
	}
	// Only the checked group lines and corresponding definition-header tokens
	// change. In particular, no memory access or branch can be dropped here.
	projected := groupPattern.ReplaceAllString(ir, "")
	projected = definitionPattern.ReplaceAllStringFunc(projected, func(header string) string {
		return regexp.MustCompile(` #[0-9]+\b`).ReplaceAllString(header, "")
	})
	if strings.Contains(projected, "target-features") || referencePattern.MatchString(projected) {
		return "", fmt.Errorf("portable packed-shift IR retains an unchecked ISA reference")
	}
	return projected, nil
}

func TestX86PackedShiftPortableProjectionConsumesNormalizedProductionIR(t *testing.T) {
	const prefix = "TEXT ·Shift(SB),4,$0-32\nMOVQ data+0(FP),AX\nMOVQ counts+8(FP),BX\n" +
		"MOVQ out+16(FP),DI\nVMOVUPS (AX),Z0\nVMOVUPS (AX),Z2\nKMOVQ mask+24(FP),K1\n"
	const suffix = "VMOVUPS Z2,(DI)\nVMOVUPS Y2,64(DI)\nVMOVUPS X2,96(DI)\nRET\n"
	const declaration = "func Shift(data,counts,out *byte,mask uint64)"
	for _, raw := range []bool{false, true} {
		t.Run(fmt.Sprint("raw_", raw), func(t *testing.T) {
			operation := "VPSLLVW.Z (BX),X0,K1,X2\n"
			if raw {
				operation = x86RawDisplacementDirectives([]byte{0x62, 0xf2, 0xfd, 0x89, 0x12, 0x13})
			}
			source := prefix + operation + suffix
			file, err := Parse(ArchAMD64, source)
			if err != nil {
				t.Fatal(err)
			}
			file, err = normalizeX86RawFile(file, "amd64")
			if err != nil || len(file.Funcs) != 1 || len(file.Funcs[0].Instrs) != 12 || file.Funcs[0].Instrs[0].Op != OpTEXT {
				t.Fatal("real normalized source lost the associated parser header or complete operation", err)
			}
			ir, err := x86ScalarFPBitViewAPI(t, source, declaration, "amd64", "x86_64-unknown-linux-gnu", "Go")
			if err != nil || ir == "" {
				t.Fatal("initialized source lost its actual public production consumer", err)
			}
			projected, err := x86PackedShiftPortableIR(source, ir)
			if err != nil || projected != ir || !strings.Contains(projected, "load volatile") || !strings.Contains(projected, "store volatile") {
				t.Fatal("already portable production IR lost a body effect at its parser header", err)
			}
			for name, changed := range map[string]string{
				"unknown-body":            strings.Replace(source, operation, "CPUID\n"+operation, 1),
				"other-function":          source + "TEXT ·Other(SB),4,$0-0\nRET\n",
				"wrong-associated-symbol": strings.Replace(source, "·Shift(SB)", "·Other(SB)", 1),
			} {
				t.Run(name, func(t *testing.T) {
					if output, err := x86PackedShiftPortableIR(changed, ir); err == nil || output != "" {
						t.Fatal("parser header exemption accepted an unknown body or foreign function", err)
					}
				})
			}
		})
	}
}

func TestX86PackedShiftPortableProjectionOwnsEveryFunctionAndEffect(t *testing.T) {
	const body = "entry:\n  %v = load volatile i8, ptr %p, align 1\n  br label %done\ndone:\n  store volatile i8 %v, ptr %p, align 1\n  ret void\n}\n"
	input := "define void @A(ptr %p) #20 {\n" + body +
		"define void @B(ptr %p) {\n" + body + "define void @C(ptr %p) #200 {\n" + body +
		"attributes #20 = { \"target-features\"=\"+avx\" }\n" +
		"attributes #200 = { \"target-features\"=\"+avx2\" }\n"
	profiles := map[string]string{"A": "+avx", "B": "", "C": "+avx2"}
	want := "define void @A(ptr %p) {\n" + body + "define void @B(ptr %p) {\n" + body +
		"define void @C(ptr %p) {\n" + body
	if actual, err := x86PackedShiftPortableFeatureProjection(input, profiles); err != nil || actual != want {
		t.Fatal("portable projection changed an access, branch or function instead of exact ISA attributes", err, actual)
	}
	for name, changed := range map[string]string{
		"missing-group":                 strings.Replace(input, "attributes #20 = { \"target-features\"=\"+avx\" }\n", "", 1),
		"duplicate-group":               input + "attributes #20 = { \"target-features\"=\"+avx\" }\n",
		"wrong-function-reference":      strings.Replace(input, "@A(ptr %p) #20", "@A(ptr %p) #200", 1),
		"unexpected-function-reference": strings.Replace(input, "@B(ptr %p) {", "@B(ptr %p) #20 {", 1),
		"foreign-reference":             input + "; foreign #20\n",
		"missing-function":              strings.Replace(input, "@B(ptr %p)", "@Unknown(ptr %p)", 1),
		"extra-definition":              input + "define void @D() {\nret void\n}\n",
		"native-asm":                    strings.Replace(input, "ret void", `call void asm sideeffect "nop", ""()`, 1),
		"native-intrinsic":              input + "declare void @llvm.x86.sse.pause()\n",
		"extra-attribute":               input + "attributes #3 = { nounwind }\n",
	} {
		t.Run(name, func(t *testing.T) {
			if actual, err := x86PackedShiftPortableFeatureProjection(changed, profiles); err == nil || actual != "" {
				t.Fatal("invalid portable inventory returned partially projected IR", err)
			}
		})
	}
	withoutFeatures := "define void @B(ptr %p) {\n" + body
	if actual, err := x86PackedShiftPortableFeatureProjection(withoutFeatures, map[string]string{"B": ""}); err != nil || actual != withoutFeatures {
		t.Fatal("already portable production IR must remain byte-for-byte unchanged", err)
	}
}
