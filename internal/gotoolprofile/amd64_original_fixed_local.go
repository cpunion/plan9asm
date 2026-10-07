package gotoolprofile

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/xgo-dev/plan9asm"
)

// AMD64OriginalFixedLocalBinding owns a separate actual Go object capability.
// The query-only observation does not become a frame certificate: the existing
// root factory and binder independently check every source/declaration/object
// and all native initialization, address, CFG and relocation requirements.
// This first handoff deliberately retains that factory's numeric, flat,
// owned-main scope. SDK imports are independently source/export-bound by that
// factory; foreign imports, DATA, composite and external roles cannot obtain a
// capability through this method.
type AMD64OriginalFixedLocalBinding struct {
	mu         sync.Mutex
	owner      *AMD64OriginalPackageObservation
	generation uint64
	source     *plan9asm.GoAMD64FrameSource
	closed     bool
}

func (observed *AMD64OriginalPackageObservation) BindFixedLocal(ctx context.Context, pkg plan9asm.GoPackage,
	file *plan9asm.File, assembly string, options plan9asm.Options,
) (_ plan9asm.Options, _ *AMD64OriginalFixedLocalBinding, result error) {
	if err := observed.Verify(ctx); err != nil {
		return plan9asm.Options{}, nil, err
	}
	observed.mu.Lock()
	request, generation, ready := observed.request, observed.generation, observed.ready
	observed.mu.Unlock()
	if !ready || pkg.Path != request.PackagePath {
		return plan9asm.Options{}, nil, fmt.Errorf("original fixed-local package identity differs")
	}
	source, err := plan9asm.CaptureGoAMD64FrameSource(ctx, pkg, plan9asm.GoAMD64FrameCaptureRequest{
		GoBinary: request.GoBinary, Directory: request.Directory, AssemblyFile: assembly,
		Target: request.Target, Environment: append([]string(nil), request.Environment...),
	})
	if err != nil {
		return plan9asm.Options{}, nil, err
	}
	// Neither a resolver nor any other external command runs under owner.mu.
	// A callback closing this observation must revoke the eventual handoff.
	defer func() {
		if result != nil {
			if err := source.Close(); err != nil {
				result = errors.Join(result, err)
			}
		}
	}()
	bound, err := source.BindFixedLocal(ctx, pkg, file, options)
	if err != nil {
		return plan9asm.Options{}, nil, err
	}
	if err := observed.Verify(ctx); err != nil {
		return plan9asm.Options{}, nil, err
	}
	lease := &AMD64OriginalFixedLocalBinding{owner: observed, generation: generation, source: source}
	observed.mu.Lock()
	defer observed.mu.Unlock()
	if !observed.ready || observed.generation != generation {
		return plan9asm.Options{}, nil, fmt.Errorf("original observation was revoked during actual frame capture")
	}
	observed.fixedLocal = append(observed.fixedLocal, lease)
	return bound, lease, nil
}

// Verify closes over both independent observations. A failed owner Verify
// permanently revokes every attached actual-object capability as well.
func (binding *AMD64OriginalFixedLocalBinding) Verify(ctx context.Context) error {
	if binding == nil {
		return fmt.Errorf("missing original fixed-local binding")
	}
	binding.mu.Lock()
	owner, source, generation, closed := binding.owner, binding.source, binding.generation, binding.closed
	binding.mu.Unlock()
	if closed || owner == nil || source == nil {
		return fmt.Errorf("original fixed-local binding is closed")
	}
	if err := owner.Verify(ctx); err != nil {
		return err
	}
	owner.mu.Lock()
	ready := owner.ready && owner.generation == generation
	owner.mu.Unlock()
	if !ready {
		return fmt.Errorf("original fixed-local owner generation changed")
	}
	return source.Verify(ctx)
}

func (binding *AMD64OriginalFixedLocalBinding) Close() error {
	if binding == nil {
		return nil
	}
	binding.mu.Lock()
	binding.closed = true
	source := binding.source
	binding.mu.Unlock()
	if source == nil {
		return nil
	}
	return source.Close()
}

func closeAMD64OriginalFixedLocal(bindings []*AMD64OriginalFixedLocalBinding) error {
	var result error
	for _, binding := range bindings {
		if err := binding.Close(); err != nil {
			result = errors.Join(result, err)
		}
	}
	return result
}
