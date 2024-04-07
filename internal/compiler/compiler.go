package compiler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/types"
	"gopkg.in/yaml.v3"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/samber/lo"
)

const fallbackContentType = "application/json" // Default content type if it omitted in spec

type Object struct {
	Object common.Renderer
	Path   []string
}

func NewModule(specPath string) *Module {
	return &Module{
		logger:             types.NewLogger("Compilation 🔨"),
		specPath:           specPath,
		objects:            make(map[string][]Object), // Object by rendered code package
		defaultContentType: fallbackContentType,
		protocols:          make(map[string]int),
	}
}

type Module struct {
	specPath          string
	logger            *types.Logger
	externalSpecPaths []string

	// Set on parsing
	parsedSpecKind SpecKind
	parsedSpec     compiledObject

	// Set during compilation
	objects            map[string][]Object // Objects by package
	defaultContentType string
	protocols          map[string]int
	promises           []common.ObjectPromise
	listPromises       []common.ObjectListPromise
	activeServers      []string // Servers in `servers` document section
	activeChannels     []string // Channels in `channels` document section
}

func (c *Module) AddObject(pkgName string, stack []string, obj common.Renderer) {
	c.objects[pkgName] = append(c.objects[pkgName], Object{Object: obj, Path: stack})
}

func (c *Module) RegisterProtocol(protoName string) {
	c.protocols[protoName]++
}

func (c *Module) AddExternalSpecPath(specPath string) {
	c.externalSpecPaths = append(c.externalSpecPaths, specPath)
}

func (c *Module) AddPromise(p common.ObjectPromise) {
	c.promises = append(c.promises, p)
}

func (c *Module) AddListPromise(p common.ObjectListPromise) {
	c.listPromises = append(c.listPromises, p)
}

func (c *Module) Protocols() []string {
	return lo.Keys(c.protocols)
}

func (c *Module) SetDefaultContentType(contentType string) {
	c.defaultContentType = contentType
}

func (c *Module) SetActiveServers(servers []string) {
	c.activeServers = servers
}

func (c *Module) ActiveServers() []string {
	return c.activeServers
}

func (c *Module) SetActiveChannels(channels []string) {
	c.activeChannels = channels
}

func (c *Module) ActiveChannels() []string {
	return c.activeChannels
}

func (c *Module) DefaultContentType() string {
	return c.defaultContentType
}

func (c *Module) PackageObjects(pkgName string) []Object {
	return c.objects[pkgName]
}

func (c *Module) Packages() []string {
	return lo.Keys(c.objects)
}

func (c *Module) AllObjects() []Object {
	return lo.Flatten(lo.Values(c.objects))
}

func (c *Module) Promises() []common.ObjectPromise {
	return c.promises
}

func (c *Module) ListPromises() []common.ObjectListPromise {
	return c.listPromises
}

func (c *Module) Load(specFileResolver SpecFileResolver) error {
	c.logger.Debug("Resolve spec", "specPath", c.specPath)

	t := time.Now()
	data, err := specFileResolver.Resolve(c.specPath)
	if err != nil {
		return fmt.Errorf("resolve spec %q: %w", c.specPath, err)
	}
	defer data.Close()

	buf, err := io.ReadAll(data)
	if err != nil {
		return fmt.Errorf("read spec data: %w", err)
	}
	c.logger.Debug("Spec resolved", "specPath", c.specPath, "duration", time.Since(t))
	c.logger.Trace("Received data", "bytes", len(buf), "data", string(buf))

	specKind, spec, err := c.parseSpec(c.specPath, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	c.logger.Debug("Spec parsed", "specPath", c.specPath, "kind", specKind)
	c.parsedSpecKind = specKind
	c.parsedSpec = spec
	return nil
}

func (c *Module) Compile(ctx *common.CompileContext) error {
	ctx = ctx.WithResultsStore(c)
	c.logger.Debug("Compile a spec", "specPath", c.specPath, "kind", c.parsedSpecKind)
	c.logger.Trace("Compile the root component", "specPath", c.specPath)
	if err := c.parsedSpec.Compile(ctx); err != nil {
		return fmt.Errorf("root component in %s schema: %w", c.parsedSpecKind, err)
	}
	c.logger.Trace("Compile nested components", "specPath", c.specPath)
	if err := WalkAndCompile(ctx, reflect.ValueOf(c.parsedSpec)); err != nil {
		return fmt.Errorf("spec: %w", err)
	}
	if !ctx.CompileOpts.NoEncodingPackage {
		c.logger.Trace("Compile the encoding package", "specPath", c.specPath)
		if err := EncodingCompile(ctx); err != nil {
			return fmt.Errorf("encoding package: %w", err)
		}
	}
	return nil
}

func (c *Module) ExternalSpecPaths() []string {
	return c.externalSpecPaths
}

func (c *Module) Stats() string {
	return fmt.Sprintf(
		"%s(%s): %d objects in %d packages, known protocols: %s",
		c.specPath, c.parsedSpecKind, len(c.AllObjects()), len(c.Packages()), strings.Join(lo.Keys(c.protocols), ","),
	)
}

func (c *Module) parseSpec(specPath string, data io.ReadSeeker) (SpecKind, compiledObject, error) {
	switch {
	case strings.HasSuffix(specPath, ".yaml") || strings.HasSuffix(specPath, ".yml"):
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
	case strings.HasSuffix(specPath, ".json"):
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

