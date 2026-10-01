//go:build unix

package gotoolprofile

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func runOwnedCommand(command *exec.Cmd) error {
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stop := func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.Cancel = stop
	err := command.Run()
	if command.Process != nil {
		if cleanup := stop(); cleanup != nil && !errors.Is(cleanup, os.ErrProcessDone) {
			err = errors.Join(err, fmt.Errorf("owned command descendants: %w", cleanup))
		}
	}
	return err
}
