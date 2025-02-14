package common

import "github.com/bdragon300/go-asyncapi/implementations"

// ObjectKind is an enumeration of object kinds that are selectable for rendering.
type ObjectKind string

const (
	// ObjectKindOther is a utility language object, not intended for selection (type, value, interface, etc.)
	ObjectKindOther ObjectKind = ""
	ObjectKindSchema = "schema"
	ObjectKindServer = "server"
	ObjectKindChannel = "channel"
	ObjectKindOperation = "operation"
	ObjectKindMessage = "message"
	ObjectKindParameter = "parameter"
	// ObjectKindAsyncAPI is a utility object represents the entire AsyncAPI document
	ObjectKindAsyncAPI = "asyncapi"
)

type Renderable interface {  // TODO: rename
	Kind() ObjectKind
	// Selectable returns true if object can be picked for selections to invoke the template. If false, the object
	// does not get to selections but still can be indirectly rendered inside the templates.
	Selectable() bool
	// Visible returns true if object contents is visible in rendered result.
	Visible() bool
	// String is just a string representation of the object for logging and debugging purposes.
	String() string
	// Name returns the name of the object as it was defined in the AsyncAPI document. Suitable if we render
	// the object through a promise -- object's name should be taken from the promise, which is also is Renderable.
	Name() string
}

type renderableWrapper interface {
	UnwrapRenderable() Renderable
}

// TODO: detect ref loops to avoid infinite recursion
func DerefRenderable(obj Renderable) Renderable {
	if w, ok := obj.(renderableWrapper); ok {
		return w.UnwrapRenderable()
	}
	return obj
}

func CheckSameRenderables(a, b Renderable) bool {
	return DerefRenderable(a) == DerefRenderable(b)
}

type (
	ConfigSelectionItem struct {
		Protocols        []string
		ObjectKinds []string
		ModuleURLRe  string
		PathRe       string
		NameRe           string
		Render           ConfigSelectionItemRender
		ReusePackagePath string

		AllSupportedProtocols []string
	}

	ConfigSelectionItemRender struct {
		Template     string
		File         string
		Package      string
		Protocols        []string
		ProtoObjectsOnly bool
	}

	ConfigImplementationProtocol struct {
		Protocol string
		Name      string
		Disable   bool
		Directory string
		Package   string
		ReusePackagePath string
	}

	ConfigInfraServer struct {
		Name           string
		VariableGroups [][]ConfigServerVariable
	}

	ConfigServerVariable struct {
		Name  string
		Value string
	}
)

func (r ConfigSelectionItem) RenderProtocols() []string {
	if len(r.Render.Protocols) > 0 {
		return r.Render.Protocols
	}
	return r.AllSupportedProtocols
}

type RenderOpts struct {
	RuntimeModule string
	ImportBase    string
	TargetDir    string
	DisableFormatting bool
	PreambleTemplate string
	Selections []ConfigSelectionItem
}

type RenderImplementationsOpts struct {
	Disable bool
	Directory string
	Protocols []ConfigImplementationProtocol
}

type ImplementationObject struct {
	Manifest implementations.ImplManifestItem
	Config   ConfigImplementationProtocol
}
