package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// A nonzero exit and an arbitrary nonempty message do not establish source
// incompatibility. Require a concrete compiler/assembler/analyzer source
// location. This is a minimum evidence check, not a new error taxonomy: an
// unrecognized diagnostic remains a failure for investigation, never a skip.
// Known infrastructure failures still take precedence over source diagnostics.
var discoveryConcreteSourceDiagnostic = regexp.MustCompile(
	`(?m)^.*?([^/\\\s:()]+\.(?:go|s)|go\.mod):([1-9][0-9]*)(?::([1-9][0-9]*))?(?::[ \t]*[^ \t\r\n]|\))`,
)

func hasDiscoveryConcreteSourceDiagnostic(diagnostic string) bool {
	return captureDiscoverySourceDiagnostic(diagnostic) != nil
}

const discoverySourceDiagnosticProtocol = "concrete_go_source_diagnostic_v1"

type discoverySourceDiagnosticLocation struct {
	File   string `json:"file"`
	Line   uint32 `json:"line"`
	Column uint32 `json:"column,omitempty"`
}

// Full diagnostics stay in frozen reports. The ledger retains a compact
// producer witness: only source basenames/positions and the diagnostic digest.
// It contains neither runner-specific paths nor arbitrary diagnostic text.
// Like the other offline proofs, it depends on frozen producer provenance;
// the digest alone cannot authenticate a diagnostic without its report bytes.
type discoverySourceDiagnosticWitness struct {
	Protocol  string                              `json:"protocol"`
	SHA256    string                              `json:"sha256"`
	Locations []discoverySourceDiagnosticLocation `json:"locations"`
}

func captureDiscoverySourceDiagnostic(diagnostic string) *discoverySourceDiagnosticWitness {
	if isDiscoveryGoBuildInfrastructureFailure(diagnostic) {
		return nil
	}
	matches := discoveryConcreteSourceDiagnostic.FindAllStringSubmatch(diagnostic, 257)
	if len(matches) == 0 || len(matches) > 256 {
		return nil
	}
	seen := make(map[discoverySourceDiagnosticLocation]bool)
	for _, match := range matches {
		line, err := strconv.ParseUint(match[2], 10, 32)
		if err != nil {
			return nil
		}
		var column uint64
		if match[3] != "" {
			column, err = strconv.ParseUint(match[3], 10, 32)
			if err != nil {
				return nil
			}
		}
		seen[discoverySourceDiagnosticLocation{File: match[1], Line: uint32(line), Column: uint32(column)}] = true
	}
	locations := make([]discoverySourceDiagnosticLocation, 0, len(seen))
	for location := range seen {
		locations = append(locations, location)
	}
	sort.Slice(locations, func(i, j int) bool { return lessDiscoveryDiagnosticLocation(locations[i], locations[j]) })
	digest := sha256.Sum256([]byte(diagnostic))
	witness := &discoverySourceDiagnosticWitness{
		Protocol: discoverySourceDiagnosticProtocol, SHA256: hex.EncodeToString(digest[:]), Locations: locations,
	}
	if validateDiscoverySourceDiagnostic(witness) != nil {
		return nil
	}
	return witness
}

func lessDiscoveryDiagnosticLocation(a, b discoverySourceDiagnosticLocation) bool {
	if a.File != b.File {
		return a.File < b.File
	}
	if a.Line != b.Line {
		return a.Line < b.Line
	}
	return a.Column < b.Column
}

func validateDiscoverySourceDiagnostic(witness *discoverySourceDiagnosticWitness) error {
	if witness == nil || witness.Protocol != discoverySourceDiagnosticProtocol ||
		!discoverySHA256Pattern.MatchString(witness.SHA256) || len(witness.Locations) == 0 || len(witness.Locations) > 256 {
		return fmt.Errorf("missing bounded concrete-source-diagnostic witness")
	}
	for i, location := range witness.Locations {
		if location.File == "" || path.Base(location.File) != location.File ||
			strings.ContainsAny(location.File, "\\:\r\n\t ()") || location.Line == 0 ||
			(location.File != "go.mod" && !strings.HasSuffix(location.File, ".go") && !strings.HasSuffix(location.File, ".s")) ||
			i > 0 && !lessDiscoveryDiagnosticLocation(witness.Locations[i-1], location) {
			return fmt.Errorf("noncanonical or nonportable source-diagnostic location")
		}
	}
	return nil
}

func validateDiscoverySourceRejectionDiagnostic(item discoverySourceNotApplicableItem, compact bool) error {
	if compact {
		if item.Reason != discoverySourceSkipReason(item.Kind) {
			return fmt.Errorf("compact source rejection lacks its stable reason")
		}
		return validateDiscoverySourceDiagnostic(item.Diagnostic)
	}
	actual := captureDiscoverySourceDiagnostic(item.Reason)
	if actual == nil {
		return fmt.Errorf("source rejection lacks a concrete source diagnostic")
	}
	if item.Diagnostic != nil {
		if err := validateDiscoverySourceDiagnostic(item.Diagnostic); err != nil {
			return err
		}
		wanted, _ := json.Marshal(actual)
		observed, _ := json.Marshal(item.Diagnostic)
		if string(wanted) != string(observed) {
			return fmt.Errorf("source diagnostic witness differs from actual report bytes")
		}
	}
	return nil
}

func discoverySourceRejectionRequiresDiagnostic(kind string) bool {
	switch kind {
	case discoverySourceNotApplicableGoBuild,
		discoverySourceNotApplicableGoAssembler,
		discoverySourceNotApplicableAsmDecl,
		nativeLayoutGoAssemblerTarget:
		return true
	default:
		// Selection and empty-object outcomes have their separate source/oracle
		// proof. They must not be confused with a failed compiler invocation.
		return false
	}
}
