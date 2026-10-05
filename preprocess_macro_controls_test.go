package plan9asm

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type ppMacroControlFixture struct {
	name, prefix, replacement, argument, suffix string
}

func ppMacroControlFixtures() []ppMacroControlFixture {
	const data = "DATA ·value(SB)/4,$42\n"
	return []ppMacroControlFixture{
		{"define", "", "#define VALUE P\n", "42", "DATA ·value(SB)/4,$VALUE\n"},
		{"undef", "#define LEAK 1\n", "#undef P\n", "LEAK", "#ifndef LEAK\n" + data + "#endif\n"},
		{"include", "", "#include P\n", `"choice.h"`, "DATA ·value(SB)/4,$VALUE\n"},
		{"line", "", "#line P \"logical source.s\"\n", "337", data},
		{"ifdef", "#define PRESENT 1\n", "#ifdef P\n" + data + "#else\nDATA ·value(SB)/4,$13\n#endif\n", "PRESENT", ""},
		{"ifndef", "", "#ifndef P\n" + data + "#else\nDATA ·value(SB)/4,$13\n#endif\n", "ABSENT", ""},
		{"open-parent-condition", "#define PRESENT 1\n", "#ifdef P\n", "PRESENT", data + "#else\nDATA ·value(SB)/4,$13\n#endif\n"},
		{"inactive-invocation", "", "#else\n", "", data + "#endif\n"},
	}
}

func ppMacroControlSource(fixture ppMacroControlFixture, style string) string {
	body, definition, invocation := fixture.replacement, "APPLY", "APPLY"
	if style == "function" {
		definition += "(P)"
		invocation += "(" + fixture.argument + ")"
	} else {
		body = strings.ReplaceAll(body, "P", fixture.argument)
	}
	if style == "alias" {
		invocation = "INVOKE"
	}
	var source strings.Builder
	source.WriteString(fixture.prefix)
	source.WriteString("#define " + definition + " \\\n")
	for _, line := range strings.Split(strings.TrimSuffix(body, "\n"), "\n") {
		source.WriteString(line + " \\\n")
	}
	source.WriteByte('\n')
	if style == "alias" {
		source.WriteString("#define INVOKE APPLY\n")
	}
	source.WriteString("TEXT ·Probe(SB),$0-0\nRET\n")
	if fixture.name == "inactive-invocation" {
		source.WriteString("#ifdef ABSENT\n")
	}
	source.WriteString(invocation + "\n")
	source.WriteString(fixture.suffix)
	source.WriteString("GLOBL ·value(SB),8,$4\n")
	return source.String()
}

func TestPreprocessAssemblyGeneratedControlGoOracle(t *testing.T) {
	for _, target := range []string{"linux/386", "linux/amd64", "linux/arm", "linux/arm64", "js/wasm"} {
		t.Run(target, func(t *testing.T) {
			for _, fixture := range ppMacroControlFixtures() {
				t.Run(fixture.name, func(t *testing.T) {
					for _, style := range []string{"object", "alias", "function"} {
						t.Run(style, func(t *testing.T) {
							source := ppMacroControlSource(fixture, style)
							headers := map[string]string{"choice.h": "#define VALUE 42\n"}
							listing := ppMacroControlGoObject(t, target, source, headers, "original")
							if !strings.Contains(string(listing), "2a 00 00 00") {
								t.Fatalf("actual Go did not select the expected marker: %s", listing)
							}
							t.Logf("original Go accepted %s %s/%s source_sha=%x listing_sha=%x",
								target, fixture.name, style, sha256.Sum256([]byte(source)), sha256.Sum256(listing))
							var includes []string
							expanded, err := PreprocessAssemblySource(source, AssemblyPreprocessOptions{
								FileName: "entry.s",
								ReadInclude: func(parent, name string) (string, []byte, error) {
									includes = append(includes, parent+":"+name)
									body, ok := headers[name]
									if !ok {
										return "", nil, fmt.Errorf("missing exact header %s", name)
									}
									return name, []byte(body), nil
								},
							})
							if err != nil {
								t.Fatalf("Go-accepted generated control failed preprocessing: %v", err)
							}
							if fixture.name == "include" && strings.Join(includes, ",") != "entry.s:choice.h" {
								t.Fatalf("generated include changed its physical invocation origin: %v", includes)
							}
							listing = ppMacroControlGoObject(t, target, expanded, nil, "expanded")
							if !strings.Contains(string(listing), "2a 00 00 00") {
								t.Fatalf("expanded Go marker differs from original: %s", listing)
							}
						})
					}
				})
			}
		})
	}
}

func ppMacroControlGoObject(t *testing.T, target, source string, headers map[string]string, stage string) []byte {
	t.Helper()
	directory := t.TempDir()
	for name, body := range headers {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	file := filepath.Join(directory, "entry.s")
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(target, "/")
	ctx, cancel := runtimeOracleStageContext(t)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "tool", "asm", "-S", "-p", "macrooracle", "-I", directory,
		"-o", filepath.Join(directory, stage+".o"), file)
	command.Dir = directory
	command.Env = append(os.Environ(), "GOOS="+parts[0], "GOARCH="+parts[1], "GOENV=off", "GOTOOLCHAIN=local")
	listing, err := runtimeOracleCombinedOutput(ctx, command)
	if err != nil {
		t.Fatalf("%s original-source Go object %s: %v\n%s\nsource:\n%s", stage, target, err, listing, source)
	}
	return listing
}
