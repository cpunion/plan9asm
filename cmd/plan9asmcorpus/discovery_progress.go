package main

import "fmt"

// The raw shard report keeps the full Go diagnostic. The committed ledger
// keeps only stable, reviewable scope and classification: raw diagnostics can
// contain runner-specific cache paths and differ across otherwise equivalent
// runs.
type discoverySourceSkipSummary struct {
	AsmFile  string   `json:"asm_file,omitempty"`
	AsmFiles []string `json:"asm_files,omitempty"`
	Targets  []string `json:"targets"`
	Kind     string   `json:"kind"`
	Reason   string   `json:"reason"`
}

func discoverySourceSkipReason(kind string) string {
	switch kind {
	case discoverySourceNotApplicableGoAssembler:
		return "Go 1.27 assembler rejected this source on all supported targets"
	case discoverySourceNotApplicableNoGoPackage:
		return "no current Go package selects this assembly source"
	case discoverySourceNotApplicableGoBuild:
		return "Go 1.27 package build rejected this target"
	case discoverySourceNotApplicableAsmDecl:
		return "Go 1.27 asmdecl rejected the source or ABI for this target"
	case discoverySourceNotApplicableNoSymbols:
		return "Go 1.27 assembler emitted no symbols for this source"
	default:
		return ""
	}
}

func summarizeDiscoverySourceSkips(items []discoverySourceNotApplicableItem) []discoverySourceSkipSummary {
	if len(items) == 0 {
		return nil
	}
	summaries := make([]discoverySourceSkipSummary, 0, len(items))
	for _, item := range items {
		summaries = append(summaries, discoverySourceSkipSummary{
			AsmFile:  item.AsmFile,
			AsmFiles: append([]string(nil), item.AsmFiles...),
			Targets:  append([]string(nil), item.Targets...),
			Kind:     item.Kind,
			Reason:   discoverySourceSkipReason(item.Kind),
		})
	}
	return summaries
}

type discoveryCandidateProgress struct {
	Module                    string                                `json:"module"`
	Version                   string                                `json:"version"`
	Status                    string                                `json:"status"`
	Translations              int                                   `json:"translations,omitempty"`
	NotApplicableTranslations int                                   `json:"not_applicable_translations,omitempty"`
	NotApplicableItems        []matrixTargetNotApplicableItem       `json:"not_applicable_items,omitempty"`
	SourceNotApplicableItems  []discoverySourceSkipSummary          `json:"source_not_applicable_items,omitempty"`
	NotApplicableReason       string                                `json:"not_applicable_reason,omitempty"`
	InvalidSourceReason       string                                `json:"invalid_source_reason,omitempty"`
	InvalidSourceEvidence     []discoveryInvalidMachineCodeEvidence `json:"invalid_source_evidence,omitempty"`
	Superseded                *discoverySupersededSkip              `json:"superseded,omitempty"`
	PrivateExtension          *discoveryPrivateExtensionSkip        `json:"private_extension,omitempty"`
	NativeLayout              *discoveryNativeLayoutSkip            `json:"native_layout,omitempty"`
}

// This is a derived view, not a mutable flag attached to a scanned version.
// Keeping test evidence outside the immutable scan ledger avoids making a
// report invalidate its own input fingerprint when it is published.
type discoveryProgress struct {
	SchemaVersion             int                          `json:"schema_version"`
	ValidationKind            string                       `json:"validation_kind"`
	Source                    discoverySourceIdentity      `json:"source"`
	LedgerSHA256              string                       `json:"ledger_sha256"`
	Targets                   []string                     `json:"targets"`
	Provenance                *discoveryCorpusProvenance   `json:"provenance,omitempty"`
	ShardCount                int                          `json:"shard_count"`
	ReportedShards            int                          `json:"reported_shards"`
	PassedShards              int                          `json:"passed_shards"`
	FailedShards              int                          `json:"failed_shards"`
	PartialShards             []int                        `json:"partial_shards"`
	PendingShards             []int                        `json:"pending_shards"`
	CandidateTotal            int                          `json:"candidate_total"`
	Passed                    int                          `json:"passed"`
	Failed                    int                          `json:"failed"`
	NotApplicable             int                          `json:"not_applicable"`
	SkippedInvalidSource      int                          `json:"skipped_invalid_source"`
	SkippedSuperseded         int                          `json:"skipped_superseded"`
	SkippedPrivateExtension   int                          `json:"skipped_private_extension"`
	SkippedNativeLayout       int                          `json:"skipped_native_layout"`
	Pending                   int                          `json:"pending"`
	DiscoveredAsmFiles        int                          `json:"discovered_asm_files"`
	Translations              int                          `json:"translations"`
	NotApplicableTranslations int                          `json:"not_applicable_translations"`
	Complete                  bool                         `json:"complete"`
	Verified                  bool                         `json:"verified"`
	Candidates                []discoveryCandidateProgress `json:"candidates"`
}

func collectDiscoveryProgress(ledgerPath, reportsPath string, targets []string, source discoverySourceIdentity, shardCount int, repoRoot ...string) (discoveryProgress, error) {
	if shardCount < 0 {
		return discoveryProgress{}, fmt.Errorf("invalid discovery shard count %d", shardCount)
	}
	if len(repoRoot) > 1 {
		return discoveryProgress{}, fmt.Errorf("discovery progress accepts at most one repository root")
	}
	var skips map[string]discoveryInvalidMachineCodeSkip
	var superseded map[string]discoverySupersededSkip
	var privateExtensions map[string]discoveryPrivateExtensionSkip
	var nativeLayouts map[string]discoveryNativeLayoutSkip
	if len(repoRoot) == 1 {
		var err error
		skips, err = loadInvalidMachineCodeSkips(repoRoot[0])
		if err != nil {
			return discoveryProgress{}, err
		}
		if skips == nil {
			skips = map[string]discoveryInvalidMachineCodeSkip{}
		}
		superseded, err = loadSupersededSkips(repoRoot[0], ledgerPath)
		if err != nil {
			return discoveryProgress{}, err
		}
		if superseded == nil {
			superseded = map[string]discoverySupersededSkip{}
		}
		privateExtensions, err = loadPrivateExtensionSkips(repoRoot[0], ledgerPath)
		if err != nil {
			return discoveryProgress{}, err
		}
		if privateExtensions == nil {
			privateExtensions = map[string]discoveryPrivateExtensionSkip{}
		}
		nativeLayouts, err = loadNativeLayoutSkips(repoRoot[0], ledgerPath)
		if err != nil {
			return discoveryProgress{}, err
		}
		if nativeLayouts == nil {
			nativeLayouts = map[string]discoveryNativeLayoutSkip{}
		}
	}
	progress := discoveryProgress{
		SchemaVersion:  1,
		ValidationKind: "translation_and_llvm22_object_compilation",
		Source:         source,
		Targets:        append([]string(nil), targets...),
		ShardCount:     shardCount,
		PartialShards:  []int{},
		PendingShards:  []int{},
		Candidates:     []discoveryCandidateProgress{},
	}
	if err := auditDiscoveryCorpusReports(ledgerPath, reportsPath, targets, source, skips, superseded, privateExtensions, nativeLayouts, &progress); err != nil {
		// Never return a plausible partial total after detecting corrupt evidence.
		return discoveryProgress{}, err
	}
	return progress, nil
}
