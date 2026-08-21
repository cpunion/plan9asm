package main

import (
	"path/filepath"
	"reflect"
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
