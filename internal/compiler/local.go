package compiler

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/samber/lo"
)

const fallbackContentType = "application/json" // Default content type if it omitted in spec

type PackageItem struct {
	Typ  common.Renderer // FIXME: rename to Item
	Path []string
}

func NewLocalFile(fileName string) *LocalFile {
	return &LocalFile{
		fileName:           fileName,
		packagesContents:   make(map[string][]PackageItem),
		defaultContentType: fallbackContentType,
		protocols:          make(map[string]int),
	}
}

type LocalFile struct {
	fileName           string
	schemaKind         SchemaKind
	spec               compiledObject
	packagesContents   map[string][]PackageItem
	defaultContentType string
	protocols          map[string]int
}

func (c *LocalFile) Add(pkgName string, stack []string, obj common.Renderer) {
	c.packagesContents[pkgName] = append(c.packagesContents[pkgName], PackageItem{Typ: obj, Path: stack})
}

func (c *LocalFile) AddProtocol(protoName string) {
	c.protocols[protoName]++
}

func (c *LocalFile) Protocols() []string {
	return lo.Keys(c.protocols)
}

func (c *LocalFile) SetDefaultContentType(contentType string) {
	c.defaultContentType = contentType
}

func (c *LocalFile) DefaultContentType() string {
	return c.defaultContentType
}

func (c *LocalFile) DirectRenderItems(pkgName string) []PackageItem {
	return c.packagesContents[pkgName]
}

func (c *LocalFile) Packages() []string {
	return lo.Keys(c.packagesContents)
}

func (c *LocalFile) AllItems() []PackageItem {
	return lo.Flatten(lo.Values(c.packagesContents))
}

func (c *LocalFile) Load() error {
	schemaKind, spec, err := parseSpecFile(c.fileName)
	if err != nil {
		return err
	}
	c.schemaKind = schemaKind
	c.spec = spec
	return nil
}

func (c *LocalFile) Compile(linker common.Linker) error {
	ctx := common.NewCompileContext(linker).WithResultsStore(c)
	logger.Debug("Schema", "kind", c.schemaKind)
	if err := c.spec.Compile(ctx); err != nil {
		return fmt.Errorf("root component in %s schema: %w", c.schemaKind, err)
	}
	if err := WalkAndCompile(ctx, reflect.ValueOf(c.spec)); err != nil {
		return fmt.Errorf("spec: %w", err)
	}
	if err := UtilsCompile(ctx); err != nil {
		return fmt.Errorf("utils package: %w", err)
	}
	return nil
}

func (c *LocalFile) Stats() string {
	return fmt.Sprintf(
		"%s(%s): %d objects in %d packages, protocols: %s",
		c.fileName, c.schemaKind, len(c.AllItems()), len(c.Packages()), strings.Join(lo.Keys(c.protocols), ","),
	)
}
