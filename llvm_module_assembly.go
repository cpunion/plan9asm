package plan9asm

/*
#include <stddef.h>
typedef struct LLVMOpaqueModule *LLVMModuleRef;
extern const char *LLVMGetModuleInlineAsm(LLVMModuleRef M, size_t *Len);
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
