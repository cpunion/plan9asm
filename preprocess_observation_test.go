package plan9asm

import (
	"crypto/sha256"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestPreprocessAssemblyCPPObservationGoObjects(t *testing.T) {
	for _, target := range []string{"linux/386", "linux/amd64", "linux/arm", "linux/arm64", "js/wasm"} {
		t.Run(target, func(t *testing.T) {
			for _, fixture := range ppMacroControlFixtures() {
				t.Run(fixture.name, func(t *testing.T) {
					for _, style := range []string{"object", "alias", "function"} {
						t.Run(style, func(t *testing.T) {
							source := ppMacroControlSource(fixture, style)
							headers := map[string]string{"choice.h": "#define VALUE 42\n"}
							original := ppMacroControlGoObject(t, target, source, headers, "observation-original")
							if !strings.Contains(string(original), "2a 00 00 00") {
								t.Fatal("actual original Go object lacks its marker")
							}
							options := AssemblyPreprocessOptions{
								FileName: "entry.s",
								ReadInclude: func(parent, name string) (string, []byte, error) {
									body, ok := headers[name]
									if !ok || parent != "entry.s" {
										return "", nil, fmt.Errorf("unbound include %s:%s", parent, name)
									}
									return name, []byte(body), nil
								},
							}
							plain, err := PreprocessAssemblySource(source, options)
							if err != nil {
								t.Fatal(err)
							}
							observed, err := ObserveAssemblyCPP(source, options)
							if err != nil {
								t.Fatal(err)
							}
							if observed.Output != plain {
								t.Fatalf("observation changed strict CPP output: %q != %q", observed.Output, plain)
							}
							ppCheckObservation(t, observed)
							if len(observed.Events) == 0 || len(observed.Sources) == 0 {
								t.Fatal("accepted source has no actual execution/source observation")
							}
							ppCheckObservedSource(t, observed, "entry.s", source)
							if fixture.name == "include" {
								ppCheckObservedSource(t, observed, "choice.h", headers["choice.h"])
							}
							normalized := ppMacroControlGoObject(t, target, observed.Output, nil, "observation-normalized")
							if !strings.Contains(string(normalized), "2a 00 00 00") {
								t.Fatal("observed CPP changed the actual Go object's marker")
							}
						})
					}
				})
			}
		})
	}
}

func TestPreprocessAssemblyCPPObservationTransitions(t *testing.T) {
	const source = "#define F(x) \\\n#define VALUE x \\\n\n#ifdef ABSENT\nF(13)\n#define DISABLED 1\n#include \"missing.h\"\n#else\nF(42)\n#endif\nDATA ·value(SB)/4,$VALUE\nGLOBL ·value(SB),8,$4\n"
	observed, err := ObserveAssemblyCPP(source, AssemblyPreprocessOptions{FileName: "entry.s"})
	if err != nil {
		t.Fatal(err)
	}
	ppCheckObservation(t, observed)
	var inactiveInvocation, activeInvocation, generatedDefinition, skippedInclude bool
	for _, event := range observed.Events {
		if event.Kind == "invoke" && event.Name == "F" {
			inactiveInvocation = inactiveInvocation || !event.EnabledBefore
			activeInvocation = activeInvocation || event.EnabledBefore
			if !event.Function || event.Definition == 0 {
				t.Fatalf("lost actual function-macro definition: %+v", event)
			}
		}
		if event.Kind == "control" && event.Directive == "define" && event.Name == "VALUE" {
			generatedDefinition = generatedDefinition || event.Expansion != 0
		}
		if event.Kind == "control" && event.Directive == "include" && !event.EnabledBefore {
			skippedInclude = true
			if event.Included != "" {
				t.Fatal("inactive include invented actual consumption")
			}
		}
	}
	if !inactiveInvocation || !activeInvocation || !generatedDefinition || !skippedInclude {
		t.Fatalf("incomplete actual execution transitions: %+v", observed.Events)
	}
}

func TestPreprocessAssemblyCPPObservationPredefines(t *testing.T) {
	defines := []string{"VALUE=13", "VALUE=42"}
	const source = "DATA ·value(SB)/4,$VALUE\nGLOBL ·value(SB),8,$4\n"
	observed, err := ObserveAssemblyCPP(source, AssemblyPreprocessOptions{FileName: "entry.s", Defines: defines})
	if err != nil {
		t.Fatal(err)
	}
	ppCheckObservation(t, observed)
	if !reflect.DeepEqual(observed.Defines, defines) || len(observed.Events) != 3 ||
		observed.Events[0].Kind != "predefine" || observed.Events[1].Kind != "predefine" ||
		observed.Events[2].Definition != 2 {
		t.Fatalf("predefined overwrite/actual macro binding lost: %+v", observed)
	}
	defines[0] = "VALUE=99"
	if observed.Defines[0] != "VALUE=13" {
		t.Fatal("observation borrowed mutable caller options")
	}
}

func TestPreprocessAssemblyCPPObservationIncludes(t *testing.T) {
	const source = "#include \"guard.h\"\nTEXT ·Probe(SB),$0-0\nRET\n"
	const header = "#ifndef GUARD\n#define GUARD 1\n#include \"guard.h\"\n#endif\n"
	options := AssemblyPreprocessOptions{FileName: "entry.s", ReadInclude: func(parent, name string) (string, []byte, error) {
		if name != "guard.h" || parent != "entry.s" && parent != "guard.h" {
			return "", nil, fmt.Errorf("unbound include")
		}
		return "guard.h", []byte(header), nil
	}}
	observed, err := ObserveAssemblyCPP(source, options)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := PreprocessAssemblySource(source, options)
	if err != nil || plain != observed.Output || len(observed.Sources) != 2 {
		t.Fatalf("guarded include cycle changed: %+v / %v", observed, err)
	}
	ppCheckObservedSource(t, observed, "guard.h", header)
	var active int
	for _, event := range observed.Events {
		if event.Directive == "include" && event.EnabledBefore {
			active++
			if event.Included != "guard.h" {
				t.Fatal("actual include identity missing")
			}
		}
	}
	if active != 2 {
		t.Fatalf("lost actual repeated include use: %d", active)
	}
}

func TestPreprocessAssemblyCPPObservationFailureIdentity(t *testing.T) {
	for _, source := range []string{
		"#define RET(x) x\nTEXT ·Probe(SB),$0-0\nRET\n",
		"#define F(x) x\n#ifdef ABSENT\nF +\n#endif\nTEXT ·Probe(SB),$0-0\nRET\n",
		"#include \"missing.h\"\n",
		"TEXT ·Probe(SB),$0-0\nRET",
	} {
		t.Run(fmt.Sprintf("%x", sha256.Sum256([]byte(source))), func(t *testing.T) {
			options := AssemblyPreprocessOptions{FileName: "entry.s"}
			_, plainErr := PreprocessAssemblySource(source, options)
			observed, observedErr := ObserveAssemblyCPP(source, options)
			if plainErr == nil {
				// Some parser failures remain invalid normalized machine source;
				// the query must still return exactly the normal engine's output.
				plain, err := PreprocessAssemblySource(source, options)
				if err != nil || observedErr != nil || observed.Output != plain {
					t.Fatalf("observation changed successful CPP: %v / %v", err, observedErr)
				}
				return
			}
			if observedErr == nil || observedErr.Error() != plainErr.Error() ||
				!reflect.DeepEqual(observed, AssemblyCPPObservation{}) {
				t.Fatalf("observation changed failure or returned a positive trace: %v / %v / %+v", plainErr, observedErr, observed)
			}
		})
	}
}

func ppCheckObservation(t *testing.T, observation AssemblyCPPObservation) {
	t.Helper()
	for index, event := range observation.Events {
		if event.Ordinal != index+1 || event.Expansion < 0 || event.Expansion >= event.Ordinal ||
			event.Definition < 0 || event.Definition >= event.Ordinal {
			t.Fatalf("noncanonical event order/lineage: %+v", event)
		}
	}
}

func ppCheckObservedSource(t *testing.T, observation AssemblyCPPObservation, file, source string) {
	t.Helper()
	for _, observed := range observation.Sources {
		if observed.File == file {
			if observed.Bytes != len(source) || observed.SHA256 != fmt.Sprintf("%x", sha256.Sum256([]byte(source))) {
				t.Fatalf("source identity mismatch: %+v", observed)
			}
			return
		}
	}
	t.Fatalf("missing actual source identity %q", file)
}
