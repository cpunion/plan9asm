package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/xgo-dev/plan9asm/internal/gotoolprofile"
)

func installDiscoveryHeaderPlanning(ctx context.Context, plan *discoveryOrdinarySelectionPlan, candidate discoveryCandidate, download moduleDownloadInfo, dir, goRoot, goBinary, markers, translator string, env []string, cache *discoveryFeatureObservationCache) error {
	if translator == "" {
		return nil // Historical callers retain the conservative all-CPU plan.
	}
	if !filepath.IsAbs(translator) || ctx == nil || cache == nil {
		return fmt.Errorf("native header planning requires the actual bounded query producer")
	}
	if _, bounded := ctx.Deadline(); !bounded || ctx.Err() != nil {
		return fmt.Errorf("native header planning lacks its live bounded context")
	}
	if plan.headerPlanningCapture != nil {
		return nil
	}
	var rules *gotoolprofile.NativeHeaderSourceRules
	module, err := ordinaryProfileDeclaredModule(plan)
	if err != nil {
		return err
	}
	execution, err := makeDiscoveryExecutionPlan(candidate, module)
	if err != nil {
		return err
	}
	queryDir, err := os.MkdirTemp(dir, "native-header-planning-")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(queryDir, "go.mod"), []byte(execution.GoMod), 0600); err != nil {
		return err
	}
	guard, err := captureDiscoveryHeaderQueryWorkspace(queryDir, execution.GoMod)
	if err != nil {
		return err
	}
	check := func() error {
		if err := guard.verify(); err != nil {
			return err
		}
		if rules != nil {
			return gotoolprofile.RecheckNativeHeaderSourceRules(goRoot, rules)
		}
		return nil
	}
	plan.headerPlanningInventory = newDiscoveryFeatureInventory()
	plan.headerPlanningRecheck = check
	plan.headerPlanningCapture = func(request discoveryFeatureProfileRequest, assembly string, tags []string) (header *gotoolprofile.GeneratedHeaderProof, result error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := check(); err != nil {
			return nil, err
		}
		defer func() {
			if err := check(); err != nil {
				header, result = nil, errors.Join(result, err)
			}
		}()
		if rules == nil {
			var err error
			rules, err = gotoolprofile.CaptureNativeHeaderSourceRules(goRoot, plan.GoVersion)
			if err != nil {
				if gotoolprofile.NativeHeaderSourceDomainUnsupported(err) {
					return nil, &discoveryHeaderPlanningUnsupported{err.Error()}
				}
				return nil, err
			}
		}
		if plan.HeaderPlanning != nil {
			if len(plan.HeaderPlanning.Queries) >= 4096 {
				return nil, fmt.Errorf("native header planning exceeds its explicit query bound")
			}
			key := discoveryHeaderPlanningKey(request, assembly, tags)
			for _, query := range plan.HeaderPlanning.Queries {
				if discoveryHeaderPlanningKey(query.Profile.Request, query.Assembly, query.BuildTags) == key {
					return replayDiscoveryPlanningHeader(plan, request, assembly, tags)
				}
			}
		}
		observed, err := cache.observe(ctx, goBinary, markers, env, request.Target, request.Overrides)
		if err != nil {
			return nil, err
		}
		profile := discoveryFeatureProfile{ID: discoveryFeatureProfileID(observed), Request: request, Observed: observed}
		queryPlan, err := discoveryHeaderQuerySourcePlan(plan, profile)
		if err != nil {
			return nil, err
		}
		if err := preflightDiscoveryHeaderQuery(queryPlan, profile, download, assembly, tags); err != nil {
			if gotoolprofile.NativeHeaderSourceDomainUnsupported(err) {
				return nil, &discoveryHeaderPlanningUnsupported{err.Error()}
			}
			return nil, err
		}
		queryPlan.GeneratedHeaders = nil
		queryPlan.GeneratedGoSources = nil
		pattern := module
		if directory := path.Dir(assembly); directory != "." {
			pattern += "/" + directory
		}
		group := discoveryPackageGroup{Pattern: pattern, AsmFiles: []string{assembly}}
		queryEnv := ordinaryProfileEnvironment(env, observed)
		ordinal := 0
		if plan.HeaderPlanning != nil {
			ordinal = len(plan.HeaderPlanning.Queries)
		}
		plan.headerPlanningQueryStarts++
		metadata, err := captureDiscoveryGeneratedHeaderQuery(ctx, queryPlan, download, profile, tags, group, queryDir, queryEnv, translator, ordinal)
		if err != nil {
			return nil, fmt.Errorf("actual native header planning query: %w", err)
		}
		if len(metadata.Headers) != 1 {
			return nil, fmt.Errorf("native header planning lacks its unique actual query")
		}
		if len(metadata.Headers[0].Imports) != 0 {
			return nil, &discoveryHeaderPlanningUnsupported{"actual native header query requires an unclosed imported source model"}
		}
		// Do not publish an unclosed query into the portable role. Independently
		// replay a private candidate first; rejected domains retain all CPUs.
		role := &discoveryCPPHeaderPlanning{Protocol: discoveryCPPHeaderPlanningProtocol, Rules: rules, Sources: make(map[string]string)}
		if plan.HeaderPlanning != nil {
			role.Queries = append(role.Queries, plan.HeaderPlanning.Queries...)
			for name, text := range plan.HeaderPlanning.Sources {
				role.Sources[name] = text
			}
		}
		for name, digest := range metadata.Headers[0].CompiledGoSHA256 {
			if !ordinarySelectionLocalPath(name) || !strings.HasSuffix(name, ".go") {
				return nil, fmt.Errorf("native planning compiler source has an unbound role")
			}
			file := filepath.Join(download.Dir, filepath.FromSlash(name))
			if err := discoveryCPPRegularSource(download.Dir, file); err != nil {
				return nil, err
			}
			text, err := readDiscoveryHeaderPlanningSource(file)
			if err != nil || discoveryFeatureBytesSHA256(text) != digest {
				return nil, fmt.Errorf("native planning source differs from selected original compiler bytes: %v", err)
			}
			if before, present := role.Sources[name]; present && before != string(text) {
				return nil, fmt.Errorf("native planning original source changed between queries")
			}
			role.Sources[name] = string(text)
		}
		if err := validateDiscoveryPlanningSources(role); err != nil {
			return nil, err
		}
		inventory := newDiscoveryFeatureInventory()
		for id, observed := range plan.headerPlanningInventory.Observations {
			inventory.Observations[id] = observed // Already privately frozen by registration.
		}
		references, err := registerDiscoveryFeatureProfiles(inventory, []discoveryFeatureProfile{profile})
		if err != nil {
			return nil, err
		}
		role.Queries = append(role.Queries, discoveryCPPHeaderPlanningQuery{
			Assembly: assembly, Profile: references[0], BuildTags: append([]string(nil), tags...), Metadata: metadata,
		})
		copy := *plan
		copy.HeaderPlanning, copy.headerPlanningInventory = role, inventory
		header, err = replayDiscoveryPlanningHeader(&copy, request, assembly, tags)
		if err != nil {
			if gotoolprofile.NativeHeaderSourceDomainUnsupported(err) {
				return nil, &discoveryHeaderPlanningUnsupported{err.Error()}
			}
			return nil, err
		}
		plan.HeaderPlanning, plan.headerPlanningInventory = role, inventory
		return header, nil
	}
	return nil
}

