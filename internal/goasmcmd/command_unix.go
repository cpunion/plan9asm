//go:build unix

package goasmcmd

import (
	"github.com/xgo-dev/plan9asm/internal/ownedprocess"
	"os/exec"
)

func run(command *exec.Cmd) error { return ownedprocess.Run(command) }
