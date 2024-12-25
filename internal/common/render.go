package common

var context RenderContext

func GetContext() RenderContext {
	return context
}

func SetContext(c RenderContext) {
	context = c
}

// ObjectKind is an enumeration of object kinds that are selectable for rendering.
type ObjectKind string

const (
	// ObjectKindOther is a utility language object, not intended for selection (type, value, interface, etc.)
	ObjectKindOther ObjectKind = ""
	ObjectKindSchema = "schema"
	ObjectKindServer = "server"
	ObjectKindChannel = "channel"
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

type (
	RenderSelectionConfigRender struct {
		Template     string
		File         string
		Package      string
		Protocols        []string
		ProtoObjectsOnly bool
	}

	RenderSelectionConfig struct {
		Protocols        []string
		ObjectKinds []string
		ModuleURLRe  string
		PathRe       string
		NameRe	   string
		Render 	 RenderSelectionConfigRender
		ReusePackagePath string

		AllSupportedProtocols []string
	}
)

func (r RenderSelectionConfig) RenderProtocols() []string {
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
	Selections []RenderSelectionConfig
}

type RenderContext interface {
	RuntimeModule(subPackage string) string

	QualifiedName(parts ...string) string
	QualifiedRuntimeName(parts ...string) string
	QualifiedGeneratedPackage(obj GolangType) (string, error)

	CurrentSelection() RenderSelectionConfig
	GetObjectName(obj Renderable) string
	Package() string

	DefineTypeInNamespace(obj GolangType, selection RenderSelectionConfig, actual bool)
	TypeDefinedInNamespace(obj GolangType) bool
	DefineNameInNamespace(name string)
	NameDefinedInNamespace(name string) bool
}

type ImportItem struct {
	Alias       string
	PackageName string
	PackagePath string
}
