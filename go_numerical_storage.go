package plan9asm

import (
	"context"
	"fmt"

	"github.com/xgo-dev/plan9asm/internal/gostaticstorage"
)

// GoNumericalStorageReference requests an actual variable, not a storage span.
type GoNumericalStorageReference struct{ PackagePath, Name string }

type GoNumericalStorageCaptureRequest struct {
	GoBinary, Directory, PackagePath, Target string
	Environment                              []string
	Variables                                []GoNumericalStorageReference
	// SDKCPU requests the separate exact original Go1.27.1 internal/cpu role.
	// It does not admit another Standard or Module=nil package.
	SDKCPU bool
}

// GoNumericalStorageOwner is an opaque live typed data observation. It is not
// a pointer, caller, initialization-completion, constant or memory permission.
// Its zero value and JSON cannot reconstruct the underlying private seal.
type GoNumericalStorageOwner struct{ observation *gostaticstorage.Observation }

func CaptureGoNumericalStorageOwner(ctx context.Context, request GoNumericalStorageCaptureRequest) (*GoNumericalStorageOwner, error) {
	actual := gostaticstorage.Request{GoBinary: request.GoBinary, Directory: request.Directory,
		PackagePath: request.PackagePath, Target: request.Target, Environment: append([]string(nil), request.Environment...)}
	for _, reference := range request.Variables {
		actual.Variables = append(actual.Variables, gostaticstorage.Reference{PackagePath: reference.PackagePath, Name: reference.Name})
	}
	var observation *gostaticstorage.Observation
	var err error
	if request.SDKCPU {
		observation, err = gostaticstorage.CaptureSDKCPU(ctx, actual)
	} else {
		observation, err = gostaticstorage.Capture(ctx, actual)
	}
	if err != nil {
		return nil, numericalSBContext("actual numerical owner: %v", err)
	}
	return &GoNumericalStorageOwner{observation: observation}, nil
}

func (owner *GoNumericalStorageOwner) Verify(ctx context.Context) error {
	if owner == nil || owner.observation == nil {
		return numericalSBContext("missing live numerical storage owner")
	}
	if err := owner.observation.Verify(ctx); err != nil {
		return numericalSBContext("numerical storage owner revoked: %v", err)
	}
	return nil
}

func (owner *GoNumericalStorageOwner) Close() {
	if owner != nil && owner.observation != nil {
		owner.observation.Close()
	}
}

func numericalSBContext(format string, args ...any) error {
	return fmt.Errorf("%w: actual numerical SB read: %s", ErrProbeNeedsContext, fmt.Sprintf(format, args...))
}
