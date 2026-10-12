package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	binary, err := exec.LookPath("go")
	if err != nil {
		panic(err)
	}
	fmt.Printf("driver=%q\n", binary)
	for _, target := range []string{"linux/amd64", "darwin/amd64", "windows/amd64"} {
		parts := strings.Split(target, "/")
		command := exec.CommandContext(ctx, binary, "env", "-json", "GOROOT", "GOTOOLDIR", "GOCACHE", "GOHOSTOS", "GOHOSTARCH")
		command.Env = append(os.Environ(), "GOOS="+parts[0], "GOARCH="+parts[1], "CGO_ENABLED=0", "GOTOOLCHAIN=local", "GOWORK=off", "GOFLAGS=")
		out, err := command.CombinedOutput()
		fmt.Printf("target=%s env=%s error=%v\n", target, out, err)
		var actual map[string]string
		if err := json.Unmarshal(out, &actual); err != nil {
			panic(err)
		}
		for _, key := range []string{"GOROOT", "GOTOOLDIR", "GOCACHE"} {
			name := actual[key]
			info, statErr := os.Lstat(name)
			resolved, resolveErr := filepath.EvalSymlinks(name)
			fmt.Printf("%s path=%q stat=%v statError=%v resolved=%q resolveError=%v\n", key, name, info, statErr, resolved, resolveErr)
		}
	}
}
