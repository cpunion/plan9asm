package gotoolprofile

import "fmt"

const EmptyAssemblyProtocol = "actual_go_symbol_free_listing_v1"

func ValidateEmptyAssembly(input *ConsumerInput, cpp CPPProof, pkg PackageProof) error {
	proof := cpp.EmptyAssembly
	if cpp.Emission != "assembly" && cpp.Emission != "symbol_free" ||
		(cpp.Emission == "symbol_free") != (proof != nil) {
		return fmt.Errorf("empty assembly requires an explicit same-scope native witness")
	}
	if proof == nil {
		return nil // Normal TEXT and data modules use the ordinary output proof.
	}
	if input == nil || input.Observed == nil || proof.Protocol != EmptyAssemblyProtocol ||
		proof.File != cpp.File || proof.SourceSHA256 != input.Sources[cpp.File] ||
		proof.PackagePath != pkg.PackagePath || proof.ProfileID != input.ID ||
		proof.Target != input.Observed.Target || proof.GoVersion != input.Observed.GoVersion ||
		proof.AsmToolSHA256 != input.Observed.ToolBinarySHA256["asm"] ||
		!discoverySHA256Pattern.MatchString(proof.ObjectSHA256) ||
		proof.ListingSHA256 != bytesSHA256(nil) {
		return fmt.Errorf("empty assembly evidence differs from its actual Go/source/package/profile scope")
	}
	return nil
}
