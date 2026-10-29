package plan9asm

/*
#include <stddef.h>
#include <stdlib.h>
typedef struct LLVMOpaqueModule *LLVMModuleRef;
typedef struct LLVMOpaqueValue *LLVMValueRef;
typedef struct LLVMOpaqueType *LLVMTypeRef;
typedef struct LLVMOpaqueContext *LLVMContextRef;
extern const char *LLVMGetModuleInlineAsm(LLVMModuleRef M, size_t *Len);
extern LLVMValueRef LLVMIsAAtomicCmpXchgInst(LLVMValueRef Val);
extern LLVMValueRef LLVMIsAAtomicRMWInst(LLVMValueRef Val);
extern unsigned LLVMGetNumArgOperands(LLVMValueRef Instr);
extern LLVMTypeRef LLVMInt8TypeInContext(LLVMContextRef C);
extern LLVMTypeRef LLVMScalableVectorType(LLVMTypeRef ElementType, unsigned ElementCount);
extern int LLVMGetTypeKind(LLVMTypeRef Ty);
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

// The pinned Go binding omits these LLVM22 instruction-class queries.
func llvmInstructionIsAtomicMemory(instruction llvm.Value) bool {
	value := C.LLVMValueRef(unsafe.Pointer(instruction.C))
	return C.LLVMIsAAtomicCmpXchgInst(value) != nil || C.LLVMIsAAtomicRMWInst(value) != nil
}

func llvmCallArgumentCount(instruction llvm.Value) int {
	return int(C.LLVMGetNumArgOperands(C.LLVMValueRef(unsafe.Pointer(instruction.C))))
}

// Obtain the missing binding enum from an actual LLVM22 scalable type. The
// existing module context owns this uniqued type; no context is allocated or
// retained. This avoids a copied enum and platform-specific header search paths.
func llvmScalableVectorKind(context llvm.Context) llvm.TypeKind {
	owner := C.LLVMContextRef(unsafe.Pointer(context.C))
	typ := C.LLVMScalableVectorType(C.LLVMInt8TypeInContext(owner), 1)
	return llvm.TypeKind(C.LLVMGetTypeKind(typ))
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
