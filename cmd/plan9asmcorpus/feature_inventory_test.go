package main

import (
	"context"
	"encoding/json"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestFeatureInventoryCanonicalIdentityAndSharing(t *testing.T) {
	goBinary, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	observed, err := captureDiscoveryTargetFeatures(ctx, goBinary, t.TempDir(), targetFeatureTestEnv(), "linux/amd64", nil)
	if err != nil {
		t.Fatal(err)
	}
	profile := discoveryFeatureProfile{ID: discoveryFeatureProfileID(observed), Request: discoveryFeatureProfileRequest{Target: observed.Target, Baseline: true}, Observed: observed}
	inventory := &discoveryFeatureInventory{}
	first, err := registerDiscoveryFeatureProfiles(inventory, []discoveryFeatureProfile{profile})
	if err != nil {
		t.Fatal(err)
	}
	second, err := registerDiscoveryFeatureProfiles(inventory, []discoveryFeatureProfile{profile})
	if err != nil || len(inventory.Observations) != 1 || !reflect.DeepEqual(first, second) {
		t.Fatalf("identical observations were not shared: %+v %v", inventory, err)
	}
	resolved, err := resolveDiscoveryFeatureProfiles(inventory, first)
	if err != nil || len(resolved) != 1 || resolved[0].ID != profile.ID {
		t.Fatalf("canonical references did not resolve: %+v %v", resolved, err)
	}
	// Caller-owned producer observations must not be aliases into inventory.
	observed.Environment["GOAMD64"] = "v4"
	if inventory.Observations[profile.ID].Environment["GOAMD64"] != "v1" {
		t.Fatal("producer pointer could rewrite shared inventory")
	}
	canonical, err := json.Marshal(inventory)
	if err != nil {
		t.Fatal(err)
	}
	referenceBytes, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	if len(referenceBytes) > 256 || len(canonical) > 8<<10 {
		t.Fatalf("shared feature evidence unexpectedly inflated: inventory=%d reference=%d bytes", len(canonical), len(referenceBytes))
	}
	t.Logf("shared actual-driver inventory=%d bytes, per-candidate baseline reference=%d bytes", len(canonical), len(referenceBytes))
	for name, mutation := range map[string]func(*discoveryFeatureInventory){
		"legacy protocol": func(i *discoveryFeatureInventory) { i.Protocol = "" },
		"same ID changed command bytes": func(i *discoveryFeatureInventory) {
			i.Observations[profile.ID].ListStderrSHA256 = strings.Repeat("e", 64)
		},
		"same ID changed actual target env": func(i *discoveryFeatureInventory) {
			i.Observations[profile.ID].Environment["GOEXPERIMENT"] = "fieldtrack"
		},
		"same ID changed source SHA": func(i *discoveryFeatureInventory) {
			i.Observations[profile.ID].ToolSourceSHA256["VERSION"] = strings.Repeat("e", 64)
		},
		"same ID changed marker bytes": func(i *discoveryFeatureInventory) {
			i.Observations[profile.ID].MarkerSourceSHA256 = strings.Repeat("e", 64)
		},
	} {
		t.Run(name, func(t *testing.T) {
			var changed discoveryFeatureInventory
			if err := json.Unmarshal(canonical, &changed); err != nil {
				t.Fatal(err)
			}
			mutation(&changed)
			if err := validateDiscoveryFeatureInventory(&changed); err == nil {
				t.Fatal("accepted same feature ID with different canonical proof bytes")
			}
		})
	}
	if _, err := resolveDiscoveryFeatureProfiles(inventory, nil); err == nil {
		t.Fatal("accepted absent profile references")
	}
	if _, err := resolveDiscoveryFeatureProfiles(inventory, append(first, first...)); err == nil {
		t.Fatal("accepted duplicate profile references")
	}
	first[0].ID = strings.Repeat("f", 64)
	if _, err := resolveDiscoveryFeatureProfiles(inventory, first); err == nil {
		t.Fatal("accepted unresolved profile references")
	}
}

func TestFeatureInventoryCPUOnlyProfileJSONRoundtrip(t *testing.T) {
	goBinary, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	baseline, err := captureDiscoveryTargetFeatures(ctx, goBinary, t.TempDir(), targetFeatureTestEnv(), "linux/amd64", nil)
	if err != nil {
		t.Fatal(err)
	}
	plan := &discoveryOrdinarySelectionPlan{GoVersion: baseline.GoVersion, Targets: []string{baseline.Target}, Sources: []ordinarySelectionSource{
		{File: "vector_amd64.s", Header: "//go:build amd64.v3\n\n"},
		{File: "vector_amd64.go", Header: "//go:build amd64.v3\n\npackage vector\n"},
	}}
	files := []string{"vector_amd64.s"}
	profiles, err := captureDiscoveryFeatureProfiles(ctx, goBinary, t.TempDir(), targetFeatureTestEnv(), plan, files)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 2 {
		t.Fatalf("actual baseline/v3 source profiles missing: %+v", profiles)
	}
	inventory := &discoveryFeatureInventory{}
	references, err := registerDiscoveryFeatureProfiles(inventory, profiles)
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := json.Marshal(struct {
		Inventory  *discoveryFeatureInventory
		References []discoveryFeatureProfileReference
	}{inventory, references})
	if err != nil {
		t.Fatal(err)
	}
	var replay struct {
		Inventory  *discoveryFeatureInventory
		References []discoveryFeatureProfileReference
	}
	if err := json.Unmarshal(bytes, &replay); err != nil {
		t.Fatal(err)
	}
	resolved, err := resolveDiscoveryFeatureProfiles(replay.Inventory, replay.References)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateDiscoveryFeatureProfiles(plan, files, resolved); err != nil {
		t.Fatalf("actual CPU-only profile lost its proof through JSON roundtrip: %v", err)
	}
}
