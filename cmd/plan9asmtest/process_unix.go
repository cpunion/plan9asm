//go:build unix

package main

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

func runOwnedCommand(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stop := func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	cmd.Cancel = stop
	err := cmd.Run()
	if cleanupErr := stop(); !errors.Is(cleanupErr, os.ErrProcessDone) {
		err = errors.Join(err, cleanupErr)
	}
	return err
}
