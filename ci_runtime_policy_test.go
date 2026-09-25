package plan9asm

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestCIFullSuitesHaveExplicitTimeout(t *testing.T) {
	data, err := os.ReadFile(".github/workflows/go-ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	commands := 0
	for line, command := range strings.Split(string(data), "\n") {
		if !strings.Contains(command, "go test ") || !strings.Contains(command, "./...") {
			continue
		}
		commands++
		var timeout time.Duration
		for _, field := range strings.Fields(command) {
			if strings.HasPrefix(field, "-timeout=") {
				timeout, err = time.ParseDuration(strings.TrimPrefix(field, "-timeout="))
				if err != nil {
					t.Errorf("line %d: invalid timeout: %v", line+1, err)
				}
			}
		}
		if timeout < 20*time.Minute {
			t.Errorf("line %d: full suite requires an explicit timeout of at least 20m: %s", line+1, strings.TrimSpace(command))
		}
	}
	if commands == 0 {
		t.Fatal("no full-suite CI commands found")
	}
}
