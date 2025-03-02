package locator

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/log"

	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"

	"github.com/samber/lo"
)

// Default is the default locator. Local spec files are read from the local filesystem, remote files are downloaded via http(s).
type Default struct {
	// Client is the http client used to download remote specs. If not set, http.DefaultClient is used.
	Client *http.Client
	// Directory is the base directory to read local files from. If empty, the current working directory is used.
	Directory string
	// Logger is the logger used to log locator actions.
	Logger *log.Logger
}

// Locate reads the given document URI. If the URI is remote, it downloads the document via http(s). Returns
// an io.ReadCloser to the document contents.
func (r Default) Locate(docURL *jsonpointer.JSONPointer) (io.ReadCloser, error) {
	if docURL.URI != nil {
		return r.locateHTTP(docURL.Location())
	}
	return r.locateFS(docURL.Location())
}

func (r Default) locateFS(filePath string) (io.ReadCloser, error) {
	dir, _ := lo.Coalesce(r.Directory, ".")
	p := path.Join(dir, filePath)
	r.Logger.Info("Reading document from filesystem", "path", p)
	return os.Open(p)
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
