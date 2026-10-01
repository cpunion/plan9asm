//go:build unix

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func runDiscoveryCommand(cmd *exec.Cmd) error {
	// Own a fresh process group: canceling only Go leaves Git, compilers or
	// LLVM children alive. Also stop descendants after ordinary parent exit,
	// including children that closed their inherited output pipes.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return stopDiscoveryCommandGroup(cmd)
	}
	err := cmd.Run()
	if cmd.Process != nil {
		if cleanupErr := stopDiscoveryCommandGroup(cmd); cleanupErr != nil && !errors.Is(cleanupErr, os.ErrProcessDone) {
			err = errors.Join(err, fmt.Errorf("%w: %v", errDiscoveryCommandCleanup, cleanupErr))
		}
	}
	return err
}

func stopDiscoveryCommandGroup(cmd *exec.Cmd) error {
	err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return os.ErrProcessDone
	}
	return err
}
