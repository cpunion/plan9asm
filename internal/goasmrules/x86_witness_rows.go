package goasmrules

import (
	"fmt"
	"strconv"
	"strings"
)

// X86PackedBinaryRow retains the owning official table row. Witnesses are
// expectations, not evidence that Go assembled or plan9asm translated them.
type X86PackedBinaryRow struct {
	ID, Opcode string
	Witnesses  []X86PackedBinaryWitness
}

// Unary and immediate AVX rows leave Second empty, as do legacy unary immediates;
// legacy RMW immediates leave First empty and retain data/destination in Second.
// Syntax fields do not certify reads: typed effects own mask and old bytes.
type X86PackedBinaryWitness struct {
	Instruction, Generation    string
	First, Second, Destination string
	Width                      int
	Memory, Masked, Zeroing    bool
	Broadcast                  bool
	MaskResult                 bool
	Encoding                   []byte
}

// X86ImmediateIntegerCompareRows retains all signed/unsigned B/W/D/Q
// immediate comparisons, with K results independent from the input width.
func X86ImmediateIntegerCompareRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `VPCMPB VPCMPUB VPCMPW VPCMPUW VPCMPD VPCMPUD VPCMPQ VPCMPUQ`)
}

// X86PackedNumericConversionRows retains ordinary packed FP/integer numeric
// conversions. The six original MMX rows keep their separate region owner.
func X86PackedNumericConversionRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		CVTPS2PL CVTTPS2PL CVTPD2PL CVTTPD2PL CVTPL2PS CVTPL2PD
		VCVTPS2DQ VCVTTPS2DQ VCVTPD2DQX VCVTPD2DQY VCVTTPD2DQX VCVTTPD2DQY VCVTPD2DQ VCVTTPD2DQ
		VCVTPD2QQ VCVTPD2UQQ VCVTTPD2QQ VCVTTPD2UQQ VCVTPS2UDQ VCVTTPS2UDQ VCVTUDQ2PS
		VCVTQQ2PD VCVTUQQ2PD VCVTPS2QQ VCVTPS2UQQ VCVTTPS2QQ VCVTTPS2UQQ VCVTUDQ2PD
		VCVTPD2UDQX VCVTTPD2UDQX VCVTQQ2PSX VCVTUQQ2PSX
		VCVTPD2UDQY VCVTTPD2UDQY VCVTQQ2PSY VCVTUQQ2PSY
		VCVTPD2UDQ VCVTTPD2UDQ VCVTQQ2PS VCVTUQQ2PS VCVTDQ2PS VCVTDQ2PD`)
}

// X86PackedIntegerMinMaxRows retains all amd64 legacy and AVX min/max rows.
func X86PackedIntegerMinMaxRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		PMAXSB PMAXSW PMAXSD PMAXUB PMAXUW PMAXUD
		PMINSB PMINSW PMINSD PMINUB PMINUW PMINUD
		VPMAXSB VPMAXSW VPMAXSD VPMAXSQ VPMAXUB VPMAXUW VPMAXUD VPMAXUQ
		VPMINSB VPMINSW VPMINSD VPMINSQ VPMINUB VPMINUW VPMINUD VPMINUQ`)
}

// X86PackedAverageMultiplyRows retains the XMM/VEX/EVEX rounded-average,
// word-product, adjacent multiply-add and widened dword-product families.
// MMX rows keep their separate region/EMMS contract and are not included.
func X86PackedAverageMultiplyRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		PAVGB PAVGW VPAVGB VPAVGW
		PMULLW PMULHW PMULHUW PMULHRSW VPMULLW VPMULHW VPMULHUW VPMULHRSW
		PMADDUBSW PMADDWL VPMADDUBSW VPMADDWD
		PMULDQ PMULULQ VPMULDQ VPMULUDQ`)
}

// X86PackedLogicalRows retains packed integer and PS/PD bitwise families,
// including the separate EVEX D/Q names. MMX regions remain a separate owner.
func X86PackedLogicalRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		PAND PANDN POR PXOR ANDPS ANDPD ANDNPS ANDNPD ORPS ORPD XORPS XORPD
		VPAND VPANDN VPOR VPXOR VANDPS VANDPD VANDNPS VANDNPD VORPS VORPD VXORPS VXORPD
		VPANDD VPANDQ VPANDND VPANDNQ VPORD VPORQ VPXORD VPXORQ`)
}

// X86PackedUnpackRows retains integer and PS/PD low/high interleave families.
// The six MMX rows retain their separate region and memory-access contracts.
func X86PackedUnpackRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		PUNPCKLBW PUNPCKHBW PUNPCKLWL PUNPCKHWL PUNPCKLLQ PUNPCKHLQ PUNPCKLQDQ PUNPCKHQDQ
		UNPCKLPS UNPCKHPS UNPCKLPD UNPCKHPD
		VPUNPCKLBW VPUNPCKHBW VPUNPCKLWD VPUNPCKHWD VPUNPCKLDQ VPUNPCKHDQ VPUNPCKLQDQ VPUNPCKHQDQ
		VUNPCKLPS VUNPCKHPS VUNPCKLPD VUNPCKHPD`)
}

// X86PackedNarrowRows retains signed/unsigned saturating dword-to-word and
// word-to-byte packs. The three MMX rows keep their separate region owner.
func X86PackedNarrowRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		PACKSSLW PACKSSWB PACKUSDW PACKUSWB VPACKSSDW VPACKSSWB VPACKUSDW VPACKUSWB`)
}

