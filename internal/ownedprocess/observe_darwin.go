//go:build darwin

package ownedprocess

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const membershipOutputLimit = 1 << 20

type snapshotOutput struct {
	buffer   bytes.Buffer
	limit    int
	overflow bool
	cancel   context.CancelFunc
}

func (output *snapshotOutput) Write(data []byte) (int, error) {
	count := len(data)
	remaining := output.limit - output.buffer.Len()
	if len(data) > remaining {
		output.overflow = true
		data = data[:remaining]
		output.cancel()
	}
	_, _ = output.buffer.Write(data)
	return count, nil
}

func observeProcessGroup(pgid int) ([]groupMember, error) {
	// Use the OS's kernel inventory interface, not a shell, PATH-selected
	// executable, group member, or inherited locale/output-width setting.
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	// Darwin ps uses inclusive OR for these two explicit selectors. This
	// suppresses its default UID filter; -x keeps every no-TTY group member.
	// The kernel enumeration is complete, but expensive Mach task queries
	// run only for the selected group and this live Go observer. See:
	// https://github.com/apple-oss-distributions/adv_cmds/blob/main/ps/ps.c#L511-L715
	callerPID := os.Getpid()
	command := exec.CommandContext(ctx, "/bin/ps",
		"-x", "-g", strconv.Itoa(pgid), "-p", strconv.Itoa(callerPID),
		"-o", "pid=", "-o", "pgid=", "-o", "stat=",
	)
	command.Env = []string{"LC_ALL=C", "LANG=C", "PATH=/usr/bin:/bin", "COMMAND_MODE=unix2003"}
	command.Dir = "/"
	data, err := captureMembership(ctx, cancel, command)
	if err != nil {
		return nil, err
	}
	return parseMembership(data, pgid, callerPID)
}

func captureMembership(ctx context.Context, cancel context.CancelFunc, command *exec.Cmd) ([]byte, error) {
	command.WaitDelay = 250 * time.Millisecond
	stdout := snapshotOutput{limit: membershipOutputLimit, cancel: cancel}
	stderr := snapshotOutput{limit: 4096, cancel: cancel}
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if stdout.overflow || stderr.overflow {
		return nil, fmt.Errorf("Darwin process membership exceeded its output bound")
	}
	if err != nil || ctx.Err() != nil {
		return nil, fmt.Errorf("Darwin process membership command: %w", errors.Join(err, ctx.Err()))
	}
	if stderr.buffer.Len() != 0 {
		return nil, fmt.Errorf("Darwin process membership produced stderr")
	}
	return stdout.buffer.Bytes(), nil
}

func parseMembership(data []byte, pgid, callerPID int) ([]groupMember, error) {
	if len(data) == 0 || len(data) > membershipOutputLimit || data[len(data)-1] != '\n' {
		return nil, fmt.Errorf("Darwin process membership is empty, oversized or truncated")
	}
	seen := make(map[int]bool)
	var members []groupMember
	for line, text := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n") {
		fields := strings.Fields(text)
		if len(fields) != 3 {
			return nil, fmt.Errorf("invalid Darwin process membership row %d", line+1)
		}
		pid, pidErr := decimalPID(fields[0])
		group, groupErr := decimalPID(fields[1])
		if pidErr != nil || groupErr != nil || pid == 0 || seen[pid] || !validStateToken(fields[2]) {
			return nil, fmt.Errorf("invalid Darwin process membership values at row %d", line+1)
		}
		seen[pid] = true
		if pid == pgid && group != pgid {
			return nil, fmt.Errorf("Darwin owned leader identity changed")
		}
		if group != pgid && pid != callerPID {
			return nil, fmt.Errorf("Darwin process membership has an unselected row")
		}
		if group == pgid {
			if !validDarwinState(fields[2]) {
				return nil, fmt.Errorf("Darwin owned member state is unknown")
			}
			members = append(members, groupMember{PID: pid, PGID: group, State: fields[2]})
		}
	}
	if !seen[callerPID] {
		return nil, fmt.Errorf("Darwin process membership omitted the observing process")
	}
	return members, nil
}

func validStateToken(text string) bool {
	if len(text) == 0 || len(text) > 16 {
		return false
	}
	for _, char := range text {
		if char < '!' || char > '~' {
			return false
		}
	}
	return true
}

func decimalPID(text string) (int, error) {
	for _, char := range text {
		if char < '0' || char > '9' {
			return 0, fmt.Errorf("nondecimal process identity")
		}
	}
	value, err := strconv.ParseUint(text, 10, 31)
	return int(value), err
}
