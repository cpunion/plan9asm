//go:build go1.27

package plan9asm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
)

// These are two different actual SDK packages, not library-name allowlists.
// Complete source/header/CPP/object observation must close before the new
// initializer factory is asked to establish its independent role.
func TestGoX86SDKReadonlyLiteralDataActualSources(t *testing.T) {
	for _, pkg := range []string{"hash/crc32", "internal/chacha8rand"} {
		t.Run(pkg, func(t *testing.T) {
			ctx, request := sdkASMDataRequest(t, "amd64", pkg)
			source, err := CaptureGoX86SDKASMDataSource(ctx, request)
			if err != nil {
				t.Fatal("actual complete SDK observation precondition", err)
			}
			t.Cleanup(func() { source.Close() })
			sdkASMDataRetain(t, source)
			owner, err := CaptureGoX86SDKReadonlyLiteralDataOwner(ctx, source)
			if err != nil || owner == nil {
				t.Fatal("actual SDK private readonly literal initializer is not certified", err)
			}
			t.Cleanup(func() { owner.Close() })
			raw, err := owner.Readback(ctx)
			if err != nil || owner.CheckReadback(ctx, raw) != nil {
				t.Fatal("live exact initializer readback is not closed", err)
			}
			var report sdkReadonlyLiteralDataReport
			if json.Unmarshal(raw, &report) != nil || report.Protocol != sdkReadonlyLiteralDataProtocol ||
				report.Target != request.Target || report.PackagePath != pkg || report.SourceReportSHA256 != source.seal ||
				report.SourceGeneration != source.generation || len(report.Initializers) == 0 {
				t.Fatal("initializer omitted the whole actual SDK source identity")
			}
			count := 0
			for _, observed := range source.observation.Objects.Sources {
				count += len(observed.ExpandedLayouts)
			}
			if len(report.Initializers) != count {
				t.Fatal("initializer omitted selected source private DATA")
			}
			for _, image := range report.Initializers {
				member := source.observation.Objects.Members[image.MemberOrdinal]
				found := false
				for _, definition := range member.Definitions {
					if definition.Namespace == image.Definition.Namespace && definition.Index == image.Definition.Index {
						found = true
						if !reflect.DeepEqual(definition, image.Definition) || int(definition.Size) != len(image.Initial) ||
							!bytes.Equal(definition.Data, image.Initial[:len(definition.Data)]) {
							t.Fatal("initializer selected another source/object-local definition")
						}
						for _, value := range image.Initial[len(definition.Data):] {
							if value != 0 {
								t.Fatal("initializer tail is not exact Go zero padding")
							}
						}
					}
				}
				if !found {
					t.Fatal("initializer object identity is absent from the whole archive")
				}
			}
			for _, name := range []string{"Bind", "BindNumericalReads", "Options", "Pointer", "Read", "Variable", "Owners", "GoPackage"} {
				if _, present := reflect.TypeOf(owner).MethodByName(name); present {
					t.Fatal("literal initialization borrowed another authority", name)
				}
			}
			if folder := os.Getenv("PLAN9ASM_STATIC_STORAGE_ARTIFACT_DIR"); folder != "" {
				name := strings.ReplaceAll(t.Name(), "/", "-")
				if err := os.WriteFile(filepath.Join(folder, name+".readonly-literal.json"), raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func readonlyLiteralCapture(t *testing.T) (context.Context, *GoX86SDKASMDataSource, *GoX86SDKReadonlyLiteralDataOwner) {
	t.Helper()
	ctx, request := sdkASMDataRequest(t, "amd64", "hash/crc32")
	source, err := CaptureGoX86SDKASMDataSource(ctx, request)
	if err != nil {
		t.Fatal("actual SDK source precondition", err)
	}
	t.Cleanup(func() { source.Close() })
	owner, err := CaptureGoX86SDKReadonlyLiteralDataOwner(ctx, source)
	if err != nil {
		t.Fatal("actual literal owner precondition", err)
	}
	t.Cleanup(func() { owner.Close() })
	return ctx, source, owner
}

func readonlyLiteralInvoke(t *testing.T, api string, owner *GoX86SDKReadonlyLiteralDataOwner, ctx context.Context, raw []byte) ([]byte, error) {
	t.Helper()
	switch api {
	case "Verify":
		return nil, owner.Verify(ctx)
	case "Readback":
		return owner.Readback(ctx)
	case "CheckReadback":
		return nil, owner.CheckReadback(ctx, raw)
	}
	t.Fatal("unknown initializer API")
	return nil, nil
}

func TestGoX86SDKReadonlyLiteralDataCopyCloseAndBytes(t *testing.T) {
	ctx, source, owner := readonlyLiteralCapture(t)
	copy := reflect.New(reflect.TypeOf(owner).Elem())
	copy.Elem().Set(reflect.ValueOf(owner).Elem())
	copied := copy.Interface().(*GoX86SDKReadonlyLiteralDataOwner)
	if !errors.Is(copied.Verify(ctx), ErrProbeNeedsContext) || !errors.Is(copied.Close(), ErrProbeNeedsContext) {
		t.Fatal("copied initializer borrowed or revoked its original token")
	}
	if err := owner.Verify(ctx); err != nil {
		t.Fatal("copied initializer affected its original owner", err)
	}
	raw, err := owner.Readback(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if owner.CheckReadback(ctx, append(raw, ' ')) == nil || !errors.Is(owner.Verify(ctx), ErrProbeNeedsContext) {
		t.Fatal("public bytes recreated an initializer or restoration revived it")
	}
	if err := source.Verify(ctx); err != nil {
		t.Fatal("initializer byte mismatch revoked its independently owned source", err)
	}
	if err := owner.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(source.guard.owned); err != nil {
		t.Fatal("initializer cleanup deleted the source artifacts", err)
	}
	if source.guard.ready || source.guard.request.NumericalReads != nil {
		t.Fatal("literal initializer granted frame or numerical-read permission")
	}
}

func TestGoX86SDKReadonlyLiteralDataSourceAndOwnerMutations(t *testing.T) {
	for name, mutate := range map[string]func(*GoX86SDKASMDataSource, *GoX86SDKReadonlyLiteralDataOwner){
		"source-close":            func(s *GoX86SDKASMDataSource, _ *GoX86SDKReadonlyLiteralDataOwner) { s.Close() },
		"owner-close":             func(_ *GoX86SDKASMDataSource, o *GoX86SDKReadonlyLiteralDataOwner) { o.Close() },
		"source-generation":       func(s *GoX86SDKASMDataSource, _ *GoX86SDKReadonlyLiteralDataOwner) { s.generation++ },
		"owner-source-generation": func(_ *GoX86SDKASMDataSource, o *GoX86SDKReadonlyLiteralDataOwner) { o.sourceGeneration++ },
		"source-bytes":            func(s *GoX86SDKASMDataSource, _ *GoX86SDKReadonlyLiteralDataOwner) { s.report[0] ^= 1 },
		"object-index": func(s *GoX86SDKASMDataSource, _ *GoX86SDKReadonlyLiteralDataOwner) {
			s.observation.Objects.Members[2].Definitions[0].Index++
		},
		"object-data": func(s *GoX86SDKASMDataSource, _ *GoX86SDKReadonlyLiteralDataOwner) {
			s.observation.Objects.Members[2].Definitions[0].Data[0] ^= 1
		},
		"original-ASM": func(s *GoX86SDKASMDataSource, _ *GoX86SDKReadonlyLiteralDataOwner) {
			s.observation.ObjectInputs.Sources[0].Original[0] ^= 1
		},
		"header": func(s *GoX86SDKASMDataSource, _ *GoX86SDKReadonlyLiteralDataOwner) {
			s.observation.Header.Actual[0] ^= 1
		},
		"source-selection": func(s *GoX86SDKASMDataSource, _ *GoX86SDKReadonlyLiteralDataOwner) { s.guard.selection.SFiles = nil },
		"archive": func(s *GoX86SDKASMDataSource, _ *GoX86SDKReadonlyLiteralDataOwner) {
			s.observation.ObjectInputs.Archive[0] ^= 1
		},
		"foreign-source": func(_ *GoX86SDKASMDataSource, o *GoX86SDKReadonlyLiteralDataOwner) {
			o.source = &GoX86SDKASMDataSource{}
		},
		"owner-bytes":   func(_ *GoX86SDKASMDataSource, o *GoX86SDKReadonlyLiteralDataOwner) { o.report[0] ^= 1 },
		"owner-context": func(_ *GoX86SDKASMDataSource, o *GoX86SDKReadonlyLiteralDataOwner) { o.context = context.Background() },
	} {
		t.Run(name, func(t *testing.T) {
			ctx, source, owner := readonlyLiteralCapture(t)
			mutate(source, owner)
			if err := owner.Verify(ctx); !errors.Is(err, ErrProbeNeedsContext) {
				t.Fatal("changed whole source or exact initializer verified", err)
			}
			if owner.ready || !errors.Is(owner.Verify(ctx), ErrProbeNeedsContext) {
				t.Fatal("revoked literal initializer revived")
			}
		})
	}
}

func TestGoX86SDKReadonlyLiteralDataContextReentry(t *testing.T) {
	for _, route := range []string{"original", "verification"} {
		for _, method := range []string{"Err", "Deadline", "Done"} {
			for _, api := range []string{"Verify", "Readback", "CheckReadback", "Close"} {
				t.Run(route+"/"+method+"/"+api, func(t *testing.T) {
					ctx, source, _ := readonlyLiteralCapture(t)
					callback := &sdkASMDataReentrantContext{Context: ctx, method: method}
					captureContext := ctx
					if route == "original" {
						captureContext = callback
					}
					owner, err := CaptureGoX86SDKReadonlyLiteralDataOwner(captureContext, source)
					if err != nil {
						t.Fatal(err)
					}
					defer owner.Close()
					raw, err := owner.Readback(ctx)
					if err != nil {
						t.Fatal(err)
					}
					var nested error
					callback.callback = func() {
						if api == "Close" {
							nested = owner.Close()
						} else {
							_, nested = readonlyLiteralInvoke(t, api, owner, ctx, raw)
						}
					}
					callback.armed.Store(true)
					finished := make(chan error, 1)
					go func() { finished <- errors.Join(owner.Verify(callback), nested) }()
					select {
					case err := <-finished:
						if api == "Close" {
							if !errors.Is(err, ErrProbeNeedsContext) || owner.ready {
								t.Fatal("reentrant Close published a revoked initializer", err)
							}
						} else if err != nil {
							t.Fatal("initializer validation callback deadlocked or rejected its actual owner", err)
						}
					case <-time.After(3 * time.Second):
						t.Fatal("initializer API synchronously deadlocked during context callback")
					}
					if err := source.Verify(ctx); err != nil {
						t.Fatal("initializer callback disposed its source", err)
					}
				})
			}
		}
	}
}

func TestGoX86SDKReadonlyLiteralDataOriginalDeadlineAndCancellation(t *testing.T) {
	for _, change := range []string{"earlier", "later", "lost", "cancel", "expired"} {
		for _, api := range []string{"Verify", "Readback", "CheckReadback"} {
			t.Run(change+"/"+api, func(t *testing.T) {
				ctx, source, _ := readonlyLiteralCapture(t)
				original, cancel := context.WithCancel(ctx)
				defer cancel()
				mutable := &sdkASMDataReentrantContext{Context: original}
				owner, err := CaptureGoX86SDKReadonlyLiteralDataOwner(mutable, source)
				if err != nil {
					t.Fatal(err)
				}
				defer owner.Close()
				raw, err := owner.Readback(ctx)
				if err != nil {
					t.Fatal(err)
				}
				want := ErrProbeNeedsContext
				verification := ctx
				switch change {
				case "earlier":
					mutable.shift.Store(-int64(time.Minute))
				case "later":
					mutable.shift.Store(int64(time.Minute))
				case "lost":
					mutable.lost.Store(true)
				case "cancel":
					cancel()
					want = context.Canceled
				case "expired":
					verification, cancel = context.WithDeadline(ctx, time.Now().Add(-time.Second))
					defer cancel()
					want = context.DeadlineExceeded
				}
				returned, err := readonlyLiteralInvoke(t, api, owner, verification, raw)
				if !errors.Is(err, want) || len(returned) != 0 {
					t.Fatal("changed lifecycle published initialized bytes or lost actual cause", err)
				}
				mutable.shift.Store(0)
				mutable.lost.Store(false)
				if !errors.Is(owner.Verify(ctx), ErrProbeNeedsContext) {
					t.Fatal("restoring the original deadline revived the initializer")
				}
			})
		}
	}
}

func TestGoX86SDKReadonlyLiteralDataConcurrentClose(t *testing.T) {
	ctx, source, owner := readonlyLiteralCapture(t)
	var workers sync.WaitGroup
	for index := 0; index < 4; index++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			if err := owner.Verify(ctx); err != nil && !errors.Is(err, ErrProbeNeedsContext) {
				t.Error(err)
			}
		}()
	}
	if err := owner.Close(); err != nil {
		t.Fatal(err)
	}
	workers.Wait()
	if !errors.Is(owner.Verify(ctx), ErrProbeNeedsContext) || source.Verify(ctx) != nil {
		t.Fatal("concurrent initializer close revived it or deleted independent source")
	}
}

func TestGoX86SDKReadonlyLiteralDataNoJSONOrOtherRole(t *testing.T) {
	ctx, source := sdkASMDataCapture(t)
	owner, err := CaptureGoX86SDKReadonlyLiteralDataOwner(ctx, source)
	if owner != nil || !errors.Is(err, ErrProbeNeedsContext) {
		t.Fatal("SDK source without literal DATA acquired initializer permission", err)
	}
	var zero GoX86SDKReadonlyLiteralDataOwner
	if !errors.Is(zero.Verify(ctx), ErrProbeNeedsContext) || !errors.Is(zero.Close(), ErrProbeNeedsContext) {
		t.Fatal("zero public token recreated a private initializer owner")
	}
	for _, input := range []*GoX86SDKASMDataSource{nil, {}} {
		if owner, err := CaptureGoX86SDKReadonlyLiteralDataOwner(ctx, input); owner != nil || !errors.Is(err, ErrProbeNeedsContext) {
			t.Fatal("unbound source or public JSON could construct a live literal owner", err)
		}
	}
	// Go storage types are still observations, not substitutes for this token.
	if _, present := reflect.TypeOf((*gostaticstorage.Variable)(nil)).MethodByName("Initializers"); present {
		t.Fatal("literal initialization replaced the original Go Variable role")
	}
}
