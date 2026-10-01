package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/sumdb/tlog"
)

func captureDiscoveryProxyGoMod(ctx context.Context, plan *discoveryOrdinarySelectionPlan, download moduleDownloadInfo, dir string, env []string) error {
	for _, source := range plan.Sources {
		if source.File == "go.mod" {
			return nil // Original ZIP metadata remains on the original protocol.
		}
	}
	if download.Path != plan.Module || download.Version != plan.Version || download.GoModSum == "" || !filepath.IsAbs(download.GoMod) {
		return fmt.Errorf("legacy proxy metadata lacks exact Go download identity and GoModSum")
	}
	output, _, err := runDiscoveryMachineCommand(ctx, dir, env, "go", "env", "-json", "GOMODCACHE")
	if err != nil {
		return err
	}
	var actual map[string]string
	if err := json.Unmarshal(output, &actual); err != nil || len(actual) != 1 || !filepath.IsAbs(actual["GOMODCACHE"]) {
		return fmt.Errorf("invalid actual metadata cache observation")
	}
	escapedModule, err := module.EscapePath(plan.Module)
	if err != nil {
		return err
	}
	escapedVersion, err := module.EscapeVersion(plan.Version)
	if err != nil {
		return err
	}
	cache := filepath.Join(actual["GOMODCACHE"], "cache", "download")
	wanted := filepath.Join(cache, filepath.FromSlash(escapedModule), "@v", escapedVersion+".mod")
	if filepath.Clean(download.GoMod) != wanted {
		return fmt.Errorf("proxy go.mod is not the actual exact-version download metadata")
	}
	info, err := os.Lstat(wanted)
	if err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("proxy metadata is not a regular exact-version download file")
	}
	canonical, err := filepath.EvalSymlinks(wanted)
	if err != nil {
		return fmt.Errorf("resolve actual metadata cache path: %w", err)
	}
	contents, err := readDiscoveryChecksumFile(canonical, 1024)
	if err != nil {
		return fmt.Errorf("read independent proxy go.mod: %w", err)
	}
	reader := &discoverySumDBReader{Context: ctx, Cache: filepath.Join(cache, "sumdb", "sum.golang.org")}
	lookup, err := reader.read("lookup/"+escapedModule+"@"+escapedVersion, 16<<10)
	if err != nil {
		return err
	}
	proof, err := gotoolprofile.CaptureProxyGoMod(plan.Module, plan.Version, plan.ModuleSum, download.GoModSum, contents, lookup, reader)
	if err != nil {
		return err
	}
	plan.ProxyGoMod, plan.proxyGoModPath = proof, canonical
	return verifyDiscoveryProxyGoMod(plan)
}

func verifyDiscoveryProxyGoMod(plan *discoveryOrdinarySelectionPlan) error {
	if plan.ProxyGoMod == nil {
		return nil
	}
	if err := validateDiscoveryProxyGoMod(plan); err != nil {
		return err
	}
	resolved, err := filepath.EvalSymlinks(plan.proxyGoModPath)
	if err != nil || resolved != plan.proxyGoModPath {
		return fmt.Errorf("independent proxy metadata is missing or redirected")
	}
	actual, err := gotoolprofile.FileSHA256(resolved)
	if err != nil || actual != plan.ProxyGoMod.SHA256 {
		return fmt.Errorf("independent proxy metadata changed after authentication")
	}
	return nil
}

func validateDiscoveryProxyGoMod(plan *discoveryOrdinarySelectionPlan) error {
	if plan.ProxyGoMod == nil {
		return nil
	}
	for _, source := range plan.Sources {
		if source.File == "go.mod" {
			return fmt.Errorf("proxy metadata cannot impersonate original ZIP go.mod")
		}
	}
	for _, dir := range plan.Directories {
		if dir.Directory == "." {
			for _, entry := range dir.Entries {
				if entry.Name == "go.mod" {
					return fmt.Errorf("proxy metadata conflicts with original module directory inventory")
				}
			}
		}
	}
	proof := plan.ProxyGoMod
	if proof.Module != plan.Module || proof.Version != plan.Version || proof.ModuleSum != plan.ModuleSum {
		return fmt.Errorf("proxy metadata differs from exact original module ZIP identity")
	}
	return gotoolprofile.ValidateProxyGoMod(proof)
}

// Sumdb responses are public source metadata, never executable dependencies.
// Cache reads are read-only; fetched bytes stay in memory and are authenticated
// before being recorded. No user/global cache is written or cleaned.
type discoverySumDBReader struct {
	Context context.Context
	Cache   string
}

func (reader *discoverySumDBReader) Height() int                     { return 8 }
func (reader *discoverySumDBReader) SaveTiles([]tlog.Tile, [][]byte) {}

func (reader *discoverySumDBReader) ReadTiles(tiles []tlog.Tile) ([][]byte, error) {
	if len(tiles) > 64 {
		return nil, fmt.Errorf("checksum tile request exceeds its explicit bound")
	}
	var result [][]byte
	for _, tile := range tiles {
		if tile.H != 8 || tile.W < 1 || tile.W > 256 {
			return nil, fmt.Errorf("invalid checksum tile extent")
		}
		if reader.Context == nil || reader.Context.Err() != nil {
			return nil, fmt.Errorf("checksum tile request has no live context")
		}
		data, err := readDiscoveryChecksumFile(filepath.Join(reader.Cache, filepath.FromSlash(tile.Path())), tile.W*tlog.HashSize)
		if os.IsNotExist(err) && tile.W < 256 {
			full := tile
			full.W = 256
			data, err = readDiscoveryChecksumFile(filepath.Join(reader.Cache, filepath.FromSlash(full.Path())), 256*tlog.HashSize)
			if err == nil && len(data) == 256*tlog.HashSize {
				data = data[:tile.W*tlog.HashSize]
			}
		}
		if os.IsNotExist(err) {
			data, err = reader.read(tile.Path(), tile.W*tlog.HashSize)
		}
		if err != nil {
			return nil, fmt.Errorf("read exact checksum tile %s: %w", tile.Path(), err)
		}
		if len(data) != tile.W*tlog.HashSize {
			return nil, fmt.Errorf("checksum tile %s has an incorrect byte extent", tile.Path())
		}
		result = append(result, data)
	}
	return result, nil
}

func (reader *discoverySumDBReader) read(name string, limit int) ([]byte, error) {
	if reader.Context == nil || reader.Context.Err() != nil || strings.Contains(name, "..") || strings.HasPrefix(name, "/") {
		return nil, fmt.Errorf("invalid bounded checksum source read")
	}
	data, err := readDiscoveryChecksumFile(filepath.Join(reader.Cache, filepath.FromSlash(name)), limit)
	if err == nil {
		return data, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	request, err := http.NewRequestWithContext(reader.Context, http.MethodGet, "https://sum.golang.org/"+name, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("read signed checksum source: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("checksum source returned HTTP %d", response.StatusCode)
	}
	data, err = io.ReadAll(io.LimitReader(response.Body, int64(limit+1)))
	if err != nil {
		return nil, fmt.Errorf("read checksum source: %w", err)
	}
	if len(data) > limit {
		return nil, fmt.Errorf("checksum source exceeds its explicit bound")
	}
	return data, nil
}

func readDiscoveryChecksumFile(name string, limit int) ([]byte, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, int64(limit+1)))
	if err != nil {
		return nil, err
	}
	if len(data) > limit {
		return nil, fmt.Errorf("checksum cache source exceeds its explicit bound")
	}
	return data, nil
}
