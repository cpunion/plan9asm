package gotoolprofile

import (
	"reflect"
	"testing"
)

const nativeHeaderGenericPropertiesSource = `package header
type Generic[T any] struct {
	Value T
	Tail [0]byte
}
type Constrained[T interface { ~uint64 | ~string }] struct { Value T }
type Pointers[T any] struct { Next *Pointers[T] }
type Concrete = Generic[uint64]
type Text = Generic[string]
type Pointer = Pointers[byte]
type Nested struct {
	Prefix byte
	Item Generic[uint32]
	Other Constrained[string]
}
func Identity[T any](value T) T { return value }
var InstantiatedFunction = Identity[int]
`

func TestGeneratedHeaderGenericPropertiesMatchActualNativeObjects(t *testing.T) {
	for _, target := range []string{"linux/386", "linux/amd64", "linux/arm", "linux/arm64", "js/wasm"} {
		t.Run(target, func(t *testing.T) {
			actual := compileNativeHeaderRules(t, target, nativeHeaderGenericPropertiesSource)
			for _, name := range []string{"Generic__size", "Constrained__size", "Pointers__size", "Identity"} {
				if _, emitted := actual[name]; emitted {
					t.Fatalf("actual Go compiler unexpectedly emitted an uninstantiated generic property: %s", name)
				}
			}
			if actual["Concrete__size"] == "" || actual["Text__size"] == "" || actual["Nested__size"] == "" {
				t.Fatalf("actual Go compiler did not emit concrete generic value properties: %v", actual)
			}
			replayed, err := ReplayNativeGeneratedHeader("example.invalid/header", target, "go1.20",
				map[string]string{"header.go": nativeHeaderGenericPropertiesSource})
			if err != nil || !reflect.DeepEqual(actual, replayed) {
				t.Fatalf("Go-accepted concrete generic properties differ from original source replay: actual=%v replay=%v error=%v",
					actual, replayed, err)
			}
		})
	}
}
