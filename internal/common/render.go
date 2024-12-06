package common

// ObjectKind is an enumeration of all possible object kinds used in the AsyncAPI specification.
type ObjectKind int

const (
	ObjectKindOther ObjectKind = iota // Utility language object, not intended for selection (type, value, interface, etc.)
	ObjectKindSchema
	ObjectKindServer
	ObjectKindServerVariable
	ObjectKindChannel
	ObjectKindMessage
	ObjectKindParameter
	ObjectKindCorrelationID
	ObjectKindAsyncAPI         // Utility object represents the entire AsyncAPI document
)

type Renderable interface {
	Kind() ObjectKind
	// Selectable returns true if object can be selected to pass to the templates for rendering.
	Selectable() bool
	String() string
}

type (
	RenderSelectionConfig struct {
		Template     string
		File         string
		Package      string
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
	CurrentDefinitionInfo() *GolangTypeDefinitionInfo
}

type ImportItem struct {
	Alias       string
	PackageName string
	PackagePath string
}

