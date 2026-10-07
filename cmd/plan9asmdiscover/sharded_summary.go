package main

// readShardedSummary validates a ledger before returning its status view.
func readShardedSummary(reportPath string) (ledgerSummary, error) {
	report, err := readShardedReport(reportPath)
	if err != nil {
		return ledgerSummary{}, err
	}
	return summarizeLedger(report), nil
}
