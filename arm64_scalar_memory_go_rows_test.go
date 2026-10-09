package plan9asm

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func arm64MOVSelectedGoSource(t *testing.T) (string, string) {
	t.Helper()
	root := testGOROOT(t)
	source, err := os.ReadFile(filepath.Join(root, "src/cmd/internal/obj/arm64/asm7.go"))
	if err != nil {
		t.Fatal(err)
	}
	version, err := os.ReadFile(filepath.Join(root, "VERSION"))
	if err != nil {
		t.Fatal(err)
	}
	name := strings.SplitN(strings.TrimSpace(string(version)), "\n", 2)[0]
	t.Logf("MOV_ROW_SOURCE version=%s source_sha256=%x", name, sha256.Sum256(source))
	return string(source), name
}

type arm64MOVAttributeMutation struct {
	name, source string
}

// These mutations edit complete selected-SDK rows, not production properties.
// Each preserves all per-opcode/case counts and the unsigned alias statements.
func arm64MOVSameCountMutations(source string) ([]arm64MOVAttributeMutation, error) {
	var changes []arm64MOVAttributeMutation
	for _, name := range []string{"param", "pool_flag", "prepost", "code_size", "class", "case_role"} {
		lines := strings.Split(source, "\n")
		changed := 0
		for index, line := range lines {
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "{AMOVH,") {
				continue
			}
			end := strings.IndexByte(trimmed, '}')
			if end < 0 {
				return nil, fmt.Errorf("unterminated original MOV row")
			}
			fields := strings.Split(trimmed[1:end], ",")
			if len(fields) != 10 && len(fields) != 11 {
				return nil, fmt.Errorf("unknown original MOV row shape")
			}
			for field := range fields {
				fields[field] = strings.TrimSpace(fields[field])
			}
			kindAt := len(fields) - 5
			kind, err := strconv.Atoi(fields[kindAt])
			if err != nil {
				return nil, err
			}
			edit := false
			switch name {
			case "param":
				edit = changed == 0 && kind == 20 && fields[4] == "C_UAUTO8K"
				if edit {
					fields[kindAt+2] = "0"
				}
			case "pool_flag":
				edit = changed == 0 && kind == 30 && fields[kindAt+3] == "LTO"
				if edit {
					fields[kindAt+3] = "0"
				}
			case "prepost", "class":
				edit = changed == 0 && kind == 22 && fields[kindAt+4] == "C_XPRE"
				if edit && name == "prepost" {
					fields[kindAt+4] = "C_XPOST"
				} else if edit {
					fields[1] = "C_NSOREG"
				}
			case "code_size":
				edit = changed == 0 && kind == 30
				if edit {
					fields[kindAt+1] = "4"
				}
			case "case_role":
				edit = changed == 0 && kind == 20 || changed == 1 && kind == 21
				if edit {
					fields[kindAt] = strconv.Itoa(41 - kind)
				}
			}
			if edit {
				lines[index] = "\t{" + strings.Join(fields, ", ") + "},"
				changed++
			}
		}
		want := 1
		if name == "case_role" {
			want = 2
		}
		if changed != want {
			return nil, fmt.Errorf("%s requires %d actual row edits, got %d", name, want, changed)
		}
		changes = append(changes, arm64MOVAttributeMutation{name, strings.Join(lines, "\n")})
	}
	return changes, nil
}

func TestARM64ScalarMemoryGoRowsRejectSameCountAttributeMutations(t *testing.T) {
	source, _ := arm64MOVSelectedGoSource(t)
	if _, err := arm64ScalarMemoryLegacyGoRows(source); err != nil {
		t.Fatalf("unmodified actual SDK prerequisites: %v", err)
	}
	mutations, err := arm64MOVSameCountMutations(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			if mutation.source == source {
				t.Fatal("attribute mutation did not change the actual input")
			}
			if _, err := arm64ScalarMemoryLegacyGoRows(mutation.source); err == nil {
				t.Fatal("MOV table audit accepted a changed row attribute with identical counts/aliases")
			}
		})
	}
}
