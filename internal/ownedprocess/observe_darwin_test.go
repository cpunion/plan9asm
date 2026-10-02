//go:build darwin

package ownedprocess

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestDarwinMembershipSnapshotRejectsUnknownEvidence(t *testing.T) {
	const callerPID = 100
	valid := "100 99 S\n4124 4123 Z+\n4125 4123 Zs\n"
	members, err := parseMembership([]byte(valid), 4123, callerPID)
	want := []groupMember{{PID: 4124, PGID: 4123, State: "Z+"}, {PID: 4125, PGID: 4123, State: "Zs"}}
	if err != nil || !reflect.DeepEqual(members, want) {
		t.Fatalf("exact owned subset: %v %v", members, err)
	}
	members, err = parseMembership([]byte("100 99 S\n"), 4123, callerPID)
	if err != nil || len(members) != 0 {
		t.Fatalf("independently absent group: %v %v", members, err)
	}
	// This Go observer is alive independently of ps's state spelling. Its
	// exact numeric group proves it is not an owned member; it need not exit.
	members, err = parseMembership([]byte("100 99 ?E\n4124 4123 Z\n"), 4123, callerPID)
	if err != nil || len(members) != 1 || members[0].PID != 4124 {
		t.Fatalf("unrelated process state broadened the owned group: %v %v", members, err)
	}
	for name, text := range map[string]string{
		"empty":             "",
		"truncated":         "100 99 S",
		"missing caller":    "4124 4123 Z\n",
		"extra column":      "100 99 S hidden-command\n",
		"missing column":    "100 99\n",
		"blank row":         "100 99 S\n\n",
		"unknown state":     "100 99 S\n4124 4123 ?E\n",
		"unknown suffix":    "100 99 S\n4124 4123 Z?\n",
		"duplicate pid":     "100 99 S\n100 4123 Z\n",
		"zero pid":          "100 99 S\n0 4123 Z\n",
		"negative pid":      "100 99 S\n-1 4123 Z\n",
		"negative pgid":     "100 -99 S\n",
		"signed pid":        "+100 99 S\n",
		"overflow pid":      "2147483648 99 S\n",
		"overflow pgid":     "100 2147483648 S\n",
		"nondecimal pid":    "0x64 99 S\n",
		"oversized":         strings.Repeat(" ", membershipOutputLimit) + "\n",
		"changed leader":    "100 99 S\n4123 3000 ?E\n",
		"extra row":         "100 99 S\n4000 3000 Z\n",
		"unknown extra row": "100 99 S\n4000 3000 ?E\n",
		"binary status":     "100 99 S\n4000 3000 \x00\n",
		"oversized status":  "100 99 S\n4000 3000 " + strings.Repeat("?", 17) + "\n",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := parseMembership([]byte(text), 4123, callerPID)
			if err == nil {
				t.Fatal("unknown/partial kernel snapshot was accepted")
			}
			if strings.Contains(err.Error(), "hidden-command") {
				t.Fatal("process output leaked through diagnostic")
			}
		})
	}
}

func TestDarwinMembershipCommandFailureBounds(t *testing.T) {
	for _, mode := range []string{"stdout-overflow", "stderr-overflow", "stderr", "exit", "timeout"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
			defer cancel()
			command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestDarwinMembershipCommandFixture$")
			command.Env = append(os.Environ(), "PLAN9ASM_MEMBERSHIP_FIXTURE="+mode)
			start := time.Now()
			data, err := captureMembership(ctx, cancel, command)
			if err == nil || data != nil {
				t.Fatalf("failed membership command accepted: bytes=%d err=%v", len(data), err)
			}
			if mode == "timeout" && !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("lost timeout: %v", err)
			}
			if time.Since(start) > 2*time.Second {
				t.Fatalf("membership command exceeded its total two-second budget: %v", time.Since(start))
			}
			if strings.Contains(err.Error(), "private-process-output") {
				t.Fatal("snapshot output leaked through diagnostic")
			}
		})
	}
}

func TestDarwinMembershipCommandFixture(t *testing.T) {
	switch os.Getenv("PLAN9ASM_MEMBERSHIP_FIXTURE") {
	case "":
		return
	case "stdout-overflow":
		fmt.Fprint(os.Stdout, strings.Repeat("x", membershipOutputLimit+1))
	case "stderr-overflow":
		fmt.Fprint(os.Stderr, strings.Repeat("x", 4097))
	case "stderr":
		fmt.Fprint(os.Stderr, "private-process-output")
	case "exit":
		fmt.Fprint(os.Stdout, "private-process-output")
		os.Exit(7)
	case "timeout":
		time.Sleep(5 * time.Second)
	default:
		os.Exit(9)
	}
	os.Exit(0)
}

func TestDarwinMembershipIncludesNoTTYLiveMembers(t *testing.T) {
	if os.Getenv("PLAN9ASM_NO_TTY_FIXTURE") == "1" {
		if err := os.WriteFile("ready", nil, 0600); err != nil {
			os.Exit(5)
		}
		time.Sleep(10 * time.Second)
		os.Exit(0)
	}
	dir := t.TempDir()
	child := exec.Command(os.Args[0], "-test.run=^TestDarwinMembershipIncludesNoTTYLiveMembers$")
	child.Dir = dir
	child.Env = append(os.Environ(), "PLAN9ASM_NO_TTY_FIXTURE=1")
	// A new session has no controlling terminal. It must not disappear
	// from the ps group selector, even when its creator is in another one.
	child.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = child.Process.Kill()
		_ = child.Wait()
	}()
	for end := time.Now().Add(3 * time.Second); ; {
		if _, err := os.Stat(filepath.Join(dir, "ready")); err == nil {
			break
		}
		if time.Now().After(end) {
			t.Fatal("no-TTY child did not start")
		}
		time.Sleep(5 * time.Millisecond)
	}
	pgid, err := syscall.Getpgid(child.Process.Pid)
	if err != nil || pgid != child.Process.Pid {
		t.Fatalf("test-owned independent group: %d %v", pgid, err)
	}
	members, err := observeProcessGroup(pgid)
	if err != nil || len(members) != 1 || members[0].PID != child.Process.Pid || members[0].State[0] == 'Z' {
		t.Fatalf("no-TTY live group member was hidden: %v %v", members, err)
	}
	permission := stopGroup(pgid, true, func(int, syscall.Signal) error { return syscall.EPERM }, observeProcessGroup)
	if !errors.Is(permission, syscall.EPERM) || errors.Is(permission, os.ErrProcessDone) {
		t.Fatalf("actual live no-TTY group suppressed a permission error: %v", permission)
	}
}
