//go:build go1.27

package plan9asm

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/goasmcmd"
	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
)

func TestSDKAMD64NativeFrameKindsActualGoObjects(t *testing.T) {
	_, profile, _, _ := numericalSBFixture(t, "amd64")
	ctx, cancel := runtimeOracleProofContext(t)
	defer cancel()
	root := filepath.Dir(filepath.Dir(profile.GoBinary))
	builtin, err := os.ReadFile(filepath.Join(root, "src/cmd/internal/goobj/builtinlist.go"))
	if err != nil {
		t.Fatal(err)
	}
	// The encoder chooses the kinds from ordinary source symbols. This is a
	// standalone compiler-format fixture, never an SDK provenance certificate.
	source := []byte(`TEXT crypto∕internal∕fips140∕framefixture·First(SB),4,$16-8
	MOVQ $11, AX
	RET
TEXT ·Ordinary(SB),4,$0-0
	RET
TEXT private<>(SB),4,$8
	RET
TEXT crypto∕internal∕fips140∕framefixture·Last(SB),4,$24-16
	RET
`)
	path := filepath.Join(profile.Directory, "mixed.asm")
	if err := os.WriteFile(path, source, 0600); err != nil {
		t.Fatal(err)
	}
	for _, goos := range []string{"linux", "darwin", "windows"} {
		t.Run(goos, func(t *testing.T) {
			env := append([]string(nil), profile.Environment...)
			for index, pair := range env {
				if strings.HasPrefix(pair, "GOOS=") {
					env[index] = "GOOS=" + goos
				}
			}
			objectPath := filepath.Join(t.TempDir(), "mixed.o")
			listing, err := goasmcmd.Run(ctx, profile.Directory, env, profile.GoBinary,
				"tool", "asm", "-p", "framefixture", "-S", "-o", objectPath, path)
			if err != nil {
				t.Fatal("actual Go assembler", err)
			}
			object, err := os.ReadFile(objectPath)
			if err != nil {
				t.Fatal(err)
			}
			observed, err := gostaticstorage.ObserveAMD64AssemblyData(source, object, builtin, "framefixture", goos+"/amd64")
			if err != nil || len(observed.Member.Texts) != 4 || len(observed.TextOrder) != 4 {
				t.Fatal("complete actual source/object precondition", err)
			}
			kinds := []uint8{2, 1, 1, 2}
			for ordinal, entry := range observed.TextOrder {
				if observed.Member.Texts[entry.ObjectOrdinal].Kind != kinds[ordinal] {
					t.Fatal("actual Go object did not emit the independent mixed-kind prior", ordinal)
				}
			}
			if bytes.Count(listing, []byte(" STEXTFIPS ")) != 2 || bytes.Count(listing, []byte(" STEXT ")) != 2 {
				t.Fatal("actual assembler listing did not preserve both TEXT kinds")
			}
			records, err := sdkAMD64FrameNativeRecordsForSource(listing, observed.Member.Texts, observed.TextOrder)
			if err != nil || len(records) != 4 {
				t.Fatal("complete actual mixed-kind listing lost native frames", err)
			}
			if records[0].Locals != 24 || records[0].Args != 8 || records[1].Locals != 0 ||
				!records[2].Static || records[2].ArgsKnown || records[2].Locals != 16 || records[3].Locals != 32 || records[3].Args != 16 {
				t.Fatal("actual ordinary/FIPS/static frame metadata changed", records)
			}
		})
	}
}

func TestSDKAMD64NativeFrameFIPSCompleteOriginalSource(t *testing.T) {
	// Share only the stock compiler cache. Each source capture retains its own
	// complete inputs, command outputs, verification and revocable lifetime.
	t.Setenv("GOCACHE", t.TempDir())
	for _, pkg := range []string{
		"crypto/internal/fips140/aes/gcm",
		"crypto/internal/fips140/sha256",
		"crypto/internal/fips140/sha512",
		"crypto/internal/fips140/nistec",
	} {
		t.Run(pkg, func(t *testing.T) {
			ctx, request := sdkASMDataRequest(t, "amd64", pkg)
			observed, err := CaptureGoAMD64SDKFrameObservation(ctx, request)
			if err != nil {
				t.Fatal("complete original SDK source/listing/object observation", err)
			}
			defer observed.Close()
			source := observed.source
			if len(source.observation.Frames) == 0 || len(source.observation.Frames) != len(source.observation.Objects.Sources) {
				t.Fatal("native frame source inventory differs")
			}
			var frames, fips int
			for ordinal, frame := range source.observation.Frames {
				object := source.observation.Objects.Sources[ordinal]
				if len(frame.Records) != len(object.Original.Texts) {
					t.Fatal("native frame TEXT inventory differs", frame.Source)
				}
				frames += len(frame.Records)
				for _, text := range object.Original.Texts {
					if text.Kind == 2 {
						fips++
					}
				}
			}
			if fips == 0 || frames == 0 {
				t.Fatal("actual original FIPS source precondition absent")
			}
			if err := observed.Verify(ctx); err != nil {
				t.Fatal("live original source verification", err)
			}
			t.Logf("complete native frames=%d FIPS=%d; observation only, not translation or execution", frames, fips)
			sdkASMDataRetain(t, source)
		})
	}
}
