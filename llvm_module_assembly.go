package plan9asm

/*
#include <stddef.h>
#include <stdlib.h>
typedef struct LLVMOpaqueModule *LLVMModuleRef;
extern const char *LLVMGetModuleInlineAsm(LLVMModuleRef M, size_t *Len);
extern char *LLVMNormalizeTargetTriple(const char *Triple);
extern void LLVMDisposeMessage(char *Message);
*/
import "C"

import (
	"unsafe"

	"github.com/xgo-dev/llvm"
)

// LLVM22 llvm-c/Core.h exposes LLVMGetModuleInlineAsm, but llvm v0.10.1
// only binds its setter. Replace this length-only bridge with the official
// getter when available; no assembly text needs copying or interpretation.
func llvmModuleHasAssembly(module llvm.Module) bool {
	var length C.size_t
	C.LLVMGetModuleInlineAsm(C.LLVMModuleRef(unsafe.Pointer(module.C)), &length)
	return length != 0
}

// LLVM22 exposes this normalizer in llvm-c/TargetMachine.h. The official
// llvm v0.10.1 binding does not expose it, so keep ownership in this small
// C API bridge until that binding supplies the query.
func llvmNormalizedTargetTriple(triple string) string {
	if triple == "" {
		return ""
	}
	input := C.CString(triple)
	defer C.free(unsafe.Pointer(input))
	normalized := C.LLVMNormalizeTargetTriple(input)
	defer C.LLVMDisposeMessage(normalized)
	return C.GoString(normalized)
}
