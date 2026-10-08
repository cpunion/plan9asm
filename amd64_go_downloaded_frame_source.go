package plan9asm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"time"

	"github.com/xgo-dev/plan9asm/internal/goasmcmd"
	"github.com/xgo-dev/plan9asm/internal/gooriginalsrc"
	"golang.org/x/mod/module"
)

// GoAMD64DownloadedFrameCaptureRequest names original exact cache inputs. It
// carries no observation, checker, object, signature or permission callback.
type GoAMD64DownloadedFrameCaptureRequest struct {
	Original                    GoAMD64FrameCaptureRequest
	SourceModule, SourceVersion string
}

type goAMD64DownloadedIdentity struct {
	module, version string
}

// CaptureGoAMD64DownloadedFrameSource independently executes stock Go and
// closes the original ZIP, extracted members, declarations and whole objects.
// Initial downloaded roles are flat, import-free primary modules and numeric
// or void ABI0 declarations. Capture itself grants no memory permission.
func CaptureGoAMD64DownloadedFrameSource(ctx context.Context, pkg GoPackage, request GoAMD64DownloadedFrameCaptureRequest) (*GoAMD64FrameSource, error) {
	if module.Check(request.SourceModule, request.SourceVersion) != nil ||
		module.CanonicalVersion(request.SourceVersion) != request.SourceVersion {
		return nil, goAMD64FrameContext("downloaded capture requires a canonical exact source identity")
	}
	return captureGoAMD64FrameSource(ctx, pkg, request.Original,
		&goAMD64DownloadedIdentity{module: request.SourceModule, version: request.SourceVersion})
}

func (source *GoAMD64FrameSource) captureDownloadedOrigin(ctx context.Context, exact goAMD64DownloadedIdentity, temporary string) (result error) {
	cache, err := filepath.EvalSymlinks(source.env["GOMODCACHE"])
	if err != nil {
		return err
	}
	origins, err := gooriginalsrc.NewOriginSet(cache, source.request.GoBinary)
	if err != nil {
		return err
	}
	boundary, err := gooriginalsrc.NewDownloadBoundary(temporary, "plan9asm-frame-download-", "example.invalid/plan9asm-frame-download")
	if err != nil {
		return err
	}
	defer func() {
		result = errors.Join(result, boundary.Close())
	}()
	// This private command boundary is not a copied main/source module. Its
	// sole purpose is to prevent stock Go mutating an ancestor go.mod/go.sum.
	if err := boundary.Prepare(); err != nil {
		return err
	}
	commandContext, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	data, err := goasmcmd.Run(commandContext, boundary.Directory(), source.request.Environment,
		source.request.GoBinary, "mod", "download", "-json", exact.module+"@"+exact.version)
	err = errors.Join(err, commandContext.Err())
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	var receipt gooriginalsrc.DownloadReceipt
	if err := decoder.Decode(&receipt); err != nil {
		return err
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return goAMD64FrameContext("actual exact download receipt has incomplete or additional JSON")
	}
	original, err := origins.Observe(ctx, exact.module, exact.version, receipt)
	if err != nil {
		return errors.Join(goAMD64FrameContext("exact original archive/source is not closed"), err)
	}
	if original.Root != source.request.Directory {
		return goAMD64FrameContext("downloaded capture must retain the actual exact primary cache root")
	}
	source.downloadedOrigin, source.downloadedModule = origins, &original
	return source.verifyDownloadedOrigin()
}

func (source *GoAMD64FrameSource) verifyDownloadedOrigin() error {
	if source.downloadedOrigin == nil && source.downloadedModule == nil {
		return nil
	}
	if source.downloadedOrigin == nil || source.downloadedModule == nil ||
		source.downloadedModule.Root != source.request.Directory {
		return goAMD64FrameContext("missing exact original download/source role")
	}
	if err := source.downloadedOrigin.Verify(); err != nil {
		return errors.Join(goAMD64FrameContext("exact original source was changed or revoked"), err)
	}
	return nil
}
