package gooriginalsrc

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// DependencyQueryEnvironment allows module lookup only in the already observed
// download cache's file-proxy index. GOPROXY=off cannot resolve legacy imports
// even when their exact ZIPs are present. There is no network/direct fallback,
// source rewrite, compiler-version change or checksum-verification exemption.
func DependencyQueryEnvironment(environment []string) ([]string, error) {
	cache := ""
	var result []string
	for _, item := range environment {
		key, value, ok := strings.Cut(item, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("malformed dependency query environment")
		}
		if key == "GOMODCACHE" {
			cache = value
		}
		if key != "GOPROXY" && key != "GONOPROXY" {
			result = append(result, item)
		}
	}
	if !filepath.IsAbs(cache) {
		return nil, fmt.Errorf("dependency query requires an absolute observed module cache")
	}
	path := filepath.ToSlash(filepath.Join(cache, "cache", "download"))
	if !strings.HasPrefix(path, "/") {
		path = "/" + path // Go's Windows file:///C:/... proxy spelling.
	}
	proxy := (&url.URL{Scheme: "file", Path: path}).String()
	// GOPROXY splits these separators before parsing each URL. A literal
	// cache directory name must not create an extra proxy or direct fallback.
	proxy = strings.NewReplacer(",", "%2C", "|", "%7C").Replace(proxy)
	return append(result, "GOPROXY="+proxy, "GONOPROXY=none"), nil
}
