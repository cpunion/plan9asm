package main

// This observer is for a complete selected assembly root, not an include
// fragment. Only the actual Go assembler can close a source rejection.
func preCPPSourceSyntax(file string, data []byte) error {
	_, err := discoveryCPPConditionsFromBytes(file, data)
	return err
}
