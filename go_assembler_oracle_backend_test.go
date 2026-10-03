package plan9asm

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

func goAssemblerOracleFixture(source string) ([]byte, bool) {
	info, err := os.Lstat(source)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 8<<20 {
		return nil, false
	}
	data, err := os.ReadFile(source)
	return data, err == nil && len(data) <= 8<<20
}

func goAssemblerOracleListingBindsSource(source, position, pc, instruction string) bool {
	pc = strings.TrimSpace(pc)
	if len(pc) < 5 {
		return false
	}
	for _, digit := range pc {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	address, err := strconv.ParseUint(pc, 10, 64)
	if err != nil || fmt.Sprintf("%05d", address) != pc {
		return false
	}
	data, ok := goAssemblerOracleFixture(source)
	line, err := strconv.Atoi(position)
	if !ok || err != nil || line <= 0 {
		return false
	}
	lines := strings.Split(string(data), "\n")
	if line > len(lines) {
		return false
	}
	op, _ := splitOpcode(strings.TrimSpace(lines[line-1]))
	fields := strings.Fields(instruction)
	return len(fields) > 1 && op != "" && strings.EqualFold(op, fields[0])
}

// Go's backends also have a few positionless source diagnostics. Do not
// accept the prefix alone: bind its exact values to an independently invalid
// operand/frame in this regular, bounded fixture. Unknown errors stay fatal.
func goAssemblerOraclePositionlessDiagnostic(source, arch, output string) bool {
	data, ok := goAssemblerOracleFixture(source)
	if !ok {
		return false
	}
	parseArch := ArchAMD64
	if arch == "arm64" {
		parseArch = ArchARM64
	} else if arch != "amd64" && arch != "386" {
		return false
	}
	file, err := Parse(parseArch, string(data))
	if err != nil {
		return false
	}
	for _, line := range strings.Split(output, "\n") {
		const framePrefix = "asm: NOFRAME functions must have a frame size of 0, not "
		if arch == "arm64" && strings.HasPrefix(line, framePrefix) {
			number := strings.TrimPrefix(line, framePrefix)
			frame, err := strconv.ParseInt(number, 10, 32)
			if err != nil || strconv.FormatInt(frame, 10) != number || frame == 0 || frame == -8 {
				continue
			}
			for _, fn := range file.Funcs {
				if fn.FrameSize == frame && arm64SourceFrameFlagsKnown(fn) && !arm64SourceGoFrame(fn).present {
					return true
				}
			}
		}

		const addressPrefix = "asm: asmidx: bad address "
		if arch == "386" && strings.HasPrefix(line, addressPrefix) {
			parts := strings.Split(strings.TrimPrefix(line, addressPrefix), "/")
			if len(parts) != 3 {
				continue
			}
			scale, err := strconv.Atoi(parts[0])
			base := Reg(parts[2])
			if err != nil || strconv.Itoa(scale) != parts[0] ||
				!isX86YrlRegisterForArch(base, "amd64") || isX86YrlRegisterForArch(base, "386") {
				continue
			}
			for _, fn := range file.Funcs {
				for _, ins := range fn.Instrs {
					for _, arg := range ins.Args {
						if arg.Kind == OpMem && arg.Mem.Base == base && string(arg.Mem.Index) == parts[1] && arg.Mem.Scale == int64(scale) {
							return true
						}
					}
				}
			}
		}

		const floatPrefix = "asm: unexpected addr1: type=5 $("
		if (arch == "amd64" || arch == "386") && strings.HasPrefix(line, floatPrefix) && strings.HasSuffix(line, ")") {
			value, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimPrefix(line, floatPrefix), ")"), 64)
			if err != nil {
				continue
			}
			for _, fn := range file.Funcs {
				for _, ins := range fn.Instrs {
					switch ins.Op {
					case "IRETW", "IRETL", "IRETQ", "RETFW", "RETFL", "RETFQ":
						if len(ins.Args) == 1 && ins.Args[0].Kind == OpImm && ins.Args[0].ImmIsFloat &&
							uint64(ins.Args[0].Imm) == math.Float64bits(value) {
							return true
						}
					}
				}
			}
		}
	}
	return false
}
