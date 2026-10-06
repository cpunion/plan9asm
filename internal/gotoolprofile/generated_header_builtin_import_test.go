package gotoolprofile

import (
	"reflect"
	"strings"
	"testing"
)

func TestGeneratedHeaderBuiltinUnsafeImportFamilyMatchesNativeObjects(t *testing.T) {
	const original = `package header
import "unsafe"
type Inner struct { Byte byte; Word uintptr }
type Values struct {
	Prefix byte
	Pointer unsafe.Pointer
	Bytes []byte
	Text string
	Map map[byte]byte
	Channel chan byte
	Function func()
	Any interface{}
	Tail [0]byte
}
const Width = unsafe.Sizeof(Values{})
const Alignment = unsafe.Alignof(Values{})
const BytesOffset = unsafe.Offsetof(Values{}.Bytes)
const EmbeddedOffset = unsafe.Offsetof(struct { Prefix byte; Inner }{}.Word)
const Word = unsafe.Sizeof(uintptr(0))
`
	for _, test := range []struct {
		name, source string
	}{
		{name: "qualified", source: original},
		{name: "alias", source: strings.ReplaceAll(strings.Replace(original,
			`import "unsafe"`, `import u "unsafe"`, 1), "unsafe.", "u.")},
		{name: "dot", source: strings.ReplaceAll(strings.Replace(original,
			`import "unsafe"`, `import . "unsafe"`, 1), "unsafe.", "")},
		{name: "blank", source: "package header\nimport _ \"unsafe\"\ntype Values struct { Bytes []byte }\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, target := range []string{"linux/386", "linux/amd64", "linux/arm", "linux/arm64", "js/wasm"} {
				t.Run(target, func(t *testing.T) {
					actual := compileNativeHeaderRules(t, target, test.source)
					replayed, err := ReplayNativeGeneratedHeader("example.invalid/header", target, "go1.20",
						map[string]string{"header.go": test.source})
					if err != nil || !reflect.DeepEqual(actual, replayed) {
						t.Fatalf("Go-accepted builtin import properties differ from original replay: actual=%v replay=%v error=%v",
							actual, replayed, err)
					}
				})
			}
		})
	}
}
