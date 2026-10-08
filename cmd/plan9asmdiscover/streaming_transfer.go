package main

// Test-first adapter reproduces the existing conversion's eager source read.
func mergeLedgerStreaming(destination, source string) error {
	report, err := readSeenReport(source)
	if err != nil {
		return err
	}
	return writeShardedReportStreaming(destination, report)
}
