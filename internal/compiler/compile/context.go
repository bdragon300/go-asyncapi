package compile

import (
	"fmt"
	"path"

	"github.com/bdragon300/go-asyncapi/internal/common"

	"github.com/bdragon300/go-asyncapi/internal/log"

	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"

	"github.com/bdragon300/go-asyncapi/internal/types"

	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
)

// CompilationStorage keeps the artifacts of the compilation process, such as compiled objects, promises, etc.
type CompilationStorage interface {
	AddArtifact(obj common.Artifact)
	AddExternalRef(ref *jsonpointer.JSONPointer)
	AddPromise(p common.ObjectPromise)
	AddListPromise(p common.ObjectListPromise)
	DocumentURL() jsonpointer.JSONPointer
}

type CompilationOpts struct {
	AllowRemoteRefs     bool
	RuntimeModule       string
	GeneratePublishers  bool
	GenerateSubscribers bool
}

type DocumentTreeItem struct {
	// Key is the key of the object in the document tree. For example, a section key in yaml file.
	Key string
	// Flags are the tool's struct tags assigned to the appropriate field.
	Flags map[common.SchemaTag]string
}

type ProtocolBuilder interface {
	Protocol() string
}

// NewCompileContext returns a new compilation context with the given document URL and compiler options.
func NewCompileContext(compileOpts CompilationOpts, protoBuilders []ProtocolBuilder) *Context {
	res := Context{CompileOpts: compileOpts, ProtocolBuilders: protoBuilders}
	res.Logger = &Logger{
		ctx:    &res,
		logger: log.GetLogger(log.LoggerPrefixCompilation),
	}
	return &res
}

// Context keeps the current state of the compilation process. Basically, its functionality is to gather the
// compilation artifacts, maintain the current position in the object tree, keep the config and so on. Context is passed
// to the compilation code.
type Context struct {
	Storage CompilationStorage
	// Stack keeps the current position in the document tree
	Stack            types.SimpleStack[DocumentTreeItem]
	Logger           *Logger
	CompileOpts      CompilationOpts
	ProtocolBuilders []ProtocolBuilder
}

type artifactPointerSetter interface {
	SetPointer(pointer jsonpointer.JSONPointer)
}

// PutArtifact adds an artifact to the storage.
func (c *Context) PutArtifact(obj common.Artifact) {
	// JSON Pointer to the current position in the document
	u := c.Storage.DocumentURL()
	u.Pointer = c.pathStack()
	obj.(artifactPointerSetter).SetPointer(u) // Every artifact must have a SetPointer method
	c.Logger.Debug(
		"Built",
		"object", obj.String(),
		"url", u.String(),
		"memoryAddress", fmt.Sprintf("%p", obj),
		"goType", fmt.Sprintf("%T", obj),
	)

	c.Storage.AddArtifact(obj)
}

// PutPromise adds a promise to the storage. If the promise points to an external document instead of the current
// document, this ref is also gets to the external refs list.
func (c *Context) PutPromise(p common.ObjectPromise) {
	ref := lo.Must(jsonpointer.Parse(p.Ref()))
	if ref.Location() != "" {
		c.Storage.AddExternalRef(ref)
	}
	c.Storage.AddPromise(p)
}

// PutListPromise adds a list promise to the storage.
func (c *Context) PutListPromise(p common.ObjectListPromise) {
	c.Storage.AddListPromise(p)
}

// CurrentPositionRef returns a $ref string to the current position in document, e.g. "#/path/to/object",
// appending the optional parts at the end.
func (c *Context) CurrentPositionRef(extraParts ...string) string {
	parts := append(c.pathStack(), extraParts...)
	return jsonpointer.PointerString(parts...)
}

func (c *Context) pathStack() []string {
	return lo.Map(c.Stack.Items(), func(item DocumentTreeItem, _ int) string { return item.Key })
}

// RuntimeModule returns the import path of the runtime module with optional subpackage,
// e.g. "github.com/bdragon300/go-asyncapi/run/kafka".
func (c *Context) RuntimeModule(subPackage string) string {
	return path.Join(c.CompileOpts.RuntimeModule, subPackage)
}

// GenerateObjName generates a valid Go object name from any given string appending the optional suffix.
// If name is empty, object's key on the current position in document is used.
func (c *Context) GenerateObjName(name, suffix string) string {
	if name == "" {
		name = c.Stack.Top().Key
	}
	return utils.ToGolangName(name, true) + suffix
}

func (c *Context) GetProtocolBuilder(protocol string) (ProtocolBuilder, bool) {
	return lo.Find(c.ProtocolBuilders, func(p ProtocolBuilder) bool { return p.Protocol() == protocol })
}

func (c *Context) SupportedProtocols() []string {
	return lo.Map(c.ProtocolBuilders, func(p ProtocolBuilder, _ int) string { return p.Protocol() })
}

func (c *Context) WithResultsStore(store CompilationStorage) *Context {
	res := Context{
		Storage:          store,
		Stack:            types.SimpleStack[DocumentTreeItem]{},
		Logger:           c.Logger,
		CompileOpts:      c.CompileOpts,
		ProtocolBuilders: c.ProtocolBuilders,
	}
	res.Logger = &Logger{
		ctx:    &res,
		logger: log.GetLogger(log.LoggerPrefixCompilation),
	}
	return &res
}
