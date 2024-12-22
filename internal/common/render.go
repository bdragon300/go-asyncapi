package common

var context RenderContext

func GetContext() RenderContext {
	return context
}

func SetContext(c RenderContext) {
	context = c
}

// ObjectKind is an enumeration of all possible object kinds used in the AsyncAPI specification.
type ObjectKind string

const (
	ObjectKindOther ObjectKind = ""// Utility language object, not intended for selection (type, value, interface, etc.)
	ObjectKindSchema = "schema"
	ObjectKindServer = "server"
	ObjectKindServerVariable = "serverVariable"
	ObjectKindChannel = "channel"
	ObjectKindMessage = "message"
	ObjectKindParameter = "parameter"
	ObjectKindCorrelationID = "correlationID"
	ObjectKindAsyncAPI = "asyncapi"	         // Utility object represents the entire AsyncAPI document
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
	RenderSelectionConfig struct {
		Template     string
		File         string
		Package      string
		Protocols   []string
		IgnoreCommon bool
		TemplateArgs map[string]string           // TODO: pass template args to templates
		ObjectKindRe string
		ModuleURLRe  string
		PathRe       string
	}
)

type RenderOpts struct {
	RuntimeModule string
	ImportBase    string
	TargetDir    string
	TemplateDir  string
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
