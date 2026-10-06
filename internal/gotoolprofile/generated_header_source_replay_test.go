package gotoolprofile

import (
	"reflect"
	"strings"
	"testing"
)

func TestGeneratedHeaderSourceReplayMatchesOriginalNativeObjects(t *testing.T) {
	// Generic declarations are outside the initial closed source domain even
	// though the compiler omits their header entries. Preserve that refusal.
	source := strings.ReplaceAll(nativeHeaderRulesSource, "type Generic[T any] struct { Value T }\n", "")
	for _, target := range []string{"linux/386", "linux/amd64", "linux/arm", "linux/arm64", "js/wasm"} {
		t.Run(target, func(t *testing.T) {
			actual := compileNativeHeaderRules(t, target, source)
			replayed, err := ReplayNativeGeneratedHeader("example.invalid/header", target, "go1.20", map[string]string{"header.go": source})
			if err != nil || !reflect.DeepEqual(actual, replayed) {
				t.Fatalf("complete native compiler header differs from independent original source replay:\nactual=%v\nreplayed=%v\nerror=%v", actual, replayed, err)
			}
		})
	}
}

func TestGeneratedHeaderSourceReplayRejectsUnclosedInputs(t *testing.T) {
	for _, source := range []string{
		"package header\nimport \"unsafe\"\nconst Size = unsafe.Sizeof(uintptr(0))\n",
		"package header\ntype Record struct { Pointer *byte }\n",
		"package header\ntype Record struct { Bytes []byte }\n",
		"package header\ntype Record struct { Text string }\n",
		"package header\ntype Record[T any] struct { Value T }\n",
		"package header\ntype Record struct { Value [1<<62]uint64 }\n",
		"package header\ntype Record struct { Self Record }\n",
		"package header\ntype A struct { B__size byte }\ntype A_B struct {}\n",
	} {
		if definitions, err := ReplayNativeGeneratedHeader("example.invalid/header", "linux/amd64", "go1.20", map[string]string{"header.go": source}); err == nil || definitions != nil {
			t.Fatalf("unclosed source acquired header definitions: %q %v %v", source, definitions, err)
		}
	}
}

func TestGeneratedHeaderSourceReplayRetainsBlankZeroTailAndMultifileIdentity(t *testing.T) {
	sources := map[string]string{
		"first.go":  "package header\nconst First = 7\ntype Record struct { First uint64; _ byte; Tail [0]byte }\n",
		"second.go": "package header\nconst Second = First + 1\ntype Alias = Record\n",
	}
	wanted := map[string]string{"const_First": "7", "const_Second": "8", "Record__size": "16", "Record_First": "0", "Record_Tail": "9", "Alias__size": "16", "Alias_First": "0", "Alias_Tail": "9"}
	actual, err := ReplayNativeGeneratedHeader("example.invalid/header", "linux/amd64", "go1.20", sources)
	if err != nil || !reflect.DeepEqual(actual, wanted) {
		t.Fatalf("multifile/blank/zero-tail native source replay: %v %v", actual, err)
	}
	for _, mutate := range []func(map[string]string){
		func(copy map[string]string) { delete(copy, "first.go") },
		func(copy map[string]string) { copy["second.go"] = "package different\nconst Second = 8\n" },
		func(copy map[string]string) { copy["second_test.go"] = copy["second.go"]; delete(copy, "second.go") },
	} {
		copy := make(map[string]string)
		for name, text := range sources {
			copy[name] = text
		}
		mutate(copy)
		if definitions, err := ReplayNativeGeneratedHeader("example.invalid/header", "linux/amd64", "go1.20", copy); err == nil || definitions != nil {
			t.Fatalf("changed original source acquired a complete header: %v %v", definitions, err)
		}
	}
}