// X86SingleSourceNarrowingRows retains every truncating and saturating store
// row, including Go-encoded memory.Z forms whose architectural result is #UD.
func X86SingleSourceNarrowingRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		VPMOVDB VPMOVDW VPMOVQB VPMOVQD VPMOVQW VPMOVWB
		VPMOVSDB VPMOVSDW VPMOVSQB VPMOVSQD VPMOVSQW VPMOVSWB
		VPMOVUSDB VPMOVUSDW VPMOVUSQB VPMOVUSQD VPMOVUSQW VPMOVUSWB`)
}

// X86PackedPerLaneShiftRows retains the complete W/D/Q logical-left,
// logical-right and arithmetic-right per-lane shift families.
func X86PackedPerLaneShiftRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		VPSLLVW VPSLLVD VPSLLVQ VPSRLVW VPSRLVD VPSRLVQ VPSRAVW VPSRAVD VPSRAVQ`)
}

// X86PackedUniformByteShiftRows retains complete XMM and AVX uniform lane
// shifts and per-128-bit byte shifts. MMX regions keep their separate owner.
func X86PackedUniformByteShiftRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		PSLLW PSLLL PSLLQ PSRLW PSRLL PSRLQ PSRAW PSRAL PSLLO PSRLO
		VPSLLW VPSLLD VPSLLQ VPSRLW VPSRLD VPSRLQ VPSRAW VPSRAD VPSRAQ VPSLLDQ VPSRLDQ`)
}

// X86PackedSignRows retains all B/W/D conditional-sign rows. Go's legacy
// table is XMM-only; VEX adds X/Y forms, without MMX, EVEX or broadcast.
func X86PackedSignRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `PSIGNB PSIGNW PSIGND VPSIGNB VPSIGNW VPSIGND`)
}

// X86PackedAbsRows retains all unary B/W/D/Q ABS rows, including each
// overlapping VEX/EVEX row and the original mask/broadcast attributes.
func X86PackedAbsRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `PABSB PABSW PABSD VPABSB VPABSW VPABSD VPABSQ`)
}

// X86PackedBitCountConflictRows retains all EVEX unary bit-count/conflict rows.
// Their shared encoding grammar does not combine the independent semantic owners.
func X86PackedBitCountConflictRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `VPOPCNTB VPOPCNTW VPOPCNTD VPOPCNTQ VPLZCNTD VPLZCNTQ
		VPCONFLICTD VPCONFLICTQ`)
}

// X86Reciprocal14Rows retains all packed/scalar EVEX reciprocal14 estimates.
// The byte grammar owns roles and tuples, not numerical or native authority.
func X86Reciprocal14Rows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx,
		"VRCP14PS VRCP14PD VRCP14SS VRCP14SD VRSQRT14PS VRSQRT14PD VRSQRT14SS VRSQRT14SD")
}

// X86PackedRotateRows retains the complete immediate and per-lane D/Q rotate family.
// Variable members remain independently owned by the EVEX binary remainder scope.
func X86PackedRotateRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `VPROLD VPROLQ VPRORD VPRORQ
		VPROLVD VPROLVQ VPRORVD VPRORVQ`)
}

// X86PackedDuplicateRows retains all legacy/VEX/EVEX duplicate-element moves.
func X86PackedDuplicateRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx,
		`MOVDDUP MOVSHDUP MOVSLDUP VMOVDDUP VMOVSHDUP VMOVSLDUP`)
}

// X86PackedImmediateShuffleRows retains every non-MMX legacy/VEX/EVEX source row.
// PSHUFL and PSHUFD are distinct optab entries, not frontend-only aliases.
func X86PackedImmediateShuffleRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx,
		`PSHUFL PSHUFD PSHUFHW PSHUFLW VPSHUFD VPSHUFHW VPSHUFLW`)
}

// X86PackedWideningRows retains the complete sign/zero-extension source family.
// Shrinking native tuples and source carriers remain separate from result width.
func X86PackedWideningRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		PMOVSXBW PMOVSXBD PMOVSXBQ PMOVSXWD PMOVSXWQ PMOVSXDQ
		PMOVZXBW PMOVZXBD PMOVZXBQ PMOVZXWD PMOVZXWQ PMOVZXDQ
		VPMOVSXBW VPMOVSXBD VPMOVSXBQ VPMOVSXWD VPMOVSXWQ VPMOVSXDQ
		VPMOVZXBW VPMOVZXBD VPMOVZXBQ VPMOVZXWD VPMOVZXWQ VPMOVZXDQ`)
}

