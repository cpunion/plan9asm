package plan9asm

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// These are pure digest fixtures, not actual SDK captures or authority tokens.
// None calls Verify, invokes a tool, supplies an initializer or grants memory.
// The independent state mirrors the complete existing digest input, not just
// selected large blobs, so an optimization cannot silently omit small fields.
type sdkASMDataDigestTestState struct {
	Request           GoX86SDKASMDataCaptureRequest
	Guard             GoAMD64FrameCaptureRequest
	Selection         goAMD64FrameSelection
	Env               map[string]string
	Imports           map[string]goAMD64FrameSelection
	Inputs            map[string][32]byte
	Entries           []string
	Misses            map[string]bool
	Observation       sdkASMDataLiveObservation
	OriginalDeadline  time.Time
	OriginalProfile   bool
	CaptureRole       sdkASMDataCaptureRole
	OriginalConfig    string
	FrameListingBytes int
}

func (state *sdkASMDataDigestTestState) publish(source *GoX86SDKASMDataSource) {
	source.request = state.Request
	source.guard.request = state.Guard
	source.guard.selection = state.Selection
	source.guard.env = state.Env
	source.guard.imports = state.Imports
	source.guard.inputs = state.Inputs
	source.guard.entries = state.Entries
	source.misses = state.Misses
	source.observation = state.Observation
	source.originalDeadline = state.OriginalDeadline
	source.originalProfile = state.OriginalProfile
	source.captureRole = state.CaptureRole
	source.originalConfig = state.OriginalConfig
	source.frameListingBytes = state.FrameListingBytes
}

func sdkASMDataDigestSeed(t *testing.T, typ reflect.Type, depth int) reflect.Value {
	t.Helper()
	if depth > 32 {
		t.Fatal("digest fixture schema unexpectedly recurses")
	}
	value := reflect.New(typ).Elem()
	if typ == reflect.TypeOf(time.Time{}) {
		value.Set(reflect.ValueOf(time.Date(2030, 1, 2, 3, 4, 5, 6, time.UTC)))
		return value
	}
	if typ == reflect.TypeOf(json.RawMessage{}) {
		value.SetBytes([]byte("{\"fixture\":1}"))
		return value
	}
	if typ == reflect.TypeOf((*GoX86NumericalReadBinding)(nil)) {
		return value // The input guard must not borrow a numerical capability.
	}
	switch typ.Kind() {
	case reflect.String:
		value.SetString("seed")
	case reflect.Bool:
		value.SetBool(true)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value.SetInt(1)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value.SetUint(1)
	case reflect.Struct:
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			if field.PkgPath == "" && strings.Split(field.Tag.Get("json"), ",")[0] != "-" {
				value.Field(i).Set(sdkASMDataDigestSeed(t, field.Type, depth+1))
			}
		}
	case reflect.Ptr:
		value.Set(reflect.New(typ.Elem()))
		value.Elem().Set(sdkASMDataDigestSeed(t, typ.Elem(), depth+1))
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			value.SetBytes([]byte{0, 1, 2, 254, 255})
			break
		}
		value.Set(reflect.MakeSlice(typ, 1, 1))
		value.Index(0).Set(sdkASMDataDigestSeed(t, typ.Elem(), depth+1))
	case reflect.Array:
		for i := 0; i < typ.Len(); i++ {
			value.Index(i).Set(sdkASMDataDigestSeed(t, typ.Elem(), depth+1))
		}
	case reflect.Map:
		if typ.Key().Kind() != reflect.String {
			t.Fatal("unexpected non-string observation map key", typ)
		}
		value.Set(reflect.MakeMap(typ))
		for _, name := range []string{"first", "second"} {
			key := reflect.New(typ.Key()).Elem()
			key.SetString(name)
			value.SetMapIndex(key, sdkASMDataDigestSeed(t, typ.Elem(), depth+1))
		}
	default:
		t.Fatal("unexpected digest fixture field type", typ)
	}
	return value
}

func sdkASMDataDigestFixture(t *testing.T) (*sdkASMDataDigestTestState, *GoX86SDKASMDataSource) {
	t.Helper()
	state := sdkASMDataDigestSeed(t, reflect.TypeOf(sdkASMDataDigestTestState{}), 0).
		Interface().(sdkASMDataDigestTestState)
	source := &GoX86SDKASMDataSource{guard: newGoOriginalInputGuard(state.Guard)}
	state.publish(source)
	return &state, source
}

