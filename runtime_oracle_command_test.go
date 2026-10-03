package plan9asm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"testing"
	"time"
)

// Keep command output and the original exit, but never wait indefinitely for
// an emulator retaining its parent's pipe. Unix also owns the fresh group.
func runtimeOracleCombinedOutput(ctx context.Context, command *exec.Cmd) ([]byte, error) {
	if ctx == nil || command == nil {
		return nil, fmt.Errorf("runtime oracle requires a command and its context")
	}
	if command.Stdout != nil || command.Stderr != nil {
		return nil, fmt.Errorf("runtime oracle command already has output writers")
	}
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	command.WaitDelay = 2 * time.Second
	err := runtimeOracleRun(command)
	if ctx.Err() != nil {
		err = errors.Join(err, ctx.Err())
	}
	return output.Bytes(), err
}

func TestRuntimeOracleCommandRequiresBoundContextAndOutput(t *testing.T) {
	ctx := context.Background()
	command := exec.CommandContext(ctx, os.Args[0])
	for _, test := range []struct {
		name    string
		ctx     context.Context
		command *exec.Cmd
	}{
		{name: "no context", command: command},
		{name: "no command", ctx: ctx},
	} {
		t.Run(test.name, func(t *testing.T) {
			if output, err := runtimeOracleCombinedOutput(test.ctx, test.command); err == nil || len(output) != 0 {
				t.Fatalf("unbound runtime command accepted: %v %q", err, output)
			}
		})
	}
	var existing bytes.Buffer
	command.Stdout = &existing
	if output, err := runtimeOracleCombinedOutput(ctx, command); err == nil || len(output) != 0 || command.Stdout != &existing || command.Process != nil {
		t.Fatalf("runtime command replaced an output writer or ran: %v %q", err, output)
	}

	expired, cancel := context.WithDeadline(ctx, time.Now().Add(-time.Second))
	defer cancel()
	command = exec.CommandContext(expired, os.Args[0])
	if output, err := runtimeOracleCombinedOutput(expired, command); !errors.Is(err, context.DeadlineExceeded) || len(output) != 0 || command.Process != nil {
		t.Fatalf("expired command was launched or lost its deadline: %v %q", err, output)
	}
}

func TestRuntimeOracleCommandConsumersRetainOwnedExecution(t *testing.T) {
	// Every Go compiler/executable/emulator command in these native oracles
	// must use the same bounded lifecycle. Do not just fix the discovering case.
	for _, name := range []string{
		"arm64_local_register_branch_test.go",
		"arm64_gp_data_runtime_test.go",
		"arm64_private_structure_runtime_test.go",
		"arm64_large_frame_runtime_test.go",
		"arm64_control_gp_move_runtime_test.go",
		"arm64_go_internal_aggregate_test.go",
		"arm64_control_scalar_multiply_runtime_test.go",
		"arm64_raw_zero_frame_runtime_test.go",
		"arm64_raw_private_frame_runtime_test.go",
		"arm64_control_memory_width_runtime_test.go",
		"arm64_raw_context_runtime_test.go",
		"arm64_sm3_runtime_test.go",
		"arm64_raw_self_bl_runtime_test.go",
	} {
		t.Run(name, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), name, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			consumers := 0
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				if function, ok := call.Fun.(*ast.Ident); ok && function.Name == "runtimeOracleCombinedOutput" {
					consumers++
					return true
				}
				method, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				switch method.Sel.Name {
				case "CombinedOutput", "Output", "Start", "Run":
				default:
					return true
				}
				if receiver, ok := method.X.(*ast.Ident); ok && receiver.Name == "t" {
					return true
				}
				t.Errorf("runtime command bypasses its owned lifecycle: %s", method.Sel.Name)
				return true
			})
			if consumers == 0 {
				t.Fatal("native oracle has no bounded command consumer")
			}
		})
	}
}