// X86PackedBroadcastRows retains scalar, block, GP and K-data broadcast rows.
func X86PackedBroadcastRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		VPBROADCASTB VPBROADCASTW VPBROADCASTD VPBROADCASTQ VBROADCASTSS VBROADCASTSD
		VBROADCASTI32X2 VBROADCASTF32X2 VBROADCASTF32X4 VBROADCASTF64X2 VBROADCASTI32X4 VBROADCASTI64X2
		VBROADCASTF32X8 VBROADCASTF64X4 VBROADCASTI32X8 VBROADCASTI64X4
		VPBROADCASTMB2Q VPBROADCASTMW2D VBROADCASTF128 VBROADCASTI128`)
}

// X86PackedMoveRows retains the twelve AVX whole-vector move names. Legacy
// moves are separate; register alternatives belong to the first store rows.
func X86PackedMoveRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `VMOVDQA VMOVDQU VMOVAPD VMOVAPS VMOVUPD VMOVUPS
		VMOVDQA32 VMOVDQA64 VMOVDQU8 VMOVDQU16 VMOVDQU32 VMOVDQU64`)
}

// X86PackedHorizontalIntegerRows retains the XMM and VEX horizontal add/sub
// rows. PHADDD's distinct MMX row keeps its independent region/EMMS owner.
func X86PackedHorizontalIntegerRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		PHADDW PHADDD PHADDSW PHSUBW PHSUBD PHSUBSW
		VPHADDW VPHADDD VPHADDSW VPHSUBW VPHSUBD VPHSUBSW`)
}

// X86PackedByteShuffleRows retains byte-indexed vector permutations. The
// separate VPSHUFBITQMB bit-to-mask family does not share this operand grammar.
func X86PackedByteShuffleRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `PSHUFB VPSHUFB`)
}

