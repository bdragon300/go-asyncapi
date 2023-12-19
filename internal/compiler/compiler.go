package compiler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/types"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
	"gopkg.in/yaml.v3"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/samber/lo"
)

const fallbackContentType = "application/json" // Default content type if it omitted in spec

type Object struct {
	Object common.Renderer
	Path   []string
}

func NewCompiler(specID string) *Compiler {
	return &Compiler{
		logger:             types.NewLogger("Compilation 🔨"),
		specID:             specID,
		packagesContents:   make(map[string][]Object),
		defaultContentType: fallbackContentType,
		protocols:          make(map[string]int),
	}
}

type Compiler struct {
	specID        string
	logger        *types.Logger
	remoteSpecIDs []string

	// Set on parsing
	parsedSpecKind SpecKind
	parsedSpec     compiledObject

	// Set during compilation
	packagesContents   map[string][]Object // Objects by package
	defaultContentType string
	protocols          map[string]int
}

func (c *Compiler) Add(pkgName string, stack []string, obj common.Renderer) {
	c.packagesContents[pkgName] = append(c.packagesContents[pkgName], Object{Object: obj, Path: stack})
}

func (c *Compiler) AddProtocol(protoName string) {
	c.protocols[protoName]++
}

func (c *Compiler) AddRemoteSpecID(specID string) {
	c.remoteSpecIDs = append(c.remoteSpecIDs, specID)
}

func (c *Compiler) Protocols() []string {
	return lo.Keys(c.protocols)
}

func (c *Compiler) SetDefaultContentType(contentType string) {
	c.defaultContentType = contentType
}

func (c *Compiler) DefaultContentType() string {
	return c.defaultContentType
}

func (c *Compiler) PackageObjects(pkgName string) []Object {
	return c.packagesContents[pkgName]
}

func (c *Compiler) Packages() []string {
	return lo.Keys(c.packagesContents)
}

func (c *Compiler) AllObjects() []Object {
	return lo.Flatten(lo.Values(c.packagesContents))
}

func (c *Compiler) Load() error {
	var f *os.File
	var err error

	if utils.IsRemoteSpecID(c.specID) {
		c.logger.Info("Download a remote spec", "specID", c.specID)
		f, err = os.CreateTemp(os.TempDir(), "asyncapi-codegen")
		if err != nil {
			return fmt.Errorf("cannot create temp file: %w", err)
		}
		defer f.Close()
		if err = downloadAndWrite(c.specID, f); err != nil {
			return fmt.Errorf("download remote file %q: %w", c.specID, err)
		}
		offset, _ := f.Seek(0, io.SeekCurrent)
		c.logger.Debug("Downloaded file", "specID", c.specID, "bytes", offset, "tmpFile", f.Name())
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("seek on %q: %w", f.Name(), err)
		}
	} else {
		c.logger.Info("Use a local spec", "specID", c.specID)
		f, err = os.Open(c.specID)
		if err != nil {
			return fmt.Errorf("open local file %s: %w", c.specID, err)
		}
		defer f.Close()
	}

	specKind, spec, err := c.parseSpecFile(c.specID, f)
	if err != nil {
		return err
	}
	c.logger.Debug("Schema parsed", "specID", c.specID, "kind", specKind)
	c.parsedSpecKind = specKind
	c.parsedSpec = spec
	return nil
}

func (c *Compiler) Compile(ctx *common.CompileContext) error {
	ctx = ctx.WithResultsStore(c)
	c.logger.Debug("Compile a spec", "specID", c.specID, "kind", c.parsedSpecKind)
	c.logger.Trace("Compile the root component", "specID", c.specID)
	if err := c.parsedSpec.Compile(ctx); err != nil {
		return fmt.Errorf("root component in %s schema: %w", c.parsedSpecKind, err)
	}
	c.logger.Trace("Compile nested components", "specID", c.specID)
	if err := WalkAndCompile(ctx, reflect.ValueOf(c.parsedSpec)); err != nil {
		return fmt.Errorf("spec: %w", err)
	}
	c.logger.Trace("Compile the utils package", "specID", c.specID)
	if err := UtilsCompile(ctx); err != nil {
		return fmt.Errorf("utils package: %w", err)
	}
	return nil
}

func (c *Compiler) RemoteSpecIDs() []string {
	return c.remoteSpecIDs
}

func (c *Compiler) Stats() string {
	return fmt.Sprintf(
		"%s(%s): %d objects in %d packages, known protocols: %s",
		c.specID, c.parsedSpecKind, len(c.AllObjects()), len(c.Packages()), strings.Join(lo.Keys(c.protocols), ","),
	)
}

func (c *Compiler) parseSpecFile(specID string, f *os.File) (SpecKind, compiledObject, error) {
	switch {
	case strings.HasSuffix(specID, ".yaml") || strings.HasSuffix(specID, ".yml"):
		c.logger.Debug("Found YAML spec file", "specID", specID, "file", f.Name())
		schemaKind, spec, err := guessSpecKind(yaml.NewDecoder(f))
		if err != nil {
			return "", nil, fmt.Errorf("guess spec kind: %w", err)
		}
		if _, err = f.Seek(0, io.SeekStart); err != nil {
			return "", nil, fmt.Errorf("file seek: %w", err)
		}
		err = yaml.NewDecoder(f).Decode(spec)
		return schemaKind, spec, err
	case strings.HasSuffix(specID, ".json"):
		c.logger.Debug("Found JSON spec file", "specID", specID, "file", f.Name())
		schemaKind, spec, err := guessSpecKind(json.NewDecoder(f))
		if err != nil {
			return "", nil, fmt.Errorf("guess spec kind: %w", err)
		}
		if _, err = f.Seek(0, io.SeekStart); err != nil {
			return "", nil, fmt.Errorf("file seek: %w", err)
		}
		err = json.NewDecoder(f).Decode(spec)
		return schemaKind, spec, err
	}

	return "", nil, fmt.Errorf("cannot determine format of a spec file: unknown filename extension: %s", specID)
}

func downloadAndWrite(uri string, f *os.File) error {
	// TODO: add additional cli settings such as headers, skip ssl check, follow redirects, allowed http response codes etc.
	resp, err := http.Get(uri)
	if err != nil {
		return fmt.Errorf("make a http request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode > 300 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("error http code: %d", resp.StatusCode)
	}
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("copy http response to file: %w", err)
	}

	return nil
}