func TestGoX86SDKASMDataDigestCompleteMutableState(t *testing.T) {
	state, source := sdkASMDataDigestFixture(t)
	original, err := json.Marshal(state)
	if err != nil {
		t.Fatal("independent complete fixture is not JSON-representable", err)
	}
	want, err := source.digest()
	if err != nil {
		t.Fatal("initial complete state digest failed", err)
	}
	var document any
	if err := json.Unmarshal(original, &document); err != nil {
		t.Fatal(err)
	}
	mutations := 0
	check := func(name string, replace func(any), before any, alternatives []any) {
		t.Run(name, func(t *testing.T) {
			defer replace(before)
			var changed sdkASMDataDigestTestState
			found := false
			for _, after := range alternatives {
				replace(after)
				raw, encodeErr := json.Marshal(document)
				if encodeErr != nil {
					t.Fatal(encodeErr)
				}
				changed = sdkASMDataDigestTestState{}
				if json.Unmarshal(raw, &changed) != nil {
					continue // Try the actual nullable field's independent JSON kind.
				}
				complete, encodeErr := json.Marshal(changed)
				if encodeErr != nil {
					t.Fatal(encodeErr)
				}
				if !bytes.Equal(complete, original) {
					found = true
					break
				}
			}
			if !found {
				t.Fatal("mutation did not change a complete typed state field")
			}
			changed.publish(source) // Keep the exact same source object: no cached digest.
			got, digestErr := source.digest()
			if digestErr != nil || got == want {
				t.Fatal("complete mutable state was omitted or reused", digestErr)
			}
			state.publish(source)
			restored, digestErr := source.digest()
			if digestErr != nil || restored != want {
				t.Fatal("pure digest is not deterministic after restoration", digestErr)
			}
			mutations++
		})
	}
	var visit func(string, any, func(any))
	visit = func(name string, value any, replace func(any)) {
		switch typed := value.(type) {
		case map[string]any:
			check(name+"/empty-object", replace, value, []any{map[string]any{}})
			keys := make([]string, 0, len(typed))
			for key := range typed {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				key := key
				visit(name+"/"+key, typed[key], func(next any) { typed[key] = next })
			}
		case []any:
			check(name+"/empty-array", replace, value, []any{[]any{}})
			for i, item := range typed {
				i := i
				visit(name+"/"+strconv.Itoa(i), item, func(next any) { typed[i] = next })
			}
		case string:
			next := "QUJD" // Valid base64 for byte fields as well as a new string value.
			if typed == next {
				next = "REVG"
			}
			if stamp, parseErr := time.Parse(time.RFC3339Nano, typed); parseErr == nil {
				next = stamp.Add(time.Second).Format(time.RFC3339Nano)
			}
			check(name, replace, value, []any{next})
		case float64:
			check(name, replace, value, []any{typed + 1})
		case bool:
			check(name, replace, value, []any{!typed})
		case nil:
			check(name+"/non-null", replace, value,
				[]any{"QUJD", []any{}, map[string]any{}, float64(1), true})
		default:
			t.Fatal("unexpected independent JSON field", name)
		}
	}
	visit("state", document, func(next any) { document = next })
	if mutations < 100 {
		t.Fatal("complete observation/schema mutation matrix was not traversed", mutations)
	}
}

