package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"go/token"
	"regexp"
	"sort"
	"strings"
)

func validName(name string) bool {
	return token.IsIdentifier(name) && (strings.HasPrefix(name, "Test") || strings.HasPrefix(name, "Example") || strings.HasPrefix(name, "Fuzz"))
}

var listFooter = regexp.MustCompile(`^ok[ \t]+([^ \t]+)[ \t]+[0-9]+(?:\.[0-9]+)?s(?:[ \t]+(?:\[no tests to run\]|coverage: (?:[0-9]+(?:\.[0-9]+)?% of statements(?: in .+)?|\[no statements\])))?$`)

// go test -list lists top-level tests/examples/fuzz targets, not dynamic
// subtests or seeds. The execution validator also checks every emitted child.
func parseList(data string) (string, []string, []string, error) {
	var pkg string
	var names []string
	var benchmarks []string
	seen := make(map[string]bool)
	for _, line := range strings.Split(strings.TrimSuffix(data, "\n"), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if pkg != "" {
			return "", nil, nil, fmt.Errorf("output after inventory footer: %q", line)
		}
		if strings.HasPrefix(line, "ok\t") || strings.HasPrefix(line, "ok ") {
			fields := listFooter.FindStringSubmatch(line)
			if fields == nil {
				return "", nil, nil, fmt.Errorf("invalid inventory footer: %q", line)
			}
			pkg = fields[1]
			continue
		}
		benchmark := token.IsIdentifier(line) && strings.HasPrefix(line, "Benchmark")
		if !validName(line) && !benchmark || seen[line] {
			return "", nil, nil, fmt.Errorf("invalid/duplicate listed test: %q", line)
		}
		seen[line] = true
		if benchmark {
			benchmarks = append(benchmarks, line)
		} else {
			names = append(names, line)
		}
	}
	if pkg == "" || len(names) == 0 {
		return "", nil, nil, fmt.Errorf("empty inventory or missing successful package footer")
	}
	sort.Strings(names)
	sort.Strings(benchmarks)
	return pkg, names, benchmarks, nil
}

func partition(names []string, count int) ([][]string, error) {
	if len(names) == 0 || count < 1 || count > 1024 {
		return nil, fmt.Errorf("nonempty inventory and 1..1024 shards required")
	}
	shards := make([][]string, count)
	seen := make(map[string]bool)
	for _, name := range names {
		if !validName(name) || seen[name] {
			return nil, fmt.Errorf("invalid/duplicate planned test: %q", name)
		}
		seen[name] = true
		hash := sha256.Sum256([]byte(name))
		index := binary.BigEndian.Uint64(hash[:8]) % uint64(count)
		shards[index] = append(shards[index], name)
	}
	for _, shard := range shards {
		sort.Strings(shard)
	}
	return shards, nil
}

func runPattern(names []string) (string, error) {
	quoted := make([]string, len(names))
	for i, name := range names {
		quoted[i] = regexp.QuoteMeta(name)
	}
	pattern := "^(" + strings.Join(quoted, "|") + ")$"
	// Conservative even on Unix: never truncate a regex to fit a platform's
	// command-line limit. More shards are required instead.
	if len(names) == 0 || len(pattern) > 16000 {
		return "", fmt.Errorf("empty/oversized exact run pattern (%d bytes); increase -shards", len(pattern))
	}
	return pattern, nil
}

// Only build/instrumentation flags are forwarded. Test filters, repeats,
// fuzzing, custom executors and arbitrary test arguments change this protocol.
func validateFlags(args []string) ([]string, string, error) {
	boolFlags := map[string]bool{"race": true, "cover": true, "trimpath": true, "asan": true, "msan": true, "work": true}
	valueFlags := map[string]bool{"covermode": true, "coverpkg": true, "tags": true, "p": true, "gcflags": true, "asmflags": true, "ldflags": true, "mod": true, "modfile": true, "overlay": true}
	var forwarded []string
	var profile string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "--") {
			return nil, "", fmt.Errorf("only supported Go flags are allowed: %q", arg)
		}
		name, value, hasValue := strings.Cut(strings.TrimPrefix(arg, "-"), "=")
		if boolFlags[name] {
			if hasValue && value != "true" && value != "false" {
				return nil, "", fmt.Errorf("invalid boolean flag %q", arg)
			}
			forwarded = append(forwarded, arg)
			continue
		}
		if !valueFlags[name] && name != "coverprofile" {
			return nil, "", fmt.Errorf("scope-changing/unknown Go flag %q", arg)
		}
		if !hasValue {
			i++
			if i == len(args) {
				return nil, "", fmt.Errorf("missing value for %q", arg)
			}
			value = args[i]
		}
		if value == "" {
			return nil, "", fmt.Errorf("empty value for %q", arg)
		}
		if name == "coverprofile" {
			if profile != "" {
				return nil, "", fmt.Errorf("duplicate coverage destination")
			}
			profile = value
			continue
		}
		forwarded = append(forwarded, arg)
		if !hasValue {
			forwarded = append(forwarded, value)
		}
	}
	return forwarded, profile, nil
}

func largeScanner(reader interface{ Read([]byte) (int, error) }) *bufio.Scanner {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 4<<20)
	return scanner
}
