package common

import (
	"fmt"
	"path"

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

type DocumentTreeItem struct {
	// Key is the key of the object in the document tree. For example, a section key in yaml file.
	Key string
	// Flags are the tool's struct tags assigned to the appropriate field.
	Flags map[SchemaTag]string
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
	Stack       types.SimpleStack[DocumentTreeItem]
	Logger      *CompilerLogger
	CompileOpts CompileOpts
	documentURL *jsonpointer.JSONPointer
}

// PutObject adds a Renderable as artifact to the storage.
func (c *CompileContext) PutObject(obj Renderable) {
	// JSON Pointer to the current position in the document
	u := c.Storage.DocumentURL()
	u.Pointer = c.pathStack()

	c.Logger.Debug(
		"Built",
		"object", obj.String(),
		"url", u.String(),
		"memoryAddress", fmt.Sprintf("%p", obj),
		"goType", fmt.Sprintf("%T", obj),
	)
	c.Storage.AddArtifact(CompileArtifact{Renderable: obj, ObjectURL: u})
}

// PutPromise adds a promise to the storage. If the promise points to an external document instead of the current
// document, this ref is also gets to the external refs list.
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

// CurrentPositionRef returns a $ref string to the current position in document, e.g. "#/path/to/object",
// appending the optional parts at the end.
func (c *CompileContext) CurrentPositionRef(extraParts ...string) string {
	parts := append(c.pathStack(), extraParts...)
	return jsonpointer.PointerString(parts...)
}

func (c *CompileContext) pathStack() []string {
	return lo.Map(c.Stack.Items(), func(item DocumentTreeItem, _ int) string { return item.Key })
}

// RuntimeModule returns the import path of the runtime module with optional subpackage,
// e.g. "github.com/bdragon300/go-asyncapi/run/kafka".
func (c *CompileContext) RuntimeModule(subPackage string) string {
	return path.Join(c.CompileOpts.RuntimeModule, subPackage)
}

// GenerateObjName generates a valid Go object name from any given string appending the optional suffix.
// If name is empty, object's key on the current position in document is used.
func (c *CompileContext) GenerateObjName(name, suffix string) string {
	if name == "" {
		name = c.Stack.Top().Key
	}
	return utils.ToGolangName(name, true) + suffix
}

func (c *CompileContext) WithResultsStore(store CompilationStorage) *CompileContext {
	res := CompileContext{
		Storage:     store,
		Stack:       types.SimpleStack[DocumentTreeItem]{},
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
