//go:build amd64

package amd64conformance

import "testing"

func TestByteMemoryDestinations(t *testing.T) {
	got := [4]byte{0x10, 0xf0, 0x5a, 0xa0}
	byteMemory(&got, 0x0f)
	want := [4]byte{0x1f, 0xff, 0x0a, 0xaf}
	if got != want {
		t.Fatalf("byteMemory() = %#v, want %#v", got, want)
	}
}

func TestPUNPCKLQDQMemorySource(t *testing.T) {
	dst := [2]uint64{0x0123456789abcdef, 0xfedcba9876543210}
	src := [2]uint64{0x1122334455667788, 0x8877665544332211}
	unpackLowQWords(&dst, &src)
	want := [2]uint64{0x0123456789abcdef, 0x1122334455667788}
	if dst != want {
		t.Fatalf("unpackLowQWords() = %#x, want %#x", dst, want)
	}
}

func TestByteMemoryFlags(t *testing.T) {
	values := [4]byte{0xff, 0xff, 0x7f, 0}
	var got [8]byte
	byteFlags(&values, &got)
	wantValues := [4]byte{0, 0, 0, 0x80}
	wantFlags := [8]byte{1, 1, 0, 1, 0, 1, 0, 1}
	if values != wantValues {
		t.Fatalf("byteFlags values = %#v, want %#v", values, wantValues)
	}
	if got != wantFlags {
		t.Fatalf("byteFlags flags = %#v, want %#v", got, wantFlags)
	}
}

func TestPUNPCKLQDQRegisterSource(t *testing.T) {
	got := [2]uint64{0x0123456789abcdef, 0xfedcba9876543210}
	unpackDuplicateLowQWord(&got)
	want := [2]uint64{0x0123456789abcdef, 0x0123456789abcdef}
	if got != want {
		t.Fatalf("unpackDuplicateLowQWord() = %#x, want %#x", got, want)
	}
}

func TestLegacyThreeOperandShift(t *testing.T) {
	if got, want := shiftLegacyThreeOperand(0x12345678, 0x89abcdef, 5), uint32(0x3579bde2); got != want {
		t.Fatalf("shiftLegacyThreeOperand() = %#x, want %#x", got, want)
	}
}

func TestDoubleShiftFamily(t *testing.T) {
	pairs32 := [][2]uint32{
		{0x12345678, 0x89abcdef},
		{0, ^uint32(0)},
		{^uint32(0), 0},
		{0x80000001, 0x7ffffffe},
	}
	for _, pair := range pairs32 {
		src, dst := pair[0], pair[1]
		for count := uint32(0); count < 128; count++ {
			var got [8]uint32
			doubleShift32(&got, src, dst, count)
			want := [8]uint32{
				shld32(src, dst, count), shld32(src, dst, 7),
				shld32(src, dst, count), shld32(src, dst, 7),
				shrd32(src, dst, count), shrd32(src, dst, 7),
				shrd32(src, dst, count), shrd32(src, dst, 7),
			}
			if got != want {
				t.Fatalf("doubleShift32(%#x, %#x, %d) = %#x, want %#x", src, dst, count, got, want)
			}
		}
	}

	pairs64 := [][2]uint64{
		{0x0123456789abcdef, 0xfedcba9876543210},
		{0, ^uint64(0)},
		{^uint64(0), 0},
		{0x8000000000000001, 0x7ffffffffffffffe},
	}
	for _, pair := range pairs64 {
		src, dst := pair[0], pair[1]
		for count := uint64(0); count < 256; count++ {
			var got [8]uint64
			doubleShift64(&got, src, dst, count)
			want := [8]uint64{
				shld64(src, dst, count), shld64(src, dst, 7),
				shld64(src, dst, count), shld64(src, dst, 7),
				shrd64(src, dst, count), shrd64(src, dst, 7),
				shrd64(src, dst, count), shrd64(src, dst, 7),
			}
			if got != want {
				t.Fatalf("doubleShift64(%#x, %#x, %d) = %#x, want %#x", src, dst, count, got, want)
			}
		}
	}
}

func shld32(src, dst, count uint32) uint32 {
	count &= 31
	if count == 0 {
		return dst
	}
	return dst<<count | src>>(32-count)
}

func shrd32(src, dst, count uint32) uint32 {
	count &= 31
	if count == 0 {
		return dst
	}
	return dst>>count | src<<(32-count)
}

func shld64(src, dst, count uint64) uint64 {
	count &= 63
	if count == 0 {
		return dst
	}
	return dst<<count | src>>(64-count)
}

func shrd64(src, dst, count uint64) uint64 {
	count &= 63
	if count == 0 {
		return dst
	}
	return dst>>count | src<<(64-count)
}