func TestGoX86SDKASMDataDigestLargePayloadTemporaryBytes(t *testing.T) {
	state, source := sdkASMDataDigestFixture(t)
	payload := bytes.Repeat([]byte{0x97, 0x31, 0x00, 0xff}, 1<<19) // One existing 2 MiB compiler-object role.
	state.Observation.Header.CompilerObject = payload
	state.publish(source)
	runtime.GC()
	want, err := source.digest() // Warm reusable encoder/buffer state outside the measurement.
	if err != nil {
		t.Fatal(err)
	}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	var sums [3][32]byte
	for i := range sums {
		sums[i], err = source.digest()
		if err != nil {
			break
		}
	}
	runtime.ReadMemStats(&after)
	runtime.KeepAlive(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, sum := range sums {
		if sum != want {
			t.Fatal("unchanged complete state has a nondeterministic digest")
		}
	}
	// This is an allocation-size regression, not a throughput benchmark or an
	// execution certificate. Every call must still inspect the current bytes.
	// Three MiB of total slack allows bounded metadata work and runtime noise,
	// but not three fresh base64/JSON copies of this single 2 MiB byte field.
	allocated := after.TotalAlloc - before.TotalAlloc
	if allocated > 3<<20 {
		t.Errorf("three complete digests allocate %d temporary bytes; want <= %d without payload-sized JSON copies",
			allocated, uint64(3<<20))
	}
	for _, index := range []int{0, len(payload) / 2, len(payload) - 1} {
		payload[index] ^= 0x80
		got, digestErr := source.digest()
		if digestErr != nil || got == want {
			t.Fatal("in-place current payload mutation escaped the digest", index, digestErr)
		}
		payload[index] ^= 0x80
	}
}

func TestGoX86SDKASMDataDigestOptimizationRetainsEveryLiveBoundary(t *testing.T) {
	// This batch optimizes only digest(), never live verification. Pin the
	// existing complete methods rather than merely counting a surviving hash
	// call that could have moved outside its command or generation boundary.
	guards := []struct {
		file, receiver, name, want string
	}{
		{file: "go_x86_sdk_asm_data_source.go", receiver: "", name: "captureGoX86SDKASMDataSourceRole",
			want: "7532bb00ebabdf1775a2d4bfb27c33e0ec2db5a6a7d29b0a97c73e809e025a3c"},
		{file: "go_x86_sdk_asm_data_source.go", receiver: "GoX86SDKASMDataSource", name: "liveContext",
			want: "7ceb17703c2f84411a492ed8d1b836aef6eb61c879dd7390f9edd8f66cec024b"},
		{file: "go_x86_sdk_asm_data_source.go", receiver: "GoX86SDKASMDataSource", name: "run",
			want: "77ac546bcae1e718c183a73263bd9f368baf93583064fd9571f53cfd592d5e75"},
		{file: "go_x86_sdk_asm_data_source.go", receiver: "GoX86SDKASMDataSource", name: "runCommand",
			want: "475339cf13e66ea1cd1da8ae4cac516d6ea47fbe5d0fb9bbbf91d5ca4b66a1d1"},
		{file: "go_x86_sdk_asm_data_source.go", receiver: "GoX86SDKASMDataSource", name: "Verify",
			want: "582072eefb9314f900f15bacd0dfbc653669c3fbdab19ed433f1981d8d179183"},
		{file: "go_x86_sdk_asm_data_source.go", receiver: "GoX86SDKASMDataSource", name: "verifyState",
			want: "427c7289796f431085ac5d2c574f30945068908dbea9564bdeb7e52094ab7ad7"},
		{file: "go_x86_sdk_asm_data_source.go", receiver: "GoX86SDKASMDataSource", name: "Close",
			want: "a6a752ce12f32861b977e0dd0e0c0e6d2ec8598e989ed74d2031763d865f8aa9"},
		{file: "go_x86_sdk_asm_data_source.go", receiver: "GoX86SDKASMDataSource", name: "Readback",
			want: "f0044ed9e953dca9f4580f9d51796787d734aee41f2268a68864bd2c975f45fa"},
		{file: "go_x86_sdk_asm_data_source.go", receiver: "GoX86SDKASMDataSource", name: "CheckReadback",
			want: "d87afa790aa89fb6530dbd226732cace6723a2369e69bcd98516369499f14892"},
		{file: "amd64_go_frame_source.go", receiver: "GoAMD64FrameSource", name: "verifyInputs",
			want: "03e42996d6837fc585e6b001afc1b3fa9d6397ce179e5bc5aea36d664bffbc2e"},
		{file: "internal/gooriginalsrc/digest.go", receiver: "", name: "HashRegularFile",
			want: "fc81069f186b180c0f9967b0af9837240d47d386e187c90b76def8ed424ab74d"},
	}
	for _, guard := range guards {
		t.Run(guard.name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.FromSlash(guard.file))
			if err != nil {
				t.Fatal(err)
			}
			// Repository Go checkout EOL is separate from external source authority.
			raw = bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))
			positions := token.NewFileSet()
			parsed, err := parser.ParseFile(positions, guard.file, raw, 0)
			if err != nil {
				t.Fatal(err)
			}
			matches := 0
			for _, declaration := range parsed.Decls {
				fn, ok := declaration.(*ast.FuncDecl)
				if !ok || fn.Name.Name != guard.name {
					continue
				}
				receiver := ""
				if fn.Recv != nil && len(fn.Recv.List) == 1 {
					if pointer, ok := fn.Recv.List[0].Type.(*ast.StarExpr); ok {
						if named, ok := pointer.X.(*ast.Ident); ok {
							receiver = named.Name
						}
					}
				}
				if receiver != guard.receiver {
					continue
				}
				first := positions.Position(fn.Pos()).Offset
				last := positions.Position(fn.End()).Offset
				sum := sha256.Sum256(raw[first:last])
				if hex.EncodeToString(sum[:]) != guard.want {
					t.Fatal("digest-only optimization changed an existing live verification boundary")
				}
				matches++
			}
			if matches != 1 {
				t.Fatal("live verification boundary is missing or duplicated", matches)
			}
		})
	}
}
