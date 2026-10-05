package gotoolprofile

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestInstrumentationOriginalGeneratedAssemblyByteBudget(t *testing.T) {
	const goLimit = 2 << 20
	const assemblyLimit = 8 << 20
	for _, test := range []struct {
		name  string
		sizes map[string]int
		valid bool
	}{
		{"assembly above Go bound", map[string]int{"probe.s": 3 << 20}, true},
		{"assembly exact bound", map[string]int{"probe.s": assemblyLimit}, true},
		{"assembly over bound", map[string]int{"probe.s": assemblyLimit + 1}, false},
		{"Go exact bound", map[string]int{"probe.go": goLimit}, true},
		{"Go over bound", map[string]int{"probe.go": goLimit + 1}, false},
		{"inventory exact bound", map[string]int{"a.s": assemblyLimit, "b.s": assemblyLimit}, true},
		{"inventory over bound", map[string]int{"a.s": assemblyLimit, "b.s": assemblyLimit, "probe.go": 16}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := &ConsumerInput{
				Module: "example.invalid/sourcebudget", SourceRoot: t.TempDir(),
				Sources: make(map[string]string), Directories: map[string][]string{".": {}},
			}
			for name, size := range test.sizes {
				data := strings.Repeat(" ", size)
				if err := os.WriteFile(filepath.Join(input.SourceRoot, name), []byte(data), 0600); err != nil {
					t.Fatal(err)
				}
				input.Sources[name] = bytesSHA256([]byte(data))
				input.Directories["."] = append(input.Directories["."], name)
			}
			sort.Strings(input.Directories["."])
			inventory, err := CaptureInstrumentationOriginalSourceInventory(input, ".")
			if !test.valid {
				if err == nil {
					t.Fatal("accepted an original source beyond its file or inventory byte bound")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(inventory)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := DecodeInstrumentationOriginalSourceInventory(bytes.NewReader(data), input, "."); err != nil {
				t.Fatal(err)
			}
			for name := range inventory.Contents {
				inventory.Contents[name] += "changed"
				if err := ValidateInstrumentationOriginalSourceInventory(input, ".", inventory); err == nil {
					t.Fatal("large-source allowance erased original byte binding")
				}
				break
			}
		})
	}
}
