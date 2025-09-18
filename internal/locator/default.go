package locator

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/log"

	"github.com/samber/lo"
)

// Default is the default locator. Local spec files are read from the local filesystem, remote files are downloaded via http(s).
type Default struct {
	// Client is the http client used to download remote specs. If not set, http.DefaultClient is used.
	Client *http.Client
	// RootDirectory is the base directory to locate the file paths. Not used if empty.
	RootDirectory string
	// Logger is the logger used to log locator actions.
	Logger *log.Logger
}

// ResolveURL joins the [jsonpointer.JSONPointer] to the base document and ref inside it and returns the pointer to the
// referenced document.
//
// If target is absolute (absolute filesystem path or URL), it is returned as is. If target is relative, it is joined to the base
// document location (filesystem path or URL).
//
// Examples:
//
//	Base: http://example.com/schemas/root.json#/components/schemas/A
//	Target: ../common.json#/components/schemas/B
//	Result: http://example.com/schemas/common.json#/components/schemas/B
//
//	Base: /home/user/schemas/root.json#/components/schemas/A
//	Target: ../common.json#/components/schemas/B
//	Result: /home/user/schemas/common.json#/components/schemas/B
//
//	Base: http://example.com/schemas/root.json#/components/schemas/A
//	Target: /home/user/schemas/common.json#/components/schemas/B
//	Result: /home/user/schemas/common.json#/components/schemas/B
//
//	Base: /home/user/schemas/root.json#/components/schemas/A
//	Target: http://example.com/schemas/common.json#/components/schemas/B
//	Result: http://example.com/schemas/common.json#/components/schemas/B
func (r Default) ResolveURL(base, target *jsonpointer.JSONPointer) (*jsonpointer.JSONPointer, error) {
	return joinBase(r.RootDirectory, base, target)
}

// Locate reads the document that is pointed by p. If the p is http url, it downloads the document via http(s). Returns
// an io.ReadCloser to the document contents.
func (r Default) Locate(p *jsonpointer.JSONPointer) (io.ReadCloser, error) {
	if p.URI != nil {
		return r.locateHTTP(p.Location())
	}

	r.Logger.Info("Reading document from filesystem", "path", p)
	return os.Open(p.FSPath)
}

func (r Default) locateHTTP(filePath string) (io.ReadCloser, error) {
	r.Logger.Info("Downloading remote file", "path", filePath)
	req, err := http.NewRequest(http.MethodGet, filePath, nil)
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

	client, _ := lo.Coalesce(r.Client, http.DefaultClient)
	resp, err := client.Do(req)
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

	// TODO: handle 3xx redirects
	if resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("error http code: %d", resp.StatusCode)
	}
	return resp.Body, nil
}

func joinBase(rootDir string, base, ref *jsonpointer.JSONPointer) (*jsonpointer.JSONPointer, error) {
	if base.URI != nil {
		if ref.URI != nil {
			// Both base and ref are URIs, join them as URLs
			joinedURL := base.URI.ResolveReference(ref.URI)
			return &jsonpointer.JSONPointer{URI: joinedURL, Pointer: ref.Pointer}, nil
		}
		// Base is URI, ref is filesystem path, join as URL with path
		refURL, err := url.Parse(ref.FSPath)
		if err != nil {
			return nil, fmt.Errorf("parse ref as url: %w", err)
		}
		joinedURL := base.URI.ResolveReference(refURL)
		return &jsonpointer.JSONPointer{URI: joinedURL, Pointer: ref.Pointer}, nil
	}
	if ref.URI != nil {
		// Base is filesystem path, ref is URI, return ref as is
		return ref, nil
	}
	// Both base and ref are filesystem paths, join them as paths
	if path.IsAbs(ref.FSPath) {
		// ref is absolute path, return as is
		return ref, nil
	}

	// ref is relative path, join with base directory
	targetPath := ref.FSPath
	basePath := path.Dir(base.FSPath)
	if rootDir != "" {
		basePath = rootDir
		targetPath = path.Clean(path.Join("/", targetPath))
	}
	joinedPath := path.Join(basePath, targetPath)

	return &jsonpointer.JSONPointer{FSPath: joinedPath, Pointer: ref.Pointer}, nil
}