func readDiscoveryHeaderPlanningSource(name string) ([]byte, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 2<<20+1))
	if err != nil || len(data) > 2<<20 {
		return nil, fmt.Errorf("planning source is unreadable or exceeds its byte bound: %v", err)
	}
	return data, nil
}

type discoveryHeaderQueryWorkspace struct {
	directory string
	physical  string
	dirs      map[string]os.FileInfo
	module    os.FileInfo
	goMod     string
}

func captureDiscoveryHeaderQueryWorkspace(directory, goMod string) (*discoveryHeaderQueryWorkspace, error) {
	physical, err := filepath.EvalSymlinks(directory)
	if err != nil {
		return nil, err
	}
	guard := &discoveryHeaderQueryWorkspace{directory: directory, physical: physical, dirs: make(map[string]os.FileInfo), goMod: goMod}
	for current := physical; ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err != nil || !info.IsDir() {
			return nil, fmt.Errorf("planning query workspace has an unbound directory")
		}
		guard.dirs[current] = info
		if filepath.Dir(current) == current {
			break
		}
	}
	guard.module, err = os.Lstat(filepath.Join(directory, "go.mod"))
	if err != nil || !guard.module.Mode().IsRegular() {
		return nil, fmt.Errorf("planning query workspace lacks its original module file")
	}
	return guard, guard.verify()
}

func (guard *discoveryHeaderQueryWorkspace) verify() error {
	physical, err := filepath.EvalSymlinks(guard.directory)
	if err != nil || physical != guard.physical {
		return fmt.Errorf("planning query workspace alias changed")
	}
	for name, before := range guard.dirs {
		after, err := os.Lstat(name)
		if err != nil || !after.IsDir() || !os.SameFile(before, after) {
			return fmt.Errorf("planning query workspace directory identity changed")
		}
	}
	name := filepath.Join(guard.directory, "go.mod")
	info, err := os.Lstat(name)
	if err != nil || !info.Mode().IsRegular() || !os.SameFile(guard.module, info) {
		return fmt.Errorf("planning query module identity changed")
	}
	data, err := readDiscoveryHeaderPlanningSource(name)
	if err != nil || string(data) != guard.goMod {
		return fmt.Errorf("planning query module bytes changed: %v", err)
	}
	return nil
}

func replayDiscoveryPlanningHeader(plan *discoveryOrdinarySelectionPlan, request discoveryFeatureProfileRequest, assembly string, tags []string) (*gotoolprofile.GeneratedHeaderProof, error) {
	copy := *plan
	copy.headerPlanningCapture = nil
	return discoveryPlanningHeader(&copy, request, assembly, tags)
}