// X86PackedIntegerCompareRows retains both vector and mask results. MMX's
// six legacy rows retain their separate region and EMMS contract.
func X86PackedIntegerCompareRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		PCMPEQB PCMPEQW PCMPEQL PCMPEQQ PCMPGTB PCMPGTW PCMPGTL PCMPGTQ
		VPCMPEQB VPCMPEQW VPCMPEQD VPCMPEQQ VPCMPGTB VPCMPGTW VPCMPGTD VPCMPGTQ`)
}

// X86PackedMaskBlendRows retains the floating-point and integer mask-blend
// families. Only the PS/PD/D/Q members permit scalar-memory broadcast.
func X86PackedMaskBlendRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `VBLENDMPS VBLENDMPD VPBLENDMB VPBLENDMW VPBLENDMD VPBLENDMQ`)
}

// X86FMA3Rows retains the original named fused floating-point rows. The
// official table owns each member, order, precision and physical encoding.
func X86FMA3Rows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	var names []string
	for _, rule := range avx {
		if strings.HasPrefix(rule.Opcode, "VFM") || strings.HasPrefix(rule.Opcode, "VFNM") {
			names = append(names, rule.Opcode)
		}
	}
	return x86PackedBinaryRows(legacy, avx, strings.Join(names, " "))
}

// X86AVXFloatBinaryRows retains the ordinary PS/PD/SS/SD arithmetic rows;
// legacy, horizontal, alternating and raw FP16 families have separate owners.
func X86AVXFloatBinaryRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `VADDPS VADDPD VADDSS VADDSD VSUBPS VSUBPD VSUBSS VSUBSD
		VMULPS VMULPD VMULSS VMULSD VDIVPS VDIVPD VDIVSS VDIVSD
		VMINPS VMINPD VMINSS VMINSD VMAXPS VMAXPD VMAXSS VMAXSD`)
}

// X86PackedEVEXBinaryRemainderRows retains the remaining _yvblendmpd members.
// Their common encoding grammar does not imply common old-destination reads.
func X86PackedEVEXBinaryRemainderRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `VPDPBUSD VPDPBUSDS VPDPWSSD VPDPWSSDS VPMADD52HUQ VPMADD52LUQ
		VPERMB VPERMW VPERMI2B VPERMI2W VPERMI2D VPERMI2Q VPERMI2PS VPERMI2PD
		VPERMT2B VPERMT2W VPERMT2D VPERMT2Q VPERMT2PS VPERMT2PD VPMULTISHIFTQB
		VPROLVD VPROLVQ VPRORVD VPRORVQ VPSHLDVW VPSHLDVD VPSHLDVQ VPSHRDVW VPSHRDVD VPSHRDVQ`)
}

// X86PackedBinaryPilotRows covers the shared packed add/subtract and D/Q
// low-multiply pilot. MMX and the separate W multiply family are not included.
// PADDD is a frontend alias, not an additional official encoder row.
func X86PackedBinaryPilotRows(legacy X86LegacyCatalog, avx []X86AVXRule) ([]X86PackedBinaryRow, error) {
	return x86PackedBinaryRows(legacy, avx, `
		PADDB PADDW PADDL PADDQ PADDSB PADDSW PADDUSB PADDUSW
		VPADDB VPADDW VPADDD VPADDQ VPADDSB VPADDSW VPADDUSB VPADDUSW
		PSUBB PSUBW PSUBL PSUBQ PSUBSB PSUBSW PSUBUSB PSUBUSW
		VPSUBB VPSUBW VPSUBD VPSUBQ VPSUBSB VPSUBSW VPSUBUSB VPSUBUSW
		PMULLD VPMULLD VPMULLQ`)
}

func x86PackedBinaryRows(legacy X86LegacyCatalog, avx []X86AVXRule, names string) ([]X86PackedBinaryRow, error) {
	// Scope names are independent of the translator's registry. All operand,
	// width, suffix and encoding axes still come from the original Go rows.
	selected := map[string]bool{}
	for _, name := range strings.Fields(names) {
		selected[name] = false
	}
	var rows []X86PackedBinaryRow
	for _, entry := range legacy.Optab {
		if _, ok := selected[entry.Opcode]; !ok {
			continue
		}
		for _, operand := range entry.Rows {
			classes := strings.Join(operand.Classes, ",")
			if (entry.Ytab == "yxcvm1" && classes == "Yxm,Ymr" || entry.Ytab == "yxcvm2" && classes == "Ymm,Yxr") &&
				operand.RowIndex == 1 && operand.ZCase == "Zm_r_xm" && operand.ZOffset == 2 && operand.UnmatchedOffset == 2 {
				continue // Original MMX conversions require a separate physical region.
			}
			if entry.Ytab == "yps" && entry.Prefix == "Py3" && entry.EscapeAdjustment == 0 &&
				(operand.RowIndex == 0 && classes == "Ymm,Ymr" && operand.ZCase == "Zm_r_xm" && operand.ZOffset == 1 && operand.UnmatchedOffset == 0 ||
					operand.RowIndex == 1 && classes == "Yi8,Ymr" && operand.ZCase == "Zibo_m_xm" && operand.ZOffset == 2 && operand.UnmatchedOffset == 1) {
				continue // The sixteen original MMX rows need their region/EMMS owner.
			}
			if operand.RowIndex == 0 && strings.Join(operand.Classes, ",") == "Ymm,Ymr" && operand.UnmatchedOffset == 0 &&
				(entry.Ytab == "ymm" && operand.ZCase == "Zm_r_xm" && operand.ZOffset == 1 ||
					entry.Ytab == "ymmxmm0f38" && entry.Prefix == "Px" && entry.EscapeAdjustment == 1 &&
						operand.ZCase == "Zlitm_r" && operand.ZOffset == 3) {
				continue // MMX regions have a separate owner and EMMS contract.
			}
			witnesses, err := x86PackedLegacyRowWitnesses(entry, operand)
			if err != nil {
				return nil, fmt.Errorf("%s/%d: %w", entry.Opcode, operand.RowIndex, err)
			}
			rows = append(rows, X86PackedBinaryRow{
				ID:     fmt.Sprintf("optab/%d/%s/%s/%d", entry.TableIndex, entry.Opcode, entry.Ytab, operand.RowIndex),
				Opcode: entry.Opcode, Witnesses: witnesses,
			})
			selected[entry.Opcode] = true
		}
	}
	for index, rule := range avx {
		if _, ok := selected[rule.Opcode]; !ok {
			continue
		}
		witnesses, err := x86PackedAVXRowWitnesses(rule)
		if err != nil {
			return nil, fmt.Errorf("%s/%d: %w", rule.Opcode, rule.RowIndex, err)
		}
		rows = append(rows, X86PackedBinaryRow{
			ID:     fmt.Sprintf("avx/%d/%s/%s/%d", index, rule.Opcode, rule.Ytab, rule.RowIndex),
			Opcode: rule.Opcode, Witnesses: witnesses,
		})
		selected[rule.Opcode] = true
	}
	for name, found := range selected {
		if !found {
			return nil, fmt.Errorf("official packed opcode %s has no retained row", name)
		}
	}
	return rows, nil
}

func x86PackedAVXRowWitnesses(rule X86AVXRule) ([]X86PackedBinaryWitness, error) {
	if rule.Ytab == "_yvpcmpb" {
		return x86ImmediateIntegerCompareRowWitnesses(rule)
	}
	if strings.HasPrefix(rule.Ytab, "_yvcvt") && x86PackedWideningRatio(rule) == 0 {
		return x86PackedNumericConversionRowWitnesses(rule)
	}
	if rule.Ytab == "_yvpmovdb" || rule.Ytab == "_yvpmovdw" {
		return x86PackedSingleSourceNarrowRowWitnesses(rule)
	}
	if strings.HasPrefix(rule.Opcode, "VBROADCAST") || strings.HasPrefix(rule.Opcode, "VPBROADCAST") {
		return x86PackedBroadcastRowWitnesses(rule)
	}
	move, store := x86PackedMoveTable(rule.Ytab), x86PackedMoveStoreRow(rule)
	shift, extensionRow := x86PackedUniformShiftTable(rule.Ytab), x86PackedImmediateExtensionRow(rule)
	registerImmediate := x86PackedImmediateRegisterRow(rule)
	widening := x86PackedWideningRatio(rule)
	immediate := extensionRow || registerImmediate
	unary := rule.Ytab == "_yvmovddup" || rule.Ytab == "_yvexpandpd" || move || registerImmediate || widening != 0
	syntax, err := x86PackedSyntaxWitnesses(rule, unary)
	if err != nil {
		return nil, err
	}
	masked := rule.ZCase == "Zevex_rm_v_k_r" || rule.ZCase == "Zevex_rm_k_r" ||
		rule.ZCase == "Zevex_i_rm_k_vo" || rule.ZCase == "Zevex_i_rm_k_r" || store && rule.ZCase == "Zevex_r_k_rm"
	kind, stride, width := "vex", 2, 128
	if strings.HasPrefix(rule.ZCase, "Zevex_") {
		kind, stride = "evex", 3
	}
	if extensionRow {
		stride++
	}
	// The validated data/result class owns vector width. Widening has a
	// smaller independent source carrier; mask results keep their data class.
	dataClass := rule.Classes[0]
	if shift || immediate {
		dataClass = rule.Classes[1]
	}
	if widening != 0 {
		dataClass = rule.Classes[len(rule.Classes)-1]
	}
	bank := strings.ToUpper(dataClass[1:2])
	if bank == "Y" {
		width = 256
	} else if bank == "Z" {
		width = 512
	}
	wantClass := map[string]string{"X": "Yxm", "Y": "Yym", "Z": "Yzm"}[bank]
	if kind == "evex" && bank != "Z" {
		wantClass += "Evex"
	}
	wideningTuple := 0
	if widening != 0 {
		wantClass, wideningTuple = x86PackedWideningSource(width/8, widening, kind == "evex")
	}
	if shift && !immediate {
		wantClass = "Yxm"
		if kind == "evex" {
			wantClass += "Evex"
		}
	}
	if extensionRow && kind == "vex" {
		wantClass = "Y" + strings.ToLower(bank) + "r"
	}
	wantOffset := stride
	if kind == "evex" && !masked && rule.Ytab != "_yvpslldq" ||
		kind == "vex" && immediate && rule.Classes[0] == "Yu8" {
		wantOffset = 0
	}
	memoryClass := rule.Classes[0]
	if immediate {
		memoryClass = rule.Classes[1]
	}
	if store {
		memoryClass = rule.Classes[len(rule.Classes)-1]
	}
	if memoryClass != wantClass || rule.ZOffset != wantOffset {
		return nil, fmt.Errorf("generation/class/row stride mismatch")
	}
	header := rule.OpBytes[rule.EncodingOffset]
	optional := []string{kind + "66", kind + "0F", kind + "0F38", kind + "W0", kind + "W1"}
	if move || rule.Ytab == "_yvaddsd" || rule.Ytab == "_yvmovddup" || registerImmediate {
		optional = append(optional, kind+"F2", kind+"F3")
	}
	atoms, err := x86PackedEncodingAtoms(header, []string{
		"avxEscape", kind + strconv.Itoa(width),
	}, optional)
	if err != nil || atoms[kind+"0F"] == atoms[kind+"0F38"] || atoms[kind+"W0"] == atoms[kind+"W1"] {
		return nil, fmt.Errorf("unmodeled packed encoding header %q", header)
	}
	if widening != 0 && (!atoms[kind+"66"] || !atoms[kind+"0F38"] || !atoms[kind+"W0"]) {
		return nil, fmt.Errorf("widening prefix, map or W mode differs from its official transport")
	}
	duplicate := rule.Ytab == "_yvmovddup" && (atoms[kind+"F2"] || atoms[kind+"F3"])
	if duplicate && (!atoms[kind+"0F"] || kind == "vex" && !atoms[kind+"W0"] ||
		kind == "evex" && atoms[kind+"F2"] != atoms[kind+"W1"]) {
		return nil, fmt.Errorf("duplicate prefix, map or W mode differs from its official transport")
	}
	if registerImmediate && (!atoms[kind+"0F"] || !atoms[kind+"W0"]) {
		return nil, fmt.Errorf("shuffle map or W mode differs from its official transport")
	}
	if (shift || extensionRow) && (!atoms[kind+"66"] || !atoms[kind+"0F"] ||
		(kind == "vex" || rule.Ytab == "_yvpslldq") && !atoms[kind+"W0"]) {
		return nil, fmt.Errorf("shift prefix or W mode differs from its official transport")
	}
	mapByte, wBit, length, pp := byte(1), byte(0), byte(0), byte(0)
	for index, prefix := range []string{"66", "F3", "F2"} {
		if atoms[kind+prefix] {
			if pp != 0 {
				return nil, fmt.Errorf("conflicting packed mandatory prefixes")
			}
			pp = byte(index + 1)
		}
	}
	if registerImmediate && pp == 0 {
		return nil, fmt.Errorf("shuffle requires its retained mandatory prefix")
	}
	if atoms[kind+"0F38"] {
		mapByte = 2
	}
	if atoms[kind+"W1"] {
		wBit = 0x80
	}
	if width == 256 {
		length = 1
	} else if width == 512 {
		length = 2
	}
	if kind == "evex" {
		features := rule.OpBytes[rule.EncodingOffset+1]
		tuple := width / 8
		optional := []string{"evexZeroingEnabled", "evexBcstN4", "evexBcstN8"}
		if widening != 0 {
			tuple, optional = wideningTuple, []string{"evexZeroingEnabled"}
		}
		if registerImmediate {
			optional = []string{"evexZeroingEnabled"}
			if atoms["evex66"] {
				optional = append(optional, "evexBcstN4")
			}
		}
		if duplicate {
			optional = []string{"evexZeroingEnabled"}
			if width == 128 && atoms["evexF2"] {
				tuple = 8 // DDUP.X reads m64; other duplicate tuples stay full width.
			}
		}
		if shift && !immediate {
			tuple = 16 // Uniform count access stays m128 for X/Y/Z data.
		}
		if move || shift && (!immediate || rule.Ytab == "_yvpslldq") {
			optional = []string{"evexZeroingEnabled"}
		}
		if rule.Ytab == "_yvaddsd" || rule.Ytab == "_yvgetexpsd" {
			tuple = 4
			if wBit != 0 {
				tuple = 8
			}
			optional = []string{"evexZeroingEnabled"}
		}
		if rule.Ytab == "_yvaddpd" && width == 512 || rule.Ytab == "_yvaddsd" {
			optional = append(optional, "evexRoundingEnabled", "evexSaeEnabled")
		}
		if rule.Ytab == "_yvpslldq" {
			optional = nil // Per-128-bit byte shifts have neither mask nor BCST.
		}
		flags, err := x86PackedEncodingAtoms(features, []string{"evexN" + strconv.Itoa(tuple)}, optional)
		if err != nil || flags["evexBcstN4"] && flags["evexBcstN8"] {
			return nil, fmt.Errorf("unmodeled packed encoding features %q", features)
		}
	}
	opcodeIndex := rule.EncodingOffset + stride - 1
	if extensionRow {
		opcodeIndex--
	}
	opcode, err := strconv.ParseUint(rule.OpBytes[opcodeIndex], 0, 8)
	if err != nil {
		return nil, fmt.Errorf("packed opcode is not a byte literal")
	}
	var extension uint64
	if extensionRow {
		extension, err = strconv.ParseUint(rule.OpBytes[opcodeIndex+1], 0, 8)
		if err != nil || extension > 7 {
			return nil, fmt.Errorf("immediate opcode extension is not a three-bit literal")
		}
	}
	var witnesses []X86PackedBinaryWitness
	for _, instruction := range syntax {
		if kind == "evex" {
			// High banks force EVEX even for unmasked X/Y forms. Low-bank
			// syntax alone cannot identify overlapping VEX/EVEX rows.
			if widening != 0 {
				sourceBank := strings.ToUpper(rule.Classes[0][1:2])
				if sourceBank != bank {
					instruction = strings.Replace(instruction, sourceBank+"0", sourceBank+"16", 1)
				}
			}
			instruction = strings.NewReplacer(bank+"0", bank+"16", bank+"1", bank+"17", bank+"2", bank+"18").Replace(instruction)
		}
		if kind == "evex" && shift && !immediate {
			instruction = strings.ReplaceAll(instruction, "X0", "X16")
		}
		fields := strings.Fields(instruction)
		operands := strings.Split(fields[1], ",")
		second, vvvv, vprime := operands[1], byte(0x70), byte(0)
		first := operands[0]
		if immediate {
			first, second = operands[1], ""
		}
		if extensionRow {
			vvvv = 0x68 // Destination 2/18 is vvvv/V' only in /digit formats.
		}
		if unary {
			// Unary encodings reserve VEX.vvvv and EVEX.V' as ones;
			// the mask and merging destination remain separate inputs.
			second, vvvv, vprime = "", 0x78, 0x08
		}
		witness := X86PackedBinaryWitness{
			Instruction: instruction, Generation: kind, Width: width,
			First: first, Second: second, Destination: operands[len(operands)-1],
			Memory: first == "(AX)", Masked: masked,
			Zeroing: strings.HasSuffix(fields[0], ".Z"), Broadcast: strings.Contains(fields[0], ".BCST"),
			MaskResult: rule.Classes[len(rule.Classes)-1] == "Yk",
		}
		modrm := byte(0xd0) // destination 2/18, r/m register 0/16.
		memory := witness.Memory || store && witness.Destination == "(AX)"
		if store {
			modrm = 0xc2 // source 0/16 owns ModRM.reg; destination 2/18 is r/m.
		}
		if memory {
			modrm = 0x10 // the only memory witness is the undisturbed (AX).
			if store {
				modrm = 0x00
			}
		}
		if extensionRow {
			modrm = 0xc0 | byte(extension)<<3
			if memory {
				modrm = byte(extension) << 3
			}
		}
		p0, p2 := byte(0), byte(0)
		if kind == "evex" {
			p0, p2 = byte(0xa0), length<<5|vprime
			if memory {
				p0 = 0xe0
			}
			if extensionRow {
				p0 |= 0x10 // ModRM.reg is /digit, not destination register 18.
			}
			if witness.MaskResult {
				p0 |= 0x10 // K2 has no destination bit 4, unlike vector 18.
			}
			if masked {
				p2 |= 1 // K1; binary V' is zero for the second source 17.
			}
			if witness.Zeroing {
				p2 |= 0x80
			}
			if witness.Broadcast {
				p2 |= 0x10
			}
			if strings.Contains(fields[0], ".SAE") {
				p2 |= 0x10 // SAE keeps LL; embedded rounding replaces it below.
			}
			for control, suffix := range []string{"RN_SAE", "RD_SAE", "RU_SAE", "RZ_SAE"} {
				if strings.Contains(fields[0], "."+suffix) {
					p2 = p2&^0x60 | byte(control)<<5 | 0x10
				}
			}
		}
		transport := x86PackedAVXOpcode{kind, mapByte, wBit, pp, length, byte(opcode)}
		witness.Encoding = transport.bytes(vvvv, p0, p2, modrm)
		if immediate {
			value, err := strconv.ParseInt(strings.TrimPrefix(operands[0], "$"), 0, 16)
			if err != nil || value < -128 || value > 255 {
				return nil, fmt.Errorf("shift witness has no byte immediate")
			}
			witness.Encoding = append(witness.Encoding, byte(value))
		}
		witnesses = append(witnesses, witness)
	}
	return witnesses, nil
}

func x86PackedEncodingAtoms(expression string, required, optional []string) (map[string]bool, error) {
	allowed, found := map[string]bool{}, map[string]bool{}
	for _, term := range append(append([]string(nil), required...), optional...) {
		allowed[term] = true
	}
	for _, term := range strings.FieldsFunc(expression, func(r rune) bool {
		return r == '|' || r == '(' || r == ')' || r == ' ' || r == '\t' || r == '\n'
	}) {
		if !allowed[term] || found[term] {
			return nil, fmt.Errorf("unrecognized or repeated encoding term %s", term)
		}
		found[term] = true
	}
	for _, term := range required {
		if !found[term] {
			return nil, fmt.Errorf("missing encoding term %s", term)
		}
	}
	return found, nil
}

func x86PackedLegacyRowWitnesses(entry X86LegacyOpcode, row X86LegacyOperand) ([]X86PackedBinaryWitness, error) {
	literal := entry.Ytab == "ymmxmm0f38" && entry.Prefix == "Px" && entry.EscapeAdjustment == 1 &&
		row.RowIndex == 1 && row.ZCase == "Zlitm_r" && row.ZOffset == 5 && row.UnmatchedOffset == 4
	shift := entry.Ytab == "yps" || entry.Ytab == "ypsdq"
	immediate := shift && strings.Join(row.Classes, ",") == "Yi8,Yxr"
	shuffle := entry.Ytab == "yxshuf" && strings.Join(row.Classes, ",") == "Yu8,Yxm,Yxr"
	conversion := entry.Ytab == "yxcvm1" || entry.Ytab == "yxcvm2"
	if entry.Kind != "encoding" || entry.EscapeAdjustment != 0 && !literal ||
		strings.Join(row.Classes, ",") != "Yxm,Yxr" && !immediate && !shuffle {
		return nil, fmt.Errorf("unmodeled legacy packed operand row")
	}
	prefix, offset, opcodeBytes := []byte{0x66, 0x0f}, 0, 1
	terminator := 0
	switch {
	case conversion && row.RowIndex == 0 && row.ZCase == "Zm_r_xm" && row.ZOffset == 2 && row.UnmatchedOffset == 0 && len(entry.OpBytes) == 4:
		if entry.Prefix == "Px" && (entry.OpBytes[2] == "Pe" || entry.OpBytes[2] == "Pm") {
			mandatory := map[string]byte{"Pe": 0x66, "Pf2": 0xf2, "Pf3": 0xf3}[entry.OpBytes[0]]
			if mandatory == 0 {
				return nil, fmt.Errorf("unmodeled packed conversion mandatory prefix")
			}
			prefix, offset = []byte{mandatory, 0x0f}, 1
		} else if entry.Ytab == "yxcvm2" && entry.Prefix == "Pm" && entry.OpBytes[1] == "0" && entry.OpBytes[3] == "0" {
			prefix = []byte{0x0f}
		} else {
			return nil, fmt.Errorf("unmodeled packed conversion complete prefix payload")
		}
	case entry.Ytab == "yps" && entry.Prefix == "Py3" && entry.EscapeAdjustment == 0:
		if len(entry.OpBytes) != 8 || entry.OpBytes[3] != "Pe" || entry.OpBytes[5] != "Pe" ||
			entry.OpBytes[0] != entry.OpBytes[4] || entry.OpBytes[1] != entry.OpBytes[6] || entry.OpBytes[2] != entry.OpBytes[7] {
			return nil, fmt.Errorf("unmodeled complete Py3 shift payload")
		}
		if !immediate && row.RowIndex == 2 && row.ZCase == "Zm_r_xm" && row.ZOffset == 2 && row.UnmatchedOffset == 3 {
			offset = 4
		} else if immediate && row.RowIndex == 3 && row.ZCase == "Zibo_m_xm" && row.ZOffset == 3 && row.UnmatchedOffset == 5 {
			offset = 6
		} else {
			return nil, fmt.Errorf("unmodeled Py3 XMM shift row")
		}
	case entry.Ytab == "ypsdq" && entry.Prefix == "Pq" && immediate && row.RowIndex == 0 &&
		row.ZCase == "Zibo_m" && row.ZOffset == 2 && row.UnmatchedOffset == 0 && len(entry.OpBytes) == 2:
	case shuffle && row.RowIndex == 0 && row.ZCase == "Zibm_r" && row.ZOffset == 2 && row.UnmatchedOffset == 0 &&
		len(entry.OpBytes) == 2 && entry.OpBytes[0] == "0x70" && (entry.OpBytes[1] == "0" || entry.OpBytes[1] == "00") &&
		(entry.Prefix == "Pq" || entry.Prefix == "Pf2" || entry.Prefix == "Pf3"):
		// The second atom is the original stride placeholder, not /digit.
		if entry.Prefix == "Pf2" {
			prefix[0] = 0xf2
		} else if entry.Prefix == "Pf3" {
			prefix[0] = 0xf3
		}
	case literal:
		// Zlitm_r emits the selected zero-terminated sequence verbatim,
		// including its mandatory prefix and opcode-map bytes.
		prefix, offset, opcodeBytes, terminator = nil, row.UnmatchedOffset, row.ZOffset-1, 1
		if len(entry.OpBytes) != offset+opcodeBytes+terminator ||
			entry.OpBytes[offset-1] != "0" || entry.OpBytes[len(entry.OpBytes)-1] != "0" {
			return nil, fmt.Errorf("unmodeled legacy packed literal boundaries")
		}
	case entry.Ytab == "ymm" && entry.Prefix == "Py1" && row.RowIndex == 1 && row.ZCase == "Zm_r_xm" && row.ZOffset == 2 && row.UnmatchedOffset == 1:
		if len(entry.OpBytes) != 3 || entry.OpBytes[1] != "Pe" || entry.OpBytes[0] != entry.OpBytes[2] {
			return nil, fmt.Errorf("unmodeled Py1 XMM opcode payload")
		}
		offset = 2 // Py1 consumes the Pe at offset 1 before the opcode.
	case entry.Ytab == "yxm" && (entry.Prefix == "Pe" || entry.Prefix == "Pq" || entry.Prefix == "Pm") &&
		row.RowIndex == 0 && row.ZCase == "Zm_r_xm" && row.ZOffset == 1 && row.UnmatchedOffset == 0:
		if entry.Prefix == "Pm" {
			prefix = []byte{0x0f}
		}
	case entry.Ytab == "yxm" && row.RowIndex == 0 && row.ZCase == "Zm_r_xm" &&
		row.ZOffset == 1 && row.UnmatchedOffset == 0 && len(entry.OpBytes) == 1 &&
		(entry.Prefix == "Pf2" && entry.OpBytes[0] == "0x12" ||
			entry.Prefix == "Pf3" && (entry.OpBytes[0] == "0x12" || entry.OpBytes[0] == "0x16")):
		// Only the duplicate move payloads join this prefix class; scalar
		// arithmetic and conversion yxm rows keep their separate contracts.
		prefix = []byte{0xf2, 0x0f}
		if entry.Prefix == "Pf3" {
			prefix[0] = 0xf3
		}
	case entry.Ytab == "yxm_q4" && entry.Prefix == "Pq4" && row.RowIndex == 0 && row.ZCase == "Zm_r" && row.ZOffset == 1 && row.UnmatchedOffset == 0:
		prefix = append(prefix, 0x38)
	case entry.Ytab == "ymshufb" && entry.Prefix == "Pq" && row.RowIndex == 0 && row.ZCase == "Zm2_r" && row.ZOffset == 2 && row.UnmatchedOffset == 0:
		// Go's Zm2_r emits two selected opcode bytes before ModRM. They
		// are payload, not another mandatory prefix or a mnemonic lookup.
		opcodeBytes = 2
	default:
		return nil, fmt.Errorf("unmodeled legacy packed prefix/offset rule")
	}
	if !shift && !shuffle && !conversion && len(entry.OpBytes) != offset+opcodeBytes+terminator {
		return nil, fmt.Errorf("unmodeled legacy packed opcode payload")
	}
	encoding := prefix
	for _, value := range entry.OpBytes[offset : offset+opcodeBytes] {
		opcode, err := strconv.ParseUint(value, 0, 8)
		if err != nil || literal && opcode == 0 {
			return nil, fmt.Errorf("legacy packed opcode is not a byte literal")
		}
		encoding = append(encoding, byte(opcode))
	}
	sources, extension := []string{"X0", "(AX)"}, uint64(0)
	if immediate {
		var err error
		extension, err = strconv.ParseUint(entry.OpBytes[offset+1], 0, 8)
		if err != nil || extension > 7 {
			return nil, fmt.Errorf("legacy shift opcode extension is not a three-bit literal")
		}
		sources = []string{""}
	}
	var witnesses []X86PackedBinaryWitness
	for _, first := range sources {
		modrm := byte(0xc8)
		if first == "(AX)" {
			modrm = 0x08
		}
		instruction := entry.Opcode + " " + first + ",X1"
		encoded := append(append([]byte(nil), encoding...), modrm)
		second := "X1"
		if shuffle {
			instruction, second = entry.Opcode+" $7,"+first+",X1", ""
			encoded = append(encoded, 7)
		}
		if immediate {
			instruction = entry.Opcode + " $-1,X1"
			encoded = append(append([]byte(nil), encoding...), 0xc1|byte(extension)<<3, 255)
		}
		witnesses = append(witnesses, X86PackedBinaryWitness{
			Instruction: instruction, Generation: "legacy", Width: 128,
			First: first, Second: second, Destination: "X1", Memory: first == "(AX)",
			Encoding: encoded,
		})
	}
	return witnesses, nil
}
