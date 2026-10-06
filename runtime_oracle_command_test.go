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

const (
	runtimeOracleStageTimeout        = 5 * time.Minute
	runtimeOracleStageCleanupReserve = 3 * time.Second
)

func runtimeOracleStageContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	deadline, hasDeadline := t.Deadline()
	return context.WithDeadline(context.Background(), runtimeOracleStageDeadline(time.Now(), deadline, hasDeadline))
}

func runtimeOracleStageDeadline(now, testDeadline time.Time, hasTestDeadline bool) time.Time {
	deadline := now.Add(runtimeOracleStageTimeout)
	if hasTestDeadline {
		// Reserve the existing bounded pipe/group cleanup before testing's
		// overall deadline. An exhausted budget stays expired: do not launch.
		latest := testDeadline.Add(-runtimeOracleStageCleanupReserve)
		if latest.Before(deadline) {
			deadline = latest
		}
	}
	return deadline
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
		"arm64_closed_tail_stdlib_test.go",
		"stdlib_arm64_test_helpers_test.go",
		"llvm_object_lifecycle_unix_test.go",
		"runtime_exec_test.go",
		"runtime_oracle_fault_core_unix_test.go",
		"amd64_go_caller_lea_test.go",
		"go_translate_test.go",
		"amd64_go_caller_lea_runtime_test.go",
		"amd64_raw_lea_runtime_test.go",
		"amd64_go_caller_acyclic_runtime_test.go",
		"amd64_go_caller_counted_loop_runtime_test.go",
		"amd64_fixed_local_runtime_test.go",
		"x86_result_initialization_test.go",
		"x86_result_auto_runtime_test.go",
		"x86_result_typed_effects_runtime_test.go",
		"amd64_legacy_packed_move_runtime_test.go",
		"amd64_legacy_integer_arithmetic_runtime_test.go",
		"amd64_packed_word_multiply_runtime_test.go",
		"amd64_packed_narrow_independent_runtime_test.go",
		"x86_explicit_imul_immediate_semantics_test.go",
		"amd64_fixed_local_frame_test.go",
		"amd64_go_caller_saved_g_runtime_test.go",
		"amd64_go_caller_logical_runtime_test.go",
		"amd64_float_add_sub_go_oracle_test.go",
		"amd64_float_add_sub_literal_runtime_test.go",
		"amd64_float_compare_oracle_test.go",
		"amd64_float_compare_original_vector_test.go",
		"amd64_float_compare_vex_numerical_test.go",
		"amd64_float_compare_transport_test.go",
		"amd64_float_compare_preflight_portability_test.go",
		"amd64_conformance_native_test.go",
		"amd64_insert_extract_native_runtime_test.go",
		"amd64_packed_double_dword_native_test.go",
		"amd64_packed_narrow_rules_test.go",
		"amd64_maddword_go_oracle_test.go",
		"amd64_madd_typed_rules_test.go",
		"amd64_madd_typed_runtime_test.go",
		"amd64_timestamp_runtime_test.go",
		"amd64_cpu_query_consumers_test.go",
		"amd64_cpu_query_runtime_unix_test.go",
		"amd64_cpu_query_native_reference_test.go",
		"amd64_cpu_query_linux_amd64_test.go",
		"amd64_timestamp_linux_amd64_test.go",
		"amd64_timestamp_rule_consumers_test.go",
		"amd64_undefined_trap_consumers_test.go",
		"amd64_undefined_trap_runtime_unix_test.go",
		"amd64_undefined_trap_linux_amd64_test.go",
		"amd64_packed_add_go_oracle_test.go",
		"amd64_auxiliary_carry_runtime_test.go",
		"amd64_scalar_flag_compare_af_runtime_test.go",
		"amd64_x87_flag_compare_test.go",
		"amd64_x87_native_ia32_test.go",
		"amd64_packed_subtract_go_oracle_test.go",
		"amd64_go_caller_maddword_runtime_test.go",
		"amd64_packed_sad_runtime_test.go",
		"amd64_conditional_set_runtime_test.go",
		"amd64_packed_sad_forms_test.go",
		"amd64_signed_integer_scalar_float_go_oracle_test.go",
		"amd64_float_binary_fp_span_test.go",
		"amd64_legacy_float_memory_test.go",
		"amd64_horizontal_float_go_oracle_test.go",
		"amd64_packed_unpack_go_oracle_test.go",
		"amd64_packed_zero_extend_runtime_test.go",
		"amd64_packed_zero_extend_raw_axes_test.go",
		"amd64_signed_packed_integer_float_runtime_test.go",
		"amd64_signed_packed_integer_float_resource_unix_test.go",
		"amd64_signed_packed_integer_float_raw_test.go",
		"amd64_scalar_vector_literal_test.go",
		"amd64_scalar_vector_literal_source_test.go",
		"amd64_scalar_vector_literal_runtime_test.go",
		"amd64_insert_extract_go_oracle_test.go",
		"amd64_go_caller_vector_zero_test.go",
		"arm64_local_register_branch_test.go",
		"arm64_gp_data_runtime_test.go",
		"arm64_gp_immediate_encoding_test.go",
		"arm64_lower_permute_test.go",
		"amd64_go_assembler_test.go",
		"go_assembler_oracle_test.go",
		"go_assembler_oracle_listing_test.go",
		"go_assembler_defect_go127_test.go",
		"go_assembler_oracle_infrastructure_go127_unix_test.go",
		"arm64_private_structure_runtime_test.go",
		"arm64_large_frame_runtime_test.go",
		"arm64_control_gp_move_runtime_test.go",
		"arm64_go_internal_aggregate_test.go",
		"arm64_control_scalar_multiply_runtime_test.go",
		"arm64_raw_zero_frame_runtime_test.go",
		"arm64_raw_private_frame_runtime_test.go",
		"wasm_packed_go_runtime_test.go",
		"wasm_packed_node_compat_test.go",
		"arm64_control_memory_width_runtime_test.go",
		"arm64_raw_context_runtime_test.go",
		"arm64_sm3_runtime_test.go",
		"arm64_raw_self_bl_runtime_test.go",
		"arm64_scalar_memory_encoding_audit_test.go",
		"arm64_fp_pair_observation_test.go",
		"arm64_fp_padding_runtime_test.go",
		"arm64_result_tail_native_runtime_test.go",
		"amd64_mov_extension_caller_test.go",
		"amd64_mov_extension_runtime_test.go",
		"amd64_raw_movsxd_runtime_test.go",
		"amd64_float_compare_native_test.go",
		"amd64_raw_insertps_mode32_test.go",
		"arm64_scalar_writeback_test.go",
		"arm64_barrier_encoding_go127_test.go",
		"arm64_ret_source_width_test.go",
		"arm64_opaque_preservation_observation_test.go",
		"arm64_private_sp_helper_test.go",
		"preprocess_macro_controls_test.go",
		"preprocess_semantic_oracle_test.go",
		"x86_official_binary_witness_test.go",
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
				if function, ok := call.Fun.(*ast.Ident); ok {
					switch function.Name {
					case "runtimeOracleCombinedOutput", "goAssemblerOracleOutput", "x86FloatCompareCaptureCommand":
						consumers++
						return true
					}
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
