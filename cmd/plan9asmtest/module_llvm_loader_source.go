package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type moduleLLVMProbeSource struct {
	Parent   string `json:"physical_artifact_parent"`
	Path     string `json:"actual_build_source"`
	Artifact string `json:"phase_source_artifact"`
	SHA      string `json:"source_sha256"`
}

type moduleLLVMProbeInput struct {
	source moduleLLVMProbeSource
	roles  []os.FileInfo
}

func moduleLLVMProbeSHA(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// Only this request's immutable source filename is shared. Every scope still
// invokes actual Go build and executes its own output. No binary, observation,
// source digest, tool route or environment proof is cached here.
func prepareModuleLLVMProbeSource(ctx context.Context, stem string) (input moduleLLVMProbeInput, resultErr error) {
	if err := ctx.Err(); err != nil {
		return input, err
	}
	parent := filepath.Dir(stem)
	info, err := os.Lstat(parent)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || !filepath.IsAbs(parent) {
		return input, fmt.Errorf("LLVM probe requires an owned regular artifact parent: %v", err)
	}
	physical, err := filepath.EvalSymlinks(parent)
	if err != nil {
		return input, err
	}
	actual, err := os.Stat(physical)
	if err != nil || !os.SameFile(info, actual) {
		return input, fmt.Errorf("LLVM probe artifact parent identity changed: %v", err)
	}
	directory := filepath.Join(physical, ".llvm-loader-source")
	if err := os.Mkdir(directory, 0700); err != nil && !os.IsExist(err) {
		return input, err
	}
	directoryInfo, err := os.Lstat(directory)
	if err != nil || !directoryInfo.IsDir() || directoryInfo.Mode()&os.ModeSymlink != 0 {
		return input, fmt.Errorf("LLVM probe source directory is not its regular owned role: %v", err)
	}
	input.source = moduleLLVMProbeSource{Parent: physical, Path: filepath.Join(directory, "main.go"),
		Artifact: filepath.Join(physical, filepath.Base(stem)+"-llvm-loader.go")}
	if _, err := os.Lstat(input.source.Path); os.IsNotExist(err) {
		staged, err := os.CreateTemp(directory, ".pending-*.go")
		if err != nil {
			return input, err
		}
		// Every creator owns only this unique staging file. Atomic hard-link
		// publication exposes complete closed bytes, even to a losing creator.
		defer func() { resultErr = errors.Join(resultErr, os.Remove(staged.Name())) }()
		n, writeErr := io.WriteString(staged, moduleLLVMLoaderProbe)
		closeErr := staged.Close()
		if writeErr == nil && n != len(moduleLLVMLoaderProbe) {
			writeErr = io.ErrShortWrite
		}
		if err := errors.Join(writeErr, closeErr, ctx.Err()); err != nil {
			return input, err
		}
		if err := os.Link(staged.Name(), input.source.Path); err != nil && !os.IsExist(err) {
			return input, err
		}
	} else if err != nil {
		return input, err
	}
	input.roles, input.source.SHA, err = readModuleLLVMProbeSource(input.source)
	if err != nil {
		return input, err
	}
	if !os.SameFile(info, input.roles[0]) || !os.SameFile(directoryInfo, input.roles[1]) {
		return input, fmt.Errorf("LLVM probe parent/source directory identity changed during publication")
	}
	return input, ctx.Err()
}

func readModuleLLVMProbeSource(source moduleLLVMProbeSource) ([]os.FileInfo, string, error) {
	var roles []os.FileInfo
	for index, path := range []string{source.Parent, filepath.Dir(source.Path), source.Path} {
		info, err := os.Lstat(path)
		if err != nil || info.Mode()&os.ModeSymlink != 0 ||
			index < 2 && !info.IsDir() || index == 2 && !info.Mode().IsRegular() {
			return nil, "", fmt.Errorf("LLVM probe physical source role changed: %v", err)
		}
		roles = append(roles, info)
	}
	file, err := os.Open(source.Path)
	if err != nil {
		return nil, "", err
	}
	opened, statErr := file.Stat()
	data, readErr := io.ReadAll(io.LimitReader(file, int64(len(moduleLLVMLoaderProbe)+1)))
	closeErr := file.Close()
	if err := errors.Join(statErr, readErr, closeErr); err != nil {
		return nil, "", err
	}
	if !os.SameFile(opened, roles[2]) || !bytes.Equal(data, []byte(moduleLLVMLoaderProbe)) {
		return nil, "", fmt.Errorf("LLVM probe actual compiled source bytes/identity differ")
	}
	return roles, moduleLLVMProbeSHA(data), nil
}

func (input moduleLLVMProbeInput) verify(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	roles, sha, err := readModuleLLVMProbeSource(input.source)
	if err != nil {
		return err
	}
	for index, role := range roles {
		if !os.SameFile(input.roles[index], role) {
			return fmt.Errorf("LLVM probe original source role replaced")
		}
	}
	if sha != input.source.SHA {
		return fmt.Errorf("LLVM probe source changed during actual build/loader execution")
	}
	return nil
}
