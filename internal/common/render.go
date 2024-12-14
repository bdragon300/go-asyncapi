package common

// ObjectKind is an enumeration of all possible object kinds used in the AsyncAPI specification.
type ObjectKind string

const (
	ObjectKindOther ObjectKind = ""// Utility language object, not intended for selection (type, value, interface, etc.)
	ObjectKindSchema = "schema"
	ObjectKindServer = "server"
	ObjectKindProtoServer = "protoServer"
	ObjectKindServerVariable = "serverVariable"
	ObjectKindChannel = "channel"
	ObjectKindProtoChannel = "protoChannel"
	ObjectKindMessage = "message"
	ObjectKindProtoMessage = "protoMessage"
	ObjectKindParameter = "parameter"
	ObjectKindCorrelationID = "correlationID"
	ObjectKindAsyncAPI = "asyncapi"	         // Utility object represents the entire AsyncAPI document
)

type Renderable interface {  // TODO: rename
	Kind() ObjectKind
	// Selectable returns true if object can be selected to pass to the templates for rendering.
	Selectable() bool
	String() string
	GetOriginalName() string
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
	CurrentObject() CompileObject
}

type ImportItem struct {
	Alias       string
	PackageName string
	PackagePath string
}

var context RenderContext

func GetContext() RenderContext {
	return context
}

func SetContext(c RenderContext) {
	context = c
}
