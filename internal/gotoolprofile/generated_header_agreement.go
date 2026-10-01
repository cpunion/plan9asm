package gotoolprofile

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// CanonicalGeneratedHeader is the effective assembly input. The original
// compiler formatting/hash remains in GeneratedHeaderProof; every definition,
// including unused names and presence, is retained in this deterministic text.
func CanonicalGeneratedHeader(definitions map[string]string) ([]byte, error) {
	if definitions == nil || len(definitions) > 4096 {
		return nil, fmt.Errorf("missing or unbounded generated definitions")
	}
	names := make([]string, 0, len(definitions))
	for name := range definitions {
		names = append(names, name)
	}
	sort.Strings(names)
	var text strings.Builder
	for _, name := range names {
		fmt.Fprintf(&text, "#define %s %s\n", name, definitions[name])
	}
	data := []byte(text.String())
	parsed, err := HeaderDefinitions(data)
	if err != nil || !reflect.DeepEqual(parsed, definitions) {
		return nil, fmt.Errorf("generated definitions cannot form an exact canonical header: %v", err)
	}
	return data, nil
}

// ValidateGeneratedHeaderAgreement compares two independently produced actual
// compiler queries. Self-consistent altered JSON is not agreement. Only the
// physical compiler header/object bytes may differ after both queries passed
// complete source/profile/role/import validation; their full definitions may not.
func ValidateGeneratedHeaderAgreement(input *ConsumerInput, expected, actual *MetadataProof, tags []string) error {
	if expected == nil || actual == nil {
		return fmt.Errorf("generated assembly requires producer and independent consumer compiler metadata")
	}
	for _, proof := range []*MetadataProof{expected, actual} {
		if err := ValidateMetadata(input, proof, tags); err != nil {
			return err
		}
	}
	if !reflect.DeepEqual(expected.Packages, actual.Packages) || !equalDiscoveryStrings(expected.CustomTags, actual.CustomTags) {
		return fmt.Errorf("generated-header compiler queries selected different package/source/tag roles")
	}
	byPackage := make(map[string]GeneratedHeaderProof, len(expected.Headers))
	for _, header := range expected.Headers {
		byPackage[header.PackagePath] = header
	}
	for _, header := range actual.Headers {
		wanted, found := byPackage[header.PackagePath]
		if !found {
			return fmt.Errorf("generated-header consumer selected an unregistered package")
		}
		wanted.HeaderSHA256, wanted.ObjectSHA256 = "", ""
		header.HeaderSHA256, header.ObjectSHA256 = "", ""
		if !reflect.DeepEqual(wanted, header) {
			return fmt.Errorf("independent generated-header definitions/imports differ from producer metadata")
		}
	}
	return nil
}

func generatedCPPSourceSHA(proof *MetadataProof, packagePath, file string) string {
	if proof == nil || file != packagePath+"/go_asm.h" {
		return ""
	}
	for _, header := range proof.Headers {
		if header.PackagePath == packagePath {
			data, err := CanonicalGeneratedHeader(header.Definitions)
			if err == nil {
				return bytesSHA256(data)
			}
		}
	}
	return ""
}
