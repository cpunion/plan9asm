package plan9asm

import (
	"errors"
	"testing"

	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
)

func TestSDKAMD64OriginalFrameNativeCompleteMetadata(t *testing.T) {
	texts := []gostaticstorage.TextRecord{{Name: "private", Size: 1, Code: []byte{0xc3}}, {Name: "sdk.Entry", Size: 1, Code: []byte{0xc3}}}
	listing := []byte("private STEXT static nosplit asm size=1 align=0x0 args=0xffffffff80000000 locals=0x0 funcid=0x0\n" +
		"sdk.Entry STEXT asm size=1 align=0x0 args=0x61 locals=0x128 funcid=0x0\n")
	records, err := sdkAMD64FrameNativeRecords(listing, texts)
	if err != nil || len(records) != 2 || !records[0].Static || !records[0].NoSplit || records[0].ArgsKnown ||
		records[0].Args != -0x80000000 || records[1].NoSplit || !records[1].ArgsKnown || records[1].Args != 97 || records[1].Locals != 296 {
		t.Fatal("private unknown args or stack-split native frame was disguised as a declared leaf", records, err)
	}
	for _, raw := range []string{
		"sdk.Entry STEXT asm size=1 args=0x61 locals=0x128\n",
		"private STEXT static nosplit asm size=1 args=0xffffffff80000000 locals=0x0\n",
		"private STEXT static nosplit asm size=2 args=0xffffffff80000000 locals=0x0\n",
		"private STEXT foreign size=1 args=0x0 locals=0x0\n",
		string(listing) + string(listing),
	} {
		if _, err := sdkAMD64FrameNativeRecords([]byte(raw), texts); !errors.Is(err, ErrProbeNeedsContext) {
			t.Fatal("partial, reordered, foreign or duplicate native frame inventory survived", raw, err)
		}
	}
}

func TestSDKAMD64OriginalFrameNativeListingByteBoundary(t *testing.T) {
	source := &GoX86SDKASMDataSource{}
	if err := source.retainSDKFrameListingBytes(sdkAMD64FrameListingBytes); err != nil {
		t.Fatal("exact SDK-primary listing byte boundary", err)
	}
	for _, added := range []int{1, -1} {
		if err := source.retainSDKFrameListingBytes(added); !errors.Is(err, ErrProbeNeedsContext) ||
			source.frameListingBytes != sdkAMD64FrameListingBytes {
			t.Fatal("oversized native listing changed its retained accounting", added, err)
		}
	}
}
