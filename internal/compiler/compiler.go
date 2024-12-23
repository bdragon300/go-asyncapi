package compiler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/specurl"

	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
)

func NewModule(specURL *specurl.URL) *Module {
	return &Module{
		logger:             types.NewLogger("Compilation 🔨"),
		specURL:            specURL,
		objects:            make([]common.CompileObject, 0),
		protocols:          make(map[string]int),
	}
}

type Module struct {
	specURL       *specurl.URL
	logger        *types.Logger
	externalSpecs []*specurl.URL

	// Set on parsing
	parsedSpecKind SpecKind
	parsedSpec     compiledObject

	// Set during compilation
	objects            []common.CompileObject
	protocols          map[string]int
	promises           []common.ObjectPromise
	listPromises       []common.ObjectListPromise
}

func (c *Module) AddObject(obj common.CompileObject) {
	c.objects = append(c.objects, obj)
}

func (c *Module) RegisterProtocol(protoName string) {
	c.protocols[protoName]++
}

func (c *Module) AddExternalSpecPath(specPath *specurl.URL) {
	c.externalSpecs = append(c.externalSpecs, specPath)
}

func (c *Module) AddPromise(p common.ObjectPromise) {
	c.promises = append(c.promises, p)
}

func (c *Module) AddListPromise(p common.ObjectListPromise) {
	c.listPromises = append(c.listPromises, p)
}

func (c *Module) SpecObjectURL() specurl.URL {
	return *c.specURL
}

func (c *Module) Protocols() []string {
	return lo.Keys(c.protocols)
}

func (c *Module) AllObjects() []common.CompileObject {
	return c.objects
}

func (c *Module) Promises() []common.ObjectPromise {
	return c.promises
}

func (c *Module) ListPromises() []common.ObjectListPromise {
	return c.listPromises
}

func (c *Module) Load(specFileResolver SpecFileResolver) error {
	c.logger.Debug("Resolve and load a spec", "specURL", c.specURL)
	buf, err := c.readSpec(c.specURL, specFileResolver)
	if err != nil {
		return err
	}

	c.logger.Trace("Received data", "bytes", len(buf), "data", string(buf))
	specKind, spec, err := c.decodeSpecFile(c.specURL.SpecID, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	c.logger.Debug("Spec parsed", "specURL", c.specURL, "kind", specKind)
	c.parsedSpecKind = specKind
	c.parsedSpec = spec
	return nil
}

func (c *Module) readSpec(specURL *specurl.URL, specFileResolver SpecFileResolver) ([]byte, error) {
	t := time.Now()
	defer func() {
		c.logger.Debug("File resolver finished", "specURL", specURL, "duration", time.Since(t))
	}()

	data, err := specFileResolver.Resolve(specURL)
	if err != nil {
		return nil, fmt.Errorf("resolve spec %q: %w", specURL, err)
	}
	defer data.Close()

	buf, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("read spec data: %w", err)
	}
	return buf, nil
}

func (c *Module) Compile(ctx *common.CompileContext) error {
	ctx = ctx.WithResultsStore(c)
	c.logger.Debug("Compile a spec", "specURL", c.specURL, "kind", c.parsedSpecKind)
	c.logger.Trace("Compile the root component", "specURL", c.specURL)
	if err := c.parsedSpec.Compile(ctx); err != nil {
		return fmt.Errorf("root component in %s schema: %w", c.parsedSpecKind, err)
	}
	c.logger.Trace("Compile nested components", "specURL", c.specURL)
	if err := WalkAndCompile(ctx, reflect.ValueOf(c.parsedSpec)); err != nil {
		return fmt.Errorf("spec: %w", err)
	}
	return nil
}

func (c *Module) ExternalSpecs() []*specurl.URL {
	return c.externalSpecs
}

func (c *Module) Stats() string {
	return fmt.Sprintf(
		"%s(%s): %d objects, known protocols: %s",
		c.specURL, c.parsedSpecKind, len(c.AllObjects()), strings.Join(lo.Keys(c.protocols), ","),
	)
}

func (c *Module) decodeSpecFile(specPath string, data io.ReadSeeker) (SpecKind, compiledObject, error) {
	switch path.Ext(specPath) {
	case ".yaml", ".yml":
		c.logger.Debug("Found YAML spec file", "specPath", specPath)
		schemaKind, spec, err := guessSpecKind(yaml.NewDecoder(data))
		if err != nil {
			return "", nil, fmt.Errorf("guess spec kind: %w", err)
		}
		if _, err = data.Seek(0, io.SeekStart); err != nil {
			return "", nil, fmt.Errorf("file seek: %w", err)
		}
		err = yaml.NewDecoder(data).Decode(spec)
		return schemaKind, spec, err
	case ".json":
		c.logger.Debug("Found JSON spec file", "specPath", specPath)
		schemaKind, spec, err := guessSpecKind(json.NewDecoder(data))
		if err != nil {
			return "", nil, fmt.Errorf("guess spec kind: %w", err)
		}
		if _, err = data.Seek(0, io.SeekStart); err != nil {
			return "", nil, fmt.Errorf("file seek: %w", err)
		}
		err = json.NewDecoder(data).Decode(spec)
		return schemaKind, spec, err
	}

	return "", nil, fmt.Errorf("cannot determine format of a spec file: unknown filename extension: %s", specPath)
}

