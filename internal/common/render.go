package common

import "github.com/bdragon300/go-asyncapi/implementations"

// ObjectKind is an enumeration of object kinds that are selectable for rendering.
type ObjectKind string

const (
	// ObjectKindOther is a utility language object, not intended for selection (type, value, interface, etc.)
	ObjectKindOther     ObjectKind = ""
	ObjectKindSchema    ObjectKind = "schema"
	ObjectKindServer    ObjectKind = "server"
	ObjectKindChannel   ObjectKind = "channel"
	ObjectKindOperation ObjectKind = "operation"
	ObjectKindMessage   ObjectKind = "message"
	ObjectKindParameter ObjectKind = "parameter"
	// ObjectKindAsyncAPI is a utility object represents the entire AsyncAPI document
	ObjectKindAsyncAPI = "asyncapi"
)

// Renderable is an interface that any compilation artifact matches.
type Renderable interface { // TODO: rename
	// Name returns the name of the object as it was defined in the AsyncAPI document. This method is suitable
	// for rendering the object through a ref. So we can render the object under ref's Name, which is necessary,
	// for example, for rendering servers, channels, etc.
	Name() string
	Kind() ObjectKind
	// Selectable returns true if object can be picked for selections to invoke the template. If false, the object
	// does not get to selections but still can be indirectly rendered inside the templates.
	Selectable() bool
	// Visible returns true if object contents is visible in rendered code.
	Visible() bool
	// String is just a string representation of the object for logging and debugging purposes.
	String() string
}

type renderableWrapper interface {
	UnwrapRenderable() Renderable
}

// DerefRenderable unwraps and the underlying object if it was wrapped in a promise or another wrapper.
func DerefRenderable(obj Renderable) Renderable {
	// TODO: detect ref loops to avoid infinite recursion
	if w, ok := obj.(renderableWrapper); ok {
		return w.UnwrapRenderable()
	}
	return obj
}

// CheckSameRenderables checks if two Renderables are eventually the same object. It extracts the object from the
// promises and wrappers if necessary.
func CheckSameRenderables(a, b Renderable) bool {
	return DerefRenderable(a) == DerefRenderable(b)
}

type (
	ConfigSelectionItem struct {
		Protocols        []string
		ObjectKinds      []string
		ModuleURLRe      string
		PathRe           string
		NameRe           string
		Render           ConfigSelectionItemRender
		ReusePackagePath string

		AllSupportedProtocols []string
	}

	ConfigSelectionItemRender struct {
		Template         string
		File             string
		Package          string
		Protocols        []string
		ProtoObjectsOnly bool
	}

	ConfigImplementationProtocol struct {
		Protocol         string
		Name             string
		Disable          bool
		Directory        string
		Package          string
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
	RuntimeModule     string
	ImportBase        string
	TargetDir         string
	DisableFormatting bool
	PreambleTemplate  string
	Selections        []ConfigSelectionItem
}

type RenderImplementationsOpts struct {
	Disable   bool
	Directory string
	Protocols []ConfigImplementationProtocol
}

type ImplementationObject struct {
	Manifest implementations.ImplManifestItem
	Config   ConfigImplementationProtocol
}
