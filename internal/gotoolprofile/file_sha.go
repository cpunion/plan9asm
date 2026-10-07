package gotoolprofile

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sync"
)

const discoveryFeatureHashBufferSize = 256 << 10

type discoveryFeatureHashBuffer = [discoveryFeatureHashBufferSize]byte

type discoveryFeatureHashBuffers interface {
	Get() any
	Put(any)
}

var discoveryFeatureFileHashBuffers = sync.Pool{
	New: func() any { return new(discoveryFeatureHashBuffer) },
}

func discoveryFeatureFileSHA256(name string) (string, error) {
	f, err := os.Open(name)
	if err != nil {
		return "", err
	}
	return discoveryFeatureReadCloserSHA256(f, &discoveryFeatureFileHashBuffers)
}

func discoveryFeatureReadCloserSHA256(input io.ReadCloser, buffers discoveryFeatureHashBuffers) (string, error) {
	h := sha256.New()
	_, readErr := io.Copy(h, input)
	closeErr := input.Close()
	if readErr != nil {
		return "", readErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
