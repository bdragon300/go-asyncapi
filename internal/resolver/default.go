package resolver

import (
	"context"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/specurl"

	"github.com/samber/lo"
)

type SpecFileResolver interface {
	Resolve(specPath *specurl.URL) (io.ReadCloser, error)
}

// DefaultSpecFileResolver is the default file resolver. Local spec files are read from the filesystem, remote spec
// files are downloaded via http(s).
type DefaultSpecFileResolver struct {
	Client  *http.Client
	Timeout time.Duration
	BaseDir string // Where to search local specs
	Logger  *log.Logger
}

func (r DefaultSpecFileResolver) Resolve(specPath *specurl.URL) (io.ReadCloser, error) {
	if specPath.IsRemote() {
		return r.resolveRemote(specPath.SpecID)
	}
	return r.resolveLocal(specPath.SpecID)
}

func (r DefaultSpecFileResolver) resolveLocal(specPath string) (io.ReadCloser, error) {
	absPath, err := filepath.Abs(filepath.Join(r.BaseDir, specPath))
	if err != nil {
		return nil, fmt.Errorf("resolve path %q: %w", filepath.Join(r.BaseDir, specPath), err)
	}
	r.Logger.Info("Reading local file", "baseDir", r.BaseDir, "specPath", specPath, "absolutePath", absPath)
	return os.Open(absPath)
}

func (r DefaultSpecFileResolver) resolveRemote(specPath string) (io.ReadCloser, error) {
	r.Logger.Info("Downloading remote file", "specPath", specPath)
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, specPath, nil)
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}
	r.Logger.Debug(
		"HTTP request",
		"method", req.Method,
		"url", req.URL,
		"headers", strings.Join(lo.MapToSlice(req.Header, func(key string, values []string) string {
			return fmt.Sprintf("%s: %s", key, strings.Join(values, ", "))
		}), "\n"),
		"host", req.Host,
	)
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	r.Logger.Debug(
		"HTTP response",
		"status", resp.Status,
		"headers", strings.Join(lo.MapToSlice(resp.Header, func(key string, values []string) string {
			return fmt.Sprintf("%s: %s", key, strings.Join(values, ", "))
		}), "\n"),
	)
	if resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("error http code: %d", resp.StatusCode)
	}
	return resp.Body, nil
}

