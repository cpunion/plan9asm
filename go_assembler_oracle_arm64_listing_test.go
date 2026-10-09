package plan9asm

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/xgo-dev/plan9asm/internal/goasmoracle"
)

// This namespace describes DRconv's diagnostic metadata, not accepted operand
// forms or ISA support. AUTO/OREG share the complete Go displacement directory.
var goAssemblerOracleARM64Classes = func() map[string]bool {
	result := make(map[string]bool)
	for _, name := range strings.Fields(`
		NONE REG ZREG RSP FREG VREG PAIR SHIFT EXTREG SPR COND SPOP ARNG ELEM LIST
		ZCON ABCON0 ADDCON0 ABCON AMCON ADDCON MBCON MOVCON BITCON ADDCON2 LCON
		MOVCON2 MOVCON3 VCON FCON VCONADDR AACON AACON2 LACON AECON SBRA LBRA
		SEXT1 SEXT2 SEXT4 SEXT8 SEXT16 LEXT ADDR GOTADDR TLS_LE TLS_IE ROFF GOK TEXTSIZE NCLASS
	`) {
		result[name] = true
	}
	for _, address := range []string{"AUTO", "OREG"} {
		for _, form := range strings.Fields(`
			Z@ NS@ NS@_16 NS@_8 NS@_4 NP@ NP@_16 NQ@_16 N@4K
			PS@ PS@_16 PS@_8 PS@_4 PP@ PP@_16 PQ@_16
			U@4K U@4K_16 U@4K_8 U@4K_4 U@4K_2
			U@8K U@8K_16 U@8K_8 U@8K_4 U@16K U@16K_16 U@16K_8
			U@32K U@32K_16 U@64K L@POOL L@
		`) {
			result[strings.Replace(form, "@", address, 1)] = true
		}
	}
	return result
}()

func goAssemblerOracleARM64SourceDiagnostic(source, arch, output string) bool {
	if arch != "arm64" {
		return false
	}
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "asm: illegal combination: ") &&
			goasmoracle.SourceDiagnosticLineWithListing(source, line, 0, goAssemblerOracleARM64Listing) {
			return true
		}
	}
	return false
}

func goAssemblerOracleARM64Suffix(suffix string) (string, bool) {
	if suffix == "" {
		return "", true
	}
	post, pre := false, false
	for _, part := range strings.Split(strings.TrimPrefix(suffix, "."), ".") {
		switch part {
		case "P":
			post = true
		case "W":
			pre = true
		default:
			return "", false
		}
	}
	canonical := ""
	if post {
		canonical += ".P"
	}
	if pre {
		canonical += ".W"
	}
	return canonical, true
}

func goAssemblerOracleARM64Listing(original, listing []string) bool {
	if len(listing) < 9 {
		return false
	}
	tail := listing[len(listing)-8:]
	for _, class := range tail[:5] {
		if !goAssemblerOracleARM64Classes[class] {
			return false
		}
	}
	if tail[5] != "," {
		return false
	}
	for _, text := range tail[6:] {
		kind, err := strconv.Atoi(text)
		if err != nil || kind < 0 || kind > 14 || strconv.Itoa(kind) != text {
			return false
		}
	}
	sourceOp, sourceSuffix, args, sourceOK := goAssemblerOracleX86Statement(original)
	listedOp, listedSuffix, listedArgs, listedOK := goAssemblerOracleX86Statement(listing[:len(listing)-8])
	if !sourceOK || !listedOK || !reflect.DeepEqual(args, listedArgs) {
		return false
	}
	canonical, valid := goAssemblerOracleARM64Suffix(sourceSuffix)
	if !valid || listedSuffix != canonical {
		return false
	}
	switch sourceOp {
	case "B":
		return listedOp == "JMP"
	case "BL":
		return listedOp == "CALL"
	}
	return false
}
