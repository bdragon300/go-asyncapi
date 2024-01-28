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

func NewModule(specID string) *Module {
	return &Module{
		logger:             types.NewLogger("Compilation 🔨"),
		specID:             specID,
		objects:            make(map[string][]Object), // Object by rendered code package
		defaultContentType: fallbackContentType,
		protocols:          make(map[string]int),
	}
}

type Module struct {
	specID          string
	logger          *types.Logger
	externalSpecIDs []string

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

func (c *Module) AddExternalSpecID(specID string) {
	c.externalSpecIDs = append(c.externalSpecIDs, specID)
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

func (c *Module) Load(specResolver SpecResolver) error {
	c.logger.Debug("Resolve spec", "specID", c.specID)
	t := time.Now()
	data, err := specResolver.Resolve(c.specID)
	if err != nil {
		return fmt.Errorf("resolve spec %q: %w", c.specID, err)
	}
	defer data.Close()
	buf, err := io.ReadAll(data)
	if err != nil {
		return fmt.Errorf("read spec data: %w", err)
	}
	c.logger.Debug("Spec resolved", "specID", c.specID, "duration", time.Since(t))
	c.logger.Trace("Received data", "bytes", len(buf), "data", string(buf))

	specKind, spec, err := c.parseSpec(c.specID, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	c.logger.Debug("Spec parsed", "specID", c.specID, "kind", specKind)
	c.parsedSpecKind = specKind
	c.parsedSpec = spec
	return nil
}

func (c *Module) Compile(ctx *common.CompileContext) error {
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
	if !ctx.CompileOpts.NoEncodingPackage {
		c.logger.Trace("Compile the encoding package", "specID", c.specID)
		if err := EncodingCompile(ctx); err != nil {
			return fmt.Errorf("encoding package: %w", err)
		}
	}
	return nil
}

func (c *Module) ExternalSpecIDs() []string {
	return c.externalSpecIDs
}

func (c *Module) Stats() string {
	return fmt.Sprintf(
		"%s(%s): %d objects in %d packages, known protocols: %s",
		c.specID, c.parsedSpecKind, len(c.AllObjects()), len(c.Packages()), strings.Join(lo.Keys(c.protocols), ","),
	)
}

func (c *Module) parseSpec(specID string, data io.ReadSeeker) (SpecKind, compiledObject, error) {
	switch {
	case strings.HasSuffix(specID, ".yaml") || strings.HasSuffix(specID, ".yml"):
		c.logger.Debug("Found YAML spec file", "specID", specID)
		schemaKind, spec, err := guessSpecKind(yaml.NewDecoder(data))
		if err != nil {
			return "", nil, fmt.Errorf("guess spec kind: %w", err)
		}
		if _, err = data.Seek(0, io.SeekStart); err != nil {
			return "", nil, fmt.Errorf("file seek: %w", err)
		}
		err = yaml.NewDecoder(data).Decode(spec)
		return schemaKind, spec, err
	case strings.HasSuffix(specID, ".json"):
		c.logger.Debug("Found JSON spec file", "specID", specID)
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

	return "", nil, fmt.Errorf("cannot determine format of a spec file: unknown filename extension: %s", specID)
}

