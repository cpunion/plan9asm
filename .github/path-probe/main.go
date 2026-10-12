package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/xgo-dev/plan9asm/gotoolprofile"
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
		env, err := gotoolprofile.FeatureCommandEnvironment(os.Environ(), target, nil)
		if err != nil {
			panic(err)
		}
		out, stderr, err := gotoolprofile.RunBounded(ctx, "", env, binary, "env", "-json", "GOROOT", "GOTOOLDIR", "GOCACHE", "GOHOSTOS", "GOHOSTARCH")
		fmt.Printf("target=%s env=%s stderr=%s error=%v\n", target, out, stderr, err)
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
		dir, err := os.MkdirTemp("", "path-probe-")
		if err != nil {
			panic(err)
		}
		observed, err := gotoolprofile.Capture(ctx, binary, dir, os.Environ(), target, nil, gotoolprofile.RunBounded)
		fmt.Printf("capture target=%s observed=%v error=%v\n", target, observed != nil, err)
		if err := os.RemoveAll(dir); err != nil {
			panic(err)
		}
	}
}
