package plan9asm

import "testing"

func TestDecodedX86VariableDwordPermuteInstructionCompleteVEXFamily(t *testing.T) {
	for _, test := range []struct {
		opcode byte
		op     Op
	}{
		{0x36, "VPERMD"},
		{0x16, "VPERMPS"},
	} {
		t.Run(string(test.op), func(t *testing.T) {
			code := []byte{0xc4, 0xe2, 0x25, test.opcode, 0xd6}
			instruction, length, ok, err := decodedX86VariableDwordPermuteInstruction(code, 64)
			if err != nil {
				t.Fatal(err)
			}
			if !ok {
				t.Fatal("encoding was not recognized")
			}
			if length != len(code) {
				t.Fatalf("length = %d, want %d", length, len(code))
			}
			want := string(test.op) + " Y6, Y11, Y2"
			if instruction.Op != test.op || instruction.Raw != want {
				t.Fatalf("instruction = %#v, want %q", instruction, want)
			}
		})
	}

	for _, test := range []struct {
		name string
		code []byte
		mode int
		want string
	}{
		{
			name: "extended_registers",
			code: []byte{0xc4, 0x42, 0x05, 0x36, 0xdb},
			mode: 64,
			want: "VPERMD Y11, Y15, Y11",
		},
		{
			name: "memory_sib_disp8",
			code: []byte{0xc4, 0x62, 0x15, 0x36, 0x5c, 0x88, 0x20},
			mode: 64,
			want: "VPERMD 32(AX)(CX*4), Y13, Y11",
		},
		{
			name: "memory_negative_disp8",
			code: []byte{0xc4, 0x62, 0x65, 0x16, 0x64, 0x35, 0xef},
			mode: 64,
			want: "VPERMPS -17(BP)(SI*1), Y3, Y12",
		},
		{
			name: "386_registers",
			code: []byte{0xc4, 0xe2, 0x65, 0x36, 0xe5},
			mode: 32,
			want: "VPERMD Y5, Y3, Y4",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			instruction, length, ok, err := decodedX86VariableDwordPermuteInstruction(test.code, test.mode)
			if err != nil {
				t.Fatal(err)
			}
			if !ok || length != len(test.code) || instruction.Raw != test.want {
				t.Fatalf("decode = (%#v, %d, %v), want %q", instruction, length, ok, test.want)
			}
		})
	}
}

func TestDecodedX86VariableDwordPermuteInstructionRejectsInvalidVEXForms(t *testing.T) {
	for _, test := range []struct {
		name string
		code []byte
		mode int
	}{
		{name: "truncated_modrm", code: []byte{0xc4, 0xe2, 0x25, 0x36}, mode: 64},
		{name: "vex_l_zero", code: []byte{0xc4, 0xe2, 0x21, 0x36, 0xd6}, mode: 64},
		{name: "vex_pp_zero", code: []byte{0xc4, 0xe2, 0x24, 0x36, 0xd6}, mode: 64},
		{name: "vex_w_one", code: []byte{0xc4, 0xe2, 0xa5, 0x36, 0xd6}, mode: 64},
		{name: "address_override", code: []byte{0x67, 0xc4, 0xe2, 0x25, 0x36, 0xd6}, mode: 64},
		{name: "extended_register_in_386", code: []byte{0xc4, 0x42, 0x05, 0x36, 0xdb}, mode: 32},
		{name: "rip_relative", code: []byte{0xc4, 0xe2, 0x25, 0x36, 0x05, 0, 0, 0, 0}, mode: 64},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, _, ok, err := decodedX86VariableDwordPermuteInstruction(test.code, test.mode)
			if !ok {
				t.Fatal("VPERMD/VPERMPS encoding was not recognized")
			}
			if err == nil {
				t.Fatal("invalid encoding was accepted")
			}
		})
	}
}

func TestDecodeX86RawDirectiveGroupVPERMDReportedSequence(t *testing.T) {
	code := []byte{
		0xc4, 0xe2, 0x25, 0x36, 0xd6,
		0xc4, 0xe2, 0x7d, 0x36, 0xd6,
		0xc4, 0xe2, 0x35, 0x36, 0xd6,
	}
	decoded, err := decodeX86RawDirectiveGroup(code, 64, 0, "go-pherence VEX VPERMD sequence", map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded) != 3 {
		t.Fatalf("decoded %d instructions, want 3: %#v", len(decoded), decoded)
	}
	for _, instruction := range decoded {
		if instruction.Op != "VPERMD" {
			t.Fatalf("decoded op = %s, want VPERMD", instruction.Op)
		}
	}
}
