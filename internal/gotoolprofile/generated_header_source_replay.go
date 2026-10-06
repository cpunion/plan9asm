package gotoolprofile

import "fmt"

// ReplayNativeGeneratedHeader queries a closed source model. Its inputs must
// independently match actual selected compiler sources and the stock SDK; it
// grants neither source authentication nor ABI/native-memory permission.
func ReplayNativeGeneratedHeader(packagePath, target, language string, sources map[string]string) (map[string]string, error) {
	return nil, fmt.Errorf("native generated-header source replay is not implemented")
}
