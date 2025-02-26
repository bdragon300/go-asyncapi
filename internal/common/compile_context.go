package common

import (
	"fmt"
	"path"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/log"

	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"

	"github.com/bdragon300/go-asyncapi/internal/types"

	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
)

// CompileArtifact keeps the compiled object and its full URL.
type CompileArtifact struct {
	Renderable
	ObjectURL jsonpointer.JSONPointer
}

// GolangType interface is a Renderable, that represents a primitive Go type, such as struct, map, type alias, etc.
// All of these types are located in [lang] package.
type GolangType interface {
	Renderable
	// CanBeAddressed returns true if value of this type could be addressed. Therefore, we're able to define a pointer
	// to this type, and we can take value's address by applying the & operator.
	//
	// Values that always *not addressable* typically are `nil`, values of interface type, constants, etc.
	CanBeAddressed() bool
	// CanBeDereferenced returns true if this type is a pointer, so we can dereference it by using * operator.
	// True basically means that the type is a pointer as well.
	CanBeDereferenced() bool
	// GoTemplate returns a template name that renders an object of this type.
	GoTemplate() string
}

// CompilationStorage keeps the artifacts of the compilation process, such as compiled objects, promises, etc.
type CompilationStorage interface {
	AddArtifact(obj CompileArtifact)
	AddExternalRef(ref *jsonpointer.JSONPointer)
	AddPromise(p ObjectPromise)
	AddListPromise(p ObjectListPromise)
	DocumentURL() jsonpointer.JSONPointer
}

type CompileOpts struct {
	AllowRemoteRefs     bool
	RuntimeModule       string
	GeneratePublishers  bool
	GenerateSubscribers bool
}

type ContextStackItem struct {
	PathItem       string
	Flags          map[SchemaTag]string
	RegisteredName string // TODO: remove in favor of OriginalName in render structures
}

// NewCompileContext returns a new compilation context with the given document URL and compiler options.
func NewCompileContext(u *jsonpointer.JSONPointer, compileOpts CompileOpts) *CompileContext {
	res := CompileContext{documentURL: u, CompileOpts: compileOpts}
	res.Logger = &CompilerLogger{
		ctx:    &res,
		logger: log.GetLogger(log.LoggerPrefixCompilation),
	}
	return &res
}

// CompileContext keeps the current state of the compilation process. Basically, its functionality is to gather the
// compilation artifacts, maintain the current position in the object tree, keep the config and so on. Context is passed
// to the whole code that compile every AsyncAPI document part.
type CompileContext struct {
	Storage CompilationStorage
	// Stack keeps the current position in the document tree
	Stack       types.SimpleStack[ContextStackItem]
	Logger      *CompilerLogger
	CompileOpts CompileOpts
	documentURL *jsonpointer.JSONPointer
}

// PutObject adds a Renderable as artifact to the storage.
func (c *CompileContext) PutObject(obj Renderable) {
	c.Logger.Debug("Built", "object", obj.String(), "addr", fmt.Sprintf("%p", obj), "type", fmt.Sprintf("%T", obj))
	c.Storage.AddArtifact(CompileArtifact{Renderable: obj, ObjectURL: c.CurrentObjectURL()})
}

// PutPromise adds a promise to the storage. If the promise points to an external document, the document path is also
// added to the list of external documents.
func (c *CompileContext) PutPromise(p ObjectPromise) {
	ref := lo.Must(jsonpointer.Parse(p.Ref()))
	if ref.Location() != "" {
		c.Storage.AddExternalRef(ref)
	}
	c.Storage.AddPromise(p)
}

// PutListPromise adds a list promise to the storage.
func (c *CompileContext) PutListPromise(p ObjectListPromise) {
	c.Storage.AddListPromise(p)
}

// PathStackRef returns the current position in document in URL fragment format, e.g. "#/path/to/object".
// The returned path is URL-encoded.
func (c *CompileContext) PathStackRef(joinParts ...string) string {
	parts := c.pathStack()
	if len(joinParts) > 0 {
		parts = append(parts, joinParts...)
	}
	return jsonpointer.PointerString(parts...)
}

func (c *CompileContext) pathStack() []string {
	return lo.Map(c.Stack.Items(), func(item ContextStackItem, _ int) string { return item.PathItem })
}

// CurrentObjectURL returns the full URL of the current object in the document, i.e. mandatory document path + fragment.
// The fragment is built from the current position in the document. E.g. "file:///path/to/spec.yaml#/path/to/object".
func (c *CompileContext) CurrentObjectURL() jsonpointer.JSONPointer {
	u := c.Storage.DocumentURL()
	u.Pointer = c.pathStack()
	return u
}

func (c *CompileContext) RegisterNameTop(n string) { // TODO: rework and remove this method?
	t := c.Stack.Top()
	t.RegisteredName = n
	c.Stack.ReplaceTop(t)
}

// RuntimeModule returns the import path of the runtime module with optional subpackage,
// e.g. "github.com/bdragon300/go-asyncapi/run/kafka".
func (c *CompileContext) RuntimeModule(subPackage string) string {
	return path.Join(c.CompileOpts.RuntimeModule, subPackage)
}

// GenerateObjName generates a valid Go object name from any given string appending the optional suffix.
// If name is empty, the method generates a name from the current object position in document.
func (c *CompileContext) GenerateObjName(name, suffix string) string {
	if name == "" {
		// Use names of registered object from current stack (that were set by RegisterNameTop call)
		items := lo.FilterMap(c.Stack.Items(), func(item ContextStackItem, _ int) (string, bool) {
			return item.RegisteredName, item.RegisteredName != ""
		})
		// Otherwise if no registered objects in stack, just use path
		if len(items) == 0 {
			items = c.pathStack()
		}
		name = strings.Join(items, "_")
	}
	return utils.ToGolangName(name, true) + suffix
}

func (c *CompileContext) WithResultsStore(store CompilationStorage) *CompileContext {
	res := CompileContext{
		Storage:     store,
		Stack:       types.SimpleStack[ContextStackItem]{},
		documentURL: c.documentURL,
		Logger:      c.Logger,
		CompileOpts: c.CompileOpts,
	}
	res.Logger = &CompilerLogger{
		ctx:    &res,
		logger: log.GetLogger(log.LoggerPrefixCompilation),
	}
	return &res
}
