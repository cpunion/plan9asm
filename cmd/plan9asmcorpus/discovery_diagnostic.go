package main

import (
	"regexp"
	"strconv"
	"strings"
)

// net/url.Error quotes the request URL. Require that HTTP request envelope
// instead of retrying any source/file diagnostic that happens to end in EOF.
var discoveryHTTPResponseEOFPattern = regexp.MustCompile(
	`(?m)(?:^|[ :\t])(?:get|head|post|put|delete|patch|options|connect|trace)` +
		` "https?://[^"\r\n]+": eof[ \t]*\r?$`,
)

func isDiscoveryRetryableNetworkFailure(diagnostic string) bool {
	_, retryable := classifyDiscoveryGoBuildFailure(diagnostic)
	return retryable
}

func isDiscoveryGoBuildInfrastructureFailure(diagnostic string) bool {
	infrastructure, _ := classifyDiscoveryGoBuildFailure(diagnostic)
	return infrastructure
}

func classifyDiscoveryGoBuildFailure(diagnostic string) (infrastructure, retryable bool) {
	diagnostic = discoveryNonSourceDiagnostics(diagnostic)
	// These failures need a different fix even if another concurrent package
	// also encounters a transient network error. Never retry past an integrity
	// failure or hide a resource/toolchain failure as source incompatibility.
	for _, marker := range []string{
		"checksum mismatch",
		"security error",
		"captured output exceeds",
		"does not match go tool version",
		"context deadline exceeded",
		"context canceled",
		"executable file not found",
		"exec format error",
		"permission denied",
		"resource temporarily unavailable",
		"out of memory",
		"no space left on device",
		"signal: killed",
	} {
		if strings.Contains(diagnostic, marker) {
			return true, false
		}
	}
	if discoveryHTTPResponseEOFPattern.MatchString(diagnostic) {
		return true, true
	}
	for _, marker := range []string{
		"too many requests",
		"service unavailable",
		"bad gateway",
		"gateway timeout",
		"i/o timeout",
		"tls handshake timeout",
		"connection reset",
		"connection closed by",
		"temporary failure",
		"unexpected eof",
	} {
		if strings.Contains(diagnostic, marker) {
			return true, true
		}
	}
	for _, marker := range []string{
		"no such host",
		"connection refused",
		"could not read from remote repository",
		"network is unreachable",
		"proxyconnect",
	} {
		if strings.Contains(diagnostic, marker) {
			return true, false
		}
	}
	return false, false
}

// A truncated assembly source and a truncated HTTP response both report
// "unexpected EOF". Exclude only the exact source diagnostic, never the
// whole command output: another package may have failed for an independent
// network, resource or integrity reason in that same build.
func discoveryNonSourceDiagnostics(diagnostic string) string {
	normalized := strings.ToLower(diagnostic)
	if !strings.Contains(normalized, "unexpected eof") {
		return normalized
	}

	lines := strings.Split(diagnostic, "\n")
	failedSources := make(map[string]bool)
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		const prefix, suffix = "asm: assembly of ", " failed"
		if strings.HasPrefix(line, prefix) && strings.HasSuffix(line, suffix) {
			filename := strings.TrimSuffix(strings.TrimPrefix(line, prefix), suffix)
			failedSources[filename] = true
		}
	}
	for index, line := range lines {
		if discoveryAssemblerEOFLine(strings.TrimSuffix(line, "\r"), failedSources) {
			lines[index] = ""
		}
	}
	return strings.ToLower(strings.Join(lines, "\n"))
}

func discoveryAssemblerEOFLine(line string, failedSources map[string]bool) bool {
	const suffix = ": unexpected eof"
	if !strings.HasSuffix(strings.ToLower(line), suffix) {
		return false
	}
	location := line[:len(line)-len(suffix)]
	separator := strings.LastIndex(location, ".s:")
	if separator < 0 || !failedSources[location[:separator+2]] {
		return false
	}
	positions := strings.Split(location[separator+3:], ":")
	if len(positions) > 2 {
		return false
	}
	for _, text := range positions {
		if _, err := strconv.ParseUint(text, 10, 32); err != nil {
			return false
		}
	}
	return true
}
