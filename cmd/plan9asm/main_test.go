package main

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestPackageSFilesAbsFiltersNonPlan9Asm(t *testing.T) {
	dir := t.TempDir()
	abs := filepath.Join(t.TempDir(), "abs", "keep.s")
	pkg := goListPackage{
		Dir: dir,
		SFiles: []string{
			"foo.s",
			"bar.S",
			"baz.Sx",
			abs,
		},
	}
	got := packageSFilesAbs(pkg)
	want := []string{
		filepath.Join(dir, "foo.s"),
		abs,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("packageSFilesAbs() = %#v, want %#v", got, want)
	}
}

func TestGoListPackagesDisablesCgo(t *testing.T) {
	dir := t.TempDir()
	name := "go"
	script := "#!/bin/sh\nprintf '%s\\n' \"{\\\"ImportPath\\\":\\\"$CGO_ENABLED\\\"}\"\n"
	if runtime.GOOS == "windows" {
		name = "go.cmd"
		script = "@echo off\r\necho {\"ImportPath\":\"%CGO_ENABLED%\"}\r\n"
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	pkgs, err := goListPackages("std", "linux", "amd64")
	if err != nil {
		t.Fatalf("goListPackages() error = %v", err)
	}
	if len(pkgs) != 1 || pkgs[0].ImportPath != "0" {
		t.Fatalf("goListPackages() = %#v, want one package with CGO disabled", pkgs)
	}
}
