package goasmcmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestActualGoCommandBoundsAndMachineStdout(t *testing.T) {
	for _, mode := range []string{"json", "stderr", "exit", "stdout overflow", "stderr overflow", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if mode == "cancel" {
				ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()
			}
			env := append(os.Environ(), "GOASM_COMMAND_FIXTURE="+mode)
			data, err := Run(ctx, t.TempDir(), env, os.Args[0], "-test.run=^TestActualGoCommandFixture$")
			if mode == "json" {
				if err != nil || string(data) != `{"actual":true}` {
					t.Fatalf("machine stdout changed: %q %v", data, err)
				}
				return
			}
			if err == nil || data != nil {
				t.Fatalf("command %s falsely succeeded: %q %v", mode, data, err)
			}
			if mode == "cancel" && !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("cancellation lost deadline evidence: %v", err)
			}
			if mode == "stderr" && !strings.Contains(err.Error(), "download diagnostic") {
				t.Fatalf("stderr was swallowed: %v", err)
			}
			if strings.HasSuffix(mode, "overflow") && !strings.Contains(err.Error(), "output bound exceeded") {
				t.Fatalf("overflow was swallowed: %v", err)
			}
		})
	}
}

func TestActualGoCommandFixture(t *testing.T) {
	switch os.Getenv("GOASM_COMMAND_FIXTURE") {
	case "json":
		_, _ = os.Stdout.WriteString(`{"actual":true}`)
	case "stderr":
		_, _ = os.Stdout.WriteString(`{"actual":true}`)
		_, _ = os.Stderr.WriteString("download diagnostic")
	case "stdout overflow":
		_, _ = os.Stdout.Write(bytes.Repeat([]byte("x"), 4<<20+1))
	case "stderr overflow":
		_, _ = os.Stderr.Write(bytes.Repeat([]byte("x"), 4<<20+1))
	case "exit":
		os.Exit(7)
	case "cancel":
		time.Sleep(time.Minute)
	default:
		return
	}
	os.Exit(0)
}

func TestActualGoCommandRequiresLiveContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, input := range []context.Context{nil, ctx} {
		if _, err := Run(input, t.TempDir(), nil, "not-started"); err == nil {
			t.Fatal("inactive context started a command")
		}
	}
}
