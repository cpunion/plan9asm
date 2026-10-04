// Package goasmcmd owns bounded, machine-readable stock-Go commands. It has no
// dependency on the translator or its proof protocols.
package goasmcmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

type output struct {
	buffer   bytes.Buffer
	overflow bool
}

func (out *output) Len() int      { return out.buffer.Len() }
func (out *output) Bytes() []byte { return out.buffer.Bytes() }

func (out *output) Write(data []byte) (int, error) {
	n := len(data)
	if remaining := 4<<20 - out.Len(); n > remaining {
		out.overflow = true
		data = data[:remaining]
	}
	_, _ = out.buffer.Write(data)
	return n, nil
}

// Run keeps stdout separate from diagnostics and owns only its fresh child.
// Unix implementations additionally terminate that child's process group.
func Run(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
	if ctx == nil || ctx.Err() != nil {
		return nil, fmt.Errorf("Go object command requires a live context")
	}
	command := exec.CommandContext(ctx, name, args...)
	command.Dir, command.Env, command.WaitDelay = dir, env, 2*time.Second
	var stdout, stderr output
	command.Stdout, command.Stderr = &stdout, &stderr
	err := run(command)
	if ctx.Err() != nil {
		err = errors.Join(err, ctx.Err())
	}
	if stdout.overflow || stderr.overflow {
		err = errors.Join(err, fmt.Errorf("Go object command output bound exceeded"))
	}
	if stderr.Len() != 0 {
		err = errors.Join(err, fmt.Errorf("Go object command diagnostics: %s", stderr.Bytes()))
	}
	if err != nil {
		return nil, fmt.Errorf("actual Go %v: %w", args, err)
	}
	return append([]byte(nil), stdout.Bytes()...), nil
}
